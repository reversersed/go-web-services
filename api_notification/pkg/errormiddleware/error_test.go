package errormiddleware

import (
	"reflect"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
)

func TestError(t *testing.T) {
	err := NewError([]string{"hello", "world"}, "IE-0001", "this is test message")

	er := &Error{
		Code:             "IE-0001",
		Message:          []string{"hello", "world"},
		DeveloperMessage: "this is test message",
	}

	if er.Error() != err.Error() {
		t.Fatalf("excepted error to be equal got not")
	}

	bytes := err.Marshall()
	if bytes == nil {
		t.Fatalf("excepted marshalled error but got nil")
	}
}

var errorsCases = []struct {
	Name         string
	Err          *Error
	ExceptedCode Code
}{
	{"Internal error test", sysError([]string{""}), "IE-0001"},
	{"Not found error test", NotFoundError([]string{""}, ""), "IE-0002"},
	{"Bad request error test", BadRequestError([]string{""}, ""), "IE-0003"},
	{"Unauthorized error test", UnauthorizedError([]string{""}, ""), "IE-0005"},
	{"Not unique error test", NotUniqueError([]string{""}, ""), "IE-0006"},
	{"Forbidden error test", ForbiddenError([]string{""}, ""), "IE-0007"},
}

func TestErrorCodes(t *testing.T) {
	for _, errorCase := range errorsCases {
		t.Run(errorCase.Name, func(t *testing.T) {
			if errorCase.Err.Code != errorCase.ExceptedCode {
				t.Errorf("excepted code %s but got %s", errorCase.ExceptedCode, errorCase.Err.Code)
			}
		})
	}
}

var validationStruct = struct {
	RequiredField string `validate:"required"`
	OneOf         string `validate:"oneof=HELLO HI THERE"`
	Min           string `validate:"min=4"`
	Max           string `validate:"max=1"`
	Email         string `validate:"email"`
	Jwt           string `validate:"jwt"`
	Lowercase     string `validate:"lowercase"`
	Uppercase     string `validate:"uppercase"`
	Digits        string `validate:"digitrequired"`
	Specials      string `validate:"specialsymbol"`
	OnlyEnglish   string `validate:"onlyenglish"`
	Default       string `validate:"ip"`
}{Max: "123", Lowercase: "A", Uppercase: "a"}

func TestValidationError(t *testing.T) {
	err := ValidationError(nil, "")
	if err.Code != "IE-0001" {
		t.Fatalf("excepted error code IE-0001 but got %s", err.Code)
	}
	errs := validator.New().Struct(validationStruct)
	if errs == nil {
		t.Fatalf("excepted error but got nil")
	}
	err = ValidationError(errs, "")
	if err.Code != "IE-0004" {
		t.Fatalf("excepted code IE-0004 but got %s", err.Code)
	}
	if fields := reflect.TypeOf(validationStruct).NumField(); fields != len(err.Message) {
		t.Fatalf("excepted %d errors but got %d", fields, len(err.Message))
	}
}
