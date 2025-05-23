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

type PlantDomain struct {
	BaseDomain          `bson:",inline"`
	PlantSlotID         *string      `json:"plant_slot_id" bson:"plant_slot_id"` // Link to PlantSlot
	MemberID            *string      `json:"member_id" bson:"member_id"`         // Link to Member
	PlantTypeID         *string      `json:"plant_type_id" bson:"plant_type_id"` // Link to PlantType
	PlantName           *string      `json:"plant_name" bson:"plant_name"`       // Optional custom name
	Status              *string      `json:"status" bson:"status"`               // growing, harvested, deceased
	PlantedDate         *time.Time   `json:"planted_date" bson:"planted_date"`
	ExpectedHarvestDate *time.Time   `json:"expected_harvest_date" bson:"expected_harvest_date"`
	ActualHarvestDate   *time.Time   `json:"actual_harvest_date" bson:"actual_harvest_date"`
	Images              *[]string    `json:"images" bson:"images"` // Image references
	LastCareDate        *time.Time   `json:"last_care_date" bson:"last_care_date"`
	Notes               *string      `json:"notes" bson:"notes"`
	HarvestID           *string      `json:"harvest_id" bson:"harvest_id"` // Link to Harvest (if harvested)
	Strain              *string      `json:"strain" bson:"strain"`         // Cannabis strain
	TenantId            *enum.Tenant `json:"tenant_id" bson:"tenant_id"`
}

type plant struct {
	repo *repo
}

func newPlant(ctx context.Context, collection *mongo.Collection) *plant {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "plant_slot_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "plant_type_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "harvest_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "strain", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create plant indexes:", err)
	}

	return &plant{repo: newrepo(collection)}
}

func (s *plant) Create(ctx context.Context, domain *PlantDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *plant) Update(ctx context.Context, id string, domain *PlantDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
	return err
}

func (s *plant) FindByID(ctx context.Context, id string) (*PlantDomain, error) {
	var domain PlantDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plant) FindByPlantSlotID(ctx context.Context, plantSlotID string) (*PlantDomain, error) {
	var domain PlantDomain
	err := s.repo.FindOne(ctx, M{"plant_slot_id": plantSlotID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plant) FindByMemberID(ctx context.Context, memberID string, offset, limit int64) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	opts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *plant) FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	query := Query{
		Filter: M{
			"status":    status,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plant) FindReadyToHarvest(ctx context.Context, tenantID enum.Tenant) ([]*PlantDomain, error) {
	var domains []*PlantDomain
	now := time.Now()

	query := Query{
		Filter: M{
			"status":                "growing",
			"tenant_id":             tenantID,
			"expected_harvest_date": M{"$lte": now},
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plant) UpdateCare(ctx context.Context, id string) error {
	now := time.Now()

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"last_care_date": now, "updated_at": now}},
	)
}

func (s *plant) AddImage(ctx context.Context, id string, imageURL string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"images": imageURL},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *plant) MarkAsHarvested(ctx context.Context, id string, harvestID string, harvestDate time.Time) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"status":              "harvested",
			"harvest_id":          harvestID,
			"actual_harvest_date": harvestDate,
			"updated_at":          time.Now(),
		}},
	)
}

func (s *plant) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plant) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
