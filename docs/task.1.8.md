# Task 1.8: Harvest Management System Enhancement

## Overview

Task 1.8 enhances the existing basic harvest functionality in the Plant Management System by creating a comprehensive, standalone harvest management system with dedicated routes, advanced processing workflows, quality control systems, and member collection interfaces.

## Current Implementation Analysis

### ✅ Existing Code to Reuse

#### 1. Core Models (100% Reuse)
- **HarvestDomain** (`store/db/harvest.go`): Complete model with all required fields
- **PlantDomain** (`store/db/plant.go`): HarvestID field and harvest integration
- **BaseDomain**: Standard domain pattern for consistency
- **Error Codes** (`pkg/ecode/cannabis.go`): Comprehensive harvest error definitions

#### 2. Database Layer (90% Reuse)
- **Harvest Repository** (`store/db/harvest.go`): All CRUD operations implemented
- **Plant Integration**: `SetHarvestID`, `FindReadyForHarvest` methods
- **Database Indexes**: Optimized for harvest queries
- **Validation System**: Complete field validation

#### 3. Basic Routes (50% Enhancement)
- **POST /plants/v1/:id/harvest**: Basic harvest creation (to be enhanced)
- **GET /plants/v1/admin/harvest-ready**: Harvest readiness detection
- **Authentication Middleware**: Bearer auth with permissions
- **Error Handling**: Standard error response patterns

#### 4. Business Logic (80% Reuse)
- **Harvest Validation**: Readiness, weight, quality checks
- **Plant Status Updates**: Automated status transitions
- **Slot Release**: Automatic slot availability on harvest
- **Audit Logging**: Complete audit trail

## Enhancement Plan

### Phase 1: Dedicated Harvest Route System (Week 8 - Days 1-2)

#### 1.1 Create Dedicated Harvest Routes ✅ TDD Implementation

**File**: `route/harvest.go` (NEW)

**Test First Approach**:
```go
// route/harvest_test.go (NEW)
func TestHarvestRoutes_MemberEndpoints(t *testing.T) {
    // Test all member harvest endpoints
}

func TestHarvestRoutes_AdminEndpoints(t *testing.T) {
    // Test all admin harvest endpoints
}

func TestHarvestRoutes_ProcessingWorkflow(t *testing.T) {
    // Test complete processing workflow
}
```

**Implementation**:
```go
// route/harvest.go
type harvest struct {
    *middleware
}

func init() {
    handlers = append(handlers, func(m *middleware, r *gin.Engine) {
        s := harvest{m}

        // Member endpoints
        v1 := r.Group("/harvest/v1")
        v1.GET("/my-harvests", s.BearerAuth(enum.PermissionHarvestView), s.v1_getMyHarvests())
        v1.GET("/:id", s.BearerAuth(enum.PermissionHarvestView), s.v1_getHarvestDetails())
        v1.PUT("/:id/status", s.BearerAuth(enum.PermissionHarvestUpdate), s.v1_updateHarvestStatus())
        v1.POST("/:id/images", s.BearerAuth(enum.PermissionHarvestUpdate), s.v1_uploadHarvestImage())
        v1.POST("/:id/collect", s.BearerAuth(enum.PermissionHarvestCollect), s.v1_collectHarvest())

        // Admin endpoints
        admin := v1.Group("/admin")
        admin.GET("/all", s.BearerAuth(enum.PermissionHarvestManage), s.v1_getAllHarvests())
        admin.GET("/processing", s.BearerAuth(enum.PermissionHarvestManage), s.v1_getProcessingHarvests())
        admin.GET("/analytics", s.BearerAuth(enum.PermissionHarvestManage), s.v1_getHarvestAnalytics())
        admin.POST("/:id/quality-check", s.BearerAuth(enum.PermissionHarvestManage), s.v1_qualityCheck())
        admin.PUT("/:id/force-status", s.BearerAuth(enum.PermissionHarvestManage), s.v1_forceStatusUpdate())
    })
}
```

**Route Specifications**:

1. **GET /harvest/v1/my-harvests** - Member's harvest history with filtering
2. **GET /harvest/v1/:id** - Detailed harvest information
3. **PUT /harvest/v1/:id/status** - Update harvest status (member-initiated)
4. **POST /harvest/v1/:id/images** - Upload harvest documentation images
5. **POST /harvest/v1/:id/collect** - Member harvest collection workflow
6. **GET /harvest/v1/admin/all** - Admin view of all harvests
7. **GET /harvest/v1/admin/processing** - Harvests in processing stages
8. **GET /harvest/v1/admin/analytics** - Harvest analytics and metrics
9. **POST /harvest/v1/admin/:id/quality-check** - Admin quality verification
10. **PUT /harvest/v1/admin/:id/force-status** - Admin force status changes

