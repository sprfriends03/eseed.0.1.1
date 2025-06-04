package ecode

import (
	"net/http"
)

// Cannabis Club Specific Error Codes
var (
	// KYC related errors
	KYCVerificationRequired = New(http.StatusForbidden, "kyc_verification_required")
	KYCRejected             = New(http.StatusForbidden, "kyc_rejected")
	KYCDocumentInvalid      = New(http.StatusBadRequest, "kyc_document_invalid")

	// Membership related errors
	MembershipExpired       = New(http.StatusForbidden, "membership_expired")
	MembershipLimitExceeded = New(http.StatusForbidden, "membership_limit_exceeded")
	MembershipRequired      = New(http.StatusForbidden, "membership_required")
	MembershipNotActive     = New(http.StatusForbidden, "membership_not_active")

	// Plant slot related errors
	PlantSlotUnavailable    = New(http.StatusConflict, "plant_slot_unavailable")
	PlantSlotAlreadyUsed    = New(http.StatusConflict, "plant_slot_already_used")
	PlantSlotNotFound       = New(http.StatusNotFound, "plant_slot_not_found")
	PlantSlotAllocationFull = New(http.StatusConflict, "plant_slot_allocation_full")

	// Plant related errors
	PlantNotFound           = New(http.StatusNotFound, "plant_not_found")
	PlantAlreadyHarvested   = New(http.StatusConflict, "plant_already_harvested")
	PlantInvalidStatus      = New(http.StatusBadRequest, "plant_invalid_status")
	PlantCareRecordNotFound = New(http.StatusNotFound, "plant_care_record_not_found")
	PlantImageUploadFailed  = New(http.StatusInternalServerError, "plant_image_upload_failed")
	PlantLimitExceeded      = New(http.StatusForbidden, "plant_limit_exceeded")

	// Harvest related errors
	HarvestNotReady         = New(http.StatusConflict, "harvest_not_ready")
	HarvestNotFound         = New(http.StatusNotFound, "harvest_not_found")
	HarvestAlreadyCollected = New(http.StatusConflict, "harvest_already_collected")
	HarvestWeightRequired   = New(http.StatusBadRequest, "harvest_weight_required")

	// NFT related errors
	NFTMintingFailed         = New(http.StatusInternalServerError, "nft_minting_failed")
	NFTAlreadyMinted         = New(http.StatusConflict, "nft_already_minted")
	NFTVerificationFailed    = New(http.StatusBadRequest, "nft_verification_failed")
	NFTOwnershipVerifyFailed = New(http.StatusForbidden, "nft_ownership_verify_failed")

	// Payment related errors
	PaymentProcessingError = New(http.StatusPaymentRequired, "payment_processing_error")
	PaymentRequired        = New(http.StatusPaymentRequired, "payment_required")
	PaymentFailed          = New(http.StatusPaymentRequired, "payment_failed")
	PaymentExpired         = New(http.StatusPaymentRequired, "payment_expired")

	// Validation errors
	ValidationError      = New(http.StatusBadRequest, "validation_error")
	DateRangeInvalid     = New(http.StatusBadRequest, "date_range_invalid")
	RequiredFieldMissing = New(http.StatusBadRequest, "required_field_missing")
	InvalidDataFormat    = New(http.StatusBadRequest, "invalid_data_format")
)

