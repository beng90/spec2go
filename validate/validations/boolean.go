package validations

import (
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

func IsBoolean(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.Bool {
		return true
	}

	return false
}
