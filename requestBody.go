package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

func getRequestBodyParameter(data yaml.MapSlice, paramName string) (param Parameter) {
	param.Name = paramName

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

func getRequestBodyParameters(data yaml.MapSlice, path []string) map[string]*Parameter {
	properties := make(map[string]*Parameter)

	for _, content := range data {
		switch content.Key {
		case "content":
			switch content.Value.(yaml.MapSlice)[0].Key {
			case "application/json":
				properties = getJSONContentProperties(content.Value.(yaml.MapSlice)[0].Value.(yaml.MapSlice))
			}
		case "$ref":
			// TODO: add handler for reference type
			fmt.Println("$ref is not supported yet")
		}
	}

	return properties
}

func getJSONContentProperties(content yaml.MapSlice) map[string]*Parameter {
	properties := make(map[string]*Parameter)

	for _, node := range content[0].Value.(yaml.MapSlice) {
		switch node.Key {
		case "required":
			for _, fieldName := range node.Value.([]interface{}) {
				if _, ok := properties[fieldName.(string)]; ok {
					properties[fieldName.(string)].Required = true
				}
			}
		case "properties":
			for _, property := range node.Value.(yaml.MapSlice) {
				param := getRequestBodyParameter(property.Value.(yaml.MapSlice), property.Key.(string))
				properties[param.Name] = &param
			}
		}
		//fmt.Println(node.Key, node.Value)
	}

	//fmt.Println("properties", properties)
	return properties
}

func generateValidatorsFromRequestBody(params map[string]*Parameter) {
	for _, param := range params {
		fmt.Printf("param %#v\n", param)
	}
}
