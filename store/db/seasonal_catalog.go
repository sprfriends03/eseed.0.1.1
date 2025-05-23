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

type SeasonalCatalogDomain struct {
	BaseDomain      `bson:",inline"`
	Season          *string      `json:"season" bson:"season"`                       // e.g., "Spring 2023"
	StartDate       *time.Time   `json:"start_date" bson:"start_date"`               // When this catalog becomes active
	EndDate         *time.Time   `json:"end_date" bson:"end_date"`                   // When this catalog expires
	Active          *bool        `json:"active" bson:"active"`                       // Whether this catalog is currently active
	PlantTypeIDs    *[]string    `json:"plant_type_ids" bson:"plant_type_ids"`       // References to available PlantTypes
	FeaturedTypeIDs *[]string    `json:"featured_type_ids" bson:"featured_type_ids"` // References to featured plant types
	Description     *string      `json:"description" bson:"description"`
	BannerImage     *string      `json:"banner_image" bson:"banner_image"`       // URL to banner image
	SlotCapacity    *int         `json:"slot_capacity" bson:"slot_capacity"`     // Maximum number of slots available
	AllocatedSlots  *int         `json:"allocated_slots" bson:"allocated_slots"` // Number of currently allocated slots
	PriceTier       *string      `json:"price_tier" bson:"price_tier"`           // e.g., "standard", "premium"
	TenantId        *enum.Tenant `json:"tenant_id" bson:"tenant_id"`
}

type seasonalCatalog struct {
	*repo
}

func newSeasonalCatalog(ctx context.Context, collection *mongo.Collection) *seasonalCatalog {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "season", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "start_date", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "end_date", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "season", Value: 1},
				{Key: "tenant_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create seasonal catalog indexes:", err)
	}

	return &seasonalCatalog{newrepo(collection)}
}

func (s *seasonalCatalog) Create(ctx context.Context, domain *SeasonalCatalogDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.Save(ctx, domain.ID, domain)
	return err
}

func (s *seasonalCatalog) Update(ctx context.Context, id string, domain *SeasonalCatalogDomain) error {
	domain.BeforeSave()
	_, err := s.Save(ctx, OID(id), domain)
	return err
}

func (s *seasonalCatalog) FindByID(ctx context.Context, id string) (*SeasonalCatalogDomain, error) {
	var domain SeasonalCatalogDomain
	err := s.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *seasonalCatalog) FindActiveCatalog(ctx context.Context, tenant enum.Tenant) (*SeasonalCatalogDomain, error) {
	var domain SeasonalCatalogDomain
	now := time.Now()

	err := s.FindOne(ctx, M{
		"tenant_id": tenant,
		"active":    true,
		"start_date": M{
			"$lte": now,
		},
		"end_date": M{
			"$gte": now,
		},
	}, &domain)

	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *seasonalCatalog) FindAll(ctx context.Context, tenant enum.Tenant, offset, limit int64) ([]*SeasonalCatalogDomain, error) {
	var domains []*SeasonalCatalogDomain

	query := Query{
		Filter: M{"tenant_id": tenant},
		Page:   offset/limit + 1,
		Limit:  limit,
		Sorts:  "start_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *seasonalCatalog) FindBySeason(ctx context.Context, season string, tenant enum.Tenant) (*SeasonalCatalogDomain, error) {
	var domain SeasonalCatalogDomain

	err := s.FindOne(ctx, M{
		"season":    season,
		"tenant_id": tenant,
	}, &domain)

	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *seasonalCatalog) FindCurrentAndUpcoming(ctx context.Context, tenant enum.Tenant) ([]*SeasonalCatalogDomain, error) {
	var domains []*SeasonalCatalogDomain
	now := time.Now()

	query := Query{
		Filter: M{
			"tenant_id": tenant,
			"end_date": M{
				"$gte": now,
			},
		},
		Sorts: "start_date.asc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *seasonalCatalog) UpdateActive(ctx context.Context, id string, active bool) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"active":     active,
			"updated_at": time.Now(),
		}},
	)
}

func (s *seasonalCatalog) AddPlantType(ctx context.Context, id string, plantTypeID string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"plant_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) RemovePlantType(ctx context.Context, id string, plantTypeID string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$pull": M{"plant_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) AddFeaturedType(ctx context.Context, id string, plantTypeID string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"featured_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) RemoveFeaturedType(ctx context.Context, id string, plantTypeID string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$pull": M{"featured_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) IncrementAllocatedSlots(ctx context.Context, id string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$inc": M{"allocated_slots": 1},
			"$set": M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) DecrementAllocatedSlots(ctx context.Context, id string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$inc": M{"allocated_slots": -1},
			"$set": M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) Delete(ctx context.Context, id string) error {
	return s.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *seasonalCatalog) Count(ctx context.Context, filter M) int64 {
	return s.CountDocuments(ctx, Query{Filter: filter})
}