#### 1.2 Enhanced Permission System ✅ Reuse Existing

**File**: `pkg/enum/index.go` (ENHANCE)

**New Permissions** (following existing pattern):
```go
// Additional harvest permissions
PermissionHarvestView    Permission = "harvest_view"    // View own harvests
PermissionHarvestUpdate  Permission = "harvest_update"  // Update harvest status/images
PermissionHarvestCollect Permission = "harvest_collect" // Collect ready harvests
PermissionHarvestManage  Permission = "harvest_manage"  // Admin harvest management

// Update role definitions
MemberPermissions = append(MemberPermissions, 
    PermissionHarvestView,
    PermissionHarvestUpdate,
    PermissionHarvestCollect,
)

AdminPermissions = append(AdminPermissions,
    PermissionHarvestManage,
)
```

### Phase 2: Advanced Processing Workflows (Week 8 - Days 3-4)

#### 2.1 Enhanced Harvest Domain ✅ Extend Existing

**File**: `store/db/harvest.go` (ENHANCE)

**Test First**:
```go
func TestHarvestDomain_ProcessingWorkflow(t *testing.T) {
    // Test status transitions: processing → curing → ready → collected
}

func TestHarvestDomain_QualityControl(t *testing.T) {
    // Test quality checks and validations
}
```

**New Methods** (following existing patterns):
```go
// Additional domain methods
func (s *harvest) UpdateProcessingStatus(ctx context.Context, id string, status string, processingNotes *string) error
func (s *harvest) RecordQualityCheck(ctx context.Context, id string, qualityData QualityCheckData) error
func (s *harvest) FindByStatusAndDateRange(ctx context.Context, status string, startDate, endDate time.Time, tenantID enum.Tenant) ([]*HarvestDomain, error)
func (s *harvest) GetProcessingMetrics(ctx context.Context, tenantID enum.Tenant, timeRange string) (map[string]interface{}, error)
func (s *harvest) GetCollectionSchedule(ctx context.Context, memberID string) ([]*HarvestDomain, error)

// New structures
type QualityCheckData struct {
    CheckedBy     string     `json:"checked_by"`
    CheckDate     time.Time  `json:"check_date"`
    VisualQuality int        `json:"visual_quality" validate:"gte=1,lte=10"`
    Moisture      *float64   `json:"moisture" validate:"omitempty,gte=0,lte=100"`
    Density       *float64   `json:"density" validate:"omitempty,gte=0"`
    Notes         *string    `json:"notes"`
    Approved      bool       `json:"approved"`
}
```

#### 2.2 Processing Status Enhancement ✅ Extend Domain

**Enhanced Status Flow**:
```
harvested → initial_processing → drying → curing → quality_check → ready → collected
```

**Domain Enhancement**:
```go
// Add to HarvestDomain
ProcessingStage    *string              `json:"processing_stage" bson:"processing_stage"`
ProcessingStarted  *time.Time           `json:"processing_started" bson:"processing_started"`
DryingCompleted    *time.Time           `json:"drying_completed" bson:"drying_completed"`
CuringCompleted    *time.Time           `json:"curing_completed" bson:"curing_completed"`
QualityChecks      *[]QualityCheckData  `json:"quality_checks" bson:"quality_checks"`
ProcessingNotes    *string              `json:"processing_notes" bson:"processing_notes"`
EstimatedReady     *time.Time           `json:"estimated_ready" bson:"estimated_ready"`
```

### Phase 3: Quality Control System (Week 8 - Day 5)

#### 3.1 Quality Control Routes ✅ TDD Implementation

**Test First**:
```go
func TestQualityControl_AdminWorkflow(t *testing.T) {
    // Test admin quality check workflow
}

func TestQualityControl_MultipleChecks(t *testing.T) {
    // Test multiple quality checks per harvest
}
```

**Implementation**:
```go
type QualityCheckRequest struct {
    VisualQuality int      `json:"visual_quality" binding:"required,gte=1,lte=10"`
    Moisture      *float64 `json:"moisture" validate:"omitempty,gte=0,lte=100"`
    Density       *float64 `json:"density" validate:"omitempty,gte=0"`
    Notes         *string  `json:"notes" validate:"omitempty,max=500"`
    Approved      bool     `json:"approved"`
}

func (s harvest) v1_qualityCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Implement quality check workflow
        // Update harvest quality checks array
        // Automatically advance status if approved
        // Generate quality certificate
    }
}
```

#### 3.2 Automated Quality Metrics ✅ Reuse Analytics Pattern

