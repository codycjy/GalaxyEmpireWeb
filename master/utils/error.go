package utils

type AppError interface {
	StatusCode() int
	Msg() string
	Error() string
	ErrorType() string
}

type ServiceError struct {
	code    int
	message string
	err     error
}

func NewServiceError(code int, message string, err error) *ServiceError {
	return &ServiceError{
		code:    code,
		message: message,
		err:     err,
	}
}

func (s *ServiceError) StatusCode() int {
	return s.code
}

func (s *ServiceError) Msg() string {
	return s.message
}

func (s *ServiceError) Error() string {
	if s.err != nil {
		return s.err.Error()
	}
	return ""
}

func (s *ServiceError) ErrorType() string {
	return "ServiceError"
}

type ApiError struct {
	Code    int
	Message string
	err     error
}

func NewApiError(code int, message string, err error) *ApiError {
	return &ApiError{
		Code:    code,
		Message: message,
		err:     err,
	}
}

func (s *ApiError) StatusCode() int {
	return s.Code
}

func (s *ApiError) Msg() string {
	return s.Message
}

func (s *ApiError) Error() string {
	return s.err.Error()
}

func (s *ApiError) ErrorType() string {
	return "ApiError"
}
