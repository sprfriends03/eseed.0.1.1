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

type MembershipDomain struct {
	BaseDomain     `bson:",inline"`
	MemberID       *string      `json:"member_id" bson:"member_id" validate:"required,len=24"`      // Link to Member
	MembershipType *string      `json:"membership_type" bson:"membership_type" validate:"required"` // standard, premium, etc.
	StartDate      *time.Time   `json:"start_date" bson:"start_date" validate:"required"`
	ExpirationDate *time.Time   `json:"expiration_date" bson:"expiration_date" validate:"required,gtfield=StartDate"`
	Status         *string      `json:"status" bson:"status" validate:"required"`                         // pending_payment, active, expired, canceled
	AllocatedSlots *int         `json:"allocated_slots" bson:"allocated_slots" validate:"required,gte=0"` // Number of plant slots
	UsedSlots      *int         `json:"used_slots" bson:"used_slots" validate:"required,gte=0"`           // Number of used plant slots
	PaymentID      *string      `json:"payment_id" bson:"payment_id" validate:"omitempty,len=24"`         // Link to Payment
	PaymentAmount  *float64     `json:"payment_amount" bson:"payment_amount" validate:"required,gte=0"`
	PaymentStatus  *string      `json:"payment_status" bson:"payment_status" validate:"required"` // pending, paid, failed
	AutoRenew      *bool        `json:"auto_renew" bson:"auto_renew" validate:"omitempty"`
	TenantId       *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *MembershipDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type membership struct {
	repo *repo
}

func newMembership(ctx context.Context, collection *mongo.Collection) *membership {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "payment_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "expiration_date", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "member_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create membership indexes:", err)
	}

	return &membership{repo: newrepo(collection)}
}

func (s *membership) Save(ctx context.Context, domain *MembershipDomain, opts ...*options.UpdateOptions) (*MembershipDomain, error) {
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

func (s *membership) Create(ctx context.Context, domain *MembershipDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *membership) Update(ctx context.Context, id string, domain *MembershipDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *membership) FindByID(ctx context.Context, id string) (*MembershipDomain, error) {
	var domain MembershipDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *membership) FindActiveByMemberID(ctx context.Context, memberID string) (*MembershipDomain, error) {
	var domain MembershipDomain
	now := time.Now()

	filter := M{
		"member_id":       memberID,
		"status":          "active",
		"expiration_date": M{"$gt": now},
	}

	err := s.repo.FindOne(ctx, filter, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *membership) FindByMemberID(ctx context.Context, memberID string) ([]*MembershipDomain, error) {
	var domains []*MembershipDomain

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *membership) FindExpiringSoon(ctx context.Context, daysThreshold int, tenantID enum.Tenant) ([]*MembershipDomain, error) {
	var domains []*MembershipDomain

	thresholdDate := time.Now().AddDate(0, 0, daysThreshold)

	query := Query{
		Filter: M{
			"status":    "active",
			"tenant_id": tenantID,
			"expiration_date": M{
				"$gt": time.Now(),
				"$lt": thresholdDate,
			},
			"auto_renew": false,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *membership) IncrementUsedSlots(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$inc": M{"used_slots": 1}},
	)
}

func (s *membership) DecrementUsedSlots(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$inc": M{"used_slots": -1}},
	)
}

func (s *membership) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"status": status, "updated_at": time.Now()}},
	)
}

func (s *membership) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *membership) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
