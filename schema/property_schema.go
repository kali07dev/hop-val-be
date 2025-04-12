package schema

import (
	"time"

	"github.com/hopekali04/valuations/models"
	"gorm.io/datatypes"
)

// Simplified User representation for embedding in responses
type UserResponse struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	Phone           string  `json:"phone"`
	Role            string  `json:"role"`                    // Consider adding formatted role here
	ProfileImageURL *string `json:"profile_image,omitempty"` // URL field
}

// Simplified Agent representation for embedding in responses
type AgentResponse struct {
	ID        uint          `json:"id"`
	Phone1    string        `json:"phone_1"`
	Headline1 *string       `json:"headline1"`
	Headline2 *string       `json:"headline2"`
	Linkedin  *string       `json:"linkedin"`
	Address   *string       `json:"address"`
	Name      string        `json:"name"`           // Name from User
	User      *UserResponse `json:"user,omitempty"` // Embed simplified user
}

// Simplified CoverPhoto representation
type CoverPhotoResponse struct {
	ID          uint   `json:"id"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

// PropertyResponse defines the structure for returning property details.
// It mirrors the desired output JSON structure more closely.
type PropertyResponse struct {
	ID                           uint                `json:"id"`
	ValuerID                     *uint               `json:"valuer_id"`
	PropertyNumber               *string             `json:"property_number"`
	ParentValuation              *uint               `json:"parent_valuation"`
	ProjectID                    *uint               `json:"project_id"`
	OwnerName                    string              `json:"owner_name"`
	PropertyType                 *string             `json:"property_type"`
	PropertyDesign               string              `json:"property_design"`
	ConstructionStage            string              `json:"construction_stage"`
	YearBuilt                    *string             `json:"year_built"`
	Age                          *int                `json:"age"`
	Eul                          *int                `json:"eul"`
	Rel                          *int                `json:"rel"`
	Measurements                 string              `json:"measurements"`
	NoRooms                      *int                `json:"no_rooms"`
	NoOfBathrooms                *int                `json:"no_of_bathrooms"`
	Occupancy                    *string             `json:"occupancy"`
	Attributes                   datatypes.JSON      `json:"attributes"` // Keep as JSON or unmarshal to struct
	TitleDeedsAvailable          string              `json:"title_deeds_available"`
	CertificateOfSearchAvailable *string             `json:"certificate_of_search_available"`
	EncumbrancesAvailable        *string             `json:"encumbrances_available"`
	Defects                      *string             `json:"defects"`
	Description                  string              `json:"description"`
	MasterBedroomEnsuite         string              `json:"master_bedroom_ensuite"`
	BuildingSize                 *float64            `json:"building_size"`
	BuildingSizeUnit             string              `json:"bulding_size_unit"`
	LandSize                     *float64            `json:"land_size"`
	LandSizeUnit                 string              `json:"land_size_unit"`
	EntryType                    string              `json:"entry_type"`
	Price                        *float64            `json:"price"`
	ListingType                  string              `json:"listing_type"`
	CreatedBy                    *uint               `json:"created_by"`
	CreatedAt                    time.Time           `json:"created_at"`
	UpdatedAt                    time.Time           `json:"updated_at"`
	IsApproved                   string              `json:"is_approved"`
	IsSubmitted                  string              `json:"is_submitted"`
	IsSaleCompleted              string              `json:"is_sale_completed"`
	HasAcceptedOffer             string              `json:"has_accepted_offer"`
	IsReferred                   string              `json:"is_referred"`
	ApprovedAt                   *time.Time          `json:"approved_at"`
	Visibility                   string              `json:"visibility"`
	Views                        int                 `json:"views"`
	Location                     *models.Location    `json:"location,omitempty"`    // Embed full location
	Agent                        *AgentResponse      `json:"agent,omitempty"`       // Embed simplified agent
	CoverPhoto                   *CoverPhotoResponse `json:"cover_photo,omitempty"` // Embed simplified cover photo
	// OpenHouses                 []models.OpenHouse `json:"open_houses"` // Embed open houses if needed
}


type CreatePropertyRequest struct {
	ID                           uint               `json:"id" validate:"required"` // Require ID for duplication check
	ValuerID                     *uint              `json:"valuer_id"`
	PropertyNumber               *string            `json:"property_number"`
	ParentValuation              *uint              `json:"parent_valuation"`
	ProjectID                    *uint              `json:"project_id"`
	OwnerName                    string             `json:"owner_name" validate:"required"`
	PropertyType                 *string            `json:"property_type"`
	PropertyDesign               string             `json:"property_design"`
	ConstructionStage            string             `json:"construction_stage"`
	YearBuilt                    *string            `json:"year_built"`
	Age                          *int               `json:"age"`
	Eul                          *int               `json:"eul"`
	Rel                          *int               `json:"rel"`
	Measurements                 string             `json:"measurements"`
	NoRooms                      *int               `json:"no_rooms"`
	NoOfBathrooms                *int               `json:"no_of_bathrooms"`
	Occupancy                    *string            `json:"occupancy"`
	Attributes                   datatypes.JSON     `json:"attributes"`
	TitleDeedsAvailable          string             `json:"title_deeds_available"`
	CertificateOfSearchAvailable *string            `json:"certificate_of_search_available"`
	EncumbrancesAvailable        *string            `json:"encumbrances_available"`
	Defects                      *string            `json:"defects"`
	Description                  string             `json:"description"`
	MasterBedroomEnsuite         string             `json:"master_bedroom_ensuite"`
	BuildingSize                 *float64           `json:"building_size"`
	BuildingSizeUnit             string             `json:"bulding_size_unit"`
	LandSize                     *float64           `json:"land_size"`
	LandSizeUnit                 string             `json:"land_size_unit"`
	EntryType                    string             `json:"entry_type"`
	Price                        *float64           `json:"price"`
	ListingType                  string             `json:"listing_type"`
	CreatedBy                    *uint              `json:"created_by"`
	IsApproved                   string             `json:"is_approved"`
	IsSubmitted                  string             `json:"is_submitted"`
	IsSaleCompleted              string             `json:"is_sale_completed"`
	HasAcceptedOffer             string             `json:"has_accepted_offer"`
	IsReferred                   string             `json:"is_referred"`
	ApprovedAt                   *time.Time         `json:"approved_at"`
	Visibility                   string             `json:"visibility"`
	Views                        int                `json:"views"`
	Location                     *models.Location   `json:"location,omitempty"`    // For creating/updating nested location
	AgentID                      *uint              `json:"agent_id,omitempty"`    // Link to existing agent
	CoverPhoto                   *models.CoverPhoto `json:"cover_photo,omitempty"` // For creating/updating nested cover photo
}

// BulkCreatePropertyRequest holds an array of properties to create.
type BulkCreatePropertyRequest struct {
	Properties []CreatePropertyRequest `json:"properties" validate:"required,dive"` // dive validates each element
}

// BulkCreatePropertyResponse defines the response for bulk creation.
type BulkCreatePropertyResponse struct {
	SuccessCount int                `json:"successCount"`
	FailCount    int                `json:"failCount"`
	Errors       []BulkErrorDetail  `json:"errors,omitempty"`
	Successful   []PropertyResponse `json:"successful,omitempty"` // Optional: return successfully created items
}

// BulkErrorDetail provides info about a failed item in a bulk operation.
type BulkErrorDetail struct {
	PropertyID uint   `json:"propertyId"` // ID of the property that failed
	Error      string `json:"error"`
}

// PropertyFilter defines available query parameters for filtering properties.
type PropertyFilter struct {
	OwnerName         *string  `query:"ownerName"`
	PropertyType      *string  `query:"propertyType"`
	ConstructionStage *string  `query:"constructionStage"`
	ListingType       *string  `query:"listingType"`
	MinPrice          *float64 `query:"minPrice"`
	MaxPrice          *float64 `query:"maxPrice"`
	District          *string  `query:"district"` // Filter by location district
	Area              *string  `query:"area"`     // Filter by location area
	AgentID           *uint    `query:"agentId"`  // Filter by agent
}
