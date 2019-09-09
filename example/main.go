package main

import (
	"bytes"
	"fmt"
	"github.com/beng90/spec2go/validators"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

// registerCustomValidations set custom validators
func registerCustomValidations(validator *validator.Validate) {
	_ = validator.RegisterValidation("ISO8601", IsISO8601Date)
}

func IsISO8601Date(fl validator.FieldLevel) bool {
	ISO8601DateRegexString := "^(-?(?:[1-9][0-9]*)?[0-9]{4})-(1[0-2]|0[1-9])-(3[01]|0[1-9]|[12][0-9])(?:T|\\s)(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])?(Z)?$"
	ISO8601DateRegex := regexp.MustCompile(ISO8601DateRegexString)

	return ISO8601DateRegex.MatchString(fl.Field().String())
}

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
	registerCustomValidations(v)

	return v
}

func main() {
	requestBody := `{"categoryId": 123, "variants": [
		{
			"content": {
				"description": "asd"
			}
		}
	]}`

	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))

	v := NewValidator()

	errs := validators.AddOfferRequestValidate(v, req)
	//fmt.Printf("errs %#v\n", errs)
	for _, e := range errs {
		fmt.Println(e)
	}

	//params := url.Values{}
	//params.Add("")
	//req.URL.RawQuery = params.Encode()
}