**Following Plant Analytics Pattern**:
```go
func (s *harvest) GetQualityMetrics(ctx context.Context, tenantID enum.Tenant, timeRange string) (map[string]interface{}, error) {
    // Average quality ratings by strain
    // Processing time analysis
    // Quality improvement trends
    // Member satisfaction metrics
}
```

### Phase 4: Member Collection Interface (Week 9 - Day 1)

#### 4.1 Collection Workflow ✅ TDD Implementation

**Test First**:
```go
func TestHarvestCollection_MemberWorkflow(t *testing.T) {
    // Test complete member collection process
}

func TestHarvestCollection_Scheduling(t *testing.T) {
    // Test collection appointment scheduling
}
```

**Collection Process**:
```go
type CollectionRequest struct {
    CollectionMethod string    `json:"collection_method" binding:"required" validate:"oneof=pickup scheduled_delivery"`
    PreferredDate    *time.Time `json:"preferred_date" validate:"omitempty"`
    DeliveryAddress  *string   `json:"delivery_address" validate:"required_if=CollectionMethod scheduled_delivery"`
    SpecialNotes     *string   `json:"special_notes" validate:"omitempty,max=500"`
}

func (s harvest) v1_collectHarvest() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Validate harvest is ready for collection
        // Process collection request
        // Update status to "collected"
        // Generate collection receipt
        // Send notification to member
    }
}
```

#### 4.2 Collection Scheduling ✅ Extend Notification System

**Reuse Existing Notification Pattern**:
```go
// New notification types
type CollectionNotification struct {
    HarvestID       string    `json:"harvest_id"`
    EstimatedReady  time.Time `json:"estimated_ready"`
    CollectionBy    time.Time `json:"collection_by"`
    Instructions    string    `json:"instructions"`
}

// Integration with existing notification system
func (s *harvest) ScheduleCollectionReminders(ctx context.Context, harvestID string) error {
    // Send notification 7 days before ready
    // Send notification 3 days before ready  
    // Send notification when ready
    // Send overdue notifications if not collected
}
```

### Phase 5: Analytics & Reporting Enhancement (Week 9 - Day 2)

#### 5.1 Comprehensive Analytics ✅ Follow Plant Analytics Pattern

**File**: `store/db/harvest.go` (ENHANCE)

**Test First**:
```go
func TestHarvestAnalytics_YieldAnalysis(t *testing.T) {
    // Test yield analytics calculations
}

func TestHarvestAnalytics_ProcessingMetrics(t *testing.T) {
    // Test processing time analytics
}
```

**Implementation** (following Plant pattern):
```go
type HarvestAnalyticsQuery struct {
    TenantId     *enum.Tenant `json:"tenant_id"`
    MemberID     *string      `json:"member_id"`
    TimeRange    *string      `json:"time_range"` // week, month, quarter, year
    ProcessingStage *string   `json:"processing_stage"`
    Strain       *string      `json:"strain"`
}

func (s *harvest) GetYieldAnalytics(ctx context.Context, query HarvestAnalyticsQuery) (map[string]interface{}, error) {
    // Average yield per plant by strain
    // Yield trends over time
    // Quality distribution analysis
    // Processing efficiency metrics
}

func (s *harvest) GetProcessingMetrics(ctx context.Context, query HarvestAnalyticsQuery) (map[string]interface{}, error) {
    // Average processing times
    // Quality improvement during processing
    // Bottleneck identification
    // Resource utilization
}
```

#### 5.2 Dashboard Integration ✅ Extend Existing Endpoints

**Route Enhancement**:
```go
func (s harvest) v1_getHarvestAnalytics() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Comprehensive harvest analytics
        // Processing stage distribution
        // Quality trends
        // Collection efficiency
        // Member satisfaction metrics
    }
}
```

### Phase 6: Enhanced Plant Integration (Week 9 - Day 3)

#### 6.1 Improved Plant Harvest Endpoint ✅ Enhance Existing

**File**: `route/plant.go` (ENHANCE)

**Test Enhancement**:
```go
func TestPlantHarvest_EnhancedWorkflow(t *testing.T) {
    // Test harvest creation with processing workflow
    // Test automatic harvest record creation
    // Test processing timeline calculation
}
```

