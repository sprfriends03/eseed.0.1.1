package db

import (
	"time"
)

// KYCDocumentUploadData represents the data for uploading KYC documents
type KYCDocumentUploadData struct {
	DocumentType string `json:"document_type" validate:"required,oneof=passport drivers_license national_id proof_of_address"`
	FileType     string `json:"file_type" validate:"required,oneof=front back document"`
}

// KYCSubmissionData represents the data for submitting KYC for verification
type KYCSubmissionData struct {
	DocumentType    string `json:"document_type" validate:"required,oneof=passport drivers_license national_id"`
	HasAllDocuments bool   `json:"has_all_documents" validate:"required"`
	ConfirmAccuracy bool   `json:"confirm_accuracy" validate:"required"`
}

// KYCVerificationData represents the data for admin verification of KYC
type KYCVerificationData struct {
	Action string  `json:"action" validate:"required,oneof=approve reject"`
	Reason *string `json:"reason" validate:"omitempty"`
	Notes  *string `json:"notes" validate:"omitempty"`
}

// KYCStatusDto represents the KYC status response for members
type KYCStatusDto struct {
	KYCStatus       string                       `json:"kyc_status"`
	CanSubmit       bool                         `json:"can_submit"`
	HasDocuments    bool                         `json:"has_documents"`
	DocumentsStatus map[string]KYCDocumentStatus `json:"documents_status"`
	Verification    *KYCVerificationStatusDto    `json:"verification,omitempty"`
	History         []KYCVerificationHistoryDto  `json:"history,omitempty"`
}

// KYCDocumentStatus represents the status of individual document types
type KYCDocumentStatus struct {
	HasFront    bool       `json:"has_front"`
	HasBack     bool       `json:"has_back"`
	HasDocument bool       `json:"has_document"` // For proof_of_address
	UploadedAt  *time.Time `json:"uploaded_at,omitempty"`
	IsComplete  bool       `json:"is_complete"`
}

// KYCVerificationStatusDto represents the verification status details
type KYCVerificationStatusDto struct {
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	VerifiedBy      *string    `json:"verified_by,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RejectedBy      *string    `json:"rejected_by,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	AdminNotes      *string    `json:"admin_notes,omitempty"`
}

// KYCVerificationHistoryDto represents a single verification history entry
type KYCVerificationHistoryDto struct {
	Action   string     `json:"action"`
	ActionBy *string    `json:"action_by,omitempty"`
	ActionAt *time.Time `json:"action_at"`
	Reason   *string    `json:"reason,omitempty"`
	Notes    *string    `json:"notes,omitempty"`
}

// KYCPendingMemberDto represents a member in the pending KYC list for admins
type KYCPendingMemberDto struct {
	ID           string                       `json:"id"`
	Email        string                       `json:"email"`
	FirstName    string                       `json:"first_name"`
	LastName     string                       `json:"last_name"`
	KYCStatus    string                       `json:"kyc_status"`
	SubmittedAt  *time.Time                   `json:"submitted_at"`
	DocumentType string                       `json:"document_type"`
	Documents    map[string]KYCDocumentStatus `json:"documents"`
}

// Helper functions for converting between domain objects and DTOs

