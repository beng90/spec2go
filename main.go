package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	SpecParameters  = "parameters"
	SpecRequestBody = "requestBody"
)

func main() {
	file, err := ioutil.ReadFile("api.yml")
	if err != nil {
		panic(err)
	}

	data := yaml.MapSlice{}
	yaml.Unmarshal(file, &data)

	//fmt.Println(data)

	path := []string{}
	walk(data, path)
}

func walk(spec yaml.MapSlice, path []string) {
	for _, node := range spec {
		path = append(path, node.Key.(string))
		switch nodeVal := node.Value.(type) {
		case string:
			//fmt.Println(path)
			//fmt.Println(node.Key, node, path)
		case yaml.MapSlice:
			if node.Key == SpecRequestBody {
				//fmt.Println(node.Key, node.Value)
				fmt.Println("generateValidatorsFromRequestBody", path)
				parameters := getRequestBodyParameters(nodeVal, path)
				//fmt.Println("parameters", parameters)
				generateValidatorsFromRequestBody(parameters)
			}
			walk(nodeVal, path)
		case []interface{}:
			switch node.Key {
			case SpecParameters:
				fmt.Println("generateValidatorsFromParameters", path)
				parameters := getParameters(nodeVal, path)
				generateValidatorsFromParameters(parameters)
			}
		default:
			//fmt.Printf("%T\n", node)
			//fmt.Println(node)
			//fmt.Println(node.Key, reflect.TypeOf(node).String())
		}
		path = nil
	}
}

type Parameter struct {
	Name        string
	In          string
	Required    bool
	Description string
	Type        string
	ArrayType   string
	Format      string
	Pattern     string
	Min         *float64
	Max         *float64
}

func getSchema(param *Parameter, schema yaml.MapSlice) {
	for _, schemaProperty := range schema {
		//fmt.Println("schemaProperty", schemaProperty.Key, schemaProperty.Value)
		switch schemaProperty.Key {
		case "type":
			param.Type = schemaProperty.Value.(string)
		case "format":
			param.Format = schemaProperty.Value.(string)
		case "pattern":
			param.Pattern = schemaProperty.Value.(string)
		case "minimum", "minLength":
			switch schemaProperty.Value.(type) {
			case int:
				v := float64(schemaProperty.Value.(float64))
				param.Min = &v
			case float64:
				v := float64(schemaProperty.Value.(float64))
				param.Min = &v
			}
		case "maximum", "maxLength":
			switch schemaProperty.Value.(type) {
			case int:
				v := float64(schemaProperty.Value.(int))
				param.Max = &v
			case float64:
				v := float64(schemaProperty.Value.(float64))
				param.Max = &v
			}
		}
	}
}

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
			fmt.Printf("param %v\n", param)
		}
	}
}

func getRequestBodyParameter(data yaml.MapSlice, paramName string, requiredFields map[string]string) (param Parameter) {
	param.Name = paramName

	_, isRequired := requiredFields[param.Name]
	param.Required = isRequired

	for _, property := range data {
		switch property.Key {
		case "description":
			param.Description = property.Value.(string)
		case "type":
			param.Type = property.Value.(string)
		case "format":
			param.Format = property.Value.(string)
		case "items":
			for _, arrayItem := range property.Value.(yaml.MapSlice) {
				switch arrayItem.Key {
				case "type":
					param.ArrayType = arrayItem.Value.(string)
				case "$ref":
					fmt.Println("$ref is not supported yet")
				}
				//param.Format = property.Value.(string)
			}
		}
	}

	getSchema(&param, data)

	//fmt.Printf("property %s\n", param)

	return
}

func getRequestBodyParameters(data yaml.MapSlice, path []string) (parameters []Parameter) {
	for _, content := range data {
		switch content.Key {
		case "content":
			switch content.Value.(yaml.MapSlice)[0].Key {
			case "application/json":
				parameters = getJSONContentProperties(content.Value.(yaml.MapSlice)[0].Value.(yaml.MapSlice))
			}
		case "$ref":
			// TODO: add handler for reference type
			fmt.Println("$ref is not supported yet")
		}
	}

	return
}

func getJSONContentProperties(content yaml.MapSlice) (properties []Parameter) {
	// go to "schema" node
	requiredFields := make(map[string]string)

	for _, node := range content[0].Value.(yaml.MapSlice) {
		switch node.Key {
		// TODO: Get required fields at first
		case "required":
			for _, fieldName := range node.Value.([]interface{}) {
				requiredFields[fieldName.(string)] = fieldName.(string)
			}
		case "properties":
			for _, property := range node.Value.(yaml.MapSlice) {
				param := getRequestBodyParameter(property.Value.(yaml.MapSlice), property.Key.(string), requiredFields)
				properties = append(properties, param)
			}
		}
		//fmt.Println(node.Key, node.Value)
	}

	//fmt.Println("properties", properties)
	return properties
}

func generateValidatorsFromRequestBody(params []Parameter) {
	for _, param := range params {
		fmt.Println("param", param)
	}
}
