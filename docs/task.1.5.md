# Task 1.5: Membership Management System - Detailed Implementation Checklist

## Overview

This task implements the membership purchase flow, renewal system, and membership status management for the Seed eG Cannabis Club Platform. The implementation builds upon the already existing membership database model, payment infrastructure, and member management system, following strict TDD principles and architectural patterns established in the codebase.

**Note**: Stripe integration will be implemented in a later phase (before NFT integration). Initial implementation focuses on core membership management functionality.

## Pre-Implementation Analysis ✅ COMPLETED

### ✅ Already Implemented Components Verified

- [x] **Database Layer (store/db)**:
  - [x] `MembershipDomain` struct with full schema (`store/db/membership.go`)
  - [x] Complete membership repository with all CRUD operations
  - [x] `PaymentDomain` struct with payment-ready fields (`store/db/payment.go`)
  - [x] Member integration with `current_membership_id` field
  - [x] All required indexes and database collections initialized

- [x] **Authentication & Authorization**:
  - [x] `RequireMembershipStatus()` middleware in `route/index.go`
  - [x] Membership status validation in JWT tokens
  - [x] Member verification functions in `store/db/member.go`

- [x] **Core Infrastructure**:
  - [x] Error handling with `pkg/ecode`
  - [x] Validation system with `pkg/validate`
  - [x] Email notification system with `pkg/mail`
  - [x] Audit logging system
  - [x] Redis caching patterns
  - [x] File storage with MinIO

## Phase 1: Enhanced Permission System & Error Handling

### Step 1.1: Add Membership Permissions to Enum

- [ ] **Update Permission System (`pkg/enum/index.go`)**
  - [ ] Add membership-specific permissions to constants:
    ```go
    // Membership permissions (add to existing Permission constants)
    PermissionMembershipView   Permission = "membership_view"
    PermissionMembershipCreate Permission = "membership_create"
    PermissionMembershipUpdate Permission = "membership_update"
    PermissionMembershipCancel Permission = "membership_cancel"
    PermissionMembershipRenew  Permission = "membership_renew"
    
    // Payment permissions  
    PermissionPaymentView    Permission = "payment_view"
    PermissionPaymentProcess Permission = "payment_process"
    PermissionPaymentRefund  Permission = "payment_refund"
    ```
  - [ ] Update `PermissionTenantValues()` function to include new permissions
  - [ ] Update `PermissionRootValues()` function to include new permissions
  - [ ] Run tests to verify enum updates work correctly

### Step 1.2: Create Error Code Extensions

- [ ] **Add Membership-Specific Error Codes (`pkg/ecode/index.go`)**
  - [ ] Add membership-related error codes:
    ```go
    // Membership errors (add to existing errors)
    var (
        MembershipRequired         = New(http.StatusForbidden, "membership_required")
        MembershipExpired         = New(http.StatusForbidden, "membership_expired")
        MembershipAlreadyActive   = New(http.StatusConflict, "membership_already_active")
        MembershipNotFound        = New(http.StatusNotFound, "membership_not_found")
        MembershipPaymentPending  = New(http.StatusPaymentRequired, "membership_payment_pending")
        MembershipInvalidType     = New(http.StatusBadRequest, "membership_invalid_type")
        MembershipNotCancellable  = New(http.StatusBadRequest, "membership_not_cancellable")
        MembershipNotRenewable    = New(http.StatusBadRequest, "membership_not_renewable")
        MembershipSlotLimitExceeded = New(http.StatusBadRequest, "membership_slot_limit_exceeded")
        
        // Payment errors
        PaymentProcessingError    = New(http.StatusInternalServerError, "payment_processing_error")
        PaymentIntentNotFound     = New(http.StatusNotFound, "payment_intent_not_found")
        PaymentWebhookError       = New(http.StatusBadRequest, "payment_webhook_error")
        PaymentConfigurationError = New(http.StatusInternalServerError, "payment_configuration_error")
    )
    ```

## Phase 2: Membership Service Layer (TDD Implementation)

