# Task 1.7: Plant Management System - Detailed Implementation Plan

## Overview
Implementation of the comprehensive Plant Management system following Test-Driven Development (TDD) principles and existing architectural patterns. This system builds upon the completed Plant Slot Management (Task 1.6) to provide full plant lifecycle management from seeding through harvest.

## Implementation Strategy

### Phase 1: Test-First Development Setup (Week 7 - Day 1)

#### 1.1 Test Infrastructure Setup
**File**: `route/plant_test.go`
**Dependencies**: Existing test patterns from `route/plant_slot_test.go`
**Reuse**: Follow exact test structure from `route/membership_test.go`

**TDD Approach**:
```go
// Create failing tests first for all endpoints
func TestPlantRoutes(t *testing.T) {
    // Test structure following plant_slot_test.go
    tests := []struct {
        name           string
        method         string
        endpoint       string
        permission     enum.Permission
        requestBody    interface{}
        expectedStatus int
        expectedFields []string
    }{
        {
            name:           "GET /plants/v1/my-plants - success",
            method:         "GET",
            endpoint:       "/plants/v1/my-plants",
            permission:     enum.PermissionPlantView,
            expectedStatus: http.StatusOK,
            expectedFields: []string{"plants", "total"},
        },
        // ... additional test cases for all 12 endpoints
    }
}
```

**Key Test Categories**:
1. **Authentication Tests**: Bearer token validation
2. **Permission Tests**: Role-based access control 
3. **Input Validation Tests**: Request body validation
4. **Business Logic Tests**: Plant lifecycle rules
5. **Error Handling Tests**: All error scenarios
6. **Integration Tests**: Plant-slot relationship

### Phase 2: Permission System Enhancement (Week 7 - Day 1)

#### 2.1 Add Plant Permissions
**File**: `pkg/enum/index.go`
**Dependencies**: None
**Reuse**: Follow exact `PermissionPlantSlot*` pattern

**TDD Implementation**:
```go
// Add to existing permission constants after PlantSlot permissions
PermissionPlantView     Permission = "plant_view"
PermissionPlantCreate   Permission = "plant_create"
PermissionPlantUpdate   Permission = "plant_update"
PermissionPlantDelete   Permission = "plant_delete"
PermissionPlantManage   Permission = "plant_manage"     // Admin-level
PermissionPlantCare     Permission = "plant_care"       // Record care activities
PermissionPlantHarvest  Permission = "plant_harvest"    // Harvest management
```

**Testing**: 
- Unit tests for permission validation
- Integration tests with authentication middleware

### Phase 3: Database Enhancement (Week 7 - Days 2-3)

#### 3.1 Enhance PlantDomain with DTO Methods
**File**: `store/db/plant.go`
**Dependencies**: Existing PlantDomain, BaseDomain
**Reuse**: Follow exact `MembershipDomain` DTO pattern

**TDD Enhancement**:
```go
// Add missing DTO structures following membership pattern
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
    PlantBaseDto     `json:",inline"`
    PlantTypeID      string     `json:"plant_type_id"`
    PlantSlotID      string     `json:"plant_slot_id"`
    MemberID         string     `json:"member_id"`
    ActualHarvest    *time.Time `json:"actual_harvest,omitempty"`
    Height           *float64   `json:"height,omitempty"`
    Images           []string   `json:"images,omitempty"`
    Notes            string     `json:"notes,omitempty"`
    HarvestID        string     `json:"harvest_id,omitempty"`
    NFTTokenID       string     `json:"nft_token_id,omitempty"`
    CreatedAt        *time.Time `json:"created_at"`
}

type PlantCareDto struct {
    PlantBaseDto `json:",inline"`
    LastCareDate *time.Time `json:"last_care_date"`
    CareActions  int        `json:"care_actions"`
    HealthTrend  string     `json:"health_trend"` // improving, stable, declining
}

// Add DTO methods following membership pattern
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
        PlantBaseDto:    *s.BaseDto(),
        PlantTypeID:     gopkg.Value(s.PlantTypeID),
        PlantSlotID:     gopkg.Value(s.PlantSlotID),
        MemberID:        gopkg.Value(s.MemberID),
        ActualHarvest:   s.ActualHarvest,
        Height:          s.Height,
        Images:          gopkg.ValueSlice(s.Images),
        Notes:           gopkg.Value(s.Notes),
        HarvestID:       gopkg.Value(s.HarvestID),
        NFTTokenID:      gopkg.Value(s.NFTTokenID),
        CreatedAt:       s.CreatedAt,
    }
}

func (s PlantDomain) CareDto() *PlantCareDto {
    // Implementation with care record aggregation
}
```

#### 3.2 Add Plant Query Support
**File**: `store/db/plant.go`
**Reuse**: Follow exact `MembershipQuery` pattern

**TDD Implementation**:
```go
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
    if s.Strain != nil {
        query.Filter["strain"] = primitive.Regex{Pattern: *s.Strain, Options: "i"}
    }
    if s.HealthMin != nil {
        query.Filter["health"] = M{"$gte": *s.HealthMin}
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

// Additional aggregation queries for analytics
type PlantAnalyticsQuery struct {
    TenantId      *enum.Tenant `json:"tenant_id"`
    MemberID      *string      `json:"member_id"`
    TimeRange     *string      `json:"time_range"` // week, month, quarter, year
    GroupBy       *string      `json:"group_by"`   // strain, status, member
}

type PlantHealthAlert struct {
    PlantID       string    `json:"plant_id"`
    MemberID      string    `json:"member_id"`
    AlertType     string    `json:"alert_type"`    // critical_health, overdue_care, harvest_ready
    Severity      string    `json:"severity"`      // low, medium, high, critical
    Message       string    `json:"message"`
    CreatedAt     time.Time `json:"created_at"`
    DaysOverdue   *int      `json:"days_overdue,omitempty"`
    HealthRating  *int      `json:"health_rating,omitempty"`
}

type PlantGrowthAnalytics struct {
    PlantID          string              `json:"plant_id"`
    GrowthStages     []GrowthStageData   `json:"growth_stages"`
    HealthHistory    []HealthDataPoint   `json:"health_history"`
    CareFrequency    map[string]int      `json:"care_frequency"`
    PredictedHarvest *time.Time          `json:"predicted_harvest"`
    YieldEstimate    *float64            `json:"yield_estimate"`
}

type GrowthStageData struct {
    Stage       string     `json:"stage"`
    StartDate   *time.Time `json:"start_date"`
    EndDate     *time.Time `json:"end_date,omitempty"`
    Duration    *int       `json:"duration_days,omitempty"`
    AvgHealth   *float64   `json:"avg_health"`
}

type HealthDataPoint struct {
    Date         time.Time `json:"date"`
    HealthRating int       `json:"health_rating"`
    CareActions  []string  `json:"care_actions"`
}
```
```

#### 3.3 Enhanced Business Logic Methods
**File**: `store/db/plant.go`
**Dependencies**: PlantSlot system, Member system
**Reuse**: Follow plant slot allocation patterns

**TDD Key Methods**:
```go
// Plant lifecycle management
func (s *plant) CreatePlantFromSlot(ctx context.Context, plantSlotID string, plantTypeID string, memberID string) (*PlantDomain, error)
func (s *plant) UpdateStatus(ctx context.Context, plantID string, newStatus enum.PlantStatus, userID string) error
func (s *plant) UpdateHealth(ctx context.Context, plantID string, healthRating int, userID string) error

