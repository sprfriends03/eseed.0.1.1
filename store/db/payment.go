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

func (s PaymentDomain) BaseDto() *PaymentBaseDto {
	return &PaymentBaseDto{
		ID:            SID(s.ID),
		MemberID:      gopkg.Value(s.MemberID),
		Amount:        gopkg.Value(s.Amount),
		Currency:      gopkg.Value(s.Currency),
		Status:        gopkg.Value(s.Status),
		PaymentMethod: gopkg.Value(s.PaymentMethod),
		PaymentDate:   gopkg.Value(s.PaymentDate),
		InvoiceNumber: gopkg.Value(s.InvoiceNumber),
	}
}

func (s PaymentDomain) DetailDto() *PaymentDetailDto {
	return &PaymentDetailDto{
		ID:                SID(s.ID),
		MemberID:          gopkg.Value(s.MemberID),
		RelatedEntityID:   gopkg.Value(s.RelatedEntityID),
		RelatedEntityType: gopkg.Value(s.RelatedEntityType),
		Amount:            gopkg.Value(s.Amount),
		Currency:          gopkg.Value(s.Currency),
		Status:            gopkg.Value(s.Status),
		PaymentMethod:     gopkg.Value(s.PaymentMethod),
		PaymentDate:       gopkg.Value(s.PaymentDate),
		ExternalID:        gopkg.Value(s.ExternalID),
		ProcessorName:     gopkg.Value(s.ProcessorName),
		BillingAddress:    s.BillingAddress,
		InvoiceNumber:     gopkg.Value(s.InvoiceNumber),
		ReceiptURL:        gopkg.Value(s.ReceiptURL),
		Notes:             gopkg.Value(s.Notes),
		RefundID:          gopkg.Value(s.RefundID),
		CreatedAt:         gopkg.Value(s.CreatedAt),
		UpdatedAt:         gopkg.Value(s.UpdatedAt),
	}
}

type PaymentBaseDto struct {
	ID            string    `json:"payment_id"`
	MemberID      string    `json:"member_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	PaymentDate   time.Time `json:"payment_date"`
	InvoiceNumber string    `json:"invoice_number,omitempty"`
}

type PaymentDetailDto struct {
	ID                string    `json:"payment_id"`
	MemberID          string    `json:"member_id"`
	RelatedEntityID   string    `json:"related_entity_id,omitempty"`
	RelatedEntityType string    `json:"related_entity_type,omitempty"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	PaymentMethod     string    `json:"payment_method"`
	PaymentDate       time.Time `json:"payment_date"`
	ExternalID        string    `json:"external_id,omitempty"`
	ProcessorName     string    `json:"processor_name,omitempty"`
	BillingAddress    *Address  `json:"billing_address,omitempty"`
	InvoiceNumber     string    `json:"invoice_number,omitempty"`
	ReceiptURL        string    `json:"receipt_url,omitempty"`
	Notes             string    `json:"notes,omitempty"`
	RefundID          string    `json:"refund_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Address represents a billing address
type Address struct {
	Street     *string `json:"street" bson:"street"`
	City       *string `json:"city" bson:"city"`
	State      *string `json:"state" bson:"state"`
	PostalCode *string `json:"postal_code" bson:"postal_code"`
	Country    *string `json:"country" bson:"country"`
}

type PaymentQuery struct {
	Query
	Search            *string      `json:"search" form:"search" validate:"omitempty"`
	MemberID          *string      `json:"member_id" form:"member_id" validate:"omitempty,len=24"`
	RelatedEntityID   *string      `json:"related_entity_id" form:"related_entity_id" validate:"omitempty,len=24"`
	RelatedEntityType *string      `json:"related_entity_type" form:"related_entity_type" validate:"omitempty"`
	Status            *string      `json:"status" form:"status" validate:"omitempty"`
	PaymentMethod     *string      `json:"payment_method" form:"payment_method" validate:"omitempty"`
	MinAmount         *float64     `json:"min_amount" form:"min_amount" validate:"omitempty"`
	MaxAmount         *float64     `json:"max_amount" form:"max_amount" validate:"omitempty"`
	StartDate         *time.Time   `json:"start_date" form:"start_date" validate:"omitempty"`
	EndDate           *time.Time   `json:"end_date" form:"end_date" validate:"omitempty"`
	TenantId          *enum.Tenant `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *PaymentQuery) Build() *PaymentQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{
			{"invoice_number": Regex(gopkg.Value(s.Search))},
			{"external_id": Regex(gopkg.Value(s.Search))},
			{"notes": Regex(gopkg.Value(s.Search))},
		}
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
	if s.Status != nil {
		s.Filter["status"] = s.Status
	}
	if s.PaymentMethod != nil {
		s.Filter["payment_method"] = s.PaymentMethod
	}
	if s.MinAmount != nil {
		s.Filter["amount"] = M{"$gte": s.MinAmount}
	}
	if s.MaxAmount != nil {
		if s.Filter["amount"] == nil {
			s.Filter["amount"] = M{"$lte": s.MaxAmount}
		} else {
			s.Filter["amount"].(M)["$lte"] = s.MaxAmount
		}
	}
	if s.StartDate != nil || s.EndDate != nil {
		dateFilter := M{}
		if s.StartDate != nil {
			dateFilter["$gte"] = s.StartDate
		}
		if s.EndDate != nil {
			dateFilter["$lte"] = s.EndDate
		}
		s.Filter["payment_date"] = dateFilter
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	return s
}

