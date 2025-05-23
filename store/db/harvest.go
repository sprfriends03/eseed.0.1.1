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

type HarvestDomain struct {
	BaseDomain         `bson:",inline"`
	PlantID            *string      `json:"plant_id" bson:"plant_id" validate:"required,len=24"`     // Link to Plant
	MemberID           *string      `json:"member_id" bson:"member_id" validate:"required,len=24"`   // Link to Member
	HarvestDate        time.Time    `json:"harvest_date" bson:"harvest_date" validate:"required"`    // When harvested
	Weight             *float64     `json:"weight" bson:"weight" validate:"required,gt=0"`           // Weight in grams
	Quality            *int         `json:"quality" bson:"quality" validate:"required,gte=1,lte=10"` // Quality rating (1-10)
	Images             *[]string    `json:"images" bson:"images" validate:"omitempty,dive,required"` // Images of the harvest
	Strain             *string      `json:"strain" bson:"strain" validate:"required"`                // Cannabis strain (denormalized from Plant)
	Status             *string      `json:"status" bson:"status" validate:"required"`                // processing, curing, ready, collected
	NFTTokenID         *string      `json:"nft_token_id" bson:"nft_token_id" validate:"omitempty"`
	NFTContractAddress *string      `json:"nft_contract_address" bson:"nft_contract_address" validate:"omitempty"`
	Notes              *string      `json:"notes" bson:"notes" validate:"omitempty"`
	CollectionDate     *time.Time   `json:"collection_date" bson:"collection_date" validate:"omitempty"` // When member collected
	TenantId           *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *HarvestDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type harvest struct {
	repo *repo
}

func newHarvest(ctx context.Context, collection *mongo.Collection) *harvest {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "plant_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "harvest_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "strain", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "nft_token_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create harvest indexes:", err)
	}

	return &harvest{repo: newrepo(collection)}
}

func (s *harvest) Save(ctx context.Context, domain *HarvestDomain, opts ...*options.UpdateOptions) (*HarvestDomain, error) {
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

func (s *harvest) Create(ctx context.Context, domain *HarvestDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *harvest) Update(ctx context.Context, id string, domain *HarvestDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *harvest) FindByID(ctx context.Context, id string) (*HarvestDomain, error) {
	var domain HarvestDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *harvest) FindByPlantID(ctx context.Context, plantID string) (*HarvestDomain, error) {
	var domain HarvestDomain
	err := s.repo.FindOne(ctx, M{"plant_id": plantID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *harvest) FindByMemberID(ctx context.Context, memberID string, offset, limit int64) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	opts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(bson.D{{Key: "harvest_date", Value: -1}})

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *harvest) FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	query := Query{
		Filter: M{
			"status":    status,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *harvest) FindReadyForCollection(ctx context.Context, memberID string) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	query := Query{
		Filter: M{
			"member_id": memberID,
			"status":    "ready",
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *harvest) UpdateStatus(ctx context.Context, id string, status string) error {
	updates := M{"status": status, "updated_at": time.Now()}

	if status == "collected" {
		updates["collection_date"] = time.Now()
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}

func (s *harvest) AddImage(ctx context.Context, id string, imageURL string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"images": imageURL},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *harvest) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *harvest) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
