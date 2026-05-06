package service

import (
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type ServerActionService struct {
	reg *registry.CrudRegistry
}

func NewServerActionService(reg *registry.CrudRegistry) *ServerActionService {
	return &ServerActionService{reg: reg}
}

func (s *ServerActionService) Handle(
	crud registry.CrudResource,
	action string,
	record interface{},
	user interface{},
	data map[string]interface{},
) (interface{}, error) {
	id := crud.GetRecordID(record.(map[string]interface{}))
	return crud.Bulk(action, []string{id}, user)
}

func (s *ServerActionService) FilterValidatedFormData(formFields interface{}, data map[string]interface{}) map[string]interface{} {
	return data
}
