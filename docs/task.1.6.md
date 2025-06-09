# Task 1.6: Plant Slot Management System - Detailed Implementation Plan

## Overview
Implementation of the Plant Slot Management system following existing architectural patterns, reusing established database models, route patterns, and business logic structures from the membership management system.

## Implementation Strategy

### Phase 1: Permission System Enhancement (Week 6 - Day 1)

#### 1.1 Extend Permission Enums
**File**: `pkg/enum/index.go`
**Dependencies**: None
**Reuse**: Follow `PermissionMembership*` pattern

**Implementation**:
```go
// Add to existing permission constants
PermissionPlantSlotView     Permission = "plant_slot_view"
PermissionPlantSlotCreate   Permission = "plant_slot_create" 
PermissionPlantSlotUpdate   Permission = "plant_slot_update"
PermissionPlantSlotDelete   Permission = "plant_slot_delete"
PermissionPlantSlotManage   Permission = "plant_slot_manage" // Admin-level
PermissionPlantSlotTransfer Permission = "plant_slot_transfer"
PermissionPlantSlotAssign   Permission = "plant_slot_assign"
```

**Testing**: Unit tests for permission validation
**Documentation**: Update API documentation with new permissions

### Phase 2: Database Schema Enhancement (Week 6 - Day 2)

#### 2.1 Enhance PlantSlotDomain Model
**File**: `store/db/plant_slot.go` 
**Dependencies**: BaseDomain, enum.Tenant
**Reuse**: Follows existing `MembershipDomain` pattern

**Current Status**: ✅ Already exists, needs enhancement
**Implementation**: Add missing DTO methods following `MembershipDomain` pattern

```go
// Add missing DTO methods to existing PlantSlotDomain
func (s PlantSlotDomain) BaseDto() *PlantSlotBaseDto {
    return &PlantSlotBaseDto{
        ID:         SID(s.ID),
        SlotNumber: gopkg.Value(s.SlotNumber),
        MemberID:   gopkg.Value(s.MemberID),
        Status:     gopkg.Value(s.Status),
        Location:   gopkg.Value(s.Location),
        Position:   s.Position,
        UpdatedAt:  gopkg.Value(s.UpdatedAt),
    }
}

func (s PlantSlotDomain) DetailDto() *PlantSlotDetailDto {
    return &PlantSlotDetailDto{
        ID:             SID(s.ID),
        SlotNumber:     gopkg.Value(s.SlotNumber),
        MemberID:       gopkg.Value(s.MemberID),
        MembershipID:   gopkg.Value(s.MembershipID),
        Status:         gopkg.Value(s.Status),
        Location:       gopkg.Value(s.Location),
        Position:       s.Position,
        Notes:          gopkg.Value(s.Notes),
        MaintenanceLog: gopkg.Value(s.MaintenanceLog),
        LastCleanDate:  gopkg.Value(s.LastCleanDate),
        CreatedAt:      gopkg.Value(s.CreatedAt),
        UpdatedAt:      gopkg.Value(s.UpdatedAt),
    }
}
```

#### 2.2 Add Query Support
**File**: `store/db/plant_slot.go`
**Reuse**: Follow `MembershipQuery` pattern

```go
type PlantSlotQuery struct {
    Query               `bson:",inline"`
    Status              *string      `json:"status" form:"status"`
    MemberID            *string      `json:"member_id" form:"member_id"`
    MembershipID        *string      `json:"membership_id" form:"membership_id"`
    Location            *string      `json:"location" form:"location"`
    AvailableOnly       *bool        `json:"available_only" form:"available_only"`
    MaintenanceRequired *bool        `json:"maintenance_required" form:"maintenance_required"`
    TenantId            *enum.Tenant `json:"tenant_id" form:"tenant_id"`
}

func (s *PlantSlotQuery) Build() *PlantSlotQuery {
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
    if s.Location != nil {
        query.Filter["location"] = *s.Location
    }
    if s.AvailableOnly != nil && *s.AvailableOnly {
        query.Filter["status"] = "available"
    }
    if s.TenantId != nil {
        query.Filter["tenant_id"] = *s.TenantId
    }

    s.Query = query
    return s
}
```

