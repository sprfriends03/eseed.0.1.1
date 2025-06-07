# Task 1.5: Membership Management System

## Status: 📋 PLANNED - READY FOR IMPLEMENTATION

## Objective
Implement comprehensive membership management system as an extension of the existing Member system, providing manual admin-controlled membership creation, renewal, and lifecycle management with full audit trails and integration with the existing Member and User models.

## Implementation Strategy
**Maximize Code Reuse**: This implementation leverages the existing `MembershipDomain` model, authentication infrastructure, member system (task 1.3), eKYC system (task 1.4), error handling, and database patterns to minimize new code and maintain architectural consistency. Manual payment processing will be handled through admin dashboard interfaces.

## Architecture Overview

### Current Infrastructure Analysis ✅ COMPLETED
**Existing Components Successfully Identified:**
- ✅ `MembershipDomain` model exists in `store/db/membership.go` with complete CRUD operations
- ✅ Member model has `CurrentMembershipID` field linking to memberships
- ✅ Authentication system supports membership status validation via `RequireMembershipStatus` middleware
- ✅ Error handling includes comprehensive membership-specific error codes in `pkg/ecode/cannabis.go`
- ✅ Audit logging system ready for membership operations
- ✅ Base permission system needs extension for membership management permissions

## Detailed Implementation Checklist

### Phase 1: Permission System Enhancement ✅ PLANNED

#### 1.1 Extend Permission Enums in `pkg/enum/index.go`
**Location: After existing KYC permissions (around line 63)**

**Tasks:**
- [ ] **Add Membership Management Permissions (after `PermissionKYCVerify`):**
  - [ ] Add `PermissionMembershipView   Permission = "membership_view"`
  - [ ] Add `PermissionMembershipCreate Permission = "membership_create"`
  - [ ] Add `PermissionMembershipUpdate Permission = "membership_update"`
  - [ ] Add `PermissionMembershipDelete Permission = "membership_delete"`
  - [ ] Add `PermissionMembershipRenew  Permission = "membership_renew"`
  - [ ] Add `PermissionMembershipManage Permission = "membership_manage"` // Admin-level permission

- [ ] **Update Permission Arrays:**
  - [ ] Add new permissions to `PermissionTenantValues()` function (around line 68)
  - [ ] Add new permissions to `PermissionRootValues()` function (around line 95)

#### 1.2 Create Membership Status Enums
**Location: After existing MemberStatus enum (around line 200)**

**Tasks:**
- [ ] **Add MembershipStatus enum type:**
  - [ ] Add `type MembershipStatus string`
  - [ ] Add constants: `MembershipStatusPending`, `MembershipStatusActive`, `MembershipStatusExpired`, `MembershipStatusCanceled`, `MembershipStatusSuspended`
  - [ ] Add `MembershipStatusValues()` function
  - [ ] Add to `Tags()` function mapping

- [ ] **Add MembershipType enum:**
  - [ ] Add `type MembershipType string`
  - [ ] Add constants: `MembershipTypeBasic`, `MembershipTypePremium`, `MembershipTypeVIP`
  - [ ] Add `MembershipTypeValues()` function
  - [ ] Add to `Tags()` function mapping

### Phase 2: Database Layer Enhancements ✅ PLANNED

#### 2.1 Extend Membership Domain in `store/db/membership.go`
**Analysis: Current model is comprehensive, needs minimal extensions**

**Tasks:**
- [ ] **Add missing database methods (after existing methods, around line 186):**
  - [ ] Add `FindAll(ctx context.Context, filter M, page, limit int64) ([]*MembershipDomain, int64, error)` for admin listing
  - [ ] Add `FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*MembershipDomain, error)` for status-based queries
  - [ ] Add `FindByMembershipType(ctx context.Context, membershipType string, tenantID enum.Tenant) ([]*MembershipDomain, error)` for type-based queries
  - [ ] Add `UpdateExpirationDate(ctx context.Context, id string, newDate time.Time) error` for renewal operations
  - [ ] Add `UpdatePaymentStatus(ctx context.Context, id string, paymentStatus string) error` for manual payment processing
  - [ ] Add `GetExpiringMemberships(ctx context.Context, days int, tenantID enum.Tenant) ([]*MembershipDomain, error)` for notifications
  - [ ] Add `GetMembershipStatistics(ctx context.Context, tenantID enum.Tenant) (map[string]interface{}, error)` for admin dashboard

