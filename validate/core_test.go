package validate_test

import (
	"bytes"
	"fmt"
	"github.com/beng90/spec2go/validate"
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
	_ = v.RegisterValidation("string", validate.IsString)

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

		fieldErr := err.(validate.ValidationErrors)[input.ErrorField][0]

		assert.Equal(t, input.Expected[fieldName][0].Field, fieldErr.Field)
		assert.Equal(t, input.Expected[fieldName][0].Rule, fieldErr.Rule)
		assert.Equal(t, input.Expected[fieldName][0].Value, fieldErr.Value)
		assert.Equal(t, input.Expected[fieldName][0].Accepted, fieldErr.Accepted)
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

		//fmt.Printf("err %#v", err)

		if len(input.Expected) > 0 {
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
