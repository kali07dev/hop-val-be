package services

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hopekali04/valuations/models"
	"github.com/hopekali04/valuations/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Helper function to setup mock DB and service
func setupServiceWithMock(t *testing.T) (PropertyService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Use postgres dialect; Disable Prepared Stmt for sqlmock compatibility
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true, // Avoids prepared statements that interfere with sqlmock
	}), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info), // Uncomment for debugging SQL
	})
	require.NoError(t, err)

	service := NewPropertyService(gormDB)
	return *service, mock
}

// Helper function to create mock property rows for sqlmock
func createMockPropertyRows(properties ...models.Property) *sqlmock.Rows {
	// Define columns matching the fields in models.Property GORM selects
	// This needs to be accurate based on your model and GORM's behavior
	cols := []string{
		"id", "valuer_id", "property_number", "parent_valuation", "project_id",
		"owner_name", "property_type", "property_design", "construction_stage", "year_built",
		"age", "eul", "rel", "measurements", "no_rooms", "no_of_bathrooms", "occupancy",
		"attributes", "title_deeds_available", "certificate_of_search_available", "encumbrances_available",
		"defects", "description", "master_bedroom_ensuite", "building_size", "building_size_unit",
		"land_size", "land_size_unit", "entry_type", "price", "listing_type", "created_by",
		"created_at", "updated_at", "is_approved", "is_submitted", "is_sale_completed",
		"has_accepted_offer", "is_referred", "approved_at", "visibility", "views", "agent_id",
		// Add columns for joined tables if you mock preloads directly (complex)
	}
	rows := sqlmock.NewRows(cols)
	for _, p := range properties {
		// Add row data matching the order of columns
		rows.AddRow(
			p.ID, p.ValuerID, p.PropertyNumber, p.ParentValuation, p.ProjectID,
			p.OwnerName, p.PropertyType, p.PropertyDesign, p.ConstructionStage, p.YearBuilt,
			p.Age, p.Eul, p.Rel, p.Measurements, p.NoRooms, p.NoOfBathrooms, p.Occupancy,
			p.Attributes, p.TitleDeedsAvailable, p.CertificateOfSearchAvailable, p.EncumbrancesAvailable,
			p.Defects, p.Description, p.MasterBedroomEnsuite, p.BuildingSize, p.BuildingSizeUnit,
			p.LandSize, p.LandSizeUnit, p.EntryType, p.Price, p.ListingType, p.CreatedBy,
			p.CreatedAt, p.UpdatedAt, p.IsApproved, p.IsSubmitted, p.IsSaleCompleted,
			p.HasAcceptedOffer, p.IsReferred, p.ApprovedAt, p.Visibility, p.Views, p.AgentID,
		)
	}
	return rows
}

// Mock Preload Queries (Simplified - adjust based on actual GORM queries)
// Mocking Preloads precisely with sqlmock can be tricky. Often easier to test the main query filter logic.
func expectPreloads(mock sqlmock.Sqlmock, propertyIDs []uint) {
	if len(propertyIDs) == 0 {
		return
	}
	// Example: Mock Location Preload
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "locations" WHERE "locations"."property_id" IN ($1)`)).
		WithArgs(propertyIDs[0]). // Adjust Args based on how GORM batches
		WillReturnRows(sqlmock.NewRows([]string{"id", "property_id", "district"}).AddRow(1, propertyIDs[0], "Test District"))

	// Example: Mock Agent Preload
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."id" IN ($1)`)). // Assuming agent IDs are known or IN clause used
													WithArgs(sqlmock.AnyArg()). // Or use specific agent IDs if available
													WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name"}).AddRow(1, 1, "Test Agent"))

	// Example: Mock User Preload (nested within Agent)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" IN ($1)`)). // Assuming user IDs are known or IN clause used
												WithArgs(sqlmock.AnyArg()). // Or use specific user IDs
												WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).AddRow(1, "Test User", "test@user.com"))

	// Example: Mock CoverPhoto Preload
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cover_photos" WHERE "cover_photos"."property_id" IN ($1)`)).
		WithArgs(propertyIDs[0]). // Adjust Args based on how GORM batches
		WillReturnRows(sqlmock.NewRows([]string{"id", "property_id", "url"}).AddRow(1, propertyIDs[0], "http://example.com/photo.jpg"))
}

// --- Tests for GetAllProperties ---

