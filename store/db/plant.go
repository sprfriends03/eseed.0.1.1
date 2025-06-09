package db

import (
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"fmt"
	"time"

	"github.com/nhnghia272/gopkg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlantDomain struct {
	BaseDomain      `bson:",inline"`
	PlantTypeID     *string      `json:"plant_type_id" bson:"plant_type_id" validate:"required,len=24"` // Link to PlantType
	PlantSlotID     *string      `json:"plant_slot_id" bson:"plant_slot_id" validate:"required,len=24"` // Link to PlantSlot
	MemberID        *string      `json:"member_id" bson:"member_id" validate:"required,len=24"`         // Link to Member (owner)
	Status          *string      `json:"status" bson:"status" validate:"required"`                      // seedling, vegetative, flowering, harvested, dead
	PlantedDate     *time.Time   `json:"planted_date" bson:"planted_date" validate:"required"`
	ExpectedHarvest *time.Time   `json:"expected_harvest" bson:"expected_harvest" validate:"required,gtfield=PlantedDate"`
	ActualHarvest   *time.Time   `json:"actual_harvest" bson:"actual_harvest" validate:"omitempty,gtfield=PlantedDate"`
	Name            *string      `json:"name" bson:"name" validate:"required"`                  // Plant nickname
	Health          *int         `json:"health" bson:"health" validate:"required,gte=1,lte=10"` // Health rating (1-10)
	Height          *float64     `json:"height" bson:"height" validate:"omitempty,gte=0"`       // in cm
	Images          *[]string    `json:"images" bson:"images" validate:"omitempty,dive,required"`
	Notes           *string      `json:"notes" bson:"notes" validate:"omitempty"`
	Strain          *string      `json:"strain" bson:"strain" validate:"required"` // Denormalized from PlantType
	HarvestID       *string      `json:"harvest_id" bson:"harvest_id" validate:"omitempty,len=24"`
	NFTTokenID      *string      `json:"nft_token_id" bson:"nft_token_id" validate:"omitempty"`
	TenantId        *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *PlantDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

// DTO structures following PlantSlotDomain pattern
type PlantBaseDto struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	Strain          string     `json:"strain"`
	Health          int        `json:"health"`
	PlantedDate     *time.Time `json:"planted_date"`
	ExpectedHarvest *time.Time `json:"expected_harvest"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type PlantDetailDto struct {
	PlantBaseDto  `json:",inline"`
	PlantTypeID   string     `json:"plant_type_id"`
	PlantSlotID   string     `json:"plant_slot_id"`
	MemberID      string     `json:"member_id"`
	ActualHarvest *time.Time `json:"actual_harvest,omitempty"`
	Height        *float64   `json:"height,omitempty"`
	Images        []string   `json:"images,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	HarvestID     string     `json:"harvest_id,omitempty"`
	NFTTokenID    string     `json:"nft_token_id,omitempty"`
	CreatedAt     *time.Time `json:"created_at"`
}

type PlantCareDto struct {
	PlantBaseDto `json:",inline"`
	LastCareDate *time.Time `json:"last_care_date"`
	CareActions  int        `json:"care_actions"`
	HealthTrend  string     `json:"health_trend"` // improving, stable, declining
}

// DTO methods following PlantSlotDomain pattern
func (s PlantDomain) BaseDto() *PlantBaseDto {
	return &PlantBaseDto{
		ID:              SID(s.ID),
		Name:            gopkg.Value(s.Name),
		Status:          gopkg.Value(s.Status),
		Strain:          gopkg.Value(s.Strain),
		Health:          gopkg.Value(s.Health),
		PlantedDate:     s.PlantedDate,
		ExpectedHarvest: s.ExpectedHarvest,
		UpdatedAt:       s.UpdatedAt,
	}
}

func (s PlantDomain) DetailDto() *PlantDetailDto {
	return &PlantDetailDto{
		PlantBaseDto:  *s.BaseDto(),
		PlantTypeID:   gopkg.Value(s.PlantTypeID),
		PlantSlotID:   gopkg.Value(s.PlantSlotID),
		MemberID:      gopkg.Value(s.MemberID),
		ActualHarvest: s.ActualHarvest,
		Height:        s.Height,
		Images:        gopkg.Value(s.Images),
		Notes:         gopkg.Value(s.Notes),
		HarvestID:     gopkg.Value(s.HarvestID),
		NFTTokenID:    gopkg.Value(s.NFTTokenID),
		CreatedAt:     s.CreatedAt,
	}
}

func (s PlantDomain) CareDto() *PlantCareDto {
	// Note: LastCareDate and CareActions would be populated from care records
	// This is a placeholder implementation - in practice, these would be calculated
	// from care record aggregations in the business logic layer
	return &PlantCareDto{
		PlantBaseDto: *s.BaseDto(),
		LastCareDate: nil,      // To be populated by business logic
		CareActions:  0,        // To be populated by business logic
		HealthTrend:  "stable", // To be calculated by business logic
	}
}