### Step 2.1: Create Comprehensive Membership Service Tests

- [ ] **Create Test File (`pkg/service/membership_test.go`)**
  - [ ] Setup test structure following existing service test patterns
  - [ ] Create mock store, payment service, and mail service
  - [ ] Implement test helper functions

- [ ] **Test Suite 1: CreateMembership Tests (12 test cases)**
  - [ ] **Test: CreateMembership_Success**
    - [ ] Valid member with verified KYC
    - [ ] Standard membership type
    - [ ] Verify membership saved with pending_payment status
    - [ ] Assert email notification sent
    - [ ] Verify audit log entry created

  - [ ] **Test: CreateMembership_Success_Premium**
    - [ ] Valid member with verified KYC
    - [ ] Premium membership type with higher slot allocation
    - [ ] Verify correct slot allocation based on type

  - [ ] **Test: CreateMembership_KYCNotVerified**
    - [ ] Member with "pending" KYC status
    - [ ] Verify 403 error returned
    - [ ] Assert no membership created

  - [ ] **Test: CreateMembership_KYCRejected**
    - [ ] Member with "rejected" KYC status
    - [ ] Verify appropriate error returned

  - [ ] **Test: CreateMembership_MemberNotFound**
    - [ ] Non-existent member ID
    - [ ] Verify 404 error returned

  - [ ] **Test: CreateMembership_ActiveMembershipExists**
    - [ ] Member already has active membership
    - [ ] Verify conflict error returned

  - [ ] **Test: CreateMembership_InvalidMembershipType**
    - [ ] Invalid membership type provided
    - [ ] Verify validation error

  - [ ] **Test: CreateMembership_InvalidSlotCount**
    - [ ] Slot count exceeding limits
    - [ ] Zero or negative slot count
    - [ ] Verify validation errors

  - [ ] **Test: CreateMembership_InvalidPaymentAmount**
    - [ ] Zero payment amount
    - [ ] Negative payment amount
    - [ ] Verify validation errors

  - [ ] **Test: CreateMembership_DatabaseError**
    - [ ] Mock database save failure
    - [ ] Verify error handling and rollback

  - [ ] **Test: CreateMembership_EmailFailure**
    - [ ] Mock email service failure
    - [ ] Verify membership still created but warning logged

  - [ ] **Test: CreateMembership_ConcurrentRequest**
    - [ ] Multiple simultaneous requests for same member
    - [ ] Verify only one membership created

- [ ] **Test Suite 2: GetMembershipDetails Tests (8 test cases)**
  - [ ] **Test: GetMembershipDetails_Success**
    - [ ] Valid membership ID and owner
    - [ ] Verify complete membership data returned

  - [ ] **Test: GetMembershipDetails_WithPaymentHistory**
    - [ ] Membership with payment records
    - [ ] Verify payment history included

  - [ ] **Test: GetMembershipDetails_NotFound**
    - [ ] Non-existent membership ID
    - [ ] Verify 404 error

  - [ ] **Test: GetMembershipDetails_AccessDenied**
    - [ ] Member trying to access another's membership
    - [ ] Verify 403 error

  - [ ] **Test: GetMembershipDetails_AdminAccess**
    - [ ] Admin accessing any membership
    - [ ] Verify access granted with proper permissions

  - [ ] **Test: GetMembershipDetails_ExpiredMembership**
    - [ ] Accessing expired membership details
    - [ ] Verify status correctly shown

  - [ ] **Test: GetMembershipDetails_CancelledMembership**
    - [ ] Accessing cancelled membership details
    - [ ] Verify cancellation details included

  - [ ] **Test: GetMembershipDetails_InvalidID**
    - [ ] Invalid membership ID format
    - [ ] Verify validation error

