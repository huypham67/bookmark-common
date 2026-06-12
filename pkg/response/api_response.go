package response

// Pagination represents pagination metadata for list responses.
type Pagination struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
	Total int64 `json:"total"`
}

// SuccessResponse is a generic response wrapper for successful API operations.
// T can be any type: a data struct, a pointer, or even basic types like string or int.
// Use this type directly or via type alias in DTOs for consistency.
//
// Example usage in DTOs:
//
//	type UserResponse = SuccessResponse[*UserData]
//	type LoginResponse = SuccessResponse[string]
//	type BookmarkListResponse = SuccessResponse[[]BookmarkData]
type SuccessResponse[T any] struct {
	Data       T           `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// MessageResponse represents a simple response containing only a message.
type MessageResponse struct {
	Message string `json:"message"`
}

// Success creates a successful response with data.
// For pointer types, call with SuccessResponse[*T]{Data: ptr, Message: msg}.
// For value types, call with Success(value, msg) or SuccessResponse[T]{Data: value, Message: msg}.
func Success[T any](data T, message string) SuccessResponse[T] {
	return SuccessResponse[T]{
		Data:    data,
		Message: message,
	}
}

// Message creates a successful response without data.
func Message(message string) MessageResponse {
	return MessageResponse{
		Message: message,
	}
}

// Paginated creates a paginated response with pagination metadata.
// For pointer types in slice, use Paginated with a slice of pointers.
func Paginated[T any](
	data T,
	page int64,
	limit int64,
	total int64,
	message string,
) SuccessResponse[T] {
	return SuccessResponse[T]{
		Data:    data,
		Message: message,
		Pagination: &Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}
}
