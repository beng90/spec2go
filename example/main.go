package main

import (
	"bytes"
	"fmt"
	"github.com/beng90/spec2go/example/validators"
	"github.com/beng90/spec2go/validate"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"net/http"
	"reflect"
	"strings"
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

func main() {
	requestBody := `{
		"categoryId": "123",
		"brand": "123",
		"productName": "test",
		"additionalInfo": [
			{
				"id": 1,
				"valuesIds": [
					222,
					333
				]
			},
			{
				"id": "2"
			}
		],
		"variants": [
			{
				"delivery": {
					"dispatchTime": 3,
					"shippingTemplateId": "5839c1a6-293f-43bf-ba8b-8a3cb19f4ea5"
				},
				"isEnabled": true,
				"content": {
					"description": "asd",
					"language": "pl"
				},
				"price": "123",
				"inventory": {
					"size": 123
				},
				"media": {
					"images": [
						{
							"type": "image",
							"url": "https://psy-pies.com/pliki/image/foto/duze/foto54eefb49dad42.jpg",
							"sortOrder": 1
						},
						{
							"sortOrder": 2,
							"url": "https://skuteczna-samoobrona.pl/wp-content/uploads/rottweiler.jpg"
						}	
					]
				},
				"tags": [
					{
						"id": "1"
					}
				]
			}
		]
	}`

	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))

	v := NewValidator()

	// flag to turn on debug mode
	validate.IsDebugMode = false
	errs := validators.AddOfferValidate(v, req)

	switch vErr := errs.(type) {
	case validate.ValidationErrors:
		for _, e := range errs.(validate.ValidationErrors) {
			fmt.Println(e)
		}
	default:
		if vErr == validate.ErrInvalidJSON {
			log.Println(vErr)
		}
	}
	//fmt.Printf("errs %#v\n", errs)
}
