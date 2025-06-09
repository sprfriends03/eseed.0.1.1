package db

import (
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HarvestDomain struct {
	BaseDomain         `bson:",inline"`
	PlantID            *string    `json:"plant_id" bson:"plant_id" validate:"required,len=24"`     // Link to Plant
	MemberID           *string    `json:"member_id" bson:"member_id" validate:"required,len=24"`   // Link to Member
	HarvestDate        time.Time  `json:"harvest_date" bson:"harvest_date" validate:"required"`    // When harvested
	Weight             *float64   `json:"weight" bson:"weight" validate:"required,gt=0"`           // Weight in grams
	Quality            *int       `json:"quality" bson:"quality" validate:"required,gte=1,lte=10"` // Quality rating (1-10)
	Images             *[]string  `json:"images" bson:"images" validate:"omitempty,dive,required"` // Images of the harvest
	Strain             *string    `json:"strain" bson:"strain" validate:"required"`                // Cannabis strain (denormalized from Plant)
	Status             *string    `json:"status" bson:"status" validate:"required"`                // processing, curing, ready, collected
	NFTTokenID         *string    `json:"nft_token_id" bson:"nft_token_id" validate:"omitempty"`
	NFTContractAddress *string    `json:"nft_contract_address" bson:"nft_contract_address" validate:"omitempty"`
	Notes              *string    `json:"notes" bson:"notes" validate:"omitempty"`
	CollectionDate     *time.Time `json:"collection_date" bson:"collection_date" validate:"omitempty"` // When member collected

	// Enhanced Processing Workflow Fields
	ProcessingStage   *string             `json:"processing_stage" bson:"processing_stage"` // harvested, initial_processing, drying, curing, quality_check, ready
	ProcessingStarted *time.Time          `json:"processing_started" bson:"processing_started"`
	DryingCompleted   *time.Time          `json:"drying_completed" bson:"drying_completed"`
	CuringCompleted   *time.Time          `json:"curing_completed" bson:"curing_completed"`
	QualityChecks     *[]QualityCheckData `json:"quality_checks" bson:"quality_checks"`
	ProcessingNotes   *string             `json:"processing_notes" bson:"processing_notes"`
	EstimatedReady    *time.Time          `json:"estimated_ready" bson:"estimated_ready"`

	// Collection Management
	CollectionMethod        *string    `json:"collection_method" bson:"collection_method"` // pickup, scheduled_delivery
	PreferredCollectionDate *time.Time `json:"preferred_collection_date" bson:"preferred_collection_date"`
	DeliveryAddress         *string    `json:"delivery_address" bson:"delivery_address"`
	CollectionScheduled     *time.Time `json:"collection_scheduled" bson:"collection_scheduled"`

	TenantId *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

// Quality Check Data Structure
type QualityCheckData struct {
	CheckedBy     string    `json:"checked_by" bson:"checked_by"`
	CheckDate     time.Time `json:"check_date" bson:"check_date"`
	VisualQuality int       `json:"visual_quality" bson:"visual_quality" validate:"gte=1,lte=10"`
	Moisture      *float64  `json:"moisture" bson:"moisture" validate:"omitempty,gte=0,lte=100"`
	Density       *float64  `json:"density" bson:"density" validate:"omitempty,gte=0"`
	Notes         *string   `json:"notes" bson:"notes"`
	Approved      bool      `json:"approved" bson:"approved"`
}

func (s *HarvestDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

type harvest struct {
	repo *repo
}

func newHarvest(ctx context.Context, collection *mongo.Collection) *harvest {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "plant_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "harvest_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "strain", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "nft_token_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create harvest indexes:", err)
	}

	return &harvest{repo: newrepo(collection)}
}

func (s *harvest) Save(ctx context.Context, domain *HarvestDomain, opts ...*options.UpdateOptions) (*HarvestDomain, error) {
	if err := domain.Validate(); err != nil {
		return nil, err
	}

	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}

	id, err := s.repo.Save(ctx, domain.ID, domain, opts...)
	if err != nil {
		return nil, err
	}
	domain.ID = id

	return s.FindByID(ctx, SID(id))
}

