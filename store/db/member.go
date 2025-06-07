package db

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MemberDomain struct {
	BaseDomain  `bson:",inline"`
	UserID      *string    `json:"user_id" bson:"user_id" validate:"required,len=24"` // Link to User
	Email       *string    `json:"email" bson:"email" validate:"required,email"`      // Email address
	Phone       *string    `json:"phone" bson:"phone" validate:"required,e164"`       // Phone number in E.164 format
	FirstName   *string    `json:"first_name" bson:"first_name" validate:"required"`
	LastName    *string    `json:"last_name" bson:"last_name" validate:"required"`
	DateOfBirth *time.Time `json:"date_of_birth" bson:"date_of_birth" validate:"required"`
	KYCStatus   *string    `json:"kyc_status" bson:"kyc_status" validate:"omitempty"`

	// KYC Documents - New field for storing uploaded documents
	KYCDocuments *struct {
		Passport *struct {
			Front      *string    `json:"front" bson:"front" validate:"omitempty"` // MinIO object path
			Back       *string    `json:"back" bson:"back" validate:"omitempty"`   // MinIO object path
			UploadedAt *time.Time `json:"uploaded_at" bson:"uploaded_at" validate:"omitempty"`
		} `json:"passport" bson:"passport" validate:"omitempty"`
		DriversLicense *struct {
			Front      *string    `json:"front" bson:"front" validate:"omitempty"` // MinIO object path
			Back       *string    `json:"back" bson:"back" validate:"omitempty"`   // MinIO object path
			UploadedAt *time.Time `json:"uploaded_at" bson:"uploaded_at" validate:"omitempty"`
		} `json:"drivers_license" bson:"drivers_license" validate:"omitempty"`
		NationalID *struct {
			Front      *string    `json:"front" bson:"front" validate:"omitempty"` // MinIO object path
			Back       *string    `json:"back" bson:"back" validate:"omitempty"`   // MinIO object path
			UploadedAt *time.Time `json:"uploaded_at" bson:"uploaded_at" validate:"omitempty"`
		} `json:"national_id" bson:"national_id" validate:"omitempty"`
		ProofOfAddress *struct {
			Document   *string    `json:"document" bson:"document" validate:"omitempty"` // MinIO object path
			UploadedAt *time.Time `json:"uploaded_at" bson:"uploaded_at" validate:"omitempty"`
		} `json:"proof_of_address" bson:"proof_of_address" validate:"omitempty"`
	} `json:"kyc_documents" bson:"kyc_documents" validate:"omitempty"`

	// KYC Verification - New field for tracking verification process
	KYCVerification *struct {
		SubmittedAt         *time.Time `json:"submitted_at" bson:"submitted_at" validate:"omitempty"`
		VerifiedAt          *time.Time `json:"verified_at" bson:"verified_at" validate:"omitempty"`
		VerifiedBy          *string    `json:"verified_by" bson:"verified_by" validate:"omitempty,len=24"` // User ID of verifier
		RejectedAt          *time.Time `json:"rejected_at" bson:"rejected_at" validate:"omitempty"`
		RejectedBy          *string    `json:"rejected_by" bson:"rejected_by" validate:"omitempty,len=24"` // User ID of rejector
		RejectionReason     *string    `json:"rejection_reason" bson:"rejection_reason" validate:"omitempty"`
		AdminNotes          *string    `json:"admin_notes" bson:"admin_notes" validate:"omitempty"`
		VerificationHistory *[]struct {
			Action   *string    `json:"action" bson:"action" validate:"required"`               // "submitted", "approved", "rejected"
			ActionBy *string    `json:"action_by" bson:"action_by" validate:"omitempty,len=24"` // User ID
			ActionAt *time.Time `json:"action_at" bson:"action_at" validate:"required"`
			Reason   *string    `json:"reason" bson:"reason" validate:"omitempty"`
			Notes    *string    `json:"notes" bson:"notes" validate:"omitempty"`
		} `json:"verification_history" bson:"verification_history" validate:"omitempty,dive"`
	} `json:"kyc_verification" bson:"kyc_verification" validate:"omitempty"`

	CurrentMembershipID *string    `json:"current_membership_id" bson:"current_membership_id" validate:"omitempty,len=24"` // Link to current active Membership
	MembershipType      *string    `json:"membership_type" bson:"membership_type" validate:"omitempty"`                    // Current membership type (denormalized)
	MemberStatus        *string    `json:"member_status" bson:"member_status" validate:"required"`                         // e.g., "active", inactive, suspended. TODO: use enum.MemberStatusActive
	JoinDate            *time.Time `json:"join_date" bson:"join_date" validate:"required"`
	Address             *struct {
		Street     *string `json:"street" bson:"street" validate:"required"`
		City       *string `json:"city" bson:"city" validate:"required"`
		State      *string `json:"state" bson:"state" validate:"required"`
		PostalCode *string `json:"postal_code" bson:"postal_code" validate:"required"`
		Country    *string `json:"country" bson:"country" validate:"required"`
	} `json:"address" bson:"address" validate:"required"`
	MedicalID      *string    `json:"medical_id" bson:"medical_id" validate:"omitempty"` // Medical cannabis card ID
	MedicalExpiry  *time.Time `json:"medical_expiry" bson:"medical_expiry" validate:"omitempty,gtfield=JoinDate"`
	Notes          *string    `json:"notes" bson:"notes" validate:"omitempty"`
	PaymentMethods *[]struct {
		Type       *string    `json:"type" bson:"type" validate:"required"`          // Credit card, PayPal, etc.
		Last4      *string    `json:"last4" bson:"last4" validate:"omitempty,len=4"` // Last 4 digits of card
		ExpiryDate *time.Time `json:"expiry_date" bson:"expiry_date" validate:"omitempty"`
		IsDefault  *bool      `json:"is_default" bson:"is_default" validate:"omitempty"`
		PaymentID  *string    `json:"payment_id" bson:"payment_id" validate:"required"` // Payment provider token
	} `json:"payment_methods" bson:"payment_methods" validate:"omitempty,dive"`
	Preferences *struct {
		NotifyHarvest   *bool   `json:"notify_harvest" bson:"notify_harvest" validate:"omitempty"`
		NotifyPlantCare *bool   `json:"notify_plant_care" bson:"notify_plant_care" validate:"omitempty"`
		Language        *string `json:"language" bson:"language" validate:"omitempty"`
	} `json:"preferences" bson:"preferences" validate:"omitempty"`
	TenantId *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

// MembershipDomainForVerification is an interface representing the expected structure for the getMembershipFunc callback.
// This avoids a direct dependency on a concrete MembershipDomain type from another package.
type MembershipDomainForVerification interface {
	GetStatus() *string
	GetExpirationDate() *time.Time
	// Add other methods if needed for verification
}

func (s *MemberDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type member struct {
	repo *repo
}

func newMember(ctx context.Context, collection *mongo.Collection) *member {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "phone", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "member_status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "membership_type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "join_date", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "medical_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "medical_expiry", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "kyc_status", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "current_membership_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create member indexes:", err)
	}

	return &member{repo: newrepo(collection)}
}

func (s *member) Save(ctx context.Context, domain *MemberDomain, opts ...*options.UpdateOptions) (*MemberDomain, error) {
	if err := domain.Validate(); err != nil {
		return nil, err
	}

	id, err := s.repo.Save(ctx, domain.ID, domain, opts...)
	if err != nil {
		return nil, err
	}
	domain.ID = id

	return s.FindByID(ctx, SID(id))
}

func (s *member) Create(ctx context.Context, domain *MemberDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *member) Update(ctx context.Context, id string, domain *MemberDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *member) FindByID(ctx context.Context, id string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindByUserID(ctx context.Context, userID string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"user_id": userID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindByEmail(ctx context.Context, email string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"email": email}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindByMedicalID(ctx context.Context, medicalID string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"medical_id": medicalID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*MemberDomain, error) {
	var domains []*MemberDomain

	query := Query{
		Filter: M{
			"member_status": status,
			"tenant_id":     tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *member) FindByMembershipType(ctx context.Context, membershipType string, tenantID enum.Tenant) ([]*MemberDomain, error) {
	var domains []*MemberDomain

	query := Query{
		Filter: M{
			"membership_type": membershipType,
			"tenant_id":       tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *member) FindWithExpiringMedical(ctx context.Context, daysThreshold int, tenantID enum.Tenant) ([]*MemberDomain, error) {
	var domains []*MemberDomain

	thresholdDate := time.Now().AddDate(0, 0, daysThreshold)

	query := Query{
		Filter: M{
			"tenant_id": tenantID,
			"medical_expiry": M{
				"$lte": thresholdDate,
			},
		},
	}
	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *member) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"member_status": status, "updated_at": time.Now()}},
	)
}

func (s *member) UpdateMembershipType(ctx context.Context, id string, membershipType string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"membership_type": membershipType, "updated_at": time.Now()}},
	)
}

