package validators

import (
	"encoding/json"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
)

type ValidationErrors map[string][]VError

type VError struct {
	Field string
	Rule string
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

{{range . }}
func {{ .Name }}(v *validator.Validate, r *http.Request) ValidationErrors {
	{{ if .Parameters }}var err error
	requestBody := getRequestBody(r){{ end }}
	errs := make(ValidationErrors)
    {{range $parameter := .Parameters }}{{range .Rules }}
    err = v.Var(requestBody.Get("{{ $parameter.Name }}"), "{{ . }}")
    try(errs, "{{ $parameter.Name }}", err)
    {{ end }}{{ end }}
	return errs
}
{{ end }}
