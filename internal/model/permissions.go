package model

import (
	"strings"
)

const Disabled = "__disabled__"

type Operation string

const (
	OpView   Operation = "view"
	OpCreate Operation = "create"
	OpEdit   Operation = "edit"
	OpDelete Operation = "delete"
	OpExport Operation = "export"
	OpDesign Operation = "design"
	OpDebug  Operation = "debug"
)

var AllOperations = []Operation{OpView, OpCreate, OpEdit, OpDelete, OpExport, OpDesign, OpDebug}

type Permissions struct {
	ViewRole        string
	CreateRole      string
	EditRole        string
	DeleteRole      string
	ExportRole      string
	DesignRole      string
	DebugRole       string
	CustomRoles     map[Operation]string

	CanView   func(op Operation, record, user interface{}) bool
	CanCreate func(op Operation, record, user interface{}) bool
	CanEdit   func(op Operation, record, user interface{}) bool
	CanDelete func(op Operation, record, user interface{}) bool
	CanExport func(op Operation, record, user interface{}) bool
	CanDesign func(op Operation, record, user interface{}) bool
	CanDebug  func(op Operation, record, user interface{}) bool

	UsesGlobalView   bool
	UsesGlobalCreate bool
	UsesGlobalEdit   bool
	UsesGlobalDelete bool
	UsesGlobalExport bool
	UsesGlobalDesign bool
	UsesGlobalDebug  bool
}

func (p *Permissions) RoleFor(op Operation) string {
	switch op {
	case OpView:
		return p.ViewRole
	case OpCreate:
		return p.CreateRole
	case OpEdit:
		return p.EditRole
	case OpDelete:
		return p.DeleteRole
	case OpExport:
		return p.ExportRole
	case OpDesign:
		return p.DesignRole
	case OpDebug:
		return p.DebugRole
	default:
		if role, ok := p.CustomRoles[op]; ok {
			return role
		}
		return Disabled
	}
}

func (p *Permissions) IsDisabled(op Operation) bool {
	return p.RoleFor(op) == Disabled
}

func (p *Permissions) IsEnabled(op Operation) bool {
	return !p.IsDisabled(op)
}

type PermissionsBuilder struct {
	p Permissions
}

func NewPermissions() *PermissionsBuilder {
	return &PermissionsBuilder{
		p: Permissions{
			ViewRole:        "ROLE_USER",
			CreateRole:      "ROLE_USER",
			EditRole:        "ROLE_USER",
			DeleteRole:      "ROLE_ADMIN",
			ExportRole:      Disabled,
			DesignRole:      "ROLE_ROOT",
			DebugRole:       "ROLE_ROOT",
			CustomRoles:     make(map[Operation]string),
			UsesGlobalView:   true,
			UsesGlobalCreate: true,
			UsesGlobalEdit:   true,
			UsesGlobalDelete: true,
			UsesGlobalExport: true,
			UsesGlobalDesign: true,
			UsesGlobalDebug:  true,
		},
	}
}

func (b *PermissionsBuilder) ViewRole(role string) *PermissionsBuilder {
	b.p.ViewRole = normalizeRole(role)
	b.p.UsesGlobalView = false
	return b
}

func (b *PermissionsBuilder) CreateRole(role string) *PermissionsBuilder {
	b.p.CreateRole = normalizeRole(role)
	b.p.UsesGlobalCreate = false
	return b
}

func (b *PermissionsBuilder) EditRole(role string) *PermissionsBuilder {
	b.p.EditRole = normalizeRole(role)
	b.p.UsesGlobalEdit = false
	return b
}

func (b *PermissionsBuilder) DeleteRole(role string) *PermissionsBuilder {
	b.p.DeleteRole = normalizeRole(role)
	b.p.UsesGlobalDelete = false
	return b
}

func (b *PermissionsBuilder) ExportRole(role string) *PermissionsBuilder {
	b.p.ExportRole = normalizeRole(role)
	b.p.UsesGlobalExport = false
	return b
}

func (b *PermissionsBuilder) DesignRole(role string) *PermissionsBuilder {
	b.p.DesignRole = normalizeRole(role)
	b.p.UsesGlobalDesign = false
	return b
}

func (b *PermissionsBuilder) DebugRole(role string) *PermissionsBuilder {
	b.p.DebugRole = normalizeRole(role)
	b.p.UsesGlobalDebug = false
	return b
}

func (b *PermissionsBuilder) CustomRole(op Operation, role string) *PermissionsBuilder {
	b.p.CustomRoles[op] = normalizeRole(role)
	return b
}

func (b *PermissionsBuilder) OnCanView(fn func(op Operation, record, user interface{}) bool) *PermissionsBuilder {
	b.p.CanView = fn
	return b
}

func (b *PermissionsBuilder) OnCanCreate(fn func(op Operation, record, user interface{}) bool) *PermissionsBuilder {
	b.p.CanCreate = fn
	return b
}

func (b *PermissionsBuilder) OnCanEdit(fn func(op Operation, record, user interface{}) bool) *PermissionsBuilder {
	b.p.CanEdit = fn
	return b
}

func (b *PermissionsBuilder) OnCanDelete(fn func(op Operation, record, user interface{}) bool) *PermissionsBuilder {
	b.p.CanDelete = fn
	return b
}

func (b *PermissionsBuilder) OnCanExport(fn func(op Operation, record, user interface{}) bool) *PermissionsBuilder {
	b.p.CanExport = fn
	return b
}

func (b *PermissionsBuilder) OnCanDesign(fn func(op Operation, record, user interface{}) bool) *PermissionsBuilder {
	b.p.CanDesign = fn
	return b
}

func (b *PermissionsBuilder) OnCanDebug(fn func(op Operation, record, user interface{}) bool) *PermissionsBuilder {
	b.p.CanDebug = fn
	return b
}

func (b *PermissionsBuilder) Build() Permissions {
	return b.p
}

func normalizeRole(role string) string {
	role = strings.TrimSpace(role)
	if role == "" || role == "null" {
		return Disabled
	}
	return role
}
