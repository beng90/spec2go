package main

import "fmt"

type SchemaType string
type SchemaFormat string
type RuleType string
type RuleFormat string
type RuleNumeric *float64

const (
	TypeString  SchemaType = "string"
	TypeNumber  SchemaType = "number"
	TypeInteger SchemaType = "integer"
	TypeBoolean SchemaType = "boolean"
	TypeArray   SchemaType = "array"
	TypeObject  SchemaType = "object"

	FormatDate     SchemaFormat = "date"
	FormatDateTime SchemaFormat = "date-time"
	FormatPassword SchemaFormat = "password"
	FormatByte     SchemaFormat = "byte"
	FormatBinary   SchemaFormat = "binary"
	FormatEmail    SchemaFormat = "email"
	FormatUuid     SchemaFormat = "uuid"
	FormatUri      SchemaFormat = "uri"
	FormatHostname SchemaFormat = "hostname"
	FormatIPv4     SchemaFormat = "ipv4"
	FormatIPv6     SchemaFormat = "ipv6"
)

var SchemaTypeToRule = map[SchemaType]RuleType{
	TypeNumber:  "numeric",
	TypeInteger: "numeric",
	TypeBoolean: "bool",
}

var SchemaFormatToRule = map[SchemaFormat]RuleFormat{
	FormatDate:     "ISO8601",
	FormatDateTime: "ISO8601",
	FormatEmail:    "email",
	FormatUuid:     "uuid",
	FormatUri:      "url",
	FormatIPv4:     "ip_v4",
	FormatIPv6:     "ip_v6",
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
	IsObject    bool
}

func (p *Parameter) Rules() (rules []string) {
	if p.Required {
		rules = append(rules, "required")
	}

	if _, hasType := SchemaTypeToRule[SchemaType(p.Type)]; hasType != false {
		rules = append(rules, fmt.Sprintf(`%s`, string(SchemaTypeToRule[SchemaType(p.Type)])))
	}

	if format, hasType := SchemaFormatToRule[SchemaFormat(p.Format)]; hasType != false {
		rules = append(rules, fmt.Sprintf(`%s`, format))
	}

	if p.Min != nil {
		rules = append(rules, fmt.Sprintf(`min=%.f`, *p.Min))
	}

	if p.Max != nil {
		rules = append(rules, fmt.Sprintf(`max=%.f`, *p.Max))
	}

	return
}
