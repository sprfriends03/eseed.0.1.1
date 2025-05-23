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
	repo *repo
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

	return &plantSlot{repo: newrepo(collection)}
}

func (s *plantSlot) Create(ctx context.Context, domain *PlantSlotDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *plantSlot) Update(ctx context.Context, id string, domain *PlantSlotDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
	return err
}

func (s *plantSlot) FindByID(ctx context.Context, id string) (*PlantSlotDomain, error) {
	var domain PlantSlotDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantSlot) FindByMembershipID(ctx context.Context, membershipID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{"membership_id": membershipID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindByMemberID(ctx context.Context, memberID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindAvailable(ctx context.Context, tenantID enum.Tenant) ([]string, error) {
	var domains []*PlantSlotDomain
	var slotNumbers []string

	query := Query{
		Filter: M{
			"status":    "available",
			"tenant_id": tenantID,
		},
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
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
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"status": status, "updated_at": time.Now()}},
	)
}

func (s *plantSlot) UpdateCurrentPlant(ctx context.Context, id string, plantID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"current_plant_id": plantID,
			"status":           "planted",
			"updated_at":       time.Now(),
		}},
	)
}

func (s *plantSlot) FindByNFTInfo(ctx context.Context, tokenID string, contractAddress string) (*PlantSlotDomain, error) {
	var domain PlantSlotDomain
	err := s.repo.FindOne(ctx, M{
		"nft_token_id":         tokenID,
		"nft_contract_address": contractAddress,
	}, &domain)

	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantSlot) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plantSlot) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
