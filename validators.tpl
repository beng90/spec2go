package validators

import (
    "github.com/beng90/spec2go/validate"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

{{range . }}{{ if .Parameters }}
func {{ .Name }}(v *validator.Validate, req *http.Request) validate.ValidationErrors {
	schemaValidator := validate.NewSchemaValidator(v, req)

    {{- range $parameter := .Parameters }}
    {{- if .Rules.String }}
    schemaValidator.Validate("{{ $parameter.Name }}", "{{ .Rules }}")
    {{- end }}{{ end }}

	return schemaValidator.Errors()
}
{{ end }}{{ end }}