func (s *member) AddPaymentMethod(ctx context.Context, id string, paymentMethod M) error {
	paymentMethod["_id"] = primitive.NewObjectID().Hex()
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"payment_methods": paymentMethod},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *member) RemovePaymentMethod(ctx context.Context, id string, paymentID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$pull": M{"payment_methods": M{"payment_id": paymentID}},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *member) SetDefaultPaymentMethod(ctx context.Context, id string, paymentID string) error {
	err := s.repo.UpdateMany(ctx,
		M{"_id": OID(id), "payment_methods.is_default": true},
		M{"$set": M{"payment_methods.$.is_default": false, "updated_at": time.Now()}},
	)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id), "payment_methods.payment_id": paymentID},
		M{
			"$set": M{
				"payment_methods.$.is_default": true,
				"updated_at":                   time.Now(),
			},
		},
	)
}

func (s *member) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *member) Count(ctx context.Context, filter M) int64 {
	return s.repo.CountDocuments(ctx, Query{Filter: filter})
}

// VerifyMemberActive checks if a member is active, KYC verified, 18+, and has a valid membership.
// It uses a callback function to fetch membership details to avoid direct dependency.
// The getMembershipFunc should return an object satisfying the MembershipDomainForVerification interface.
func VerifyMemberActive(ctx context.Context, member *MemberDomain, getMembershipFunc func(ctx context.Context, membershipID string) (MembershipDomainForVerification, error)) error {
	if member == nil {
		return ecode.New(http.StatusBadRequest, "member_nil_for_verification")
	}

	// Check MemberStatus
	if member.MemberStatus == nil || *member.MemberStatus != "active" {
		if member.MemberStatus != nil {
			switch *member.MemberStatus {
			case "pending_verification":
				return ecode.New(http.StatusForbidden, "member_pending_verification")
			case "pending_approval":
				return ecode.New(http.StatusForbidden, "member_pending_approval")
			case "inactive":
				return ecode.New(http.StatusForbidden, "member_inactive")
			case "suspended":
				return ecode.New(http.StatusForbidden, "member_suspended")
			case "terminated":
				return ecode.New(http.StatusForbidden, "member_terminated")
			default:
				return ecode.New(http.StatusForbidden, "member_status_unknown").Desc(fmt.Errorf("Member status: %s", *member.MemberStatus))
			}
		}
		return ecode.New(http.StatusForbidden, "member_status_not_active_or_missing")
	}

	// Check KYCStatus
	if member.KYCStatus == nil || *member.KYCStatus != "verified" {
		if member.KYCStatus != nil {
			switch *member.KYCStatus {
			case "pending_kyc", "submitted", "in_review", "not_started":
				return ecode.New(http.StatusForbidden, "kyc_pending_or_not_started")
			case "rejected", "requires_resubmission":
				return ecode.New(http.StatusForbidden, "kyc_failed_or_needs_resubmit")
			default:
				return ecode.New(http.StatusForbidden, "kyc_status_unknown").Desc(fmt.Errorf("KYC status: %s", *member.KYCStatus))
			}
		}
		return ecode.New(http.StatusForbidden, "kyc_not_verified_or_missing")
	}

	// Check Age (18+)
	if member.DateOfBirth == nil {
		return ecode.New(http.StatusForbidden, "member_dob_missing")
	}
	if time.Now().Sub(*member.DateOfBirth).Hours()/(24*365.25) < 18 {
		return ecode.New(http.StatusForbidden, "member_underage")
	}

	// Check CurrentMembershipID and its validity via callback
	if member.CurrentMembershipID == nil || *member.CurrentMembershipID == "" {
		return ecode.New(http.StatusForbidden, "member_missing_current_membership_id")
	}

	membership, err := getMembershipFunc(ctx, *member.CurrentMembershipID)
	if err != nil {
		if _, ok := err.(*ecode.Error); ok {
			return err
		}
		return ecode.InternalServerError.Desc(err)
	}
	if membership == nil {
		return ecode.New(http.StatusNotFound, "membership_details_not_found_via_callback")
	}

	membershipStatusStr := membership.GetStatus()
	if membershipStatusStr == nil || *membershipStatusStr != "active" {
		return ecode.New(http.StatusForbidden, "membership_not_active_from_callback")
	}

	membershipExpiry := membership.GetExpirationDate()
	if membershipExpiry == nil || time.Now().After(*membershipExpiry) {
		return ecode.New(http.StatusForbidden, "membership_expired_from_callback")
	}

	return nil
}

