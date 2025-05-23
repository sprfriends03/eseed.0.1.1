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

func (s PlantTypeDomain) BaseDto() *PlantTypeBaseDto {
	return &PlantTypeBaseDto{
		ID:                SID(s.ID),
		Name:              gopkg.Value(s.Name),
		Strain:            gopkg.Value(s.Strain),
		Category:          gopkg.Value(s.Category),
		ThcContent:        gopkg.Value(s.ThcContent),
		CbdContent:        gopkg.Value(s.CbdContent),
		GrowthDifficulty:  gopkg.Value(s.GrowthDifficulty),
		AverageYield:      gopkg.Value(s.AverageYield),
		SeasonalAvailable: gopkg.Value(s.SeasonalAvailable),
		UpdatedAt:         gopkg.Value(s.UpdatedAt),
	}
}

func (s PlantTypeDomain) DetailDto() *PlantTypeDetailDto {
	return &PlantTypeDetailDto{
		ID:                SID(s.ID),
		Name:              gopkg.Value(s.Name),
		Strain:            gopkg.Value(s.Strain),
		Category:          gopkg.Value(s.Category),
		ThcContent:        gopkg.Value(s.ThcContent),
		CbdContent:        gopkg.Value(s.CbdContent),
		GrowthDifficulty:  gopkg.Value(s.GrowthDifficulty),
		AverageYield:      gopkg.Value(s.AverageYield),
		FloweringTime:     gopkg.Value(s.FloweringTime),
		GrowthPhases:      gopkg.Value(s.GrowthPhases),
		Description:       gopkg.Value(s.Description),
		CareInstructions:  gopkg.Value(s.CareInstructions),
		Images:            gopkg.Value(s.Images),
		SeasonalAvailable: gopkg.Value(s.SeasonalAvailable),
		Effects:           gopkg.Value(s.Effects),
		MedicalBenefits:   gopkg.Value(s.MedicalBenefits),
		Metadata:          gopkg.Value(s.Metadata),
		UpdatedAt:         gopkg.Value(s.UpdatedAt),
	}
}

type PlantTypeBaseDto struct {
	ID                string    `json:"plant_type_id"`
	Name              string    `json:"name"`
	Strain            string    `json:"strain"`
	Category          string    `json:"category"`
	ThcContent        float64   `json:"thc_content"`
	CbdContent        float64   `json:"cbd_content"`
	GrowthDifficulty  int       `json:"growth_difficulty"`
	AverageYield      float64   `json:"average_yield"`
	SeasonalAvailable bool      `json:"seasonal_available"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PlantTypeDetailDto struct {
	ID                string            `json:"plant_type_id"`
	Name              string            `json:"name"`
	Strain            string            `json:"strain"`
	Category          string            `json:"category"`
	ThcContent        float64           `json:"thc_content"`
	CbdContent        float64           `json:"cbd_content"`
	GrowthDifficulty  int               `json:"growth_difficulty"`
	AverageYield      float64           `json:"average_yield"`
	FloweringTime     int               `json:"flowering_time"`
	GrowthPhases      []string          `json:"growth_phases"`
	Description       string            `json:"description"`
	CareInstructions  string            `json:"care_instructions"`
	Images            []string          `json:"images"`
	SeasonalAvailable bool              `json:"seasonal_available"`
	Effects           []string          `json:"effects"`
	MedicalBenefits   []string          `json:"medical_benefits"`
	Metadata          map[string]string `json:"metadata"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type PlantTypeQuery struct {
	Query
	Search        *string          `json:"search" form:"search" validate:"omitempty"`
	Category      *string          `json:"category" form:"category" validate:"omitempty"`
	Strain        *string          `json:"strain" form:"strain" validate:"omitempty"`
	SeasonalOnly  *bool            `json:"seasonal_only" form:"seasonal_only" validate:"omitempty"`
	MinThcContent *float64         `json:"min_thc_content" form:"min_thc_content" validate:"omitempty"`
	MaxThcContent *float64         `json:"max_thc_content" form:"max_thc_content" validate:"omitempty"`
	DataStatus    *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
	TenantId      *enum.Tenant     `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *PlantTypeQuery) Build() *PlantTypeQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{
			{"name": Regex(gopkg.Value(s.Search))},
			{"strain": Regex(gopkg.Value(s.Search))},
			{"category": Regex(gopkg.Value(s.Search))},
		}
	}
	if s.Category != nil {
		s.Filter["category"] = s.Category
	}
	if s.Strain != nil {
		s.Filter["strain"] = s.Strain
	}
	if s.SeasonalOnly != nil && *s.SeasonalOnly {
		s.Filter["seasonal_available"] = true
	}
	if s.MinThcContent != nil {
		s.Filter["thc_content"] = M{"$gte": s.MinThcContent}
	}
	if s.MaxThcContent != nil {
		if s.Filter["thc_content"] == nil {
			s.Filter["thc_content"] = M{"$lte": s.MaxThcContent}
		} else {
			s.Filter["thc_content"].(M)["$lte"] = s.MaxThcContent
		}
	}
	if s.DataStatus != nil {
		s.Filter["data_status"] = s.DataStatus
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	return s
}

type plantType struct {
	repo *repo
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

	return &plantType{repo: newrepo(collection)}
}

func (s *plantType) Create(ctx context.Context, domain *PlantTypeDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *plantType) Update(ctx context.Context, id string, domain *PlantTypeDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
	return err
}

func (s *plantType) FindByID(ctx context.Context, id string) (*PlantTypeDomain, error) {
	var domain PlantTypeDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantType) FindByName(ctx context.Context, name string) (*PlantTypeDomain, error) {
	var domain PlantTypeDomain
	err := s.repo.FindOne(ctx, M{"name": name}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantType) FindByStrain(ctx context.Context, strain string) (*PlantTypeDomain, error) {
	var domain PlantTypeDomain
	err := s.repo.FindOne(ctx, M{"strain": strain}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantType) FindByCategory(ctx context.Context, category string, tenant enum.Tenant) ([]*PlantTypeDomain, error) {
	query := &PlantTypeQuery{
		Category: gopkg.Pointer(category),
		TenantId: gopkg.Pointer(tenant),
	}
	return s.FindAll(ctx, query)
}

func (s *plantType) FindAll(ctx context.Context, q *PlantTypeQuery, opts ...*options.FindOptions) ([]*PlantTypeDomain, error) {
	domains := make([]*PlantTypeDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s *plantType) FindSeasonal(ctx context.Context, tenant enum.Tenant) ([]*PlantTypeDomain, error) {
	query := &PlantTypeQuery{
		SeasonalOnly: gopkg.Pointer(true),
		TenantId:     gopkg.Pointer(tenant),
	}
	return s.FindAll(ctx, query)
}

func (s *plantType) AddImage(ctx context.Context, id string, imageURL string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"images": imageURL},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *plantType) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plantType) Count(ctx context.Context, q *PlantTypeQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}
