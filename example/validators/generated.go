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

	//schemaValidator.AddRule("additionalInfo[]", "omitempty")
	//schemaValidator.AddRule("additionalInfo[].id", "required,string")
	//schemaValidator.AddRule("additionalInfo[].valuesIds[]", "required,min=1,string")
	//schemaValidator.AddRule("brand", "omitempty,string")
	//schemaValidator.AddRule("categoryId", "required,string,max=16")
	//schemaValidator.AddRule("defaultLanguage", "omitempty,string,min=2,max=2")
	//schemaValidator.AddRule("productName", "required,string,min=1,max=255")
	schemaValidator.AddRule("variants[]", "required")
	schemaValidator.AddRule("variants[].content[]", "required")
	schemaValidator.AddRule("variants[].content[].description", "required,string,min=1,max=1024")
	schemaValidator.AddRule("variants[].content[].language", "required,string,min=2,max=2")
	//schemaValidator.AddRule("variants[].delivery", "required")
	//schemaValidator.AddRule("variants[].delivery.additionalInfo", "omitempty,string")
	//schemaValidator.AddRule("variants[].delivery.dispatchTime", "required,integer,min=1,max=64")
	//schemaValidator.AddRule("variants[].delivery.shippingTemplateId", "required,string,uuid")
	//schemaValidator.AddRule("variants[].ean", "omitempty,string,min=13,max=13")
	//schemaValidator.AddRule("variants[].inventory", "required")
	//schemaValidator.AddRule("variants[].inventory.size", "required,integer,min=1,max=4294967295")
	//schemaValidator.AddRule("variants[].isEnabled", "required,boolean")
	schemaValidator.AddRule("variants[].media[]", "required")
	schemaValidator.AddRule("variants[].media[].images", "required,min=2")
	schemaValidator.AddRule("variants[].media[].images[]", "required")
	//schemaValidator.AddRule("variants[].media[].images[].sortOrder", "omitempty,integer,min=1,max=64")
	schemaValidator.AddRule("variants[].media[].images[].url", "required,string,url,max=255")
	//schemaValidator.AddRule("variants[].price", "required,string")
	//schemaValidator.AddRule("variants[].sku", "omitempty,string")
	schemaValidator.AddRule("variants[].tags[]", "omitempty")
	schemaValidator.AddRule("variants[].tags[].id", "required,string")
	schemaValidator.AddRule("variants[].tags[].valueId", "required,string")

	err = schemaValidator.Validate()

	return err
}
