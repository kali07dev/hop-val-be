// services/sync_service.go
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hopekali04/valuations/config"
	"github.com/hopekali04/valuations/models"
	"github.com/hopekali04/valuations/schema"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SyncService struct {
	DB     *gorm.DB
	Client *http.Client 
}

func NewSyncService(db *gorm.DB) *SyncService {
	return &SyncService{
		DB: db,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchAndSyncProperties fetches data from the external API and syncs it.
func (s *SyncService) FetchAndSyncProperties() (*schema.SyncResult, error) {
	cfg := config.GetConfig() // Get loaded config
	if cfg.ExternalAPI.PropertiesURL == "" {
		return nil, errors.New("external API properties URL not configured")
	}

	result := &schema.SyncResult{Status: "in_progress"}
	var allErrors []string
	processedIDs := make(map[uint]bool) // Keep track of processed properties across pages

	nextURL := cfg.ExternalAPI.PropertiesURL // Start with the base URL

	for nextURL != "" { // Loop through pages
		log.Printf("Fetching data from: %s\n", nextURL)
		result.TotalPagesFetched++

		resp, err := s.Client.Get(nextURL)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Failed to fetch data from %s: %v", nextURL, err))
			// Decide if you want to stop on fetch error or try next page (if applicable later)
			// For now, let's stop if a page fetch fails.
			result.Status = "failed"
			result.Errors = allErrors
			result.ErrorCount = len(allErrors)
			return result, fmt.Errorf("failed to fetch page %d: %w", result.TotalPagesFetched, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errMsg := fmt.Sprintf("API request to %s failed with status %d: %s", nextURL, resp.StatusCode, string(bodyBytes))
			allErrors = append(allErrors, errMsg)
			result.Status = "failed"
			result.Errors = allErrors
			result.ErrorCount = len(allErrors)
			// Stop if API returns non-OK status
			return result, errors.New(errMsg)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read response body from %s: %v", nextURL, err)
			allErrors = append(allErrors, errMsg)
			result.Status = "failed"
			result.Errors = allErrors
			result.ErrorCount = len(allErrors)
			return result, fmt.Errorf("failed to read response body for page %d: %w", result.TotalPagesFetched, err)
		}

		var apiResponse schema.PropertiesAPIResponse
		if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
			errMsg := fmt.Sprintf("Failed to unmarshal JSON from %s: %v", nextURL, err)
			allErrors = append(allErrors, errMsg)
			result.Status = "failed"
			result.Errors = allErrors
			result.ErrorCount = len(allErrors)
			return result, fmt.Errorf("failed to unmarshal JSON for page %d: %w", result.TotalPagesFetched, err)
		}

		if result.TotalPagesFetched == 1 { // Set total count from the first page meta
			result.TotalProperties = apiResponse.Meta.Total
		}

		log.Printf("Processing %d properties from page %d...\n", len(apiResponse.Data), apiResponse.Meta.CurrentPage)
		for _, extProp := range apiResponse.Data {
			result.FetchedCount++
			if _, processed := processedIDs[extProp.ID]; processed {
				log.Printf("Skipping already processed property ID %d from a previous page.\n", extProp.ID)
				continue // Skip if already processed in this sync run (safety for overlapping pages)
			}

			syncErr := s.syncSingleProperty(&extProp)
			if syncErr != nil {
				if errors.Is(syncErr, gorm.ErrRecordNotFound) { // Specific check for "already exists"
					log.Printf("Property ID %d already exists, skipping.\n", extProp.ID)
					result.SkippedCount++
				} else {
					errorMsg := fmt.Sprintf("Failed to sync property ID %d: %v", extProp.ID, syncErr)
					log.Println(errorMsg)
					allErrors = append(allErrors, errorMsg)
					result.ErrorCount++
				}
			} else {
				log.Printf("Successfully synced property ID %d.\n", extProp.ID)
				result.SyncedCount++
			}
			processedIDs[extProp.ID] = true // Mark as processed
		}

		// Prepare for the next iteration
		if apiResponse.Meta.NextPageURL != nil {
			nextURL = *apiResponse.Meta.NextPageURL
		} else {
			nextURL = "" // No more pages
		}
	}

	result.Status = "completed"
	if result.ErrorCount > 0 {
		result.Status = "completed_with_errors"
	}
	result.Errors = allErrors

	log.Printf("Sync finished. Fetched: %d, Synced: %d, Skipped: %d, Errors: %d\n",
		result.FetchedCount, result.SyncedCount, result.SkippedCount, result.ErrorCount)

	return result, nil
}

// syncSingleProperty handles the logic for checking and inserting/updating one property and its relations.
// Returns gorm.ErrRecordNotFound if the property already exists (used as a signal to skip).
func (s *SyncService) syncSingleProperty(extProp *schema.ExternalProperty) error {
    // Check if Property already exists by ID
    var existingProperty models.Property
    err := s.DB.Select("id").First(&existingProperty, extProp.ID).Error
    if err == nil {
        return gorm.ErrRecordNotFound // Property found, signal skip
    } else if !errors.Is(err, gorm.ErrRecordNotFound) {
        return fmt.Errorf("failed to check for existing property: %w", err)
    }

    // Use a transaction
    return s.DB.Transaction(func(tx *gorm.DB) error {
        var mappedUser *models.User
        var mappedAgent *models.Agent
        var txErr error

        // --- Upsert User and Agent first (these don't depend on Property) ---
        if extProp.Agent != nil && extProp.Agent.User != nil {
            mappedUser, txErr = s.upsertUser(tx, extProp.Agent.User)
            if txErr != nil {
                return fmt.Errorf("failed to upsert user %d for property %d: %w", extProp.Agent.User.ID, extProp.ID, txErr)
            }

            if mappedUser != nil {
                mappedAgent, txErr = s.upsertAgent(tx, extProp.Agent, mappedUser.ID)
                if txErr != nil {
                    return fmt.Errorf("failed to upsert agent %d for property %d: %w", extProp.Agent.ID, extProp.ID, txErr)
                }
            }
        }

        // --- Upsert Cover Photo that doesn't depend on Property ---
        var mappedCoverPhoto *models.CoverPhoto
        if extProp.CoverPhoto != nil {
            mappedCoverPhoto, txErr = s.upsertCoverPhoto(tx, extProp.CoverPhoto)
            if txErr != nil {
                return fmt.Errorf("failed to upsert cover photo %d for property %d: %w", extProp.CoverPhoto.ID, extProp.ID, txErr)
            }
        }

        // --- Map and Create Property first ---
        dbProperty, mapErr := mapExternalToDBProperty(extProp, mappedAgent, mappedCoverPhoto)
        if mapErr != nil {
            return fmt.Errorf("failed to map external property %d to db model: %w", extProp.ID, mapErr)
        }

        // Create the property without associations
        result := tx.Omit(clause.Associations).Create(&dbProperty)
        if result.Error != nil {
            return fmt.Errorf("failed to create property %d: %w", dbProperty.ID, result.Error)
        }

        // --- Now that property exists, upsert Location ---
        if extProp.Location != nil {
            extProp.Location.PropertyID = extProp.ID
            _, txErr = s.upsertLocation(tx, extProp.Location)
            if txErr != nil {
                return fmt.Errorf("failed to upsert location %d for property %d: %w", extProp.Location.ID, extProp.ID, txErr)
            }
        }

        return nil
    })
}

// Helper to parse string dates from API (adjust format if needed)
func parseAPITime(apiTime *string) (*time.Time, error) {
	if apiTime == nil || *apiTime == "" {
		return nil, nil
	}
	// Try different common formats if necessary
	layouts := []string{
		time.RFC3339Nano,      // "2025-04-01T11:44:25.000000Z"
		"2006-01-02 15:04:05", // "2025-04-01 12:31:28" (for approved_at)
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, *apiTime)
		if err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("failed to parse time string: %s", *apiTime)
}

// Helper function to map ExternalProperty to models.Property
func mapExternalToDBProperty(extProp *schema.ExternalProperty, agent *models.Agent, coverPhoto *models.CoverPhoto) (models.Property, error) {
	prop := models.Property{
		ID:                extProp.ID, // Explicitly set ID
		ValuerID:          extProp.ValuerID,
		PropertyNumber:    extProp.PropertyNumber,
		ParentValuation:   extProp.ParentValuation,
		ProjectID:         extProp.ProjectID,
		OwnerName:         extProp.OwnerName,
		PropertyType:      extProp.PropertyType,
		PropertyDesign:    extProp.PropertyDesign,
		ConstructionStage: extProp.ConstructionStage,
		YearBuilt:         extProp.YearBuilt,
		Age:               extProp.Age,
		Eul:               extProp.Eul,
		Rel:               extProp.Rel,
		Measurements:      extProp.Measurements,
		NoRooms:           extProp.NoRooms,
		NoOfBathrooms:     extProp.NoOfBathrooms,
		Occupancy:         extProp.Occupancy,
		// Attributes JSON string needs to be converted to datatypes.JSON
		Attributes:                   datatypes.JSON(extProp.Attributes),
		TitleDeedsAvailable:          extProp.TitleDeedsAvailable,
		CertificateOfSearchAvailable: extProp.CertificateOfSearchAvailable,
		EncumbrancesAvailable:        extProp.EncumbrancesAvailable,
		Defects:                      extProp.Defects,
		Description:                  extProp.Description,
		MasterBedroomEnsuite:         extProp.MasterBedroomEnsuite,
		BuildingSize:                 extProp.BuildingSize,
		BuildingSizeUnit:             extProp.BuildingSizeUnit,
		LandSize:                     extProp.LandSize,
		LandSizeUnit:                 extProp.LandSizeUnit,
		EntryType:                    extProp.EntryType,
		Price:                        extProp.Price,
		ListingType:                  extProp.ListingType,
		CreatedBy:                    extProp.CreatedBy,
		IsApproved:                   extProp.IsApproved,
		IsSubmitted:                  extProp.IsSubmitted,
		IsSaleCompleted:              extProp.IsSaleCompleted,
		HasAcceptedOffer:             extProp.HasAcceptedOffer,
		IsReferred:                   extProp.IsReferred,
		Visibility:                   extProp.Visibility,
		Views:                        extProp.Views,
	}

	// Parse time strings
	createdAt, err := parseAPITime(&extProp.CreatedAt)
	if err != nil {
		log.Printf("Warning: Could not parse CreatedAt for property %d: %v", extProp.ID, err)
		// Decide default or skip - let's default to Now() for DB if unparseable
		prop.CreatedAt = time.Now()
	} else if createdAt != nil {
		prop.CreatedAt = *createdAt
	}

	updatedAt, err := parseAPITime(&extProp.UpdatedAt)
	if err != nil {
		log.Printf("Warning: Could not parse UpdatedAt for property %d: %v", extProp.ID, err)
		prop.UpdatedAt = time.Now()
	} else if updatedAt != nil {
		prop.UpdatedAt = *updatedAt
	}

	approvedAt, err := parseAPITime(extProp.ApprovedAt)
	if err != nil {
		log.Printf("Warning: Could not parse ApprovedAt for property %d: %v", extProp.ID, err)
		prop.ApprovedAt = nil
	} else {
		prop.ApprovedAt = approvedAt // Assign pointer directly
	}

	// Link Agent via AgentID if agent was successfully upserted/found
	if agent != nil {
		prop.AgentID = &agent.ID
		// Don't assign prop.Agent = *agent here if using FK; GORM handles loading
	}


	return prop, nil // Return nil error if mapping succeeds
}

// --- Upsert Helper Functions ---

func (s *SyncService) upsertUser(tx *gorm.DB, extUser *schema.ExternalUser) (*models.User, error) {
	if extUser == nil {
		return nil, errors.New("cannot upsert nil user")
	}

	// Map ExternalUser to models.User
	user := models.User{
		ID:                     extUser.ID, // Set ID for lookup/update
		Name:                   extUser.Name,
		Email:                  extUser.Email,
		Status:                 extUser.Status,
		Phone:                  extUser.Phone,
		FinancialInstitutionID: extUser.FinancialInstitutionID,
		Role:                   extUser.Role,
		SignatureStorageURL:    extUser.SignatureStorageURL,
		ProfileImageStorageURL: extUser.ProfileImageStorageURL,
	}

	// Parse time strings for User
	createdAt, err := parseAPITime(&extUser.CreatedAt)
	if err == nil && createdAt != nil {
		user.CreatedAt = *createdAt
	} else { /* handle error or default */
	}
	updatedAt, err := parseAPITime(&extUser.UpdatedAt)
	if err == nil && updatedAt != nil {
		user.UpdatedAt = *updatedAt
	} else { /* handle error or default */
	}
	emailVerifiedAt, err := parseAPITime(extUser.EmailVerifiedAt)
	if err == nil {
		user.EmailVerifiedAt = emailVerifiedAt
	} else { /* handle error or default */
	}
	pwLastUpdAt, err := parseAPITime(extUser.PasswordLastUpdatedAt)
	if err == nil {
		user.PasswordLastUpdatedAt = pwLastUpdAt
	} else { /* handle error or default */
	}
	// Handle DeletedAt if needed

	// Use Clauses(clause.OnConflict) for Upsert
	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, // Conflict on primary key
		DoUpdates: clause.AssignmentColumns([]string{ // List columns to update on conflict
			"name", "email", "status", "phone", "financial_institution_id", "role",
			"signature_storage_url", "profile_image_storage_url", "updated_at",
			"email_verified_at", "password_last_updated_at", // Add other updatable fields
		}),
	}).Create(&user).Error // Create attempts insert, OnConflict handles update

	if err != nil {
		return nil, fmt.Errorf("upsert user failed: %w", err)
	}
	return &user, nil // Return the upserted user
}