// Cannabis club error codes - 3000 to 3999
const (
	// Member errors (3000-3099)
	ErrMemberNotFound         = 3000 // Member not found
	ErrMemberAlreadyExists    = 3001 // Member already exists
	ErrMemberEmailInvalid     = 3002 // Invalid email address
	ErrMemberPhoneInvalid     = 3003 // Invalid phone number
	ErrMemberMedicalIDExpired = 3004 // Medical ID is expired
	ErrMemberMedicalIDInvalid = 3005 // Invalid medical ID
	ErrMemberInactive         = 3006 // Member is inactive
	ErrMemberSuspended        = 3007 // Member is suspended
	ErrMemberAgeRestriction   = 3008 // Member does not meet age requirements
	ErrMemberAddressInvalid   = 3009 // Invalid member address

	// Membership errors (3100-3199)
	ErrMembershipNotFound        = 3100 // Membership not found
	ErrMembershipExpired         = 3101 // Membership is expired
	ErrMembershipPaymentFailed   = 3102 // Membership payment failed
	ErrMembershipLimit           = 3103 // Member has reached membership limit
	ErrMembershipTypeInvalid     = 3104 // Invalid membership type
	ErrMembershipDatesInvalid    = 3105 // Invalid membership dates
	ErrMembershipAlreadyActive   = 3106 // Member already has active membership
	ErrMembershipPaymentRequired = 3107 // Payment required for membership

	// Plant slot errors (3200-3299)
	ErrPlantSlotNotFound        = 3200 // Plant slot not found
	ErrPlantSlotNotAvailable    = 3201 // Plant slot not available
	ErrPlantSlotAlreadyOccupied = 3202 // Plant slot already occupied
	ErrPlantSlotLimitReached    = 3203 // Member has reached plant slot limit
	ErrPlantSlotInMaintenance   = 3204 // Plant slot is in maintenance
	ErrPlantSlotLocationInvalid = 3205 // Invalid plant slot location
	ErrPlantSlotPositionInvalid = 3206 // Invalid plant slot position

	// Plant errors (3300-3399)
	ErrPlantNotFound           = 3300 // Plant not found
	ErrPlantAlreadyHarvested   = 3301 // Plant already harvested
	ErrPlantTypeInvalid        = 3302 // Invalid plant type
	ErrPlantStatusInvalid      = 3303 // Invalid plant status
	ErrPlantHealthCritical     = 3304 // Plant health is critical
	ErrPlantAlreadyExists      = 3305 // Plant already exists
	ErrPlantNotReadyForHarvest = 3306 // Plant not ready for harvest
	ErrPlantNameInvalid        = 3307 // Invalid plant name

	// Plant type errors (3400-3499)
	ErrPlantTypeNotFound        = 3400 // Plant type not found
	ErrPlantTypeAlreadyExists   = 3401 // Plant type already exists
	ErrPlantTypeInvalidCategory = 3402 // Invalid plant type category
	ErrPlantTypeInvalidContent  = 3403 // Invalid THC/CBD content
	ErrPlantTypeNotAvailable    = 3404 // Plant type not currently available
	ErrPlantTypeInvalidStrain   = 3405 // Invalid strain name

	// Care record errors (3500-3599)
	ErrCareRecordNotFound       = 3500 // Care record not found
	ErrCareRecordInvalid        = 3501 // Invalid care record
	ErrCareRecordTypeInvalid    = 3502 // Invalid care record type
	ErrCareRecordMeasurement    = 3503 // Invalid measurement values
	ErrCareRecordDateInvalid    = 3504 // Invalid care date
	ErrCareRecordMemberMismatch = 3505 // Care record member doesn't match plant owner

	// Harvest errors (3600-3699)
	ErrHarvestNotFound       = 3600 // Harvest not found
	ErrHarvestAlreadyExists  = 3601 // Harvest already exists
	ErrHarvestInvalidWeight  = 3602 // Invalid harvest weight
	ErrHarvestQualityInvalid = 3603 // Invalid harvest quality
	ErrHarvestStatusInvalid  = 3604 // Invalid harvest status
	ErrHarvestCollected      = 3605 // Harvest already collected
	ErrHarvestNotReady       = 3606 // Harvest not ready for collection
	ErrHarvestMemberMismatch = 3607 // Harvest member doesn't match plant owner

	// NFT errors (3700-3799)
	ErrNftAlreadyMinted    = 3700 // NFT already minted
	ErrNftTokenInvalid     = 3701 // Invalid NFT token
	ErrNftContractInvalid  = 3702 // Invalid NFT contract
	ErrNftMintingFailed    = 3703 // NFT minting failed
	ErrNftNotFound         = 3704 // NFT not found
	ErrNftOwnershipInvalid = 3705 // Invalid NFT ownership

	// System errors (3800-3899)
	ErrSystemMaintenance      = 3800 // System under maintenance
	ErrSystemCapacityReached  = 3801 // System capacity reached
	ErrSystemOperationFailed  = 3802 // System operation failed
	ErrSystemUnexpected       = 3803 // Unexpected system error
	ErrSystemPermissionDenied = 3804 // Permission denied
)

