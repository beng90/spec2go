package validate

import (
	"encoding/json"
	"strings"
)

type Rules []string

func (r Rules) String() string {
	return strings.Join(r, ",")
}

func (r Rules) Required() bool {
	for _, rule := range r {
		if rule == "required" {
			return true
		}
	}

	return false
}

type FieldSchema struct {
	Type       string
	Name       string
	Value      interface{}
	Rules      Rules
	Properties MapField
	Items      FieldArray
}

type SliceField interface {
	Get(index string) interface{}
}

type FieldArray []MapField
type MapField map[string]*FieldSchema

func (m MapField) Get(index string) *FieldSchema {
	if _, ok := m[index]; ok {
		v := m[index]

		return v
	}

	return nil
}

func (m FieldArray) Get(index int) *FieldSchema {
	//fmt.Println("Get", index, len(m))

	if index >= len(m) {
		return nil
	}

	field := &FieldSchema{
		Type: "",
		Name: "",
		//Value:      m[index]["value"].Value,
		Value:      "",
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
			case string, float64:
				f.Items = append(f.Items, MapField{
					"value": &FieldSchema{
						Type:  "item",
						Name:  "value",
						Value: vv,
					},
				})

				break
				//default:
				//	fmt.Printf("Type: %T\n", vv)
			}
		}

		//fmt.Printf("r: %#v\n", r)
		f.Type = "array"
		f.Name = "array"
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

func (f *FieldSchema) IsRequired() bool {
	for _, rule := range f.Rules {
		if rule == "required" {
			return true
		}
	}

	return false
}