**Enhanced Implementation**:
```go
func (s plant) v1_harvestPlant() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Existing validation logic (REUSE 100%)
        
        // ENHANCEMENT: Create comprehensive harvest record
        harvestRecord := &db.HarvestDomain{
            PlantID:         &plantID,
            MemberID:        &db.SID(member.ID),
            HarvestDate:     time.Now(),
            Weight:          &req.Weight,
            Quality:         &req.Quality,
            Strain:          plant.Strain,
            Status:          gopkg.Pointer("initial_processing"),
            ProcessingStage: gopkg.Pointer("drying"),
            Notes:           &req.Notes,
            TenantId:        &session.TenantId,
        }

        // Calculate processing timeline
        estimatedReady := time.Now().AddDate(0, 0, 21) // 3 weeks typical processing
        harvestRecord.EstimatedReady = &estimatedReady

        // Save harvest record
        savedHarvest, err := s.store.Db.Harvest.Save(ctx, harvestRecord)
        
        // Update plant with harvest ID (REUSE existing logic)
        err = s.store.Db.Plant.SetHarvestID(ctx, plantID, db.SID(savedHarvest.ID))
        
        // ENHANCEMENT: Schedule processing workflow
        s.scheduleProcessingWorkflow(ctx, savedHarvest)
        
        // ENHANCEMENT: Enhanced response
        c.JSON(http.StatusOK, gin.H{
            "message": "Plant harvested successfully",
            "harvest": savedHarvest.Dto(), // Use existing DTO pattern
            "processing_timeline": map[string]interface{}{
                "estimated_ready": estimatedReady,
                "current_stage":   "drying",
                "next_milestone":  "curing",
            },
        })
    }
}
```

### Phase 7: API Documentation & Testing (Week 9 - Day 4)

#### 7.1 Comprehensive API Documentation ✅ Follow Existing Pattern

**File**: `docs/api-harvest-management.md` (NEW)

**Structure** (following `api-plant-management.md` pattern):
```markdown
# Harvest Management API Documentation

## Overview
Complete harvest lifecycle management from plant harvest through member collection.

## Features
- Complete harvest lifecycle tracking
- Advanced processing workflows
- Quality control systems
- Member collection interface
- Comprehensive analytics

## API Endpoints

### Member Endpoints
- GET /harvest/v1/my-harvests - Harvest history with filtering
- GET /harvest/v1/:id - Detailed harvest information
- PUT /harvest/v1/:id/status - Update harvest status
- POST /harvest/v1/:id/images - Upload harvest images
- POST /harvest/v1/:id/collect - Collection workflow

### Admin Endpoints
- GET /harvest/v1/admin/all - All harvests overview
- GET /harvest/v1/admin/processing - Processing stage tracking
- GET /harvest/v1/admin/analytics - Harvest analytics
- POST /harvest/v1/admin/:id/quality-check - Quality verification
- PUT /harvest/v1/admin/:id/force-status - Force status updates

## Business Rules
[Complete business rule documentation]
```

#### 7.2 Swagger Documentation ✅ Extend Existing

**File**: `docs/swagger.yaml` (ENHANCE)

**Following existing pattern for all new endpoints**:
```yaml
/harvest/v1/my-harvests:
  get:
    summary: Get My Harvests
    parameters:
      - name: status
        in: query
        type: string
        enum: [processing, curing, ready, collected]
      - name: strain
        in: query
        type: string
      # ... additional parameters
    responses:
      200:
        description: Harvest list retrieved successfully
        schema:
          $ref: '#/definitions/HarvestListResponse'
```

### Phase 8: Advanced Features & Integration (Week 9 - Day 5)

#### 8.1 NFT Integration Preparation ✅ Future-Ready

**Domain Enhancement**:
```go
// HarvestDomain already has NFT fields (REUSE)
NFTTokenID         *string `json:"nft_token_id" bson:"nft_token_id"`
NFTContractAddress *string `json:"nft_contract_address" bson:"nft_contract_address"`

// Add NFT workflow integration points
func (s *harvest) PrepareNFTMinting(ctx context.Context, harvestID string) (*NFTMetadata, error) {
    // Prepare harvest data for NFT minting
    // Include quality verification
    // Processing timeline
    // Provenance data
}
```

#### 8.2 Marketplace Integration Preparation ✅ Future-Ready

**Domain Enhancement**:
```go
// Add marketplace preparation fields
MarketplaceEligible *bool      `json:"marketplace_eligible" bson:"marketplace_eligible"`
ProductBatches      *[]string  `json:"product_batches" bson:"product_batches"`
SalePrice          *float64   `json:"sale_price" bson:"sale_price"`

// Integration methods
func (s *harvest) CalculateMarketplaceValue(ctx context.Context, harvestID string) (*MarketplaceValuation, error) {
    // Calculate base price from weight and quality
    // Apply strain premium/discount
    // Consider processing quality
    // Market demand factors
}
```

## Testing Strategy (TDD Throughout)

### Unit Tests ✅ Complete Coverage

**Test Files**:
1. `route/harvest_test.go` - Route testing (following `plant_test.go` pattern)
2. `store/db/harvest_test.go` - Domain testing (following existing patterns)

