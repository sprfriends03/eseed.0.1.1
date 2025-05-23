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

type PaymentDomain struct {
	BaseDomain        `bson:",inline"`
	MemberID          *string         `json:"member_id" bson:"member_id"`                     // Link to Member
	RelatedEntityID   *string         `json:"related_entity_id" bson:"related_entity_id"`     // Link to Membership or other related entity
	RelatedEntityType *string         `json:"related_entity_type" bson:"related_entity_type"` // Type of related entity (e.g., "membership")
	Amount            *float64        `json:"amount" bson:"amount"`                           // Payment amount
	Currency          *string         `json:"currency" bson:"currency"`                       // Currency code (e.g., EUR)
	Status            *string         `json:"status" bson:"status"`                           // pending, completed, failed, refunded
	PaymentMethod     *string         `json:"payment_method" bson:"payment_method"`           // credit_card, bank_transfer, crypto, etc.
	PaymentDate       *time.Time      `json:"payment_date" bson:"payment_date"`
	ExternalID        *string         `json:"external_id" bson:"external_id"`       // ID from payment processor
	ProcessorName     *string         `json:"processor_name" bson:"processor_name"` // Name of payment processor
	BillingAddress    *Address        `json:"billing_address" bson:"billing_address"`
	InvoiceNumber     *string         `json:"invoice_number" bson:"invoice_number"`
	ReceiptURL        *string         `json:"receipt_url" bson:"receipt_url"`
	Notes             *string         `json:"notes" bson:"notes"`
	RefundID          *string         `json:"refund_id" bson:"refund_id"` // If refunded, reference to refund
	Metadata          *map[string]any `json:"metadata" bson:"metadata"`   // Additional data from payment processor
	TenantId          *enum.Tenant    `json:"tenant_id" bson:"tenant_id"`
}

// Address represents a billing address
type Address struct {
	Street     *string `json:"street" bson:"street"`
	City       *string `json:"city" bson:"city"`
	State      *string `json:"state" bson:"state"`
	PostalCode *string `json:"postal_code" bson:"postal_code"`
	Country    *string `json:"country" bson:"country"`
}

type payment struct {
	*repo
}

func newPayment(ctx context.Context, collection *mongo.Collection) *payment {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "related_entity_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "payment_date", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "external_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "invoice_number", Value: 1}},
			Options: options.Index().SetSparse(true).SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create payment indexes:", err)
	}

	return &payment{newrepo(collection)}
}

func (s *payment) Create(ctx context.Context, domain *PaymentDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.Save(ctx, domain.ID, domain)
	return err
}

func (s *payment) Update(ctx context.Context, id string, domain *PaymentDomain) error {
	domain.BeforeSave()
	_, err := s.Save(ctx, OID(id), domain)
	return err
}

func (s *payment) FindByID(ctx context.Context, id string) (*PaymentDomain, error) {
	var domain PaymentDomain
	err := s.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *payment) FindByExternalID(ctx context.Context, externalID string) (*PaymentDomain, error) {
	var domain PaymentDomain
	err := s.FindOne(ctx, M{"external_id": externalID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *payment) FindByInvoiceNumber(ctx context.Context, invoiceNumber string) (*PaymentDomain, error) {
	var domain PaymentDomain
	err := s.FindOne(ctx, M{"invoice_number": invoiceNumber}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *payment) FindByMemberID(ctx context.Context, memberID string, offset, limit int64) ([]*PaymentDomain, error) {
	var domains []*PaymentDomain

	query := Query{
		Filter: M{"member_id": memberID},
		Page:   offset/limit + 1,
		Limit:  limit,
		Sorts:  "payment_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *payment) FindByRelatedEntity(ctx context.Context, entityID string, entityType string) ([]*PaymentDomain, error) {
	var domains []*PaymentDomain

	query := Query{
		Filter: M{
			"related_entity_id":   entityID,
			"related_entity_type": entityType,
		},
		Sorts: "payment_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *payment) FindByStatus(ctx context.Context, status string, tenant enum.Tenant, offset, limit int64) ([]*PaymentDomain, error) {
	var domains []*PaymentDomain

	query := Query{
		Filter: M{
			"status":    status,
			"tenant_id": tenant,
		},
		Page:  offset/limit + 1,
		Limit: limit,
		Sorts: "payment_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *payment) FindByDateRange(ctx context.Context, startDate, endDate time.Time, tenant enum.Tenant) ([]*PaymentDomain, error) {
	var domains []*PaymentDomain

	query := Query{
		Filter: M{
			"tenant_id": tenant,
			"payment_date": M{
				"$gte": startDate,
				"$lte": endDate,
			},
		},
		Sorts: "payment_date.desc",
	}

	err := s.repo.FindAll(ctx, query, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *payment) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"status":     status,
			"updated_at": time.Now(),
		}},
	)
}

func (s *payment) RefundPayment(ctx context.Context, id string, refundID string) error {
	return s.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"status":     "refunded",
			"refund_id":  refundID,
			"updated_at": time.Now(),
		}},
	)
}

func (s *payment) Delete(ctx context.Context, id string) error {
	return s.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *payment) Count(ctx context.Context, filter M) int64 {
	return s.CountDocuments(ctx, Query{Filter: filter})
}
