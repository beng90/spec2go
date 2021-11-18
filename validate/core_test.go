package validate_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/beng90/spec2go/validate"
	"github.com/beng90/spec2go/validate/validations"
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
		_, err := validate.NewSchemaValidator(v, req, context.Background())

		assert.Equal(t, validate.ErrInvalidJSON, err)
	}
}

func getSchemaValidator(requestBody string) *validate.SchemaValidator {
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))
	v := NewValidator()
	schemaValidator, _ := validate.NewSchemaValidator(v, req, context.Background())

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
	rules      string
	pattern    string
	input      string
	want       error
	errorField string
}

func (i Input) Test(t *testing.T, err error) error {
	// fmt.Printf("err %#v\n", err)

	if i.want != nil && len(i.want.(validate.ValidationErrors)) > 0 {
		expected := i.want.(validate.ValidationErrors)

		switch err.(type) {
		case validate.ValidationErrors:
			if x := err.(validate.ValidationErrors)[i.errorField]; x == nil {
				return errors.New(fmt.Sprintf(`Wrong testing rule. Field "%s", rules "%s", value "%v".`, i.errorField, i.rules, i.input))
			}

			fieldErr := err.(validate.ValidationErrors)[i.errorField][0]

			if expectedError := expected[i.errorField]; expectedError == nil {
				return errors.New(fmt.Sprintf(`want error does not exist. Field "%s", rules "%s", value "%v".`, i.errorField, i.rules, i.input))
			}

			assert.Equal(t, expected[i.errorField][0].Field, fieldErr.Field)
			assert.Equal(t, expected[i.errorField][0].Rule, fieldErr.Rule)
			assert.Equal(t, expected[i.errorField][0].Value, fieldErr.Value)
			assert.Equal(t, expected[i.errorField][0].Accepted, fieldErr.Accepted)
		default:
			t.Errorf("error is %T, not validate.ValidationErrors\n", err)
		}
	} else {
		assert.Equal(t, i.want, err)
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
		rules:      "required,integer,min=1,max=999",
		pattern:    "",
		input:      `{"categoryId": 123}`,
		want:       nil,
		errorField: "categoryId",
	}

	schemaValidator := getSchemaValidator(input.input)
	schemaValidator.AddRule(input.errorField, input.rules, &input.pattern)
	err := schemaValidator.Validate()

	assert.Equal(t, nil, err)
}

