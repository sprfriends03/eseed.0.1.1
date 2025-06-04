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

type PlantDomain struct {
	BaseDomain      `bson:",inline"`
	PlantTypeID     *string      `json:"plant_type_id" bson:"plant_type_id" validate:"required,len=24"` // Link to PlantType
	PlantSlotID     *string      `json:"plant_slot_id" bson:"plant_slot_id" validate:"required,len=24"` // Link to PlantSlot
	MemberID        *string      `json:"member_id" bson:"member_id" validate:"required,len=24"`         // Link to Member (owner)
	Status          *string      `json:"status" bson:"status" validate:"required"`                      // seedling, vegetative, flowering, harvested, dead
	PlantedDate     *time.Time   `json:"planted_date" bson:"planted_date" validate:"required"`
	ExpectedHarvest *time.Time   `json:"expected_harvest" bson:"expected_harvest" validate:"required,gtfield=PlantedDate"`
	ActualHarvest   *time.Time   `json:"actual_harvest" bson:"actual_harvest" validate:"omitempty,gtfield=PlantedDate"`
	Name            *string      `json:"name" bson:"name" validate:"required"`                  // Plant nickname
	Health          *int         `json:"health" bson:"health" validate:"required,gte=1,lte=10"` // Health rating (1-10)
	Height          *float64     `json:"height" bson:"height" validate:"omitempty,gte=0"`       // in cm
	Images          *[]string    `json:"images" bson:"images" validate:"omitempty,dive,required"`
	Notes           *string      `json:"notes" bson:"notes" validate:"omitempty"`
	Strain          *string      `json:"strain" bson:"strain" validate:"required"` // Denormalized from PlantType
	HarvestID       *string      `json:"harvest_id" bson:"harvest_id" validate:"omitempty,len=24"`
	NFTTokenID      *string      `json:"nft_token_id" bson:"nft_token_id" validate:"omitempty"`
	TenantId        *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *PlantDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type plant struct {
	repo *repo
}

func newPlant(ctx context.Context, collection *mongo.Collection) *plant {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "plant_type_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "plant_slot_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "strain", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "planted_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "expected_harvest", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "harvest_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "nft_token_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create plant indexes:", err)
	}

	return &plant{repo: newrepo(collection)}
}

func (s *plant) Save(ctx context.Context, domain *PlantDomain, opts ...*options.UpdateOptions) (*PlantDomain, error) {
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

func (s *plant) Create(ctx context.Context, domain *PlantDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *plant) Update(ctx context.Context, id string, domain *PlantDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
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

func (s *plant) FindByMemberID(ctx context.Context, memberID string) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	opts := options.Find().SetSort(bson.D{{Key: "planted_date", Value: -1}})
	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *plant) FindByPlantSlotID(ctx context.Context, plantSlotID string) (*PlantDomain, error) {
	var domain PlantDomain

	// Find active plant in this slot
	query := M{
		"plant_slot_id": plantSlotID,
		"status": M{
			"$nin": []string{"harvested", "dead"},
		},
	}

	err := s.repo.FindOne(ctx, query, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plant) FindActiveByMemberID(ctx context.Context, memberID string) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	opts := options.Find().SetSort(bson.D{{Key: "planted_date", Value: -1}})
	query := Query{
		Filter: M{
			"member_id": memberID,
			"status": M{
				"$nin": []string{"harvested", "dead"},
			},
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *plant) FindReadyForHarvest(ctx context.Context, tenantID enum.Tenant) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	now := time.Now()
	query := Query{
		Filter: M{
			"tenant_id": tenantID,
			"status":    "flowering",
			"expected_harvest": M{
				"$lte": now,
			},
			"harvest_id": M{
				"$exists": false,
			},
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plant) UpdateStatus(ctx context.Context, id string, status string) error {
	updates := M{"status": status, "updated_at": time.Now()}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}

func (s *plant) UpdateHealth(ctx context.Context, id string, health int) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"health": health, "updated_at": time.Now()}},
	)
}

func (s *plant) UpdateHeight(ctx context.Context, id string, height float64) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"height": height, "updated_at": time.Now()}},
	)
}

func (s *plant) SetHarvestID(ctx context.Context, id string, harvestID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$set": M{
				"harvest_id":     harvestID,
				"status":         "harvested",
				"actual_harvest": time.Now(),
				"updated_at":     time.Now(),
			},
		},
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

func (s *plant) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plant) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
