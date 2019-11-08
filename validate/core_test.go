package validate_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/beng90/spec2go/validate"
	"github.com/beng90/spec2go/validate/validations"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func NewValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// custom validations
	_ = v.RegisterValidation("string", validations.IsString)
	_ = v.RegisterValidation("notblank", validations.NotBlank)
	_ = v.RegisterValidation("boolean", validations.IsBoolean)

	return v
}

func TestNewSchemaValidator_InvalidJSON(t *testing.T) {
	testData := []string{
		``,
		`<xml>`,
	}

	for _, data := range testData {
		req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(data)))
		v := NewValidator()
		_, err := validate.NewSchemaValidator(v, req, nil)

		assert.Equal(t, validate.ErrInvalidJSON, err)
	}
}

func getSchemaValidator(requestBody string) *validate.SchemaValidator {
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))
	v := NewValidator()
	schemaValidator, _ := validate.NewSchemaValidator(v, req, nil)

	return schemaValidator
}

func TestSchemaValidator_AddGetHasRule(t *testing.T) {
	schemaValidator := getSchemaValidator(`{}`)

	ruleString := "required,string,max=16"
	fieldPathString := "media[].images[].url"
	schemaValidator.AddRule(fieldPathString, ruleString, nil)
	fieldPath := validate.FieldPath{"media[]", "images[]", "url"}

	rule := schemaValidator.GetRule(fieldPath)

	assert.Equal(t, true, schemaValidator.HasRule(fieldPath))

	expected := &validate.Rule{
		Path:  fieldPath,
		Rules: validate.Rules{"required", "string", "max=16"},
	}

	assert.Equal(t, expected, rule)
}

type Input struct {
	Rules      string
	Pattern    string
	Input      string
	Expected   error
	ErrorField string
}

func (i Input) Test(t *testing.T, err error) error {
	//fmt.Printf("err %#v\n", err)

	if i.Expected != nil && len(i.Expected.(validate.ValidationErrors)) > 0 {
		expected := i.Expected.(validate.ValidationErrors)

		switch err.(type) {
		case validate.ValidationErrors:
			if x := err.(validate.ValidationErrors)[i.ErrorField]; x == nil {
				return errors.New(fmt.Sprintf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, i.ErrorField, i.Rules, i.Input))
			}

			fieldErr := err.(validate.ValidationErrors)[i.ErrorField][0]

			if expectedError := expected[i.ErrorField]; expectedError == nil {
				return errors.New(fmt.Sprintf(`Expected error does not exist. Field "%s", rules "%s", value "%v".`, i.ErrorField, i.Rules, i.Input))
			}

			assert.Equal(t, expected[i.ErrorField][0].Field, fieldErr.Field)
			assert.Equal(t, expected[i.ErrorField][0].Rule, fieldErr.Rule)
			assert.Equal(t, expected[i.ErrorField][0].Value, fieldErr.Value)
			assert.Equal(t, expected[i.ErrorField][0].Accepted, fieldErr.Accepted)
		default:
			t.Errorf("error is %T, not validate.ValidationErrors\n", err)
		}
	} else {
		assert.Equal(t, i.Expected, err)
	}

	return nil
}

func getExpectedError(fieldName, rule string, value interface{}, accepted string) validate.ValidationErrors {
	return validate.ValidationErrors{fieldName: []validate.FieldError{{
		Field:            fieldName,
		Rule:             rule,
		Value:            value,
		Accepted:         accepted,
		ValidationErrors: nil,
	}}}
}

func TestNewSchemaValidator_Validate_Passed(t *testing.T) {
	input := Input{
		Rules:      "required,integer,min=1,max=999",
		Pattern:    "",
		Input:      `{"categoryId": 123}`,
		Expected:   nil,
		ErrorField: "categoryId",
	}

	schemaValidator := getSchemaValidator(input.Input)
	schemaValidator.AddRule(input.ErrorField, input.Rules, &input.Pattern)
	err := schemaValidator.Validate()

	assert.Equal(t, err, nil)
}

