package service

import (
	"fmt"
	"github.com/gmcorenet/bundle-crud/internal/export"
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type BulkActionService struct {
	reg *registry.CrudRegistry
}

func NewBulkActionService(reg *registry.CrudRegistry) *BulkActionService {
	return &BulkActionService{reg: reg}
}

func (b *BulkActionService) DownloadExport(resource, action, content string) (*export.ExportResult, error) {
	exporter, ok := export.GlobalRegistry().Get(action)
	if !ok {
		return nil, fmt.Errorf("unknown export format: %s", action)
	}
	return exporter.Export(export.ExportRequest{
		Resource: resource,
		Action:   action,
		Content:  content,
	})
}