type payment struct {
	repo *repo
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

	return &payment{repo: newrepo(collection)}
}

func (s *payment) Create(ctx context.Context, domain *PaymentDomain) error {
	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}
	domain.BeforeSave()

	_, err := s.repo.Save(ctx, domain.ID, domain)
	return err
}

func (s *payment) Update(ctx context.Context, id string, domain *PaymentDomain) error {
	domain.BeforeSave()
	_, err := s.repo.Save(ctx, OID(id), domain)
	return err
}

func (s *payment) FindByID(ctx context.Context, id string) (*PaymentDomain, error) {
	var domain PaymentDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *payment) FindByExternalID(ctx context.Context, externalID string) (*PaymentDomain, error) {
	var domain PaymentDomain
	err := s.repo.FindOne(ctx, M{"external_id": externalID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *payment) FindByInvoiceNumber(ctx context.Context, invoiceNumber string) (*PaymentDomain, error) {
	var domain PaymentDomain
	err := s.repo.FindOne(ctx, M{"invoice_number": invoiceNumber}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *payment) FindByMemberID(ctx context.Context, memberID string) ([]*PaymentDomain, error) {
	query := &PaymentQuery{
		MemberID: gopkg.Pointer(memberID),
		Query: Query{
			Sorts: "payment_date.desc",
		},
	}
	return s.FindAll(ctx, query)
}

func (s *payment) FindAll(ctx context.Context, q *PaymentQuery, opts ...*options.FindOptions) ([]*PaymentDomain, error) {
	domains := make([]*PaymentDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s *payment) FindByRelatedEntity(ctx context.Context, entityID string, entityType string) ([]*PaymentDomain, error) {
	query := &PaymentQuery{
		RelatedEntityID:   gopkg.Pointer(entityID),
		RelatedEntityType: gopkg.Pointer(entityType),
		Query: Query{
			Sorts: "payment_date.desc",
		},
	}
	return s.FindAll(ctx, query)
}

func (s *payment) FindByStatus(ctx context.Context, status string, tenant enum.Tenant) ([]*PaymentDomain, error) {
	query := &PaymentQuery{
		Status:   gopkg.Pointer(status),
		TenantId: gopkg.Pointer(tenant),
		Query: Query{
			Sorts: "payment_date.desc",
		},
	}
	return s.FindAll(ctx, query)
}

func (s *payment) FindByDateRange(ctx context.Context, startDate, endDate time.Time, tenant enum.Tenant) ([]*PaymentDomain, error) {
	query := &PaymentQuery{
		StartDate: gopkg.Pointer(startDate),
		EndDate:   gopkg.Pointer(endDate),
		TenantId:  gopkg.Pointer(tenant),
		Query: Query{
			Sorts: "payment_date.desc",
		},
	}
	return s.FindAll(ctx, query)
}

func (s *payment) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"status":     status,
			"updated_at": time.Now(),
		}},
	)
}

func (s *payment) RefundPayment(ctx context.Context, id string, refundID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{
			"status":     "refunded",
			"refund_id":  refundID,
			"updated_at": time.Now(),
		}},
	)
}

func (s *payment) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *payment) Count(ctx context.Context, q *PaymentQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}
