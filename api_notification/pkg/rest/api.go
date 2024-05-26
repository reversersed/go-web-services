package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
func (r *CustomResponse) ReadBody() ([]byte, error) {
	defer r.response.Body.Close()
	return io.ReadAll(r.response.Body)
}
func (r *CustomResponse) StatusCode() int {
	return r.response.StatusCode
}
func (r *CustomResponse) Location() (*url.URL, error) {
	return r.response.Location()
}

type CustomError struct {
	Message          []string `json:"messages,omitempty"`
	ErrorCode        string   `json:"code,omitempty"`
	DeveloperMessage string   `json:"dev_message,omitempty"`
}

func (e *CustomError) ToString() string {
	return fmt.Sprintf("Error code: %s, Error: %s, Dev message: %s", e.ErrorCode, strings.Join(e.Message, ", "), e.DeveloperMessage)
}
func (e CustomError) Error() string {
	return fmt.Sprintf("Error code: %s, Error: %s, Dev message: %s", e.ErrorCode, strings.Join(e.Message, ", "), e.DeveloperMessage)
}
