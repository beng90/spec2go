package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/beng90/spec2go/example/validators"
	"github.com/beng90/spec2go/validate"
	"github.com/beng90/spec2go/validate/validations"
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
	_ = v.RegisterValidation("string", validations.IsString)

	return v
}

func main() {
	// example request body
	requestBody := `{
		"categoryId": "123_12",
		"variants": [{"inventory": {}}],
		"variant": {"tags": false}
	}`

	var errs error

	// mocked request
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))

	// create new validator
	v := NewValidator()

	// read body
	buffer, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	// restore body in request
	req.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))

	// verify if json is correct before validation
	if json.Valid(buffer) == false {
		panic(validate.ErrInvalidJSON)
	}

	// final validation
	errs = validators.AddOfferValidate(v, req, context.Background())

	// validator can return two types of error
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
}