// UpdateKYCDocuments updates the KYC documents for a member
func (s *member) UpdateKYCDocuments(ctx context.Context, memberID string, documents map[string]string) error {
	now := time.Now()

	// Build the update based on document type and file type
	update := M{"$set": M{"updated_at": now}}

	for key, objectPath := range documents {
		// Parse key format: "passport_front", "drivers_license_back", etc.
		if objectPath != "" {
			update["$set"].(M)[fmt.Sprintf("kyc_documents.%s", key)] = objectPath
			update["$set"].(M)[fmt.Sprintf("kyc_documents.%s.uploaded_at", getDocumentTypeFromKey(key))] = now
		}
	}

	return s.repo.UpdateOne(ctx, M{"_id": OID(memberID)}, update)
}

// UpdateKYCStatus updates the KYC status and verification information for a member
func (s *member) UpdateKYCStatus(ctx context.Context, memberID, status, verifiedBy string, verification interface{}) error {
	now := time.Now()

	update := M{
		"$set": M{
			"kyc_status": status,
			"updated_at": now,
		},
	}

	// Update verification fields based on status
	switch status {
	case "submitted":
		update["$set"].(M)["kyc_verification.submitted_at"] = now
		// Add to verification history
		historyEntry := M{
			"action":    "submitted",
			"action_by": verifiedBy,
			"action_at": now,
		}
		update["$push"] = M{"kyc_verification.verification_history": historyEntry}

	case "verified":
		update["$set"].(M)["kyc_verification.verified_at"] = now
		update["$set"].(M)["kyc_verification.verified_by"] = verifiedBy
		// Add admin notes if provided
		if verification != nil {
			if verifyData, ok := verification.(map[string]interface{}); ok {
				if notes, exists := verifyData["notes"]; exists {
					update["$set"].(M)["kyc_verification.admin_notes"] = notes
				}
			}
		}
		// Add to verification history
		historyEntry := M{
			"action":    "approved",
			"action_by": verifiedBy,
			"action_at": now,
		}
		if verification != nil {
			if verifyData, ok := verification.(map[string]interface{}); ok {
				if notes, exists := verifyData["notes"]; exists {
					historyEntry["notes"] = notes
				}
			}
		}
		update["$push"] = M{"kyc_verification.verification_history": historyEntry}

	case "rejected":
		update["$set"].(M)["kyc_verification.rejected_at"] = now
		update["$set"].(M)["kyc_verification.rejected_by"] = verifiedBy
		// Handle rejection reason and notes
		if verification != nil {
			if verifyData, ok := verification.(map[string]interface{}); ok {
				if reason, exists := verifyData["reason"]; exists {
					update["$set"].(M)["kyc_verification.rejection_reason"] = reason
				}
				if notes, exists := verifyData["notes"]; exists {
					update["$set"].(M)["kyc_verification.admin_notes"] = notes
				}
			}
		}
		// Add to verification history
		historyEntry := M{
			"action":    "rejected",
			"action_by": verifiedBy,
			"action_at": now,
		}
		if verification != nil {
			if verifyData, ok := verification.(map[string]interface{}); ok {
				if reason, exists := verifyData["reason"]; exists {
					historyEntry["reason"] = reason
				}
				if notes, exists := verifyData["notes"]; exists {
					historyEntry["notes"] = notes
				}
			}
		}
		update["$push"] = M{"kyc_verification.verification_history": historyEntry}
	}

	return s.repo.UpdateOne(ctx, M{"_id": OID(memberID)}, update)
}

