// services/property_service.go
package services

import (
	"errors"
	"fmt"

	"github.com/hopekali04/valuations/models"
	"github.com/hopekali04/valuations/schema"
	"github.com/hopekali04/valuations/utils"
	"gorm.io/gorm"
)

type PropertyService struct {
	DB *gorm.DB
}

func NewPropertyService(db *gorm.DB) *PropertyService {
	return &PropertyService{DB: db}
}

// --- Create Operations ---

// CreateProperty handles creation of a single property, checking for duplicates.
func (s *PropertyService) CreateProperty(req *schema.CreatePropertyRequest) (*models.Property, error) {
	// 1. Check if property with this ID already exists
	var existing models.Property
	// Use .Unscoped() if you might have soft-deleted records with the same ID you want to prevent reusing
	err := s.DB.Select("id").First(&existing, req.ID).Error
	if err == nil {
		// Record found, return specific conflict error
		return nil, utils.ErrPropertyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// An actual DB error occurred during the check
		return nil, fmt.Errorf("failed to check for existing property: %w", err)
	}

	newProperty := models.Property{
		ID:                           req.ID, // Set explicitly from request
		ValuerID:                     req.ValuerID,
		PropertyNumber:               req.PropertyNumber,
		ParentValuation:              req.ParentValuation,
		ProjectID:                    req.ProjectID,
		OwnerName:                    req.OwnerName,
		PropertyType:                 req.PropertyType,
		PropertyDesign:               req.PropertyDesign,
		ConstructionStage:            req.ConstructionStage,
		YearBuilt:                    req.YearBuilt,
		Age:                          req.Age,
		Eul:                          req.Eul,
		Rel:                          req.Rel,
		Measurements:                 req.Measurements,
		NoRooms:                      req.NoRooms,
		NoOfBathrooms:                req.NoOfBathrooms,
		Occupancy:                    req.Occupancy,
		Attributes:                   req.Attributes,
		TitleDeedsAvailable:          req.TitleDeedsAvailable,
		CertificateOfSearchAvailable: req.CertificateOfSearchAvailable,
		EncumbrancesAvailable:        req.EncumbrancesAvailable,
		Defects:                      req.Defects,
		Description:                  req.Description,
		MasterBedroomEnsuite:         req.MasterBedroomEnsuite,
		BuildingSize:                 req.BuildingSize,
		BuildingSizeUnit:             req.BuildingSizeUnit,
		LandSize:                     req.LandSize,
		LandSizeUnit:                 req.LandSizeUnit,
		EntryType:                    req.EntryType,
		Price:                        req.Price,
		ListingType:                  req.ListingType,
		CreatedBy:                    req.CreatedBy,
		IsApproved:                   req.IsApproved,
		IsSubmitted:                  req.IsSubmitted,
		IsSaleCompleted:              req.IsSaleCompleted,
		HasAcceptedOffer:             req.HasAcceptedOffer,
		IsReferred:                   req.IsReferred,
		ApprovedAt:                   req.ApprovedAt,
		Visibility:                   req.Visibility,
		Views:                        req.Views,

	}

	// If Location data is provided in the request, assign it. GORM handles association.
	if req.Location != nil {

		req.Location.PropertyID = newProperty.ID

		newProperty.Location = *req.Location
	}

	// Similarly for CoverPhoto
	if req.CoverPhoto != nil {
		// Set the PropertyID for the cover photo
		req.CoverPhoto.PropertyID = newProperty.ID
		newProperty.CoverPhoto = *req.CoverPhoto
	}

	result := s.DB.Omit("Agent", "CoverPhoto.PropertyID", "Location.PropertyID").Create(&newProperty) // Omit relations that are handled via FK fields
	if result.Error != nil {

		return nil, fmt.Errorf("failed to create property: %w", result.Error)
	}


	err = s.DB.Preload("Location").
		Preload("Agent.User"). // Preload User within Agent
		Preload("CoverPhoto").
		First(&newProperty, newProperty.ID).Error
	if err != nil {
		// Log this error, but potentially still return the created property without associations
		fmt.Printf("Warning: failed to preload associations for created property %d: %v\n", newProperty.ID, err)
		// return &newProperty, nil // Return without associations
		return nil, fmt.Errorf("failed to fetch created property with associations: %w", err) // Or fail fully
	}

	return &newProperty, nil
}

