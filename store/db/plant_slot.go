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

type PlantSlotDomain struct {
	BaseDomain         `bson:",inline"`
	MembershipID       *string      `json:"membership_id" bson:"membership_id"` // Link to Membership
	MemberID           *string      `json:"member_id" bson:"member_id"`         // Link to Member (denormalized for query efficiency)
	SlotNumber         *string      `json:"slot_number" bson:"slot_number"`     // Unique identifier within a farm location
	Status             *string      `json:"status" bson:"status"`               // allocated, planted, harvested, available
	FarmLocation       *string      `json:"farm_location" bson:"farm_location"` // Physical farm location
	AreaDesignation    *string      `json:"area_designation" bson:"area_designation"`
	CurrentPlantID     *string      `json:"current_plant_id" bson:"current_plant_id"` // Reference to currently planted Plant
	NFTTokenID         *string      `json:"nft_token_id" bson:"nft_token_id"`
	NFTContractAddress *string      `json:"nft_contract_address" bson:"nft_contract_address"`
	TenantId           *enum.Tenant `json:"tenant_id" bson:"tenant_id"`
}

type plantSlot struct {
	collection *mongo.Collection
}

func newPlantSlot(ctx context.Context, collection *mongo.Collection) *plantSlot {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "membership_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "farm_location", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "current_plant_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "nft_token_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "farm_location", Value: 1},
				{Key: "slot_number", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create plant slot indexes:", err)
	}

	return &plantSlot{collection}
}

func (s *plantSlot) Create(ctx context.Context, domain *PlantSlotDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.collection.InsertOne(ctx, domain)
	return err
}

func (s *plantSlot) Update(ctx context.Context, id string, domain *PlantSlotDomain) error {
	domain.BeforeSave()

	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$set": domain},
	)

	return err
}

func (s *plantSlot) FindByID(ctx context.Context, id string) (*PlantSlotDomain, error) {
	var domain PlantSlotDomain
	err := s.collection.FindOne(ctx, bson.M{"_id": OID(id)}).Decode(&domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantSlot) FindByMembershipID(ctx context.Context, membershipID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	cur, err := s.collection.Find(ctx, bson.M{"membership_id": membershipID})

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *plantSlot) FindByMemberID(ctx context.Context, memberID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	cur, err := s.collection.Find(ctx, bson.M{"member_id": memberID})

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *plantSlot) FindAvailable(ctx context.Context, tenantID enum.Tenant) ([]string, error) {
	var domains []*PlantSlotDomain
	var slotNumbers []string

	cur, err := s.collection.Find(ctx, bson.M{
		"status":    "available",
		"tenant_id": tenantID,
	})

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &domains); err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if domain.SlotNumber != nil {
			slotNumbers = append(slotNumbers, *domain.SlotNumber)
		}
	}

	return slotNumbers, nil
}

func (s *plantSlot) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}},
	)

	return err
}

func (s *plantSlot) UpdateCurrentPlant(ctx context.Context, id string, plantID string) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": OID(id)},
		bson.M{"$set": bson.M{
			"current_plant_id": plantID,
			"status":           "planted",
			"updated_at":       time.Now(),
		}},
	)

	return err
}

func (s *plantSlot) FindByNFTInfo(ctx context.Context, tokenID string, contractAddress string) (*PlantSlotDomain, error) {
	var domain PlantSlotDomain
	err := s.collection.FindOne(ctx, bson.M{
		"nft_token_id":         tokenID,
		"nft_contract_address": contractAddress,
	}).Decode(&domain)

	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantSlot) Delete(ctx context.Context, id string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": OID(id)})
	return err
}

func (s *plantSlot) Count(ctx context.Context, filter bson.M) (int64, error) {
	return s.collection.CountDocuments(ctx, filter)
}
