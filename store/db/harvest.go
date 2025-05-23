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

type HarvestDomain struct {
	BaseDomain         `bson:",inline"`
	PlantID            *string      `json:"plant_id" bson:"plant_id"`         // Link to Plant
	MemberID           *string      `json:"member_id" bson:"member_id"`       // Link to Member
	HarvestDate        time.Time    `json:"harvest_date" bson:"harvest_date"` // When harvested
	Weight             *float64     `json:"weight" bson:"weight"`             // Weight in grams
	Quality            *int         `json:"quality" bson:"quality"`           // Quality rating (1-10)
	Images             *[]string    `json:"images" bson:"images"`             // Images of the harvest
	Strain             *string      `json:"strain" bson:"strain"`             // Cannabis strain (denormalized from Plant)
	Status             *string      `json:"status" bson:"status"`             // processing, curing, ready, collected
	NFTTokenID         *string      `json:"nft_token_id" bson:"nft_token_id"`
	NFTContractAddress *string      `json:"nft_contract_address" bson:"nft_contract_address"`
	Notes              *string      `json:"notes" bson:"notes"`
	CollectionDate     *time.Time   `json:"collection_date" bson:"collection_date"` // When member collected
	TenantId           *enum.Tenant `json:"tenant_id" bson:"tenant_id"`
}

type harvest struct {
	collection *mongo.Collection
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

	return &harvest{collection}
}

func (s *harvest) Create(ctx context.Context, domain *HarvestDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.collection.InsertOne(ctx, domain)
	return err
}

func (s *harvest) Update(ctx context.Context, id string, domain *HarvestDomain) error {
	domain.BeforeSave()

	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$set": domain},
	)

	return err
}

func (s *harvest) FindByID(ctx context.Context, id string) (*HarvestDomain, error) {
	var domain HarvestDomain
	err := s.collection.FindOne(ctx, bson.M{"_id": OID(id)}).Decode(&domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *harvest) FindByPlantID(ctx context.Context, plantID string) (*HarvestDomain, error) {
	var domain HarvestDomain
	err := s.collection.FindOne(ctx, bson.M{"plant_id": plantID}).Decode(&domain)
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

	cur, err := s.collection.Find(ctx,
		bson.M{"member_id": memberID},
		opts,
	)

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *harvest) FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	cur, err := s.collection.Find(ctx, bson.M{
		"status":    status,
		"tenant_id": tenantID,
	})

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *harvest) FindReadyForCollection(ctx context.Context, memberID string) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	cur, err := s.collection.Find(ctx, bson.M{
		"member_id": memberID,
		"status":    "ready",
	})

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *harvest) UpdateStatus(ctx context.Context, id string, status string) error {
	updates := bson.M{"status": status, "updated_at": time.Now()}

	if status == "collected" {
		updates["collection_date"] = time.Now()
	}

	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$set": updates},
	)

	return err
}

func (s *harvest) AddImage(ctx context.Context, id string, imageURL string) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{
			"$push": bson.M{"images": imageURL},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)

	return err
}

func (s *harvest) Delete(ctx context.Context, id string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": OID(id)})
	return err
}

func (s *harvest) Count(ctx context.Context, filter bson.M) (int64, error) {
	return s.collection.CountDocuments(ctx, filter)
}
