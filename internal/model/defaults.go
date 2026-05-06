package model

type ConfigDefaults struct {
	Version                    string            `yaml:"-"`
	DefaultTheme               string            `yaml:"default_theme"`
	PaginationDefaultPerPage   int               `yaml:"-"`
	PaginationAllowedPerPage   []int             `yaml:"-"`
	PaginationMaxPerPage       int               `yaml:"-"`
	PaginationDefaultMode      string            `yaml:"-"`
	PermissionDefaults         map[Operation]string `yaml:"-"`
	DebugEnabled               bool              `yaml:"-"`
	DebugShowPolicyReasons     bool              `yaml:"-"`
	FilterSearchDelayMs        int               `yaml:"-"`
	FilterMaxRelationOptPerPage int              `yaml:"-"`
	RelationPageSize           int               `yaml:"-"`
	RelationLoadAllLimit       int               `yaml:"-"`
	DesignerEnabled            bool              `yaml:"-"`
	ExportsEnabled             bool              `yaml:"-"`
	ExportsFormats             []string          `yaml:"-"`
	CSRFEnabled                bool              `yaml:"-"`

	raw map[string]interface{}
}

func NewConfigDefaults() *ConfigDefaults {
	return &ConfigDefaults{
		DefaultTheme:               "gmcore",
		PaginationDefaultPerPage:   25,
		PaginationAllowedPerPage:   []int{10, 25, 50, 100},
		PaginationMaxPerPage:       500,
		PaginationDefaultMode:      "exact",
		DebugEnabled:               true,
		DebugShowPolicyReasons:     true,
		FilterSearchDelayMs:        250,
		FilterMaxRelationOptPerPage: 20,
		RelationPageSize:           20,
		RelationLoadAllLimit:       500,
		DesignerEnabled:            true,
		ExportsEnabled:             true,
		ExportsFormats:             []string{"txt", "csv", "pdf"},
		CSRFEnabled:                true,
		PermissionDefaults: map[Operation]string{
			OpView:   "ROLE_ADMIN",
			OpCreate: Disabled,
			OpEdit:   Disabled,
			OpDelete: Disabled,
			OpExport: Disabled,
			OpDesign: Disabled,
			OpDebug:  Disabled,
		},
		raw: make(map[string]interface{}),
	}
}

func (d *ConfigDefaults) Apply(config *ResourceConfig) *ResourceConfig {
	if config.UsesGlobalDefaultPerPage {
		config.DefaultPerPage = d.PaginationDefaultPerPage
	}
	if config.UsesGlobalPerPageOptions {
		config.PerPageOptions = d.PaginationAllowedPerPage
	}
	if config.UsesGlobalPaginationMode {
		config.PaginationMode = PaginationMode(d.PaginationDefaultMode)
	}
	if config.UsesGlobalTheme {
		config.Theme = d.DefaultTheme
	}

	merged := *config
	merged.Features = make(map[string]bool)
	for k, v := range config.Features {
		merged.Features[k] = v
	}
	if _, ok := merged.Features["debug"]; !ok {
		merged.Features["debug"] = d.DebugEnabled
	}
	if _, ok := merged.Features["design"]; !ok {
		merged.Features["design"] = d.DesignerEnabled
	}
	if _, ok := merged.Features["filters"]; !ok {
		merged.Features["filters"] = true
	}
	if _, ok := merged.Features["bulk"]; !ok {
		merged.Features["bulk"] = true
	}
	if _, ok := merged.Features["pagination"]; !ok {
		merged.Features["pagination"] = true
	}
	if _, ok := merged.Features["per_page"]; !ok {
		merged.Features["per_page"] = true
	}
	if _, ok := merged.Features["sort"]; !ok {
		merged.Features["sort"] = true
	}
	if _, ok := merged.Features["search"]; !ok {
		merged.Features["search"] = true
	}

	perm := merged.Permissions
	if perm.UsesGlobalView {
		if role, ok := d.PermissionDefaults[OpView]; ok {
			perm.ViewRole = role
		}
	}
	if perm.UsesGlobalCreate {
		if role, ok := d.PermissionDefaults[OpCreate]; ok {
			perm.CreateRole = role
		}
	}
	if perm.UsesGlobalEdit {
		if role, ok := d.PermissionDefaults[OpEdit]; ok {
			perm.EditRole = role
		}
	}
	if perm.UsesGlobalDelete {
		if role, ok := d.PermissionDefaults[OpDelete]; ok {
			perm.DeleteRole = role
		}
	}
	if perm.UsesGlobalExport {
		if role, ok := d.PermissionDefaults[OpExport]; ok {
			perm.ExportRole = role
		}
	}
	if perm.UsesGlobalDesign {
		if role, ok := d.PermissionDefaults[OpDesign]; ok {
			perm.DesignRole = role
		}
	}
	if perm.UsesGlobalDebug {
		if role, ok := d.PermissionDefaults[OpDebug]; ok {
			perm.DebugRole = role
		}
	}
	merged.Permissions = perm

	return &merged
}

func (d *ConfigDefaults) SearchDelayMs() int {
	return d.FilterSearchDelayMs
}

func (d *ConfigDefaults) MaxRelationOptionsPerPage() int {
	return d.FilterMaxRelationOptPerPage
}

func (d *ConfigDefaults) GetRelationPageSize() int {
	return d.RelationPageSize
}
