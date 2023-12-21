package api

type ErrorResponse struct {
	Succeed bool   `json:"succeed"`
	Error   string `json:"error"`
	Message string `json:"message"`
}
