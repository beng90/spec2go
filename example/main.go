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

func main() {
	requestBody := `{
	}`

	//requestBody2 := `{
	//}`

	var errs error
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody)))
	//req, _ = http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(requestBody2)))

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
