package main

import (
    "context"
    "github.com/beng90/spec2go/validate"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ValidationRule struct {
	Field   string
	Rule    string
	Pattern *string
}
{{ range . }}{{ if .Parameters }}
var {{ .Name }}Rules = []ValidationRule{
    {{- range $parameter := .Parameters }}
    {{- if .Rules.String }}
    {"{{ $parameter.Name }}", "{{ .Rules }}", nil},
    {{- end }}{{ end }}
}

func {{ .Name }}(v *validator.Validate, req *http.Request, ctx context.Context) error {
	schemaValidator, err := validate.NewSchemaValidator(v, req, ctx)
    if err != nil {
        return err
    }

    for _, vRule := range {{ .Name }}Rules {
        schemaValidator.AddRule(vRule.Field, vRule.Rule, vRule.Pattern)
    }

	err = schemaValidator.Validate()

    return err
}
{{ end }}{{ end }}