func TestSchemaValidator_Validate_String(t *testing.T) {
	fieldName := "categoryId"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      "{}",
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      fmt.Sprintf(`{"%s": 123}`, fieldName),
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "string", float64(123), ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      fmt.Sprintf(`{"%s": "123456"}`, fieldName),
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "max", "123456", "5"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_Integer(t *testing.T) {
	fieldName := "categoryId"

	testData := []Input{
		{
			Rules:      "required,integer,min=1,max=5",
			Input:      "{}",
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,integer,min=1,max=5",
			Input:      fmt.Sprintf(`{"%s": 0}`, fieldName),
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", float64(0), ""),
		},
		{
			Rules:      "notblank,integer,min=1,max=5",
			Input:      "{}",
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "notblank", nil, ""),
		},
		{
			Rules:      "notblank,min=1",
			Input:      fmt.Sprintf(`{"%s": 0}`, fieldName),
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "min", float64(0), "1"),
		},
		{
			Rules:      "notblank,integer,min=1,max=5",
			Input:      fmt.Sprintf(`{"%s": "123"}`, fieldName),
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "integer", "123", ""),
		},
		{
			Rules:      "notblank,integer,min=2,max=12345",
			Input:      fmt.Sprintf(`{"%s": 1}`, fieldName),
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "min", float64(1), "2"),
		},
		{
			Rules:      "notblank,integer,min=1,max=12345",
			Input:      fmt.Sprintf(`{"%s": 123456}`, fieldName),
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "max", float64(123456), "12345"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_Boolean(t *testing.T) {
	fieldName := "isEnabled"

	testData := []Input{
		{
			Rules:      "required,boolean",
			Input:      "{}",
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": null}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": ""}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", "", ""),
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": 0}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", float64(0), ""),
		},
		{
			Rules:      "notblank,boolean",
			Input:      `{"isEnabled": false}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": true}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "boolean",
			Input:      `{"isEnabled": 0}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "boolean", float64(0), ""),
		},
		{
			Rules:      "boolean",
			Input:      `{"isEnabled": ""}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "boolean", "", ""),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_Pattern(t *testing.T) {
	fieldName := "countryCode"
	pattern := `^[a-zA-Z]{2}$`

	testData := []Input{
		{
			Rules:      "required,string",
			Pattern:    pattern,
			Input:      `{}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ``),
		},
		{
			Rules:      "required,string",
			Pattern:    pattern,
			Input:      `{"countryCode": null}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ``),
		},
		{
			Rules:      "required,string",
			Pattern:    pattern,
			Input:      `{"countryCode": ""}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", "", ``),
		},
		{
			Rules:      "required,string",
			Pattern:    pattern,
			Input:      `{"countryCode": "pln"}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "regexp", "pln", pattern),
		},
		{
			Rules:      "required,string",
			Pattern:    pattern,
			Input:      `{"countryCode": "USA"}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "regexp", "USA", pattern),
		},
		{
			Rules:      "required,string",
			Pattern:    pattern,
			Input:      `{"countryCode": "pl"}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string",
			Pattern:    pattern,
			Input:      `{"countryCode": "US"}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, &input.Pattern)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_ObjectItem(t *testing.T) {
	fieldName := "category.id"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      `{"category": {}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"category": {"id": 123}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "string", float64(123), ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"category": {"id": "123456"}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "max", "123456", "5"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_ArrayField(t *testing.T) {
	fieldName := "categories[].id"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": []}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": [{}]}`,
			ErrorField: "categories[0].id",
			Expected:   getExpectedError("categories[0].id", "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": [{"id": 123456}]}`,
			ErrorField: "categories[0].id",
			Expected:   getExpectedError("categories[0].id", "string", float64(123456), ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": [{"id": "123456"}]}`,
			ErrorField: "categories[0].id",
			Expected:   getExpectedError("categories[0].id", "max", "123456", "5"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_NestedArray(t *testing.T) {
	fieldName := "product.categories[]"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      `{}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": null}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": {}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": null}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": true}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": []}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ 123 ]}}`,
			ErrorField: "product.categories[0]",
			Expected:   getExpectedError("product.categories[0]", "string", float64(123), ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ "asd", 123 ]}}`,
			ErrorField: "product.categories[1]",
			Expected:   getExpectedError("product.categories[1]", "string", float64(123), ""),
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ "1" ]}}`,
			ErrorField: "product.categories[0]",
			Expected:   getExpectedError("product.categories[0]", "min", "1", "2"),
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ "123456" ]}}`,
			ErrorField: "product.categories[0]",
			Expected:   getExpectedError("product.categories[0]", "max", "123456", "5"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_NestedArrayWithField(t *testing.T) {
	fieldName := "product.categories[].id"
	pattern := `^\d+$`

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      `{}`,
			Pattern:    pattern,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": null}`,
			Pattern:    pattern,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": {}}`,
			Pattern:    pattern,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": null}}`,
			Pattern:    pattern,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [] }}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": true }}`,
			Pattern:    pattern,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "someField": "123" }}`,
			Pattern:    pattern,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": {} }}`,
			Pattern:    pattern,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [{}]}}`,
			ErrorField: "product.categories[0].id",
			Expected:   getExpectedError("product.categories[0].id", "required", nil, ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ { "id": 123 } ]}}`,
			Pattern:    pattern,
			ErrorField: "product.categories[0].id",
			Expected:   getExpectedError("product.categories[0].id", "string", float64(123), ""),
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ {"id": "asd"}, {"id": 123} ]}}`,
			Pattern:    pattern,
			ErrorField: "product.categories[1].id",
			Expected:   getExpectedError("product.categories[1].id", "string", float64(123), ""),
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ {"id": "1"} ]}}`,
			Pattern:    pattern,
			ErrorField: "product.categories[0].id",
			Expected:   getExpectedError("product.categories[0].id", "min", "1", "2"),
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ {"id": "123456"} ]}}`,
			Pattern:    pattern,
			ErrorField: "product.categories[0].id",
			Expected:   getExpectedError("product.categories[0].id", "max", "123456", "5"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules, &input.Pattern)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_NestedArray_ParentRequired(t *testing.T) {
	fieldName := "product.categories[]"
	rulesString := "required,string,min=2,max=5"

	testData := []Input{
		{
			Rules:      rulesString,
			Input:      `{}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": null}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": true}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "string", true, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": {}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": null}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": true}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": []}}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ 123 ]}}`,
			ErrorField: "product.categories[0]",
			Expected:   getExpectedError("product.categories[0]", "string", float64(123), ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ "asd", 123 ]}}`,
			ErrorField: "product.categories[1]",
			Expected:   getExpectedError("product.categories[1]", "string", float64(123), ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ "1" ]}}`,
			ErrorField: "product.categories[0]",
			Expected:   getExpectedError("product.categories[0]", "min", "1", "2"),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ "123456" ]}}`,
			ErrorField: "product.categories[0]",
			Expected:   getExpectedError("product.categories[0]", "max", "123456", "5"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule("product", "required", nil)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}

