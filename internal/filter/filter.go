package filter

import (
	"github.com/gmcorenet/bundle-crud/internal/model"
)

type FilterType interface {
	Supports(field model.Field) bool
	Operators(field model.Field) []string
	NormalizeValue(value interface{}, field model.Field) interface{}
	WidgetKey() string
}

type FilterTypeRegistry struct {
	types []FilterType
}

func NewFilterTypeRegistry() *FilterTypeRegistry {
	r := &FilterTypeRegistry{}
	r.types = []FilterType{
		&StringFilterType{},
		&NumericFilterType{},
		&BoolFilterType{},
		&DateFilterType{},
		&ChoiceFilterType{},
		&RelationFilterType{},
	}
	return r
}

func (r *FilterTypeRegistry) Register(ft FilterType) {
	r.types = append(r.types, ft)
}

func (r *FilterTypeRegistry) Widget(field model.Field) string {
	for _, ft := range r.types {
		if ft.Supports(field) {
			return ft.WidgetKey()
		}
	}
	return "text"
}

func (r *FilterTypeRegistry) Operators(field model.Field) []string {
	for _, ft := range r.types {
		if ft.Supports(field) {
			return ft.Operators(field)
		}
	}
	return []string{"eq", "neq", "like"}
}

func (r *FilterTypeRegistry) NormalizeOperator(field model.Field, operator string, value interface{}) string {
	allowed := r.Operators(field)
	for _, op := range allowed {
		if op == operator {
			return operator
		}
	}
	if len(allowed) > 0 {
		return allowed[0]
	}
	return "eq"
}

type StringFilterType struct{}

func (s *StringFilterType) Supports(field model.Field) bool {
	switch field.Type {
	case model.FieldTypeString, model.FieldTypeText:
		return true
	}
	return false
}

func (s *StringFilterType) Operators(field model.Field) []string {
	return []string{"eq", "neq", "like", "not_like", "is_null", "is_not_null"}
}

func (s *StringFilterType) NormalizeValue(value interface{}, field model.Field) interface{} {
	if v, ok := value.(string); ok {
		return v
	}
	return ""
}

func (s *StringFilterType) WidgetKey() string { return "text" }

type NumericFilterType struct{}

func (n *NumericFilterType) Supports(field model.Field) bool {
	switch field.Type {
	case model.FieldTypeInt, model.FieldTypeFloat:
		return true
	}
	return false
}

func (n *NumericFilterType) Operators(field model.Field) []string {
	return []string{"eq", "neq", "gt", "gte", "lt", "lte", "is_null", "is_not_null"}
}

func (n *NumericFilterType) NormalizeValue(value interface{}, field model.Field) interface{} {
	return value
}

func (n *NumericFilterType) WidgetKey() string { return "number" }

type BoolFilterType struct{}

func (b *BoolFilterType) Supports(field model.Field) bool {
	return field.Type == model.FieldTypeBool
}

func (b *BoolFilterType) Operators(field model.Field) []string {
	return []string{"eq", "neq"}
}

func (b *BoolFilterType) NormalizeValue(value interface{}, field model.Field) interface{} {
	return value
}

func (b *BoolFilterType) WidgetKey() string { return "boolean" }

type DateFilterType struct{}

func (d *DateFilterType) Supports(field model.Field) bool {
	switch field.Type {
	case model.FieldTypeDate, model.FieldTypeDateTime, model.FieldTypeTime:
		return true
	}
	return false
}

func (d *DateFilterType) Operators(field model.Field) []string {
	return []string{"eq", "neq", "gt", "gte", "lt", "lte", "is_null", "is_not_null"}
}

func (d *DateFilterType) NormalizeValue(value interface{}, field model.Field) interface{} {
	return value
}

func (d *DateFilterType) WidgetKey() string { return "date" }

type ChoiceFilterType struct{}

func (c *ChoiceFilterType) Supports(field model.Field) bool {
	return field.Type == model.FieldTypeChoice
}

func (c *ChoiceFilterType) Operators(field model.Field) []string {
	if field.Multiple {
		return []string{"in", "not_in", "is_null", "is_not_null"}
	}
	return []string{"eq", "neq", "is_null", "is_not_null"}
}

func (c *ChoiceFilterType) NormalizeValue(value interface{}, field model.Field) interface{} {
	return value
}

func (c *ChoiceFilterType) WidgetKey() string { return "choice" }

type RelationFilterType struct{}

func (r *RelationFilterType) Supports(field model.Field) bool {
	return field.Type == model.FieldTypeRelation
}

func (r *RelationFilterType) Operators(field model.Field) []string {
	if field.Multiple {
		return []string{"in", "not_in", "is_null", "is_not_null"}
	}
	return []string{"eq", "neq", "is_null", "is_not_null"}
}

func (r *RelationFilterType) NormalizeValue(value interface{}, field model.Field) interface{} {
	return value
}

func (r *RelationFilterType) WidgetKey() string { return "relation" }
