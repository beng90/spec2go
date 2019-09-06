package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

func getParameter(data yaml.MapSlice) Parameter {
	param := &Parameter{}

	for _, property := range data {
		switch property.Key {
		case "schema":
			getSchema(param, property.Value.(yaml.MapSlice))
		case "name":
			param.Name = property.Value.(string)
		case "in":
			param.In = property.Value.(string)
		case "required":
			param.Required = property.Value.(bool)
		case "description":
			param.Description = property.Value.(string)
		}
	}

	//fmt.Printf("property %v\n", param)

	return *param
}

func getParameters(data []interface{}, path []string) (parameters []Parameter) {
	for _, param := range data {
		switch paramVal := param.(type) {
		case yaml.MapSlice:
			parameter := getParameter(paramVal)
			path = append(path, parameter.Name)
			parameters = append(parameters, parameter)
		}
	}

	return
}

func generateValidatorsFromParameters(params []Parameter) {
	for _, param := range params {
		if param.Min != nil {
			fmt.Printf("param %#v\n", param)
		}
	}
}
