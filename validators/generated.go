package validators

import (
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)


func AddOfferValidate(v *validator.Validate, req *http.Request) ValidationErrors {
	schemaValidator := NewSchemaValidator(v, req)
    //schemaValidator.validate("additionalInfo", "omitempty")
    schemaValidator.validate("additionalInfo[].id", "omitempty,string")
    //schemaValidator.validate("additionalInfo[].valuesIds", "omitempty")
    //schemaValidator.validate("additionalInfo[].valuesIds[]", "omitempty,string")
    //schemaValidator.validate("brand", "omitempty,string")
    //schemaValidator.validate("categoryId", "required,string,max=16")
    //schemaValidator.validate("defaultLanguage", "omitempty,string,min=2,max=2")
    //schemaValidator.validate("productName", "required,string,min=1,max=255")
    //schemaValidator.validate("variants", "required")
    //schemaValidator.validate("variants[].content", "required")
    //schemaValidator.validate("variants[].content[].description", "required,string,min=1,max=1024")
    //schemaValidator.validate("variants[].content[].language", "required,string,min=2,max=2")
    //schemaValidator.validate("variants[].delivery", "required")
    //schemaValidator.validate("variants[].delivery.additionalInfo", "omitempty,string")
    //schemaValidator.validate("variants[].delivery.dispatchTime", "required,integer,min=1,max=64")
    //schemaValidator.validate("variants[].delivery.shippingTemplateId", "required,string,uuid")
    //schemaValidator.validate("variants[].ean", "omitempty,string,min=13,max=13")
    //schemaValidator.validate("variants[].inventory", "required")
    //schemaValidator.validate("variants[].inventory.size", "required,integer,min=1,max=4294967295")
    //schemaValidator.validate("variants[].inventory.sold", "required,integer,min=1,max=4294967295")
    //schemaValidator.validate("variants[].isEnabled", "required,boolean")
    //schemaValidator.validate("variants[].media", "required")
    //schemaValidator.validate("variants[].media.images", "required")
    //schemaValidator.validate("variants[].media.images[].sortOrder", "omitempty,integer,min=1,max=64")
    //schemaValidator.validate("variants[].media.images[].url", "required,string,url,max=255")
    //schemaValidator.validate("variants[].price", "required,string")
    //schemaValidator.validate("variants[].sku", "omitempty,string")
    //schemaValidator.validate("variants[].tags", "omitempty")
    //schemaValidator.validate("variants[].tags[].id", "omitempty,string")
    //schemaValidator.validate("variants[].tags[].valueId", "omitempty,string")

	return schemaValidator.Errors()
}

