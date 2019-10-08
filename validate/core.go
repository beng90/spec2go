package validate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

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

type TreeField struct {
	Field string
	Value interface{}
	Rule  Rules
}

func (s *SchemaValidator) walk(data MapField, path []string, tree []TreeField) {
	for fieldName, field := range data {
		//fmt.Println("fieldName", fieldName)
		if field.Properties != nil {
			path = append(path, fieldName)
			rule := s.getRule(path)
			treeRule := []string{}
			if rule != nil {
				treeRule = rule.Rules
			}

			tree = append(tree, TreeField{
				Field: fieldName,
				Value: data,
				Rule:  treeRule,
			})

			s.walk(field.Properties, path, tree)
			path = path[:len(path)-1]
			tree = tree[:len(tree)-1]
		} else if field.Items != nil {
			rule := s.getRule(append(path, fieldName+"[]"))

			if rule != nil {
				field.Rules = rule.Rules
				//fmt.Println("ruleName", s.ruleName(path), rule, prev)
				//fmt.Printf("%#v\n", field.Value)
			}

			//fmt.Println("fieldName", fieldName, field.Value)
			tree = append(tree, TreeField{
				Field: fieldName,
				Value: field.Value,
				Rule:  field.Rules,
			})

			for i, item := range field.Items {
				treeRule := []string{}
				if rule != nil {
					treeRule = rule.Rules
				}

				tree = append(tree, TreeField{
					Field: fieldName + "[" + strconv.Itoa(i) + "]",
					Value: item.Get("value"),
					Rule:  treeRule,
				})

				path = append(path, fieldName+"["+strconv.Itoa(i)+"]")
				s.walk(item, path, tree)

				path = path[:len(path)-1]
				tree = tree[:len(tree)-1]
			}
		} else {
			field.Name = fieldName
			s.processField(field, tree, path)
		}
	}
}

func (s *SchemaValidator) processField(field *FieldSchema, tree []TreeField, path []string) {
	prev := TreeField{}
	if len(tree) > 0 {
		if field.Name == "value" {
			prev = tree[len(tree)-2]
		} else if len(tree) >= 2 {
			prev = tree[len(tree)-2]
		}

		//fmt.Println("prev", s.fieldPath(path), prev.Value, prev.Rule)
	}

	// TODO: change hardcode
	if field.Name != "value" {
		path = append(path, field.Name)
	}

	s.valuesMap.Add(s.fieldPath(path), field.Value)

	rule := s.getRule(path)
	if rule != nil {
		field.Rules = rule.Rules
	}

	//fmt.Println("current", s.fieldPath(path), field.Rules)

	if field.IsRequired() && field.Value == nil {
		if prev.Rule.Required() {
			vErr := s.validator.Var(field.Value, field.Rules.String())
			//fmt.Println("vErr", vErr)
			try(s.errors, field.Name, vErr)
		}
	} else {
		vErr := s.validator.Var(field.Value, field.Rules.String())
		//fmt.Println("vErr", vErr)
		try(s.errors, s.fieldPath(path), vErr)
	}

	tree = append(tree, TreeField{
		Field: field.Name,
		Value: field.Value,
		Rule:  field.Rules,
	})

	path = path[:len(path)-1]
	tree = tree[:len(tree)-1]
}

func (s *SchemaValidator) Validate() error {
	s.walk(s.requestBody, []string{}, []TreeField{})
	//s.requestBody["additionalInfo"].Rules = []string{"zzz", "yyy"}
	//fmt.Printf("additionalInfo %#v\n", s.requestBody["additionalInfo"].Rules)
	//fmt.Printf("brand %#v\n", s.requestBody["brand"].Rules)
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

	return s.errors
}