// GetPendingKYCVerifications returns members with pending KYC verifications
func (s *member) GetPendingKYCVerifications(ctx context.Context, tenantID enum.Tenant, page, limit int64) ([]*MemberDomain, int64, error) {
	var domains []*MemberDomain

	query := Query{
		Filter: M{
			"kyc_status": M{"$in": []string{"submitted", "in_review"}},
			"tenant_id":  tenantID,
		},
		Page:  page,
		Limit: limit,
		Sorts: "kyc_verification.submitted_at.asc", // Oldest first
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount := s.repo.CountDocuments(ctx, Query{
		Filter: M{
			"kyc_status": M{"$in": []string{"submitted", "in_review"}},
			"tenant_id":  tenantID,
		},
	})

	return domains, totalCount, nil
}

// CountKYCByStatus returns count of members by KYC status for a tenant
func (s *member) CountKYCByStatus(ctx context.Context, tenantID enum.Tenant, status string) int64 {
	return s.repo.CountDocuments(ctx, Query{
		Filter: M{
			"kyc_status": status,
			"tenant_id":  tenantID,
		},
	})
}

// Helper function to extract document type from key
func getDocumentTypeFromKey(key string) string {
	// Extract base document type from keys like "passport.front", "drivers_license.back"
	if len(key) > 0 {
		parts := strings.Split(key, ".")
		if len(parts) >= 1 {
			return parts[0]
		}
	}
	return key
}

// GetKYCStatistics returns comprehensive KYC statistics for a tenant
func (s *member) GetKYCStatistics(ctx context.Context, tenantID enum.Tenant) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by each status
	statuses := []string{"not_started", "pending_kyc", "submitted", "in_review", "verified", "rejected"}

	for _, status := range statuses {
		count := s.CountKYCByStatus(ctx, tenantID, status)
		stats[status] = count
	}

	// Total members
	stats["total"] = s.repo.CountDocuments(ctx, Query{
		Filter: M{"tenant_id": tenantID},
	})

	// Members with any KYC documents uploaded
	stats["with_documents"] = s.repo.CountDocuments(ctx, Query{
		Filter: M{
			"tenant_id":     tenantID,
			"kyc_documents": M{"$exists": true, "$ne": nil},
		},
	})

	return stats, nil
}

