package validate_test

import (
	"bytes"
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
		_, err := validate.NewSchemaValidator(v, req)

		assert.Equal(t, validate.ErrInvalidJSON, err)
	}
}

func getSchemaValidator(requestBody string) *validate.SchemaValidator {
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))
	v := NewValidator()
	schemaValidator, _ := validate.NewSchemaValidator(v, req)

	return schemaValidator
}

func TestSchemaValidator_AddGetHasRule(t *testing.T) {
	schemaValidator := getSchemaValidator(`{}`)

	ruleString := "required,string,max=16"
	fieldPathString := "media[].images[].url"
	schemaValidator.AddRule(fieldPathString, ruleString)
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
	Input      string
	Expected   validate.ValidationErrors
	ErrorField string
}

func TestSchemaValidator_Validate_String(t *testing.T) {
	fieldName := "categoryId"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      "{}",
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      fmt.Sprintf(`{"%s": 123}`, fieldName),
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "string",
				Value:            float64(123),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      fmt.Sprintf(`{"%s": "123456"}`, fieldName),
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "max",
				Value:            "123456",
				Accepted:         "5",
				ValidationErrors: nil,
			}}},
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules)
		err := schemaValidator.Validate()

		fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

		assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
		assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
		assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
		assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
	}
}

func TestSchemaValidator_Validate_Integer(t *testing.T) {
	fieldName := "categoryId"

	testData := []Input{
		{
			Rules:      "required,integer,min=1,max=5",
			Input:      "{}",
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,integer,min=1,max=5",
			Input:      fmt.Sprintf(`{"%s": 0}`, fieldName),
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            float64(0),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "notblank,integer,min=1,max=5",
			Input:      "{}",
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "notblank",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "notblank,min=1",
			Input:      fmt.Sprintf(`{"%s": 0}`, fieldName),
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "min",
				Value:            float64(0),
				Accepted:         "1",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "notblank,integer,min=1,max=5",
			Input:      fmt.Sprintf(`{"%s": "123"}`, fieldName),
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "integer",
				Value:            "123",
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "notblank,integer,min=2,max=12345",
			Input:      fmt.Sprintf(`{"%s": 1}`, fieldName),
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "min",
				Value:            float64(1),
				Accepted:         "2",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "notblank,integer,min=1,max=12345",
			Input:      fmt.Sprintf(`{"%s": 123456}`, fieldName),
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "max",
				Value:            float64(123456),
				Accepted:         "12345",
				ValidationErrors: nil,
			}}},
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules)
		err := schemaValidator.Validate()

		if x := err.(validate.ValidationErrors)[input.ErrorField]; x == nil {
			t.Errorf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, input.ErrorField, input.Rules, input.Input)

			return
		}

		fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

		assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
		assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
		assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
		assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
	}
}

func TestSchemaValidator_Validate_Boolean(t *testing.T) {
	fieldName := "isEnabled"

	testData := []Input{
		{
			Rules:      "required,boolean",
			Input:      "{}",
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": null}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": ""}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            "",
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": 0}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            float64(0),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "notblank,boolean",
			Input:      `{"isEnabled": false}`,
			ErrorField: fieldName,
			Expected:   validate.ValidationErrors{},
		},
		{
			Rules:      "required,boolean",
			Input:      `{"isEnabled": true}`,
			ErrorField: fieldName,
			Expected:   validate.ValidationErrors{},
		},
		{
			Rules:      "boolean",
			Input:      `{"isEnabled": 0}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "boolean",
				Value:            float64(0),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "boolean",
			Input:      `{"isEnabled": ""}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "boolean",
				Value:            "",
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules)
		err := schemaValidator.Validate()

		if len(input.Expected) > 0 {
			if x := err.(validate.ValidationErrors)[input.ErrorField]; x == nil {
				t.Errorf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, input.ErrorField, input.Rules, input.Input)

				return
			}

			fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

			assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
			assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
			assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
			assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
		} else {
			assert.Equal(t, input.Expected, err)
		}
	}
}

