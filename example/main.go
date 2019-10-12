package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/beng90/spec2go/example/validators"
	"github.com/beng90/spec2go/validate"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
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

	_ = v.RegisterValidation("string", validate.IsString)

	// custom validations
	return v
}

type SingleRule struct {
	Rule     string
	Children interface{}
}
type RulesMap map[string]SingleRule

var somerules RulesMap = RulesMap{
	"categoryId": SingleRule{Rule: "required,string"},
	"brand":      SingleRule{Rule: "required,string"},
	"additionalInfo": SingleRule{
		Rule: "",
		Children: map[string]SingleRule{
			"id": {
				Rule: "required,string",
			},
			"valuesIds": {
				Rule: "required,min=1",
				Children: []SingleRule{
					{
						Rule: "string",
					},
				},
			},
		},
	},
}

type testObject struct {
	CategoryId     interface{} `validate:"required,string"`
	Brand          interface{} `validate:"required,string"`
	AdditionalInfo []struct {
		Id        interface{}   `validate:"required,string"`
		ValuesIds []interface{} `validate:"required,min=1"`
	}
}

func main() {
	requestBody := `{
		"categoryId": "123",
		"brand": "123",
		"additionalInfo": [
			{
				"id": "1",
				"valuesIds": [
					"222",
					"333",
					"444"
				]
			}
		],
		"variants": [
			{
				"isEnabled": true,
				"content": [
					{
						"language": "pl",
						"description": "asd"
					}
				],
				"tags": [
					{
						"id": "123",
						"valueId": "123"
					}
				],
				"media": [
					{
						"images": [
							{
								"url": "http://asd"
							}
						]
					}
				]
			}
		]
	}`

	var errs error
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))

	v := NewValidator()

	// Read body
	buffer, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	// restore body in request
	req.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))

	if json.Valid(buffer) == false {
		panic(validate.ErrInvalidJSON)
	}

	var x testObject
	if err := json.Unmarshal(buffer, &x); err != nil {
		panic(err)
	}

	errs = v.Struct(x)

	//fmt.Printf("x: %#v\n", errs)
	//return

	// flag to turn on debug mode
	validate.IsDebugMode = false
	errs = validators.AddOfferValidate(v, req)

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