func TestSchemaValidator_Validate_String(t *testing.T) {
	fieldName := "categoryId"

	testData := []Input{
		{
			rules:      "required,string,max=5",
			input:      "{}",
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      fmt.Sprintf(`{"%s": 123}`, fieldName),
			errorField: fieldName,
			want:       getExpectedError(fieldName, "string", float64(123), ""),
		},
		{
			rules:      "required,string,max=5",
			input:      fmt.Sprintf(`{"%s": "123456"}`, fieldName),
			errorField: fieldName,
			want:       getExpectedError(fieldName, "max", "123456", "5"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_Integer(t *testing.T) {
	fieldName := "categoryId"

	testData := []Input{
		{
			rules:      "required,integer,min=1,max=5",
			input:      "{}",
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,integer,min=1,max=5",
			input:      fmt.Sprintf(`{"%s": 0}`, fieldName),
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", float64(0), ""),
		},
		{
			rules:      "notblank,integer,min=1,max=5",
			input:      "{}",
			errorField: fieldName,
			want:       getExpectedError(fieldName, "notblank", nil, ""),
		},
		{
			rules:      "notblank,min=1",
			input:      fmt.Sprintf(`{"%s": 0}`, fieldName),
			errorField: fieldName,
			want:       getExpectedError(fieldName, "min", float64(0), "1"),
		},
		{
			rules:      "notblank,integer,min=1,max=5",
			input:      fmt.Sprintf(`{"%s": "123"}`, fieldName),
			errorField: fieldName,
			want:       getExpectedError(fieldName, "integer", "123", ""),
		},
		{
			rules:      "notblank,integer,min=2,max=12345",
			input:      fmt.Sprintf(`{"%s": 1}`, fieldName),
			errorField: fieldName,
			want:       getExpectedError(fieldName, "min", float64(1), "2"),
		},
		{
			rules:      "notblank,integer,min=1,max=12345",
			input:      fmt.Sprintf(`{"%s": 123456}`, fieldName),
			errorField: fieldName,
			want:       getExpectedError(fieldName, "max", float64(123456), "12345"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_Boolean(t *testing.T) {
	fieldName := "isEnabled"

	testData := []Input{
		{
			rules:      "required,boolean",
			input:      "{}",
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,boolean",
			input:      `{"isEnabled": null}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,boolean",
			input:      `{"isEnabled": ""}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", "", ""),
		},
		{
			rules:      "required,boolean",
			input:      `{"isEnabled": 0}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", float64(0), ""),
		},
		{
			rules:      "notblank,boolean",
			input:      `{"isEnabled": false}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,boolean",
			input:      `{"isEnabled": true}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "boolean",
			input:      `{"isEnabled": 0}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "boolean", float64(0), ""),
		},
		{
			rules:      "boolean",
			input:      `{"isEnabled": ""}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "boolean", "", ""),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_Pattern(t *testing.T) {
	fieldName := "countryCode"
	pattern := `^[a-zA-Z]{2}$`

	testData := []Input{
		{
			rules:      "required,string",
			pattern:    pattern,
			input:      `{}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ``),
		},
		{
			rules:      "required,string",
			pattern:    pattern,
			input:      `{"countryCode": null}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ``),
		},
		{
			rules:      "required,string",
			pattern:    pattern,
			input:      `{"countryCode": ""}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", "", ``),
		},
		{
			rules:      "required,string",
			pattern:    pattern,
			input:      `{"countryCode": "pln"}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "regexp", "pln", pattern),
		},
		{
			rules:      "required,string",
			pattern:    pattern,
			input:      `{"countryCode": "USA"}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "regexp", "USA", pattern),
		},
		{
			rules:      "required,string",
			pattern:    pattern,
			input:      `{"countryCode": "pl"}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string",
			pattern:    pattern,
			input:      `{"countryCode": "US"}`,
			errorField: fieldName,
			want:       nil,
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, &tt.pattern)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_ObjectItem(t *testing.T) {
	fieldName := "category.id"

	testData := []Input{
		{
			rules:      "required,string,max=5",
			input:      `{"category": {}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"category": {"id": 123}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "string", float64(123), ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"category": {"id": "123456"}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "max", "123456", "5"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_ArrayField(t *testing.T) {
	fieldName := "categories[].id"

	testData := []Input{
		{
			rules:      "required,string,max=5",
			input:      `{"categories": []}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"categories": [{}]}`,
			errorField: "categories[0].id",
			want:       getExpectedError("categories[0].id", "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"categories": [{"id": 123456}]}`,
			errorField: "categories[0].id",
			want:       getExpectedError("categories[0].id", "string", float64(123456), ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"categories": [{"id": "123456"}]}`,
			errorField: "categories[0].id",
			want:       getExpectedError("categories[0].id", "max", "123456", "5"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_NestedArray(t *testing.T) {
	fieldName := "product.categories[]"

	testData := []Input{
		{
			rules:      "required,string,max=5",
			input:      `{}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": null}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": {}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": null}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": true}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": []}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": [ 123 ]}}`,
			errorField: "product.categories[0]",
			want:       getExpectedError("product.categories[0]", "string", float64(123), ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": [ "asd", 123 ]}}`,
			errorField: "product.categories[1]",
			want:       getExpectedError("product.categories[1]", "string", float64(123), ""),
		},
		{
			rules:      "required,string,min=2,max=5",
			input:      `{"product": { "categories": [ "1" ]}}`,
			errorField: "product.categories[0]",
			want:       getExpectedError("product.categories[0]", "min", "1", "2"),
		},
		{
			rules:      "required,string,min=2,max=5",
			input:      `{"product": { "categories": [ "123456" ]}}`,
			errorField: "product.categories[0]",
			want:       getExpectedError("product.categories[0]", "max", "123456", "5"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_NestedArrayWithField(t *testing.T) {
	fieldName := "product.categories[].id"
	pattern := `^\d+$`

	testData := []Input{
		{
			rules:      "required,string,max=5",
			input:      `{}`,
			pattern:    pattern,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": null}`,
			pattern:    pattern,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": {}}`,
			pattern:    pattern,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": null}}`,
			pattern:    pattern,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": [] }}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": true }}`,
			pattern:    pattern,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "someField": "123" }}`,
			pattern:    pattern,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": {} }}`,
			pattern:    pattern,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": [{}]}}`,
			errorField: "product.categories[0].id",
			want:       getExpectedError("product.categories[0].id", "required", nil, ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": [ { "id": 123 } ]}}`,
			pattern:    pattern,
			errorField: "product.categories[0].id",
			want:       getExpectedError("product.categories[0].id", "string", float64(123), ""),
		},
		{
			rules:      "required,string,max=5",
			input:      `{"product": { "categories": [ {"id": "asd"}, {"id": 123} ]}}`,
			pattern:    pattern,
			errorField: "product.categories[1].id",
			want:       getExpectedError("product.categories[1].id", "string", float64(123), ""),
		},
		{
			rules:      "required,string,min=2,max=5",
			input:      `{"product": { "categories": [ {"id": "1"} ]}}`,
			pattern:    pattern,
			errorField: "product.categories[0].id",
			want:       getExpectedError("product.categories[0].id", "min", "1", "2"),
		},
		{
			rules:      "required,string,min=2,max=5",
			input:      `{"product": { "categories": [ {"id": "123456"} ]}}`,
			pattern:    pattern,
			errorField: "product.categories[0].id",
			want:       getExpectedError("product.categories[0].id", "max", "123456", "5"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule(fieldName, tt.rules, &tt.pattern)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_NestedArray_ParentRequired(t *testing.T) {
	fieldName := "product.categories[]"
	rulesString := "required,string,min=2,max=5"

	testData := []Input{
		{
			rules:      rulesString,
			input:      `{}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": null}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": true}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "string", true, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": {}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": null}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": true}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": []}}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ 123 ]}}`,
			errorField: "product.categories[0]",
			want:       getExpectedError("product.categories[0]", "string", float64(123), ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ "asd", 123 ]}}`,
			errorField: "product.categories[1]",
			want:       getExpectedError("product.categories[1]", "string", float64(123), ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ "1" ]}}`,
			errorField: "product.categories[0]",
			want:       getExpectedError("product.categories[0]", "min", "1", "2"),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ "123456" ]}}`,
			errorField: "product.categories[0]",
			want:       getExpectedError("product.categories[0]", "max", "123456", "5"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule("product", "required", nil)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSchemaValidator_Validate_NestedArrayField_ParentRequired(t *testing.T) {
	fieldName := "product.categories[].id"
	rulesString := "required,string,min=2,max=5"

	tests := []Input{
		{
			rules:      rulesString,
			input:      `{}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": null}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "required", nil, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": true}`,
			errorField: fieldName,
			want:       getExpectedError(fieldName, "string", true, ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": {}}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": null}}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": true}}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": []}}`,
			errorField: fieldName,
			want:       nil,
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ {"id": 123} ]}}`,
			errorField: "product.categories[0].id",
			want:       getExpectedError("product.categories[0].id", "string", float64(123), ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ {"id": "asd"}, {"id": 123} ]}}`,
			errorField: "product.categories[1].id",
			want:       getExpectedError("product.categories[1].id", "string", float64(123), ""),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ {"id": "1"} ]}}`,
			errorField: "product.categories[0].id",
			want:       getExpectedError("product.categories[0].id", "min", "1", "2"),
		},
		{
			rules:      rulesString,
			input:      `{"product": { "categories": [ {"id": "123456"} ]}}`,
			errorField: "product.categories[0].id",
			want:       getExpectedError("product.categories[0].id", "max", "123456", "5"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			schemaValidator := getSchemaValidator(tt.input)
			schemaValidator.AddRule("product", "required", nil)
			schemaValidator.AddRule(fieldName, tt.rules, nil)
			err := schemaValidator.Validate()

			if err := tt.Test(t, err); err != nil {
				t.Error(err)
			}
		})
	}
}