### Phase 3: Business Logic Implementation (Week 6 - Days 3-4)

#### 3.1 Slot Allocation Service
**File**: `store/db/plant_slot.go`
**Dependencies**: Membership system, Member system
**Reuse**: Follow membership allocation patterns

**Key Methods**:
```go
// Enhanced allocation with membership validation
func (s *plantSlot) AllocateToMember(ctx context.Context, memberID string, membershipID string, quantity int) ([]*PlantSlotDomain, error)

// Validate allocation capacity
func (s *plantSlot) ValidateAllocation(ctx context.Context, memberID string, requestedSlots int) error

// Release slots on membership expiry
func (s *plantSlot) ReleaseSlots(ctx context.Context, membershipID string) error

// Transfer between members
func (s *plantSlot) TransferSlots(ctx context.Context, fromMemberID, toMemberID string, slotIDs []string) error
```

#### 3.2 Status Management System
**File**: `store/db/plant_slot.go`
**Reuse**: Follow membership status patterns

**Status Transitions**:
- `available` → `allocated` (member gets slot)
- `allocated` → `occupied` (plant assigned)
- `occupied` → `maintenance` (maintenance required)
- `maintenance` → `available` (ready for reallocation)
- `allocated` → `available` (membership expired)

### Phase 4: API Route Implementation (Week 6 - Day 5)

#### 4.1 Plant Slot Routes
**File**: `route/plant_slot.go`
**Dependencies**: Authentication, permissions, database
**Reuse**: Follow `route/membership.go` pattern exactly

**Route Structure**: `/plant-slots/v1/*`

```go
type plantSlot struct {
    *middleware
}

func init() {
    handlers = append(handlers, func(m *middleware, r *gin.Engine) {
        s := plantSlot{m}

        v1 := r.Group("/plant-slots/v1")
        {
            // Member endpoints
            v1.GET("/my-slots", s.BearerAuth(enum.PermissionPlantSlotView), s.v1_getMySlots())
            v1.POST("/request", s.BearerAuth(enum.PermissionPlantSlotCreate), s.v1_requestSlots())
            v1.GET("/:id", s.BearerAuth(enum.PermissionPlantSlotView), s.v1_getSlotDetails())
            v1.PUT("/:id/status", s.BearerAuth(enum.PermissionPlantSlotUpdate), s.v1_updateSlotStatus())
            v1.POST("/:id/maintenance", s.BearerAuth(enum.PermissionPlantSlotUpdate), s.v1_reportMaintenance())
            
            // Transfer functionality
            v1.POST("/transfer", s.BearerAuth(enum.PermissionPlantSlotTransfer), s.v1_transferSlots())
            
            // Admin endpoints
            admin := v1.Group("/admin")
            {
                admin.GET("/all", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_getAllSlots())
                admin.POST("/assign", s.BearerAuth(enum.PermissionPlantSlotAssign), s.v1_assignSlots())
                admin.GET("/maintenance", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_getMaintenanceSlots())
                admin.GET("/analytics", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_getSlotAnalytics())
                admin.PUT("/:id/force-status", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_forceStatusUpdate())
            }
        }
    })
}
```

#### 4.2 Request/Response Structures
**Reuse**: Follow membership request/response patterns

```go
type SlotRequestRequest struct {
    Quantity int    `json:"quantity" binding:"required,min=1,max=10"`
    Location string `json:"preferred_location" validate:"omitempty"`
}

type TransferSlotsRequest struct {
    ToMemberID string   `json:"to_member_id" binding:"required,len=24"`
    SlotIDs    []string `json:"slot_ids" binding:"required,min=1,dive,len=24"`
    Reason     string   `json:"reason" binding:"required"`
}

type MaintenanceRequest struct {
    Description string `json:"description" binding:"required"`
    Priority    string `json:"priority" validate:"oneof=low normal high"`
}
```

