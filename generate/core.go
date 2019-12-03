package generate

import (
	"gopkg.in/yaml.v2"
	"strings"
)

const (
	SpecParameters  = "parameters"
	SpecRequestBody = "requestBody"
)

type Validator struct {
	Name       string
	Parameters map[string]*Parameter
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
				v := float64(schemaProperty.Value.(int))
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

func walk(validators *[]Validator, spec yaml.MapSlice, path []string) {
	for _, node := range spec {
		switch nodeVal := node.Value.(type) {
		case string:
			if node.Key.(string) == "operationId" {
				path = append(path, node.Value.(string))
			}
			//fmt.Println(path)
			//fmt.Println(node.Key, node, path)
		case yaml.MapSlice:
			//path = append(path, node.Key.(string))

			if node.Key == SpecRequestBody {
				//fmt.Println(node.Key, node.Value)
				//fmt.Println("generateValidatorsFromRequestBody", path)
				parameters := GetRequestBodyParameters(nodeVal, path)
				*validators = append(*validators, Validator{
					Name:       strings.Title(path[0]) + "Validate",
					Parameters: parameters,
				})

				//for _, pam := range parameters {
				//	fmt.Println()
				//	fmt.Printf("%#v\n", pam)
				//}
				//generateValidatorsFromRequestBody(parameters)
			}
			walk(validators, nodeVal, path)
			//path = nil
		case []interface{}:
			switch node.Key {
			case SpecParameters:
				//fmt.Println("generateValidatorsFromParameters", path)
				//parameters := getParameters(nodeVal, path)
				//generateValidatorsFromParameters(parameters)
			}
		default:
			//fmt.Printf("%T\n", node)
			//fmt.Println(node)
			//fmt.Println(node.Key, reflect.TypeOf(node).String())
		}
		//path = nil
	}
}

func Generate(validators *[]Validator, spec yaml.MapSlice) []string {
	var path []string
	walk(validators, spec, path)

	return path
}