- [ ] **Test Suite 3: ListMemberMemberships Tests (6 test cases)**
  - [ ] **Test: ListMemberMemberships_Success**
    - [ ] Member with multiple memberships
    - [ ] Verify all memberships returned, sorted by date

  - [ ] **Test: ListMemberMemberships_Empty**
    - [ ] Member with no memberships
    - [ ] Verify empty list returned

  - [ ] **Test: ListMemberMemberships_Pagination**
    - [ ] Member with many memberships
    - [ ] Test pagination parameters
    - [ ] Verify correct page returned

  - [ ] **Test: ListMemberMemberships_FilterByStatus**
    - [ ] Filter by active/expired/cancelled status
    - [ ] Verify filtering works correctly

  - [ ] **Test: ListMemberMemberships_AccessControl**
    - [ ] Verify member can only see own memberships
    - [ ] Admin can see filtered results

  - [ ] **Test: ListMemberMemberships_DatabaseError**
    - [ ] Mock database query failure
    - [ ] Verify error handling

- [ ] **Test Suite 4: RenewMembership Tests (10 test cases)**
  - [ ] **Test: RenewMembership_Success**
    - [ ] Expiring membership renewal
    - [ ] Verify new membership created with extended dates

  - [ ] **Test: RenewMembership_AutoRenewEnabled**
    - [ ] Membership with auto-renew enabled
    - [ ] Verify automatic renewal process

  - [ ] **Test: RenewMembership_NotExpiring**
    - [ ] Membership not close to expiration
    - [ ] Verify early renewal allowed

  - [ ] **Test: RenewMembership_AlreadyExpired**
    - [ ] Expired membership renewal
    - [ ] Verify grace period handling

  - [ ] **Test: RenewMembership_NotFound**
    - [ ] Non-existent membership ID
    - [ ] Verify 404 error

  - [ ] **Test: RenewMembership_AccessDenied**
    - [ ] Member trying to renew another's membership
    - [ ] Verify 403 error

  - [ ] **Test: RenewMembership_CancelledMembership**
    - [ ] Attempting to renew cancelled membership
    - [ ] Verify appropriate error

  - [ ] **Test: RenewMembership_PaymentFailure**
    - [ ] Mock payment processing failure
    - [ ] Verify membership remains in original state

  - [ ] **Test: RenewMembership_EmailNotification**
    - [ ] Verify renewal confirmation email sent
    - [ ] Test email failure handling

  - [ ] **Test: RenewMembership_SlotTransfer**
    - [ ] Verify plant slots transferred to new membership
    - [ ] Test slot allocation updates

- [ ] **Test Suite 5: CancelMembership Tests (8 test cases)**
  - [ ] **Test: CancelMembership_Success**
    - [ ] Active membership cancellation
    - [ ] Verify status updated to cancelled

  - [ ] **Test: CancelMembership_WithReason**
    - [ ] Cancellation with reason provided
    - [ ] Verify reason stored

  - [ ] **Test: CancelMembership_RefundEligible**
    - [ ] Cancellation eligible for refund
    - [ ] Verify refund amount calculated

  - [ ] **Test: CancelMembership_NotRefundable**
    - [ ] Cancellation past refund period
    - [ ] Verify no refund processed

  - [ ] **Test: CancelMembership_AlreadyCancelled**
    - [ ] Attempting to cancel already cancelled membership
    - [ ] Verify appropriate error

  - [ ] **Test: CancelMembership_HasActivePlants**
    - [ ] Membership with active plant slots
    - [ ] Verify warning about plant slot loss

  - [ ] **Test: CancelMembership_AccessDenied**
    - [ ] Member trying to cancel another's membership
    - [ ] Verify 403 error

  - [ ] **Test: CancelMembership_EmailNotification**
    - [ ] Verify cancellation confirmation email
    - [ ] Include refund information if applicable

### Step 2.2: Create Membership Service Implementation

- [ ] **Create Service Package Directory**
  - [ ] Create `pkg/service/` directory if it doesn't exist
  - [ ] Follow existing package organization patterns

