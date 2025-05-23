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

type MemberDomain struct {
	BaseDomain          `bson:",inline"`
	Name                *string          `json:"name" bson:"name"`
	Email               *string          `json:"email" bson:"email"`
	Phone               *string          `json:"phone" bson:"phone"`
	Address             *string          `json:"address" bson:"address"`
	DateOfBirth         *time.Time       `json:"date_of_birth" bson:"date_of_birth"`
	KYCStatus           *string          `json:"kyc_status" bson:"kyc_status"` // pending, verified, rejected
	KYCDocuments        *[]string        `json:"kyc_documents" bson:"kyc_documents"`
	DataStatus          *enum.DataStatus `json:"data_status" bson:"data_status"`
	UserID              *string          `json:"user_id" bson:"user_id"` // Link to User collection
	TenantId            *enum.Tenant     `json:"tenant_id" bson:"tenant_id"`
	CurrentMembershipID *string          `json:"current_membership_id" bson:"current_membership_id"` // Link to current active membership
}

type member struct {
	repo *repo
}

func newMember(ctx context.Context, collection *mongo.Collection) *member {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "kyc_status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "current_membership_id", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create member indexes:", err)
	}

	return &member{repo: newrepo(collection)}
}

func (s *member) Create(ctx context.Context, domain *MemberDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *member) Update(ctx context.Context, id string, domain *MemberDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
	return err
}

func (s *member) FindByID(ctx context.Context, id string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindByEmail(ctx context.Context, email string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"email": email}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindByPhone(ctx context.Context, phone string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"phone": phone}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindByUserID(ctx context.Context, userID string) (*MemberDomain, error) {
	var domain MemberDomain
	err := s.repo.FindOne(ctx, M{"user_id": userID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *member) FindWithKYCStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*MemberDomain, error) {
	var domains []*MemberDomain
	filter := M{
		"kyc_status": status,
		"tenant_id":  tenantID,
	}

	return domains, s.repo.FindAll(ctx, Query{Filter: filter}, &domains)
}

func (s *member) FindByTenantID(ctx context.Context, tenantID enum.Tenant, offset, limit int64) ([]*MemberDomain, error) {
	var domains []*MemberDomain

	opts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	return domains, s.repo.FindAll(ctx, Query{Filter: M{"tenant_id": tenantID}}, &domains, opts)
}

func (s *member) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *member) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
