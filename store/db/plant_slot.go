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

type PlantSlotDomain struct {
	BaseDomain   `bson:",inline"`
	SlotNumber   *int    `json:"slot_number" bson:"slot_number" validate:"required,gte=1"`      // Unique slot number
	MemberID     *string `json:"member_id" bson:"member_id" validate:"required,len=24"`         // Link to Member
	MembershipID *string `json:"membership_id" bson:"membership_id" validate:"required,len=24"` // Link to Membership
	Status       *string `json:"status" bson:"status" validate:"required"`                      // available, occupied, maintenance
	Location     *string `json:"location" bson:"location" validate:"required"`                  // greenhouse-1, greenhouse-2, etc.
	Position     *struct {
		Row    *int `json:"row" bson:"row" validate:"required,gte=0"`
		Column *int `json:"column" bson:"column" validate:"required,gte=0"`
	} `json:"position" bson:"position" validate:"required"`
	Notes          *string `json:"notes" bson:"notes" validate:"omitempty"`
	MaintenanceLog *[]struct {
		Date        *time.Time `json:"date" bson:"date" validate:"required"`
		Description *string    `json:"description" bson:"description" validate:"required"`
		PerformedBy *string    `json:"performed_by" bson:"performed_by" validate:"required,len=24"` // Staff member ID
	} `json:"maintenance_log" bson:"maintenance_log" validate:"omitempty,dive"`
	LastCleanDate *time.Time   `json:"last_clean_date" bson:"last_clean_date" validate:"omitempty"`
	TenantId      *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *PlantSlotDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type plantSlot struct {
	repo *repo
}

func newPlantSlot(ctx context.Context, collection *mongo.Collection) *plantSlot {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "membership_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "location", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "slot_number", Value: 1},
				{Key: "tenant_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "position.row", Value: 1},
				{Key: "position.column", Value: 1},
				{Key: "location", Value: 1},
				{Key: "tenant_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create plant slot indexes:", err)
	}

	return &plantSlot{repo: newrepo(collection)}
}

func (s *plantSlot) Save(ctx context.Context, domain *PlantSlotDomain, opts ...*options.UpdateOptions) (*PlantSlotDomain, error) {
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

func (s *plantSlot) Create(ctx context.Context, domain *PlantSlotDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *plantSlot) Update(ctx context.Context, id string, domain *PlantSlotDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *plantSlot) FindByID(ctx context.Context, id string) (*PlantSlotDomain, error) {
	var domain PlantSlotDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantSlot) FindByMemberID(ctx context.Context, memberID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindByMembershipID(ctx context.Context, membershipID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{"membership_id": membershipID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{
			"status":    status,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindByLocation(ctx context.Context, location string, tenantID enum.Tenant) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{
			"location":  location,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"status": status, "updated_at": time.Now()}},
	)
}

func (s *plantSlot) AddMaintenanceLog(ctx context.Context, id string, description string, staffID string) error {
	maintenance := M{
		"date":         time.Now(),
		"description":  description,
		"performed_by": staffID,
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"maintenance_log": maintenance},
			"$set": M{
				"status":     "maintenance",
				"updated_at": time.Now(),
			},
		},
	)
}

func (s *plantSlot) MarkCleaned(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$set": M{
				"last_clean_date": time.Now(),
				"status":          "available",
				"updated_at":      time.Now(),
			},
		},
	)
}

func (s *plantSlot) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plantSlot) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