func TestSchemaValidator_Validate_NestedArrayField_ParentRequired(t *testing.T) {
	fieldName := "product.categories[].id"
	rulesString := "required,string,min=2,max=5"

	testData := []Input{
		{
			Rules:      rulesString,
			Input:      `{}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": null}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "required", nil, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": true}`,
			ErrorField: fieldName,
			Expected:   getExpectedError(fieldName, "string", true, ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": {}}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": null}}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": true}}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": []}}`,
			ErrorField: fieldName,
			Expected:   nil,
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ {"id": 123} ]}}`,
			ErrorField: "product.categories[0].id",
			Expected:   getExpectedError("product.categories[0].id", "string", float64(123), ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ {"id": "asd"}, {"id": 123} ]}}`,
			ErrorField: "product.categories[1].id",
			Expected:   getExpectedError("product.categories[1].id", "string", float64(123), ""),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ {"id": "1"} ]}}`,
			ErrorField: "product.categories[0].id",
			Expected:   getExpectedError("product.categories[0].id", "min", "1", "2"),
		},
		{
			Rules:      rulesString,
			Input:      `{"product": { "categories": [ {"id": "123456"} ]}}`,
			ErrorField: "product.categories[0].id",
			Expected:   getExpectedError("product.categories[0].id", "max", "123456", "5"),
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule("product", "required", nil)
		schemaValidator.AddRule(fieldName, input.Rules, nil)
		err := schemaValidator.Validate()

		if err := input.Test(t, err); err != nil {
			t.Error(err)
		}
	}
}
