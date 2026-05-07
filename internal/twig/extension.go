package twig

import (
	"context"
	"html/template"
)

type CrudExtension struct {
	runtime *CrudRuntime
}

func NewCrudExtension(runtime *CrudRuntime) *CrudExtension {
	return &CrudExtension{runtime: runtime}
}

func (e *CrudExtension) RegisterFuncs(ctx context.Context, funcs map[string]interface{}) {
	funcs["gmcore_crud"] = func(resource string, withAssets bool) template.HTML {
		return e.runtime.ShowCrud(ctx, resource, withAssets, "anonymous")
	}
	funcs["gmcore_crud_as"] = func(resource string, withAssets bool, userIdentifier string) template.HTML {
		return e.runtime.ShowCrud(ctx, resource, withAssets, userIdentifier)
	}
}
