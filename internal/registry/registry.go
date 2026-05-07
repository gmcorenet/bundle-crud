package registry

import (
	"fmt"
	"sort"
	"sync"

	gmcore_crud "github.com/gmcorenet/sdk-gmcore-crud"
)

type CrudResource interface {
	Name() string
	Config() gmcore_crud.Config
	List(params gmcore_crud.ListParams, user interface{}) ([]gmcore_crud.Record, error)
	Find(id string, user interface{}) (gmcore_crud.Record, error)
	FindForUser(id string, user interface{}) (gmcore_crud.Record, error)
	IsRecordVisible(id string, user interface{}) bool
	Create(data map[string]interface{}, user interface{}) (gmcore_crud.Record, error)
	Update(record gmcore_crud.Record, data map[string]interface{}, user interface{}) error
	Delete(record gmcore_crud.Record, user interface{}) error
	Bulk(action string, ids []string, user interface{}) (interface{}, error)
	RelationOptions(relation, query string, page, limit int) ([]gmcore_crud.RelationOption, bool, error)
	RelationOptionLabels(relation string, values []string) (map[string]string, error)
	GetRecordID(record gmcore_crud.Record) string
	Can(operation string, record interface{}, user interface{}) bool
	ResolveRowActions(record gmcore_crud.Record) ([]gmcore_crud.Action, error)
}

type CrudRegistry struct {
	mu    sync.RWMutex
	items map[string]CrudResource
}

func NewCrudRegistry() *CrudRegistry {
	return &CrudRegistry{
		items: make(map[string]CrudResource),
	}
}

func (r *CrudRegistry) Register(resource CrudResource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := resource.Name()
	if _, exists := r.items[name]; exists {
		return fmt.Errorf("crud: duplicate resource name %q", name)
	}
	r.items[name] = resource
	return nil
}

func (r *CrudRegistry) Get(name string) (CrudResource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, ok := r.items[name]
	if !ok {
		return nil, fmt.Errorf("crud: unknown resource %q", name)
	}
	return res, nil
}

func (r *CrudRegistry) All() []CrudResource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.items))
	for name := range r.items {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]CrudResource, 0, len(names))
	for _, name := range names {
		result = append(result, r.items[name])
	}
	return result
}

func (r *CrudRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.items[name]
	return ok
}

func (r *CrudRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}
