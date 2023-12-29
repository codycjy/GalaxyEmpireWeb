package api

// ErrorResponse godoc
// All the error response
type ErrorResponse struct {
	Succeed bool   `json:"succeed"`
	Error   string `json:"error"`
	Message string `json:"message"`
	TraceID string `json:"traceID"`
}
