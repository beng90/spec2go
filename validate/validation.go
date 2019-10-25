package validate

import (
	"errors"
	"github.com/beng90/spec2go/validate/validations"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"regexp"
)

var (
	ErrInvalidJSON = errors.New("invalid json")
)

func registerCustomValidations(validator *validator.Validate) {
	_ = validator.RegisterValidation("ISO8601", IsISO8601Date)
	_ = validator.RegisterValidation("boolean", validations.IsBoolean)
	_ = validator.RegisterValidation("string", validations.IsString)
	_ = validator.RegisterValidation("integer", IsNumber)
	_ = validator.RegisterValidation("object", IsObject)
	_ = validator.RegisterValidation("notblank", validations.NotBlank)
}

func IsISO8601Date(fl validator.FieldLevel) bool {
	ISO8601DateRegexString := "^(-?(?:[1-9][0-9]*)?[0-9]{4})-(1[0-2]|0[1-9])-(3[01]|0[1-9]|[12][0-9])(?:T|\\s)(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])?(Z)?$"
	ISO8601DateRegex := regexp.MustCompile(ISO8601DateRegexString)

	return ISO8601DateRegex.MatchString(fl.Field().String())
}

func IsNumber(fl validator.FieldLevel) bool {
	switch fl.Field().Kind() {
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}

	return false
}

func IsObject(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.Map {
		return true
	}

	return false
}

type ValidationErrors map[string][]FieldError

func (v ValidationErrors) Error() string {
	return ""
}

func (errs ValidationErrors) try(fieldName string, err error) {
	if err != nil {
		e := err.(validator.ValidationErrors)

		errs[fieldName] = append(errs[fieldName], FieldError{
			Field:            fieldName,
			Rule:             e[0].Tag(),
			Value:            e[0].Value(),
			Accepted:         e[0].Param(),
			ValidationErrors: e,
		})
	}
}
