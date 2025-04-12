// api/property_handler.go
package api

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/hopekali04/valuations/schema"
	"github.com/hopekali04/valuations/services"
	"github.com/hopekali04/valuations/utils"
)

type PropertyHandler struct {
	Service   *services.PropertyService
	Validator *validator.Validate // Add validator instance
}

func NewPropertyHandler(service *services.PropertyService) *PropertyHandler {
	return &PropertyHandler{
		Service:   service,
		Validator: validator.New(), // Initialize validator
	}
}

// CreateProperty handles POST /properties
func (h *PropertyHandler) CreateProperty(c *fiber.Ctx) error {
	var req schema.CreatePropertyRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return utils.HandleError(c, utils.NewBadRequestError("Invalid request body: "+err.Error()))
	}

	// Validate request struct
	if err := h.Validator.Struct(req); err != nil {
		return utils.HandleError(c, err.(validator.ValidationErrors)) // Let HandleError format validation errors
	}

	// Call the service
	createdProperty, err := h.Service.CreateProperty(&req)
	if err != nil {
		return utils.HandleError(c, err) // Use centralized error handler
	}

	// Map model to response DTO
	response := services.MapPropertyToResponse(createdProperty)

	return c.Status(fiber.StatusCreated).JSON(response)
}

// CreateMultipleProperties handles POST /properties/bulk
func (h *PropertyHandler) CreateMultipleProperties(c *fiber.Ctx) error {
	var bulkReq schema.BulkCreatePropertyRequest

	if err := c.BodyParser(&bulkReq); err != nil {
		return utils.HandleError(c, utils.NewBadRequestError("Invalid request body: "+err.Error()))
	}

	// Validate the wrapper struct and the nested slice
	if err := h.Validator.Struct(bulkReq); err != nil {
		return utils.HandleError(c, err.(validator.ValidationErrors))
	}

	// Call the service
	successful, errorsReport := h.Service.CreateMultipleProperties(bulkReq.Properties)

	// Prepare response
	response := schema.BulkCreatePropertyResponse{
		SuccessCount: len(successful),
		FailCount:    len(errorsReport),
		Errors:       errorsReport,
		Successful:   services.MapPropertiesToResponse(successful), // Map successful ones to response DTOs
	}

	// Determine status code: Accepted if there were partial failures, Created if all succeeded
	statusCode := fiber.StatusCreated
	if len(errorsReport) > 0 {
		statusCode = fiber.StatusAccepted // Indicate partial success/failure
	}

	return c.Status(statusCode).JSON(response)
}

// GetPropertyByID handles GET /properties/:id
func (h *PropertyHandler) GetPropertyByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32) // Use uint32 or uint64 based on your ID type
	if err != nil {
		return utils.HandleError(c, utils.NewBadRequestError("Invalid property ID format"))
	}

	property, err := h.Service.GetPropertyByID(uint(id))
	if err != nil {
		return utils.HandleError(c, err)
	}

	response := services.MapPropertyToResponse(property)

	return c.JSON(response)
}

// GetAllProperties handles GET /properties
func (h *PropertyHandler) GetAllProperties(c *fiber.Ctx) error {
	// Get pagination params
	paginationParams := utils.GetPaginationParams(c)

	// Parse filter params
	var filterParams schema.PropertyFilter
	if err := c.QueryParser(&filterParams); err != nil {
		return utils.HandleError(c, utils.NewBadRequestError("Invalid filter parameters: "+err.Error()))
	}

	// Call service
	properties, totalItems, err := h.Service.GetAllProperties(paginationParams, filterParams)
	if err != nil {
		return utils.HandleError(c, err)
	}

	// Map models to response DTOs
	propertyResponses := services.MapPropertiesToResponse(properties)

	// Create paginated response
	paginatedResponse := utils.CreatePaginatedResponse(propertyResponses, totalItems, paginationParams.Page, paginationParams.PageSize)

	return c.JSON(paginatedResponse)
}

// SearchProperties handles GET /properties/search
func (h *PropertyHandler) SearchProperties(c *fiber.Ctx) error {
	// Get search term
	searchTerm := c.Query("q", "") // Default to empty string if 'q' param is missing

	// Get pagination params
	paginationParams := utils.GetPaginationParams(c)

	// Call service
	properties, totalItems, err := h.Service.SearchProperties(searchTerm, paginationParams)
	if err != nil {
		return utils.HandleError(c, err)
	}

	// Map models to response DTOs
	propertyResponses := services.MapPropertiesToResponse(properties)

	// Create paginated response
	paginatedResponse := utils.CreatePaginatedResponse(propertyResponses, totalItems, paginationParams.Page, paginationParams.PageSize)

	return c.JSON(paginatedResponse)
}
