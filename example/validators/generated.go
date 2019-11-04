package validators

import (
	"context"
	"github.com/beng90/spec2go/validate"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ValidationRule struct {
	Field   string
	Rule    string
	Pattern *string
}

var AddOfferValidationRules = []ValidationRule{
	{"features", "omitempty", nil},
	{"features", "omitempty", nil},
	{"features[].id", "required,string,min=1,max=10", validate.Pattern(`^\d+$`)},
	{"features[].valuesIds", "required,min=1", nil},
	{"features[].valuesIds[]", "string", nil},
	{"brandId", "omitempty,string,min=1,max=10", validate.Pattern(`^\d+$`)},
	{"categoryId", "required,string,max=16", validate.Pattern(`^\d+_\d+$`)},
	{"defaultLanguage", "omitempty,string,min=2,max=2", validate.Pattern(`^[a-zA-Z]{2}$`)},
	{"productName", "required,string,min=1,max=255", nil},
	{"variants", "required,min=1,max=1", nil},
	{"variants[].content", "required", nil},
	{"variants[].content[].description", "required,string,min=1,max=65536", nil},
	{"variants[].content[].language", "required,string,min=2,max=2", validate.Pattern(`^[a-zA-Z]{2}$`)},
	{"variants[].delivery", "required,object", nil},
	{"variants[].delivery.additionalInfo", "omitempty,string", nil},
	{"variants[].delivery.dispatchTime", "required,integer,min=1,max=64", nil},
	{"variants[].delivery.shippingTemplateId", "required,string,uuid", nil},
	{"variants[].ean", "omitempty,string", validate.Pattern(`^(\d{13})?$`)},
	{"variants[].inventory", "required,object", nil},
	{"variants[].inventory.size", "required,integer,min=0,max=4294967295", nil},
	{"variants[].isEnabled", "required,boolean", nil},
	{"variants[].media", "required", nil},
	{"variants[].media.images", "required", nil},
	{"variants[].media.images[].sortOrder", "omitempty,integer,min=1,max=64", nil},
	{"variants[].media.images[].url", "required,string,url,max=255", nil},
	{"variants[].price", "required", nil},
	{"variants[].sku", "omitempty,string,min=1,max=64", nil},
	{"variants[].tags", "omitempty", nil},
	{"variants[].tags[].id", "required,string,min=1,max=10", validate.Pattern(`^\d+$`)},
	{"variants[].tags[].valueId", "required,string,min=1,max=10", validate.Pattern(`^\d+$`)},
}

func AddOfferValidate(v *validator.Validate, req *http.Request, ctx context.Context) error {
	schemaValidator, err := validate.NewSchemaValidator(v, req, ctx)
	if err != nil {
		return err
	}

	for _, vRule := range AddOfferValidationRules {
		schemaValidator.AddRule(vRule.Field, vRule.Rule, vRule.Pattern)
	}

	err = schemaValidator.Validate()

	return err
}
