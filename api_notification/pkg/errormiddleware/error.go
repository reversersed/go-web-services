package errormiddleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Code string

const (
	InternalErrorCode     Code = "IE-0001"
	NotFoundErrorCode     Code = "IE-0002"
	BadRequestErrorCode   Code = "IE-0003"
	ValidationErrorCode   Code = "IE-0004"
	UnauthorizedErrorCode Code = "IE-0005"
	NotUniqueErrorCode    Code = "IE-0006"
	ForbiddenErrorCode    Code = "IE-0007"
)

type Error struct {
	Message          []string `json:"messages,omitempty"`
	DeveloperMessage string   `json:"dev_message,omitempty"`
	Code             Code     `json:"code,omitempty"`
}

func NewError(message []string, code Code, dev_message string) *Error {
	return &Error{
		Code:             code,
		Message:          message,
		DeveloperMessage: dev_message,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error code: %s, Error: %s, Dev message: %s", e.Code, strings.Join(e.Message, ", "), e.DeveloperMessage)
}

func (e *Error) Marshall() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}
func sysError(message []string) *Error {
	return NewError(message, InternalErrorCode, "Something wrong happened while service executing")
}
func NotFoundError(message []string, dev_message string) *Error {
	return NewError(message, NotFoundErrorCode, dev_message)
}
func BadRequestError(message []string, dev_message string) *Error {
	return NewError(message, BadRequestErrorCode, dev_message)
}
func ValidationError(errors error, dev_message string) *Error {
	valErrors, ok := errors.(validator.ValidationErrors)
	if !ok {
		return sysError([]string{"unhandled errors type"})
	}
	var errs []string
	for _, err := range valErrors {
		switch err.Tag() {
		case "required":
			errs = append(errs, fmt.Sprintf("%s: field is required", err.Field()))
		case "oneof":
			errs = append(errs, fmt.Sprintf("%s: field can only be: %s", err.Field(), err.Param()))
		case "min":
			errs = append(errs, fmt.Sprintf("%s must be at least %s characters length", err.Field(), err.Param()))
		case "max":
			errs = append(errs, fmt.Sprintf("%s can't be more that %s characters length", err.Field(), err.Param()))
		case "email":
			errs = append(errs, fmt.Sprintf("%s must be a valid email", err.Field()))
		case "jwt":
			errs = append(errs, fmt.Sprintf("%s must be a JWT token", err.Field()))
		case "lowercase":
			errs = append(errs, fmt.Sprintf("%s must contain at least one lowercase character", err.Field()))
		case "uppercase":
			errs = append(errs, fmt.Sprintf("%s must contain at least one uppercase character", err.Field()))
		case "digitrequired":
			errs = append(errs, fmt.Sprintf("%s must contain at least one digit", err.Field()))
		case "specialsymbol":
			errs = append(errs, fmt.Sprintf("%s must contain at least one special symbol", err.Field()))
		case "onlyenglish":
			errs = append(errs, fmt.Sprintf("%s must contain only latin characters", err.Field()))
		default:
			errs = append(errs, err.Error())
		}
	}
	return NewError(errs, ValidationErrorCode, dev_message)
}
func UnauthorizedError(message []string, dev_message string) *Error {
	return NewError(message, UnauthorizedErrorCode, dev_message)
}
func NotUniqueError(message []string, dev_message string) *Error {
	return NewError(message, NotUniqueErrorCode, dev_message)
}
func ForbiddenError(message []string, dev_message string) *Error {
	return NewError(message, ForbiddenErrorCode, dev_message)
}