### Phase 5: Error Handling Enhancement (Week 6 - Day 6)

#### 5.1 Enhance Cannabis Error Codes
**File**: `pkg/ecode/cannabis.go`
**Dependencies**: None
**Reuse**: Existing error patterns

**Additional Error Codes**:
```go
// Plant slot allocation errors
PlantSlotInsufficientSlots   = New(http.StatusConflict, "plant_slot_insufficient_slots")
PlantSlotMembershipRequired  = New(http.StatusForbidden, "plant_slot_membership_required")
PlantSlotTransferFailed      = New(http.StatusConflict, "plant_slot_transfer_failed")
PlantSlotMaintenanceRequired = New(http.StatusConflict, "plant_slot_maintenance_required")
```

### Phase 6: Cache Integration (Week 7 - Day 1)

#### 6.1 Redis Cache Implementation
**File**: `store/index.go`
**Dependencies**: Redis, existing cache patterns
**Reuse**: Follow user/membership cache patterns

```go
// Cache key patterns
const (
    PlantSlotCacheKey    = "plant_slot:%s"
    MemberSlotsCacheKey  = "member_slots:%s"
    LocationSlotsCacheKey = "location_slots:%s:%s" // location:status
)

// Cache methods
func (s Store) GetMemberSlots(ctx context.Context, memberID string) ([]*db.PlantSlotBaseDto, error)
func (s Store) InvalidateMemberSlots(ctx context.Context, memberID string) error
func (s Store) CacheSlotStatus(ctx context.Context, slotID string, status string) error
```

### Phase 7: Integration with Membership System (Week 7 - Day 2)

#### 7.1 Membership-PlantSlot Lifecycle
**File**: `route/membership.go`
**Dependencies**: PlantSlot system
**Enhancement**: Add slot allocation to membership purchase

```go
// Enhance v1_purchaseMembership to auto-allocate slots
func (s membership) v1_purchaseMembership() gin.HandlerFunc {
    // ... existing membership creation code ...
    
    // After membership created successfully:
    slots, err := s.store.Db.PlantSlot.AllocateToMember(ctx, 
        db.SID(member.ID), 
        db.SID(savedMembership.ID), 
        config.SlotAllocation)
    if err != nil {
        // Handle allocation failure - may need to rollback membership
        s.store.Db.Membership.UpdateStatus(ctx, db.SID(savedMembership.ID), "pending_slots")
    }
}
```

### Phase 8: Test-Driven Development (TDD) Implementation (Week 7 - Days 3-4)

#### 8.1 TDD Approach: Write Tests First
**File**: `route/plant_slot_test.go`, `store/db/plant_slot_test.go`
**Dependencies**: Test infrastructure
**Reuse**: Follow `route/membership_test.go` pattern

#### 8.2 Business Logic Test Coverage (100% Required)

##### 8.2.1 Slot Allocation Business Logic Tests
```go
func TestPlantSlot_AllocateToMember(t *testing.T) {
    tests := []struct {
        name           string
        memberID       string
        membershipID   string
        quantity       int
        existingSlots  int
        availableSlots int
        wantErr        bool
        expectedError  string
    }{
        {
            name:           "successful_allocation_basic_member",
            memberID:       "member123",
            membershipID:   "membership456", 
            quantity:       2,
            existingSlots:  0,
            availableSlots: 10,
            wantErr:        false,
        },
        {
            name:           "insufficient_available_slots",
            memberID:       "member123",
            membershipID:   "membership456",
            quantity:       5,
            existingSlots:  0,
            availableSlots: 3,
            wantErr:        true,
            expectedError:  "plant_slot_insufficient_slots",
        },
        {
            name:           "member_already_has_slots",
            memberID:       "member123",
            membershipID:   "membership456",
            quantity:       2,
            existingSlots:  3,
            availableSlots: 10,
            wantErr:        true,
            expectedError:  "plant_slot_already_allocated",
        },
        {
            name:           "concurrent_allocation_conflict",
            memberID:       "member123",
            membershipID:   "membership456",
            quantity:       1,
            existingSlots:  0,
            availableSlots: 1,
            wantErr:        false, // First request should succeed
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // TDD: Test implementation before business logic
            result, err := plantSlotService.AllocateToMember(ctx, tt.memberID, tt.membershipID, tt.quantity)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.Len(t, result, tt.quantity)
                // Verify slot status is 'allocated'
                for _, slot := range result {
                    assert.Equal(t, "allocated", *slot.Status)
                    assert.Equal(t, tt.memberID, *slot.MemberID)
                    assert.Equal(t, tt.membershipID, *slot.MembershipID)
                }
            }
        })
    }
}
```

