package service

import (
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
)

type DiagnosticsService struct{}

func NewDiagnosticsService() *DiagnosticsService {
	return &DiagnosticsService{}
}

func (d *DiagnosticsService) DebugPayload(config *model.ResourceConfig) map[string]interface{} {
	warnings := make([]string, 0)

	if len(config.Fields) == 0 {
		warnings = append(warnings, "No fields configured")
	}

	if config.Permissions.IsDisabled(model.OpView) {
		warnings = append(warnings, "View operation is disabled")
	}

	return map[string]interface{}{
		"name":           config.Name,
		"entity_class":    config.EntityClass,
		"fields_count":   len(config.Fields),
		"columns_count":  len(config.Columns),
		"relations_count": len(config.Relations),
		"permissions":    d.permissionSummary(&config.Permissions),
		"warnings":       warnings,
	}
}

func (d *DiagnosticsService) permissionSummary(perms *model.Permissions) map[string]string {
	return map[string]string{
		"view":   perms.ViewRole,
		"create": perms.CreateRole,
		"edit":   perms.EditRole,
		"delete": perms.DeleteRole,
		"export": perms.ExportRole,
		"design": perms.DesignRole,
		"debug":  perms.DebugRole,
	}
}

func (d *DiagnosticsService) ValidateResource(crud registry.CrudResource) []string {
	warnings := make([]string, 0)
	config := crud.Config()

	if config.Name == "" {
		warnings = append(warnings, "Resource has no name")
	}
	if config.PrimaryKey == "" {
		warnings = append(warnings, "Primary key not defined")
	}

	hasPrimaryField := false
	for _, f := range config.Fields {
		if f.Name == config.PrimaryKey {
			hasPrimaryField = true
			break
		}
	}
	if !hasPrimaryField {
		warnings = append(warnings, "Primary key field not found in field list")
	}

	return warnings
}
