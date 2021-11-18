package validations

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func IsBoolean(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.Bool {
		return true
	}

	return false
}
