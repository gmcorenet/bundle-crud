package crud

import (
	"context"
	"html/template"
	"net/http"

	gmcore_bundle "github.com/gmcorenet/sdk/gmcore-bundle"
	gmcore_templating "github.com/gmcorenet/sdk/gmcore-templating"
	"github.com/gmcorenet/bundle-crud/internal/controller"
	"github.com/gmcorenet/bundle-crud/internal/export"
	"github.com/gmcorenet/bundle-crud/internal/filter"
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
	"github.com/gmcorenet/bundle-crud/internal/service"
	"github.com/gmcorenet/bundle-crud/internal/twig"
)

var (
	Registry         *registry.CrudRegistry
	ConfigDefaults   *model.ConfigDefaults
	FilterRegistry   *filter.FilterTypeRegistry
	ExporterRegistry *export.ExportRegistry
	CrudTwigRuntime  *twig.CrudRuntime
	CrudTwigExtension *twig.CrudExtension
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
	twigRuntime      *twig.CrudRuntime
	twigExtension    *twig.CrudExtension
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
	b.twigRuntime = twig.NewCrudRuntime(Registry, engine, ConfigDefaults, "gmcore")
	b.twigExtension = twig.NewCrudExtension(b.twigRuntime)
	CrudTwigRuntime = b.twigRuntime
	CrudTwigExtension = b.twigExtension
}

func (b *Bundle) TemplateFunctions() map[string]interface{} {
	if b.twigRuntime == nil {
		return nil
	}
	funcs := make(map[string]interface{})
	funcs["gmcore_crud"] = func(resource string, withAssets bool) template.HTML {
		return b.twigRuntime.ShowCrud(context.Background(), resource, withAssets, "anonymous")
	}
	funcs["gmcore_crud_as"] = func(resource string, withAssets bool, userIdentifier string) template.HTML {
		return b.twigRuntime.ShowCrud(context.Background(), resource, withAssets, userIdentifier)
	}
	return funcs
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