#### Route Tests - Comprehensive Coverage

```go
// Authentication & Authorization Tests
func TestHarvestRoutes_Authentication(t *testing.T) {
    t.Run("No auth token", func(t *testing.T) {
        // Test all endpoints return 401 without token
    })
    t.Run("Invalid token", func(t *testing.T) {
        // Test all endpoints reject invalid tokens
    })
    t.Run("Expired token", func(t *testing.T) {
        // Test all endpoints reject expired tokens
    })
    t.Run("Insufficient permissions", func(t *testing.T) {
        // Test permission restrictions per endpoint
    })
}

// Member Endpoint Tests
func TestHarvestRoutes_GetMyHarvests(t *testing.T) {
    t.Run("Get harvests with filters", func(t *testing.T) {
        // Test status filter (processing, curing, ready, collected)
        // Test strain filter
        // Test date range filter
        // Test pagination
    })
    t.Run("Empty harvest list", func(t *testing.T) {
        // Test response when no harvests exist
    })
    t.Run("Access control", func(t *testing.T) {
        // Ensure member only sees own harvests
    })
}

func TestHarvestRoutes_GetHarvestDetails(t *testing.T) {
    t.Run("Valid harvest ID", func(t *testing.T) {
        // Test complete harvest detail response
        // Verify all fields present
    })
    t.Run("Invalid harvest ID", func(t *testing.T) {
        // Test 404 response
    })
    t.Run("Unauthorized access", func(t *testing.T) {
        // Test member cannot access others' harvests
    })
    t.Run("Processing timeline display", func(t *testing.T) {
        // Test processing stage information
    })
}

func TestHarvestRoutes_UpdateHarvestStatus(t *testing.T) {
    t.Run("Valid status transition", func(t *testing.T) {
        // Test member-allowed status updates
        // Test status validation
    })
    t.Run("Invalid status transition", func(t *testing.T) {
        // Test blocked transitions
        // Test invalid status values
    })
    t.Run("Status with notes", func(t *testing.T) {
        // Test status update with notes
    })
    t.Run("Unauthorized update", func(t *testing.T) {
        // Test member cannot update others' harvests
    })
}

func TestHarvestRoutes_UploadHarvestImage(t *testing.T) {
    t.Run("Valid image upload", func(t *testing.T) {
        // Test successful image upload
        // Test image URL validation
        // Test description handling
    })
    t.Run("Invalid image data", func(t *testing.T) {
        // Test invalid URL format
        // Test missing required fields
    })
    t.Run("Image limit validation", func(t *testing.T) {
        // Test maximum images per harvest
    })
    t.Run("Ownership validation", func(t *testing.T) {
        // Test member can only upload to own harvests
    })
}

func TestHarvestRoutes_CollectHarvest(t *testing.T) {
    t.Run("Collection workflow - pickup", func(t *testing.T) {
        // Test pickup collection method
        // Test status update to collected
        // Test collection date recording
    })
    t.Run("Collection workflow - delivery", func(t *testing.T) {
        // Test delivery collection method
        // Test address validation
        // Test scheduling
    })
    t.Run("Harvest not ready", func(t *testing.T) {
        // Test collection attempt on non-ready harvest
    })
    t.Run("Already collected", func(t *testing.T) {
        // Test duplicate collection attempt
    })
}

// Admin Endpoint Tests
func TestHarvestRoutes_AdminGetAll(t *testing.T) {
    t.Run("All harvests with filters", func(t *testing.T) {
        // Test tenant-scoped results
        // Test member filter
        // Test status filter
        // Test date range filter
        // Test pagination
    })
    t.Run("Admin permissions", func(t *testing.T) {
        // Test admin role requirement
        // Test member access denied
    })
}

func TestHarvestRoutes_AdminGetProcessing(t *testing.T) {
    t.Run("Processing stage filtering", func(t *testing.T) {
        // Test drying stage filter
        // Test curing stage filter
        // Test quality_check stage filter
    })
    t.Run("Processing metrics", func(t *testing.T) {
        // Test processing time calculations
        // Test overdue processing alerts
    })
}

func TestHarvestRoutes_AdminAnalytics(t *testing.T) {
    t.Run("Yield analytics", func(t *testing.T) {
        // Test average yield calculations
        // Test yield by strain
        // Test yield trends
    })
    t.Run("Quality analytics", func(t *testing.T) {
        // Test quality distribution
        // Test quality trends
        // Test quality by processing stage
    })
    t.Run("Processing analytics", func(t *testing.T) {
        // Test processing time metrics
        // Test bottleneck identification
        // Test efficiency metrics
    })
    t.Run("Collection analytics", func(t *testing.T) {
        // Test collection rates
        // Test collection method distribution
        // Test overdue collections
    })
}

func TestHarvestRoutes_AdminQualityCheck(t *testing.T) {
    t.Run("Quality check workflow", func(t *testing.T) {
        // Test quality check recording
        // Test automatic status advancement
        // Test quality check history
    })
    t.Run("Quality check validation", func(t *testing.T) {
        // Test visual quality range (1-10)
        // Test moisture validation (0-100%)
        // Test density validation
    })
    t.Run("Multiple quality checks", func(t *testing.T) {
        // Test multiple checks per harvest
        // Test check history tracking
        // Test average quality calculation
    })
    t.Run("Quality check approval", func(t *testing.T) {
        // Test approved quality check
        // Test rejected quality check
        // Test status transitions
    })
}

func TestHarvestRoutes_AdminForceStatus(t *testing.T) {
    t.Run("Force status update", func(t *testing.T) {
        // Test admin override capabilities
        // Test reason requirement
        // Test audit logging
    })
    t.Run("Force status validation", func(t *testing.T) {
        // Test valid status values
        // Test reason validation
    })
    t.Run("Force status audit", func(t *testing.T) {
        // Test audit trail creation
        // Test admin user tracking
    })
}
```

