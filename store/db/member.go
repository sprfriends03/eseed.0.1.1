package db

import (
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MemberDomain struct {
	BaseDomain     `bson:",inline"`
	UserID         *string    `json:"user_id" bson:"user_id" validate:"required,len=24"` // Link to User
	Email          *string    `json:"email" bson:"email" validate:"required,email"`      // Email address
	Phone          *string    `json:"phone" bson:"phone" validate:"required,e164"`       // Phone number in E.164 format
	FirstName      *string    `json:"first_name" bson:"first_name" validate:"required"`
	LastName       *string    `json:"last_name" bson:"last_name" validate:"required"`
	DateOfBirth    *time.Time `json:"date_of_birth" bson:"date_of_birth" validate:"required"`
	MembershipType *string    `json:"membership_type" bson:"membership_type" validate:"omitempty"` // Current membership type (denormalized)
	MemberStatus   *string    `json:"member_status" bson:"member_status" validate:"required"`      // active, inactive, suspended, etc.
	JoinDate       *time.Time `json:"join_date" bson:"join_date" validate:"required"`
	Address        *struct {
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

	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
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
				"$gt": time.Now(),
				"$lt": thresholdDate,
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
	// First, unset any existing default
	err := s.repo.UpdateOne(ctx,
		M{
			"_id":                        OID(id),
			"payment_methods.is_default": true,
		},
		M{"$set": M{"payment_methods.$.is_default": false}},
	)
	if err != nil {
		return err
	}

	// Then set the new default
	return s.repo.UpdateOne(ctx,
		M{
			"_id":                        OID(id),
			"payment_methods.payment_id": paymentID,
		},
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

func (s *member) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
