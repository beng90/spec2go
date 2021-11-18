package validations

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func IsString(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.String {
		return true
	}

	return false
}