- [ ] **Create Membership Service (`pkg/service/membership.go`)**
  - [ ] Implement service struct following existing patterns:
    ```go
    package service
    
    import (
        "app/pkg/ecode"
        "app/pkg/mail"
        "app/store"
        "app/store/db"
        "context"
        "fmt"
        "net/http"
        "time"
        
        "github.com/nhnghia272/gopkg"
    )
    
    type MembershipService struct {
        store *store.Store
        mail  *mail.Mail
    }
    
    func NewMembershipService(store *store.Store) *MembershipService {
        return &MembershipService{
            store: store,
            mail:  mail.New(store),
        }
    }
    ```

- [ ] **Implement CreateMembership Method**
  - [ ] Add comprehensive input validation
  - [ ] Check member exists and KYC is verified
  - [ ] Verify no active membership exists
  - [ ] Create membership with pending_payment status
  - [ ] Send confirmation email
  - [ ] Handle all error cases with proper error codes

- [ ] **Implement GetMembershipDetails Method**
  - [ ] Retrieve membership by ID
  - [ ] Verify member ownership or admin permission
  - [ ] Return detailed membership information
  - [ ] Include payment history and status

- [ ] **Implement ListMemberMemberships Method**
  - [ ] Get all memberships for authenticated member
  - [ ] Sort by creation date (newest first)
  - [ ] Include basic membership info and status
  - [ ] Support pagination

- [ ] **Implement RenewMembership Method**
  - [ ] Find existing membership by ID
  - [ ] Validate membership belongs to authenticated member
  - [ ] Check membership is eligible for renewal
  - [ ] Create new membership with extended dates
  - [ ] Update membership status
  - [ ] Send renewal confirmation email

- [ ] **Implement CancelMembership Method**
  - [ ] Find membership by ID and verify ownership
  - [ ] Check membership is active and cancellable
  - [ ] Update membership status to "cancelled"
  - [ ] Calculate refund if applicable
  - [ ] Send cancellation confirmation email
  - [ ] Update member's current_membership_id

- [ ] **Run Membership Service Tests**
  - [ ] Execute all 44 membership service tests
  - [ ] Verify 100% test coverage
  - [ ] Fix any failing tests

## Phase 3: API Routes Implementation (TDD Driven)

### Step 3.1: Create Comprehensive Membership Route Tests

- [ ] **Create Test File (`route/membership_test.go`)**
  - [ ] Follow structure from `route/auth_test.go` and `route/kyc_test.go`
  - [ ] Setup test dependencies and mock services
  - [ ] Create helper functions for test data generation

