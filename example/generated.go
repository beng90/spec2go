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

func addPetValidate(v *validator.Validate, r *http.Request) map[string]validator.ValidationErrors {
	var err error
	errs := make(ValidationErrors)

	err = v.Var(r.Form.Get("id"), "required")
	try(errs, "id", err)

	err = v.Var(r.Form.Get("id"), "min=1.20")
	try(errs, "id", err)

	err = v.Var(r.Form.Get("name"), "required")
	try(errs, "id", err)

	err = v.Var(r.Form.Get("photoUrls"), "required")
	try(errs, "id", err)

	return errs
}
