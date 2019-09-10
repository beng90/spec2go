package validators

import (
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

func AddOfferValidate(v *validator.Validate, req *http.Request) ValidationErrors {
	schemaValidator := NewSchemaValidator(v, req)
	schemaValidator.validate("additionalInfo[].id", "string")
	schemaValidator.validate("additionalInfo[].valuesIds[]", "string")
	schemaValidator.validate("brand", "string")
	schemaValidator.validate("categoryId", "required,string,max=16")
	schemaValidator.validate("defaultLanguage", "string,min=2,max=2")
	schemaValidator.validate("productName", "required,string,min=1,max=255")
	schemaValidator.validate("variants", "required")
	schemaValidator.validate("variants[].content", "required")
	schemaValidator.validate("variants[].content[].description", "required,string,min=1,max=1024")
	schemaValidator.validate("variants[].content[].language", "required,string,min=2,max=2")
	schemaValidator.validate("variants[].delivery", "required")
	schemaValidator.validate("variants[].delivery.additionalInfo", "string")
	schemaValidator.validate("variants[].delivery.dispatchTime", "required,numeric,min=1,max=64")
	schemaValidator.validate("variants[].delivery.shippingTemplateId", "required,string,uuid")
	schemaValidator.validate("variants[].ean", "string,min=13,max=13")
	schemaValidator.validate("variants[].inventory", "required")
	schemaValidator.validate("variants[].inventory.size", "required,numeric,min=1,max=4294967295")
	schemaValidator.validate("variants[].inventory.sold", "required,numeric,min=1,max=4294967295")
	schemaValidator.validate("variants[].isEnabled", "required,boolean")
	schemaValidator.validate("variants[].media", "required")
	schemaValidator.validate("variants[].media.images", "required")
	schemaValidator.validate("variants[].media.images[].sortOrder", "numeric,min=1,max=64")
	schemaValidator.validate("variants[].media.images[].url", "required,string,url,max=255")
	schemaValidator.validate("variants[].price", "required,string")
	schemaValidator.validate("variants[].sku", "string")
	schemaValidator.validate("variants[].tags[].id", "string")
	schemaValidator.validate("variants[].tags[].valueId", "string")

	return schemaValidator.errors
}