// FindMembersByKYCStatus returns members filtered by KYC status with pagination
func (s *member) FindMembersByKYCStatus(ctx context.Context, tenantID enum.Tenant, status string, page, limit int64) ([]*MemberDomain, int64, error) {
	var domains []*MemberDomain

	query := Query{
		Filter: M{
			"kyc_status": status,
			"tenant_id":  tenantID,
		},
		Page:  page,
		Limit: limit,
		Sorts: "created_at.desc", // Newest first
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount := s.CountKYCByStatus(ctx, tenantID, status)

	return domains, totalCount, nil
}

// DeleteKYCDocuments removes specific KYC documents for a member
func (s *member) DeleteKYCDocuments(ctx context.Context, memberID string, documentType string) error {
	now := time.Now()

	update := M{
		"$unset": M{},
		"$set":   M{"updated_at": now},
	}

	// Remove specific document type
	switch documentType {
	case "passport":
		update["$unset"].(M)["kyc_documents.passport"] = ""
	case "drivers_license":
		update["$unset"].(M)["kyc_documents.drivers_license"] = ""
	case "national_id":
		update["$unset"].(M)["kyc_documents.national_id"] = ""
	case "proof_of_address":
		update["$unset"].(M)["kyc_documents.proof_of_address"] = ""
	case "all":
		update["$unset"].(M)["kyc_documents"] = ""
	default:
		return fmt.Errorf("invalid document type: %s", documentType)
	}

	return s.repo.UpdateOne(ctx, M{"_id": OID(memberID)}, update)
}

// GetKYCVerificationHistory returns the verification history for a member
func (s *member) GetKYCVerificationHistory(ctx context.Context, memberID string) ([]map[string]interface{}, error) {
	member, err := s.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	var history []map[string]interface{}

	if member.KYCVerification != nil && member.KYCVerification.VerificationHistory != nil {
		for _, entry := range *member.KYCVerification.VerificationHistory {
			historyItem := map[string]interface{}{
				"action":    getStringValue(entry.Action),
				"action_by": getStringValue(entry.ActionBy),
				"action_at": entry.ActionAt,
				"reason":    getStringValue(entry.Reason),
				"notes":     getStringValue(entry.Notes),
			}
			history = append(history, historyItem)
		}
	}

	return history, nil
}

// UpdateKYCDocumentStatus updates the status of a specific document during verification
func (s *member) UpdateKYCDocumentStatus(ctx context.Context, memberID string, documentType string, status string, notes *string) error {
	now := time.Now()

	update := M{
		"$set": M{
			fmt.Sprintf("kyc_documents.%s.verification_status", documentType): status,
			fmt.Sprintf("kyc_documents.%s.verified_at", documentType):         now,
			"updated_at": now,
		},
	}

	if notes != nil {
		update["$set"].(M)[fmt.Sprintf("kyc_documents.%s.verification_notes", documentType)] = *notes
	}

	return s.repo.UpdateOne(ctx, M{"_id": OID(memberID)}, update)
}

// GetMembersRequiringKYCReview returns members that need KYC review based on various criteria
func (s *member) GetMembersRequiringKYCReview(ctx context.Context, tenantID enum.Tenant, criteria map[string]interface{}) ([]*MemberDomain, error) {
	var domains []*MemberDomain

	filter := M{"tenant_id": tenantID}

	// Add criteria to filter
	if submittedBefore, exists := criteria["submitted_before"]; exists {
		filter["kyc_verification.submitted_at"] = M{"$lt": submittedBefore}
	}

	if statuses, exists := criteria["statuses"]; exists {
		filter["kyc_status"] = M{"$in": statuses}
	}

	if hasDocuments, exists := criteria["has_documents"]; exists && hasDocuments.(bool) {
		filter["kyc_documents"] = M{"$exists": true, "$ne": nil}
	}

	query := Query{
		Filter: filter,
		Sorts:  "kyc_verification.submitted_at.asc", // Oldest first
	}

	err := s.repo.FindAll(ctx, query, &domains)
	return domains, err
}

// Helper function to safely get string value from pointer
