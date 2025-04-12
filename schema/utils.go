package schema

// CustomError represents a structured API error response.
type CustomError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"` // Optional additional details
}

// Error implements error.
func (c *CustomError) Error() string {
	panic("unimplemented")
}

// PaginationRequest holds pagination parameters from the request query.
type PaginationRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"pageSize"`
}

// PaginatedResponse wraps list responses with pagination metadata.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalItems int64       `json:"totalItems"`
	TotalPages int         `json:"totalPages"`
}
