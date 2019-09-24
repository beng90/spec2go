package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type FieldSchema struct {
	Type       string
	Name       string
	Value      interface{}
	Rules      []string
	Properties MapField
	Items      FieldArray
}

type FieldArray []MapField
type MapField map[string]FieldSchema

func (m MapField) Get(index string) *FieldSchema {
	if _, ok := m[index]; ok {
		v := m[index]

		return &v
	}

	return nil
}

func (m FieldArray) Get(index int) *FieldSchema {
	//fmt.Println("Get", index, len(m))

	if index >= len(m) {
		return nil
	}

	field := &FieldSchema{
		Type:       "",
		Name:       "",
		Value:      m[index]["value"].Value,
		Rules:      nil,
		Properties: nil,
		Items:      nil,
	}

	return field
}

func (f *FieldSchema) UnmarshalJSON(data []byte) error {
	var r interface{}
	_ = json.Unmarshal(data, &r)
	switch v := r.(type) {
	case []interface{}:
		for _, vv := range v {
			switch vv.(type) {
			case string:
				f.Items = append(f.Items, MapField{
					"value": FieldSchema{
						Type:       "",
						Name:       "value",
						Value:      vv,
						Rules:      nil,
						Properties: nil,
						Items:      nil,
					},
				})

				break
			}
		}

		f.Value = v

		if len(f.Items) == 0 {
			_ = json.Unmarshal(data, &f.Items)
		}
	case map[string]interface{}:
		f.Value = v
		_ = json.Unmarshal(data, &f.Properties)
	default:
		f.Value = v
		//fmt.Printf("%T s:%s\n", v, data)
	}

	return nil
}

var jsonBody = `
	{
		"categoryId": "123",
		"features": [
			{
				"key1": "value1",
				"key2": "value2"
			}
		],
		"variants": [
			{
				"inventory": {
					"size": 5
				},
				"tags": [
					"asd",
					"qwe"
				]
			}
		]
	}
`

var rules = &MapField{
	"categoryId": FieldSchema{
		Type:       "string",
		Name:       "categoryId",
		Rules:      []string{"required", "string"},
		Properties: nil,
		Items:      nil,
	},
	"variants": FieldSchema{
		Type:       "array",
		Name:       "variants",
		Rules:      []string{"required", "min=1"},
		Properties: nil,
		Items: []MapField{
			map[string]FieldSchema{
				"inventory": {
					Type:  "object",
					Name:  "inventory",
					Rules: []string{"required"},
					Properties: MapField{
						"size": FieldSchema{
							Type:       "integer",
							Name:       "size",
							Rules:      []string{"required", "integer", "min=1"},
							Properties: nil,
							Items:      nil,
						},
					},
					Items: nil,
				},
				"tags": {
					Type:       "array",
					Name:       "tags",
					Rules:      []string{"string"},
					Properties: nil,
					Items:      nil,
				},
			},
		},
	},
}

func main() {
	var req MapField

	err := json.Unmarshal([]byte(jsonBody), &req)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%#v\n", req["features"].Items[0]["key1"].Value)
	//fmt.Printf("%#v\n", req["features"].Items[0].Get("key2").Value)
	//fmt.Printf("%#v\n", req["variants"].Items[0]["tags"].Items.Get(0).Value)
	//fmt.Printf("%#v\n", req["variants"].Items[0]["tags"].Items.Get(1).Value)
	//fmt.Printf("%#v\n", req["variants"].Items[0]["tags"].Items.Get(2))
	//fmt.Printf("%#v\n", req["variants"].Items[0]["tags"].Items[1].Value)
	//fmt.Printf("%#v\n", req["variants"].Items[0]["tags"].Items[0].Get("asd"))

	walk(req, *rules, []string{})
}

func walk(req MapField, rules MapField, path []string) {
	for _, field := range rules {
		path = append(path, field.Name)
		//fmt.Println(i, req[item.Name])
		//fmt.Println(i)

		if field.Properties != nil {
			walk(req, field.Properties, path)
		}

		if field.Type == "array" {
			path[len(path)-1] = path[len(path)-1] + "[]"
		}

		if field.Items != nil {
			for _, item := range field.Items {
				walk(req, item, path)
			}
		}

		fmt.Println("fieldName:", strings.Join(path, "."), "rule:", field.Rules)
		fmt.Println("value", req.Get(strings.Join(path, ".")))
		path = path[:len(path)-1]
	}
}
