package db

import (
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"encoding/json"
	"time"

	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuditLogDomain struct {
	BaseDomain `bson:",inline"`
	Name       *string          `json:"name,omitempty" validate:"omitempty"`
	Url        *string          `json:"url,omitempty" validate:"omitempty"`
	Method     *string          `json:"method,omitempty" validate:"omitempty"`
	Body       *string          `json:"body,omitempty" validate:"omitempty"`
	Data       *[]byte          `json:"data,omitempty" validate:"omitempty"`
	Domain     *[]byte          `json:"domain,omitempty" validate:"omitempty"`
	DomainId   *string          `json:"domain_id,omitempty" validate:"omitempty,len=24"`
	UserId     *string          `json:"user_id,omitempty" validate:"omitempty,len=24"`
	UserEmail  *string          `json:"user_email,omitempty" validate:"omitempty"`
	UserName   *string          `json:"user_name,omitempty" validate:"omitempty"`
	Action     *enum.DataAction `json:"action,omitempty" validate:"omitempty,data_action"`
	TenantId   *enum.Tenant     `json:"tenant_id,omitempty" validate:"omitempty,len=24"`
}

func (s *AuditLogDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

func (s AuditLogDomain) CmsDto() *AuditLogCmsDto {
	data := M{}
	json.Unmarshal(gopkg.Value(s.Data), &data)

	return &AuditLogCmsDto{
		ID:        SID(s.ID),
		Name:      gopkg.Value(s.Name),
		Url:       gopkg.Value(s.Url),
		Method:    gopkg.Value(s.Method),
		Action:    gopkg.Value(s.Action),
		Data:      data,
		DomainId:  gopkg.Value(s.DomainId),
		UpdatedBy: gopkg.Value(s.UpdatedBy),
		UpdatedAt: gopkg.Value(s.UpdatedAt),
	}
}

type AuditLogCmsDto struct {
	ID        string          `json:"log_id" example:"812db9eca1f1b1bdbf3d4617"`
	Name      string          `json:"name" example:"Role"`
	Url       string          `json:"url" example:"/v1/cms/roles"`
	Method    string          `json:"method" example:"GET"`
	Action    enum.DataAction `json:"action"`
	Data      M               `json:"data"`
	DomainId  string          `json:"domain_id" example:"654db9eca1f1b1bdbf3d4617"`
	UpdatedBy string          `json:"updated_by" example:"editor"`
	UpdatedAt time.Time       `json:"updated_at" example:"2006-01-02T03:04:05Z"`
}

type AuditLogQuery struct {
	Query
	Search   *string      `json:"search" form:"search" validate:"omitempty"`
	Name     *string      `json:"name" form:"name" validate:"omitempty"`
	DomainId *string      `json:"domain_id" form:"domain_id" validate:"omitempty,len=24"`
	TenantId *enum.Tenant `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *AuditLogQuery) Build() *AuditLogQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{{"url": Regex(gopkg.Value(s.Search))}, {"name": Regex(gopkg.Value(s.Search))}, {"method": Regex(gopkg.Value(s.Search))}}
	}
	if s.Name != nil {
		s.Filter["name"] = s.Name
	}
	if s.DomainId != nil {
		s.Filter["domain_id"] = s.DomainId
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	return s
}

type AuditLogCmsQuery struct {
	Query
	Search   *string `json:"search" form:"search" validate:"omitempty"`
	Name     *string `json:"name" form:"name" validate:"omitempty"`
	DomainId *string `json:"domain_id" form:"domain_id" validate:"omitempty,len=24"`
}

func (s *AuditLogCmsQuery) BuildCore() *AuditLogQuery {
	return &AuditLogQuery{Query: s.Query, Search: s.Search, Name: s.Name, DomainId: s.DomainId}
}

type audit_log struct {
	repo *repo
}

func newAuditLog(ctx context.Context, col *mongo.Collection) *audit_log {
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "tenant_id", Value: 1}},
	})
	return &audit_log{newrepo(col)}
}

func (s audit_log) CollectionName() string { return s.repo.col.Name() }

func (s audit_log) Save(ctx context.Context, domain *AuditLogDomain, opts ...*options.UpdateOptions) (*AuditLogDomain, error) {
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

func (s audit_log) FindOne(ctx context.Context, filter M, opts ...*options.FindOneOptions) (*AuditLogDomain, error) {
	domain := &AuditLogDomain{}
	return domain, s.repo.FindOne(ctx, filter, &domain, opts...)
}

func (s audit_log) Count(ctx context.Context, q *AuditLogQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}

func (s audit_log) FindAll(ctx context.Context, q *AuditLogQuery, opts ...*options.FindOptions) ([]*AuditLogDomain, error) {
	domains := make([]*AuditLogDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s audit_log) FindOneById(ctx context.Context, id string) (*AuditLogDomain, error) {
	return s.FindOne(ctx, M{"_id": OID(id)})
}
