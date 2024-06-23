package validator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ValidationErrors validator.ValidationErrors
type Validator struct {
	*validator.Validate
}

func New() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	v.RegisterValidation("primitiveid", validate_PrimitiveId)
	v.RegisterValidation("lowercase", validate_LowercaseCharacter)
	v.RegisterValidation("uppercase", validate_UppercaseCharacter)
	v.RegisterValidation("digitrequired", validate_AtLeastOneDigit)
	v.RegisterValidation("specialsymbol", validate_SpecialSymbol)
	v.RegisterValidation("onlyenglish", validate_OnlyEnglish)
	return &Validator{v}
}
func validate_PrimitiveId(field validator.FieldLevel) bool {
	var obj primitive.ObjectID

	_, err := primitive.ObjectIDFromHex(field.Field().String())
	return (err == nil) || (field.Field().Kind() == reflect.TypeOf(obj).Kind())
}
func validate_OnlyEnglish(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString(`^[a-zA-Z]+$`, field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
func validate_LowercaseCharacter(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[a-z]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
func validate_UppercaseCharacter(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[A-Z]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
func validate_AtLeastOneDigit(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[0-9]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
func validate_SpecialSymbol(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[!@#\\$%\\^&*()_\\+-.,]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
