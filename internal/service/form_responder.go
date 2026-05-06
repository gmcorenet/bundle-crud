package service

import (
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type FormResponder struct {
	reg *registry.CrudRegistry
}

func NewFormResponder(reg *registry.CrudRegistry) *FormResponder {
	return &FormResponder{reg: reg}
}

func (f *FormResponder) CreateForm(crud registry.CrudResource, record interface{}, mode string) *FormDefinition {
	config := crud.Config()
	fields := make([]FormField, 0)

	for _, field := range config.Fields {
		if !field.Editable && mode != "show" {
			continue
		}

		formField := FormField{
			Name:     field.Name,
			LabelKey: field.LabelKey,
			Type:     mapFieldType(field.Type),
			Required: field.Required,
		}

		if record != nil {
			if r, ok := record.(map[string]interface{}); ok {
				if val, exists := r[field.Name]; exists {
					formField.Value = val
				}
			}
		}

		fields = append(fields, formField)
	}

	return &FormDefinition{
		Name:   config.Name,
		Mode:   mode,
		Fields: fields,
		Title:  config.Label,
	}
}

func (f *FormResponder) AjaxForm(
	form *FormDefinition,
	config *model.ResourceConfig,
	mode string,
	record interface{},
	status int,
) map[string]interface{} {
	_ = config

	return map[string]interface{}{
		"ok":     status < 400,
		"form":   form,
		"mode":   mode,
		"record": record,
		"status": status,
	}
}

type FormDefinition struct {
	Name   string      `json:"name"`
	Mode   string      `json:"mode"`
	Fields []FormField `json:"fields"`
	Title  string      `json:"title"`
}

type FormField struct {
	Name     string      `json:"name"`
	LabelKey string      `json:"label_key"`
	Type     string      `json:"type"`
	Required bool        `json:"required"`
	Value    interface{} `json:"value,omitempty"`
	Options  interface{} `json:"options,omitempty"`
}

func mapFieldType(fieldType string) string {
	return fieldType
}
