package generate

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
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

func GetRequestBodyParameters(data yaml.MapSlice, path []string) map[string]*Parameter {
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
				path := []string{}
				properties = getJSONProperties(properties, property, path)
			}
		}
		//fmt.Println(node.Key, node.Value)
	}

	//fmt.Println("properties", properties)
	return properties
}

func getJSONProperties(properties map[string]*Parameter, node interface{}, path []string) map[string]*Parameter {
	switch element := node.(type) {
	case yaml.MapItem:
		data := element.Value.(yaml.MapSlice)
		path = append(path, element.Key.(string))
		paramName := strings.Join(path, ".")
		//fmt.Println("paramName", paramName)
		param := getRequestBodyParameter(data, paramName)
		if val, ok := properties[paramName]; ok {
			param.Required = val.Required
		}

		//fmt.Println("param.Name", paramName)

		for _, embeded := range data {
			switch embeded.Key {
			case "properties":
				//fmt.Println("param.Name2", param.Name)
				param.IsObject = true
				getJSONProperties(properties, data, path)
			//case "description":
			//	fmt.Println("desc", embeded.Value)
			case "items":
				path[len(path)-1] = path[len(path)-1] + "[]"
				fmt.Println("param.Name", path)
				getJSONProperties(properties, embeded.Value, path)
			}
		}

		properties[param.Name] = &param
	case yaml.MapSlice:
		//fmt.Println("222")
		//fmt.Printf("Type: %T\n", element)

		for _, embeded := range element {
			//fmt.Println("333", embeded.Key)
			switch embeded.Key {
			case "type":
				if embeded.Value.(string) != "object" && embeded.Value.(string) != "array" {
					paramName := strings.Join(path, ".")
					param := getRequestBodyParameter(element, paramName)
					properties[paramName] = &param
					break
				}
			case "properties":
				for _, property := range embeded.Value.(yaml.MapSlice) {
					//fmt.Println("property", property.Key)
					//fmt.Println("path", path)
					getJSONProperties(properties, property, path)
				}
			case "required":
				for _, fieldName := range embeded.Value.([]interface{}) {
					fieldPath := append(path, fieldName.(string))
					paramName := strings.Join(fieldPath, ".")
					if _, ok := properties[paramName]; ok {
						properties[paramName].Required = true
					}
				}
			}
		}
	}

	return properties
}
