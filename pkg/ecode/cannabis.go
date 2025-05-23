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
