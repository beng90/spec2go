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
	"strconv"
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

type jsonField struct {
	name  string
	value interface{}
}

func (j MapField) GetVal(fieldName string) interface{} {
	exploded := strings.Split(fieldName, ".")
	if len(exploded) > 0 {
		var val []jsonField
		var prev interface{}
		for _, part := range exploded {
			fn := strings.Trim(part, "[]")
			fmt.Println("part", fn, prev)
			var current interface{} = j[fn]

			if current != nil {
				switch v := current.(type) {
				case FieldSchema:
					if v.Items != nil {
						for _, vv := range v.Items {
							fmt.Println("vv", vv)
						}
					}
				default:
					fmt.Printf("%T", v)
				}
			}

			prev = current
		}

		return val
	}

	return nil
}

func getRequestBody(req *http.Request) (requestBody MapField, err error) {
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
	requestBody MapField
	rules       RulesMap
	valuesMap   ValuesMap
	errors      ValidationErrors
}

type ValuesMap map[string]interface{}

func (v ValuesMap) Add(path string, value interface{}) {
	v[path] = value
}

type RulesMap map[string]Rule
type Rule struct {
	Path   string
	Rules  []string
	Passed bool
}

func NewSchemaValidator(v *validator.Validate, req *http.Request) (schemaValidator *SchemaValidator, err error) {
	// custom validations
	registerCustomValidations(v)

	requestBody, err := getRequestBody(req)
	if err != nil {
		return nil, err
	}

	schemaValidator = &SchemaValidator{
		v,
		requestBody,
		make(RulesMap),
		make(ValuesMap),
		make(ValidationErrors),
	}

	return
}

func (s *SchemaValidator) AddRule(path string, rule string) {
	if s.rules == nil {
		s.rules = make(RulesMap)
	}

	rulesSlice := strings.Split(rule, ",")
	s.rules[path] = Rule{path, rulesSlice, false}
}

func (s *SchemaValidator) hasRule(path []string) bool {
	if _, ok := s.rules[s.fieldPath(path)]; ok {
		return true
	}

	return false
}

func (s *SchemaValidator) getRule(path []string) *Rule {
	if rule, ok := s.rules[s.ruleName(path)]; ok {
		return &rule
	}

	return nil
}

func (s *SchemaValidator) fieldPath(path []string) string {
	return strings.Join(path, ".")
}

func (s *SchemaValidator) ruleName(path []string) string {
	var re = regexp.MustCompile(`(?m)\[(\d)\]`)
	ruleName := re.ReplaceAllString(strings.Join(path, "."), `[]`)

	return ruleName
}

func (s *SchemaValidator) walk(data MapField, path []string) {
	for fieldName, field := range data {
		if field.Properties != nil {
			path = append(path, fieldName)
			//rule := s.getRule(path)
			//if rule != nil && len(rule.Rules) > 0 {
			//	field.Rules = rule.Rules
			//	fmt.Println("ruleName", s.ruleName(path), len(rule.Rules))
			//}

			s.walk(field.Properties, path)
			path = path[:len(path)-1]
		} else if field.Items != nil {
			rule := s.getRule(append(path, fieldName+"[]"))
			//fmt.Println("fieldName", fieldName, path)
			if rule != nil {
				field.Rules = rule.Rules
				//fmt.Println("ruleName", s.ruleName(path), rule)
				//fmt.Printf("%#v\n", field.Value)
			}

			for i, item := range field.Items {
				path = append(path, fieldName+"["+strconv.Itoa(i)+"]")
				s.walk(item, path)

				path = path[:len(path)-1]
			}
		} else {
			// TODO: change hardcode
			if fieldName != "value" {
				path = append(path, fieldName)
			}

			s.valuesMap.Add(s.fieldPath(path), field.Value)
			//fmt.Println("ruleName", s.ruleName(path))
			// add rules to struct
			rule := s.getRule(path)
			if rule != nil {
				field.Rules = rule.Rules
			}

			path = path[:len(path)-1]
		}
	}
}

func (s *SchemaValidator) Validate() error {
	errs := make(ValidationErrors)

	s.walk(s.requestBody, []string{})
	//s.requestBody["additionalInfo"].Rules = []string{"zzz", "yyy"}
	fmt.Printf("additionalInfo %#v\n", s.requestBody["additionalInfo"].Rules)
	fmt.Printf("brand %#v\n", s.requestBody["brand"].Rules)
	//fmt.Printf("s.requestBody:\n %#v\n", s.requestBody["additionalInfo"].Items[0]["id"].Rules)

	//for fieldName, rule := range s.rules {
	//	//fmt.Println("field", fieldName, s.valuesMap["additionalInfo[0].id"])
	//	if value, ok := s.valuesMap[fieldName]; ok {
	//		err := s.validator.Var(value, rule.Rules)
	//		try(errs, rule.Path, err)
	//	} else {
	//		err := s.validator.Var(nil, rule.Rules)
	//		try(errs, rule.Path, err)
	//	}
	//}

	//for _, rule := range s.rules {
	//	value := s.requestBody.GetVal(rule.Path)
	//
	//	switch v := value.(type) {
	//	case []jsonField:
	//		if len(v) > 0 {
	//			for _, vv := range v {
	//				debug("val", vv.value, "rule", rule.Rules)
	//				switch vvv := vv.value.(type) {
	//				case []interface{}:
	//					for _, singleValue := range vvv {
	//						err := s.validator.Var(singleValue, rule.Rules)
	//						try(errors, rule.Path, err)
	//					}
	//				default:
	//					err := s.validator.Var(vv.value, rule.Rules)
	//					try(errors, rule.Path, err)
	//				}
	//			}
	//		} else {
	//			err := s.validator.Var(v, rule.Rules)
	//			try(errors, rule.Path, err)
	//		}
	//	default:
	//		fmt.Printf("Default: %T\n", v)
	//		err := s.validator.Var(v, rule.Rules)
	//		try(errors, rule.Path, err)
	//	}
	//}

	return errs
}
