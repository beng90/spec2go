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
{{range . }}
func {{ .Name }}(v *validator.Validate, r *http.Request) map[string]validator.ValidationErrors {
	var err error
	errs := make(ValidationErrors)
    {{range $parameter := .Parameters }}{{range .Rules }}
    err = v.Var(r.Form.Get("{{ $parameter.Name }}"), "{{ . }}")
    try(errs, "{{ $parameter.Name }}", err)
    {{ end }}{{ end }}
	return errs
}
{{ end }}
