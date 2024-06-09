package rest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type CustomResponse struct {
	Valid    bool
	response *http.Response
	Error    CustomError
}

func (r *CustomResponse) Body() io.ReadCloser {
	return r.response.Body
}
func (r *CustomResponse) StatusCode() int {
	return r.response.StatusCode
}

type CustomError struct {
	Message          []string `json:"messages,omitempty"`
	ErrorCode        string   `json:"code,omitempty"`
	DeveloperMessage string   `json:"dev_message,omitempty"`
}

func (e CustomError) Error() string {
	return fmt.Sprintf("Error code: %s, Error: %s, Dev message: %s", e.ErrorCode, strings.Join(e.Message, ", "), e.DeveloperMessage)
}
