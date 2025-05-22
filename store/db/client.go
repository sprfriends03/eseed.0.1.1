package db

import (
	"app/pkg/encryption"
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"time"

	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientDomain struct {
	BaseDomain   `json:"inline"`
	Name         *string      `json:"name,omitempty" validate:"omitempty"`
	ClientId     *string      `json:"client_id,omitempty" validate:"omitempty"`
	ClientSecret *string      `json:"client_secret,omitempty" validate:"omitempty"`
	SecureKey    *string      `json:"secure_key,omitempty" validate:"omitempty"`
	IsRoot       *bool        `json:"is_root,omitempty" validate:"omitempty"`
	TenantId     *enum.Tenant `json:"tenant_id,omitempty" validate:"omitempty,len=24"`
}

func (s *ClientDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

func (s ClientDomain) CmsDto() *ClientCmsDto {
	return &ClientCmsDto{
		Name:         gopkg.Value(s.Name),
		ClientId:     gopkg.Value(s.ClientId),
		ClientSecret: encryption.Decrypt(gopkg.Value(s.ClientSecret), gopkg.Value(s.ClientId)),
		SecureKey:    encryption.Decrypt(gopkg.Value(s.SecureKey), gopkg.Value(s.ClientId)),
		UpdatedBy:    gopkg.Value(s.UpdatedBy),
		UpdatedAt:    gopkg.Value(s.UpdatedAt),
	}
}

func (s ClientDomain) Cache() *ClientCache {
	return &ClientCache{
		ClientId:     gopkg.Value(s.ClientId),
		ClientSecret: encryption.Decrypt(gopkg.Value(s.ClientSecret), gopkg.Value(s.ClientId)),
		SecureKey:    encryption.Decrypt(gopkg.Value(s.SecureKey), gopkg.Value(s.ClientId)),
		TenantId:     gopkg.Value(s.TenantId),
	}
}

type ClientCmsDto struct {
	Name         string    `json:"name" example:"Aloha"`
	ClientId     string    `json:"client_id" example:"ogy64Ji1E4VY0S8b99oGDlDCRk5ZO3"`
	ClientSecret string    `json:"client_secret" example:"oSAa14Q1Ne6iSqVLs4nfG7p12K6cyv67PyV3L509"`
	SecureKey    string    `json:"secure_key" example:"QX9f276HW4fyL38Jto0pi9WVa40yLRpW0jsKN033"`
	UpdatedBy    string    `json:"updated_by" example:"editor"`
	UpdatedAt    time.Time `json:"updated_at" example:"2006-01-02T03:04:05Z"`
}

type ClientCache struct {
	ClientId     string      `json:"client_id"`
	ClientSecret string      `json:"client_secret"`
	SecureKey    string      `json:"secure_key"`
	TenantId     enum.Tenant `json:"tenant_id"`
}

type ClientCmsData struct {
	Name string `json:"name" validate:"required" example:"Aloha"`
}

func (s ClientCmsData) Domain(domain *ClientDomain) *ClientDomain {
	domain.Name = gopkg.Pointer(s.Name)
	return domain
}

type ClientQuery struct {
	Query
	Search   *string      `json:"search" form:"search" validate:"omitempty"`
	TenantId *enum.Tenant `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *ClientQuery) Build() *ClientQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{{"name": Regex(gopkg.Value(s.Search))}}
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	s.Filter["is_root"] = false
	return s
}

type ClientCmsQuery struct {
	Query
	Search *string `json:"search" form:"search" validate:"omitempty"`
}

func (s *ClientCmsQuery) BuildCore() *ClientQuery {
	return &ClientQuery{Query: s.Query, Search: s.Search}
}

type client struct {
	repo *repo
}

func newClient(ctx context.Context, col *mongo.Collection) *client {
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "client_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "tenant_id", Value: 1}},
	})
	return &client{newrepo(col)}
}

func (s client) CollectionName() string { return s.repo.col.Name() }

func (s client) Save(ctx context.Context, domain *ClientDomain, opts ...*options.UpdateOptions) (*ClientDomain, error) {
	if err := domain.Validate(); err != nil {
		return nil, err
	}
	id, err := s.repo.Save(ctx, domain.ID, domain, opts...)
	if err != nil {
		return nil, err
	}
	domain.ID = id
	return s.FindOneById(ctx, SID(id))
}

func (s client) FindOne(ctx context.Context, filter M, opts ...*options.FindOneOptions) (*ClientDomain, error) {
	domain := &ClientDomain{}
	return domain, s.repo.FindOne(ctx, filter, &domain, opts...)
}

func (s client) Count(ctx context.Context, q *ClientQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}

func (s client) FindAll(ctx context.Context, q *ClientQuery, opts ...*options.FindOptions) ([]*ClientDomain, error) {
	domains := make([]*ClientDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s client) FindOneById(ctx context.Context, id string) (*ClientDomain, error) {
	return s.FindOne(ctx, M{"_id": OID(id)})
}

func (s client) FindOneByClientId(ctx context.Context, clientId string) (*ClientDomain, error) {
	return s.FindOne(ctx, M{"client_id": clientId})
}

func (s client) FindAllByTenant(ctx context.Context, tenant enum.Tenant) ([]*ClientDomain, error) {
	return s.FindAll(ctx, &ClientQuery{Query: Query{Filter: M{"tenant_id": tenant}}})
}

func (s client) DeleteOne(ctx context.Context, domain *ClientDomain) error {
	return s.repo.DeleteOne(ctx, M{"_id": domain.ID})
}
