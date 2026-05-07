package twig

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"

	gmcore_crud "github.com/gmcorenet/sdk/gmcore-crud"
	gmcore_templating "github.com/gmcorenet/sdk/gmcore-templating"
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type CrudRuntime struct {
	registry      *registry.CrudRegistry
	engine        *gmcore_templating.Engine
	defaults      *model.ConfigDefaults
	theme         string
	bundleVersion string
}

func NewCrudRuntime(
	reg *registry.CrudRegistry,
	engine *gmcore_templating.Engine,
	defaults *model.ConfigDefaults,
	theme string,
) *CrudRuntime {
	return &CrudRuntime{
		registry:      reg,
		engine:        engine,
		defaults:      defaults,
		theme:         theme,
		bundleVersion: "1.0.0",
	}
}

func (r *CrudRuntime) ShowCrud(ctx context.Context, resource string, withAssets bool, userIdentifier string) template.HTML {
	if r.engine == nil {
		return template.HTML("")
	}

	crud, err := r.registry.Get(resource)
	if err != nil {
		return template.HTML("")
	}

	cfg := crud.Config()

	if !crud.Can("view", nil, nil) {
		return template.HTML("")
	}

	params := gmcore_crud.ListParams{
		Page:    1,
		PerPage: 25,
	}

	records, listErr := crud.List(params, nil)
	if listErr != nil {
		return template.HTML(fmt.Sprintf("<!-- CRUD list error: %v -->", listErr))
	}

	instanceID := resource + "-" + randomHex(6)

	payload := map[string]interface{}{
		"config":            cfg,
		"crudInstanceId":    instanceID,
		"params":            params,
		"result":            buildListResult(records, params),
		"filterSummaries":   []interface{}{},
		"crudBundleVersion": r.bundleVersion,
		"crudThemeVersion":  "gmcore dev",
		"crudTheme": map[string]interface{}{
			"rootClass":   "gmcrud-theme-gmcore",
			"footerLabel": "gmcore dev",
		},
		"crudConfigDefaults": r.defaults,
		"withAssets":         withAssets,
		"userIdentifier":     userIdentifier,
	}

	rendered, renderErr := r.engine.RenderContext(ctx, "crud/_partials/show_crud.html", payload)
	if renderErr != nil {
		return template.HTML(fmt.Sprintf("<!-- CRUD render error: %v -->", renderErr))
	}

	return template.HTML(rendered)
}

func buildListResult(records []gmcore_crud.Record, params gmcore_crud.ListParams) map[string]interface{} {
	displayRows := make([]map[string]interface{}, 0, len(records))
	for _, rec := range records {
		id, _ := rec["id"]
		cells := make([]map[string]interface{}, 0)
		for k, v := range rec {
			if k == "id" || k == "ID" {
				continue
			}
			cells = append(cells, map[string]interface{}{
				"field":    k,
				"value":    v,
				"labelKey": k,
			})
		}
		displayRows = append(displayRows, map[string]interface{}{
			"id":      id,
			"cells":   cells,
			"actions": []interface{}{},
		})
	}

	total := len(records)
	totalPages := (total + params.PerPage - 1) / params.PerPage
	if totalPages < 1 {
		totalPages = 1
	}

	return map[string]interface{}{
		"displayRows":   displayRows,
		"total":         total,
		"totalPages":    totalPages,
		"page":          params.Page,
		"perPage":       params.PerPage,
		"hasExactTotal": true,
		"hasMore":       false,
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