func (s *harvest) Create(ctx context.Context, domain *HarvestDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *harvest) Update(ctx context.Context, id string, domain *HarvestDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *harvest) FindByID(ctx context.Context, id string) (*HarvestDomain, error) {
	var domain HarvestDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *harvest) FindByPlantID(ctx context.Context, plantID string) (*HarvestDomain, error) {
	var domain HarvestDomain
	err := s.repo.FindOne(ctx, M{"plant_id": plantID}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *harvest) FindByMemberID(ctx context.Context, memberID string, offset, limit int64) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	opts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(bson.D{{Key: "harvest_date", Value: -1}})

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *harvest) FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	query := Query{
		Filter: M{
			"status":    status,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *harvest) FindReadyForCollection(ctx context.Context, memberID string) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	query := Query{
		Filter: M{
			"member_id": memberID,
			"status":    "ready",
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *harvest) UpdateStatus(ctx context.Context, id string, status string) error {
	updates := M{"status": status, "updated_at": time.Now()}

	if status == "collected" {
		updates["collection_date"] = time.Now()
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}

func (s *harvest) AddImage(ctx context.Context, id string, imageURL string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"images": imageURL},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *harvest) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *harvest) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}

// Enhanced methods for processing workflow
func (s *harvest) UpdateProcessingStatus(ctx context.Context, id string, stage string, processingNotes *string) error {
	updates := M{
		"processing_stage": stage,
		"updated_at":       time.Now(),
	}

	if processingNotes != nil {
		updates["processing_notes"] = *processingNotes
	}

	// Set stage-specific timestamps
	switch stage {
	case "initial_processing":
		if updates["processing_started"] == nil {
			updates["processing_started"] = time.Now()
		}
	case "drying":
		updates["drying_completed"] = time.Now()
	case "curing":
		updates["curing_completed"] = time.Now()
	case "ready":
		updates["estimated_ready"] = time.Now()
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}

func (s *harvest) RecordQualityCheck(ctx context.Context, id string, qualityData QualityCheckData) error {
	updates := M{
		"$push": M{"quality_checks": qualityData},
		"$set":  M{"updated_at": time.Now()},
	}

	// If approved, advance to ready status
	if qualityData.Approved {
		updates["$set"].(M)["status"] = "ready"
		updates["$set"].(M)["processing_stage"] = "ready"
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		updates,
	)
}

func (s *harvest) FindByStatusAndDateRange(ctx context.Context, status string, startDate, endDate time.Time, tenantID enum.Tenant) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	query := Query{
		Filter: M{
			"status":    status,
			"tenant_id": tenantID,
			"harvest_date": M{
				"$gte": startDate,
				"$lte": endDate,
			},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "harvest_date", Value: -1}})

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *harvest) FindByProcessingStage(ctx context.Context, stage string, tenantID enum.Tenant) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	query := Query{
		Filter: M{
			"processing_stage": stage,
			"tenant_id":        tenantID,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "processing_started", Value: 1}})

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *harvest) GetProcessingMetrics(ctx context.Context, tenantID enum.Tenant, timeRange string) (map[string]interface{}, error) {
	// Build date filter based on time range
	var startDate time.Time
	now := time.Now()

	switch timeRange {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
	default:
		startDate = now.AddDate(-1, 0, 0) // Year default
	}

	// Aggregate pipeline for metrics
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "tenant_id", Value: tenantID},
			{Key: "harvest_date", Value: bson.D{{Key: "$gte", Value: startDate}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "totalWeight", Value: bson.D{{Key: "$sum", Value: "$weight"}}},
			{Key: "avgQuality", Value: bson.D{{Key: "$avg", Value: "$quality"}}},
		}}},
	}

	cursor, err := s.repo.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	results := make(map[string]interface{})
	var statusMetrics []bson.M
	if err := cursor.All(ctx, &statusMetrics); err != nil {
		return nil, err
	}

	results["statusMetrics"] = statusMetrics
	results["timeRange"] = timeRange
	results["generated"] = time.Now()

	return results, nil
}

func (s *harvest) GetCollectionSchedule(ctx context.Context, memberID string) ([]*HarvestDomain, error) {
	var domains []*HarvestDomain

	query := Query{
		Filter: M{
			"member_id":       memberID,
			"status":          "ready",
			"collection_date": M{"$exists": false},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "estimated_ready", Value: 1}})

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *harvest) ScheduleCollection(ctx context.Context, id string, collectionMethod string, preferredDate *time.Time, deliveryAddress *string) error {
	updates := M{
		"collection_method": collectionMethod,
		"updated_at":        time.Now(),
	}

	if preferredDate != nil {
		updates["preferred_collection_date"] = *preferredDate
		updates["collection_scheduled"] = *preferredDate
	}

	if deliveryAddress != nil && collectionMethod == "scheduled_delivery" {
		updates["delivery_address"] = *deliveryAddress
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}

func (s *harvest) CompleteCollection(ctx context.Context, id string) error {
	updates := M{
		"status":          "collected",
		"collection_date": time.Now(),
		"updated_at":      time.Now(),
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}