- [ ] **Test Suite 1: Member API Endpoints (25 test cases)**

  **POST /api/v1/memberships Tests (8 cases):**
  - [ ] **Test: CreateMembership_Success**
    - [ ] Valid request with verified KYC member
    - [ ] Verify 201 Created response
    - [ ] Assert response contains membership details

  - [ ] **Test: CreateMembership_KYCNotVerified**
    - [ ] Request from member without verified KYC
    - [ ] Verify 403 Forbidden response

  - [ ] **Test: CreateMembership_InvalidPayload**
    - [ ] Missing required fields
    - [ ] Invalid membership type
    - [ ] Negative payment amount
    - [ ] Invalid slot count

  - [ ] **Test: CreateMembership_Unauthorized**
    - [ ] Request without authentication token
    - [ ] Invalid/expired token
    - [ ] Verify 401 Unauthorized response

  - [ ] **Test: CreateMembership_ActiveMembershipExists**
    - [ ] Member already has active membership
    - [ ] Verify 409 Conflict response

  - [ ] **Test: CreateMembership_ValidationErrors**
    - [ ] Test all field validation scenarios
    - [ ] Verify 400 Bad Request responses

  - [ ] **Test: CreateMembership_InternalError**
    - [ ] Mock service failure
    - [ ] Verify 500 error handling

  - [ ] **Test: CreateMembership_RateLimiting**
    - [ ] Multiple rapid requests
    - [ ] Verify rate limiting works

  **GET /api/v1/memberships Tests (5 cases):**
  - [ ] **Test: GetMyMemberships_Success**
    - [ ] Member with multiple memberships
    - [ ] Verify all memberships returned

  - [ ] **Test: GetMyMemberships_Empty**
    - [ ] Member with no memberships
    - [ ] Verify empty array returned

  - [ ] **Test: GetMyMemberships_Pagination**
    - [ ] Test pagination parameters
    - [ ] Verify pagination headers

  - [ ] **Test: GetMyMemberships_Unauthorized**
    - [ ] Request without token
    - [ ] Verify 401 response

  - [ ] **Test: GetMyMemberships_Sorting**
    - [ ] Verify memberships sorted by date
    - [ ] Test different sort options

  **GET /api/v1/memberships/:id Tests (6 cases):**
  - [ ] **Test: GetMembershipDetails_Success**
    - [ ] Valid membership ID owned by member
    - [ ] Verify complete details returned

  - [ ] **Test: GetMembershipDetails_NotFound**
    - [ ] Non-existent membership ID
    - [ ] Verify 404 response

  - [ ] **Test: GetMembershipDetails_AccessDenied**
    - [ ] Member trying to access another's membership
    - [ ] Verify 403 response

  - [ ] **Test: GetMembershipDetails_InvalidID**
    - [ ] Invalid ObjectID format
    - [ ] Verify 400 response

  - [ ] **Test: GetMembershipDetails_Unauthorized**
    - [ ] Request without token
    - [ ] Verify 401 response

  - [ ] **Test: GetMembershipDetails_WithPaymentHistory**
    - [ ] Membership with payment records
    - [ ] Verify payment data included

  **POST /api/v1/memberships/:id/renew Tests (3 cases):**
  - [ ] **Test: RenewMembership_Success**
    - [ ] Valid renewal request
    - [ ] Verify new membership created

  - [ ] **Test: RenewMembership_NotEligible**
    - [ ] Membership not eligible for renewal
    - [ ] Verify appropriate error

  - [ ] **Test: RenewMembership_AccessDenied**
    - [ ] Member trying to renew another's membership
    - [ ] Verify 403 response

  **POST /api/v1/memberships/:id/cancel Tests (3 cases):**
  - [ ] **Test: CancelMembership_Success**
    - [ ] Valid cancellation request
    - [ ] Verify status updated

  - [ ] **Test: CancelMembership_NotCancellable**
    - [ ] Already cancelled membership
    - [ ] Verify appropriate error

  - [ ] **Test: CancelMembership_AccessDenied**
    - [ ] Member trying to cancel another's membership
    - [ ] Verify 403 response

- [ ] **Test Suite 2: CMS Admin API Endpoints (15 test cases)**

  **GET /api/v1/cms/memberships Tests (5 cases):**
  - [ ] **Test: AdminListMemberships_Success**
    - [ ] Admin with proper permissions
    - [ ] Verify all memberships returned

  - [ ] **Test: AdminListMemberships_Filtering**
    - [ ] Filter by status, type, member
    - [ ] Verify filtering works

  - [ ] **Test: AdminListMemberships_Unauthorized**
    - [ ] Non-admin user
    - [ ] Verify 403 response

  - [ ] **Test: AdminListMemberships_Pagination**
    - [ ] Large dataset pagination
    - [ ] Verify pagination works

  - [ ] **Test: AdminListMemberships_Search**
    - [ ] Search by member name/email
    - [ ] Verify search functionality

  **GET /api/v1/cms/memberships/:id Tests (3 cases):**
  - [ ] **Test: AdminGetMembershipDetails_Success**
    - [ ] Admin viewing any membership
    - [ ] Verify access granted

  - [ ] **Test: AdminGetMembershipDetails_NotFound**
    - [ ] Non-existent membership
    - [ ] Verify 404 response

  - [ ] **Test: AdminGetMembershipDetails_Unauthorized**
    - [ ] Non-admin user
    - [ ] Verify 403 response

  **PUT /api/v1/cms/memberships/:id Tests (4 cases):**
  - [ ] **Test: AdminUpdateMembership_Success**
    - [ ] Valid update request
    - [ ] Verify membership updated

  - [ ] **Test: AdminUpdateMembership_InvalidData**
    - [ ] Invalid update data
    - [ ] Verify validation errors

  - [ ] **Test: AdminUpdateMembership_NotFound**
    - [ ] Non-existent membership
    - [ ] Verify 404 response

  - [ ] **Test: AdminUpdateMembership_Unauthorized**
    - [ ] Non-admin user
    - [ ] Verify 403 response

  **POST /api/v1/cms/memberships/:id/extend Tests (3 cases):**
  - [ ] **Test: AdminExtendMembership_Success**
    - [ ] Valid extension request
    - [ ] Verify expiration date extended

  - [ ] **Test: AdminExtendMembership_InvalidPeriod**
    - [ ] Invalid extension period
    - [ ] Verify validation error

  - [ ] **Test: AdminExtendMembership_Unauthorized**
    - [ ] Non-admin user
    - [ ] Verify 403 response

