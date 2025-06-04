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
	// KindMemberStatus Kind = "member_status" // Reverted
	// KindKYCStatus    Kind = "kyc_status"    // Reverted
)

func Tags() map[string][]string {
	return map[string][]string{
		string(KindPermission): gopkg.MapFunc(PermissionValues(), func(e Permission) string { return string(e) }),
		string(KindDataStatus): gopkg.MapFunc(DataStatusValues(), func(e DataStatus) string { return string(e) }),
		string(KindDataAction): gopkg.MapFunc(DataActionValues(), func(e DataAction) string { return string(e) }),
		// string(KindMemberStatus): gopkg.MapFunc(MemberStatusValues(), func(e MemberStatus) string { return string(e) }), // Reverted
		// string(KindKYCStatus):    gopkg.MapFunc(KYCStatusValues(), func(e KYCStatus) string { return string(e) }),       // Reverted
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

// PlantStatus represents the current status of a plant
type PlantStatus string

const (
	PlantStatusGrowing   PlantStatus = "growing"
	PlantStatusFlowering PlantStatus = "flowering"
	PlantStatusHarvested PlantStatus = "harvested"
	PlantStatusDormant   PlantStatus = "dormant"
	PlantStatusDiseased  PlantStatus = "diseased"
	PlantStatusDead      PlantStatus = "dead"
)

func PlantStatusValues() []PlantStatus {
	return []PlantStatus{
		PlantStatusGrowing,
		PlantStatusFlowering,
		PlantStatusHarvested,
		PlantStatusDormant,
		PlantStatusDiseased,
		PlantStatusDead,
	}
}

// NotificationStatus represents the current status of a notification
type NotificationStatus string

const (
	NotificationStatusUnread NotificationStatus = "unread"
	NotificationStatusRead   NotificationStatus = "read"
)

func NotificationStatusValues() []NotificationStatus {
	return []NotificationStatus{NotificationStatusUnread, NotificationStatusRead}
}

// MemberStatus represents the current status of a club member
type MemberStatus string

const (
	// MemberStatusPendingVerification MemberStatus = "pending_verification" // Reverted
	// MemberStatusPendingApproval     MemberStatus = "pending_approval"     // Reverted
	MemberStatusActive     MemberStatus = "active"
	MemberStatusInactive   MemberStatus = "inactive"
	MemberStatusSuspended  MemberStatus = "suspended"
	MemberStatusTerminated MemberStatus = "terminated"
)

func MemberStatusValues() []MemberStatus {
	return []MemberStatus{
		// MemberStatusPendingVerification, // Reverted
		// MemberStatusPendingApproval,     // Reverted
		MemberStatusActive,
		MemberStatusInactive,
		MemberStatusSuspended,
		MemberStatusTerminated,
	}
}

// KYCStatus enum and KYCStatusValues function fully removed as per instruction to revert.
