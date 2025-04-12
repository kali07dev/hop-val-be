package schema

// --- Structs mirroring the EXACT structure of the external API response ---

type ExternalUser struct {
	ID                     uint    `json:"id"`
	Name                   string  `json:"name"`
	Email                  string  `json:"email"`
	Status                 string  `json:"status"`
	EmailVerifiedAt        *string `json:"email_verified_at"` // API might send null or string date
	Phone                  string  `json:"phone"`
	PasswordLastUpdatedAt  *string `json:"password_last_updated_at"`
	FinancialInstitutionID *uint   `json:"financial_institution_id"`
	DeletedAt              *string `json:"deleted_at"`
	CreatedAt              string  `json:"created_at"` // API sends string date
	UpdatedAt              string  `json:"updated_at"` // API sends string date
	Role                   string  `json:"role"`
	SignatureStorageURL    *string `json:"signature_storage_url"`
	ProfileImageStorageURL *string `json:"profile_image_storage_url"`
	// Ignored calculated/formatted fields from API not needed
}

type ExternalAgent struct {
	ID                uint          `json:"id"`
	UserID            string        `json:"user_id"` // API sends user_id as string! Need conversion.
	Phone1            string        `json:"phone_1"`
	Phone2            *string       `json:"phone_2"`
	Headline1         *string       `json:"headline1"`
	Headline2         *string       `json:"headline2"`
	About             *string       `json:"about"`
	IsAgreementSigned string        `json:"is_agreement_signed"`
	AgentType         string        `json:"agent_type"`
	BankName          *string       `json:"bank_name"`
	AccountName       *string       `json:"account_name"`
	AccountNumber     *string       `json:"account_number"`
	AccountType       *string       `json:"account_type"`
	AccountBranch     *string       `json:"account_branch"`
	Phone             string        `json:"phone"` // Duplicated?
	Email             string        `json:"email"` // Duplicated?
	Linkedin          *string       `json:"linkedin"`
	Address           *string       `json:"address"`
	CreatedAt         string        `json:"created_at"` // API sends string date
	UpdatedAt         string        `json:"updated_at"` // API sends string date
	CoverageArea      *string       `json:"coverage_area"`
	Name              string        `json:"name"` // Name provided directly
	User              *ExternalUser `json:"user"` // Nested user object
}

type ExternalLocation struct {
	ID            uint    `json:"id"`
	PropertyID    uint    `json:"property_id"`
	Region        string  `json:"region"`
	District      string  `json:"district"`
	Area          string  `json:"area"`
	Postcode      *string `json:"postcode"`
	SubArea       *string `json:"sub_area"`
	GoogleMapLink *string `json:"google_map_link"`
	Latitude      *string `json:"latitude"`
	Longitude     *string `json:"longitude"`
	ZoneCategory  string  `json:"zone_category"`
	Zoning        string  `json:"zoning"`
	CreatedAt     string  `json:"created_at"` // API sends string date
	UpdatedAt     string  `json:"updated_at"` // API sends string date
}

type ExternalCoverPhoto struct {
	ID          uint   `json:"id"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

type ExternalProperty struct {
	ID                           uint                `json:"id"` // Use this as the primary key check
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
	Attributes                   string              `json:"attributes"` // Comes as JSON string, needs conversion
	TitleDeedsAvailable          string              `json:"title_deeds_available"`
	CertificateOfSearchAvailable *string             `json:"certificate_of_search_available"`
	EncumbrancesAvailable        *string             `json:"encumbrances_available"`
	Defects                      *string             `json:"defects"`
	Description                  string              `json:"description"`
	MasterBedroomEnsuite         string              `json:"master_bedroom_ensuite"`
	BuildingSize                 *float64            `json:"building_size"`     // API seems to send int, json unmarshal handles float conversion
	BuildingSizeUnit             string              `json:"bulding_size_unit"` // Note typo in API "bulding"
	LandSize                     *float64            `json:"land_size"`         // API seems to send int
	LandSizeUnit                 string              `json:"land_size_unit"`
	EntryType                    string              `json:"entry_type"`
	Price                        *float64            `json:"price"` // API sends int
	ListingType                  string              `json:"listing_type"`
	CreatedBy                    *uint               `json:"created_by"`
	CreatedAt                    string              `json:"created_at"` // API sends string date
	UpdatedAt                    string              `json:"updated_at"` // API sends string date
	IsApproved                   string              `json:"is_approved"`
	IsSubmitted                  string              `json:"is_submitted"`
	IsSaleCompleted              string              `json:"is_sale_completed"`
	HasAcceptedOffer             string              `json:"has_accepted_offer"`
	IsReferred                   string              `json:"is_referred"`
	ApprovedAt                   *string             `json:"approved_at"` // API sends string date
	Visibility                   string              `json:"visibility"`
	Views                        int                 `json:"views"`
	Location                     *ExternalLocation   `json:"location"`
	Agent                        *ExternalAgent      `json:"agent"`
	CoverPhoto                   *ExternalCoverPhoto `json:"cover_photo"`
	OpenHouses                   []interface{}       `json:"open_houses"` // Keep as interface{} if structure varies or is unused
}

// Represents the overall structure of the API response
type PropertiesAPIResponse struct {
	Data []ExternalProperty `json:"data"`
	Meta struct {
		CurrentPage int     `json:"current_page"`
		LastPage    int     `json:"last_page"`
		PerPage     int     `json:"per_page"`
		Total       int     `json:"total"`
		NextPageURL *string `json:"next_page_url"`
		PrevPageURL *string `json:"prev_page_url"`
	} `json:"meta"`
}

// --- Sync Operation Response ---
type SyncResult struct {
	Status            string   `json:"status"`
	TotalPagesFetched int      `json:"totalPagesFetched"`
	TotalProperties   int      `json:"totalProperties"` // Total from API meta
	FetchedCount      int      `json:"fetchedCount"`    // Actual items processed from API pages
	SyncedCount       int      `json:"syncedCount"`     // Successfully inserted
	SkippedCount      int      `json:"skippedCount"`    // Duplicates skipped
	ErrorCount        int      `json:"errorCount"`
	Errors            []string `json:"errors,omitempty"` // List of specific errors encountered
}
