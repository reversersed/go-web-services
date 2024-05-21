package errormiddleware

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Err              error  `json:"-"`
	Message          string `json:"message,omitempty"`
	DeveloperMessage string `json:"dev_message,omitempty"`
	Code             string `json:"code,omitempty"`
}

func NewError(message, code, dev_message string) *Error {
	return &Error{
		Err:              fmt.Errorf(message),
		Code:             code,
		Message:          message,
		DeveloperMessage: dev_message,
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}
func (e *Error) Unwrap() error { return e.Err }

func (e *Error) Marshall() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return bytes
}
func sysError(message string) *Error {
	return NewError(message, "IE-0001", "something wrong happened while program executing")
}