#### 2.2 Create Membership DTOs in `store/db/membership.go`
**Location: After existing domain, around line 32**

**Tasks:**
- [ ] **Create MembershipCreateData struct:**
  - [ ] Add `type MembershipCreateData struct {`
  - [ ] Add `MemberID string` with validation `validate:"required,len=24"`
  - [ ] Add `MembershipType string` with validation `validate:"required,oneof=basic premium vip"`
  - [ ] Add `StartDate string` with validation `validate:"required"` // Date format YYYY-MM-DD
  - [ ] Add `DurationMonths int` with validation `validate:"required,min=1,max=24"`
  - [ ] Add `AllocatedSlots int` with validation `validate:"required,min=0,max=10"`
  - [ ] Add `PaymentAmount float64` with validation `validate:"required,min=0"`
  - [ ] Add `Notes string` with validation `validate:"omitempty,max=500"`
  - [ ] Add closing `}`

- [ ] **Create MembershipUpdateData struct:**
  - [ ] Add `type MembershipUpdateData struct {`
  - [ ] Add `Status string` with validation `validate:"omitempty,oneof=pending_payment active expired canceled suspended"`
  - [ ] Add `PaymentStatus string` with validation `validate:"omitempty,oneof=pending paid failed"`
  - [ ] Add `Notes string` with validation `validate:"omitempty,max=500"`
  - [ ] Add `AutoRenew *bool` with validation `validate:"omitempty"`
  - [ ] Add closing `}`

- [ ] **Create MembershipRenewalData struct:**
  - [ ] Add `type MembershipRenewalData struct {`
  - [ ] Add `DurationMonths int` with validation `validate:"required,min=1,max=24"`
  - [ ] Add `PaymentAmount float64` with validation `validate:"required,min=0"`
  - [ ] Add `KeepSlots bool` with validation `validate:"omitempty"` // Keep current slot allocation
  - [ ] Add `Notes string` with validation `validate:"omitempty,max=500"`
  - [ ] Add closing `}`

- [ ] **Create MembershipListDto struct:**
  - [ ] Add `type MembershipListDto struct {`
  - [ ] Add embedded `MembershipDomain`
  - [ ] Add `MemberName string` with tag `json:"member_name"`
  - [ ] Add `MemberEmail string` with tag `json:"member_email"`
  - [ ] Add `DaysUntilExpiration int` with tag `json:"days_until_expiration"`
  - [ ] Add `AvailableSlots int` with tag `json:"available_slots"` // AllocatedSlots - UsedSlots
  - [ ] Add closing `}`

### Phase 3: Service Layer Implementation ✅ PLANNED

#### 3.1 Create new file `pkg/service/membership.go`
**Following existing service patterns from KYC and profile implementations**

**Tasks:**
- [ ] **File setup:**
  - [ ] Add package declaration: `package service`
  - [ ] Add imports: `context`, `time`, `fmt`, error packages
  - [ ] Add imports: `app/store/db`, `app/pkg/ecode`, `app/pkg/enum`

- [ ] **Create MembershipService struct:**
  - [ ] Add `type MembershipService struct {`
  - [ ] Add `store *db.Db` field
  - [ ] Add closing `}`
  - [ ] Add constructor: `func NewMembershipService(store *db.Db) *MembershipService`