// CreateMultipleProperties handles bulk creation with individual error reporting.
func (s *PropertyService) CreateMultipleProperties(requests []schema.CreatePropertyRequest) ([]models.Property, []schema.BulkErrorDetail) {
	var successfulProperties []models.Property
	var errorsReport []schema.BulkErrorDetail

	for _, req := range requests {
		// Use a pointer for the request in the loop if modifying it (not needed here)
		propertyReq := req // Create a copy for this iteration

		// Try to create the property using the single create logic
		createdProperty, err := s.CreateProperty(&propertyReq) // Pass address of the copy

		if err != nil {
			// Log the error and add to the error report
			var errMsg string
			var customErr *schema.CustomError
			if errors.As(err, &customErr) {
				errMsg = fmt.Sprintf("%s: %s", customErr.Message, customErr.Details)
			} else {
				errMsg = err.Error()
			}
			errorsReport = append(errorsReport, schema.BulkErrorDetail{
				PropertyID: propertyReq.ID, // Use the ID from the request
				Error:      errMsg,
			})
		} else if createdProperty != nil {
			// Add successfully created property to the list
			successfulProperties = append(successfulProperties, *createdProperty)
		}
	}

	return successfulProperties, errorsReport
}

// --- Read Operations ---

// GetPropertyByID retrieves a single property by its ID with associations.
func (s *PropertyService) GetPropertyByID(id uint) (*models.Property, error) {
	var property models.Property
	// Preload associations for the response
	err := s.DB.Preload("Location").
		Preload("Agent.User").
		Preload("CoverPhoto").
		// Preload("OpenHouses"). // Add if OpenHouse model exists
		First(&property, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError("Property")
		}
		return nil, fmt.Errorf("database error retrieving property: %w", err)
	}
	return &property, nil
}

// GetAllProperties retrieves properties with pagination and filtering.
func (s *PropertyService) GetAllProperties(pag schema.PaginationRequest, filter schema.PropertyFilter) ([]models.Property, int64, error) {
	var properties []models.Property
	var totalItems int64

	query := s.DB.Model(&models.Property{})

	// Apply Filters
	query = applyPropertyFilters(query, filter)

	// Get total count (before pagination)
	err := query.Count(&totalItems).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count properties: %w", err)
	}

	// Apply pagination and preloads
	err = query.Scopes(utils.PaginateScope(pag.Page, pag.PageSize)).
		Order("created_at DESC"). // Example ordering
		Preload("Location").
		Preload("Agent.User").
		Preload("CoverPhoto").
		Find(&properties).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve properties: %w", err)
	}

	return properties, totalItems, nil
}

// SearchProperties performs a simple text search across relevant fields.
func (s *PropertyService) SearchProperties(searchTerm string, pag schema.PaginationRequest) ([]models.Property, int64, error) {
	var properties []models.Property
	var totalItems int64

	// Base query
	query := s.DB.Model(&models.Property{})

	// Apply search condition (case-insensitive)
	if searchTerm != "" {
		// Join with location to search location fields too
		query = query.Joins("JOIN locations ON locations.property_id = properties.id").
			Joins("LEFT JOIN agents ON agents.id = properties.agent_id"). // LEFT JOIN in case agent is null
			Joins("LEFT JOIN users ON users.id = agents.user_id")         // LEFT JOIN for agent's user

		// Use ILIKE for case-insensitive search in PostgreSQL
		// Adjust fields as needed
		searchPattern := "%" + searchTerm + "%"
		query = query.Where(
			s.DB.Where("properties.owner_name ILIKE ?", searchPattern).
				Or("properties.description ILIKE ?", searchPattern).
				Or("properties.property_design ILIKE ?", searchPattern).
				Or("locations.district ILIKE ?", searchPattern).
				Or("locations.area ILIKE ?", searchPattern).
				Or("locations.sub_area ILIKE ?", searchPattern).
				Or("users.name ILIKE ?", searchPattern), // Search by agent name
		)
	}

	// Count total matching items
	countQuery := query // Create a separate query for counting before applying limits/offsets/preloads
	err := countQuery.Count(&totalItems).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Apply pagination, ordering, and preloads to the original query
	err = query.Scopes(utils.PaginateScope(pag.Page, pag.PageSize)).
		Order("properties.created_at DESC"). // Order by property creation date
		Preload("Location").
		Preload("Agent.User").
		Preload("CoverPhoto").
		Find(&properties).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve search results: %w", err)
	}

	return properties, totalItems, nil
}

