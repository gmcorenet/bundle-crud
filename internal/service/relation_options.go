package service

import (
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type RelationOptionsService struct {
	reg *registry.CrudRegistry
}

func NewRelationOptionsService(reg *registry.CrudRegistry) *RelationOptionsService {
	return &RelationOptionsService{reg: reg}
}

func (s *RelationOptionsService) Options(
	crud registry.CrudResource,
	relation string,
	query string,
	page int,
	limit int,
) ([]map[string]interface{}, bool, error) {
	options, hasMore, err := crud.RelationOptions(relation, query, page, limit)
	if err != nil {
		return nil, false, err
	}

	result := make([]map[string]interface{}, len(options))
	for i, opt := range options {
		result[i] = map[string]interface{}{
			"value": opt.Value,
			"label": opt.Label,
		}
	}

	return result, hasMore, nil
}