- [ ] **Implement Core Service Methods:**
  - [ ] **CreateMembership method:**
    - [ ] Function: `func (s *MembershipService) CreateMembership(ctx context.Context, tenantID enum.Tenant, data *db.MembershipCreateData) (*db.MembershipDomain, error)`
    - [ ] Validate member exists and has verified KYC
    - [ ] Check for existing active membership
    - [ ] Parse start date and calculate expiration date
    - [ ] Create membership with "pending_payment" status
    - [ ] Update member's CurrentMembershipID
    - [ ] Return created membership

  - [ ] **RenewMembership method:**
    - [ ] Function: `func (s *MembershipService) RenewMembership(ctx context.Context, membershipID string, data *db.MembershipRenewalData) (*db.MembershipDomain, error)`
    - [ ] Validate existing membership exists
    - [ ] Calculate new expiration date from current expiration
    - [ ] Update membership details
    - [ ] Maintain slot allocation if requested
    - [ ] Return updated membership

  - [ ] **UpdateMembershipStatus method:**
    - [ ] Function: `func (s *MembershipService) UpdateMembershipStatus(ctx context.Context, membershipID string, data *db.MembershipUpdateData) error`
    - [ ] Validate status transitions (pending -> active, active -> expired, etc.)
    - [ ] Update membership status and payment status
    - [ ] Handle member's CurrentMembershipID updates
    - [ ] Trigger status change notifications

  - [ ] **GetMembershipWithDetails method:**
    - [ ] Function: `func (s *MembershipService) GetMembershipWithDetails(ctx context.Context, membershipID string) (*db.MembershipListDto, error)`
    - [ ] Fetch membership and member details
    - [ ] Calculate days until expiration
    - [ ] Calculate available slots
    - [ ] Return enriched DTO

### Phase 4: API Layer Implementation ✅ PLANNED

#### 4.1 Create new file `route/membership.go`
**Following exact patterns from `route/kyc.go` and `route/profile.go`**

**Tasks:**
- [ ] **File setup (copy structure from route/kyc.go):**
  - [ ] Copy import block from kyc.go exactly
  - [ ] Add service import: `"app/pkg/service"`
  - [ ] Create `type membership struct { *middleware }` (following kyc struct pattern)
  - [ ] Add init() function with handlers registration following kyc.go pattern

#### 4.2 Member-Facing Endpoints
**Location: Following patterns from kyc.go and profile.go**

**Tasks:**
- [ ] **Register member routes in init() function:**
  - [ ] Add line: `v1 := r.Group("/membership/v1")`
  - [ ] Add line: `v1.GET("/current", s.BearerAuth(enum.PermissionMembershipView), s.v1_GetCurrentMembership())`
  - [ ] Add line: `v1.GET("/history", s.BearerAuth(enum.PermissionMembershipView), s.v1_GetMembershipHistory())`

- [ ] **Implement GET `/membership/v1/current` (following kyc.go v1_GetStatus pattern):**
  - [ ] Function signature: `func (s membership) v1_GetCurrentMembership() gin.HandlerFunc {`
  - [ ] Get session: `session := s.Session(c)`
  - [ ] Find member: `member, err := s.store.Db.Member.FindByUserID(c.Request.Context(), session.UserId)`
  - [ ] Get current membership if exists
  - [ ] Return membership details with status
  - [ ] Handle case where no active membership exists
  - [ ] Use exact error handling pattern from kyc.go

- [ ] **Implement GET `/membership/v1/history` (following profile.go patterns):**
  - [ ] Function signature: `func (s membership) v1_GetMembershipHistory() gin.HandlerFunc {`
  - [ ] Get session and find member
  - [ ] Get all memberships for member sorted by creation date
  - [ ] Return membership history array
  - [ ] Use exact error handling from profile.go

#### 4.3 Admin/CMS Endpoints
**Location: Following patterns from kyc.go admin endpoints**

**Tasks:**
- [ ] **Register admin routes in init() function:**
  - [ ] Add line: `admin := v1.Group("/admin")`
  - [ ] Add line: `admin.GET("/memberships", s.BearerAuth(enum.PermissionMembershipView), s.admin_ListMemberships())`
  - [ ] Add line: `admin.POST("/memberships", s.BearerAuth(enum.PermissionMembershipCreate), s.admin_CreateMembership())`
  - [ ] Add line: `admin.GET("/memberships/:id", s.BearerAuth(enum.PermissionMembershipView), s.admin_GetMembership())`
  - [ ] Add line: `admin.PUT("/memberships/:id", s.BearerAuth(enum.PermissionMembershipUpdate), s.admin_UpdateMembership())`
  - [ ] Add line: `admin.POST("/memberships/:id/renew", s.BearerAuth(enum.PermissionMembershipRenew), s.admin_RenewMembership())`
  - [ ] Add line: `admin.DELETE("/memberships/:id", s.BearerAuth(enum.PermissionMembershipDelete), s.admin_DeleteMembership())`
  - [ ] Add line: `admin.GET("/memberships/expiring", s.BearerAuth(enum.PermissionMembershipView), s.admin_GetExpiringMemberships())`
  - [ ] Add line: `admin.GET("/memberships/statistics", s.BearerAuth(enum.PermissionMembershipView), s.admin_GetMembershipStatistics())`

