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

type CareRecordDomain struct {
	BaseDomain   `bson:",inline"`
	PlantID      *string    `json:"plant_id" bson:"plant_id" validate:"required,len=24"`   // Link to Plant
	MemberID     *string    `json:"member_id" bson:"member_id" validate:"required,len=24"` // Link to Member who performed care
	CareType     *string    `json:"care_type" bson:"care_type" validate:"required"`        // watering, fertilizing, pruning, etc.
	CareDate     *time.Time `json:"care_date" bson:"care_date" validate:"required"`
	Notes        *string    `json:"notes" bson:"notes" validate:"omitempty"`
	Images       *[]string  `json:"images" bson:"images" validate:"omitempty,dive,required"`
	Measurements *struct {
		Temperature *float64 `json:"temperature" bson:"temperature" validate:"omitempty"`         // in Celsius
		Humidity    *float64 `json:"humidity" bson:"humidity" validate:"omitempty,gte=0,lte=100"` // in percentage
		SoilPH      *float64 `json:"soil_ph" bson:"soil_ph" validate:"omitempty,gte=0,lte=14"`
		WaterAmount *float64 `json:"water_amount" bson:"water_amount" validate:"omitempty,gte=0"` // in milliliters
	} `json:"measurements" bson:"measurements" validate:"omitempty"`
	Products *[]string    `json:"products" bson:"products" validate:"omitempty,dive,required"` // fertilizers, pesticides used
	TenantId *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *CareRecordDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type careRecord struct {
	repo *repo
}

func newCareRecord(ctx context.Context, collection *mongo.Collection) *careRecord {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "plant_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "care_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "care_type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "plant_id", Value: 1},
				{Key: "care_date", Value: -1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create care record indexes:", err)
	}

	return &careRecord{repo: newrepo(collection)}
}

func (s *careRecord) Save(ctx context.Context, domain *CareRecordDomain, opts ...*options.UpdateOptions) (*CareRecordDomain, error) {
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

func (s *careRecord) Create(ctx context.Context, domain *CareRecordDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *careRecord) Update(ctx context.Context, id string, domain *CareRecordDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *careRecord) FindByID(ctx context.Context, id string) (*CareRecordDomain, error) {
	var domain CareRecordDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *careRecord) FindByPlantID(ctx context.Context, plantID string, limit int64) ([]*CareRecordDomain, error) {
	var domains []*CareRecordDomain

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "care_date", Value: -1}})

	query := Query{
		Filter: M{"plant_id": plantID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *careRecord) FindByMemberID(ctx context.Context, memberID string, limit int64) ([]*CareRecordDomain, error) {
	var domains []*CareRecordDomain

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "care_date", Value: -1}})

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *careRecord) FindByDateRange(ctx context.Context, plantID string, startDate, endDate time.Time) ([]*CareRecordDomain, error) {
	var domains []*CareRecordDomain

	query := Query{
		Filter: M{
			"plant_id": plantID,
			"care_date": M{
				"$gte": startDate,
				"$lte": endDate,
			},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "care_date", Value: -1}})

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *careRecord) FindByCareType(ctx context.Context, plantID string, careType string) ([]*CareRecordDomain, error) {
	var domains []*CareRecordDomain

	query := Query{
		Filter: M{
			"plant_id":  plantID,
			"care_type": careType,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "care_date", Value: -1}})

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *careRecord) AddImage(ctx context.Context, id string, imageURL string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"images": imageURL},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *careRecord) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *careRecord) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
