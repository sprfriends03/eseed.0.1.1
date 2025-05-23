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

func (s NFTRecordDomain) BaseDto() *NFTRecordBaseDto {
	return &NFTRecordBaseDto{
		ID:              SID(s.ID),
		TokenID:         gopkg.Value(s.TokenID),
		ContractAddress: gopkg.Value(s.ContractAddress),
		OwnerAddress:    gopkg.Value(s.OwnerAddress),
		MemberID:        gopkg.Value(s.MemberID),
		TokenType:       gopkg.Value(s.TokenType),
		MintDate:        gopkg.Value(s.MintDate),
		Status:          gopkg.Value(s.Status),
		ImageURL:        gopkg.Value(s.ImageURL),
	}
}

func (s NFTRecordDomain) DetailDto() *NFTRecordDetailDto {
	return &NFTRecordDetailDto{
		ID:                SID(s.ID),
		TokenID:           gopkg.Value(s.TokenID),
		ContractAddress:   gopkg.Value(s.ContractAddress),
		OwnerAddress:      gopkg.Value(s.OwnerAddress),
		MemberID:          gopkg.Value(s.MemberID),
		RelatedEntityID:   gopkg.Value(s.RelatedEntityID),
		RelatedEntityType: gopkg.Value(s.RelatedEntityType),
		TokenURI:          gopkg.Value(s.TokenURI),
		TokenType:         gopkg.Value(s.TokenType),
		ChainID:           gopkg.Value(s.ChainID),
		MintDate:          gopkg.Value(s.MintDate),
		Status:            gopkg.Value(s.Status),
		TransactionHash:   gopkg.Value(s.TransactionHash),
		BlockNumber:       gopkg.Value(s.BlockNumber),
		ImageURL:          gopkg.Value(s.ImageURL),
		CreatedAt:         gopkg.Value(s.CreatedAt),
		UpdatedAt:         gopkg.Value(s.UpdatedAt),
	}
}

type NFTRecordBaseDto struct {
	ID              string    `json:"nft_record_id"`
	TokenID         string    `json:"token_id"`
	ContractAddress string    `json:"contract_address"`
	OwnerAddress    string    `json:"owner_address"`
	MemberID        string    `json:"member_id"`
	TokenType       string    `json:"token_type"`
	MintDate        time.Time `json:"mint_date"`
	Status          string    `json:"status"`
	ImageURL        string    `json:"image_url,omitempty"`
}

type NFTRecordDetailDto struct {
	ID                string    `json:"nft_record_id"`
	TokenID           string    `json:"token_id"`
	ContractAddress   string    `json:"contract_address"`
	OwnerAddress      string    `json:"owner_address"`
	MemberID          string    `json:"member_id"`
	RelatedEntityID   string    `json:"related_entity_id,omitempty"`
	RelatedEntityType string    `json:"related_entity_type,omitempty"`
	TokenURI          string    `json:"token_uri,omitempty"`
	TokenType         string    `json:"token_type"`
	ChainID           int       `json:"chain_id"`
	MintDate          time.Time `json:"mint_date"`
	Status            string    `json:"status"`
	TransactionHash   string    `json:"transaction_hash,omitempty"`
	BlockNumber       int64     `json:"block_number,omitempty"`
	ImageURL          string    `json:"image_url,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type NFTRecordQuery struct {
	Query
	Search            *string      `json:"search" form:"search" validate:"omitempty"`
	TokenID           *string      `json:"token_id" form:"token_id" validate:"omitempty"`
	ContractAddress   *string      `json:"contract_address" form:"contract_address" validate:"omitempty"`
	OwnerAddress      *string      `json:"owner_address" form:"owner_address" validate:"omitempty"`
	MemberID          *string      `json:"member_id" form:"member_id" validate:"omitempty,len=24"`
	RelatedEntityID   *string      `json:"related_entity_id" form:"related_entity_id" validate:"omitempty,len=24"`
	RelatedEntityType *string      `json:"related_entity_type" form:"related_entity_type" validate:"omitempty"`
	TokenType         *string      `json:"token_type" form:"token_type" validate:"omitempty"`
	ChainID           *int         `json:"chain_id" form:"chain_id" validate:"omitempty"`
	Status            *string      `json:"status" form:"status" validate:"omitempty"`
	StartDate         *time.Time   `json:"start_date" form:"start_date" validate:"omitempty"`
	EndDate           *time.Time   `json:"end_date" form:"end_date" validate:"omitempty"`
	TenantId          *enum.Tenant `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *NFTRecordQuery) Build() *NFTRecordQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{
			{"token_id": Regex(gopkg.Value(s.Search))},
			{"transaction_hash": Regex(gopkg.Value(s.Search))},
		}
	}
	if s.TokenID != nil {
		s.Filter["token_id"] = s.TokenID
	}
	if s.ContractAddress != nil {
		s.Filter["contract_address"] = s.ContractAddress
	}
	if s.OwnerAddress != nil {
		s.Filter["owner_address"] = s.OwnerAddress
	}
	if s.MemberID != nil {
		s.Filter["member_id"] = s.MemberID
	}
	if s.RelatedEntityID != nil {
		s.Filter["related_entity_id"] = s.RelatedEntityID
	}
	if s.RelatedEntityType != nil {
		s.Filter["related_entity_type"] = s.RelatedEntityType
	}
	if s.TokenType != nil {
		s.Filter["token_type"] = s.TokenType
	}
	if s.ChainID != nil {
		s.Filter["chain_id"] = s.ChainID
	}
	if s.Status != nil {
		s.Filter["status"] = s.Status
	}
	if s.StartDate != nil || s.EndDate != nil {
		dateFilter := M{}
		if s.StartDate != nil {
			dateFilter["$gte"] = s.StartDate
		}
		if s.EndDate != nil {
			dateFilter["$lte"] = s.EndDate
		}
		s.Filter["mint_date"] = dateFilter
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	return s
}

