package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var cases = []struct {
	name  string
	field string
	tag   string
	err   bool
}{
	{"primitive id tag testing", "6665976c2691650b53a24009", "primitiveid", false},
	{"primitive id tag fail testing", "", "primitiveid", true},
	{"lowercase tag testing", "This Is Lower Containing String With 1 Number And !@#$% Specials", "lowercase", false},
	{"lowercase tag fail testing", "THIS IS NOT LOWERCASE WITH 2 NUMBER AND @#%$#@ SPECIALS", "lowercase", true},
	{"uppercase tag testing", "This Is Upper Containing String With 1 Number And !@#$% Specials", "uppercase", false},
	{"uppercase tag fail testing", "this is not uppercase containig string with 252 numbers and $!@ specials", "uppercase", true},
	{"only english tag testing", "OnlyEnglishLetters", "onlyenglish", false},
	{"only english tag fail testing on spaces", "this is only english letters testing but with spaces", "onlyenglish", true},
	{"only english tag fail testing on numbers", "this is only english letters testing but with 513 number", "onlyenglish", true},
	{"only english tag fail testing on other letters", "этот тест тэга onlyenglish должен провалиться, т.к. тут есть русские символы", "onlyenglish", true},
	{"only english tag fail testing on specials", "this is only english letters testing but with %@#!&*(@#", "onlyenglish", true},
	{"digit required tag testing", "there is number 2", "digitrequired", false},
	{"digit required tag fail testing", "there is not number", "digitrequired", true},
	{"specials required tag testing", "there is special !", "specialsymbol", false},
	{"specials required tag fail testing", "there is not specials", "specialsymbol", true},
}

func TestValidator(t *testing.T) {
	valid := New()

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			err := valid.Var(v.field, v.tag)

			if v.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type fieldStruct struct {
	Field string `json:"field" validate:"required,digitrequired"`
}

func TestFieldName(t *testing.T) {
	str := &fieldStruct{Field: "there is supposed to be an error"}
	valid := New()

	err := valid.Struct(str)
	assert.Error(t, err)

	errs, ok := err.(validator.ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, errs[0].Field(), "field")
}

type emptyFieldStruct struct {
	Field string `json:"-" validate:"required,digitrequired"`
}

func TestEmptyFieldName(t *testing.T) {
	str := &emptyFieldStruct{Field: "there is supposed to be an error"}
	valid := New()

	err := valid.Struct(str)
	assert.Error(t, err)

	errs, ok := err.(validator.ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, errs[0].Field(), "Field")
}
