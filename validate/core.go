package validate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
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

type FieldError struct {
	Field            string
	Rule             string
	Value            interface{}
	Accepted         string
	ValidationErrors validator.ValidationErrors
}

func (v FieldError) Error() string {
	msg := fmt.Sprintf(`Field '%s' failed in '%s' rule`, v.Field, v.Rule)

	values := v.Accepted
	if values != "" {
		msg += ", available values: " + values
	}

	return msg
}

func try(errs ValidationErrors, fieldName string, err error) {
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
		debug("--- jsonMap ---")

		// take next index from map and go through
		//fmt.Println("exploded", fn, v[fn])
		if i+1 < len(exploded) {
			j.getVal(exploded, i+1, v[fn], data)
		} else {
			j.getVal(exploded, i, v[fn], data)
		}
	case map[string]interface{}:
		debug("--- map[string]interface{} ---")
		//debug("v", fn, v[fn])
		//debug("exploded", exploded[i], v)

		// for 1 level
		if len(exploded) == 1 {
			*data = append(*data, jsonField{
				name:  fn,
				value: v,
			})
		} else {

			if i+1 < len(exploded) {
				j.getVal(exploded, i+1, v[fn], data)
			} else {
				//j.getVal(exploded, i, v[fn], data)
				//is last element from exploded
				debug("v", fn, v[fn])
				*data = append(*data, jsonField{
					name:  fn,
					value: v[fn],
				})
			}
		}
	case []interface{}:
		debug("--- []interface{} ---")

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

		debug("val", val)
		return val
	}

	return nil
}

func getRequestBody(req *http.Request) (requestBody jsonMap, err error) {
	// Read body
	buffer, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// restore body in request
	req.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))

	if json.Valid(buffer) == false {
		return nil, ErrInvalidJSON
	}

	if err := json.Unmarshal(buffer, &requestBody); err != nil {
		panic(err)
	}

	return requestBody, nil
}

type SchemaValidator struct {
	validator   *validator.Validate
	requestBody jsonMap
	rules       []Rule
}

type Rule struct {
	Path  string
	Rules string
}

func NewSchemaValidator(v *validator.Validate, req *http.Request) (schemaValidator *SchemaValidator, err error) {
	// custom validations
	registerCustomValidations(v)

	requestBody, err := getRequestBody(req)
	if err != nil {
		return nil, err
	}

	schemaValidator = &SchemaValidator{v, requestBody, nil}

	return
}

func (s *SchemaValidator) AddRule(path string, rule string) {
	s.rules = append(s.rules, Rule{path, rule})
}

func (s *SchemaValidator) Validate() error {
	errors := make(ValidationErrors)

	for _, rule := range s.rules {
		value := s.requestBody.Get(rule.Path)

		switch v := value.(type) {
		case []jsonField:
			if len(v) > 0 {
				for _, vv := range v {
					debug("val", vv.value, "rule", rule.Rules)
					switch vvv := vv.value.(type) {
					case []interface{}:
						for _, singleValue := range vvv {
							err := s.validator.Var(singleValue, rule.Rules)
							try(errors, rule.Path, err)
						}
					default:
						err := s.validator.Var(vv.value, rule.Rules)
						try(errors, rule.Path, err)
					}
				}
			} else {
				err := s.validator.Var(v, rule.Rules)
				try(errors, rule.Path, err)
			}
		default:
			fmt.Printf("Default: %T\n", v)
			err := s.validator.Var(v, rule.Rules)
			try(errors, rule.Path, err)
		}
	}

	return errors
}