func TestSchemaValidator_Validate_Object(t *testing.T) {
	fieldName := "category.id"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      `{"category": {}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"category": {"id": 123}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "string",
				Value:            float64(123),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"category": {"id": "123456"}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "max",
				Value:            "123456",
				Accepted:         "5",
				ValidationErrors: nil,
			}}},
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules)
		err := schemaValidator.Validate()

		if len(input.Expected) > 0 {
			if x := err.(validate.ValidationErrors)[input.ErrorField]; x == nil {
				t.Errorf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, input.ErrorField, input.Rules, input.Input)

				return
			}

			fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

			assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
			assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
			assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
			assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
		} else {
			assert.Equal(t, input.Expected, err)
		}
	}
}

func TestSchemaValidator_Validate_Array(t *testing.T) {
	fieldName := "categories[].id"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": []}`,
			ErrorField: "categories[].id",
			Expected:   validate.ValidationErrors{},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": [{}]}`,
			ErrorField: "categories[0].id",
			Expected: validate.ValidationErrors{"categories[].id": []validate.FieldError{{ // TODO: wrong field name
				Field:            "categories[0].id",
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": [{"id": 123456}]}`,
			ErrorField: "categories[0].id",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            "categories[0].id",
				Rule:             "string",
				Value:            float64(123456),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"categories": [{"id": "123456"}]}`,
			ErrorField: "categories[0].id",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            "categories[0].id",
				Rule:             "max",
				Value:            "123456",
				Accepted:         "5",
				ValidationErrors: nil,
			}}},
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules)
		err := schemaValidator.Validate()

		if len(input.Expected) > 0 {
			if x := err.(validate.ValidationErrors)[input.ErrorField]; x == nil {
				t.Errorf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, input.ErrorField, input.Rules, input.Input)

				return
			}

			fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

			assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
			assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
			assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
			assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
		} else {
			assert.Equal(t, input.Expected, err)
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
			Expected:   validate.ValidationErrors{},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": null}`,
			ErrorField: fieldName,
			Expected:   validate.ValidationErrors{},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": {}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": null}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": true}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": []}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ 123 ]}}`,
			ErrorField: "product.categories[0]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            "product.categories[0]",
				Rule:             "string",
				Value:            float64(123),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ "asd", 123 ]}}`,
			ErrorField: "product.categories[1]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            "product.categories[1]",
				Rule:             "string",
				Value:            float64(123),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ "1" ]}}`,
			ErrorField: "product.categories[0]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            "product.categories[0]",
				Rule:             "min",
				Value:            "1",
				Accepted:         "2",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ "123456" ]}}`,
			ErrorField: "product.categories[0]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{ // TODO: wrong field name
				Field:            "product.categories[0]",
				Rule:             "max",
				Value:            "123456",
				Accepted:         "5",
				ValidationErrors: nil,
			}}},
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule(fieldName, input.Rules)
		err := schemaValidator.Validate()

		//fmt.Printf("err %#v", err)

		if len(input.Expected) > 0 {
			if x := err.(validate.ValidationErrors)[input.ErrorField]; x == nil {
				t.Errorf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, input.ErrorField, input.Rules, input.Input)

				return
			}

			fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

			assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
			assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
			assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
			assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
		} else {
			assert.Equal(t, input.Expected, err)
		}
	}
}

func TestSchemaValidator_Validate_NestedArray_ParentRequired(t *testing.T) {
	fieldName := "product.categories[]"

	testData := []Input{
		{
			Rules:      "required,string,max=5",
			Input:      `{}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{
				fieldName: []validate.FieldError{{
					Field:            fieldName,
					Rule:             "required",
					Value:            nil,
					Accepted:         "",
					ValidationErrors: nil,
				}},
			},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": null}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{
				fieldName: []validate.FieldError{{
					Field:            fieldName,
					Rule:             "required",
					Value:            nil,
					Accepted:         "",
					ValidationErrors: nil,
				}},
			},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": {}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": null}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": true}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": []}}`,
			ErrorField: fieldName,
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            fieldName,
				Rule:             "required",
				Value:            nil,
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ 123 ]}}`,
			ErrorField: "product.categories[0]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            "product.categories[0]",
				Rule:             "string",
				Value:            float64(123),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,max=5",
			Input:      `{"product": { "categories": [ "asd", 123 ]}}`,
			ErrorField: "product.categories[1]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            "product.categories[1]",
				Rule:             "string",
				Value:            float64(123),
				Accepted:         "",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ "1" ]}}`,
			ErrorField: "product.categories[0]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            "product.categories[0]",
				Rule:             "min",
				Value:            "1",
				Accepted:         "2",
				ValidationErrors: nil,
			}}},
		},
		{
			Rules:      "required,string,min=2,max=5",
			Input:      `{"product": { "categories": [ "123456" ]}}`,
			ErrorField: "product.categories[0]",
			Expected: validate.ValidationErrors{fieldName: []validate.FieldError{{
				Field:            "product.categories[0]",
				Rule:             "max",
				Value:            "123456",
				Accepted:         "5",
				ValidationErrors: nil,
			}}},
		},
	}

	for _, input := range testData {
		schemaValidator := getSchemaValidator(input.Input)
		schemaValidator.AddRule("product", "required")
		schemaValidator.AddRule(fieldName, input.Rules)
		err := schemaValidator.Validate()

		//fmt.Printf("err %#v\n", err)

		if len(input.Expected) > 0 {
			if x := err.(validate.ValidationErrors)[input.ErrorField]; x == nil {
				t.Errorf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, input.ErrorField, input.Rules, input.Input)

				return
			}

			fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

			assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
			assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
			assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
			assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
		} else {
			assert.Equal(t, input.Expected, err)
		}
	}
}
