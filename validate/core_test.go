package validate

import (
	"bytes"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"gotest.tools/assert"
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
	return v
}

func getSchemaValidator(requestBody string) *SchemaValidator {
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))
	v := NewValidator()

	schemaValidator, err := NewSchemaValidator(v, req)
	if err != nil {
		panic(err)
	}

	schemaValidator.AddRule("additionalInfo", "omitempty")
	schemaValidator.AddRule("additionalInfo[].id", "required,string")
	schemaValidator.AddRule("additionalInfo[].valuesIds", "required")
	schemaValidator.AddRule("additionalInfo[].valuesIds[]", "required,min=1,string")
	schemaValidator.AddRule("brand", "omitempty,string")
	schemaValidator.AddRule("categoryId", "required,string,max=16")
	schemaValidator.AddRule("defaultLanguage", "omitempty,string,min=2,max=2")
	schemaValidator.AddRule("productName", "required,string,min=1,max=255")
	schemaValidator.AddRule("variants", "required")
	schemaValidator.AddRule("variants[].content", "required")
	schemaValidator.AddRule("variants[].content[].description", "required,string,min=1,max=1024")
	schemaValidator.AddRule("variants[].content[].language", "required,string,min=2,max=2")
	schemaValidator.AddRule("variants[].delivery", "required")
	schemaValidator.AddRule("variants[].delivery.additionalInfo", "omitempty,string")
	schemaValidator.AddRule("variants[].delivery.dispatchTime", "required,integer,min=1,max=64")
	schemaValidator.AddRule("variants[].delivery.shippingTemplateId", "required,string,uuid")
	schemaValidator.AddRule("variants[].ean", "omitempty,string,min=13,max=13")
	schemaValidator.AddRule("variants[].inventory", "required")
	schemaValidator.AddRule("variants[].inventory.size", "required,integer,min=1,max=4294967295")
	schemaValidator.AddRule("variants[].isEnabled", "required,boolean")
	schemaValidator.AddRule("variants[].media", "required")
	schemaValidator.AddRule("variants[].media.images", "required")
	schemaValidator.AddRule("variants[].media.images[].sortOrder", "omitempty,integer,min=1,max=64")
	schemaValidator.AddRule("variants[].media.images[].url", "required,string,url,max=255")
	schemaValidator.AddRule("variants[].price", "required,string")
	schemaValidator.AddRule("variants[].sku", "omitempty,string")
	schemaValidator.AddRule("variants[].tags", "omitempty")
	schemaValidator.AddRule("variants[].tags[].id", "omitempty,string")
	schemaValidator.AddRule("variants[].tags[].valueId", "omitempty,string")

	return schemaValidator
}

func TestSchemaValidator_Validate(t *testing.T) {
	requestBody := `{
	}`

	schemaValidator := getSchemaValidator(requestBody)
	err := schemaValidator.Validate()

	vErrors := err.(ValidationErrors)
	fmt.Println("vErrors", vErrors)

	assert.Equal(t, hasError(vErrors, "additionalInfo[]"), false)
	assert.Equal(t, hasError(vErrors, "additionalInfo[].id"), false)
	assert.Equal(t, hasError(vErrors, "additionalInfo[].valuesIds"), false)
	assert.Equal(t, hasError(vErrors, "additionalInfo[].valuesIds[]"), false)
	assert.Equal(t, hasError(vErrors, "brand"), false)
	assert.Equal(t, hasError(vErrors, "categoryId"), true)
	assert.Equal(t, hasError(vErrors, "defaultLanguage"), false)
	assert.Equal(t, hasError(vErrors, "productName"), true)
	assert.Equal(t, hasError(vErrors, "variants"), true)
	assert.Equal(t, hasError(vErrors, "variants[].content"), false)
	assert.Equal(t, hasError(vErrors, "variants[].content[].description"), false)
	assert.Equal(t, hasError(vErrors, "variants[].content[].language"), false)
	assert.Equal(t, hasError(vErrors, "variants[].delivery"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].delivery.additionalInfo"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].delivery.dispatchTime"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].delivery.shippingTemplateId"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].ean"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].inventory"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].inventory.size"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].isEnabled"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].media"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].media.images"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].media.images[].sortOrder"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].media.images[].url"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].price"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].sku"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].tags"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].tags[].id"), false)
	//assert.Equal(t, hasError(vErrors, "variants[].tags[].valueId"), false)
}

func hasError(vErrors ValidationErrors, fieldPath string) bool {
	_, hasField := vErrors[fieldPath]
	//fmt.Println(fieldPath, hasField)

	return hasField
}