// PlantQuery following PlantSlotQuery pattern
type PlantQuery struct {
	Query           `bson:",inline"`
	Status          *string      `json:"status" form:"status"`
	MemberID        *string      `json:"member_id" form:"member_id"`
	PlantSlotID     *string      `json:"plant_slot_id" form:"plant_slot_id"`
	Strain          *string      `json:"strain" form:"strain"`
	HealthMin       *int         `json:"health_min" form:"health_min"`
	HealthMax       *int         `json:"health_max" form:"health_max"`
	ReadyForHarvest *bool        `json:"ready_for_harvest" form:"ready_for_harvest"`
	TenantId        *enum.Tenant `json:"tenant_id" form:"tenant_id"`
}

func (s *PlantQuery) Build() *PlantQuery {
	query := Query{
		Page:   s.Page,
		Limit:  s.Limit,
		Sorts:  s.Sorts,
		Filter: M{},
	}

	if s.Status != nil {
		query.Filter["status"] = *s.Status
	}
	if s.MemberID != nil {
		query.Filter["member_id"] = *s.MemberID
	}
	if s.PlantSlotID != nil {
		query.Filter["plant_slot_id"] = *s.PlantSlotID
	}
	if s.Strain != nil {
		query.Filter["strain"] = primitive.Regex{Pattern: *s.Strain, Options: "i"}
	}
	if s.HealthMin != nil {
		if existing, ok := query.Filter["health"].(M); ok {
			existing["$gte"] = *s.HealthMin
		} else {
			query.Filter["health"] = M{"$gte": *s.HealthMin}
		}
	}
	if s.HealthMax != nil {
		if existing, ok := query.Filter["health"].(M); ok {
			existing["$lte"] = *s.HealthMax
		} else {
			query.Filter["health"] = M{"$lte": *s.HealthMax}
		}
	}
	if s.ReadyForHarvest != nil && *s.ReadyForHarvest {
		query.Filter["status"] = "flowering"
		query.Filter["expected_harvest"] = M{"$lte": time.Now()}
	}
	if s.TenantId != nil {
		query.Filter["tenant_id"] = *s.TenantId
	}

	s.Query = query
	return s
}

// Additional analytics and reporting structures
type PlantAnalyticsQuery struct {
	TenantId  *enum.Tenant `json:"tenant_id"`
	MemberID  *string      `json:"member_id"`
	TimeRange *string      `json:"time_range"` // week, month, quarter, year
	GroupBy   *string      `json:"group_by"`   // strain, status, member
}

type PlantHealthAlert struct {
	PlantID      string    `json:"plant_id"`
	MemberID     string    `json:"member_id"`
	AlertType    string    `json:"alert_type"` // critical_health, overdue_care, harvest_ready
	Severity     string    `json:"severity"`   // low, medium, high, critical
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
	DaysOverdue  *int      `json:"days_overdue,omitempty"`
	HealthRating *int      `json:"health_rating,omitempty"`
}

type PlantGrowthAnalytics struct {
	PlantID          string            `json:"plant_id"`
	GrowthStages     []GrowthStageData `json:"growth_stages"`
	HealthHistory    []HealthDataPoint `json:"health_history"`
	CareFrequency    map[string]int    `json:"care_frequency"`
	PredictedHarvest *time.Time        `json:"predicted_harvest"`
	YieldEstimate    *float64          `json:"yield_estimate"`
}

