package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PageParams holds pagination state extracted from query params.
type PageParams struct {
	Page     int
	PageSize int
}

// ParsePagination extracts page and page_size from the request query string
// with sensible defaults and bounds.
func ParsePagination(c *gin.Context) PageParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return PageParams{Page: page, PageSize: pageSize}
}

// Offset returns the GORM offset value.
func (p PageParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PaginatedResult wraps a data slice with pagination metadata.
func PaginatedResult(data interface{}, total int64, p PageParams) gin.H {
	return gin.H{
		"data":      data,
		"total":     total,
		"page":      p.Page,
		"page_size": p.PageSize,
	}
}