func (s *SyncService) upsertAgent(tx *gorm.DB, extAgent *schema.ExternalAgent, userID uint) (*models.Agent, error) {
	if extAgent == nil {
		return nil, errors.New("cannot upsert nil agent")
	}

	// Convert API's string UserID to uint (handle potential error)
	apiUserID, err := strconv.ParseUint(extAgent.UserID, 10, 32)
	if err != nil {
		log.Printf("Warning: Could not parse Agent UserID '%s' for agent %d: %v. Using linked User ID %d.", extAgent.UserID, extAgent.ID, err, userID)
		// Fallback or error handling strategy needed here.
		// For now, we trust the userID passed from the upserted user.
	} else if uint(apiUserID) != userID {
		// This indicates a potential data inconsistency between nested user and agent user_id
		log.Printf("Warning: Agent %d UserID mismatch. API Agent.UserID ('%s') != API Agent.User.ID (%d). Using User ID %d.", extAgent.ID, extAgent.UserID, userID, userID)
	}

	agent := models.Agent{
		ID:                extAgent.ID, // Set ID for lookup/update
		UserID:            userID,      // Use the ID from the already upserted user
		Phone1:            extAgent.Phone1,
		Phone2:            extAgent.Phone2,
		Headline1:         extAgent.Headline1,
		Headline2:         extAgent.Headline2,
		About:             extAgent.About,
		IsAgreementSigned: extAgent.IsAgreementSigned,
		AgentType:         extAgent.AgentType,
		BankName:          extAgent.BankName,
		AccountName:       extAgent.AccountName,
		AccountNumber:     extAgent.AccountNumber,
		AccountType:       extAgent.AccountType,
		AccountBranch:     extAgent.AccountBranch,
		Linkedin:          extAgent.Linkedin,
		Address:           extAgent.Address,
		CoverageArea:      extAgent.CoverageArea,
	}

	// Parse time strings for Agent
	createdAt, err := parseAPITime(&extAgent.CreatedAt)
	if err == nil && createdAt != nil {
		agent.CreatedAt = *createdAt
	} else { /* handle error or default */
	}
	updatedAt, err := parseAPITime(&extAgent.UpdatedAt)
	if err == nil && updatedAt != nil {
		agent.UpdatedAt = *updatedAt
	} else { /* handle error or default */
	}

	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{ // Update relevant fields
			"user_id", "phone1", "phone2", "headline1", "headline2", "about",
			"is_agreement_signed", "agent_type", "bank_name", "account_name",
			"account_number", "account_type", "account_branch", "linkedin",
			"address", "coverage_area", "updated_at",
		}),
	}).Create(&agent).Error

	if err != nil {
		return nil, fmt.Errorf("upsert agent failed: %w", err)
	}
	return &agent, nil
}