// BuildKYCStatusDto converts a MemberDomain to a KYCStatusDto
func BuildKYCStatusDto(member *MemberDomain) KYCStatusDto {
	status := KYCStatusDto{
		KYCStatus:       getKYCStatusOrDefault(member.KYCStatus),
		DocumentsStatus: make(map[string]KYCDocumentStatus),
	}

	// Build document status
	if member.KYCDocuments != nil {
		if member.KYCDocuments.Passport != nil {
			status.DocumentsStatus["passport"] = KYCDocumentStatus{
				HasFront:   member.KYCDocuments.Passport.Front != nil && *member.KYCDocuments.Passport.Front != "",
				HasBack:    member.KYCDocuments.Passport.Back != nil && *member.KYCDocuments.Passport.Back != "",
				UploadedAt: member.KYCDocuments.Passport.UploadedAt,
				IsComplete: member.KYCDocuments.Passport.Front != nil && *member.KYCDocuments.Passport.Front != "" &&
					member.KYCDocuments.Passport.Back != nil && *member.KYCDocuments.Passport.Back != "",
			}
		}

		if member.KYCDocuments.DriversLicense != nil {
			status.DocumentsStatus["drivers_license"] = KYCDocumentStatus{
				HasFront:   member.KYCDocuments.DriversLicense.Front != nil && *member.KYCDocuments.DriversLicense.Front != "",
				HasBack:    member.KYCDocuments.DriversLicense.Back != nil && *member.KYCDocuments.DriversLicense.Back != "",
				UploadedAt: member.KYCDocuments.DriversLicense.UploadedAt,
				IsComplete: member.KYCDocuments.DriversLicense.Front != nil && *member.KYCDocuments.DriversLicense.Front != "" &&
					member.KYCDocuments.DriversLicense.Back != nil && *member.KYCDocuments.DriversLicense.Back != "",
			}
		}

		if member.KYCDocuments.NationalID != nil {
			status.DocumentsStatus["national_id"] = KYCDocumentStatus{
				HasFront:   member.KYCDocuments.NationalID.Front != nil && *member.KYCDocuments.NationalID.Front != "",
				HasBack:    member.KYCDocuments.NationalID.Back != nil && *member.KYCDocuments.NationalID.Back != "",
				UploadedAt: member.KYCDocuments.NationalID.UploadedAt,
				IsComplete: member.KYCDocuments.NationalID.Front != nil && *member.KYCDocuments.NationalID.Front != "" &&
					member.KYCDocuments.NationalID.Back != nil && *member.KYCDocuments.NationalID.Back != "",
			}
		}

		if member.KYCDocuments.ProofOfAddress != nil {
			status.DocumentsStatus["proof_of_address"] = KYCDocumentStatus{
				HasDocument: member.KYCDocuments.ProofOfAddress.Document != nil && *member.KYCDocuments.ProofOfAddress.Document != "",
				UploadedAt:  member.KYCDocuments.ProofOfAddress.UploadedAt,
				IsComplete:  member.KYCDocuments.ProofOfAddress.Document != nil && *member.KYCDocuments.ProofOfAddress.Document != "",
			}
		}
	}

	// Check if has any documents
	status.HasDocuments = false
	for _, docStatus := range status.DocumentsStatus {
		if docStatus.IsComplete {
			status.HasDocuments = true
			break
		}
	}

	// Determine if can submit
	status.CanSubmit = status.HasDocuments && (status.KYCStatus == "not_started" || status.KYCStatus == "pending_kyc" || status.KYCStatus == "rejected")

	// Build verification info
	if member.KYCVerification != nil {
		status.Verification = &KYCVerificationStatusDto{
			SubmittedAt:     member.KYCVerification.SubmittedAt,
			VerifiedAt:      member.KYCVerification.VerifiedAt,
			VerifiedBy:      member.KYCVerification.VerifiedBy,
			RejectedAt:      member.KYCVerification.RejectedAt,
			RejectedBy:      member.KYCVerification.RejectedBy,
			RejectionReason: member.KYCVerification.RejectionReason,
			AdminNotes:      member.KYCVerification.AdminNotes,
		}

		// Build history
		if member.KYCVerification.VerificationHistory != nil {
			for _, historyItem := range *member.KYCVerification.VerificationHistory {
				status.History = append(status.History, KYCVerificationHistoryDto{
					Action:   getStringValue(historyItem.Action),
					ActionBy: historyItem.ActionBy,
					ActionAt: historyItem.ActionAt,
					Reason:   historyItem.Reason,
					Notes:    historyItem.Notes,
				})
			}
		}
	}

	return status
}

// BuildKYCPendingMemberDto converts a MemberDomain to a KYCPendingMemberDto for admin views
func BuildKYCPendingMemberDto(member *MemberDomain) KYCPendingMemberDto {
	dto := KYCPendingMemberDto{
		ID:        SID(member.ID),
		Email:     getStringValue(member.Email),
		FirstName: getStringValue(member.FirstName),
		LastName:  getStringValue(member.LastName),
		KYCStatus: getKYCStatusOrDefault(member.KYCStatus),
		Documents: make(map[string]KYCDocumentStatus),
	}

	// Get submission date
	if member.KYCVerification != nil && member.KYCVerification.SubmittedAt != nil {
		dto.SubmittedAt = member.KYCVerification.SubmittedAt
	}

	// Determine primary document type
	dto.DocumentType = "unknown"
	if member.KYCDocuments != nil {
		if member.KYCDocuments.Passport != nil && member.KYCDocuments.Passport.Front != nil {
			dto.DocumentType = "passport"
		} else if member.KYCDocuments.DriversLicense != nil && member.KYCDocuments.DriversLicense.Front != nil {
			dto.DocumentType = "drivers_license"
		} else if member.KYCDocuments.NationalID != nil && member.KYCDocuments.NationalID.Front != nil {
			dto.DocumentType = "national_id"
		}
	}

	// Build document status (reuse from BuildKYCStatusDto)
	statusDto := BuildKYCStatusDto(member)
	dto.Documents = statusDto.DocumentsStatus

	return dto
}

// Helper functions

func getKYCStatusOrDefault(status *string) string {
	if status != nil {
		return *status
	}
	return "not_started"
}

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
