package crud

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"

	gmcore_crud_sdk "github.com/gmcorenet/sdk/gmcore-crud"
	gmcore_bundle "github.com/gmcorenet/sdk/gmcore-bundle"
	gmcore_templating "github.com/gmcorenet/sdk/gmcore-templating"
	"github.com/gmcorenet/bundle-crud/internal/controller"
	"github.com/gmcorenet/bundle-crud/internal/export"
	"github.com/gmcorenet/bundle-crud/internal/filter"
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
	"github.com/gmcorenet/bundle-crud/internal/service"
)

var (
	Registry         *registry.CrudRegistry
	ConfigDefaults   *model.ConfigDefaults
	FilterRegistry   *filter.FilterTypeRegistry
	ExporterRegistry *export.ExportRegistry
	crudEngine       *gmcore_templating.Engine
)

func init() {
	Registry = registry.NewCrudRegistry()
	ConfigDefaults = model.NewConfigDefaults()
	FilterRegistry = filter.NewFilterTypeRegistry()
	ExporterRegistry = export.NewExportRegistry()

	gmcore_templating.RegisterFunc("gmcore_crud", func(resource string, withAssets bool) template.HTML {
		return showCrud(context.Background(), resource, withAssets, "anonymous")
	})
	gmcore_templating.RegisterFunc("gmcore_crud_as", func(resource string, withAssets bool, userIdentifier string) template.HTML {
		return showCrud(context.Background(), resource, withAssets, userIdentifier)
	})
}

type Bundle struct {
	gmcore_bundle.BaseBundle
	Registry         *registry.CrudRegistry
	ConfigDefaults   *model.ConfigDefaults
	FilterRegistry   *filter.FilterTypeRegistry
	ExporterRegistry *export.ExportRegistry
	controllers      *controller.CrudRouter
}

func NewBundle() *Bundle {
	return &Bundle{
		Registry:         Registry,
		ConfigDefaults:   ConfigDefaults,
		FilterRegistry:   FilterRegistry,
		ExporterRegistry: ExporterRegistry,
	}
}

func (b *Bundle) Name() string { return "crud-bundle" }

func (b *Bundle) Boot(ctx context.Context) error {
	controller.SetRegistry(Registry)
	controller.SetDefaults(ConfigDefaults)
	controller.SetFilterRegistry(FilterRegistry)

	b.controllers = controller.NewCrudRouter(
		Registry,
		service.NewListResponder(Registry),
		service.NewFormResponder(Registry),
		service.NewBulkActionService(Registry),
		service.NewServerActionService(Registry),
		service.NewRelationOptionsService(Registry),
		service.NewDesignerService(),
		service.NewDiagnosticsService(),
		ConfigDefaults,
		FilterRegistry,
	)

	return nil
}

func (b *Bundle) InitTemplating(engine *gmcore_templating.Engine) {
	crudEngine = engine
}

func (b *Bundle) Shutdown() error {
	return nil
}

func (b *Bundle) RegisterRoutes(mux *http.ServeMux) {
	b.controllers.Register(mux)
}

func (b *Bundle) RegisterResource(resource registry.CrudResource) error {
	return Registry.Register(resource)
}

func (b *Bundle) GetResource(name string) (registry.CrudResource, error) {
	return Registry.Get(name)
}

func (b *Bundle) AllResources() []registry.CrudResource {
	return Registry.All()
}

func showCrud(ctx context.Context, resource string, withAssets bool, userIdentifier string) template.HTML {
	if crudEngine == nil || Registry == nil {
		return template.HTML("")
	}

	crudRes, err := Registry.Get(resource)
	if err != nil {
		return template.HTML("")
	}

	cfg := crudRes.Config()

	if !crudRes.Can("view", nil, nil) {
		return template.HTML("")
	}

	params := gmcore_crud_sdk.ListParams{
		Page:    1,
		PerPage: 25,
	}

	records, listErr := crudRes.List(params, nil)
	if listErr != nil {
		return template.HTML(fmt.Sprintf("<!-- CRUD list error: %v -->", listErr))
	}

	instanceID := resource + "-" + randomHex(6)

	payload := map[string]interface{}{
		"config":            cfg,
		"crudInstanceId":    instanceID,
		"params":            params,
		"result":            buildCrudListResult(records, params),
		"filterSummaries":   []interface{}{},
		"crudBundleVersion": "1.0.0",
		"crudThemeVersion":  "gmcore dev",
		"crudTheme": map[string]interface{}{
			"rootClass":   "gmcrud-theme-gmcore",
			"footerLabel": "gmcore dev",
		},
		"crudConfigDefaults": ConfigDefaults,
		"withAssets":         withAssets,
		"userIdentifier":     userIdentifier,
	}

	rendered, renderErr := crudEngine.RenderContext(ctx, "crud/_partials/show_crud.html", payload)
	if renderErr != nil {
		return template.HTML(fmt.Sprintf("<!-- CRUD render error: %v -->", renderErr))
	}

	return template.HTML(rendered)
}

func buildCrudListResult(records []gmcore_crud_sdk.Record, params gmcore_crud_sdk.ListParams) map[string]interface{} {
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
