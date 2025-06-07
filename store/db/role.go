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

type RoleDomain struct {
	BaseDomain  `bson:",inline"`
	Name        *string            `json:"name,omitempty" validate:"omitempty"`
	Permissions *[]enum.Permission `json:"permissions,omitempty" validate:"omitempty,dive,permission"`
	DataStatus  *enum.DataStatus   `json:"data_status,omitempty" validate:"omitempty,data_status"`
	TenantId    *enum.Tenant       `json:"tenant_id,omitempty" validate:"omitempty,len=24"`
}

func (s *RoleDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

func (s RoleDomain) BaseDto() *RoleBaseDto {
	return &RoleBaseDto{
		ID:         SID(s.ID),
		Name:       gopkg.Value(s.Name),
		DataStatus: gopkg.Value(s.DataStatus),
	}
}

func (s RoleDomain) CmsDto() *RoleCmsDto {
	return &RoleCmsDto{
		ID:          SID(s.ID),
		Name:        gopkg.Value(s.Name),
		Permissions: gopkg.Value(s.Permissions),
		DataStatus:  gopkg.Value(s.DataStatus),
		UpdatedBy:   gopkg.Value(s.UpdatedBy),
		UpdatedAt:   gopkg.Value(s.UpdatedAt),
	}
}

type RoleBaseDto struct {
	ID         string          `json:"role_id" example:"671db9eca1f1b1bdbf3d4617"`
	Name       string          `json:"name" example:"Aloha"`
	DataStatus enum.DataStatus `json:"data_status" example:"active"`
}

type RoleCmsDto struct {
	ID          string            `json:"role_id" example:"671db9eca1f1b1bdbf3d4617"`
	Name        string            `json:"name" example:"Aloha"`
	Permissions []enum.Permission `json:"permissions"`
	DataStatus  enum.DataStatus   `json:"data_status"`
	UpdatedBy   string            `json:"updated_by" example:"editor"`
	UpdatedAt   time.Time         `json:"updated_at" example:"2006-01-02T03:04:05Z"`
}

type RoleCmsData struct {
	Name        string            `json:"name" validate:"required" example:"Aloha"`
	Permissions []enum.Permission `json:"permissions" validate:"required,min=1,dive,permission"`
	DataStatus  enum.DataStatus   `json:"data_status" validate:"required,data_status"`
}

func (s RoleCmsData) Domain(domain *RoleDomain) *RoleDomain {
	domain.Name = gopkg.Pointer(s.Name)
	domain.Permissions = gopkg.Pointer(s.Permissions)
	domain.DataStatus = gopkg.Pointer(s.DataStatus)
	return domain
}

type RoleQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
	TenantId   *enum.Tenant     `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *RoleQuery) Build() *RoleQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{{"name": Regex(gopkg.Value(s.Search))}}
	}
	if s.DataStatus != nil {
		s.Filter["data_status"] = s.DataStatus
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	return s
}

type RoleCmsQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
}

func (s *RoleCmsQuery) BuildCore() *RoleQuery {
	return &RoleQuery{Query: s.Query, Search: s.Search, DataStatus: s.DataStatus}
}

type role struct {
	repo *repo
}

func newRole(ctx context.Context, col *mongo.Collection) *role {
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "tenant_id", Value: 1}},
	})
	return &role{newrepo(col)}
}

func (s role) CollectionName() string { return s.repo.col.Name() }

func (s role) Save(ctx context.Context, domain *RoleDomain, opts ...*options.UpdateOptions) (*RoleDomain, error) {
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

func (s role) FindOne(ctx context.Context, filter M, opts ...*options.FindOneOptions) (*RoleDomain, error) {
	domain := &RoleDomain{}
	return domain, s.repo.FindOne(ctx, filter, &domain, opts...)
}

func (s role) Count(ctx context.Context, q *RoleQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}

func (s role) FindAll(ctx context.Context, q *RoleQuery, opts ...*options.FindOptions) ([]*RoleDomain, error) {
	domains := make([]*RoleDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s role) FindOneById(ctx context.Context, id string) (*RoleDomain, error) {
	return s.FindOne(ctx, M{"_id": OID(id)})
}

func (s role) FindAllByIds(ctx context.Context, ids []string) ([]*RoleDomain, error) {
	return s.FindAll(ctx, &RoleQuery{Query: Query{Filter: M{"_id": M{"$in": OIDs(ids)}}}})
}