type GrowthStageData struct {
	Stage     string     `json:"stage"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Duration  *int       `json:"duration_days,omitempty"`
	AvgHealth *float64   `json:"avg_health"`
}

type HealthDataPoint struct {
	Date         time.Time `json:"date"`
	HealthRating int       `json:"health_rating"`
	CareActions  []string  `json:"care_actions"`
}

type plant struct {
	repo *repo
}

func newPlant(ctx context.Context, collection *mongo.Collection) *plant {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "plant_type_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "plant_slot_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "strain", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "planted_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "expected_harvest", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "harvest_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "nft_token_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create plant indexes:", err)
	}

	return &plant{repo: newrepo(collection)}
}

func (s *plant) Save(ctx context.Context, domain *PlantDomain, opts ...*options.UpdateOptions) (*PlantDomain, error) {
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

func (s *plant) Create(ctx context.Context, domain *PlantDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *plant) Update(ctx context.Context, id string, domain *PlantDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *plant) FindByID(ctx context.Context, id string) (*PlantDomain, error) {
	var domain PlantDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plant) FindByMemberID(ctx context.Context, memberID string) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	opts := options.Find().SetSort(bson.D{{Key: "planted_date", Value: -1}})
	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *plant) FindByPlantSlotID(ctx context.Context, plantSlotID string) (*PlantDomain, error) {
	var domain PlantDomain

	// Find active plant in this slot
	query := M{
		"plant_slot_id": plantSlotID,
		"status": M{
			"$nin": []string{"harvested", "dead"},
		},
	}

	err := s.repo.FindOne(ctx, query, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plant) FindActiveByMemberID(ctx context.Context, memberID string) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	opts := options.Find().SetSort(bson.D{{Key: "planted_date", Value: -1}})
	query := Query{
		Filter: M{
			"member_id": memberID,
			"status": M{
				"$nin": []string{"harvested", "dead"},
			},
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains, opts)
}

func (s *plant) FindReadyForHarvest(ctx context.Context, tenantID enum.Tenant) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	now := time.Now()
	query := Query{
		Filter: M{
			"tenant_id": tenantID,
			"status":    "flowering",
			"expected_harvest": M{
				"$lte": now,
			},
			"harvest_id": M{
				"$exists": false,
			},
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plant) UpdateStatus(ctx context.Context, id string, status string) error {
	updates := M{"status": status, "updated_at": time.Now()}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}

func (s *plant) UpdateHealth(ctx context.Context, id string, health int) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"health": health, "updated_at": time.Now()}},
	)
}

func (s *plant) UpdateHeight(ctx context.Context, id string, height float64) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"height": height, "updated_at": time.Now()}},
	)
}

func (s *plant) SetHarvestID(ctx context.Context, id string, harvestID string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$set": M{
				"harvest_id":     harvestID,
				"status":         "harvested",
				"actual_harvest": time.Now(),
				"updated_at":     time.Now(),
			},
		},
	)
}

func (s *plant) AddImage(ctx context.Context, id string, imageURL string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"images": imageURL},
			"$set":  M{"updated_at": time.Now()},
		},
	)
}

func (s *plant) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plant) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}

// FindAll with query support following PlantSlotDomain pattern
func (s *plant) FindAll(ctx context.Context, query *PlantQuery) ([]*PlantDomain, error) {
	var domains []*PlantDomain

	opts := options.Find().SetSort(bson.D{{Key: "planted_date", Value: -1}})

	// Apply pagination
	if query.Page > 0 && query.Limit > 0 {
		skip := (query.Page - 1) * query.Limit
		opts.SetSkip(skip).SetLimit(query.Limit)
	}

	return domains, s.repo.FindAll(ctx, query.Query, &domains, opts)
}

// UpdateFields for partial updates
func (s *plant) UpdateFields(ctx context.Context, id string, updates M) error {
	updates["updated_at"] = time.Now()
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": updates},
	)
}

// Analytics and reporting methods following plant slot pattern
func (s *plant) GetStatusStatistics(ctx context.Context, tenantID enum.Tenant, memberID *string) (map[string]int64, error) {
	pipeline := []bson.M{
		{"$match": M{"tenant_id": tenantID}},
	}

	if memberID != nil {
		pipeline[0]["$match"].(M)["member_id"] = *memberID
	}

	pipeline = append(pipeline,
		bson.M{"$group": bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}},
	)

	cursor, err := s.repo.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int64)
	for cursor.Next(ctx) {
		var doc struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		result[doc.ID] = doc.Count
	}

	return result, nil
}

func (s *plant) GetHealthStatistics(ctx context.Context, tenantID enum.Tenant, memberID *string) (map[string]int64, error) {
	pipeline := []bson.M{
		{"$match": M{"tenant_id": tenantID}},
	}

	if memberID != nil {
		pipeline[0]["$match"].(M)["member_id"] = *memberID
	}

	pipeline = append(pipeline,
		bson.M{"$bucket": bson.M{
			"groupBy":    "$health",
			"boundaries": []int{1, 4, 7, 11}, // Poor (1-3), Fair (4-6), Good (7-10)
			"default":    "unknown",
			"output": bson.M{
				"count": bson.M{"$sum": 1},
			},
		}},
	)

	cursor, err := s.repo.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int64)
	healthRanges := map[int]string{1: "poor", 4: "fair", 7: "good"}

	for cursor.Next(ctx) {
		var doc struct {
			ID    interface{} `bson:"_id"`
			Count int64       `bson:"count"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		if boundary, ok := doc.ID.(int32); ok {
			if label, exists := healthRanges[int(boundary)]; exists {
				result[label] = doc.Count
			}
		}
	}

	return result, nil
}

