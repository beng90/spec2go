package validators

import (
	"encoding/json"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

// registerCustomValidations set custom validators
func registerCustomValidations(validator *validator.Validate) {
	_ = validator.RegisterValidation("ISO8601", IsISO8601Date)
	_ = validator.RegisterValidation("boolean", IsBoolean)
	_ = validator.RegisterValidation("string", IsString)
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

type ValidationErrors map[string][]VError

type VError struct {
	Field            string
	Rule             string
	Value            interface{}
	ValidationErrors validator.ValidationErrors
}

func (v VError) Error() string {
	if v.Value != nil {
		return fmt.Sprintf(`Field "%s" failed in "%s" rule, given "%v"`, v.Field, v.Rule, v.Value)
	}

	return fmt.Sprintf(`Field "%s" failed in "%s" rule`, v.Field, v.Rule)
}

func try(errs ValidationErrors, fieldName string, err error) {
	if err != nil {
		e := err.(validator.ValidationErrors)

		errs[fieldName] = append(errs[fieldName], VError{
			Field:            fieldName,
			Rule:             e[0].Tag(),
			Value:            e[0].Value(),
			ValidationErrors: e,
		})
	}
}

type jsonMap map[string]interface{}

type jsonField struct {
	name  string
	value interface{}
}

func (j jsonMap) getVal(exploded []string, i int, prev interface{}, data *[]jsonField) {
	if len(exploded) <= i {
		return
	}

	fn := strings.Trim(exploded[i], "[]")
	//fmt.Println("fn", fn)
	//fmt.Println("prev", prev)

	switch v := prev.(type) {
	case string, float64, int, bool, nil:
		*data = append(*data, jsonField{
			name:  fn,
			value: v,
		})
	case jsonMap:
		//fmt.Println("--- jsonMap ---")

		// take next index from map and go through
		//fmt.Println("exploded", fn, v[fn])
		if i+1 < len(exploded) {
			j.getVal(exploded, i+1, v[fn], data)
		} else {
			j.getVal(exploded, i, v[fn], data)
		}
		//*data = v[exploded[i]]
		//for fnn, vv := range v[fn] {
		//	fmt.Println("vv", fnn, vv)
		//}
	case map[string]interface{}:
		//fmt.Println("--- map[string]interface{} ---")
		//fmt.Println("v", fn, v[fn])
		if i+1 < len(exploded) {
			j.getVal(exploded, i+1, v[fn], data)
		} else {
			j.getVal(exploded, i, v[fn], data)
		}
		//}
	case []interface{}:
		//fmt.Println("--- []interface{} ---")

		for _, vv := range v {
			//fmt.Println("exploded[i]", i, exploded[i], vv)
			j.getVal(exploded, i, vv, data)
		}
	default:
		fmt.Printf("Default: %T\n", v)
	}
}

func (j jsonMap) Get(fieldName string) interface{} {
	exploded := strings.Split(fieldName, ".")
	if len(exploded) > 0 {
		var val []jsonField
		j.getVal(exploded, 0, j, &val)

		//fmt.Println("val", val)
		return val
	}

	return nil
}

func getRequestBody(r *http.Request) jsonMap {
	// Read body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	var requestBody jsonMap
	if err := json.Unmarshal(body, &requestBody); err != nil {
		panic(err)
	}

	return requestBody
}

type SchemaValidator struct {
	validator   *validator.Validate
	requestBody jsonMap
	errors      ValidationErrors
}

func NewSchemaValidator(v *validator.Validate, req *http.Request) *SchemaValidator {
	// custom validations
	registerCustomValidations(v)

	requestBody := getRequestBody(req)
	errs := make(ValidationErrors)

	return &SchemaValidator{v, requestBody, errs}
}

func (s *SchemaValidator) validate(fieldName string, rule string) {
	value := s.requestBody.Get(fieldName)

	switch v := value.(type) {
	case []jsonField:
		for _, vv := range v {
			//fmt.Println("===", vv)
			err := s.validator.Var(vv.value, rule)
			//fn := fmt.Sprintf("%s[%d]", strings.Trim(fieldName, "[]"), i)
			try(s.errors, fieldName, err)
		}
	default:
		fmt.Printf("Default: %T\n", v)
		err := s.validator.Var(v, rule)
		try(s.errors, fieldName, err)
	}

	return
}
