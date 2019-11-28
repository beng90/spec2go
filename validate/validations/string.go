package validations

import (
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

func IsString(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.String {
		return true
	}

	return false
}