func TestGetAllProperties_NoFilters(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 5}
	filter := schema.PropertyFilter{} // Empty filter

	// Mock Count Query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "properties"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2)) // Total 2 items

	// Mock Data Query (with pagination)
	mockProperties := []models.Property{
		{ID: 1, OwnerName: "Owner A", CreatedAt: time.Now()},
		{ID: 2, OwnerName: "Owner B", CreatedAt: time.Now()},
	}
	rows := createMockPropertyRows(mockProperties...)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "properties" ORDER BY created_at DESC LIMIT 5`)). // OFFSET 0 implied for page 1
														WillReturnRows(rows)

	// Mock Preloads (Optional but good practice if your handler relies on them)
	expectPreloads(mock, []uint{1, 2})

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, properties, 2)
	assert.Equal(t, uint(1), properties[0].ID)
	assert.Equal(t, uint(2), properties[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet()) // Verify all expectations were met
}

func TestGetAllProperties_FilterByOwnerName(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 10}
	ownerName := "Specific Owner"
	filter := schema.PropertyFilter{OwnerName: &ownerName}

	// Mock Count
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "properties" WHERE properties.owner_name ILIKE $1`)).
		WithArgs("%" + ownerName + "%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock Data Fetch
	mockProperties := []models.Property{{ID: 3, OwnerName: ownerName}}
	rows := createMockPropertyRows(mockProperties...)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "properties" WHERE properties.owner_name ILIKE $1 ORDER BY created_at DESC LIMIT 10`)).
		WithArgs("%" + ownerName + "%").
		WillReturnRows(rows)

	expectPreloads(mock, []uint{3})

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, properties, 1)
	assert.Equal(t, uint(3), properties[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllProperties_FilterByPropertyType(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 10}
	propType := "Commercial"
	filter := schema.PropertyFilter{PropertyType: &propType}

	// Mock Count
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "properties" WHERE properties.property_type = $1`)).
		WithArgs(propType).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock Data Fetch
	mockProperties := []models.Property{{ID: 4, PropertyType: &propType}}
	rows := createMockPropertyRows(mockProperties...)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "properties" WHERE properties.property_type = $1 ORDER BY created_at DESC LIMIT 10`)).
		WithArgs(propType).
		WillReturnRows(rows)

	expectPreloads(mock, []uint{4})

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, properties, 1)
	assert.Equal(t, uint(4), properties[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllProperties_FilterByMinMaxPrice(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 10}
	minPrice := 100000.0
	maxPrice := 500000.0
	filter := schema.PropertyFilter{MinPrice: &minPrice, MaxPrice: &maxPrice}

	// Mock Count
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "properties" WHERE properties.price >= $1 AND properties.price <= $2`)).
		WithArgs(minPrice, maxPrice).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock Data Fetch
	mockProperties := []models.Property{{ID: 5, Price: &minPrice}}
	rows := createMockPropertyRows(mockProperties...)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "properties" WHERE properties.price >= $1 AND properties.price <= $2 ORDER BY created_at DESC LIMIT 10`)).
		WithArgs(minPrice, maxPrice).
		WillReturnRows(rows)

	expectPreloads(mock, []uint{5})

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, properties, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllProperties_FilterByDistrict(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 10}
	district := "Downtown"
	filter := schema.PropertyFilter{District: &district}

	// Mock Count
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "properties" JOIN locations ON locations.property_id = properties.id WHERE locations.district ILIKE $1`)).
		WithArgs("%" + district + "%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock Data Fetch
	mockProperties := []models.Property{{ID: 6}} // Location data handled by preload mock
	rows := createMockPropertyRows(mockProperties...)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "properties" JOIN locations ON locations.property_id = properties.id WHERE locations.district ILIKE $1 ORDER BY created_at DESC LIMIT 10`)).
		WithArgs("%" + district + "%").
		WillReturnRows(rows)

	expectPreloads(mock, []uint{6}) // Expect preloads after main query

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, properties, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllProperties_FilterByAgentID(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 10}
	agentID := uint(15)
	filter := schema.PropertyFilter{AgentID: &agentID}

	// Mock Count
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "properties" WHERE properties.agent_id = $1`)).
		WithArgs(agentID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock Data Fetch
	mockProperties := []models.Property{{ID: 7, AgentID: &agentID}}
	rows := createMockPropertyRows(mockProperties...)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "properties" WHERE properties.agent_id = $1 ORDER BY created_at DESC LIMIT 10`)).
		WithArgs(agentID).
		WillReturnRows(rows)

	expectPreloads(mock, []uint{7})

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, properties, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllProperties_AllFiltersCombined(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 10}
	ownerName := "Combined Owner"
	propType := "Residential"
	minPrice := 50000.0
	district := "Suburb"
	agentID := uint(20)
	filter := schema.PropertyFilter{
		OwnerName:    &ownerName,
		PropertyType: &propType,
		MinPrice:     &minPrice,
		District:     &district,
		AgentID:      &agentID,
	}

	// Construct expected SQL part (order might vary slightly, use regex if needed)
	expectedCountSQL := `SELECT count(*) FROM "properties" JOIN locations ON locations.property_id = properties.id WHERE properties.owner_name ILIKE $1 AND properties.property_type = $2 AND properties.price >= $3 AND properties.agent_id = $4 AND locations.district ILIKE $5`
	expectedSelectSQL := `SELECT * FROM "properties" JOIN locations ON locations.property_id = properties.id WHERE properties.owner_name ILIKE $1 AND properties.property_type = $2 AND properties.price >= $3 AND properties.agent_id = $4 AND locations.district ILIKE $5 ORDER BY created_at DESC LIMIT 10`

	// Mock Count
	mock.ExpectQuery(regexp.QuoteMeta(expectedCountSQL)).
		WithArgs("%"+ownerName+"%", propType, minPrice, agentID, "%"+district+"%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock Data Fetch
	mockProperties := []models.Property{{ID: 8, OwnerName: ownerName, PropertyType: &propType, Price: &minPrice, AgentID: &agentID}}
	rows := createMockPropertyRows(mockProperties...)
	mock.ExpectQuery(regexp.QuoteMeta(expectedSelectSQL)).
		WithArgs("%"+ownerName+"%", propType, minPrice, agentID, "%"+district+"%").
		WillReturnRows(rows)

	expectPreloads(mock, []uint{8})

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, properties, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllProperties_NoResults(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 10}
	ownerName := "NonExistentOwner"
	filter := schema.PropertyFilter{OwnerName: &ownerName}

	// Mock Count returning 0
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "properties" WHERE properties.owner_name ILIKE $1`)).
		WithArgs("%" + ownerName + "%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock Data Fetch returning no rows
	rows := createMockPropertyRows() // Empty rows
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "properties" WHERE properties.owner_name ILIKE $1 ORDER BY created_at DESC LIMIT 10`)).
		WithArgs("%" + ownerName + "%").
		WillReturnRows(rows)

	// No preloads expected if no main rows found

	// --- Act ---
	properties, total, err := service.GetAllProperties(pagination, filter)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, properties, 0) // Expect empty slice
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Tests for SearchProperties ---

func TestSearchProperties_Success(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 5}
	searchTerm := "Zingwangwa"

	// Regex for JOINs and WHERE clauses in search
	expectedJoins := `JOIN locations ON locations.property_id = properties.id LEFT JOIN agents ON agents.id = properties.agent_id LEFT JOIN users ON users.id = agents.user_id`
	_ = `properties.owner_name ILIKE $1 OR properties.description ILIKE $2 OR properties.property_design ILIKE $3 OR locations.district ILIKE $4 OR locations.area ILIKE $5 OR locations.sub_area ILIKE $6 OR users.name ILIKE $7`
	searchArg := "%" + searchTerm + "%"

	// Mock Count Query
	// Using regex because the exact query string might be complex
	mock.ExpectQuery(`SELECT count\(\*\) FROM "properties" `+expectedJoins+` WHERE \(.*`+regexp.QuoteMeta(searchTerm)+`.*\)`). // Check for joins and WHERE ILIKE
																	WithArgs(searchArg, searchArg, searchArg, searchArg, searchArg, searchArg, searchArg). // 7 arguments for ILIKE
																	WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))                           // Found 1 match

	// Mock Data Query
	mockProperties := []models.Property{{ID: 154, OwnerName: "Some Match"}}
	rows := createMockPropertyRows(mockProperties...)
	// Using regex again for flexibility
	mock.ExpectQuery(`SELECT .* FROM "properties" `+expectedJoins+` WHERE \(.*`+regexp.QuoteMeta(searchTerm)+`.*\) ORDER BY properties.created_at DESC LIMIT 5`).
		WithArgs(searchArg, searchArg, searchArg, searchArg, searchArg, searchArg, searchArg).
		WillReturnRows(rows)

	expectPreloads(mock, []uint{154})

	// --- Act ---
	properties, total, err := service.SearchProperties(searchTerm, pagination)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, properties, 1)
	assert.Equal(t, uint(154), properties[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProperties_NoResults(t *testing.T) {
	service, mock := setupServiceWithMock(t)
	pagination := schema.PaginationRequest{Page: 1, PageSize: 5}
	searchTerm := "xyzNonExistent123"
	searchArg := "%" + searchTerm + "%"

	expectedJoins := `JOIN locations ON locations.property_id = properties.id LEFT JOIN agents ON agents.id = properties.agent_id LEFT JOIN users ON users.id = agents.user_id`

	// Mock Count Query returning 0
	mock.ExpectQuery(`SELECT count\(\*\) FROM "properties" `+expectedJoins+` WHERE \(.*`+regexp.QuoteMeta(searchTerm)+`.*\)`).
		WithArgs(searchArg, searchArg, searchArg, searchArg, searchArg, searchArg, searchArg).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Mock Data Query returning no rows
	rows := createMockPropertyRows() // Empty
	mock.ExpectQuery(`SELECT .* FROM "properties" `+expectedJoins+` WHERE \(.*`+regexp.QuoteMeta(searchTerm)+`.*\) ORDER BY properties.created_at DESC LIMIT 5`).
		WithArgs(searchArg, searchArg, searchArg, searchArg, searchArg, searchArg, searchArg).
		WillReturnRows(rows)

	// No preloads expected

	// --- Act ---
	properties, total, err := service.SearchProperties(searchTerm, pagination)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, properties, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