#### Domain Tests - Complete Coverage

```go
// Basic CRUD Operations
func TestHarvestDomain_CRUD(t *testing.T) {
    t.Run("Create harvest", func(t *testing.T) {
        // Test harvest creation
        // Test field validation
        // Test tenant isolation
    })
    t.Run("Read harvest", func(t *testing.T) {
        // Test FindByID
        // Test FindByPlantID
        // Test FindByMemberID
        // Test pagination
    })
    t.Run("Update harvest", func(t *testing.T) {
        // Test field updates
        // Test validation on update
        // Test partial updates
    })
    t.Run("Delete harvest", func(t *testing.T) {
        // Test soft delete if implemented
        // Test cascade effects
    })
}

// Processing Workflow Tests
func TestHarvestDomain_ProcessingWorkflow(t *testing.T) {
    t.Run("Status transitions", func(t *testing.T) {
        // Test valid transitions: processing → drying → curing → quality_check → ready → collected
        // Test invalid transitions
        // Test transition validation
    })
    t.Run("Processing timeline", func(t *testing.T) {
        // Test processing stage durations
        // Test estimated ready calculation
        // Test overdue processing detection
    })
    t.Run("Processing notes", func(t *testing.T) {
        // Test processing notes storage
        // Test notes validation
        // Test notes history
    })
}

// Quality Control Tests
func TestHarvestDomain_QualityControl(t *testing.T) {
    t.Run("Quality check recording", func(t *testing.T) {
        // Test quality check data storage
        // Test quality check validation
        // Test multiple checks per harvest
    })
    t.Run("Quality metrics", func(t *testing.T) {
        // Test visual quality validation (1-10)
        // Test moisture validation (0-100%)
        // Test density validation
    })
    t.Run("Quality approval workflow", func(t *testing.T) {
        // Test approval process
        // Test rejection handling
        // Test re-check workflow
    })
}

// Analytics Tests
func TestHarvestDomain_Analytics(t *testing.T) {
    t.Run("Yield analytics", func(t *testing.T) {
        // Test GetYieldAnalytics method
        // Test yield calculations
        // Test strain-based analytics
    })
    t.Run("Processing metrics", func(t *testing.T) {
        // Test GetProcessingMetrics method
        // Test processing time calculations
        // Test efficiency metrics
    })
    t.Run("Quality metrics", func(t *testing.T) {
        // Test GetQualityMetrics method
        // Test quality distribution
        // Test quality trends
    })
}

// Advanced Query Tests
func TestHarvestDomain_AdvancedQueries(t *testing.T) {
    t.Run("Status and date range", func(t *testing.T) {
        // Test FindByStatusAndDateRange
        // Test complex filters
        // Test performance with large datasets
    })
    t.Run("Collection schedule", func(t *testing.T) {
        // Test GetCollectionSchedule
        // Test member-specific schedules
        // Test overdue collections
    })
    t.Run("Processing metrics", func(t *testing.T) {
        // Test GetProcessingMetrics
        // Test tenant isolation
        // Test time range filtering
    })
}
```

#### Integration Tests - End-to-End Coverage

