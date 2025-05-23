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

type CareRecordDomain struct {
	BaseDomain   `bson:",inline"`
	PlantID      *string         `json:"plant_id" bson:"plant_id"`         // Link to Plant
	MemberID     *string         `json:"member_id" bson:"member_id"`       // Link to Member who performed care
	CareDate     time.Time       `json:"care_date" bson:"care_date"`       // When the care was performed
	CareType     *string         `json:"care_type" bson:"care_type"`       // watering, pruning, fertilizing, pest control, etc.
	Description  *string         `json:"description" bson:"description"`   // Description of care performed
	Images       *[]string       `json:"images" bson:"images"`             // Images of the care/plant
	Measurements *map[string]any `json:"measurements" bson:"measurements"` // Flexible measurements (height, pH, etc.)
	Health       *int            `json:"health" bson:"health"`             // Health rating (1-10)
	Notes        *string         `json:"notes" bson:"notes"`
	TenantId     *enum.Tenant    `json:"tenant_id" bson:"tenant_id"`
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

func (s *careRecord) Create(ctx context.Context, domain *CareRecordDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *careRecord) Update(ctx context.Context, id string, domain *CareRecordDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
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

func (s *careRecord) FindByPlantID(ctx context.Context, plantID string, offset, limit int64) ([]*CareRecordDomain, error) {
	var domains []*CareRecordDomain

	opts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(bson.D{{Key: "care_date", Value: -1}})

	query := Query{
		Filter: M{"plant_id": plantID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *careRecord) FindByMemberID(ctx context.Context, memberID string, offset, limit int64) ([]*CareRecordDomain, error) {
	var domains []*CareRecordDomain

	opts := options.Find().
		SetSkip(offset).
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

	return domains, s.repo.FindAll(ctx, query, &domains)
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
