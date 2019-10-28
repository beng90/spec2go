package validate

import (
	"bytes"
	"encoding/json"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type SchemaValidator struct {
	validator   *validator.Validate
	requestBody MapField
	rules       RulesMap
	errors      ValidationErrors
}

type RulesMap map[string]Rule
type Rule struct {
	Path    FieldPath
	Rules   Rules
	Pattern *string
}

func (r *Rule) Has(name string) bool {
	if r == nil {
		return false
	}

	for _, rule := range r.Rules {
		if rule == name {
			return true
		}
	}

	return false
}

type TreeField struct {
	Field string
	Value interface{}
	Rule  Rules
}

type FieldPath []string

func (path *FieldPath) add(s string) {
	*path = append(*path, s)
}

func (path FieldPath) last() string {
	return path[len(path)-1]
}

func (path *FieldPath) String() string {
	return strings.Join(*path, ".")
}

func Pattern(val string) *string {
	return &val
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
		make(ValidationErrors),
	}

	return
}

func (s *SchemaValidator) AddRule(path string, rule string, pattern *string) {
	if s.rules == nil {
		s.rules = make(RulesMap)
	}

	rulesSlice := strings.Split(rule, ",")
	pathSlice := strings.Split(path, ".")
	s.rules[path] = Rule{pathSlice, rulesSlice, pattern}
}

func (s *SchemaValidator) HasRule(path []string) bool {
	if _, ok := s.rules[s.fieldPath(path)]; ok {
		return true
	}

	return false
}

func (s *SchemaValidator) GetRule(path []string) *Rule {
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

func (s *SchemaValidator) getValue(exploded FieldPath, index int, fieldsTree FieldsArray, values *[]FieldSchema, path FieldPath) {
	fieldName := exploded[index]
	lastValue := fieldsTree.last()
	rule := s.rules[exploded.String()]
	rules := rule.Rules
	parent := lastValue.Get(fieldName)
	parent.Name = path.String()
	parent.Rules = rules

	//fmt.Println("field", fieldName)

	// last path element
	if index == len(exploded)-1 {
		path.add(fieldName)

		// for arrays
		if strings.Contains(exploded[index], "[]") {
			if parent.Items != nil && len(parent.Items) > 0 {
				path[len(path)-1] = strings.Trim(path[len(path)-1], "[]")
				for i, item := range parent.Items {
					singleItem := item.Get("arrayItem")
					// check if its array of strings
					singleItem.Value = item.Get("arrayItem").Value
					singleItem.Name = path.String() + "[" + strconv.Itoa(i) + "]"
					singleItem.Rules = rules
					singleItem.Rule = rule
					*values = append(*values, singleItem)
				}

				return
			}

			// there is no items in array
			parent.Value = nil
			parent.Name = path.String()
			parent.Rule = rule
			*values = append(*values, parent)

			return
		}

		current := lastValue[fieldName]
		current.Name = path.String()
		current.Rules = rules
		current.Rule = rule
		*values = append(*values, current)

		return
	}

	// has properties
	if lastValue.Get(fieldName).Properties != nil {
		path.add(fieldName)
		fieldsMap := fieldsTree.last().Get(fieldName).Properties
		fieldsTree = append(fieldsTree, fieldsMap)

		s.getValue(exploded, index+1, fieldsTree, values, path)

		path = path[:len(path)-1]

	} else if lastValue.Get(fieldName).Items != nil {
		// has items
		for i, item := range fieldsTree.last().Get(fieldName).Items {
			path.add(strings.Trim(fieldName, "[]") + "[" + strconv.Itoa(i) + "]")
			fieldsTree = append(fieldsTree, item)

			s.getValue(exploded, index+1, fieldsTree, values, path)

			fieldsTree = fieldsTree[:len(fieldsTree)-1]
			path = path[:len(path)-1]
		}
	} else {
		// not last element - without nodes
		if len(path) > 0 {
			path[len(path)-1] = strings.Trim(path[len(path)-1], "[]")
		} else {
			path.add(fieldName)
		}

		for j := index; j < len(exploded); j++ {
			if parent.Name == "" {
				parent.Name += exploded[j]
			} else {
				parent.Name += "." + exploded[j]
			}
		}

		parentRules := s.GetRule(path)
		if parentRules.Has("required") {
			*values = append(*values, parent)
		}
	}

}

func (s *SchemaValidator) Validate() error {
	data := FieldsArray{s.requestBody}
	values := &[]FieldSchema{}

	for _, rule := range s.rules {
		s.getValue(rule.Path, 0, data, values, []string{})
	}

	for _, field := range *values {
		switch field.Value.(type) {
		case bool:
			err := s.validator.Var(field.Value, field.Rules.ForBool().String())
			s.errors.try(field.Name, err)
		default:
			err := s.validator.Var(field.Value, field.Rules.String())
			s.errors.try(field.Name, err)

			if field.Rule.Pattern != nil && field.Value != nil {
				switch field.Value.(type) {
				case string:
					if field.Value.(string) == "" {
						break
					}
					err := s.validatePattern(field.Name, *field.Rule.Pattern, field.Value.(string))
					if err != nil {
						s.errors[field.Name] = append(s.errors[field.Name], *err)
					}
				}
			}
		}
	}

	// TODO: sort errors by fieldName

	if len(s.errors) > 0 {
		return s.errors
	}

	return nil
}

func (s *SchemaValidator) validatePattern(fieldName, pattern, value string) *FieldError {
	re := regexp.MustCompile(pattern)
	isValid := re.MatchString(value)

	if !isValid {
		// return error
		return &FieldError{
			Field:            fieldName,
			Rule:             "regexp",
			Value:            value,
			Accepted:         pattern,
			ValidationErrors: nil,
		}
	}

	return nil
}
