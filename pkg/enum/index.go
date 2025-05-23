package enum

import (
	"slices"

	"github.com/nhnghia272/gopkg"
)

// Tenant
type Tenant string

// Kind
type Kind string

const (
	KindPermission Kind = "permission"
	KindDataStatus Kind = "data_status"
	KindDataAction Kind = "data_action"
)

func Tags() map[string][]string {
	return map[string][]string{
		string(KindPermission): gopkg.MapFunc(PermissionValues(), func(e Permission) string { return string(e) }),
		string(KindDataStatus): gopkg.MapFunc(DataStatusValues(), func(e DataStatus) string { return string(e) }),
		string(KindDataAction): gopkg.MapFunc(DataActionValues(), func(e DataAction) string { return string(e) }),
	}
}

// Permission
type Permission string

const (
	PermissionSystemSetting  Permission = "system_setting"
	PermissionSystemAuditLog Permission = "system_audit_log"

	PermissionClientView   Permission = "client_view"
	PermissionClientCreate Permission = "client_create"
	PermissionClientDelete Permission = "client_delete"

	PermissionRoleView   Permission = "role_view"
	PermissionRoleCreate Permission = "role_create"
	PermissionRoleUpdate Permission = "role_update"

	PermissionUserView   Permission = "user_view"
	PermissionUserCreate Permission = "user_create"
	PermissionUserUpdate Permission = "user_update"

	PermissionTenantView   Permission = "tenant_view"
	PermissionTenantCreate Permission = "tenant_create"
	PermissionTenantUpdate Permission = "tenant_update"
)

func PermissionTenantValues() []Permission {
	permissions := []Permission{
		PermissionSystemSetting,
		PermissionSystemAuditLog,

		PermissionClientView,
		PermissionClientCreate,
		PermissionClientDelete,

		PermissionRoleView,
		PermissionRoleCreate,
		PermissionRoleUpdate,

		PermissionUserView,
		PermissionUserCreate,
		PermissionUserUpdate,
	}
	return gopkg.UniqueFunc(slices.Sorted(slices.Values(permissions)), func(e Permission) Permission { return e })
}

func PermissionRootValues() []Permission {
	permissions := []Permission{
		PermissionSystemSetting,
		PermissionSystemAuditLog,

		PermissionClientView,
		PermissionClientCreate,
		PermissionClientDelete,

		PermissionRoleView,
		PermissionRoleCreate,
		PermissionRoleUpdate,

		PermissionUserView,
		PermissionUserCreate,
		PermissionUserUpdate,

		PermissionTenantView,
		PermissionTenantCreate,
		PermissionTenantUpdate,
	}
	return gopkg.UniqueFunc(slices.Sorted(slices.Values(permissions)), func(e Permission) Permission { return e })
}

func PermissionValues() []Permission {
	permissions := slices.Concat(PermissionRootValues(), PermissionTenantValues())
	return gopkg.UniqueFunc(slices.Sorted(slices.Values(permissions)), func(e Permission) Permission { return e })
}

// DataStatus
type DataStatus string

const (
	DataStatusEnable  DataStatus = "enable"
	DataStatusDisable DataStatus = "disable"
)

func DataStatusValues() []DataStatus {
	return []DataStatus{DataStatusEnable, DataStatusDisable}
}

// DataAction
type DataAction string

const (
	DataActionCreate        DataAction = "create"
	DataActionUpdate        DataAction = "update"
	DataActionDelete        DataAction = "delete"
	DataActionResetPassword DataAction = "reset_password"
)

func DataActionValues() []DataAction {
	return []DataAction{DataActionCreate, DataActionUpdate, DataActionDelete, DataActionResetPassword}
}