```go
// Plant-Harvest Integration
func TestHarvestIntegration_PlantWorkflow(t *testing.T) {
    t.Run("Complete plant to harvest workflow", func(t *testing.T) {
        // Create plant → grow to flowering → harvest
        // Test plant status updates
        // Test harvest record creation
        // Test plant-harvest linking
    })
    t.Run("Enhanced plant harvest endpoint", func(t *testing.T) {
        // Test enhanced harvest creation
        // Test processing timeline calculation
        // Test automatic processing workflow initiation
    })
    t.Run("Harvest validation", func(t *testing.T) {
        // Test plant readiness validation
        // Test harvest requirements
        // Test quality/weight validation
    })
}

// Plant Slot Integration
func TestHarvestIntegration_SlotRelease(t *testing.T) {
    t.Run("Slot status on harvest", func(t *testing.T) {
        // Test slot release on plant harvest
        // Test slot availability update
        // Test slot reallocation capability
    })
    t.Run("Slot release timing", func(t *testing.T) {
        // Test immediate slot release
        // Test delayed slot release scenarios
    })
}

// Member Integration
func TestHarvestIntegration_MemberWorkflow(t *testing.T) {
    t.Run("Member harvest ownership", func(t *testing.T) {
        // Test member can only access own harvests
        // Test membership validation
        // Test tenant isolation
    })
    t.Run("Member collection workflow", func(t *testing.T) {
        // Test complete collection process
        // Test collection method validation
        // Test collection confirmation
    })
}

// Notification Integration
func TestHarvestIntegration_Notifications(t *testing.T) {
    t.Run("Processing notifications", func(t *testing.T) {
        // Test processing stage notifications
        // Test ready for collection notifications
        // Test overdue collection alerts
    })
    t.Run("Quality check notifications", func(t *testing.T) {
        // Test quality approval notifications
        // Test quality rejection notifications
        // Test re-check notifications
    })
}

// Analytics Integration
func TestHarvestIntegration_Analytics(t *testing.T) {
    t.Run("Cross-system analytics", func(t *testing.T) {
        // Test plant-harvest yield correlation
        // Test member harvest history
        // Test strain performance analytics
    })
    t.Run("Real-time metrics", func(t *testing.T) {
        // Test live processing metrics
        // Test real-time collection status
        // Test dynamic quality trends
    })
}
```

#### Error Handling Tests - Complete Coverage

```go
// Error Response Tests
func TestHarvestRoutes_ErrorHandling(t *testing.T) {
    t.Run("Validation errors", func(t *testing.T) {
        // Test all validation error responses
        // Test error message clarity
        // Test error code consistency
    })
    t.Run("Not found errors", func(t *testing.T) {
        // Test harvest not found
        // Test plant not found
        // Test member not found
    })
    t.Run("Permission errors", func(t *testing.T) {
        // Test unauthorized access
        // Test insufficient permissions
        // Test role-based restrictions
    })
    t.Run("Business logic errors", func(t *testing.T) {
        // Test harvest already collected
        // Test harvest not ready
        // Test invalid status transitions
    })
}

// Edge Case Tests
func TestHarvestRoutes_EdgeCases(t *testing.T) {
    t.Run("Concurrent access", func(t *testing.T) {
        // Test concurrent status updates
        // Test concurrent collection attempts
        // Test race condition handling
    })
    t.Run("Data consistency", func(t *testing.T) {
        // Test transaction rollback scenarios
        // Test partial update failures
        // Test data integrity constraints
    })
    t.Run("Large dataset handling", func(t *testing.T) {
        // Test pagination with large datasets
        // Test query performance
        // Test memory usage
    })
}
```

#### Performance Tests

```go
// Performance and Load Tests
func TestHarvestRoutes_Performance(t *testing.T) {
    t.Run("Response time benchmarks", func(t *testing.T) {
        // Test all endpoints under load
        // Verify sub-200ms response times
        // Test database query optimization
    })
    t.Run("Concurrent user simulation", func(t *testing.T) {
        // Test multiple members accessing harvests
        // Test admin operations under load
        // Test system stability
    })
    t.Run("Large dataset performance", func(t *testing.T) {
        // Test with thousands of harvests
        // Test pagination performance
        // Test analytics calculation speed
    })
}
```

### Test Data Management

```go
// Test Setup and Teardown
func setupHarvestTestData(t *testing.T) *HarvestTestSetup {
    // Create test tenant
    // Create test members
    // Create test plants
    // Create test harvests in various stages
    // Create test quality checks
}

func teardownHarvestTestData(t *testing.T, setup *HarvestTestSetup) {
    // Clean up test data
    // Reset collections
    // Clear cache
}

// Test Data Factories
func createTestHarvest(t *testing.T, options HarvestOptions) *db.HarvestDomain
func createTestQualityCheck(t *testing.T, harvestID string) *QualityCheckData
func createTestCollectionRequest(t *testing.T, method string) *CollectionRequest
```

