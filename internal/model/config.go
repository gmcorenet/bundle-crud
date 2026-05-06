package model

import (
	"sort"
	"strconv"
	"strings"
)

type PaginationMode string

const (
	PaginationExact  PaginationMode = "exact"
	PaginationHasMore PaginationMode = "has_more"
)

type IndexColumn struct {
	Field          string `json:"field" yaml:"field"`
	LabelKey       string `json:"label_key" yaml:"label_key"`
	Sortable       bool   `json:"sortable" yaml:"sortable"`
	Visible        bool   `json:"visible" yaml:"visible"`
	FormatTemplate string `json:"format_template" yaml:"format_template"`
}

type ActionKind string

const (
	ActionKindDefault  ActionKind = "default"
	ActionKindDanger   ActionKind = "danger"
	ActionKindDownload ActionKind = "download"
)

type Action struct {
	Name           string            `json:"name" yaml:"name"`
	LabelKey       string            `json:"label_key" yaml:"label_key"`
	Kind           ActionKind        `json:"kind" yaml:"kind"`
	Method         string            `json:"method" yaml:"method"`
	Order          int               `json:"order" yaml:"order"`
	ConfirmKey     string            `json:"confirm_key" yaml:"confirm_key"`
	PermissionKey  string            `json:"permission_key" yaml:"permission_key"`
	FormFields     []Field           `json:"form_fields" yaml:"form_fields"`
	Icon           string            `json:"icon" yaml:"icon"`
	VisibleWhen    string            `json:"visible_when" yaml:"visible_when"`
	Metadata       map[string]interface{} `json:"metadata" yaml:"metadata"`
}

type Relation struct {
	Name             string `json:"name" yaml:"name"`
	LabelKey         string `json:"label_key" yaml:"label_key"`
	TargetEntity     string `json:"target_entity" yaml:"target_entity"`
	LocalField       string `json:"local_field" yaml:"local_field"`
	ForeignKey       string `json:"foreign_key" yaml:"foreign_key"`
	ValueField       string `json:"value_field" yaml:"value_field"`
	DisplayField     string `json:"display_field" yaml:"display_field"`
	Async            bool   `json:"async" yaml:"async"`
	AsyncDebounce    int    `json:"async_debounce" yaml:"async_debounce"`
	Limit            int    `json:"limit" yaml:"limit"`
}

type ResourceConfig struct {
	Name              string            `json:"name" yaml:"name"`
	TitleKey          string            `json:"title_key" yaml:"title_key"`
	PrimaryKey        string            `json:"primary_key" yaml:"primary_key"`
	PrimaryKeyType    string            `json:"primary_key_type" yaml:"primary_key_type"`
	EntityClass       string            `json:"entity_class" yaml:"entity_class"`
	Fields            []Field           `json:"fields" yaml:"fields"`
	Columns           []IndexColumn     `json:"columns" yaml:"columns"`
	RowActions        []Action          `json:"row_actions" yaml:"row_actions"`
	BulkActions       []Action          `json:"bulk_actions" yaml:"bulk_actions"`
	Relations         []Relation        `json:"relations" yaml:"relations"`
	Features          map[string]bool   `json:"features" yaml:"features"`
	Permissions       Permissions       `json:"-" yaml:"-"`
	DefaultPerPage    int               `json:"default_per_page" yaml:"default_per_page"`
	PerPageOptions    []int             `json:"per_page_options" yaml:"per_page_options"`
	SearchPlaceholder string            `json:"search_placeholder" yaml:"search_placeholder"`
	PaginationMode    PaginationMode    `json:"pagination_mode" yaml:"pagination_mode"`
	Theme             string            `json:"theme" yaml:"theme"`
	Heavy             bool              `json:"heavy" yaml:"heavy"`
	Scope             func(interface{}) interface{} `json:"-" yaml:"-"`
	RowPolicy         func(record interface{}, user interface{}) bool `json:"-" yaml:"-"`
	Hooks             map[string][]func(record interface{}) error `json:"-" yaml:"-"`
	ServerActions     map[string]func(record interface{}, data map[string]interface{}, user interface{}) (interface{}, error) `json:"-" yaml:"-"`
	FormClientHooks   map[string]string `json:"form_client_hooks" yaml:"form_client_hooks"`

	UsesGlobalDefaultPerPage    bool `json:"-" yaml:"-"`
	UsesGlobalPerPageOptions    bool `json:"-" yaml:"-"`
	UsesGlobalPaginationMode    bool `json:"-" yaml:"-"`
	UsesGlobalTheme             bool `json:"-" yaml:"-"`
}

