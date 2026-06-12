// pkg/responses/responses.go — add paginated response
package responses

type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type PaginatedResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    T      `json:"data,omitempty"`
	Meta    Meta   `json:"meta"`
}

type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func Paginated[T any](message string, data T, total, limit, offset int) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    Meta{Total: total, Limit: limit, Offset: offset},
	}
}

func Success[T any](message string, data T) *Response[T] {
	return &Response[T]{Success: true, Message: message, Data: data}
}

func Created[T any](message string, data T) *Response[T] {
	return &Response[T]{Success: true, Message: message, Data: data}
}

func BadRequest(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "bad_request"}
}

func Unauthorized(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "unauthorized"}
}

func Forbidden(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "forbidden"}
}

func NotFound(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "not_found"}
}

func Conflict(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "conflict"}
}

func Validation(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "validation_error"}
}

func InternalServerError(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "internal_server_error"}
}

func TooManyRequests(message string) *Response[any] {
	return &Response[any]{Success: false, Message: message, Error: "too_many_requests"}
}