- [ ] **Implement admin_ListMemberships (following kyc.go admin listing pattern):**
  - [ ] Function signature: `func (s membership) admin_ListMemberships() gin.HandlerFunc {`
  - [ ] Parse query parameters: page, limit, status, membership_type
  - [ ] Get tenant from session
  - [ ] Call membership service with filters
  - [ ] Return paginated results with member details
  - [ ] Use exact pagination pattern from kyc.go

- [ ] **Implement admin_CreateMembership (following kyc.go admin creation pattern):**
  - [ ] Function signature: `func (s membership) admin_CreateMembership() gin.HandlerFunc {`
  - [ ] Parse request body using ShouldBind exactly like kyc.go
  - [ ] Validate using membership service
  - [ ] Create membership through service layer
  - [ ] Create audit log using s.AuditLog exactly like kyc.go
  - [ ] Return created membership JSON

- [ ] **Implement admin_UpdateMembership (following profile.go update pattern):**
  - [ ] Function signature: `func (s membership) admin_UpdateMembership() gin.HandlerFunc {`
  - [ ] Get membership ID from URL params
  - [ ] Parse update data from request body
  - [ ] Update through membership service
  - [ ] Create audit log for the update
  - [ ] Return updated membership details

- [ ] **Implement admin_RenewMembership (new pattern):**
  - [ ] Function signature: `func (s membership) admin_RenewMembership() gin.HandlerFunc {`
  - [ ] Get membership ID from URL params
  - [ ] Parse renewal data from request body
  - [ ] Process renewal through service layer
  - [ ] Create audit log for renewal action
  - [ ] Return renewed membership details

### Phase 5: Integration Layer ✅ PLANNED

#### 5.1 Enhance Member Model Integration
**Location: `store/db/member.go`**

**Tasks:**
- [ ] **Add membership helper methods (after existing methods, around line 720):**
  - [ ] Add `UpdateCurrentMembership(ctx context.Context, memberID, membershipID string) error`
  - [ ] Add `ClearCurrentMembership(ctx context.Context, memberID string) error`
  - [ ] Add `GetMembersWithExpiredMemberships(ctx context.Context, tenantID enum.Tenant) ([]*MemberDomain, error)`
  - [ ] Add `GetMembersWithoutActiveMembership(ctx context.Context, tenantID enum.Tenant) ([]*MemberDomain, error)`

#### 5.2 Notification Integration
**Location: Create new file `pkg/notification/membership.go`**

**Tasks:**
- [ ] **Create membership notification service:**
  - [ ] Add membership expiration notifications
  - [ ] Add membership activation notifications
  - [ ] Add renewal reminder notifications
  - [ ] Integration with existing notification system

#### 5.3 Cache Integration
**Location: Enhance existing cache patterns**

**Tasks:**
- [ ] **Add membership caching (following existing Redis patterns):**
  - [ ] Cache current membership details
  - [ ] Cache membership lists for admin dashboard
  - [ ] Invalidate cache on membership updates
  - [ ] Cache expiring memberships for notifications

### Phase 6: Testing Implementation ✅ PLANNED

#### 6.1 Create `route/membership_test.go`
**Following exact patterns from `route/kyc_test.go` - All tests consolidated in route layer**

**Tasks:**
- [ ] **File setup (copy structure from kyc_test.go):**
  - [ ] Copy test setup patterns from kyc_test.go exactly
  - [ ] Create mock membership service following KYC patterns
  - [ ] Set up test database and cleanup functions
  - [ ] Use same mock OAuth, Storage, and Email services from KYC tests