### Step 3.2: Create Membership Routes Implementation

- [ ] **Create Route File (`route/membership.go`)**
  - [ ] Follow structure from `route/kyc.go` and `route/profile.go`
  - [ ] Implement route registration in init() function
  - [ ] Setup route groups for member and CMS endpoints

- [ ] **Implement Route Registration**
  ```go
  func init() {
      handlers = append(handlers, func(m *middleware, r *gin.Engine) {
          // Member API routes - self-service
          v1 := r.Group("/api/v1/memberships")
          {
              v1.POST("", m.BearerAuth(enum.PermissionUserViewSelf), m.RequireKYCStatus("verified"), m.v1_CreateMembership())
              v1.GET("", m.BearerAuth(enum.PermissionUserViewSelf), m.v1_GetMyMemberships())
              v1.GET("/:membership_id", m.BearerAuth(enum.PermissionUserViewSelf), m.v1_GetMembershipDetails())
              v1.POST("/:membership_id/renew", m.BearerAuth(enum.PermissionUserViewSelf), m.v1_RenewMembership())
              v1.POST("/:membership_id/cancel", m.BearerAuth(enum.PermissionUserViewSelf), m.v1_CancelMembership())
          }
  
          // CMS Admin routes - administrative management
          v1cms := r.Group("/api/v1/cms/memberships")
          {
              v1cms.GET("", m.BearerAuth(enum.PermissionMembershipView), m.v1cms_ListMemberships())
              v1cms.GET("/:membership_id", m.BearerAuth(enum.PermissionMembershipView), m.v1cms_GetMembershipDetails())
              v1cms.PUT("/:membership_id", m.BearerAuth(enum.PermissionMembershipUpdate), m.v1cms_UpdateMembership())
              v1cms.POST("/:membership_id/extend", m.BearerAuth(enum.PermissionMembershipUpdate), m.v1cms_ExtendMembership())
          }
      })
  }
  ```

- [ ] **Implement Request/Response Structures**
  - [ ] Create `CreateMembershipRequest` with validation tags
  - [ ] Create `RenewMembershipRequest` struct
  - [ ] Create `CancelMembershipRequest` struct
  - [ ] Create response DTOs following existing patterns

- [ ] **Implement All Handler Methods (10 handlers)**
  - [ ] `v1_CreateMembership` - Member membership creation
  - [ ] `v1_GetMyMemberships` - Member membership listing
  - [ ] `v1_GetMembershipDetails` - Member membership details
  - [ ] `v1_RenewMembership` - Member membership renewal
  - [ ] `v1_CancelMembership` - Member membership cancellation
  - [ ] `v1cms_ListMemberships` - Admin membership listing
  - [ ] `v1cms_GetMembershipDetails` - Admin membership details
  - [ ] `v1cms_UpdateMembership` - Admin membership updates
  - [ ] `v1cms_ExtendMembership` - Admin membership extension

- [ ] **Add Swagger Documentation Comments**
  - [ ] Add comprehensive swagger comments to all handlers
  - [ ] Document request/response schemas
  - [ ] Include error response examples
  - [ ] Follow existing swagger documentation patterns