type nftRecord struct {
	repo *repo
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

	return &nftRecord{repo: newrepo(collection)}
}

func (s *nftRecord) Create(ctx context.Context, domain *NFTRecordDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *nftRecord) Update(ctx context.Context, id string, domain *NFTRecordDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
	return err
}

func (s *nftRecord) FindByID(ctx context.Context, id string) (*NFTRecordDomain, error) {
	var domain NFTRecordDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *nftRecord) FindByTokenInfo(ctx context.Context, tokenID, contractAddress string) (*NFTRecordDomain, error) {
	var domain NFTRecordDomain
	query := &NFTRecordQuery{
		TokenID:         gopkg.Pointer(tokenID),
		ContractAddress: gopkg.Pointer(contractAddress),
	}

	err := s.repo.FindOne(ctx, query.Build().Filter, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *nftRecord) FindByRelatedEntity(ctx context.Context, entityID string, entityType string) (*NFTRecordDomain, error) {
	var domain NFTRecordDomain
	query := &NFTRecordQuery{
		RelatedEntityID:   gopkg.Pointer(entityID),
		RelatedEntityType: gopkg.Pointer(entityType),
	}

	err := s.repo.FindOne(ctx, query.Build().Filter, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *nftRecord) FindAll(ctx context.Context, q *NFTRecordQuery, opts ...*options.FindOptions) ([]*NFTRecordDomain, error) {
	domains := make([]*NFTRecordDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s *nftRecord) FindByMemberID(ctx context.Context, memberID string) ([]*NFTRecordDomain, error) {
	query := &NFTRecordQuery{
		MemberID: gopkg.Pointer(memberID),
		Query: Query{
			Sorts: "mint_date.desc",
		},
	}
	return s.FindAll(ctx, query)
}

func (s *nftRecord) FindByOwnerAddress(ctx context.Context, ownerAddress string) ([]*NFTRecordDomain, error) {
	query := &NFTRecordQuery{
		OwnerAddress: gopkg.Pointer(ownerAddress),
		Query: Query{
			Sorts: "mint_date.desc",
		},
	}
	return s.FindAll(ctx, query)
}

func (s *nftRecord) FindRecentlyMinted(ctx context.Context, tenant enum.Tenant, limit int64) ([]*NFTRecordDomain, error) {
	query := &NFTRecordQuery{
		TenantId: gopkg.Pointer(tenant),
		Query: Query{
			Limit: limit,
			Sorts: "mint_date.desc",
		},
	}
	return s.FindAll(ctx, query)
}

func (s *nftRecord) UpdateOwner(ctx context.Context, id string, newOwnerAddress string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"owner_address": newOwnerAddress,
			"status":        "transferred",
			"updated_at":    time.Now(),
		}},
	)
}

func (s *nftRecord) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"status":     status,
			"updated_at": time.Now(),
		}},
	)
}

func (s *nftRecord) UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"metadata":   metadata,
			"updated_at": time.Now(),
		}},
	)
}

func (s *nftRecord) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *nftRecord) Count(ctx context.Context, q *NFTRecordQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}