func NewResourceConfig(name string) *ResourceConfig {
	return &ResourceConfig{
		Name:           name,
		PrimaryKey:     "id",
		PrimaryKeyType: "int",
		Fields:         make([]Field, 0),
		Columns:        make([]IndexColumn, 0),
		RowActions:     make([]Action, 0),
		BulkActions:    make([]Action, 0),
		Relations:      make([]Relation, 0),
		Features:       make(map[string]bool),
		Permissions:    NewPermissions().Build(),
		PerPageOptions: make([]int, 0),
		PaginationMode: PaginationExact,
		Theme:          "gmcore",
		Hooks:          make(map[string][]func(record interface{}) error),
		ServerActions:  make(map[string]func(record interface{}, data map[string]interface{}, user interface{}) (interface{}, error)),
		FormClientHooks: make(map[string]string),
		UsesGlobalDefaultPerPage: true,
		UsesGlobalPerPageOptions: true,
		UsesGlobalPaginationMode: true,
		UsesGlobalTheme:          true,
	}
}

func (c *ResourceConfig) ResolvedRowActions() []Action {
	actions := make(map[string]Action)
	for _, a := range c.RowActions {
		actions[a.Name] = a
	}

	builtIns := []struct {
		name      string
		operation Operation
		action    Action
	}{
		{"view", OpView, Action{Name: "view", LabelKey: "crud.action.view", Order: 10}},
		{"edit", OpEdit, Action{Name: "edit", LabelKey: "crud.action.edit", Order: 20}},
		{"clone", OpCreate, Action{Name: "clone", LabelKey: "crud.action.clone", Order: 30}},
		{"delete", OpDelete, Action{Name: "delete", LabelKey: "crud.action.delete", Kind: ActionKindDanger, Method: "POST", Order: 40, ConfirmKey: "crud.confirm.delete"}},
	}

	for _, bi := range builtIns {
		if c.Permissions.IsEnabled(bi.operation) {
			if _, exists := actions[bi.name]; !exists {
				actions[bi.name] = bi.action
			}
		} else {
			delete(actions, bi.name)
		}
	}

	result := make([]Action, 0, len(actions))
	for _, a := range actions {
		result = append(result, a)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func (c *ResourceConfig) ResolvedBulkActions() []Action {
	actions := make(map[string]Action)
	for _, a := range c.BulkActions {
		actions[a.Name] = a
	}

	if c.Permissions.IsEnabled(OpExport) {
		if _, exists := actions["export_txt"]; !exists {
			actions["export_txt"] = Action{Name: "export_txt", LabelKey: "crud.action.export_txt", Kind: ActionKindDownload, Method: "POST", Order: 100}
		}
		if _, exists := actions["export_csv"]; !exists {
			actions["export_csv"] = Action{Name: "export_csv", LabelKey: "crud.action.export_csv", Kind: ActionKindDownload, Method: "POST", Order: 110}
		}
		if _, exists := actions["export_pdf"]; !exists {
			actions["export_pdf"] = Action{Name: "export_pdf", LabelKey: "crud.action.export_pdf", Kind: ActionKindDownload, Method: "POST", Order: 120}
		}
	} else {
		for name := range actions {
			if strings.HasPrefix(name, "export_") {
				delete(actions, name)
			}
		}
	}

	if c.Permissions.IsEnabled(OpDelete) {
		if _, exists := actions["delete_bulk"]; !exists {
			actions["delete_bulk"] = Action{Name: "delete_bulk", LabelKey: "crud.action.delete_bulk", Kind: ActionKindDanger, Method: "POST", Order: 900, ConfirmKey: "crud.confirm.bulk_message"}
		}
	} else {
		delete(actions, "delete_bulk")
	}

	result := make([]Action, 0, len(actions))
	for _, a := range actions {
		result = append(result, a)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func (c *ResourceConfig) FieldByName(name string) (Field, bool) {
	for _, f := range c.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

func (c *ResourceConfig) RelationByName(name string) (Relation, bool) {
	for _, r := range c.Relations {
		if r.Name == name {
			return r, true
		}
	}
	return Relation{}, false
}

func (c *ResourceConfig) FeatureEnabled(feature string) bool {
	if enabled, ok := c.Features[feature]; ok {
		return enabled
	}
	switch feature {
	case "debug", "design", "filters", "bulk", "pagination", "per_page", "sort", "search":
		return true
	default:
		return false
	}
}

type ListParams struct {
	Page            int
	PerPage         int
	Search          string
	Sort            []string
	Filters         map[string]string
	AdvancedFilters map[string]AdvancedFilter
}

type AdvancedFilter struct {
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func ParsePerPage(value string, defaultVal int, options []int) int {
	if value == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	for _, opt := range options {
		if n == opt {
			return n
		}
	}
	return defaultVal
}

func ParsePage(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 1
	}
	return n
}
