package errormiddleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/validator"
)

var errorCodeCases = []struct {
	Name           string
	Err            error
	ExceptedStatus int
}{
	{"Internal default error", errors.New("internal error"), http.StatusInternalServerError},
	{"Internal custom error", sysError([]string{""}), http.StatusInternalServerError},
	{"Not found custom error", NotFoundError([]string{""}, ""), http.StatusNotFound},
	{"Bad request custom error", BadRequestError([]string{""}, ""), http.StatusBadRequest},
	{"Unauthorized custom error", UnauthorizedError([]string{""}, ""), http.StatusUnauthorized},
	{"NotUnique custom error", NotUniqueError([]string{""}, ""), http.StatusConflict},
	{"Forbidden custom error", ForbiddenError([]string{""}, ""), http.StatusForbidden},
	{"Unknowed custom error code", NewError([]string{""}, "0000", ""), http.StatusBadRequest},
	{"Successful response", nil, http.StatusOK},
}

func TestMiddleWareErrorCodes(t *testing.T) {
	for _, errorCase := range errorCodeCases {
		t.Run(errorCase.Name, func(t *testing.T) {
			handler := Middleware(func(w http.ResponseWriter, r *http.Request) error {
				if errorCase.Err != nil {
					return errorCase.Err
				}
				w.WriteHeader(http.StatusOK)
				return nil
			})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://test", http.NoBody)
			err := handler(w, r)

			assert.Equal(t, w.Result().StatusCode, errorCase.ExceptedStatus)
			assert.Equal(t, errorCase.Err, err)
			if err != nil {
				Err, errOk := err.(*Error)
				Error, ErrorOk := errorCase.Err.(*Error)
				if !errOk || !ErrorOk {
					assert.Equal(t, err.Error(), errorCase.Err.Error())
				} else {
					assert.Equal(t, Err.Code, Error.Code)
				}
			}
		})
	}
}

func TestValidationMiddleware(t *testing.T) {
	handler := Middleware(func(w http.ResponseWriter, r *http.Request) error {
		return ValidationError(validator.New().Struct(validationStruct), "")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	err := handler(w, r)

	assert.Error(t, err)
	assert.Equal(t, w.Result().StatusCode, http.StatusNotImplemented)

	Err, ok := err.(*Error)
	assert.True(t, ok, "excepted custom error but got %v", err)

	fields := reflect.TypeOf(validationStruct).NumField()
	assert.Equal(t, fields, len(Err.Message), "excepted %d errors but got %d", fields, len(Err.Message))

	var bodyError Error
	errs := json.NewDecoder(w.Result().Body).Decode(&bodyError)
	assert.NoError(t, errs)
	assert.Equal(t, bodyError.Code, Err.Code)
}
