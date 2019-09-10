package main

import (
	"bytes"
	"fmt"
	"github.com/beng90/spec2go/validators"
	"gopkg.in/go-playground/validator.v9"
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
					"333"
				]
			},
			{
				"id": 2,
				"valuesIds": [
					"66",
					99
				]
			}
		],
		"variants": [
		{
			"isEnabled": true,
			"content": {
				"description": "asd",
				"language": "pl"
			},
			"price": "123",
			"inventory": {
				"size": 123
			}
		},
		{
			"isEnabled": true,
			"content": {
				"language": "asd"
			},
			"price": "123",
			"inventory": {
				"size": 123
			}
		}
	]}`

	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))

	v := NewValidator()

	errs := validators.AddOfferValidate(v, req)
	//fmt.Printf("errs %#v\n", errs)
	for _, e := range errs {
		fmt.Println(e)
	}
}