func (s *SyncService) upsertLocation(tx *gorm.DB, extLocation *schema.ExternalLocation) (*models.Location, error) {
	if extLocation == nil {
		return nil, errors.New("cannot upsert nil location")
	}

	location := models.Location{
		ID:            extLocation.ID,
		PropertyID:    extLocation.PropertyID, // This should match the target property ID
		Region:        extLocation.Region,
		District:      extLocation.District,
		Area:          extLocation.Area,
		Postcode:      extLocation.Postcode,
		SubArea:       extLocation.SubArea,
		GoogleMapLink: extLocation.GoogleMapLink,
		Latitude:      extLocation.Latitude,
		Longitude:     extLocation.Longitude,
		ZoneCategory:  extLocation.ZoneCategory,
		Zoning:        extLocation.Zoning,
	}

	// Parse time strings for Location
	createdAt, err := parseAPITime(&extLocation.CreatedAt)
	if err == nil && createdAt != nil {
		location.CreatedAt = *createdAt
	} else { /* handle error or default */
	}
	updatedAt, err := parseAPITime(&extLocation.UpdatedAt)
	if err == nil && updatedAt != nil {
		location.UpdatedAt = *updatedAt
	} else { /* handle error or default */
	}


	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, // Assuming Location ID is PK
		// Alternative if composite key (property_id, some_other_field) or if ID isn't reliable
		// Columns:   []clause.Column{{Name: "property_id"}}, // Example: conflict on property_id
		DoUpdates: clause.AssignmentColumns([]string{
			"property_id", "region", "district", "area", "postcode", "sub_area",
			"google_map_link", "latitude", "longitude", "zone_category", "zoning",
			"updated_at",
		}),
	}).Create(&location).Error

	if err != nil {
		return nil, fmt.Errorf("upsert location failed: %w", err)
	}
	return &location, nil
}

func (s *SyncService) upsertCoverPhoto(tx *gorm.DB, extPhoto *schema.ExternalCoverPhoto) (*models.CoverPhoto, error) {
	if extPhoto == nil {
		return nil, errors.New("cannot upsert nil cover photo")
	}

	photo := models.CoverPhoto{
		ID:          extPhoto.ID,
		Url:         extPhoto.Url,
		Description: extPhoto.Description,

	}

	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"url", "description"}),
	}).Create(&photo).Error

	if err != nil {
		return nil, fmt.Errorf("upsert cover photo failed: %w", err)
	}
	return &photo, nil
}