// Care management
func (s *plant) AddCareRecord(ctx context.Context, plantID string, careType string, memberID string, data map[string]interface{}) error
func (s *plant) GetCareHistory(ctx context.Context, plantID string) ([]*CareRecordDomain, error)

// Harvest management  
func (s *plant) MarkReadyForHarvest(ctx context.Context, plantID string) error
func (s *plant) ProcessHarvest(ctx context.Context, plantID string, harvestData *HarvestDomain) error

// Analytics and reporting
func (s *plant) GetHealthAnalytics(ctx context.Context, memberID string) (*PlantHealthAnalytics, error)
func (s *plant) GetGrowthAnalytics(ctx context.Context, plantID string) (*PlantGrowthAnalytics, error)
```

**Testing Focus**:
- Unit tests for each business method
- Integration tests with plant slot system
- Edge case handling (invalid states, permissions)

### Phase 4: Error Code Management (Week 7 - Day 3)

#### 4.1 Add Plant-Specific Error Codes
**File**: `pkg/ecode/cannabis.go`
**Dependencies**: Existing error patterns
**Reuse**: Follow exact plant slot error pattern

**TDD Implementation**:
```go
// Add after existing plant slot errors
PlantNotFound              = New(http.StatusNotFound, "plant_not_found")
PlantSlotRequired          = New(http.StatusBadRequest, "plant_slot_required")
PlantInvalidStatus         = New(http.StatusBadRequest, "plant_invalid_status")
PlantUnauthorizedOwner     = New(http.StatusForbidden, "plant_unauthorized_owner")
PlantAlreadyHarvested      = New(http.StatusConflict, "plant_already_harvested")
PlantNotReadyForHarvest    = New(http.StatusConflict, "plant_not_ready_for_harvest")
PlantHealthCritical        = New(http.StatusConflict, "plant_health_critical")
PlantCareRecordInvalid     = New(http.StatusBadRequest, "plant_care_record_invalid")
PlantTypeNotAvailable      = New(http.StatusConflict, "plant_type_not_available")
PlantSlotOccupied          = New(http.StatusConflict, "plant_slot_occupied")
PlantImageUploadFailed     = New(http.StatusBadRequest, "plant_image_upload_failed")
PlantLifecycleViolation    = New(http.StatusConflict, "plant_lifecycle_violation")
```

**Testing**: Error handling integration tests

### Phase 5: API Route Implementation (Week 7 - Days 4-5)

#### 5.1 Plant Management Routes
**File**: `route/plant.go`
**Dependencies**: Authentication, permissions, database
**Reuse**: Follow exact `route/plant_slot.go` pattern

**TDD Route Structure**: `/plants/v1/*`

```go
type plant struct {
    *middleware
}

func init() {
    handlers = append(handlers, func(m *middleware, r *gin.Engine) {
        s := plant{m}

        v1 := r.Group("/plants/v1")
        {
            // Member endpoints (7 endpoints)
            v1.GET("/my-plants", s.BearerAuth(enum.PermissionPlantView), s.v1_getMyPlants())
            v1.POST("/create", s.BearerAuth(enum.PermissionPlantCreate), s.v1_createPlant())
            v1.GET("/:id", s.BearerAuth(enum.PermissionPlantView), s.v1_getPlantDetails())
            v1.PUT("/:id/status", s.BearerAuth(enum.PermissionPlantUpdate), s.v1_updatePlantStatus())
            v1.PUT("/:id/care", s.BearerAuth(enum.PermissionPlantCare), s.v1_recordCare())
            v1.POST("/:id/images", s.BearerAuth(enum.PermissionPlantUpdate), s.v1_uploadPlantImage())
            v1.POST("/:id/harvest", s.BearerAuth(enum.PermissionPlantHarvest), s.v1_harvestPlant())

            // Admin endpoints (5 endpoints)
            admin := v1.Group("/admin")
            {
                admin.GET("/all", s.BearerAuth(enum.PermissionPlantManage), s.v1_getAllPlants())
                admin.GET("/analytics", s.BearerAuth(enum.PermissionPlantManage), s.v1_getPlantAnalytics())
                admin.GET("/health-alerts", s.BearerAuth(enum.PermissionPlantManage), s.v1_getHealthAlerts())
                admin.PUT("/:id/force-status", s.BearerAuth(enum.PermissionPlantManage), s.v1_forceStatusUpdate())
                admin.GET("/harvest-ready", s.BearerAuth(enum.PermissionPlantManage), s.v1_getHarvestReady())
            }
        }
    })
}
```

#### 5.2 Request/Response Structures
**Reuse**: Follow exact membership/plant slot request patterns

**TDD Structures**:
```go
type CreatePlantRequest struct {
    PlantSlotID  string `json:"plant_slot_id" binding:"required,len=24"`
    PlantTypeID  string `json:"plant_type_id" binding:"required,len=24"`
    Name         string `json:"name" binding:"required,min=2,max=50"`
    Notes        string `json:"notes" validate:"omitempty,max=500"`
}

type UpdatePlantStatusRequest struct {
    Status string  `json:"status" binding:"required" validate:"oneof=seedling vegetative flowering harvested dead"`
    Reason *string `json:"reason" validate:"omitempty,max=255"`
}

type RecordCareRequest struct {
    CareType     string             `json:"care_type" binding:"required" validate:"oneof=watering fertilizing pruning inspection pest_control"`
    Notes        string             `json:"notes" validate:"omitempty,max=500"`
    Measurements *CareMeasurements  `json:"measurements" validate:"omitempty"`
    Products     []string           `json:"products" validate:"omitempty,dive,required"`
}

type CareMeasurements struct {
    Temperature *float64 `json:"temperature" validate:"omitempty,gte=-10,lte=50"`
    Humidity    *float64 `json:"humidity" validate:"omitempty,gte=0,lte=100"`
    SoilPH      *float64 `json:"soil_ph" validate:"omitempty,gte=0,lte=14"`
    WaterAmount *float64 `json:"water_amount" validate:"omitempty,gte=0,lte=10000"`
}

type HarvestPlantRequest struct {
    Weight         float64  `json:"weight" binding:"required,gt=0"`
    Quality        int      `json:"quality" binding:"required,gte=1,lte=10"`
    Notes          string   `json:"notes" validate:"omitempty,max=500"`
    ProcessingType string   `json:"processing_type" binding:"required" validate:"oneof=self_process sell_to_seedeg"`
}
```

#### 5.3 TDD Endpoint Implementation

**Member Endpoints**:

1. **GET /plants/v1/my-plants**
   - **Test First**: Write failing test for member plant listing
   - **Implementation**: Follow exact `plant_slot.go` my-slots pattern
   - **Tests**: Authentication, pagination, filtering

2. **POST /plants/v1/create**
   - **Test First**: Write failing test for plant creation
   - **Business Logic**: Validate plant slot availability, member ownership
   - **Tests**: Slot validation, duplicate prevention, permission checks

3. **GET /plants/v1/:id**
   - **Test First**: Write failing test for plant detail retrieval
   - **Implementation**: Ownership verification, detailed DTO response
   - **Tests**: Authorization, data completeness

4. **PUT /plants/v1/:id/status**
   - **Test First**: Write failing test for status updates
   - **Business Logic**: Status transition validation, lifecycle rules
   - **Tests**: Invalid transitions, permission checks

5. **PUT /plants/v1/:id/care**
   - **Test First**: Write failing test for care record creation
   - **Implementation**: Care record validation, measurement tracking
   - **Tests**: Data validation, care type restrictions

6. **POST /plants/v1/:id/images**
   - **Test First**: Write failing test for image upload
   - **Implementation**: MinIO storage integration, image validation
   - **Tests**: File type validation, size limits

7. **POST /plants/v1/:id/harvest**
   - **Test First**: Write failing test for harvest processing
   - **Business Logic**: Readiness validation, harvest record creation
   - **Tests**: Status requirements, data validation

**Admin Endpoints**:

8. **GET /plants/v1/admin/all**
   - **Test First**: Write failing test for admin plant listing
   - **Implementation**: Advanced filtering, multi-tenant support
   - **Tests**: Permission checks, filtering accuracy

9. **GET /plants/v1/admin/analytics**
   - **Test First**: Write failing test for analytics retrieval
   - **Implementation**: Health statistics, growth metrics
   - **Tests**: Data accuracy, performance

10. **GET /plants/v1/admin/health-alerts**
    - **Test First**: Write failing test for health alert system
    - **Implementation**: Critical health detection, alert generation
    - **Tests**: Threshold accuracy, alert categorization

11. **PUT /plants/v1/admin/:id/force-status**
    - **Test First**: Write failing test for admin override
    - **Implementation**: Admin-only status forcing, audit logging
    - **Tests**: Admin permission enforcement

12. **GET /plants/v1/admin/harvest-ready**
     - **Test First**: Write failing test for harvest readiness
     - **Implementation**: Harvest schedule calculation
     - **Tests**: Date calculations, status filtering

#### 5.4 Detailed Endpoint Implementations

**TDD Implementation Pattern for Each Endpoint**:

1. **Write Failing Test**: Create test that expects functionality
2. **Implement Minimal Code**: Make test pass with simplest implementation
3. **Refactor**: Optimize and enhance while keeping tests green
4. **Add Edge Cases**: Test error conditions and boundaries
5. **Integration Testing**: Test with other systems

##### v1_getMyPlants Implementation

```go
func (s plant) v1_getMyPlants() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        session := s.Session(c)
        
        // Parse query parameters
        var query db.PlantQuery
        if err := c.ShouldBindQuery(&query); err != nil {
            c.Error(ecode.BadRequest.Desc(err))
            return
        }
        
        // Get member info
        member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
        if err != nil {
            c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
            return
        }
        
        // Set member filter
        query.MemberID = gopkg.Pointer(db.SID(member.ID))
        query.TenantId = gopkg.Pointer(session.TenantId)
        
        // Get plants with pagination
        plants, err := s.store.Db.Plant.FindAll(ctx, query.Build())
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        
        // Convert to DTOs
        plantDtos := make([]*db.PlantBaseDto, len(plants))
        for i, plant := range plants {
            plantDtos[i] = plant.BaseDto()
        }
        
        // Get total count for pagination
        totalCount, err := s.store.Db.Plant.Count(ctx, query.Filter)
        if err != nil {
            logrus.Warnf("Failed to get plant count: %v", err)
        }
        
        c.JSON(http.StatusOK, gin.H{
            "plants": plantDtos,
            "total":  len(plantDtos),
            "count":  totalCount,
            "page":   gopkg.Value(query.Page, 1),
            "limit":  gopkg.Value(query.Limit, 20),
        })
    }
}
```

##### v1_createPlant Implementation

```go
func (s plant) v1_createPlant() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        session := s.Session(c)
        
        var req CreatePlantRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.Error(ecode.BadRequest.Desc(err))
            return
        }
        
        // Get member info
        member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
        if err != nil {
            c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
            return
        }
        
        // Validate plant slot ownership and availability
        plantSlot, err := s.store.Db.PlantSlot.FindByID(ctx, req.PlantSlotID)
        if err != nil {
            c.Error(ecode.PlantSlotNotFound.Desc(err))
            return
        }
        
        // Verify slot ownership
        if plantSlot.MemberID == nil || *plantSlot.MemberID != db.SID(member.ID) {
            c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Plant slot not owned by member")))
            return
        }
        
        // Verify slot is available
        if plantSlot.Status == nil || *plantSlot.Status != "allocated" {
            c.Error(ecode.PlantSlotOccupied.Desc(fmt.Errorf("Plant slot is not available for planting")))
            return
        }
        
        // Check if slot already has a plant
        existingPlant, err := s.store.Db.Plant.FindByPlantSlotID(ctx, req.PlantSlotID)
        if err == nil && existingPlant != nil {
            c.Error(ecode.PlantSlotOccupied.Desc(fmt.Errorf("Plant slot already has an active plant")))
            return
        }
        
        // Validate plant type availability
        plantType, err := s.store.Db.PlantType.FindByID(ctx, req.PlantTypeID)
        if err != nil {
            c.Error(ecode.PlantTypeNotAvailable.Desc(err))
            return
        }
        
        // Create plant domain
        now := time.Now()
        expectedHarvest := now.AddDate(0, 0, 90) // Default 90 days growth cycle
        if plantType.FloweringTime != nil {
            expectedHarvest = now.AddDate(0, 0, *plantType.FloweringTime)
        }
        
        plant := &db.PlantDomain{
            PlantTypeID:     &req.PlantTypeID,
            PlantSlotID:     &req.PlantSlotID,
            MemberID:        gopkg.Pointer(db.SID(member.ID)),
            Status:          gopkg.Pointer("seedling"),
            PlantedDate:     &now,
            ExpectedHarvest: &expectedHarvest,
            Name:            &req.Name,
            Health:          gopkg.Pointer(8), // Default healthy rating
            Strain:          plantType.Strain,
            Notes:           &req.Notes,
            TenantId:        &session.TenantId,
        }
        
        // Save plant
        savedPlant, err := s.store.Db.Plant.Save(ctx, plant)
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        
        // Update plant slot status to occupied
        err = s.store.Db.PlantSlot.UpdateStatus(ctx, req.PlantSlotID, "occupied")
        if err != nil {
            // Log error but don't fail the request
            logrus.Errorf("Failed to update plant slot status after plant creation: %v", err)
        }
        
        // Audit log
        s.AuditLog(c, "plant", enum.DataActionCreate, savedPlant, savedPlant, db.SID(savedPlant.ID))
        
        c.JSON(http.StatusCreated, gin.H{
            "message": "Plant created successfully",
            "plant":   savedPlant.DetailDto(),
        })
    }
}
```

##### v1_recordCare Implementation

```go
func (s plant) v1_recordCare() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        session := s.Session(c)
        plantID := c.Param("id")
        
        var req RecordCareRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.Error(ecode.BadRequest.Desc(err))
            return
        }
        
        // Get member info
        member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
        if err != nil {
            c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
            return
        }
        
        // Get plant and verify ownership
        plant, err := s.store.Db.Plant.FindByID(ctx, plantID)
        if err != nil {
            c.Error(ecode.PlantNotFound.Desc(err))
            return
        }
        
        if plant.MemberID == nil || *plant.MemberID != db.SID(member.ID) {
            c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Not authorized to record care for this plant")))
            return
        }
        
        // Validate plant is not harvested or dead
        if plant.Status != nil && (*plant.Status == "harvested" || *plant.Status == "dead") {
            c.Error(ecode.PlantLifecycleViolation.Desc(fmt.Errorf("Cannot record care for harvested or dead plants")))
            return
        }
        
        // Create care record
        now := time.Now()
        careRecord := &db.CareRecordDomain{
            PlantID:  &plantID,
            MemberID: gopkg.Pointer(db.SID(member.ID)),
            CareType: &req.CareType,
            CareDate: &now,
            Notes:    &req.Notes,
            TenantId: &session.TenantId,
        }
        
        // Add measurements if provided
        if req.Measurements != nil {
            careRecord.Measurements = &struct {
                Temperature *float64 `json:"temperature" bson:"temperature" validate:"omitempty"`
                Humidity    *float64 `json:"humidity" bson:"humidity" validate:"omitempty,gte=0,lte=100"`
                SoilPH      *float64 `json:"soil_ph" bson:"soil_ph" validate:"omitempty,gte=0,lte=14"`
                WaterAmount *float64 `json:"water_amount" bson:"water_amount" validate:"omitempty,gte=0"`
            }{
                Temperature: req.Measurements.Temperature,
                Humidity:    req.Measurements.Humidity,
                SoilPH:      req.Measurements.SoilPH,
                WaterAmount: req.Measurements.WaterAmount,
            }
        }
        
        // Add products if provided
        if len(req.Products) > 0 {
            careRecord.Products = &req.Products
        }
        
        // Save care record
        savedCareRecord, err := s.store.Db.CareRecord.Save(ctx, careRecord)
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        
        // Update plant's last care information and potentially health
        updateData := map[string]interface{}{
            "updated_at": now,
        }
        
        // Health improvement logic based on care type
        if plant.Health != nil {
            currentHealth := *plant.Health
            switch req.CareType {
            case "watering":
                if currentHealth < 10 {
                    updateData["health"] = currentHealth + 1
                }
            case "fertilizing":
                if currentHealth < 9 {
                    updateData["health"] = currentHealth + 2
                }
            case "pest_control":
                if currentHealth < 8 {
                    updateData["health"] = currentHealth + 3
                }
            }
        }
        
        // Update plant
        err = s.store.Db.Plant.UpdateFields(ctx, plantID, updateData)
        if err != nil {
            logrus.Errorf("Failed to update plant after care record: %v", err)
        }
        
        // Audit log
        s.AuditLog(c, "care_record", enum.DataActionCreate, savedCareRecord, savedCareRecord, db.SID(savedCareRecord.ID))
        
        c.JSON(http.StatusOK, gin.H{
            "message":     "Care record added successfully",
            "care_record": savedCareRecord.BaseDto(),
        })
    }
}
```

##### v1_getPlantAnalytics Implementation

```go
func (s plant) v1_getPlantAnalytics() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        session := s.Session(c)
        
        // Parse query parameters
        var query PlantAnalyticsQuery
        if err := c.ShouldBindQuery(&query); err != nil {
            c.Error(ecode.BadRequest.Desc(err))
            return
        }
        
        query.TenantId = &session.TenantId
        
        // Get plant statistics
        stats := map[string]interface{}{}
        
        // Total plants by status
        statusStats, err := s.store.Db.Plant.GetStatusStatistics(ctx, *query.TenantId, query.MemberID)
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        stats["status_distribution"] = statusStats
        
        // Health distribution
        healthStats, err := s.store.Db.Plant.GetHealthStatistics(ctx, *query.TenantId, query.MemberID)
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        stats["health_distribution"] = healthStats
        
        // Strain popularity
        strainStats, err := s.store.Db.Plant.GetStrainStatistics(ctx, *query.TenantId, query.MemberID)
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        stats["strain_popularity"] = strainStats
        
        // Growth cycle metrics
        cycleStats, err := s.store.Db.Plant.GetGrowthCycleMetrics(ctx, *query.TenantId, query.TimeRange)
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        stats["growth_metrics"] = cycleStats
        
        // Upcoming harvests
        upcomingHarvests, err := s.store.Db.Plant.GetUpcomingHarvests(ctx, *query.TenantId, 30) // Next 30 days
        if err != nil {
            c.Error(ecode.InternalServerError.Desc(err))
            return
        }
        stats["upcoming_harvests"] = upcomingHarvests
        
        c.JSON(http.StatusOK, gin.H{
            "analytics": stats,
            "generated_at": time.Now(),
        })
    }
}
```

### Phase 6: Business Logic Integration (Week 7 - Day 5)

#### 6.1 Plant-Slot Integration
**Dependencies**: Plant Slot Management System (Task 1.6)
**Implementation**: Seamless integration with slot allocation

**TDD Integration Points**:
1. **Slot Occupancy Management**: When plant is created, slot status → "occupied"
2. **Slot Release Management**: When plant is harvested/dies, slot status → "available"
3. **Transfer Validation**: Cannot transfer occupied plant slots
4. **Maintenance Coordination**: Plant health affects slot maintenance requirements

#### 6.2 Member-Plant Relationship
**Dependencies**: Membership system
**Implementation**: Ownership validation and access control

**TDD Business Rules**:
1. **Ownership Verification**: Only plant owners can modify their plants
2. **Membership Validation**: Active membership required for plant operations
3. **Slot Allocation Limits**: Respect membership tier slot limits
4. **Transfer Restrictions**: Plant ownership follows slot transfers

#### 6.3 Plant Lifecycle Management
**Implementation**: Status transition validation and automation

**TDD Lifecycle Rules**:
```go
// Valid status transitions
var validTransitions = map[string][]string{
    "seedling":   {"vegetative", "dead"},
    "vegetative": {"flowering", "dead"},
    "flowering":  {"harvested", "dead"},
    "harvested":  {"seedling"},  // New cycle
    "dead":       {"seedling"},  // Replacement
}

func ValidateStatusTransition(current, new string) bool {
    allowed, exists := validTransitions[current]
    if !exists {
        return false
    }
    return slices.Contains(allowed, new)
}
```

### Phase 7: Documentation & API Specification (Week 7 - Day 5)

#### 7.1 API Documentation
**File**: `docs/api-plant-management.md`
**Reuse**: Follow exact `api-plant-slot-management.md` structure

**TDD Documentation Sections**:
1. **Overview**: Plant management capabilities
2. **Endpoints Summary**: All 12 endpoints with descriptions
3. **Authentication & Authorization**: Permission matrix
4. **Business Rules**: Plant lifecycle, care requirements
5. **Error Handling**: Complete error code reference
6. **Usage Examples**: cURL examples for all endpoints
7. **Integration Points**: Plant slot, membership, harvest integration

#### 7.2 Swagger Documentation
**File**: `docs/swagger.yaml`
**Implementation**: Complete OpenAPI specification

**TDD Swagger Additions**:
```yaml
# Plant Management Paths
/plants/v1/my-plants:
  get:
    summary: Get member's plants
    tags: [Plants]
    security:
      - BearerAuth: []
    parameters:
      - name: status
        in: query
        schema:
          type: string
          enum: [seedling, vegetative, flowering, harvested, dead]
    responses:
      200:
        description: Plants retrieved successfully
        content:
          application/json:
            schema:
              type: object
              properties:
                plants:
                  type: array
                  items:
                    $ref: '#/components/schemas/PlantBaseDto'
                total:
                  type: integer

# ... Complete definitions for all 12 endpoints
```

### Phase 8: Comprehensive Testing Strategy (Week 7 - Days 1-5)

#### 8.1 Unit Testing Approach
**File**: `route/plant_test.go`
**Coverage Target**: 95%+ test coverage

**TDD Test Categories**:

1. **Authentication Tests**:
```go
func TestPlantRoutes_Authentication(t *testing.T) {
    tests := []struct {
        name       string
        endpoint   string
        method     string
        authHeader string
        expectCode int
    }{
        {
            name:       "Missing auth header",
            endpoint:   "/plants/v1/my-plants",
            method:     "GET",
            authHeader: "",
            expectCode: 401,
        },
        // ... more auth tests
    }
}
```

2. **Permission Tests**:
```go
func TestPlantRoutes_Permissions(t *testing.T) {
    tests := []struct {
        name         string
        endpoint     string
        userRole     string
        expectedCode int
    }{
        {
            name:         "Member can view own plants",
            endpoint:     "/plants/v1/my-plants",
            userRole:     "member",
            expectedCode: 200,
        },
        {
            name:         "Member cannot access admin endpoints",
            endpoint:     "/plants/v1/admin/all",
            userRole:     "member",
            expectedCode: 403,
        },
        // ... more permission tests
    }
}
```

3. **Business Logic Tests**:
```go
func TestPlantRoutes_BusinessLogic(t *testing.T) {
    tests := []struct {
        name           string
        setupFunc      func(*testing.T) TestSetup
        endpoint       string
        method         string
        requestBody    interface{}
        expectedCode   int
        validateFunc   func(*testing.T, *httptest.ResponseRecorder)
    }{
        {
            name: "Cannot create plant in unowned slot",
            setupFunc: func(t *testing.T) TestSetup {
                // Setup test data
            },
            endpoint:     "/plants/v1/create",
            method:       "POST",
            requestBody:  CreatePlantRequest{...},
            expectedCode: 403,
        },
        // ... more business logic tests
    }
}
```

4. **Integration Tests**:
```go
func TestPlantRoutes_Integration(t *testing.T) {
    // Test full workflows:
    // 1. Plant creation → slot occupation
    // 2. Plant harvest → slot release
    // 3. Plant transfer → ownership change
    // 4. Plant health → maintenance triggers
}
```

#### 8.2 Performance Testing
**Implementation**: Load testing for plant endpoints

**TDD Performance Tests**:
```go
func BenchmarkPlantRoutes_GetMyPlants(b *testing.B) {
    // Benchmark plant listing with various data sizes
}

func BenchmarkPlantRoutes_PlantCreation(b *testing.B) {
    // Benchmark plant creation process
}
```

#### 8.3 Edge Case Testing
**Focus**: Error conditions and boundary cases

**TDD Edge Cases**:
1. **Invalid Status Transitions**: Attempt invalid lifecycle changes
2. **Concurrent Operations**: Multiple users accessing same plant
3. **Data Consistency**: Plant-slot relationship integrity
4. **Resource Limits**: Maximum plants per member
5. **Network Failures**: Partial operation recovery

#### 8.4 Comprehensive Test Examples

**TDD Test Implementation**:

```go
// Complete test file structure following plant_slot_test.go pattern
package route

import (
    "app/pkg/ecode"
    "app/pkg/enum"
    "app/store/db"
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPlantRoutes_Authentication(t *testing.T) {
    tests := []struct {
        name       string
        endpoint   string
        method     string
        authHeader string
        expectCode int
    }{
        {
            name:       "Missing auth header - my plants",
            endpoint:   "/plants/v1/my-plants",
            method:     "GET",
            authHeader: "",
            expectCode: 401,
        },
        {
            name:       "Missing auth header - create plant",
            endpoint:   "/plants/v1/create",
            method:     "POST",
            authHeader: "",
            expectCode: 401,
        },
        {
            name:       "Invalid auth header",
            endpoint:   "/plants/v1/my-plants",
            method:     "GET",
            authHeader: "Bearer invalid_token",
            expectCode: 401,
        },
        {
            name:       "Valid auth header",
            endpoint:   "/plants/v1/my-plants",
            method:     "GET",
            authHeader: "Bearer valid_token",
            expectCode: 200,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
            router := setupTestRouter()
            req := httptest.NewRequest(tt.method, tt.endpoint, nil)
            if tt.authHeader != "" {
                req.Header.Set("Authorization", tt.authHeader)
            }
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectCode, w.Code)
        })
    }
}

func TestPlantRoutes_Permissions(t *testing.T) {
    tests := []struct {
        name         string
        endpoint     string
        method       string
        userRole     string
        expectedCode int
    }{
        {
            name:         "Member can view own plants",
            endpoint:     "/plants/v1/my-plants",
            method:       "GET",
            userRole:     "member",
            expectedCode: 200,
        },
        {
            name:         "Member cannot access admin endpoints",
            endpoint:     "/plants/v1/admin/all",
            method:       "GET",
            userRole:     "member",
            expectedCode: 403,
        },
        {
            name:         "Admin can access all endpoints",
            endpoint:     "/plants/v1/admin/all",
            method:       "GET",
            userRole:     "admin",
            expectedCode: 200,
        },
        {
            name:         "Member cannot force status update",
            endpoint:     "/plants/v1/admin/plant123/force-status",
            method:       "PUT",
            userRole:     "member",
            expectedCode: 403,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation with role-based authentication
            router := setupTestRouterWithAuth(tt.userRole)
            req := httptest.NewRequest(tt.method, tt.endpoint, nil)
            req.Header.Set("Authorization", "Bearer "+generateTestToken(tt.userRole))
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedCode, w.Code)
        })
    }
}

func TestPlantRoutes_BusinessLogic(t *testing.T) {
    tests := []struct {
        name           string
        setupFunc      func(*testing.T) TestSetup
        endpoint       string
        method         string
        requestBody    interface{}
        expectedCode   int
        validateFunc   func(*testing.T, *httptest.ResponseRecorder)
    }{
        {
            name: "Cannot create plant in unowned slot",
            setupFunc: func(t *testing.T) TestSetup {
                return TestSetup{
                    Member:    createTestMember(t, "member1"),
                    PlantSlot: createTestPlantSlot(t, "member2"), // Owned by different member
                    PlantType: createTestPlantType(t),
                }
            },
            endpoint: "/plants/v1/create",
            method:   "POST",
            requestBody: CreatePlantRequest{
                PlantSlotID: "slot123",
                PlantTypeID: "type123",
                Name:        "Test Plant",
            },
            expectedCode: 403,
            validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &response)
                require.NoError(t, err)
                assert.Contains(t, response["error"], "unauthorized")
            },
        },
        {
            name: "Cannot create plant in occupied slot",
            setupFunc: func(t *testing.T) TestSetup {
                return TestSetup{
                    Member:       createTestMember(t, "member1"),
                    PlantSlot:    createTestPlantSlot(t, "member1", "occupied"),
                    PlantType:    createTestPlantType(t),
                    ExistingPlan: createTestPlant(t, "slot123"),
                }
            },
            endpoint: "/plants/v1/create",
            method:   "POST",
            requestBody: CreatePlantRequest{
                PlantSlotID: "slot123",
                PlantTypeID: "type123",
                Name:        "Test Plant",
            },
            expectedCode: 409,
            validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &response)
                require.NoError(t, err)
                assert.Contains(t, response["error"], "occupied")
            },
        },
        {
            name: "Successful plant creation",
            setupFunc: func(t *testing.T) TestSetup {
                return TestSetup{
                    Member:    createTestMember(t, "member1"),
                    PlantSlot: createTestPlantSlot(t, "member1", "allocated"),
                    PlantType: createTestPlantType(t),
                }
            },
            endpoint: "/plants/v1/create",
            method:   "POST",
            requestBody: CreatePlantRequest{
                PlantSlotID: "slot123",
                PlantTypeID: "type123",
                Name:        "Test Plant",
                Notes:       "Initial planting notes",
            },
            expectedCode: 201,
            validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &response)
                require.NoError(t, err)
                
                assert.Equal(t, "Plant created successfully", response["message"])
                
                plant := response["plant"].(map[string]interface{})
                assert.Equal(t, "Test Plant", plant["name"])
                assert.Equal(t, "seedling", plant["status"])
                assert.Equal(t, float64(8), plant["health"]) // Default health
                assert.NotEmpty(t, plant["id"])
                assert.NotEmpty(t, plant["planted_date"])
                assert.NotEmpty(t, plant["expected_harvest"])
            },
        },
        {
            name: "Invalid status transition",
            setupFunc: func(t *testing.T) TestSetup {
                return TestSetup{
                    Member: createTestMember(t, "member1"),
                    Plant:  createTestPlant(t, "member1", "seedling"),
                }
            },
            endpoint: "/plants/v1/plant123/status",
            method:   "PUT",
            requestBody: UpdatePlantStatusRequest{
                Status: "harvested", // Invalid: cannot go from seedling to harvested
                Reason: ptr("Testing invalid transition"),
            },
            expectedCode: 400,
            validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &response)
                require.NoError(t, err)
                assert.Contains(t, response["error"], "lifecycle_violation")
            },
        },
        {
            name: "Valid status transition",
            setupFunc: func(t *testing.T) TestSetup {
                return TestSetup{
                    Member: createTestMember(t, "member1"),
                    Plant:  createTestPlant(t, "member1", "seedling"),
                }
            },
            endpoint: "/plants/v1/plant123/status",
            method:   "PUT",
            requestBody: UpdatePlantStatusRequest{
                Status: "vegetative",
                Reason: ptr("Plant has developed first true leaves"),
            },
            expectedCode: 200,
            validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &response)
                require.NoError(t, err)
                assert.Equal(t, "Plant status updated successfully", response["message"])
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test data
            setup := tt.setupFunc(t)
            defer cleanup(t, setup)
            
            // Create request
            var reqBody bytes.Buffer
            if tt.requestBody != nil {
                reqBodyBytes, err := json.Marshal(tt.requestBody)
                require.NoError(t, err)
                reqBody = *bytes.NewBuffer(reqBodyBytes)
            }
            
            router := setupTestRouterWithAuth("member")
            req := httptest.NewRequest(tt.method, tt.endpoint, &reqBody)
            req.Header.Set("Authorization", "Bearer "+generateTestToken("member"))
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedCode, w.Code)
            
            if tt.validateFunc != nil {
                tt.validateFunc(t, w)
            }
        })
    }
}

func TestPlantRoutes_CareRecording(t *testing.T) {
    tests := []struct {
        name         string
        careType     string
        measurements *CareMeasurements
        expectedCode int
        healthChange int
    }{
        {
            name:     "Basic watering",
            careType: "watering",
            measurements: &CareMeasurements{
                WaterAmount: ptr(500.0), // 500ml
            },
            expectedCode: 200,
            healthChange: 1,
        },
        {
            name:     "Fertilizing with soil pH measurement",
            careType: "fertilizing",
            measurements: &CareMeasurements{
                SoilPH:      ptr(6.5),
                Temperature: ptr(24.0),
                Humidity:    ptr(65.0),
            },
            expectedCode: 200,
            healthChange: 2,
        },
        {
            name:     "Pest control treatment",
            careType: "pest_control",
            measurements: &CareMeasurements{
                Temperature: ptr(22.0),
                Humidity:    ptr(70.0),
            },
            expectedCode: 200,
            healthChange: 3,
        },
        {
            name:         "Invalid care type",
            careType:     "invalid_care",
            expectedCode: 400,
            healthChange: 0,
        },
        {
            name:     "Invalid pH measurement",
            careType: "fertilizing",
            measurements: &CareMeasurements{
                SoilPH: ptr(15.0), // Invalid: pH cannot be > 14
            },
            expectedCode: 400,
            healthChange: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test plant with known health
            setup := createTestSetupWithPlant(t, 5) // Health rating of 5
            defer cleanup(t, setup)
            
            reqBody := RecordCareRequest{
                CareType:     tt.careType,
                Notes:        "Test care notes",
                Measurements: tt.measurements,
            }
            
            // Make request
            response := makeTestRequest(t, "PUT", fmt.Sprintf("/plants/v1/%s/care", setup.Plant.ID), reqBody, "member")
            
            assert.Equal(t, tt.expectedCode, response.Code)
            
            if tt.expectedCode == 200 {
                // Verify care record was created
                var responseBody map[string]interface{}
                err := json.Unmarshal(response.Body.Bytes(), &responseBody)
                require.NoError(t, err)
                
                assert.Equal(t, "Care record added successfully", responseBody["message"])
                assert.NotNil(t, responseBody["care_record"])
                
                // Verify plant health updated
                updatedPlant := getPlantFromDB(t, setup.Plant.ID)
                expectedHealth := 5 + tt.healthChange
                if expectedHealth > 10 {
                    expectedHealth = 10 // Health cap
                }
                assert.Equal(t, expectedHealth, *updatedPlant.Health)
            }
        })
    }
}

func TestPlantRoutes_HarvestWorkflow(t *testing.T) {
    t.Run("Complete harvest workflow", func(t *testing.T) {
        // Setup plant ready for harvest
        setup := createTestSetupWithPlant(t, 8)
        
        // Set plant to flowering status and past expected harvest date
        plantID := setup.Plant.ID
        err := updatePlantInDB(t, plantID, map[string]interface{}{
            "status":           "flowering",
            "expected_harvest": time.Now().AddDate(0, 0, -1), // Yesterday
        })
        require.NoError(t, err)
        
        // Attempt harvest
        harvestReq := HarvestPlantRequest{
            Weight:         45.5,
            Quality:        8,
            Notes:          "Good quality harvest",
            ProcessingType: "self_process",
        }
        
        response := makeTestRequest(t, "POST", fmt.Sprintf("/plants/v1/%s/harvest", plantID), harvestReq, "member")
        
        assert.Equal(t, 200, response.Code)
        
        // Verify response
        var responseBody map[string]interface{}
        err = json.Unmarshal(response.Body.Bytes(), &responseBody)
        require.NoError(t, err)
        
        assert.Equal(t, "Plant harvested successfully", responseBody["message"])
        assert.NotNil(t, responseBody["harvest"])
        
        // Verify plant status updated
        updatedPlant := getPlantFromDB(t, plantID)
        assert.Equal(t, "harvested", *updatedPlant.Status)
        assert.NotNil(t, updatedPlant.ActualHarvest)
        assert.NotNil(t, updatedPlant.HarvestID)
        
        // Verify plant slot released
        plantSlot := getPlantSlotFromDB(t, setup.PlantSlot.ID)
        assert.Equal(t, "available", *plantSlot.Status)
        
        // Verify harvest record created
        harvest := getHarvestFromDB(t, *updatedPlant.HarvestID)
        assert.Equal(t, 45.5, *harvest.Weight)
        assert.Equal(t, 8, *harvest.Quality)
        assert.Equal(t, "Good quality harvest", *harvest.Notes)
    })
}

// Helper functions for testing
type TestSetup struct {
    Member      *db.MemberDomain
    PlantSlot   *db.PlantSlotDomain
    PlantType   *db.PlantTypeDomain
    Plant       *db.PlantDomain
    Membership  *db.MembershipDomain
}

func createTestMember(t *testing.T, memberID string) *db.MemberDomain {
    // Implementation for creating test member
    return &db.MemberDomain{
        // Test member data
    }
}

func createTestPlantSlot(t *testing.T, ownerID string, status ...string) *db.PlantSlotDomain {
    // Implementation for creating test plant slot
    return &db.PlantSlotDomain{
        // Test plant slot data
    }
}

func createTestPlantType(t *testing.T) *db.PlantTypeDomain {
    // Implementation for creating test plant type
    return &db.PlantTypeDomain{
        // Test plant type data
    }
}

func createTestPlant(t *testing.T, ownerID string, status string) *db.PlantDomain {
    // Implementation for creating test plant
    return &db.PlantDomain{
        // Test plant data
    }
}

func ptr[T any](v T) *T {
    return &v
}
```

### Phase 9: Security & Compliance (Week 7 - Day 5)

#### 9.1 Security Testing
**Implementation**: Comprehensive security validation

**TDD Security Tests**:
1. **Authorization Bypass**: Attempt to access unauthorized resources
2. **Input Injection**: SQL injection, script injection attempts
3. **Data Leakage**: Cross-tenant data access attempts
4. **Rate Limiting**: API abuse prevention testing

#### 9.2 Data Privacy
**Implementation**: GDPR compliance for plant data

**TDD Privacy Controls**:
1. **Data Minimization**: Only collect necessary plant information
2. **Access Logging**: Track all plant data access
3. **Data Retention**: Implement plant data lifecycle policies
4. **Export/Deletion**: Member data portability and deletion

## Success Criteria

### ✅ TDD Success Metrics

1. **Test Coverage**: 95%+ code coverage across all components
2. **Test-First Development**: All code written after failing tests
3. **Regression Prevention**: Comprehensive test suite prevents breaks
4. **Performance Benchmarks**: All endpoints meet response time SLAs

### ✅ Functional Success Criteria

1. **Complete API Implementation**: All 12 endpoints fully functional
2. **Business Rule Enforcement**: Plant lifecycle rules properly implemented
3. **Integration Success**: Seamless plant-slot system integration
4. **Data Integrity**: Consistent plant-slot-member relationships

### ✅ Quality Success Criteria

1. **Code Quality**: 100% adherence to existing patterns
2. **Error Handling**: Comprehensive error coverage
3. **Documentation**: Complete API and business rule documentation
4. **Performance**: Sub-200ms response times for all endpoints

### ✅ Security Success Criteria

1. **Authentication**: 100% endpoint protection
2. **Authorization**: Role-based access properly enforced
3. **Data Protection**: Plant data properly secured
4. **Audit Trail**: Complete activity logging

## Dependencies & Prerequisites

### ✅ Completed Dependencies
1. **Plant Slot Management** (Task 1.6): ✅ Complete
2. **Membership System** (Task 1.5): ✅ Complete 
3. **Authentication System** (Task 1.2): ✅ Complete
4. **eKYC System** (Task 1.4): ✅ Complete

### 📋 Required Resources
1. **PlantDomain Model**: ✅ Exists, needs DTO enhancement
2. **CareRecordDomain Model**: ✅ Exists, ready for integration
3. **HarvestDomain Model**: ✅ Exists, ready for integration
4. **Error Code Framework**: ✅ Exists, needs plant-specific codes

## Risk Mitigation

### 🔒 Technical Risks
1. **Plant-Slot Consistency**: Mitigated by transaction-based operations
2. **Performance Impact**: Mitigated by proper indexing and caching
3. **Data Complexity**: Mitigated by following established patterns

### 🔒 Business Risks
1. **Lifecycle Violations**: Mitigated by comprehensive status validation
2. **Access Control**: Mitigated by robust permission testing
3. **Data Loss**: Mitigated by audit trail and backup procedures

## Timeline & Milestones

### Week 7 Development Schedule

**Day 1 (TDD Setup)**:
- ✅ Morning: Test infrastructure setup (`route/plant_test.go`)
- ✅ Afternoon: Permission system enhancement (`pkg/enum/index.go`)

**Day 2 (Database Enhancement)**:
- ✅ Morning: PlantDomain DTO methods (`store/db/plant.go`)
- ✅ Afternoon: PlantQuery implementation and testing

**Day 3 (Business Logic)**:
- ✅ Morning: Enhanced plant business methods
- ✅ Afternoon: Error code implementation (`pkg/ecode/cannabis.go`)

**Day 4 (API Implementation)**:
- ✅ Morning: Member endpoints (7 endpoints)
- ✅ Afternoon: Admin endpoints (5 endpoints)

**Day 5 (Integration & Documentation)**:
- ✅ Morning: Plant-slot integration testing
- ✅ Afternoon: API documentation and Swagger updates

## Implementation Notes

### 🔧 Technical Specifications

1. **Database Indexes**: Optimize for member queries and status filtering
2. **Caching Strategy**: Cache plant details and care records for performance
3. **File Storage**: Integrate MinIO for plant images with proper access control
4. **Audit Logging**: Track all plant operations for compliance

### 🔧 Business Rule Implementation

1. **Plant Creation**: Require available plant slot and active membership
2. **Status Transitions**: Validate against allowed lifecycle progressions
3. **Care Tracking**: Aggregate care activities for health analytics
4. **Harvest Management**: Coordinate with slot release and member benefits

### 🔧 Integration Points

1. **Plant Slot System**: Bidirectional status synchronization
2. **Membership System**: Access control and benefit calculation
3. **Harvest System**: Lifecycle completion and yield tracking
4. **NFT System**: Token generation for plant ownership (future)

## Conclusion

Task 1.7 implements a comprehensive Plant Management system using strict TDD methodology while maintaining 100% compliance with established architectural patterns. The implementation provides full plant lifecycle management with robust business rule enforcement, comprehensive error handling, and seamless integration with existing systems.

**Next Steps**: Upon completion, the system will be ready for Task 1.8 (Harvest Management System enhancement) and frontend integration for the complete cannabis cultivation club management platform.

**Architecture Compliance**: ✅ 100% adherence to existing patterns, ✅ Zero architectural debt, ✅ Production-ready implementation 