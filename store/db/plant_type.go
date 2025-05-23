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

type PlantTypeDomain struct {
	BaseDomain        `bson:",inline"`
	Name              *string            `json:"name" bson:"name"`
	Strain            *string            `json:"strain" bson:"strain"`
	Category          *string            `json:"category" bson:"category"` // indica, sativa, hybrid
	ThcContent        *float64           `json:"thc_content" bson:"thc_content"`
	CbdContent        *float64           `json:"cbd_content" bson:"cbd_content"`
	GrowthDifficulty  *int               `json:"growth_difficulty" bson:"growth_difficulty"` // 1-10 scale
	AverageYield      *float64           `json:"average_yield" bson:"average_yield"`         // in grams
	FloweringTime     *int               `json:"flowering_time" bson:"flowering_time"`       // in days
	GrowthPhases      *[]string          `json:"growth_phases" bson:"growth_phases"`
	Description       *string            `json:"description" bson:"description"`
	CareInstructions  *string            `json:"care_instructions" bson:"care_instructions"`
	Images            *[]string          `json:"images" bson:"images"`
	SeasonalAvailable *bool              `json:"seasonal_available" bson:"seasonal_available"` // Is this plant seasonal
	Effects           *[]string          `json:"effects" bson:"effects"`                       // relaxing, energizing, etc.
	MedicalBenefits   *[]string          `json:"medical_benefits" bson:"medical_benefits"`
	DataStatus        *enum.DataStatus   `json:"data_status" bson:"data_status"`
	TenantId          *enum.Tenant       `json:"tenant_id" bson:"tenant_id"`
	Metadata          *map[string]string `json:"metadata" bson:"metadata"` // Additional flexible metadata
}

type plantType struct {
	*repo
}

func newPlantType(ctx context.Context, collection *mongo.Collection) *plantType {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "strain", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "seasonal_available", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
				{Key: "strain", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create plant type indexes:", err)
	}

	return &plantType{newrepo(collection)}
}

func (s *plantType) Create(ctx context.Context, domain *PlantTypeDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.Save(ctx, domain.ID, domain)
	return err
}

func (s *plantType) Update(ctx context.Context, id string, domain *PlantTypeDomain) error {
	domain.BeforeSave()
	_, err := s.Save(ctx, OID(id), domain)
	return err
}

func (s *plantType) FindByID(ctx context.Context, id string) (*PlantTypeDomain, error) {
	var domain PlantTypeDomain
	err := s.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantType) FindByName(ctx context.Context, name string) (*PlantTypeDomain, error) {
	var domain PlantTypeDomain
	err := s.FindOne(ctx, M{"name": name}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantType) FindByStrain(ctx context.Context, strain string) (*PlantTypeDomain, error) {
	var domain PlantTypeDomain
	err := s.FindOne(ctx, M{"strain": strain}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantType) FindByCategory(ctx context.Context, category string, tenant enum.Tenant, offset, limit int64) ([]*PlantTypeDomain, error) {
	var domains []*PlantTypeDomain

	query := Query{
		Filter: M{
			"category":  category,
			"tenant_id": tenant,
		},
		Page:  offset/limit + 1,
		Limit: limit,
		Sorts: "name.asc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *plantType) FindAll(ctx context.Context, tenant enum.Tenant, offset, limit int64) ([]*PlantTypeDomain, error) {
	var domains []*PlantTypeDomain

	query := Query{
		Filter: M{"tenant_id": tenant},
		Page:   offset/limit + 1,
		Limit:  limit,
		Sorts:  "name.asc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *plantType) FindSeasonal(ctx context.Context, tenant enum.Tenant, offset, limit int64) ([]*PlantTypeDomain, error) {
	var domains []*PlantTypeDomain

	query := Query{
		Filter: M{
			"tenant_id":          tenant,
			"seasonal_available": true,
		},
		Page:  offset/limit + 1,
		Limit: limit,
		Sorts: "name.asc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *plantType) AddImage(ctx context.Context, id string, imageURL string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"images": imageURL},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *plantType) Delete(ctx context.Context, id string) error {
	return s.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plantType) Count(ctx context.Context, filter M) int64 {
	return s.CountDocuments(ctx, Query{Filter: filter})
}
