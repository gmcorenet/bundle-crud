package model

import (
	"sort"
)

type FieldType string

const (
	FieldTypeString   FieldType = "string"
	FieldTypeInt      FieldType = "int"
	FieldTypeFloat    FieldType = "float"
	FieldTypeBool     FieldType = "bool"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeTime     FieldType = "time"
	FieldTypeChoice   FieldType = "choice"
	FieldTypeRelation FieldType = "relation"
	FieldTypeText     FieldType = "text"
	FieldTypeFile     FieldType = "file"
	FieldTypeImage    FieldType = "image"
)

var filterOperatorsText = []string{"eq", "neq", "like", "not_like", "is_null", "is_not_null"}
var filterOperatorsNumeric = []string{"eq", "neq", "gt", "gte", "lt", "lte", "is_null", "is_not_null"}
var filterOperatorsBool = []string{"eq", "neq"}
var filterOperatorsChoice = []string{"eq", "neq", "is_null", "is_not_null"}
var filterOperatorsChoiceMulti = []string{"in", "not_in", "is_null", "is_not_null"}

type Field struct {
	Name             string   `json:"name" yaml:"name"`
	LabelKey         string   `json:"label_key" yaml:"label_key"`
	Type             FieldType `json:"type" yaml:"type"`
	Required         bool     `json:"required" yaml:"required"`
	Searchable       bool     `json:"searchable" yaml:"searchable"`
	Sortable         bool     `json:"sortable" yaml:"sortable"`
	Writable         bool     `json:"writable" yaml:"writable"`
	WritableOnCreate bool     `json:"writable_on_create" yaml:"writable_on_create"`
	WritableOnEdit   bool     `json:"writable_on_edit" yaml:"writable_on_edit"`
	WritableOnClone  bool     `json:"writable_on_clone" yaml:"writable_on_clone"`
	Visible          bool     `json:"visible" yaml:"visible"`
	Filterable       bool     `json:"filterable" yaml:"filterable"`
	FilterOperators  []string `json:"filter_operators" yaml:"filter_operators"`
	FilterWidget     string   `json:"filter_widget" yaml:"filter_widget"`
	Relation         string   `json:"relation" yaml:"relation"`
	Choices          map[string]string `json:"choices" yaml:"choices"`
	Multiple         bool     `json:"multiple" yaml:"multiple"`
	Expanded         bool     `json:"expanded" yaml:"expanded"`
	FormType         string   `json:"form_type" yaml:"form_type"`
	FormOptions      map[string]interface{} `json:"form_options" yaml:"form_options"`
	HelpKey          string   `json:"help_key" yaml:"help_key"`
	DefaultVal       interface{} `json:"default" yaml:"default"`

	originalFilterOperators []string
}

func NewField(name string, fieldType FieldType) *Field {
	return &Field{
		Name:              name,
		LabelKey:          "field." + name,
		Type:              fieldType,
		Writable:          true,
		WritableOnCreate:  true,
		WritableOnEdit:    true,
		WritableOnClone:   true,
		Visible:           true,
		Filterable:        false,
	}
}

func (f *Field) IsWritableInMode(mode string) bool {
	if !f.Writable {
		return false
	}
	switch mode {
	case "create":
		return f.WritableOnCreate
	case "edit":
		return f.WritableOnEdit
	case "clone":
		return f.WritableOnClone
	default:
		return f.Writable
	}
}

func (f *Field) ResolvedFilterOperators() []string {
	if len(f.originalFilterOperators) > 0 && !isDefaultOperators(f.originalFilterOperators) {
		return f.FilterOperators
	}
	switch f.Type {
	case FieldTypeBool:
		return filterOperatorsBool
	case FieldTypeInt, FieldTypeFloat, FieldTypeDateTime, FieldTypeDate, FieldTypeTime:
		return filterOperatorsNumeric
	case FieldTypeChoice, FieldTypeRelation:
		if f.Multiple {
			return filterOperatorsChoiceMulti
		}
		return filterOperatorsChoice
	default:
		return filterOperatorsText
	}
}

func (f *Field) ResolvedFilterWidget() string {
	if f.FilterWidget != "" {
		return f.FilterWidget
	}
	switch f.Type {
	case FieldTypeBool:
		return "boolean"
	case FieldTypeChoice:
		return "choice"
	case FieldTypeRelation:
		if f.Multiple {
			return "relation_multiple"
		}
		return "relation"
	case FieldTypeInt, FieldTypeFloat:
		return "number"
	case FieldTypeDateTime:
		return "datetime"
	case FieldTypeDate:
		return "date"
	case FieldTypeTime:
		return "time"
	default:
		return "text"
	}
}

func isDefaultOperators(ops []string) bool {
	defaults := []string{"eq", "neq", "like", "gt", "lt"}
	if len(ops) != len(defaults) {
		return false
	}
	sort.Strings(ops)
	sorted := make([]string, len(defaults))
	copy(sorted, defaults)
	sort.Strings(sorted)
	for i := range ops {
		if ops[i] != sorted[i] {
			return false
		}
	}
	return true
}

func (f *Field) SetOriginalFilterOperators(ops []string) {
	f.originalFilterOperators = make([]string, len(ops))
	copy(f.originalFilterOperators, ops)
}
