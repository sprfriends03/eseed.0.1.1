package db

import (
	"app/pkg/enum"
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
	MemberID       *string      `json:"member_id" bson:"member_id"`             // Link to Member
	MembershipType *string      `json:"membership_type" bson:"membership_type"` // standard, premium, etc.
	StartDate      *time.Time   `json:"start_date" bson:"start_date"`
	ExpirationDate *time.Time   `json:"expiration_date" bson:"expiration_date"`
	Status         *string      `json:"status" bson:"status"`                   // pending_payment, active, expired, canceled
	AllocatedSlots *int         `json:"allocated_slots" bson:"allocated_slots"` // Number of plant slots
	UsedSlots      *int         `json:"used_slots" bson:"used_slots"`           // Number of used plant slots
	PaymentID      *string      `json:"payment_id" bson:"payment_id"`           // Link to Payment
	PaymentAmount  *float64     `json:"payment_amount" bson:"payment_amount"`
	PaymentStatus  *string      `json:"payment_status" bson:"payment_status"` // pending, paid, failed
	AutoRenew      *bool        `json:"auto_renew" bson:"auto_renew"`
	TenantId       *enum.Tenant `json:"tenant_id" bson:"tenant_id"`
}

type membership struct {
	collection *mongo.Collection
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

	return &membership{collection}
}

func (s *membership) Create(ctx context.Context, domain *MembershipDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.collection.InsertOne(ctx, domain)
	return err
}

func (s *membership) Update(ctx context.Context, id string, domain *MembershipDomain) error {
	domain.BeforeSave()

	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$set": domain},
	)

	return err
}

func (s *membership) FindByID(ctx context.Context, id string) (*MembershipDomain, error) {
	var domain MembershipDomain
	err := s.collection.FindOne(ctx, bson.M{"_id": OID(id)}).Decode(&domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *membership) FindActiveByMemberID(ctx context.Context, memberID string) (*MembershipDomain, error) {
	var domain MembershipDomain
	now := time.Now()

	err := s.collection.FindOne(ctx, bson.M{
		"member_id":       memberID,
		"status":          "active",
		"expiration_date": bson.M{"$gt": now},
	}).Decode(&domain)

	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *membership) FindByMemberID(ctx context.Context, memberID string) ([]*MembershipDomain, error) {
	var domains []*MembershipDomain

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cur, err := s.collection.Find(ctx,
		bson.M{"member_id": memberID},
		opts,
	)

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *membership) FindExpiringSoon(ctx context.Context, daysThreshold int, tenantID enum.Tenant) ([]*MembershipDomain, error) {
	var domains []*MembershipDomain

	thresholdDate := time.Now().AddDate(0, 0, daysThreshold)

	cur, err := s.collection.Find(ctx, bson.M{
		"status":    "active",
		"tenant_id": tenantID,
		"expiration_date": bson.M{
			"$gt": time.Now(),
			"$lt": thresholdDate,
		},
		"auto_renew": false,
	})

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *membership) IncrementUsedSlots(ctx context.Context, id string) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$inc": bson.M{"used_slots": 1}},
	)

	return err
}

func (s *membership) DecrementUsedSlots(ctx context.Context, id string) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$inc": bson.M{"used_slots": -1}},
	)

	return err
}

func (s *membership) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}},
	)

	return err
}

func (s *membership) Delete(ctx context.Context, id string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": OID(id)})
	return err
}

func (s *membership) Count(ctx context.Context, filter bson.M) (int64, error) {
	return s.collection.CountDocuments(ctx, filter)
}