// Error messages for cannabis club errors
var cannabisErrorMessages = map[int]string{
	// Member errors
	ErrMemberNotFound:         "Member not found",
	ErrMemberAlreadyExists:    "Member already exists",
	ErrMemberEmailInvalid:     "Invalid email address",
	ErrMemberPhoneInvalid:     "Invalid phone number",
	ErrMemberMedicalIDExpired: "Medical ID is expired",
	ErrMemberMedicalIDInvalid: "Invalid medical ID",
	ErrMemberInactive:         "Member is inactive",
	ErrMemberSuspended:        "Member is suspended",
	ErrMemberAgeRestriction:   "Member does not meet age requirements",
	ErrMemberAddressInvalid:   "Invalid member address",

	// Membership errors
	ErrMembershipNotFound:        "Membership not found",
	ErrMembershipExpired:         "Membership is expired",
	ErrMembershipPaymentFailed:   "Membership payment failed",
	ErrMembershipLimit:           "Member has reached membership limit",
	ErrMembershipTypeInvalid:     "Invalid membership type",
	ErrMembershipDatesInvalid:    "Invalid membership dates",
	ErrMembershipAlreadyActive:   "Member already has active membership",
	ErrMembershipPaymentRequired: "Payment required for membership",

	// Plant slot errors
	ErrPlantSlotNotFound:        "Plant slot not found",
	ErrPlantSlotNotAvailable:    "Plant slot not available",
	ErrPlantSlotAlreadyOccupied: "Plant slot already occupied",
	ErrPlantSlotLimitReached:    "Member has reached plant slot limit",
	ErrPlantSlotInMaintenance:   "Plant slot is in maintenance",
	ErrPlantSlotLocationInvalid: "Invalid plant slot location",
	ErrPlantSlotPositionInvalid: "Invalid plant slot position",

	// Plant errors
	ErrPlantNotFound:           "Plant not found",
	ErrPlantAlreadyHarvested:   "Plant already harvested",
	ErrPlantTypeInvalid:        "Invalid plant type",
	ErrPlantStatusInvalid:      "Invalid plant status",
	ErrPlantHealthCritical:     "Plant health is critical",
	ErrPlantAlreadyExists:      "Plant already exists",
	ErrPlantNotReadyForHarvest: "Plant not ready for harvest",
	ErrPlantNameInvalid:        "Invalid plant name",

	// Plant type errors
	ErrPlantTypeNotFound:        "Plant type not found",
	ErrPlantTypeAlreadyExists:   "Plant type already exists",
	ErrPlantTypeInvalidCategory: "Invalid plant type category",
	ErrPlantTypeInvalidContent:  "Invalid THC/CBD content",
	ErrPlantTypeNotAvailable:    "Plant type not currently available",
	ErrPlantTypeInvalidStrain:   "Invalid strain name",

	// Care record errors
	ErrCareRecordNotFound:       "Care record not found",
	ErrCareRecordInvalid:        "Invalid care record",
	ErrCareRecordTypeInvalid:    "Invalid care record type",
	ErrCareRecordMeasurement:    "Invalid measurement values",
	ErrCareRecordDateInvalid:    "Invalid care date",
	ErrCareRecordMemberMismatch: "Care record member doesn't match plant owner",

	// Harvest errors
	ErrHarvestNotFound:       "Harvest not found",
	ErrHarvestAlreadyExists:  "Harvest already exists",
	ErrHarvestInvalidWeight:  "Invalid harvest weight",
	ErrHarvestQualityInvalid: "Invalid harvest quality",
	ErrHarvestStatusInvalid:  "Invalid harvest status",
	ErrHarvestCollected:      "Harvest already collected",
	ErrHarvestNotReady:       "Harvest not ready for collection",
	ErrHarvestMemberMismatch: "Harvest member doesn't match plant owner",

	// NFT errors
	ErrNftAlreadyMinted:    "NFT already minted",
	ErrNftTokenInvalid:     "Invalid NFT token",
	ErrNftContractInvalid:  "Invalid NFT contract",
	ErrNftMintingFailed:    "NFT minting failed",
	ErrNftNotFound:         "NFT not found",
	ErrNftOwnershipInvalid: "Invalid NFT ownership",

	// System errors
	ErrSystemMaintenance:      "System under maintenance",
	ErrSystemCapacityReached:  "System capacity reached",
	ErrSystemOperationFailed:  "System operation failed",
	ErrSystemUnexpected:       "Unexpected system error",
	ErrSystemPermissionDenied: "Permission denied",
}

func init() {
	// Register cannabis error messages
	for code, message := range cannabisErrorMessages {
		RegisterError(code, message)
	}
}
