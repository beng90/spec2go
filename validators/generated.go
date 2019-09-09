package validators

import (
	"encoding/json"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"strings"
)

type ValidationErrors map[string][]VError

type VError struct {
	Field            string
	Rule             string
	ValidationErrors validator.ValidationErrors
}

func (v VError) Error() string {
	return fmt.Sprintf(`Field "%s" failed in "%s" rule`, v.Field, v.Rule)
}

func try(errs ValidationErrors, fieldName string, err error) {
	if err != nil {
		e := err.(validator.ValidationErrors)

		errs[fieldName] = append(errs[fieldName], VError{
			Field:            fieldName,
			Rule:             e[0].Tag(),
			ValidationErrors: e,
		})
	}
}

type jsonMap map[string]interface{}

func (j jsonMap) getVal(exploded []string, i int, prev interface{}, data *interface{}) {
	if len(exploded) <= i {
		return
	}

	switch v := prev.(type) {
	case string:
		*data = nil
	case jsonMap:
		*data = v[exploded[i]]
		j.getVal(exploded, i+1, v[exploded[i]], data)
	case map[string]interface{}:
		*data = v[exploded[i]]
		j.getVal(exploded, i+1, v[exploded[i]], data)
	case []interface{}:
		j.getVal(exploded, i, v[0], data)
		//default:
		//	fmt.Printf("Type %T\n", v)
		//	fmt.Println("val", v)
	}
}

func (j jsonMap) Get(fieldName string) interface{} {
	exploded := strings.Split(fieldName, ".")
	if len(exploded) > 0 {
		var val interface{}
		j.getVal(exploded, 0, j, &val)

		return val
	}

	return nil
}

func getRequestBody(r *http.Request) jsonMap {
	// Read body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	var requestBody jsonMap
	if err := json.Unmarshal(body, &requestBody); err != nil {
		panic(err)
	}

	return requestBody
}

func AddOfferRequestValidate(v *validator.Validate, r *http.Request) ValidationErrors {
	var err error
	requestBody := getRequestBody(r)
	errs := make(ValidationErrors)

	err = v.Var(requestBody.Get("id"), "required")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("id"), "numeric")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("id"), "min=1")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("id"), "max=4294967295")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("variants"), "required")
	try(errs, "variants", err)

	err = v.Var(requestBody.Get("variants.content"), "required")
	try(errs, "variants.content", err)

	err = v.Var(requestBody.Get("variants.content.description"), "required")
	try(errs, "variants.content.description", err)

	err = v.Var(requestBody.Get("variants.content.description"), "min=1")
	try(errs, "variants.content.description", err)

	err = v.Var(requestBody.Get("variants.content.description"), "max=1024")
	try(errs, "variants.content.description", err)

	err = v.Var(requestBody.Get("variants.content.language"), "required")
	try(errs, "variants.content.language", err)

	err = v.Var(requestBody.Get("variants.content.language"), "min=2")
	try(errs, "variants.content.language", err)

	err = v.Var(requestBody.Get("variants.content.language"), "max=2")
	try(errs, "variants.content.language", err)

	err = v.Var(requestBody.Get("variants.inventory"), "required")
	try(errs, "variants.inventory", err)

	err = v.Var(requestBody.Get("variants.inventory.size"), "required")
	try(errs, "variants.inventory.size", err)

	err = v.Var(requestBody.Get("variants.inventory.size"), "numeric")
	try(errs, "variants.inventory.size", err)

	err = v.Var(requestBody.Get("variants.inventory.size"), "min=1")
	try(errs, "variants.inventory.size", err)

	err = v.Var(requestBody.Get("variants.inventory.size"), "max=4294967295")
	try(errs, "variants.inventory.size", err)

	err = v.Var(requestBody.Get("variants.inventory.sold"), "required")
	try(errs, "variants.inventory.sold", err)

	err = v.Var(requestBody.Get("variants.inventory.sold"), "numeric")
	try(errs, "variants.inventory.sold", err)

	err = v.Var(requestBody.Get("variants.inventory.sold"), "min=1")
	try(errs, "variants.inventory.sold", err)

	err = v.Var(requestBody.Get("variants.inventory.sold"), "max=4294967295")
	try(errs, "variants.inventory.sold", err)

	err = v.Var(requestBody.Get("variants.media"), "required")
	try(errs, "variants.media", err)

	err = v.Var(requestBody.Get("variants.media.images"), "required")
	try(errs, "variants.media.images", err)

	err = v.Var(requestBody.Get("variants.media.images.sortOrder"), "numeric")
	try(errs, "variants.media.images.sortOrder", err)

	err = v.Var(requestBody.Get("variants.media.images.sortOrder"), "min=1")
	try(errs, "variants.media.images.sortOrder", err)

	err = v.Var(requestBody.Get("variants.media.images.sortOrder"), "max=64")
	try(errs, "variants.media.images.sortOrder", err)

	err = v.Var(requestBody.Get("variants.media.images.url"), "required")
	try(errs, "variants.media.images.url", err)

	err = v.Var(requestBody.Get("variants.media.images.url"), "max=255")
	try(errs, "variants.media.images.url", err)

	err = v.Var(requestBody.Get("variants.price"), "required")
	try(errs, "variants.price", err)

	return errs
}

func GetTokenValidate(v *validator.Validate, r *http.Request) ValidationErrors {

	errs := make(ValidationErrors)

	return errs
}
