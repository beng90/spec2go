package validators

import (
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

{{range . }}{{ if .Parameters }}
func {{ .Name }}(v *validator.Validate, req *http.Request) ValidationErrors {
	schemaValidator := NewSchemaValidator(v, req)

    {{- range $parameter := .Parameters }}
    {{- if .Rules.String }}
    schemaValidator.validate("{{ $parameter.Name }}", "{{ .Rules }}")
    {{- end }}{{ end }}

	return schemaValidator.Errors()
}
{{ end }}{{ end }}
