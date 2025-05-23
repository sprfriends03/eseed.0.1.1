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

type NFTRecordDomain struct {
	BaseDomain        `bson:",inline"`
	TokenID           *string         `json:"token_id" bson:"token_id"`                       // NFT Token ID
	ContractAddress   *string         `json:"contract_address" bson:"contract_address"`       // Smart contract address
	OwnerAddress      *string         `json:"owner_address" bson:"owner_address"`             // Current owner's address
	MemberID          *string         `json:"member_id" bson:"member_id"`                     // Associated member ID
	RelatedEntityID   *string         `json:"related_entity_id" bson:"related_entity_id"`     // ID of related entity
	RelatedEntityType *string         `json:"related_entity_type" bson:"related_entity_type"` // Type of related entity (plant_slot, plant, harvest)
	TokenURI          *string         `json:"token_uri" bson:"token_uri"`                     // URI for the token metadata
	TokenType         *string         `json:"token_type" bson:"token_type"`                   // ERC721 or ERC1155
	ChainID           *int            `json:"chain_id" bson:"chain_id"`                       // Blockchain network ID
	MintDate          *time.Time      `json:"mint_date" bson:"mint_date"`                     // Date when NFT was minted
	Status            *string         `json:"status" bson:"status"`                           // minted, transferred, burned
	Metadata          *map[string]any `json:"metadata" bson:"metadata"`                       // Token metadata
	TransactionHash   *string         `json:"transaction_hash" bson:"transaction_hash"`       // Transaction hash of mint
	BlockNumber       *int64          `json:"block_number" bson:"block_number"`               // Block number of mint transaction
	ImageURL          *string         `json:"image_url" bson:"image_url"`                     // URL of associated image
	TenantId          *enum.Tenant    `json:"tenant_id" bson:"tenant_id"`
}

type nftRecord struct {
	*repo
}

func newNFTRecord(ctx context.Context, collection *mongo.Collection) *nftRecord {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_id", Value: 1}, {Key: "contract_address", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "owner_address", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "related_entity_id", Value: 1}, {Key: "related_entity_type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "mint_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create NFT record indexes:", err)
	}

	return &nftRecord{newrepo(collection)}
}

func (s *nftRecord) Create(ctx context.Context, domain *NFTRecordDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.Save(ctx, domain.ID, domain)
	return err
}

func (s *nftRecord) Update(ctx context.Context, id string, domain *NFTRecordDomain) error {
	domain.BeforeSave()
	_, err := s.Save(ctx, OID(id), domain)
	return err
}

func (s *nftRecord) FindByID(ctx context.Context, id string) (*NFTRecordDomain, error) {
	var domain NFTRecordDomain
	err := s.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *nftRecord) FindByTokenInfo(ctx context.Context, tokenID, contractAddress string) (*NFTRecordDomain, error) {
	var domain NFTRecordDomain
	err := s.FindOne(ctx, M{
		"token_id":         tokenID,
		"contract_address": contractAddress,
	}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *nftRecord) FindByRelatedEntity(ctx context.Context, entityID string, entityType string) (*NFTRecordDomain, error) {
	var domain NFTRecordDomain
	err := s.FindOne(ctx, M{
		"related_entity_id":   entityID,
		"related_entity_type": entityType,
	}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *nftRecord) FindByMemberID(ctx context.Context, memberID string, offset, limit int64) ([]*NFTRecordDomain, error) {
	var domains []*NFTRecordDomain

	query := Query{
		Filter: M{"member_id": memberID},
		Page:   offset/limit + 1,
		Limit:  limit,
		Sorts:  "mint_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *nftRecord) FindByOwnerAddress(ctx context.Context, ownerAddress string, offset, limit int64) ([]*NFTRecordDomain, error) {
	var domains []*NFTRecordDomain

	query := Query{
		Filter: M{"owner_address": ownerAddress},
		Page:   offset/limit + 1,
		Limit:  limit,
		Sorts:  "mint_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *nftRecord) FindRecentlyMinted(ctx context.Context, tenant enum.Tenant, limit int64) ([]*NFTRecordDomain, error) {
	var domains []*NFTRecordDomain

	query := Query{
		Filter: M{"tenant_id": tenant},
		Limit:  limit,
		Sorts:  "mint_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *nftRecord) UpdateOwner(ctx context.Context, id string, newOwnerAddress string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"owner_address": newOwnerAddress,
			"status":        "transferred",
			"updated_at":    time.Now(),
		}},
	)
}

func (s *nftRecord) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"status":     status,
			"updated_at": time.Now(),
		}},
	)
}

func (s *nftRecord) UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"metadata":   metadata,
			"updated_at": time.Now(),
		}},
	)
}

func (s *nftRecord) Delete(ctx context.Context, id string) error {
	return s.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *nftRecord) Count(ctx context.Context, filter M) int64 {
	return s.CountDocuments(ctx, Query{Filter: filter})
}
