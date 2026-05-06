package crud

import (
	"context"
	"net/http"

	gmcore_bundle "github.com/gmcorenet/sdk/gmcore-bundle"
	"github.com/gmcorenet/bundle-crud/internal/controller"
	"github.com/gmcorenet/bundle-crud/internal/export"
	"github.com/gmcorenet/bundle-crud/internal/filter"
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
	"github.com/gmcorenet/bundle-crud/internal/service"
)

var (
	Registry        *registry.CrudRegistry
	ConfigDefaults  *model.ConfigDefaults
	FilterRegistry  *filter.FilterTypeRegistry
	ExporterRegistry *export.ExportRegistry
)

func init() {
	Registry = registry.NewCrudRegistry()
	ConfigDefaults = model.NewConfigDefaults()
	FilterRegistry = filter.NewFilterTypeRegistry()
	ExporterRegistry = export.NewExportRegistry()
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