// applyPropertyFilters builds the WHERE clauses based on filter criteria.
func applyPropertyFilters(query *gorm.DB, filter schema.PropertyFilter) *gorm.DB {
	if filter.OwnerName != nil && *filter.OwnerName != "" {
		query = query.Where("properties.owner_name ILIKE ?", "%"+*filter.OwnerName+"%")
	}
	if filter.PropertyType != nil && *filter.PropertyType != "" {
		query = query.Where("properties.property_type = ?", *filter.PropertyType)
	}
	if filter.ConstructionStage != nil && *filter.ConstructionStage != "" {
		query = query.Where("properties.construction_stage = ?", *filter.ConstructionStage)
	}
	if filter.ListingType != nil && *filter.ListingType != "" {
		query = query.Where("properties.listing_type = ?", *filter.ListingType)
	}
	if filter.MinPrice != nil {
		query = query.Where("properties.price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("properties.price <= ?", *filter.MaxPrice)
	}
	if filter.AgentID != nil {
		query = query.Where("properties.agent_id = ?", *filter.AgentID)
	}

	// For location filters, we need to join the tables
	needsLocationJoin := (filter.District != nil && *filter.District != "") || (filter.Area != nil && *filter.Area != "")
	if needsLocationJoin {
		query = query.Joins("JOIN locations ON locations.property_id = properties.id")
		if filter.District != nil && *filter.District != "" {
			query = query.Where("locations.district ILIKE ?", "%"+*filter.District+"%")
		}
		if filter.Area != nil && *filter.Area != "" {
			query = query.Where("locations.area ILIKE ?", "%"+*filter.Area+"%")
		}
	}

	return query
}

// Helper function to map model to response DTO
func MapPropertyToResponse(p *models.Property) schema.PropertyResponse {
	// Basic mapping, can be enhanced
	resp := schema.PropertyResponse{
		ID:                           p.ID,
		ValuerID:                     p.ValuerID,
		PropertyNumber:               p.PropertyNumber,
		ParentValuation:              p.ParentValuation,
		ProjectID:                    p.ProjectID,
		OwnerName:                    p.OwnerName,
		PropertyType:                 p.PropertyType,
		PropertyDesign:               p.PropertyDesign,
		ConstructionStage:            p.ConstructionStage,
		YearBuilt:                    p.YearBuilt,
		Age:                          p.Age,
		Eul:                          p.Eul,
		Rel:                          p.Rel,
		Measurements:                 p.Measurements,
		NoRooms:                      p.NoRooms,
		NoOfBathrooms:                p.NoOfBathrooms,
		Occupancy:                    p.Occupancy,
		Attributes:                   p.Attributes,
		TitleDeedsAvailable:          p.TitleDeedsAvailable,
		CertificateOfSearchAvailable: p.CertificateOfSearchAvailable,
		EncumbrancesAvailable:        p.EncumbrancesAvailable,
		Defects:                      p.Defects,
		Description:                  p.Description,
		MasterBedroomEnsuite:         p.MasterBedroomEnsuite,
		BuildingSize:                 p.BuildingSize,
		BuildingSizeUnit:             p.BuildingSizeUnit,
		LandSize:                     p.LandSize,
		LandSizeUnit:                 p.LandSizeUnit,
		EntryType:                    p.EntryType,
		Price:                        p.Price,
		ListingType:                  p.ListingType,
		CreatedBy:                    p.CreatedBy,
		CreatedAt:                    p.CreatedAt,
		UpdatedAt:                    p.UpdatedAt,
		IsApproved:                   p.IsApproved,
		IsSubmitted:                  p.IsSubmitted,
		IsSaleCompleted:              p.IsSaleCompleted,
		HasAcceptedOffer:             p.HasAcceptedOffer,
		IsReferred:                   p.IsReferred,
		ApprovedAt:                   p.ApprovedAt,
		Visibility:                   p.Visibility,
		Views:                        p.Views,
		// Embed associated data if it was preloaded
		Location: &p.Location, // Embed directly if not null
	}

	if p.Agent.ID != 0 { // Check if Agent was preloaded
		resp.Agent = &schema.AgentResponse{
			ID:        p.Agent.ID,
			Phone1:    p.Agent.Phone1,
			Headline1: p.Agent.Headline1,
			Headline2: p.Agent.Headline2,
			Linkedin:  p.Agent.Linkedin,
			Address:   p.Agent.Address,
		}
		if p.Agent.User.ID != 0 { // Check if User within Agent was preloaded
			resp.Agent.Name = p.Agent.User.Name // Get name from User
			resp.Agent.User = &schema.UserResponse{
				ID:              p.Agent.User.ID,
				Name:            p.Agent.User.Name,
				Email:           p.Agent.User.Email,
				Phone:           p.Agent.User.Phone,
				Role:            p.Agent.User.Role,
				ProfileImageURL: p.Agent.User.ProfileImageStorageURL, // Use the storage URL
			}
		}
	}

	if p.CoverPhoto.ID != 0 { // Check if CoverPhoto was preloaded
		resp.CoverPhoto = &schema.CoverPhotoResponse{
			ID:          p.CoverPhoto.ID,
			Url:         p.CoverPhoto.Url, // Populate URL from model
			Description: p.CoverPhoto.Description,
		}
	}
	return resp
}

// Helper to map multiple properties
func MapPropertiesToResponse(properties []models.Property) []schema.PropertyResponse {
	responses := make([]schema.PropertyResponse, len(properties))
	for i, p := range properties {
		responses[i] = MapPropertyToResponse(&p)
	}
	return responses
}
