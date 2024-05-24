package errormiddleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Message          string   `json:"message,omitempty"`
	DeveloperMessage []string `json:"dev_messages,omitempty"`
	Code             string   `json:"code,omitempty"`
}

func NewError(message, code string, dev_message []string) *Error {
	return &Error{
		Code:             code,
		Message:          message,
		DeveloperMessage: dev_message,
	}
}

func (e *Error) Error() string {
	return strings.Join(e.DeveloperMessage, ", ")
}
func (e *Error) Unwrap() error { return fmt.Errorf(strings.Join(e.DeveloperMessage, ", ")) }

func (e *Error) Marshall() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return bytes
}
func sysError(dev_message []string) *Error {
	return NewError("Something wrong happened", "IE-0001", dev_message)
}
func NotFoundError(message string, dev_message []string) *Error {
	return NewError(message, "IE-0002", dev_message)
}
func BadRequestError(message string, dev_message []string) *Error {
	return NewError(message, "IE-0003", dev_message)
}
func ValidationError(errors validator.ValidationErrors, message string) *Error {
	var errs []string
	for _, err := range errors {
		errs = append(errs, err.Error())
	}
	return NewError(message, "IE-0004", errs)
}
func ValidationErrorByString(errors []string, message string) *Error {
	return NewError(message, "IE-0004", errors)
}
func UnauthorizedError(errors []string, message string) *Error {
	return NewError(message, "IE-0005", errors)
}
func NotUniqueError(errors []string, message string) *Error {
	return NewError(message, "IE-0006", errors)
}
