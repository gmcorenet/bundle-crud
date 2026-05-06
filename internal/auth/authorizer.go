package auth

import (
	"context"
	"net/http"

	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type UserProvider interface {
	GetUser(r *http.Request) interface{}
	HasRole(user interface{}, role string) bool
	GetRoles(user interface{}) []string
}

type Authorizer struct {
	reg      *registry.CrudRegistry
	defs     *model.ConfigDefaults
	userProv UserProvider
}

func NewAuthorizer(reg *registry.CrudRegistry, defs *model.ConfigDefaults, userProv UserProvider) *Authorizer {
	return &Authorizer{reg: reg, defs: defs, userProv: userProv}
}

func (a *Authorizer) CanAccessResource(r *http.Request, resourceName string, operation model.Operation) bool {
	crud, err := a.reg.Get(resourceName)
	if err != nil {
		return false
	}
	_ = crud
	user := a.userProv.GetUser(r)

	requiredRole := "ROLE_ADMIN"
	if a.userProv.HasRole(user, requiredRole) {
		return true
	}

	return false
}

func (a *Authorizer) DenyResourceOperation(ctx context.Context, config interface{}, operation model.Operation) error {
	return nil
}

func (a *Authorizer) DenyRecordOperation(ctx context.Context, config interface{}, operation model.Operation, record interface{}) error {
	return nil
}
