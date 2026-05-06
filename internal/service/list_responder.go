package service

import (
	gmcore_crud "github.com/gmcorenet/sdk/gmcore-crud"
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type ListResponder struct {
	reg *registry.CrudRegistry
}

func NewListResponder(reg *registry.CrudRegistry) *ListResponder {
	return &ListResponder{reg: reg}
}

func (l *ListResponder) IndexVariables(
	crud registry.CrudResource,
	params gmcore_crud.ListParams,
	user interface{},
	isAjax bool,
) map[string]interface{} {
	records, err := crud.List(params, user)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	total := len(records)
	return map[string]interface{}{
		"records":     records,
		"total":       total,
		"page":        params.Page,
		"per_page":    params.PerPage,
		"total_pages": (total + params.PerPage - 1) / params.PerPage,
		"params":      params,
		"ajax":        isAjax,
	}
}

func (l *ListResponder) BuildDisplayRows(records []gmcore_crud.Record, config *model.ResourceConfig) []map[string]interface{} {
	rows := make([]map[string]interface{}, len(records))
	for i, record := range records {
		cells := make([]map[string]interface{}, len(config.Columns))
		for j, col := range config.Columns {
			val, _ := record[col.Field]
			cells[j] = map[string]interface{}{
				"field":    col.Field,
				"label":    col.LabelKey,
				"value":    val,
			}
		}
		rows[i] = map[string]interface{}{
			"id":     record["id"],
			"cells":  cells,
		}
	}
	return rows
}
