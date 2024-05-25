package errormiddleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Message          []string `json:"messages,omitempty"`
	DeveloperMessage string   `json:"dev_message,omitempty"`
	Code             string   `json:"code,omitempty"`
}

func NewError(message []string, code, dev_message string) *Error {
	return &Error{
		Code:             code,
		Message:          message,
		DeveloperMessage: dev_message,
	}
}

func (e *Error) Error() string {
	return strings.Join(e.Message, ", ")
}
func (e *Error) Unwrap() error { return fmt.Errorf(strings.Join(e.Message, ", ")) }

func (e *Error) Marshall() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return bytes
}
func sysError(message []string) *Error {
	return NewError(message, "IE-0001", "Something wrong happened while service executing")
}
func NotFoundError(message []string, dev_message string) *Error {
	return NewError(message, "IE-0002", dev_message)
}
func BadRequestError(message []string, dev_message string) *Error {
	return NewError(message, "IE-0003", dev_message)
}
func ValidationError(errors validator.ValidationErrors, dev_message string) *Error {
	var errs []string
	for _, err := range errors {
		errs = append(errs, err.Error())
	}
	return NewError(errs, "IE-0004", dev_message)
}
func ValidationErrorByString(errors []string, dev_message string) *Error {
	return NewError(errors, "IE-0004", dev_message)
}
func UnauthorizedError(errors []string, dev_message string) *Error {
	return NewError(errors, "IE-0005", dev_message)
}
func NotUniqueError(errors []string, dev_message string) *Error {
	return NewError(errors, "IE-0006", dev_message)
}
func ForbiddenError(errors []string, dev_message string) *Error {
	return NewError(errors, "IE-0007", dev_message)
}
