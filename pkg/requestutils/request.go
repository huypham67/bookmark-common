package requestutils

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Bind extracts and binds data from multiple sources (URI, JSON, Query, Header),
// then validates using `validate` struct tags.
//
// Binding order:
// 1. JSON body (for non-GET requests)
// 2. URI parameters
// 3. Query parameters
// 4. Headers
// 5. Struct validation using `validate` tags (AFTER all sources are bound)
//
// Important:
// - Use binding:"..." tags ONLY for type conversion during ShouldBind* calls
// - Use validate:"..." tags for business logic validation that runs AFTER binding
// - Do NOT use binding:"required" for fields from different sources
//
// Example:
//
//	type UpdateRequest struct {
//	    ID   string `uri:"id" validate:"required"`                   // ID from URI, must exist
//	    Name string `json:"name" validate:"omitempty,min=2,max=100"` // Name from JSON, optional
//	}
func Bind[T any](c *gin.Context) (*T, error) {
	req := new(T)

	// Step 1: Bind JSON body (for non-GET requests)
	// Only extracts and type-converts, does NOT validate binding tags
	if c.Request.Method != http.MethodGet {
		if err := c.ShouldBindJSON(req); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	// Step 2: Bind URI parameters
	if err := c.ShouldBindUri(req); err != nil {
		return nil, err
	}

	// Step 3: Bind query parameters
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}

	// Step 4: Bind headers
	if err := c.ShouldBindHeader(req); err != nil {
		return nil, err
	}

	// Step 5: Validate using `validate` tags (AFTER ALL sources are bound)
	v := validator.New()
	if err := v.Struct(req); err != nil {
		return nil, err
	}

	return req, nil
}