##### 8.2.2 Status Transition Business Logic Tests
```go
func TestPlantSlot_StatusTransitions(t *testing.T) {
    tests := []struct {
        name           string
        currentStatus  string
        targetStatus   string
        wantErr        bool
        expectedError  string
    }{
        {"available_to_allocated", "available", "allocated", false, ""},
        {"allocated_to_occupied", "allocated", "occupied", false, ""},
        {"occupied_to_maintenance", "occupied", "maintenance", false, ""},
        {"maintenance_to_available", "maintenance", "available", false, ""},
        {"invalid_available_to_occupied", "available", "occupied", true, "invalid_status_transition"},
        {"invalid_occupied_to_allocated", "occupied", "allocated", true, "invalid_status_transition"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup slot with current status
            slot := createTestSlot(t, tt.currentStatus)
            
            // Test status transition
            err := plantSlotService.UpdateStatus(ctx, slot.ID, tt.targetStatus)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                // Verify status updated
                updatedSlot, _ := plantSlotService.FindByID(ctx, slot.ID)
                assert.Equal(t, tt.targetStatus, *updatedSlot.Status)
            }
        })
    }
}
```

##### 8.2.3 Transfer Business Logic Tests
```go
func TestPlantSlot_TransferSlots(t *testing.T) {
    tests := []struct {
        name              string
        fromMemberID      string
        toMemberID        string
        slotIDs           []string
        fromMembershipValid bool
        toMembershipValid   bool
        wantErr           bool
        expectedError     string
    }{
        {
            name:                "successful_transfer",
            fromMemberID:        "member1",
            toMemberID:          "member2", 
            slotIDs:             []string{"slot1", "slot2"},
            fromMembershipValid: true,
            toMembershipValid:   true,
            wantErr:             false,
        },
        {
            name:                "transfer_to_member_without_membership",
            fromMemberID:        "member1",
            toMemberID:          "member2",
            slotIDs:             []string{"slot1"},
            fromMembershipValid: true,
            toMembershipValid:   false,
            wantErr:             true,
            expectedError:       "membership_required",
        },
        {
            name:                "transfer_occupied_slots",
            fromMemberID:        "member1",
            toMemberID:          "member2",
            slotIDs:             []string{"slot_occupied"},
            fromMembershipValid: true,
            toMembershipValid:   true,
            wantErr:             true,
            expectedError:       "plant_slot_occupied_cannot_transfer",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test data
            setupTransferTestData(t, tt)
            
            // Test transfer
            err := plantSlotService.TransferSlots(ctx, tt.fromMemberID, tt.toMemberID, tt.slotIDs)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                // Verify ownership transfer
                for _, slotID := range tt.slotIDs {
                    slot, _ := plantSlotService.FindByID(ctx, slotID)
                    assert.Equal(t, tt.toMemberID, *slot.MemberID)
                }
            }
        })
    }
}
```

