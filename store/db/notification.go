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

type NotificationDomain struct {
	BaseDomain   `bson:",inline"`
	MemberID     *string      `json:"member_id" bson:"member_id" validate:"required,len=24"` // Link to Member
	Title        *string      `json:"title" bson:"title" validate:"required"`
	Message      *string      `json:"message" bson:"message" validate:"required"`
	Type         *string      `json:"type" bson:"type" validate:"required"`         // plant_care, harvest_ready, membership, system
	Status       *string      `json:"status" bson:"status" validate:"required"`     // unread, read, archived
	Priority     *string      `json:"priority" bson:"priority" validate:"required"` // high, normal, low
	ReadAt       *time.Time   `json:"read_at" bson:"read_at" validate:"omitempty"`
	RelatedID    *string      `json:"related_id" bson:"related_id" validate:"omitempty"`     // ID of related entity (plant, harvest, etc.)
	RelatedType  *string      `json:"related_type" bson:"related_type" validate:"omitempty"` // Type of related entity
	ExpiresAt    *time.Time   `json:"expires_at" bson:"expires_at" validate:"omitempty,gtfield=CreatedAt"`
	DeliveryType *string      `json:"delivery_type" bson:"delivery_type" validate:"required"` // app, email, sms, push
	TenantId     *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *NotificationDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type notification struct {
	repo *repo
}

func newNotification(ctx context.Context, collection *mongo.Collection) *notification {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "priority", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "related_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
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
		logrus.Errorln("Failed to create notification indexes:", err)
	}

	return &notification{repo: newrepo(collection)}
}

func (s *notification) Save(ctx context.Context, domain *NotificationDomain, opts ...*options.UpdateOptions) (*NotificationDomain, error) {
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

func (s *notification) Create(ctx context.Context, domain *NotificationDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *notification) Update(ctx context.Context, id string, domain *NotificationDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *notification) FindByID(ctx context.Context, id string) (*NotificationDomain, error) {
	var domain NotificationDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *notification) FindByMemberID(ctx context.Context, memberID string, limit int64) ([]*NotificationDomain, error) {
	var domains []*NotificationDomain

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *notification) FindUnreadByMemberID(ctx context.Context, memberID string) ([]*NotificationDomain, error) {
	var domains []*NotificationDomain

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	query := Query{
		Filter: M{
			"member_id": memberID,
			"status":    "unread",
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *notification) FindByType(ctx context.Context, notificationType string, tenantID enum.Tenant) ([]*NotificationDomain, error) {
	var domains []*NotificationDomain

	query := Query{
		Filter: M{
			"type":      notificationType,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *notification) FindByRelatedEntity(ctx context.Context, relatedID string, relatedType string) ([]*NotificationDomain, error) {
	var domains []*NotificationDomain

	query := Query{
		Filter: M{
			"related_id":   relatedID,
			"related_type": relatedType,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *notification) MarkAsRead(ctx context.Context, id string) error {
	now := time.Now()
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$set": M{
				"status":     "read",
				"read_at":    now,
				"updated_at": now,
			},
		},
	)
}

func (s *notification) Archive(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"status": "archived", "updated_at": time.Now()}},
	)
}

func (s *notification) DeleteExpired(ctx context.Context, tenantID enum.Tenant) error {
	now := time.Now()
	return s.repo.DeleteMany(ctx, M{
		"expires_at": M{"$lt": now},
		"tenant_id":  tenantID,
	})
}

func (s *notification) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *notification) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