- [ ] **Member endpoint tests (following kyc_test.go patterns):**
  - [ ] Test `GET /membership/v1/current` with and without active membership
  - [ ] Test `GET /membership/v1/history` with pagination
  - [ ] Test authentication and permission requirements
  - [ ] Test error cases and edge conditions

- [ ] **Admin endpoint tests (following kyc_test.go admin patterns):**
  - [ ] Test membership creation with valid and invalid data
  - [ ] Test membership listing with filters and pagination
  - [ ] Test membership updates and status changes
  - [ ] Test renewal functionality
  - [ ] Test permission validation for each endpoint
  - [ ] Test audit logging for all admin operations

- [ ] **Service layer business logic tests (within route tests):**
  - [ ] Test membership creation business logic
  - [ ] Test renewal calculations and validations
  - [ ] Test status transition validation
  - [ ] Test edge cases and error conditions
  - [ ] Test integration with existing Member and User models

- [ ] **Security and integration tests:**
  - [ ] Test tenant isolation across all endpoints
  - [ ] Test KYC status validation for membership operations
  - [ ] Test member authentication and session handling
  - [ ] Test cache invalidation on membership updates

### Phase 7: Documentation & Configuration ✅ PLANNED

#### 7.1 API Documentation
**Location: Following existing Swagger patterns**

**Tasks:**
- [ ] **Add Swagger annotations to all endpoints:**
  - [ ] Document request/response schemas
  - [ ] Document error responses
  - [ ] Document permission requirements
  - [ ] Add examples for each endpoint

#### 7.2 Database Documentation
**Location: Update architecture documentation**

**Tasks:**
- [ ] **Document membership data model:**
  - [ ] Update relationship diagrams
  - [ ] Document status flows
  - [ ] Document business rules

#### 7.3 Configuration Enhancement
**Location: Environment configuration**

**Tasks:**
- [ ] **Add membership-related configuration:**
  - [ ] Default membership types and pricing
  - [ ] Maximum slot allocations per type
  - [ ] Renewal grace period settings
  - [ ] Notification timing configuration

## Validation Checklist

### Core Functionality Validation
- [ ] **Membership Creation:**
  - [ ] Admin can create memberships for verified members
  - [ ] Proper validation of member KYC status
  - [ ] Correct calculation of expiration dates
  - [ ] Proper slot allocation based on membership type

- [ ] **Membership Management:**
  - [ ] Status updates work correctly
  - [ ] Payment status tracking functions
  - [ ] Member CurrentMembershipID updates properly
  - [ ] Audit trails are created for all operations

- [ ] **Membership Renewal:**
  - [ ] Proper extension of expiration dates
  - [ ] Slot allocation preservation option works
  - [ ] Grace period handling for expired memberships
  - [ ] Notification triggers for expiring memberships

### Security Validation
- [ ] **Permission System:**
  - [ ] All endpoints require proper authentication
  - [ ] Permission checks prevent unauthorized access
  - [ ] Tenant isolation maintained across all operations
  - [ ] Audit logging captures all administrative actions

- [ ] **Data Validation:**
  - [ ] Input validation prevents invalid data entry
  - [ ] Business rule validation enforced
  - [ ] Error messages are informative but secure
  - [ ] No sensitive data exposure in API responses

### Integration Validation
- [ ] **Member System Integration:**
  - [ ] Member KYC status properly validated
  - [ ] CurrentMembershipID correctly maintained
  - [ ] Member dashboard shows accurate membership info
  - [ ] Member can view membership history

- [ ] **Database Consistency:**
  - [ ] Foreign key relationships maintained
  - [ ] No orphaned records created
  - [ ] Transaction integrity maintained
  - [ ] Cache invalidation works properly

## Success Criteria

### Technical Success Metrics
- [ ] **Code Quality:**
  - [ ] All code passes linting (golangci-lint)
  - [ ] No compilation errors
  - [ ] Test coverage > 80% for all endpoint tests
  - [ ] All API endpoints documented with Swagger

- [ ] **Performance:**
  - [ ] Membership listing loads < 500ms for 1000 records
  - [ ] Membership creation completes < 200ms
  - [ ] Cache hit ratio > 85% for membership queries
  - [ ] Database queries optimized with proper indexes