##### 8.2.4 Maintenance Business Logic Tests
```go
func TestPlantSlot_MaintenanceTracking(t *testing.T) {
    tests := []struct {
        name            string
        slotStatus      string
        description     string
        staffID         string
        wantErr         bool
        expectedError   string
    }{
        {
            name:        "successful_maintenance_log",
            slotStatus:  "occupied",
            description: "Weekly cleaning and inspection",
            staffID:     "staff123",
            wantErr:     false,
        },
        {
            name:          "maintenance_on_available_slot",
            slotStatus:    "available", 
            description:   "Routine maintenance",
            staffID:       "staff123",
            wantErr:       false,
        },
        {
            name:          "invalid_staff_id",
            slotStatus:    "occupied",
            description:   "Maintenance",
            staffID:       "invalid",
            wantErr:       true,
            expectedError: "invalid_staff_id",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            slot := createTestSlot(t, tt.slotStatus)
            
            err := plantSlotService.AddMaintenanceLog(ctx, slot.ID, tt.description, tt.staffID)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                
                // Verify maintenance log added
                updatedSlot, _ := plantSlotService.FindByID(ctx, slot.ID)
                assert.NotEmpty(t, *updatedSlot.MaintenanceLog)
                
                latestLog := (*updatedSlot.MaintenanceLog)[len(*updatedSlot.MaintenanceLog)-1]
                assert.Equal(t, tt.description, *latestLog.Description)
                assert.Equal(t, tt.staffID, *latestLog.PerformedBy)
            }
        })
    }
}
```

#### 8.3 Integration Tests with Business Logic Coverage

##### 8.3.1 Membership-PlantSlot Integration Tests
```go
func TestIntegration_MembershipPlantSlotLifecycle(t *testing.T) {
    // Test complete lifecycle: membership purchase -> slot allocation -> slot usage -> membership expiry -> slot release
    
    // 1. Create member with KYC approved
    member := createTestMember(t, "approved")
    
    // 2. Purchase membership - should auto-allocate slots
    membership := purchaseTestMembership(t, member.ID, "basic") // 2 slots
    
    // 3. Verify slots allocated
    slots, err := plantSlotService.FindByMemberID(ctx, member.ID)
    assert.NoError(t, err)
    assert.Len(t, slots, 2)
    assert.Equal(t, "allocated", *slots[0].Status)
    
    // 4. Assign plant to slot
    err = plantSlotService.UpdateStatus(ctx, slots[0].ID, "occupied")
    assert.NoError(t, err)
    
    // 5. Expire membership
    err = membershipService.ExpireMembership(ctx, membership.ID)
    assert.NoError(t, err)
    
    // 6. Verify slots released
    updatedSlots, err := plantSlotService.FindByMemberID(ctx, member.ID)
    assert.NoError(t, err)
    for _, slot := range updatedSlots {
        assert.Equal(t, "available", *slot.Status)
        assert.Nil(t, slot.MemberID)
    }
}
```

##### 8.3.2 Concurrency Tests for Business Logic
```go
func TestConcurrency_SlotAllocation(t *testing.T) {
    // Test concurrent allocation requests for the same slots
    const numConcurrentRequests = 10
    const availableSlots = 5
    
    // Setup available slots
    setupAvailableSlots(t, availableSlots)
    
    var wg sync.WaitGroup
    var successCount int64
    var errorCount int64
    
    for i := 0; i < numConcurrentRequests; i++ {
        wg.Add(1)
        go func(memberIndex int) {
            defer wg.Done()
            
            memberID := fmt.Sprintf("member%d", memberIndex)
            membershipID := fmt.Sprintf("membership%d", memberIndex)
            
            _, err := plantSlotService.AllocateToMember(ctx, memberID, membershipID, 1)
            if err != nil {
                atomic.AddInt64(&errorCount, 1)
            } else {
                atomic.AddInt64(&successCount, 1)
            }
        }(i)
    }
    
    wg.Wait()
    
    // Only availableSlots should succeed
    assert.Equal(t, int64(availableSlots), successCount)
    assert.Equal(t, int64(numConcurrentRequests-availableSlots), errorCount)
    
    // Verify no overselling
    totalAllocated := countAllocatedSlots(t)
    assert.Equal(t, availableSlots, totalAllocated)
}
```

