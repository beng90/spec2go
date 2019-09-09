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

	err = v.Var(requestBody.Get("categoryId"), "required")
	try(errs, "categoryId", err)

	err = v.Var(requestBody.Get("categoryId"), "numeric")
	try(errs, "categoryId", err)

	err = v.Var(requestBody.Get("categoryId"), "min=1")
	try(errs, "categoryId", err)

	err = v.Var(requestBody.Get("categoryId"), "max=4294967295")
	try(errs, "categoryId", err)

	err = v.Var(requestBody.Get("categoryTree"), "required")
	try(errs, "categoryTree", err)

	err = v.Var(requestBody.Get("categoryTree.id"), "required")
	try(errs, "categoryTree.id", err)

	err = v.Var(requestBody.Get("categoryTree.id"), "numeric")
	try(errs, "categoryTree.id", err)

	err = v.Var(requestBody.Get("categoryTree.id"), "min=1")
	try(errs, "categoryTree.id", err)

	err = v.Var(requestBody.Get("categoryTree.id"), "max=4294967295")
	try(errs, "categoryTree.id", err)

	err = v.Var(requestBody.Get("categoryTree.name"), "required")
	try(errs, "categoryTree.name", err)

	err = v.Var(requestBody.Get("categoryTree.name"), "max=128")
	try(errs, "categoryTree.name", err)

	err = v.Var(requestBody.Get("createdAt"), "required")
	try(errs, "createdAt", err)

	err = v.Var(requestBody.Get("createdAt"), "ISO8601")
	try(errs, "createdAt", err)

	err = v.Var(requestBody.Get("expireAt"), "required")
	try(errs, "expireAt", err)

	err = v.Var(requestBody.Get("expireAt"), "ISO8601")
	try(errs, "expireAt", err)

	err = v.Var(requestBody.Get("id"), "required")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("id"), "numeric")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("id"), "min=1")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("id"), "max=4294967295")
	try(errs, "id", err)

	err = v.Var(requestBody.Get("productId"), "required")
	try(errs, "productId", err)

	err = v.Var(requestBody.Get("productId"), "numeric")
	try(errs, "productId", err)

	err = v.Var(requestBody.Get("productId"), "min=1")
	try(errs, "productId", err)

	err = v.Var(requestBody.Get("productId"), "max=99999999999999")
	try(errs, "productId", err)

	err = v.Var(requestBody.Get("productName"), "min=1")
	try(errs, "productName", err)

	err = v.Var(requestBody.Get("productName"), "max=255")
	try(errs, "productName", err)

	err = v.Var(requestBody.Get("productTypeId"), "required")
	try(errs, "productTypeId", err)

	err = v.Var(requestBody.Get("productTypeId"), "numeric")
	try(errs, "productTypeId", err)

	err = v.Var(requestBody.Get("productTypeId"), "min=1")
	try(errs, "productTypeId", err)

	err = v.Var(requestBody.Get("productTypeId"), "max=4294967295")
	try(errs, "productTypeId", err)

	err = v.Var(requestBody.Get("updatedAt"), "required")
	try(errs, "updatedAt", err)

	err = v.Var(requestBody.Get("updatedAt"), "ISO8601")
	try(errs, "updatedAt", err)

	err = v.Var(requestBody.Get("variants"), "required")
	try(errs, "variants", err)

	err = v.Var(requestBody.Get("variants.additionalInfo.id"), "required")
	try(errs, "variants.additionalInfo.id", err)

	err = v.Var(requestBody.Get("variants.additionalInfo.id"), "max=32")
	try(errs, "variants.additionalInfo.id", err)

	err = v.Var(requestBody.Get("variants.additionalInfo.value"), "required")
	try(errs, "variants.additionalInfo.value", err)

	err = v.Var(requestBody.Get("variants.additionalInfo.value"), "max=64")
	try(errs, "variants.additionalInfo.value", err)

	err = v.Var(requestBody.Get("variants.attributes"), "required")
	try(errs, "variants.attributes", err)

	err = v.Var(requestBody.Get("variants.attributes.id"), "required")
	try(errs, "variants.attributes.id", err)

	err = v.Var(requestBody.Get("variants.attributes.id"), "max=32")
	try(errs, "variants.attributes.id", err)

	err = v.Var(requestBody.Get("variants.attributes.value"), "required")
	try(errs, "variants.attributes.value", err)

	err = v.Var(requestBody.Get("variants.attributes.value"), "max=64")
	try(errs, "variants.attributes.value", err)

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

	err = v.Var(requestBody.Get("variants.delivery"), "required")
	try(errs, "variants.delivery", err)

	err = v.Var(requestBody.Get("variants.delivery.dispatchTime"), "required")
	try(errs, "variants.delivery.dispatchTime", err)

	err = v.Var(requestBody.Get("variants.delivery.dispatchTime"), "numeric")
	try(errs, "variants.delivery.dispatchTime", err)

	err = v.Var(requestBody.Get("variants.delivery.dispatchTime"), "min=1")
	try(errs, "variants.delivery.dispatchTime", err)

	err = v.Var(requestBody.Get("variants.delivery.dispatchTime"), "max=64")
	try(errs, "variants.delivery.dispatchTime", err)

	err = v.Var(requestBody.Get("variants.delivery.shippingTemplateId"), "required")
	try(errs, "variants.delivery.shippingTemplateId", err)

	err = v.Var(requestBody.Get("variants.delivery.shippingTemplateId"), "uuid")
	try(errs, "variants.delivery.shippingTemplateId", err)

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

	err = v.Var(requestBody.Get("variants.isEnabled"), "required")
	try(errs, "variants.isEnabled", err)

	err = v.Var(requestBody.Get("variants.isEnabled"), "oneof:true false")
	try(errs, "variants.isEnabled", err)

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