- [ ] **Run Membership Route Tests**
  - [ ] Execute all 40 membership route tests
  - [ ] Verify all endpoints work correctly
  - [ ] Test error handling and edge cases
  - [ ] Validate authentication and authorization

## Phase 4: Email Notification System

### Step 4.1: Create Email Templates

- [ ] **Membership Purchase Confirmation Email**
  - [ ] Create HTML template for purchase confirmation
  - [ ] Include membership details and payment information
  - [ ] Add next steps instructions
  - [ ] Follow existing email template structure

- [ ] **Membership Activation Email**
  - [ ] Create template for membership activation
  - [ ] Include membership details and benefits
  - [ ] Add account dashboard link
  - [ ] Include plant slot allocation information

- [ ] **Membership Renewal Reminder Email**
  - [ ] Create template for renewal reminders
  - [ ] Include expiration date and renewal options
  - [ ] Add auto-renewal status information
  - [ ] Include pricing and benefits

- [ ] **Membership Expiration Warning Email**
  - [ ] Create template for expiration warnings
  - [ ] Include grace period information
  - [ ] Add urgent renewal call-to-action
  - [ ] Include consequences of expiration

- [ ] **Membership Cancellation Confirmation Email**
  - [ ] Create template for cancellation confirmation
  - [ ] Include cancellation details and effective date
  - [ ] Add refund information if applicable
  - [ ] Include feedback survey link

### Step 4.2: Implement Email Service Extensions

- [ ] **Extend Mail Service (`pkg/mail/index.go`)**
  - [ ] Add membership-specific email methods
  - [ ] Follow existing email sending patterns
  - [ ] Implement template rendering for membership emails

- [ ] **Add Email Methods**
  ```go
  func (s *Mail) SendMembershipPurchaseConfirmation(ctx context.Context, member *db.MemberDomain, membership *db.MembershipDomain) error
  func (s *Mail) SendMembershipActivationNotification(ctx context.Context, member *db.MemberDomain, membership *db.MembershipDomain) error
  func (s *Mail) SendMembershipRenewalReminder(ctx context.Context, member *db.MemberDomain, membership *db.MembershipDomain) error
  func (s *Mail) SendMembershipExpirationWarning(ctx context.Context, member *db.MemberDomain, membership *db.MembershipDomain) error
  func (s *Mail) SendMembershipCancellationConfirmation(ctx context.Context, member *db.MemberDomain, membership *db.MembershipDomain) error
  ```

## Phase 5: Background Jobs & Automation

### Step 5.1: Create Membership Monitoring Jobs

- [ ] **Create Job Package (`pkg/jobs/membership.go`)**
  - [ ] Implement membership expiration monitoring
  - [ ] Create renewal reminder scheduling
  - [ ] Add automatic status updates

- [ ] **Implement Expiration Monitor Job**
  - [ ] Find memberships expiring in 30, 7, and 1 days
  - [ ] Send appropriate reminder emails
  - [ ] Update membership status for expired memberships
  - [ ] Handle grace period logic

- [ ] **Implement Auto-Renewal Job**
  - [ ] Find memberships eligible for auto-renewal
  - [ ] Create renewal entries
  - [ ] Handle auto-renewal notifications
  - [ ] Send renewal confirmation emails

- [ ] **Add Job Scheduling**
  - [ ] Setup cron job scheduling
  - [ ] Configure job execution intervals
  - [ ] Add job monitoring and logging

## Phase 6: API Documentation & Testing

### Step 6.1: Complete API Documentation

- [ ] **Update Swagger Specification (`docs/swagger.yaml`)**
  - [ ] Add all membership endpoints
  - [ ] Document request/response schemas
  - [ ] Include authentication requirements
  - [ ] Add error response examples

