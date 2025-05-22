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

type UserDomain struct {
	BaseDomain   `json:"inline"`
	Name         *string          `json:"name,omitempty" validate:"omitempty"`
	Phone        *string          `json:"phone,omitempty" validate:"omitempty,lowercase"`
	Email        *string          `json:"email,omitempty" validate:"omitempty,lowercase"`
	Username     *string          `json:"username,omitempty" validate:"omitempty,lowercase"`
	Password     *string          `json:"password,omitempty" validate:"omitempty"`
	DataStatus   *enum.DataStatus `json:"data_status,omitempty" validate:"omitempty,data_status"`
	RoleIds      *[]string        `json:"role_ids,omitempty" validate:"omitempty,dive,len=24"`
	IsRoot       *bool            `json:"is_root,omitempty" validate:"omitempty"`
	TenantId     *enum.Tenant     `json:"tenant_id,omitempty" validate:"omitempty,len=24"`
	VersionToken *int64           `json:"version_token,omitempty" validate:"omitempty"`
}

func (s *UserDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

func (s UserDomain) CmsDto() *UserCmsDto {
	return &UserCmsDto{
		ID:         SID(s.ID),
		Name:       gopkg.Value(s.Name),
		Phone:      gopkg.Value(s.Phone),
		Email:      gopkg.Value(s.Email),
		Username:   gopkg.Value(s.Username),
		DataStatus: gopkg.Value(s.DataStatus),
		RoleIds:    gopkg.Value(s.RoleIds),
		UpdatedBy:  gopkg.Value(s.UpdatedBy),
		UpdatedAt:  gopkg.Value(s.UpdatedAt),
	}
}

func (s *UserDomain) Cache() *UserCache {
	return &UserCache{
		ID:           SID(s.ID),
		Name:         gopkg.Value(s.Name),
		Phone:        gopkg.Value(s.Phone),
		Email:        gopkg.Value(s.Email),
		Username:     gopkg.Value(s.Username),
		DataStatus:   gopkg.Value(s.DataStatus),
		RoleIds:      gopkg.Value(s.RoleIds),
		IsRoot:       gopkg.Value(s.IsRoot),
		TenantId:     gopkg.Value(s.TenantId),
		VersionToken: gopkg.Value(s.VersionToken),
		Permissions:  make([]enum.Permission, 0),
	}
}

type UserCmsDto struct {
	ID         string          `json:"user_id" example:"671db9eca1f1b1bdbf3d4618"`
	Name       string          `json:"name" example:"Aloha"`
	Phone      string          `json:"phone" example:"0973123456"`
	Email      string          `json:"email" example:"aloha@email.com"`
	Username   string          `json:"username" example:"aloha"`
	DataStatus enum.DataStatus `json:"data_status"`
	RoleIds    []string        `json:"role_ids" example:"671db9eca1f1b1bdbf3d4617"`
	UpdatedBy  string          `json:"updated_by" example:"editor"`
	UpdatedAt  time.Time       `json:"updated_at" example:"2006-01-02T03:04:05Z"`
}

type UserCache struct {
	ID           string            `json:"user_id"`
	Name         string            `json:"name"`
	Phone        string            `json:"phone"`
	Email        string            `json:"email"`
	Username     string            `json:"username"`
	DataStatus   enum.DataStatus   `json:"data_status"`
	RoleIds      []string          `json:"role_ids"`
	IsRoot       bool              `json:"is_root"`
	TenantId     enum.Tenant       `json:"tenant_id"`
	VersionToken int64             `json:"version_token"`
	IsTenant     bool              `json:"is_tenant"`
	Permissions  []enum.Permission `json:"permissions"`
}

type UserCmsData struct {
	Name       string          `json:"name" validate:"required" example:"Aloha"`
	Phone      string          `json:"phone" validate:"required,lowercase" example:"0973123456"`
	Email      string          `json:"email" validate:"required,lowercase,email" example:"aloha@email.com"`
	Username   string          `json:"username" validate:"required,lowercase" example:"aloha"`
	DataStatus enum.DataStatus `json:"data_status" validate:"required,data_status"`
	RoleIds    []string        `json:"role_ids" validate:"required,min=1,dive,len=24" example:"671db9eca1f1b1bdbf3d4617"`
}

func (s UserCmsData) Domain(domain *UserDomain) *UserDomain {
	domain.Name = gopkg.Pointer(s.Name)
	domain.Phone = gopkg.Pointer(s.Phone)
	domain.Email = gopkg.Pointer(s.Email)
	domain.Username = gopkg.Pointer(s.Username)
	domain.DataStatus = gopkg.Pointer(s.DataStatus)
	domain.RoleIds = gopkg.Pointer(s.RoleIds)
	return domain
}

type UserQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	RoleId     *string          `json:"role_id" form:"role_id" validate:"omitempty,len=24"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
	TenantId   *enum.Tenant     `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *UserQuery) Build() *UserQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{{"name": Regex(gopkg.Value(s.Search))}, {"email": Regex(gopkg.Value(s.Search))}, {"phone": Regex(gopkg.Value(s.Search))}, {"username": Regex(gopkg.Value(s.Search))}}
	}
	if s.RoleId != nil {
		s.Filter["role_ids"] = s.RoleId
	}
	if s.DataStatus != nil {
		s.Filter["data_status"] = s.DataStatus
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	s.Filter["is_root"] = false
	return s
}

type UserCmsQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	RoleId     *string          `json:"role_id" form:"role_id" validate:"omitempty,len=24"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
}

func (s *UserCmsQuery) BuildCore() *UserQuery {
	return &UserQuery{Query: s.Query, Search: s.Search, RoleId: s.RoleId, DataStatus: s.DataStatus}
}

type user struct {
	repo *repo
}

func newUser(ctx context.Context, col *mongo.Collection) *user {
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"username": bson.M{"$exists": true, "$gt": ""}}),
	})
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"phone": bson.M{"$exists": true, "$gt": ""}}),
	})
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"email": bson.M{"$exists": true, "$gt": ""}}),
	})
	return &user{newrepo(col)}
}

func (s user) CollectionName() string { return s.repo.col.Name() }

func (s user) Save(ctx context.Context, domain *UserDomain, opts ...*options.UpdateOptions) (*UserDomain, error) {
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

func (s user) FindOne(ctx context.Context, filter M, opts ...*options.FindOneOptions) (*UserDomain, error) {
	domain := &UserDomain{}
	return domain, s.repo.FindOne(ctx, filter, &domain, opts...)
}

func (s user) UpdateOne(ctx context.Context, filter M, update M, opts ...*options.UpdateOptions) error {
	return s.repo.UpdateOne(ctx, filter, update, opts...)
}

func (s user) Count(ctx context.Context, q *UserQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}

func (s user) FindAll(ctx context.Context, q *UserQuery, opts ...*options.FindOptions) ([]*UserDomain, error) {
	domains := make([]*UserDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s user) FindOneById(ctx context.Context, id string) (*UserDomain, error) {
	return s.FindOne(ctx, M{"_id": OID(id)})
}

func (s user) FindAllByRole(ctx context.Context, roleId string) ([]*UserDomain, error) {
	return s.FindAll(ctx, &UserQuery{Query: Query{Filter: M{"role_ids": roleId}}})
}

func (s user) FindAllByTenant(ctx context.Context, tenant enum.Tenant) ([]*UserDomain, error) {
	return s.FindAll(ctx, &UserQuery{Query: Query{Filter: M{"tenant_id": tenant}}})
}

func (s user) FindOneByTenant_Username(ctx context.Context, tenant enum.Tenant, username string) (*UserDomain, error) {
	return s.FindOne(ctx, M{"tenant_id": tenant, "username": username})
}

func (s user) IncrementVersionToken(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx, M{"_id": OID(id)}, M{"$inc": M{"version_token": 1}})
}
