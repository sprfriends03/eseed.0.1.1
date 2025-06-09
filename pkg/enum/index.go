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
	KindPermission       Kind = "permission"
	KindDataStatus       Kind = "data_status"
	KindDataAction       Kind = "data_action"
	KindMembershipStatus Kind = "membership_status"
	KindMembershipType   Kind = "membership_type"
	KindPlantSlotStatus  Kind = "plant_slot_status"
	// KindMemberStatus Kind = "member_status" // Reverted
	// KindKYCStatus    Kind = "kyc_status"    // Reverted
)

func Tags() map[string][]string {
	return map[string][]string{
		string(KindPermission):       gopkg.MapFunc(PermissionValues(), func(e Permission) string { return string(e) }),
		string(KindDataStatus):       gopkg.MapFunc(DataStatusValues(), func(e DataStatus) string { return string(e) }),
		string(KindDataAction):       gopkg.MapFunc(DataActionValues(), func(e DataAction) string { return string(e) }),
		string(KindMembershipStatus): gopkg.MapFunc(MembershipStatusValues(), func(e MembershipStatus) string { return string(e) }),
		string(KindMembershipType):   gopkg.MapFunc(MembershipTypeValues(), func(e MembershipType) string { return string(e) }),
		string(KindPlantSlotStatus):  gopkg.MapFunc(PlantSlotStatusValues(), func(e PlantSlotStatus) string { return string(e) }),
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

	// Self-service permissions
	PermissionUserViewSelf    Permission = "user_view_self"
	PermissionUserUpdateSelf  Permission = "user_update_self"
	PermissionUserDeleteSelf  Permission = "user_delete_self"
	PermissionUserPrivacySelf Permission = "user_privacy_self"

	// KYC permissions
	PermissionKYCView   Permission = "kyc_view"
	PermissionKYCVerify Permission = "kyc_verify"

	// Membership Management Permissions
	PermissionMembershipView   Permission = "membership_view"
	PermissionMembershipCreate Permission = "membership_create"
	PermissionMembershipUpdate Permission = "membership_update"
	PermissionMembershipDelete Permission = "membership_delete"
	PermissionMembershipRenew  Permission = "membership_renew"
	PermissionMembershipManage Permission = "membership_manage" // Admin-level permission

	// Plant Slot Management Permissions
	PermissionPlantSlotView     Permission = "plant_slot_view"
	PermissionPlantSlotCreate   Permission = "plant_slot_create"
	PermissionPlantSlotUpdate   Permission = "plant_slot_update"
	PermissionPlantSlotDelete   Permission = "plant_slot_delete"
	PermissionPlantSlotManage   Permission = "plant_slot_manage" // Admin-level
	PermissionPlantSlotTransfer Permission = "plant_slot_transfer"
	PermissionPlantSlotAssign   Permission = "plant_slot_assign"
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

		// Self-service permissions
		PermissionUserViewSelf,
		PermissionUserUpdateSelf,
		PermissionUserDeleteSelf,
		PermissionUserPrivacySelf,

		// KYC permissions
		PermissionKYCView,
		PermissionKYCVerify,

		// Membership Management Permissions
		PermissionMembershipView,
		PermissionMembershipCreate,
		PermissionMembershipUpdate,
		PermissionMembershipDelete,
		PermissionMembershipRenew,
		PermissionMembershipManage,

		// Plant Slot Management Permissions
		PermissionPlantSlotView,
		PermissionPlantSlotCreate,
		PermissionPlantSlotUpdate,
		PermissionPlantSlotDelete,
		PermissionPlantSlotManage,
		PermissionPlantSlotTransfer,
		PermissionPlantSlotAssign,
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

		// Self-service permissions
		PermissionUserViewSelf,
		PermissionUserUpdateSelf,
		PermissionUserDeleteSelf,
		PermissionUserPrivacySelf,

		// KYC permissions
		PermissionKYCView,
		PermissionKYCVerify,

		// Membership Management Permissions
		PermissionMembershipView,
		PermissionMembershipCreate,
		PermissionMembershipUpdate,
		PermissionMembershipDelete,
		PermissionMembershipRenew,
		PermissionMembershipManage,

		// Plant Slot Management Permissions
		PermissionPlantSlotView,
		PermissionPlantSlotCreate,
		PermissionPlantSlotUpdate,
		PermissionPlantSlotDelete,
		PermissionPlantSlotManage,
		PermissionPlantSlotTransfer,
		PermissionPlantSlotAssign,
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

// MembershipStatus represents the current status of a membership
type MembershipStatus string

const (
	MembershipStatusPending   MembershipStatus = "pending_payment"
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusExpired   MembershipStatus = "expired"
	MembershipStatusCanceled  MembershipStatus = "canceled"
	MembershipStatusSuspended MembershipStatus = "suspended"
)

func MembershipStatusValues() []MembershipStatus {
	return []MembershipStatus{
		MembershipStatusPending,
		MembershipStatusActive,
		MembershipStatusExpired,
		MembershipStatusCanceled,
		MembershipStatusSuspended,
	}
}

// MembershipType represents the type of membership
type MembershipType string

const (
	MembershipTypeBasic   MembershipType = "basic"
	MembershipTypePremium MembershipType = "premium"
	MembershipTypeVIP     MembershipType = "vip"
)

func MembershipTypeValues() []MembershipType {
	return []MembershipType{
		MembershipTypeBasic,
		MembershipTypePremium,
		MembershipTypeVIP,
	}
}

// PlantSlotStatus represents the current status of a plant slot
type PlantSlotStatus string

const (
	PlantSlotStatusAvailable    PlantSlotStatus = "available"
	PlantSlotStatusAllocated    PlantSlotStatus = "allocated"
	PlantSlotStatusOccupied     PlantSlotStatus = "occupied"
	PlantSlotStatusMaintenance  PlantSlotStatus = "maintenance"
	PlantSlotStatusOutOfService PlantSlotStatus = "out_of_service"
)

func PlantSlotStatusValues() []PlantSlotStatus {
	return []PlantSlotStatus{
		PlantSlotStatusAvailable,
		PlantSlotStatusAllocated,
		PlantSlotStatusOccupied,
		PlantSlotStatusMaintenance,
		PlantSlotStatusOutOfService,
	}
}

// KYCStatus enum and KYCStatusValues function fully removed as per instruction to revert.