- [ ] **Document All Endpoints**
  - [ ] `POST /api/v1/memberships` - Create membership
  - [ ] `GET /api/v1/memberships` - List member memberships
  - [ ] `GET /api/v1/memberships/{id}` - Get membership details
  - [ ] `POST /api/v1/memberships/{id}/renew` - Renew membership
  - [ ] `POST /api/v1/memberships/{id}/cancel` - Cancel membership
  - [ ] All CMS admin endpoints

### Step 6.2: Integration Testing

- [ ] **Create Integration Test Suite**
  - [ ] Setup test environment with real database
  - [ ] Test complete membership management flow

- [ ] **Test Complete User Flows**
  - [ ] Member registration → KYC → Membership creation → Management
  - [ ] Membership renewal flow
  - [ ] Membership cancellation flow
  - [ ] Admin membership management flow

- [ ] **Performance Testing**
  - [ ] Test API endpoints under load
  - [ ] Verify database query performance
  - [ ] Validate email sending performance

- [ ] **Security Testing**
  - [ ] Verify authentication on all endpoints
  - [ ] Test permission enforcement
  - [ ] Validate input sanitization

## Final Validation Checklist

### Functional Testing ✅

- [ ] **Member Self-Service Functions**
  - [ ] Member can create new membership
  - [ ] Member can view own membership details
  - [ ] Member can renew expiring membership
  - [ ] Member can cancel active membership
  - [ ] Member cannot access other members' data

- [ ] **Email Notifications**
  - [ ] Purchase confirmation emails sent
  - [ ] Renewal reminder emails sent
  - [ ] Cancellation confirmation emails sent

- [ ] **Admin CMS Functions**
  - [ ] Admin can view all memberships
  - [ ] Admin can update membership details
  - [ ] Admin can extend memberships

### Technical Validation ✅

- [ ] **Code Quality**
  - [ ] All tests pass (100% success rate)
  - [ ] Code coverage > 90%
  - [ ] No linting errors
  - [ ] Performance benchmarks met

- [ ] **Security Validation**
  - [ ] All endpoints require proper authentication
  - [ ] Permission checks work correctly
  - [ ] Input validation prevents injection attacks

- [ ] **Architecture Compliance**
  - [ ] Follows existing database patterns
  - [ ] Uses established error handling
  - [ ] Implements proper caching
  - [ ] Maintains audit logging

## Future Phase: Payment Integration (Pre-NFT)

### Payment Integration Checklist (Future Implementation)
- [ ] Stripe SDK integration
- [ ] Payment intent creation
- [ ] Webhook handling
- [ ] Payment status updates
- [ ] Refund processing

**Note**: Payment integration will be implemented as a separate phase before NFT integration, allowing the core membership system to be functional and testable independently.

## Success Criteria Summary

### ✅ Functional Requirements Met
1. Complete membership management system
2. Member self-service operations
3. Admin CMS for membership oversight
4. Email notification system
5. Complete audit trail

### ✅ Technical Requirements Met
1. 100% test coverage (84+ comprehensive tests)
2. All existing architectural patterns followed
3. Performance benchmarks achieved
4. Security requirements satisfied
5. Complete API documentation

### ✅ Code Quality Requirements Met
1. Maximum reuse of existing components
2. Established coding patterns followed
3. Comprehensive error handling
4. Proper validation and sanitization
5. Complete logging and monitoring

## Implementation Timeline

**Total Estimated Duration: 2.5 weeks**

### Week 1: Foundation (Days 1-7)
- Days 1-2: Permission system and error handling
- Days 3-5: Membership service with 44 comprehensive tests
- Days 6-7: Service implementation and test validation

### Week 2: API & Integration (Days 8-14)
- Days 8-11: API routes with 40 comprehensive tests
- Days 12-14: Email system and background jobs

### Week 2.5: Validation & Documentation (Days 15-18)
- Days 15-16: Integration testing and bug fixes
- Days 17-18: Documentation and deployment preparation

This comprehensive checklist ensures Task 1.5 implements a complete membership management system with maximum code reuse, strict TDD adherence (84+ tests), and full architectural compliance, ready for future payment integration. 