#### 8.4 API Route Tests with Business Logic
```go
func TestAPI_PlantSlotRoutes(t *testing.T) {
    router, store := setupPlantSlotTestRouter(t)
    
    tests := []struct {
        name           string
        method         string
        url            string
        body           interface{}
        setupFunc      func(t *testing.T) string // Returns auth token
        expectedStatus int
        validateFunc   func(t *testing.T, response *httptest.ResponseRecorder)
    }{
        {
            name:   "request_slots_success",
            method: "POST",
            url:    "/plant-slots/v1/request",
            body: map[string]interface{}{
                "quantity": 2,
                "preferred_location": "greenhouse-1",
            },
            setupFunc: func(t *testing.T) string {
                return createMemberWithMembership(t, "basic") // Returns auth token
            },
            expectedStatus: http.StatusCreated,
            validateFunc: func(t *testing.T, response *httptest.ResponseRecorder) {
                var result map[string]interface{}
                json.Unmarshal(response.Body.Bytes(), &result)
                
                slots := result["slots"].([]interface{})
                assert.Len(t, slots, 2)
                
                // Verify slots are allocated
                for _, slot := range slots {
                    slotData := slot.(map[string]interface{})
                    assert.Equal(t, "allocated", slotData["status"])
                }
            },
        },
        {
            name:   "request_slots_without_membership",
            method: "POST", 
            url:    "/plant-slots/v1/request",
            body: map[string]interface{}{
                "quantity": 1,
            },
            setupFunc: func(t *testing.T) string {
                return createMemberWithoutMembership(t) // Returns auth token
            },
            expectedStatus: http.StatusForbidden,
            validateFunc: func(t *testing.T, response *httptest.ResponseRecorder) {
                var result map[string]interface{}
                json.Unmarshal(response.Body.Bytes(), &result)
                assert.Equal(t, "membership_required", result["error"])
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            token := tt.setupFunc(t)
            
            body, _ := json.Marshal(tt.body)
            req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(body))
            req.Header.Set("Authorization", "Bearer "+token)
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            if tt.validateFunc != nil {
                tt.validateFunc(t, w)
            }
        })
    }
}
```

#### 8.5 Error Handling Tests
```go
func TestErrorHandling_PlantSlotOperations(t *testing.T) {
    tests := []struct {
        name          string
        operation     func() error
        expectedError string
        shouldLog     bool
    }{
        {
            name: "allocation_insufficient_slots",
            operation: func() error {
                return plantSlotService.AllocateToMember(ctx, "member1", "membership1", 100)
            },
            expectedError: "plant_slot_insufficient_slots",
            shouldLog:     true,
        },
        {
            name: "transfer_to_nonexistent_member",
            operation: func() error {
                return plantSlotService.TransferSlots(ctx, "member1", "nonexistent", []string{"slot1"})
            },
            expectedError: "member_not_found",
            shouldLog:     true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.operation()
            
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedError)
            
            if tt.shouldLog {
                // Verify error was logged properly
                // Implementation depends on logging framework
            }
        })
    }
}
```

#### 8.6 Performance Tests for Business Logic
```go
func BenchmarkPlantSlot_AllocationPerformance(b *testing.B) {
    // Setup
    setupPerformanceTestData(b)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        memberID := fmt.Sprintf("member%d", i)
        membershipID := fmt.Sprintf("membership%d", i)
        
        plantSlotService.AllocateToMember(ctx, memberID, membershipID, 1)
    }
}

func TestPerformance_ConcurrentSlotOperations(t *testing.T) {
    const numOperations = 1000
    const concurrency = 50
    
    start := time.Now()
    
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, concurrency)
    
    for i := 0; i < numOperations; i++ {
        wg.Add(1)
        go func(index int) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Perform slot operation
            memberID := fmt.Sprintf("member%d", index)
            plantSlotService.FindByMemberID(ctx, memberID)
        }(i)
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    // Assert performance requirements
    assert.Less(t, duration, 10*time.Second, "Operations should complete within 10 seconds")
    
    operationsPerSecond := float64(numOperations) / duration.Seconds()
    assert.Greater(t, operationsPerSecond, 100.0, "Should handle at least 100 operations per second")
}
```

