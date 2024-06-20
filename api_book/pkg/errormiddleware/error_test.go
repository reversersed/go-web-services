package errormiddleware

import (
	"reflect"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_book/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := NewError([]string{"hello", "world"}, "IE-0001", "this is test message")

	er := &Error{
		Code:             "IE-0001",
		Message:          []string{"hello", "world"},
		DeveloperMessage: "this is test message",
	}

	assert.EqualError(t, err, er.Error())

	bytes := err.Marshall()
	assert.NotNil(t, bytes)
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
			assert.Equal(t, errorCase.Err.Code, errorCase.ExceptedCode)
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
	assert.Equal(t, err.Code, InternalErrorCode)

	errs := validator.New().Struct(validationStruct)
	if assert.NotNil(t, errs) {
		err = ValidationError(errs, "")

		assert.Equal(t, err.Code, ValidationErrorCode)
		assert.Equal(t, reflect.TypeOf(validationStruct).NumField(), len(err.Message), "excepted %d errors but got %d", reflect.TypeOf(validationStruct).NumField(), len(err.Message))
	}
}
