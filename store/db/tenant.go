package db

import (
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"time"

	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TenantDomain struct {
	BaseDomain `json:"inline"`
	Name       *string          `json:"name,omitempty" validate:"omitempty"`
	Keycode    *string          `json:"keycode,omitempty" validate:"omitempty,lowercase"`
	Username   *string          `json:"username,omitempty" validate:"omitempty,lowercase"`
	Phone      *string          `json:"phone,omitempty" validate:"omitempty,lowercase"`
	Email      *string          `json:"email,omitempty" validate:"omitempty,lowercase"`
	Address    *string          `json:"address,omitempty" validate:"omitempty"`
	DataStatus *enum.DataStatus `json:"data_status,omitempty" validate:"omitempty,data_status"`
	IsRoot     *bool            `json:"is_root,omitempty" validate:"omitempty"`
}

func (s *TenantDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

func (s TenantDomain) CmsDto() *TenantCmsDto {
	return &TenantCmsDto{
		ID:         SID(s.ID),
		Name:       gopkg.Value(s.Name),
		Keycode:    gopkg.Value(s.Keycode),
		Username:   gopkg.Value(s.Username),
		Phone:      gopkg.Value(s.Phone),
		Email:      gopkg.Value(s.Email),
		Address:    gopkg.Value(s.Address),
		DataStatus: gopkg.Value(s.DataStatus),
		UpdatedBy:  gopkg.Value(s.UpdatedBy),
		UpdatedAt:  gopkg.Value(s.UpdatedAt),
	}
}

func (s TenantDomain) Cache() *TenantCache {
	return &TenantCache{
		ID:         SID(s.ID),
		Keycode:    gopkg.Value(s.Keycode),
		DataStatus: gopkg.Value(s.DataStatus),
		IsRoot:     gopkg.Value(s.IsRoot),
	}
}

type TenantCmsDto struct {
	ID         string          `json:"tenant_id" example:"671dfc49f06ba89b1811cc5a"`
	Name       string          `json:"name" example:"Aloha"`
	Keycode    string          `json:"keycode" example:"aloha"`
	Username   string          `json:"username" example:"aloha"`
	Phone      string          `json:"phone" example:"0973123456"`
	Email      string          `json:"email" example:"aloha@email.com"`
	Address    string          `json:"address" example:"Aloha City"`
	DataStatus enum.DataStatus `json:"data_status"`
	UpdatedBy  string          `json:"updated_by" example:"editor"`
	UpdatedAt  time.Time       `json:"updated_at" example:"2006-01-02T03:04:05Z"`
}

type TenantCache struct {
	ID         string          `json:"tenant_id"`
	Keycode    string          `json:"keycode"`
	DataStatus enum.DataStatus `json:"data_status"`
	IsRoot     bool            `json:"is_root"`
}

type TenantCmsData struct {
	Name       string          `json:"name" validate:"required" example:"Aloha"`
	Keycode    string          `json:"keycode" validate:"required,lowercase" example:"aloha"`
	Username   string          `json:"username" validate:"required,lowercase" example:"aloha"`
	Phone      string          `json:"phone" validate:"required,lowercase" example:"0973123456"`
	Email      string          `json:"email" validate:"required,lowercase" example:"aloha@email.com"`
	Address    string          `json:"address" validate:"required" example:"Aloha City"`
	DataStatus enum.DataStatus `json:"data_status" validate:"required,data_status"`
}

func (s TenantCmsData) Domain(domain *TenantDomain) *TenantDomain {
	domain.Name = gopkg.Pointer(s.Name)
	domain.Keycode = gopkg.Pointer(s.Keycode)
	domain.Username = gopkg.Pointer(s.Username)
	domain.Phone = gopkg.Pointer(s.Phone)
	domain.Email = gopkg.Pointer(s.Email)
	domain.Address = gopkg.Pointer(s.Address)
	domain.DataStatus = gopkg.Pointer(s.DataStatus)
	return domain
}

type TenantQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
}

func (s *TenantQuery) Build() *TenantQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{{"name": Regex(gopkg.Value(s.Search))}}
	}
	if s.DataStatus != nil {
		s.Filter["data_status"] = s.DataStatus
	}
	s.Filter["is_root"] = false
	return s
}

type TenantCmsQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
}

func (s *TenantCmsQuery) BuildCore() *TenantQuery {
	return &TenantQuery{Query: s.Query, Search: s.Search, DataStatus: s.DataStatus}
}

type tenant struct {
	repo *repo
}

func newTenant(ctx context.Context, col *mongo.Collection) *tenant {
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "keycode", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &tenant{newrepo(col)}
}

func (s tenant) CollectionName() string { return s.repo.col.Name() }

func (s tenant) Save(ctx context.Context, domain *TenantDomain, opts ...*options.UpdateOptions) (*TenantDomain, error) {
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

func (s tenant) FindOne(ctx context.Context, filter M, opts ...*options.FindOneOptions) (*TenantDomain, error) {
	domain := &TenantDomain{}
	return domain, s.repo.FindOne(ctx, filter, &domain, opts...)
}

func (s tenant) Count(ctx context.Context, q *TenantQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}

func (s tenant) FindAll(ctx context.Context, q *TenantQuery, opts ...*options.FindOptions) ([]*TenantDomain, error) {
	domains := make([]*TenantDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s tenant) FindOneById(ctx context.Context, id string) (*TenantDomain, error) {
	return s.FindOne(ctx, M{"_id": OID(id)})
}

func (s tenant) FindOneByKeycode(ctx context.Context, keycode string) (*TenantDomain, error) {
	return s.FindOne(ctx, M{"keycode": keycode})
}