### Business Success Metrics
- [ ] **Administrative Efficiency:**
  - [ ] Admin can create membership in < 30 seconds
  - [ ] Bulk operations support for membership management
  - [ ] Comprehensive reporting and statistics available
  - [ ] Automated expiration notifications working

- [ ] **Member Experience:**
  - [ ] Members can view current membership status
  - [ ] Membership history easily accessible
  - [ ] Clear status indicators and expiration dates
  - [ ] Proper error messages for membership issues

## Implementation Timeline

### Week 1: Foundation (Phases 1-2)
- [ ] Day 1-2: Permission system enhancement and enum definitions
- [ ] Day 3-4: Database layer enhancements and DTO creation
- [ ] Day 5: Service layer structure and core business logic

### Week 2: Core Implementation (Phases 3-4)
- [ ] Day 1-3: Complete service layer implementation
- [ ] Day 4-5: API layer implementation (member and admin endpoints)

### Week 3: Integration & Testing (Phases 5-6)
- [ ] Day 1-2: Integration layer and notification system
- [ ] Day 3-4: Comprehensive testing implementation following KYC patterns
- [ ] Day 5: Bug fixes and performance optimization

### Week 4: Documentation & Deployment (Phase 7)
- [ ] Day 1-2: Documentation completion and configuration
- [ ] Day 3-4: Final validation and security review
- [ ] Day 5: Production readiness assessment

## Dependencies

### Prerequisites (Must be completed)
- [x] Task 1.1: Core Infrastructure Setup
- [x] Task 1.2: Authentication & Authorization
- [x] Task 1.3: Member Management
- [x] Task 1.4: eKYC Integration

### External Dependencies
- [ ] Admin dashboard UI for membership management (separate frontend task)
- [ ] Email templates for membership notifications
- [ ] Documentation site updates for new API endpoints
- [ ] Monitoring and alerting setup for membership operations

## Risk Mitigation

### Technical Risks
- **Database Performance**: Implement proper indexing and caching strategies
- **Data Consistency**: Use database transactions for multi-step operations
- **API Security**: Comprehensive permission testing and audit logging
- **Cache Invalidation**: Clear cache invalidation strategy for membership updates

### Business Risks
- **Manual Payment Processing**: Clear workflow documentation for admin operations
- **Membership Conflicts**: Validation rules to prevent duplicate active memberships
- **Data Migration**: Plan for migrating any existing membership data
- **User Training**: Documentation and training for administrative staff

## Next Steps

### Phase 2 Preparations
- [ ] **Integration with Plant Slot Management (Task 1.6):**
  - [ ] Membership slot allocation validation
  - [ ] Slot availability checking based on membership type
  - [ ] Automated slot assignment on membership activation

- [ ] **Advanced Features Planning:**
  - [ ] Membership transfer between members
  - [ ] Membership upgrade/downgrade functionality
  - [ ] Bulk membership operations
  - [ ] Advanced reporting and analytics

### Monitoring and Maintenance
- [ ] **Operational Monitoring:**
  - [ ] Membership expiration alerts
  - [ ] Payment status monitoring
  - [ ] System health checks for membership services
  - [ ] Performance monitoring for admin operations

## Notes

### Implementation Approach
- **Code Reuse**: Maximum utilization of existing patterns from KYC and profile systems
- **Manual Processing**: Payment processing handled through admin interface without automated payment systems
- **Security First**: Comprehensive permission system and audit logging for all operations
- **Testing Strategy**: TDD approach with comprehensive testing consolidated in route layer following KYC patterns
- **Documentation**: Complete API documentation and business process documentation

### Architectural Decisions
- **Service Layer**: Dedicated membership service for business logic isolation
- **DTO Pattern**: Clear separation between domain models and API data transfer objects
- **Audit Trail**: Complete audit logging for all membership operations
- **Cache Strategy**: Redis caching for frequently accessed membership data
- **Error Handling**: Comprehensive error codes and user-friendly error messages
- **Testing Pattern**: All tests consolidated in route layer, following established KYC testing patterns 