func (s *plant) GetStrainStatistics(ctx context.Context, tenantID enum.Tenant, memberID *string) ([]map[string]interface{}, error) {
	pipeline := []bson.M{
		{"$match": M{"tenant_id": tenantID}},
	}

	if memberID != nil {
		pipeline[0]["$match"].(M)["member_id"] = *memberID
	}

	pipeline = append(pipeline,
		bson.M{"$group": bson.M{
			"_id":        "$strain",
			"count":      bson.M{"$sum": 1},
			"avg_health": bson.M{"$avg": "$health"},
		}},
		bson.M{"$sort": bson.M{"count": -1}},
		bson.M{"$limit": 10}, // Top 10 strains
	)

	cursor, err := s.repo.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results = append(results, doc)
	}

	return results, nil
}

func (s *plant) GetGrowthCycleMetrics(ctx context.Context, tenantID enum.Tenant, timeRange *string) (map[string]interface{}, error) {
	now := time.Now()
	var startDate time.Time

	switch gopkg.Value(timeRange) {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0) // Default to month
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"tenant_id":    tenantID,
			"planted_date": bson.M{"$gte": startDate},
		}},
		{"$group": bson.M{
			"_id":          nil,
			"total_plants": bson.M{"$sum": 1},
			"avg_health":   bson.M{"$avg": "$health"},
			"avg_days_to_harvest": bson.M{"$avg": bson.M{
				"$divide": []interface{}{
					bson.M{"$subtract": []interface{}{"$expected_harvest", "$planted_date"}},
					86400000, // Convert milliseconds to days
				},
			}},
			"completed_cycles": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$status", "harvested"}},
					1,
					0,
				},
			}},
		}},
	}

	cursor, err := s.repo.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result map[string]interface{}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		delete(result, "_id") // Remove the null _id field
		return result, nil
	}

	return map[string]interface{}{}, nil
}

func (s *plant) GetUpcomingHarvests(ctx context.Context, tenantID enum.Tenant, daysAhead int) ([]map[string]interface{}, error) {
	endDate := time.Now().AddDate(0, 0, daysAhead)

	pipeline := []bson.M{
		{"$match": bson.M{
			"tenant_id": tenantID,
			"status":    bson.M{"$in": []string{"flowering", "vegetative"}},
			"expected_harvest": bson.M{
				"$gte": time.Now(),
				"$lte": endDate,
			},
		}},
		{"$lookup": bson.M{
			"from":         "members",
			"localField":   "member_id",
			"foreignField": "_id",
			"as":           "member",
		}},
		{"$unwind": "$member"},
		{"$project": bson.M{
			"plant_id":         "$_id",
			"name":             1,
			"strain":           1,
			"status":           1,
			"expected_harvest": 1,
			"member_email":     "$member.email",
			"days_until_harvest": bson.M{
				"$divide": []interface{}{
					bson.M{"$subtract": []interface{}{"$expected_harvest", "$$NOW"}},
					86400000, // Convert to days
				},
			},
		}},
		{"$sort": bson.M{"expected_harvest": 1}},
	}

	cursor, err := s.repo.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results = append(results, doc)
	}

	return results, nil
}

func (s *plant) GetHealthAlerts(ctx context.Context, tenantID enum.Tenant) ([]*PlantHealthAlert, error) {
	now := time.Now()
	var alerts []*PlantHealthAlert

	// Find plants with critical health (health <= 3)
	criticalHealthPlants, err := s.FindAll(ctx, &PlantQuery{
		Query: Query{
			Filter: M{
				"tenant_id": tenantID,
				"status":    M{"$nin": []string{"harvested", "dead"}},
				"health":    M{"$lte": 3},
			},
		},
	})
	if err == nil {
		for _, plant := range criticalHealthPlants {
			alerts = append(alerts, &PlantHealthAlert{
				PlantID:      SID(plant.ID),
				MemberID:     gopkg.Value(plant.MemberID),
				AlertType:    "critical_health",
				Severity:     "high",
				Message:      fmt.Sprintf("Plant '%s' has critical health rating of %d", gopkg.Value(plant.Name), gopkg.Value(plant.Health)),
				CreatedAt:    now,
				HealthRating: plant.Health,
			})
		}
	}

	// Find plants ready for harvest (flowering status + past expected harvest date)
	readyForHarvest, err := s.FindReadyForHarvest(ctx, tenantID)
	if err == nil {
		for _, plant := range readyForHarvest {
			daysOverdue := int(now.Sub(*plant.ExpectedHarvest).Hours() / 24)
			severity := "medium"
			if daysOverdue > 7 {
				severity = "high"
			}

			alerts = append(alerts, &PlantHealthAlert{
				PlantID:     SID(plant.ID),
				MemberID:    gopkg.Value(plant.MemberID),
				AlertType:   "harvest_ready",
				Severity:    severity,
				Message:     fmt.Sprintf("Plant '%s' is ready for harvest (%d days overdue)", gopkg.Value(plant.Name), daysOverdue),
				CreatedAt:   now,
				DaysOverdue: &daysOverdue,
			})
		}
	}

	return alerts, nil
}
