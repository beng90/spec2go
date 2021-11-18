package validations

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// NotBlank is the validation function for validating if the current field
// has a value or length greater than zero, or is not a space only string.
func NotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(field.String())) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	case reflect.Bool:
		return true
	default:
		if field.IsValid() {
			myType := reflect.TypeOf(field.Interface())

			switch myType.Kind() {
			case reflect.Int:
				return field.Int() >= int64(0)
			case reflect.Float64:
				return field.Float() >= float64(0)
			}

			return field.Interface() != reflect.Zero(field.Type()).Interface()
		}

		return false
	}
}
