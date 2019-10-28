package validators

import (
	"github.com/beng90/spec2go/validate"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

func AddOfferValidate(v *validator.Validate, req *http.Request) error {
	schemaValidator, err := validate.NewSchemaValidator(v, req)
	if err != nil {
		return err
	}

	schemaValidator.AddRule("additionalInfo[].id", "required,string", nil)
	schemaValidator.AddRule("additionalInfo[].valuesIds", "required,min=1", nil)
	schemaValidator.AddRule("additionalInfo[].valuesIds[]", "string", nil)
	schemaValidator.AddRule("brand", "omitempty,string", nil)
	schemaValidator.AddRule("categoryId", "required,string,max=16", validate.Pattern(`^\d+_\d+$`))
	schemaValidator.AddRule("defaultLanguage", "required,string,min=2,max=2", validate.Pattern(`^[a-zA-Z]{2}$`))
	schemaValidator.AddRule("productName", "required,string,min=1,max=255", nil)
	schemaValidator.AddRule("variants", "required,min=1,max=1", nil)
	schemaValidator.AddRule("variants[].content[]", "required", nil)
	schemaValidator.AddRule("variants[].content[].description", "required,string,min=1,max=1024", nil)
	schemaValidator.AddRule("variants[].content[].language", "required,string,min=2,max=2", nil)
	schemaValidator.AddRule("variants[].delivery", "required", nil)
	schemaValidator.AddRule("variants[].delivery.additionalInfo", "string", nil)
	schemaValidator.AddRule("variants[].delivery.dispatchTime", "required,integer,min=1,max=64", nil)
	schemaValidator.AddRule("variants[].delivery.shippingTemplateId", "required,string,uuid", nil)
	schemaValidator.AddRule("variants[].ean", "omitempty,string,min=13,max=13", nil)
	schemaValidator.AddRule("variants[].inventory", "required,object", nil)
	schemaValidator.AddRule("variants[].inventory.size", "required,integer,min=1,max=4294967295", nil)
	schemaValidator.AddRule("variants[].isEnabled", "required,boolean", nil)
	schemaValidator.AddRule("variants[].media", "required", nil)
	schemaValidator.AddRule("variants[].media[].images", "required", nil)
	schemaValidator.AddRule("variants[].media[].images[].sortOrder", "omitempty,integer,min=1,max=64", nil)
	schemaValidator.AddRule("variants[].media[].images[].url", "required,string,url,max=255", nil)
	schemaValidator.AddRule("variants[].price", "required,string", nil)
	schemaValidator.AddRule("variants[].sku", "omitempty,string", nil)
	schemaValidator.AddRule("variants[].tags[]", "omitempty", nil)
	schemaValidator.AddRule("variants[].tags[].id", "required,string", nil)
	schemaValidator.AddRule("variants[].tags[].valueId", "required,string", nil)

	err = schemaValidator.Validate()

	return err
}
