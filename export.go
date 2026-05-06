package crud

import (
	"github.com/gmcorenet/bundle-crud/internal/export"
)

func NewExportRegistry() *export.ExportRegistry {
	return export.NewExportRegistry()
}
