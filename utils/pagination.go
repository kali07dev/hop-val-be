package utils

import (
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/hopekali04/valuations/schema"
	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// GetPaginationParams extracts pagination info from Fiber context.
func GetPaginationParams(c *fiber.Ctx) schema.PaginationRequest {
	page := c.QueryInt("page", DefaultPage)
	pageSize := c.QueryInt("pageSize", DefaultPageSize)

	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return schema.PaginationRequest{
		Page:     page,
		PageSize: pageSize,
	}
}

// PaginateScope returns a GORM scope function for pagination.
func PaginateScope(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// CreatePaginatedResponse creates the standard paginated response structure.
func CreatePaginatedResponse(data interface{}, totalItems int64, page, pageSize int) schema.PaginatedResponse {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}

	return schema.PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
