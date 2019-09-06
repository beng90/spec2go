package validators

import (
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ValidationErrors map[string]validator.ValidationErrors

func try(errs ValidationErrors, fieldName string, err error) {
	if err != nil {
		if _, ok := errs[fieldName]; ok {
			errs[fieldName] = append(errs[fieldName], err.(validator.ValidationErrors)[0])
		} else {
			errs[fieldName] = err.(validator.ValidationErrors)
		}
	}
}

func AddOfferRequestValidate(v *validator.Validate, r *http.Request) map[string]validator.ValidationErrors {
	var err error
	errs := make(ValidationErrors)

	err = v.Var(r.Form.Get("categoryId"), "required")
	try(errs, "categoryId", err)

	err = v.Var(r.Form.Get("categoryId"), "numeric")
	try(errs, "categoryId", err)

	err = v.Var(r.Form.Get("categoryId"), "min=1.00")
	try(errs, "categoryId", err)

	err = v.Var(r.Form.Get("categoryId"), "max=4294967295.00")
	try(errs, "categoryId", err)

	err = v.Var(r.Form.Get("categoryTree"), "required")
	try(errs, "categoryTree", err)

	err = v.Var(r.Form.Get("createdAt"), "required")
	try(errs, "createdAt", err)

	err = v.Var(r.Form.Get("createdAt"), "date:dd-mm-yyyy H:i:s")
	try(errs, "createdAt", err)

	err = v.Var(r.Form.Get("expireAt"), "required")
	try(errs, "expireAt", err)

	err = v.Var(r.Form.Get("expireAt"), "date:dd-mm-yyyy H:i:s")
	try(errs, "expireAt", err)

	err = v.Var(r.Form.Get("id"), "required")
	try(errs, "id", err)

	err = v.Var(r.Form.Get("id"), "numeric")
	try(errs, "id", err)

	err = v.Var(r.Form.Get("id"), "min=1.00")
	try(errs, "id", err)

	err = v.Var(r.Form.Get("id"), "max=4294967295.00")
	try(errs, "id", err)

	err = v.Var(r.Form.Get("productId"), "required")
	try(errs, "productId", err)

	err = v.Var(r.Form.Get("productId"), "numeric")
	try(errs, "productId", err)

	err = v.Var(r.Form.Get("productId"), "min=1.00")
	try(errs, "productId", err)

	err = v.Var(r.Form.Get("productId"), "max=99999999999999.00")
	try(errs, "productId", err)

	err = v.Var(r.Form.Get("productName"), "min=1.00")
	try(errs, "productName", err)

	err = v.Var(r.Form.Get("productName"), "max=255.00")
	try(errs, "productName", err)

	err = v.Var(r.Form.Get("productTypeId"), "required")
	try(errs, "productTypeId", err)

	err = v.Var(r.Form.Get("productTypeId"), "numeric")
	try(errs, "productTypeId", err)

	err = v.Var(r.Form.Get("productTypeId"), "min=1.00")
	try(errs, "productTypeId", err)

	err = v.Var(r.Form.Get("productTypeId"), "max=4294967295.00")
	try(errs, "productTypeId", err)

	err = v.Var(r.Form.Get("updatedAt"), "required")
	try(errs, "updatedAt", err)

	err = v.Var(r.Form.Get("updatedAt"), "date:dd-mm-yyyy H:i:s")
	try(errs, "updatedAt", err)

	err = v.Var(r.Form.Get("variants"), "required")
	try(errs, "variants", err)

	return errs
}

func GetTokenValidate(v *validator.Validate, r *http.Request) map[string]validator.ValidationErrors {
	var err error
	errs := make(ValidationErrors)

	return errs
}