#### 8.7 Test Coverage Requirements
- **Business Logic Coverage**: 100% required
- **API Route Coverage**: 95% minimum
- **Error Path Coverage**: 90% minimum
- **Integration Test Coverage**: 85% minimum

#### 8.8 Test Data Management
```go
// Test helpers for consistent data setup
func createTestMember(t *testing.T, kycStatus string) *db.MemberDomain
func createTestMembership(t *testing.T, memberID string, tier string) *db.MembershipDomain
func createTestSlot(t *testing.T, status string) *db.PlantSlotDomain
func setupAvailableSlots(t *testing.T, count int) []*db.PlantSlotDomain
func cleanupTestData(t *testing.T)
```

### Phase 9: Documentation and Monitoring (Week 7 - Day 5)

#### 9.1 API Documentation
**File**: `docs/swagger.yaml`
**Update**: Add PlantSlot endpoints following existing patterns

#### 9.2 Architecture Documentation
**File**: `docs/architecture.md`
**Update**: Document PlantSlot integration points

#### 9.3 Monitoring Integration
**Implementation**: Add metrics for slot allocation, utilization rates, maintenance tracking

## Technical Specifications

### Database Indexes
```javascript
// Required indexes for performance
{
    {member_id: 1, status: 1},           // Member slot queries
    {membership_id: 1},                  // Membership integration
    {status: 1, location: 1},            // Availability queries
    {last_clean_date: 1},                // Maintenance scheduling
    {slot_number: 1, tenant_id: 1}       // Unique constraint
}
```

### Caching Strategy
- **Member Slots**: 4-hour TTL
- **Slot Status**: 1-hour TTL
- **Location Availability**: 30-minute TTL
- **Maintenance Queue**: 15-minute TTL

### Error Handling Strategy
- Use existing `ecode.Error` pattern
- Cannabis-specific error codes
- Comprehensive error context
- Audit trail integration

### Security Considerations
- Permission-based access control
- Tenant isolation
- Audit logging for all operations
- Rate limiting on allocation endpoints

## Integration Points

### With Membership System
- Auto-allocation on membership purchase
- Auto-release on membership expiry
- Slot count validation
- Transfer permissions

### With Plant Management System (Future)
- Plant assignment to slots
- Occupancy tracking
- Harvest integration
- Growth monitoring

### With NFT System (Future)
- Slot tokenization
- Ownership verification
- Transfer mechanisms
- Blockchain integration

## Success Criteria

### Functional Requirements
- ✅ Slot allocation following membership tiers
- ✅ Status management with proper transitions
- ✅ Transfer functionality between members
- ✅ Maintenance tracking and scheduling
- ✅ Integration with membership lifecycle

### Non-Functional Requirements
- ✅ Sub-second response times for slot queries
- ✅ Concurrent allocation handling
- ✅ 99.9% data consistency
- ✅ Comprehensive audit trails
- ✅ Scalable to 10,000+ slots per tenant

### Quality Assurance
- ✅ 100% business logic test coverage
- ✅ Zero security vulnerabilities
- ✅ Complete API documentation
- ✅ Performance benchmarks met
- ✅ Integration tests passing

## Risk Mitigation

### Concurrency Issues
- Optimistic locking on slot allocation
- Transaction rollback mechanisms
- Queue-based allocation for high load

### Data Integrity
- Foreign key constraints
- Status transition validation
- Audit trail verification

### Performance Bottlenecks
- Database query optimization
- Caching layer implementation
- Background job processing

## Maintenance and Operations

### Monitoring Metrics
- Slot allocation rate
- Maintenance queue length
- Transfer success rate
- Cache hit ratio
- Error rate by endpoint

### Operational Procedures
- Daily slot status reconciliation
- Weekly maintenance scheduling
- Monthly utilization reporting
- Quarterly capacity planning

This implementation plan ensures 100% compliance with existing architectural patterns while providing a robust, scalable plant slot management system that integrates seamlessly with the current membership and authentication systems. The TDD approach guarantees comprehensive test coverage of all business logic.