### Test Coverage Requirements

**Minimum Coverage Targets**:
- **Route Tests**: 100% endpoint coverage, 95% line coverage
- **Domain Tests**: 100% method coverage, 95% line coverage  
- **Integration Tests**: 100% workflow coverage
- **Error Handling**: 100% error scenario coverage
- **Edge Cases**: 90% edge case coverage
- **Performance**: All endpoints under 200ms response time

**Coverage Verification**:
```bash
# Run tests with coverage
go test -cover ./route/harvest_test.go
go test -cover ./store/db/harvest_test.go

# Generate coverage reports
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests ✅ End-to-End Coverage

**Test Scenarios**:
1. Complete harvest lifecycle (plant → harvest → processing → collection)
2. Quality control workflow
3. Member collection process
4. Admin management workflows
5. Analytics and reporting
6. Error handling and edge cases

## Implementation Timeline

### Week 8 - Backend Enhancement
- **Day 1**: Dedicated harvest routes implementation (TDD)
- **Day 2**: Enhanced permission system and route completion
- **Day 3**: Advanced processing workflows
- **Day 4**: Processing status enhancement and validation
- **Day 5**: Quality control system implementation

### Week 9 - Advanced Features & Documentation
- **Day 1**: Member collection interface
- **Day 2**: Analytics & reporting enhancement
- **Day 3**: Enhanced plant integration
- **Day 4**: API documentation and Swagger updates
- **Day 5**: Advanced features and integration preparation

## Dependencies

### ✅ Satisfied Dependencies
- Plant Management System (Task 1.7) - ✅ Complete
- Plant Slot Management System - ✅ Complete  
- Membership Management System - ✅ Complete
- Authentication & Authorization - ✅ Complete
- Database Schema (HarvestDomain) - ✅ Complete

### Integration Points
- **Plant System**: Harvest creation and slot release
- **Member System**: Ownership verification and notifications
- **Storage System**: Image upload for harvest documentation
- **Notification System**: Processing updates and collection reminders

## Code Reuse Summary

### 100% Reuse (No Changes)
- `HarvestDomain` core structure
- `BaseDomain` inheritance
- Database connection and transactions
- Authentication middleware
- Error handling patterns
- Audit logging system

### 90% Reuse (Minor Extensions)
- Database repository methods (add new query methods)
- Validation system (extend existing patterns)
- DTO/Response patterns (follow existing structure)

### 50% Reuse (Enhancement)
- Plant harvest endpoint (enhanced workflow)
- Permission system (add harvest permissions)
- Analytics patterns (extend to harvest data)

### New Implementation (Following Patterns)
- Dedicated harvest routes (following plant route patterns)
- Quality control workflows (following care record patterns)
- Collection interface (following membership patterns)

## Success Criteria

### ✅ Technical Requirements
1. **Complete API Coverage**: 10 new endpoints following existing patterns
2. **TDD Implementation**: 100% test coverage for all new functionality
3. **Code Reuse**: 80%+ reuse of existing codebase patterns
4. **Performance**: Sub-200ms response times for all endpoints
5. **Security**: Complete authentication and authorization
6. **Documentation**: Comprehensive API and integration documentation

### ✅ Business Requirements
1. **Complete Harvest Lifecycle**: From plant harvest to member collection
2. **Quality Control**: Admin verification and approval workflows
3. **Processing Tracking**: Detailed status and timeline management
4. **Member Experience**: Easy collection scheduling and status tracking
5. **Analytics**: Comprehensive yield and processing metrics
6. **Integration Ready**: Prepared for NFT and marketplace integration

## Architecture Compliance

### ✅ Pattern Compliance
- **Domain Models**: 100% BaseDomain inheritance
- **Repository Pattern**: Standard CRUD and query methods
- **Route Structure**: Version-prefixed endpoints with middleware
- **Error Handling**: Standard ecode.Error patterns
- **Authentication**: Bearer token with permission-based access
- **Validation**: Comprehensive input validation
- **Audit Logging**: Complete action tracking

### ✅ Database Design
- **Tenant Isolation**: All operations scoped to tenant
- **Relationships**: Proper foreign key references
- **Indexing**: Optimized for query performance
- **Validation**: Domain-level and database-level constraints

### ✅ Security Implementation
- **Authorization**: Role-based access control
- **Input Validation**: Comprehensive request validation
- **Data Isolation**: Tenant-specific data access
- **Audit Trail**: Complete operation logging

This comprehensive enhancement maintains 100% architectural compliance while providing a complete, production-ready harvest management system that seamlessly integrates with existing components and prepares for future NFT and marketplace integration.
