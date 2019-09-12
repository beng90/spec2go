package validators

import (
	"github.com/beng90/spec2go/validate"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

func AddOfferValidate(v *validator.Validate, req *http.Request) validate.ValidationErrors {
	schemaValidator := validate.NewSchemaValidator(v, req)
	schemaValidator.Validate("additionalInfo", "omitempty")
	schemaValidator.Validate("additionalInfo[].id", "omitempty,string")
	schemaValidator.Validate("additionalInfo[].valuesIds", "omitempty")
	schemaValidator.Validate("additionalInfo[].valuesIds[]", "omitempty,string")
	schemaValidator.Validate("brand", "omitempty,string")
	schemaValidator.Validate("categoryId", "required,string,max=16")
	schemaValidator.Validate("defaultLanguage", "omitempty,string,min=2,max=2")
	schemaValidator.Validate("productName", "required,string,min=1,max=255")
	schemaValidator.Validate("variants", "required")
	schemaValidator.Validate("variants[].content", "required")
	schemaValidator.Validate("variants[].content[].description", "required,string,min=1,max=1024")
	schemaValidator.Validate("variants[].content[].language", "required,string,min=2,max=2")
	schemaValidator.Validate("variants[].delivery", "required")
	schemaValidator.Validate("variants[].delivery.additionalInfo", "omitempty,string")
	schemaValidator.Validate("variants[].delivery.dispatchTime", "required,integer,min=1,max=64")
	schemaValidator.Validate("variants[].delivery.shippingTemplateId", "required,string,uuid")
	schemaValidator.Validate("variants[].ean", "omitempty,string,min=13,max=13")
	schemaValidator.Validate("variants[].inventory", "required")
	schemaValidator.Validate("variants[].inventory.size", "required,integer,min=1,max=4294967295")
	schemaValidator.Validate("variants[].isEnabled", "required,boolean")
	schemaValidator.Validate("variants[].media", "required")
	schemaValidator.Validate("variants[].media.images", "required")
	schemaValidator.Validate("variants[].media.images[].sortOrder", "omitempty,integer,min=1,max=64")
	schemaValidator.Validate("variants[].media.images[].url", "required,string,url,max=255")
	schemaValidator.Validate("variants[].price", "required,string")
	schemaValidator.Validate("variants[].sku", "omitempty,string")
	schemaValidator.Validate("variants[].tags", "omitempty")
	schemaValidator.Validate("variants[].tags[].id", "omitempty,string")
	schemaValidator.Validate("variants[].tags[].valueId", "omitempty,string")

	return schemaValidator.Errors()
}
