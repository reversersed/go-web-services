package errormiddleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
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
			r := httptest.NewRequest(http.MethodGet, "http://test", nil)
			err := handler(w, r)
			if w.Result().StatusCode != errorCase.ExceptedStatus {
				t.Fatalf("excepted status code %d but got %d", errorCase.ExceptedStatus, w.Result().StatusCode)
			}
			if errorCase.Err != nil && err == nil {
				t.Fatal("excepted error but got nil")
			}
			if errorCase.Err == nil && err == nil {
				t.SkipNow()
			}
			Err, errOk := err.(*Error)
			Error, ErrorOk := errorCase.Err.(*Error)
			if !errOk || !ErrorOk {
				if err.Error() != errorCase.Err.Error() {
					t.Errorf("excepted error %s but got %s", errorCase.Err.Error(), err.Error())
				}
			} else {
				if Err.Code != Error.Code {
					t.Errorf("excepeted error code %s but got %s", Error.Code, Err.Code)
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

	if err == nil {
		t.Fatal("excepted error but got nil")
	}
	if w.Result().StatusCode != http.StatusNotImplemented {
		t.Fatalf("excepeted status code %d but got %d", http.StatusNotImplemented, w.Result().StatusCode)
	}
	Err, ok := err.(*Error)
	if !ok {
		t.Fatalf("excepted custom error but got %v", err)
	}
	if fields := reflect.TypeOf(validationStruct).NumField(); fields != len(Err.Message) {
		t.Fatalf("excepted %d errors but got %d", fields, len(Err.Message))
	}
	var bodyError Error
	if errs := json.NewDecoder(w.Result().Body).Decode(&bodyError); errs != nil {
		t.Fatalf("excepted body error but got %v", errs)
	}
	if bodyError.Code != Err.Code {
		t.Fatalf("excepeted body error code %s but got %s", Err.Code, bodyError.Code)
	}
}
