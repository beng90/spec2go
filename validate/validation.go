package validate

import (
	"errors"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"regexp"
)

var (
	ErrInvalidJSON = errors.New("invalid json")
)

func registerCustomValidations(validator *validator.Validate) {
	_ = validator.RegisterValidation("ISO8601", IsISO8601Date)
	_ = validator.RegisterValidation("boolean", IsBoolean)
	_ = validator.RegisterValidation("string", IsString)
	_ = validator.RegisterValidation("integer", IsNumber)
}

func IsISO8601Date(fl validator.FieldLevel) bool {
	ISO8601DateRegexString := "^(-?(?:[1-9][0-9]*)?[0-9]{4})-(1[0-2]|0[1-9])-(3[01]|0[1-9]|[12][0-9])(?:T|\\s)(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])?(Z)?$"
	ISO8601DateRegex := regexp.MustCompile(ISO8601DateRegexString)

	return ISO8601DateRegex.MatchString(fl.Field().String())
}

func IsBoolean(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.Bool {
		return true
	}

	return false
}

func IsString(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.String {
		return true
	}

	return false
}

func IsNumber(fl validator.FieldLevel) bool {
	switch fl.Field().Kind() {
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}

	return false
}

type ValidationErrors map[string][]FieldError

func (v ValidationErrors) Error() string {
	return ""
}
