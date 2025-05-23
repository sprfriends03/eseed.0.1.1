package db

import (
	"app/pkg/enum"
	"context"
	"time"

	"github.com/nhnghia272/gopkg"
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

func (s SeasonalCatalogDomain) BaseDto() *SeasonalCatalogBaseDto {
	return &SeasonalCatalogBaseDto{
		ID:             SID(s.ID),
		Season:         gopkg.Value(s.Season),
		StartDate:      gopkg.Value(s.StartDate),
		EndDate:        gopkg.Value(s.EndDate),
		Active:         gopkg.Value(s.Active),
		Description:    gopkg.Value(s.Description),
		BannerImage:    gopkg.Value(s.BannerImage),
		SlotCapacity:   gopkg.Value(s.SlotCapacity),
		AllocatedSlots: gopkg.Value(s.AllocatedSlots),
		PriceTier:      gopkg.Value(s.PriceTier),
	}
}

func (s SeasonalCatalogDomain) DetailDto() *SeasonalCatalogDetailDto {
	return &SeasonalCatalogDetailDto{
		ID:              SID(s.ID),
		Season:          gopkg.Value(s.Season),
		StartDate:       gopkg.Value(s.StartDate),
		EndDate:         gopkg.Value(s.EndDate),
		Active:          gopkg.Value(s.Active),
		PlantTypeIDs:    gopkg.Value(s.PlantTypeIDs),
		FeaturedTypeIDs: gopkg.Value(s.FeaturedTypeIDs),
		Description:     gopkg.Value(s.Description),
		BannerImage:     gopkg.Value(s.BannerImage),
		SlotCapacity:    gopkg.Value(s.SlotCapacity),
		AllocatedSlots:  gopkg.Value(s.AllocatedSlots),
		PriceTier:       gopkg.Value(s.PriceTier),
		CreatedAt:       gopkg.Value(s.CreatedAt),
		UpdatedAt:       gopkg.Value(s.UpdatedAt),
	}
}

type SeasonalCatalogBaseDto struct {
	ID             string    `json:"catalog_id"`
	Season         string    `json:"season"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Active         bool      `json:"active"`
	Description    string    `json:"description"`
	BannerImage    string    `json:"banner_image"`
	SlotCapacity   int       `json:"slot_capacity"`
	AllocatedSlots int       `json:"allocated_slots"`
	PriceTier      string    `json:"price_tier"`
}

type SeasonalCatalogDetailDto struct {
	ID              string    `json:"catalog_id"`
	Season          string    `json:"season"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Active          bool      `json:"active"`
	PlantTypeIDs    []string  `json:"plant_type_ids"`
	FeaturedTypeIDs []string  `json:"featured_type_ids"`
	Description     string    `json:"description"`
	BannerImage     string    `json:"banner_image"`
	SlotCapacity    int       `json:"slot_capacity"`
	AllocatedSlots  int       `json:"allocated_slots"`
	PriceTier       string    `json:"price_tier"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SeasonalCatalogQuery struct {
	Query
	Search            *string      `json:"search" form:"search" validate:"omitempty"`
	Season            *string      `json:"season" form:"season" validate:"omitempty"`
	ActiveOnly        *bool        `json:"active_only" form:"active_only" validate:"omitempty"`
	CurrentAndFuture  *bool        `json:"current_and_future" form:"current_and_future" validate:"omitempty"`
	PlantTypeID       *string      `json:"plant_type_id" form:"plant_type_id" validate:"omitempty,len=24"`
	HasAvailableSlots *bool        `json:"has_available_slots" form:"has_available_slots" validate:"omitempty"`
	TenantId          *enum.Tenant `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *SeasonalCatalogQuery) Build() *SeasonalCatalogQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{
			{"season": Regex(gopkg.Value(s.Search))},
			{"description": Regex(gopkg.Value(s.Search))},
		}
	}
	if s.Season != nil {
		s.Filter["season"] = s.Season
	}
	if s.ActiveOnly != nil && *s.ActiveOnly {
		now := time.Now()
		s.Filter["active"] = true
		s.Filter["start_date"] = M{"$lte": now}
		s.Filter["end_date"] = M{"$gte": now}
	}
	if s.CurrentAndFuture != nil && *s.CurrentAndFuture {
		now := time.Now()
		s.Filter["end_date"] = M{"$gte": now}
	}
	if s.PlantTypeID != nil {
		s.Filter["plant_type_ids"] = s.PlantTypeID
	}
	if s.HasAvailableSlots != nil && *s.HasAvailableSlots {
		s.Filter["$expr"] = M{"$lt": []string{"$allocated_slots", "$slot_capacity"}}
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	return s
}

type seasonalCatalog struct {
	repo *repo
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

	return &seasonalCatalog{repo: newrepo(collection)}
}

func (s *seasonalCatalog) Create(ctx context.Context, domain *SeasonalCatalogDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *seasonalCatalog) Update(ctx context.Context, id string, domain *SeasonalCatalogDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
	return err
}

func (s *seasonalCatalog) FindByID(ctx context.Context, id string) (*SeasonalCatalogDomain, error) {
	var domain SeasonalCatalogDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *seasonalCatalog) FindActiveCatalog(ctx context.Context, tenant enum.Tenant) (*SeasonalCatalogDomain, error) {
	var domain SeasonalCatalogDomain
	query := &SeasonalCatalogQuery{
		ActiveOnly: gopkg.Pointer(true),
		TenantId:   gopkg.Pointer(tenant),
	}

	err := s.repo.FindOne(ctx, query.Build().Filter, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *seasonalCatalog) FindAll(ctx context.Context, q *SeasonalCatalogQuery, opts ...*options.FindOptions) ([]*SeasonalCatalogDomain, error) {
	domains := make([]*SeasonalCatalogDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s *seasonalCatalog) FindBySeason(ctx context.Context, season string, tenant enum.Tenant) (*SeasonalCatalogDomain, error) {
	var domain SeasonalCatalogDomain
	query := &SeasonalCatalogQuery{
		Season:   gopkg.Pointer(season),
		TenantId: gopkg.Pointer(tenant),
	}

	err := s.repo.FindOne(ctx, query.Build().Filter, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *seasonalCatalog) FindCurrentAndUpcoming(ctx context.Context, tenant enum.Tenant) ([]*SeasonalCatalogDomain, error) {
	query := &SeasonalCatalogQuery{
		CurrentAndFuture: gopkg.Pointer(true),
		TenantId:         gopkg.Pointer(tenant),
		Query: Query{
			Sorts: "start_date.asc",
		},
	}

	return s.FindAll(ctx, query)
}

func (s *seasonalCatalog) UpdateActive(ctx context.Context, id string, active bool) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"active":     active,
			"updated_at": time.Now(),
		}},
	)
}

func (s *seasonalCatalog) AddPlantType(ctx context.Context, id string, plantTypeID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"plant_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) RemovePlantType(ctx context.Context, id string, plantTypeID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$pull": M{"plant_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) AddFeaturedType(ctx context.Context, id string, plantTypeID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"featured_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) RemoveFeaturedType(ctx context.Context, id string, plantTypeID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$pull": M{"featured_type_ids": plantTypeID},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) IncrementAllocatedSlots(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$inc": M{"allocated_slots": 1},
			"$set": M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) DecrementAllocatedSlots(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$inc": M{"allocated_slots": -1},
			"$set": M{"updated_at": time.Now()},
		},
	)
}

func (s *seasonalCatalog) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *seasonalCatalog) Count(ctx context.Context, q *SeasonalCatalogQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}
