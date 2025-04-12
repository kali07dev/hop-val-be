package models

import (
	"time"

	"gorm.io/datatypes"
)

// Struct to help unmarshal the JSON string in 'attributes'
type Attribute struct {
	Name string `json:"name"`
}
type AttributeList struct {
	Attributes []Attribute `json:"attributes"`
}

type Property struct {
	ID                            uint           `gorm:"primaryKey" json:"id"`
	ValuerID                      *uint          `json:"valuer_id"`        // Nullable uint
	PropertyNumber                *string        `json:"property_number"`  // Nullable string
	ParentValuation               *uint          `json:"parent_valuation"` // Nullable uint
	ProjectID                     *uint          `json:"project_id"`       // Nullable uint
	OwnerName                     string         `json:"owner_name"`
	PropertyType                  *string        `json:"property_type"`
	PropertyDesign                string         `json:"property_design"`
	ConstructionStage             string         `json:"construction_stage"`
	YearBuilt                     *string        `json:"year_built"` 
	Age                           *int           `json:"age"`        // Nullable int
	Eul                           *int           `json:"eul"`        // Estimated Useful Life
	Rel                           *int           `json:"rel"`        // Remaining Economic Life
	Measurements                  string         `json:"measurements"`
	NoRooms                       *int           `json:"no_rooms"`
	NoOfBathrooms                 *int           `json:"no_of_bathrooms"`
	Occupancy                     *string        `json:"occupancy"`

	Attributes                    datatypes.JSON `gorm:"type:jsonb" json:"attributes"` // Use jsonb for PSQL efficiency
	TitleDeedsAvailable           string         `json:"title_deeds_available"` // Consider bool or enum
	CertificateOfSearchAvailable  *string        `json:"certificate_of_search_available"`
	EncumbrancesAvailable         *string        `json:"encumbrances_available"`
	Defects                       *string        `json:"defects"`
	Description                   string         `gorm:"type:text" json:"description"`
	MasterBedroomEnsuite          string         `json:"master_bedroom_ensuite"` // Consider bool or enum
	BuildingSize                  *float64       `json:"building_size"`          // Use float for sizes
	BuildingSizeUnit              string         `json:"bulding_size_unit"`
	LandSize                      *float64       `json:"land_size"`
	LandSizeUnit                  string         `json:"land_size_unit"`
	EntryType                     string         `json:"entry_type"`
	Price                         *float64       `json:"price"` // Use float for price/currency
	ListingType                   string         `json:"listing_type"`
	CreatedBy                     *uint          `json:"created_by"` // Link to User ID? Assuming nullable uint
	CreatedAt                     time.Time      `json:"created_at"`
	UpdatedAt                     time.Time      `json:"updated_at"`
	IsApproved                    string         `json:"is_approved"` // Consider bool or enum
	IsSubmitted                   string         `json:"is_submitted"`
	IsSaleCompleted               string         `json:"is_sale_completed"`
	HasAcceptedOffer              string         `json:"has_accepted_offer"`
	IsReferred                    string         `json:"is_referred"`
	ApprovedAt                    *time.Time     `json:"approved_at"` // Nullable timestamp
	Visibility                    string         `json:"visibility"`
	Views                         int            `json:"views"`

	Location Location `gorm:"foreignKey:PropertyID"`

	AgentID *uint `json:"-"`
	Agent   Agent `gorm:"foreignKey:AgentID"`


	CoverPhoto CoverPhoto `gorm:"foreignKey:PropertyID"` 


}