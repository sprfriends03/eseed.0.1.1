# Task 1.4: eKYC Integration

## Status: 📋 PLANNING COMPLETE - AWAITING IMPLEMENTATION APPROVAL

## Objective
Implement electronic Know Your Customer (eKYC) verification system as an extension of the existing Member system, providing document upload, manual verification workflow, and administrative review capabilities.

## Implementation Strategy
**Maximize Code Reuse**: This implementation leverages existing infrastructure including storage buckets, error handling, authentication, and database patterns to minimize new code and maintain architectural consistency.

## Detailed Implementation Checklist

### Phase 1: Infrastructure Assessment ✅ COMPLETED
**Reusing Existing Components:**
- ✅ Storage: KYC documents bucket already configured (`"kyc-documents"`)
- ✅ Member Model: KYC status field exists (`KYCStatus *string`)
- ✅ Error Handling: KYC error codes exist (`KYCVerificationRequired`, `KYCRejected`, `KYCDocumentInvalid`)
- ✅ Authentication: JWT includes KYC status in `MemberAccessClaims.KYCStatus`
- ✅ Middleware: `RequireKYCStatus()` already implemented
- ✅ Validation: `VerifyMemberActive()` already checks KYC status
- ✅ Database: Indexes for KYC status already created

### Phase 2: Database Schema Extensions 📋 PLANNED

#### 2.1 Extend MemberDomain in `store/db/member.go`
- [ ] **Add KYCDocuments struct to MemberDomain (after line 28 Preferences struct):**
  - [ ] Add `KYCDocuments *struct {` field after Preferences struct
  - [ ] Add `    IDDocumentFront *string` with tags `json:"id_document_front,omitempty" bson:"id_document_front,omitempty"`
  - [ ] Add `    IDDocumentBack *string` with tags `json:"id_document_back,omitempty" bson:"id_document_back,omitempty"`
  - [ ] Add `    SelfiePhoto *string` with tags `json:"selfie_photo,omitempty" bson:"selfie_photo,omitempty"`
  - [ ] Add `    DocumentType *string` with tags `json:"document_type,omitempty" bson:"document_type,omitempty" validate:"omitempty,oneof=passport driver_license national_id"`
  - [ ] Add `    UploadedAt *time.Time` with tags `json:"uploaded_at,omitempty" bson:"uploaded_at,omitempty"`
  - [ ] Add `    SubmittedAt *time.Time` with tags `json:"submitted_at,omitempty" bson:"submitted_at,omitempty"`
  - [ ] Add closing `}` with tags `json:"kyc_documents,omitempty" bson:"kyc_documents,omitempty" validate:"omitempty"`

- [ ] **Add KYCVerification struct to MemberDomain (after KYCDocuments):**
  - [ ] Add `KYCVerification *struct {` field
  - [ ] Add `    VerifiedAt *time.Time` with tags `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
  - [ ] Add `    VerifiedBy *string` with tags `json:"verified_by,omitempty" bson:"verified_by,omitempty"`
  - [ ] Add `    RejectedAt *time.Time` with tags `json:"rejected_at,omitempty" bson:"rejected_at,omitempty"`
  - [ ] Add `    RejectedBy *string` with tags `json:"rejected_by,omitempty" bson:"rejected_by,omitempty"`
  - [ ] Add `    RejectionReason *string` with tags `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
  - [ ] Add `    Notes *string` with tags `json:"notes,omitempty" bson:"notes,omitempty"`
  - [ ] Add `    AttemptCount *int` with tags `json:"attempt_count,omitempty" bson:"attempt_count,omitempty"`
  - [ ] Add closing `}` with tags `json:"kyc_verification,omitempty" bson:"kyc_verification,omitempty" validate:"omitempty"`

#### 2.2 Create new file `store/db/kyc.go`
- [ ] **File setup:**
  - [ ] Add exact package declaration: `package db`
  - [ ] Add import block with: `"time"`, `"context"`, validation package
  - [ ] Add import: `"github.com/nhnghia272/gopkg"` for pointer utilities

- [ ] **Create KYCDocumentUploadData struct (following existing DTO patterns):**
  - [ ] Add `type KYCDocumentUploadData struct {`
  - [ ] Add `DocumentType string` with tag `json:"document_type" validate:"required,oneof=passport driver_license national_id" example:"passport"`
  - [ ] Add `FileType string` with tag `json:"file_type" validate:"required,oneof=front back selfie" example:"front"`
  - [ ] Add closing `}` 

- [ ] **Create KYCSubmissionData struct:**
  - [ ] Add `type KYCSubmissionData struct {`
  - [ ] Add `DocumentType string` with tag `json:"document_type" validate:"required,oneof=passport driver_license national_id" example:"passport"`
  - [ ] Add `HasAllDocuments bool` with tag `json:"has_all_documents" validate:"required" example:"true"`
  - [ ] Add `ConfirmAccuracy bool` with tag `json:"confirm_accuracy" validate:"required" example:"true"`
  - [ ] Add closing `}`

- [ ] **Create KYCVerificationData struct:**
  - [ ] Add `type KYCVerificationData struct {`
  - [ ] Add `Action string` with tag `json:"action" validate:"required,oneof=approve reject" example:"approve"`
  - [ ] Add `Notes string` with tag `json:"notes,omitempty" example:"Documents verified successfully"`
  - [ ] Add `Reason string` with tag `json:"reason,omitempty" example:"Document quality insufficient"`
  - [ ] Add closing `}`
  - [ ] Add method `func (v *KYCVerificationData) Validate() error` to check reason required for rejection

- [ ] **Create KYCStatusDto struct (following UserProfileDto pattern):**
  - [ ] Add `type KYCStatusDto struct {`
  - [ ] Add `Status string` with tag `json:"status" example:"pending_kyc"`
  - [ ] Add `SubmittedAt *time.Time` with tag `json:"submitted_at,omitempty"`
  - [ ] Add `VerifiedAt *time.Time` with tag `json:"verified_at,omitempty"`
  - [ ] Add `RejectedAt *time.Time` with tag `json:"rejected_at,omitempty"`
  - [ ] Add `RejectionReason string` with tag `json:"rejection_reason,omitempty"`
  - [ ] Add `AttemptCount int` with tag `json:"attempt_count" example:"1"`
  - [ ] Add `RequiredDocs []string` with tag `json:"required_documents"`
  - [ ] Add `UploadedDocs []string` with tag `json:"uploaded_documents"`
  - [ ] Add `CanSubmit bool` with tag `json:"can_submit" example:"false"`
  - [ ] Add closing `}`

- [ ] **Add helper functions:**
  - [ ] Create `func GetRequiredDocuments(docType string) []string` returning appropriate doc types
  - [ ] Create `func ValidateDocumentType(docType string) bool` for validation
  - [ ] Create `func CheckCompleteness(docs *KYCDocuments) bool` for submission readiness

### Phase 3: Storage Layer Extensions 📋 PLANNED

#### 3.1 Extend `store/storage/index.go` (following UploadProfileImage pattern exactly)
- [ ] **Add UploadKYCDocument method (after UploadProfileImage method, around line 240):**
  - [ ] Function signature: `func (s Storage) UploadKYCDocument(ctx context.Context, memberID, docType, fileType string, reader io.Reader, filename string) (string, error) {`
  - [ ] Generate secure filename: `objectName := fmt.Sprintf("%s/%s_%s_%d_%s", memberID, docType, fileType, time.Now().Unix(), filename)`
  - [ ] Read data: `data, err := io.ReadAll(reader)` (following exact profile image pattern)
  - [ ] Create new reader: `newReader := strings.NewReader(string(data))`
  - [ ] Get object size: `objectSize := int64(len(data))`
  - [ ] Determine content type from filename extension (copy logic from UploadProfileImage)
  - [ ] Call `_, err = s.UploadToTypeBucket(ctx, "kyc", objectName, newReader, objectSize, contentType)`
  - [ ] Return `objectName, err`

- [ ] **Add DeleteKYCDocument method (following DeleteProfileImage pattern exactly):**
  - [ ] Function signature: `func (s Storage) DeleteKYCDocument(ctx context.Context, objectName string) error {`
  - [ ] Get bucket: `bucket := s.GetBucketForType("kyc")`
  - [ ] Call: `return s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})`

- [ ] **Add GetKYCDocumentURL method (new, but following presigned URL patterns):**
  - [ ] Function signature: `func (s Storage) GetKYCDocumentURL(ctx context.Context, objectName string) (string, error) {`
  - [ ] Get bucket: `bucket := s.GetBucketForType("kyc")`
  - [ ] Call: `return s.client.PresignedGetObject(ctx, bucket, objectName, time.Hour*24, nil)`

- [ ] **Add file validation helper (following profile image validation logic):**
  - [ ] Function: `func ValidateKYCFile(header *multipart.FileHeader) error`
  - [ ] Check content type: allow `image/jpeg`, `image/png`, `application/pdf`
  - [ ] Check file size: maximum 10MB per document
  - [ ] Validate filename and extension
  - [ ] Return appropriate error messages

### Phase 4: API Endpoints 📋 PLANNED

#### 4.1 Create new file `route/kyc.go` (following route/profile.go structure EXACTLY)
- [ ] **File setup (copy exact structure from route/profile.go):**
  - [ ] Copy import block from profile.go exactly
  - [ ] Create `type kyc struct { *middleware }` (following profile struct)
  - [ ] Add init() function with handlers registration following profile.go pattern
  - [ ] Register route groups: `v1 := r.Group("/kyc/v1")` and `admin := v1.Group("/admin")`

#### 4.2 Member Endpoints (following exact middleware and response patterns from profile.go)
- [ ] **Implement POST `/kyc/v1/documents/upload` (following v1_UploadProfilePicture pattern):**
  - [ ] Add line: `v1.POST("/documents/upload", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_UploadDocument())`
  - [ ] Function signature: `func (s kyc) v1_UploadDocument() gin.HandlerFunc {`
  - [ ] Get session: `session := s.Session(c)`
  - [ ] Parse form file: `file, header, err := c.Request.FormFile("file")`
  - [ ] Parse form values for document_type and file_type
  - [ ] Validate using ValidateKYCFile helper
  - [ ] Call storage.UploadKYCDocument method
  - [ ] Update member documents in database using member.UpdateKYCDocuments
  - [ ] Return JSON response with document info
  - [ ] Use exact error handling pattern from profile picture upload

- [ ] **Implement GET `/kyc/v1/status` (following v1_GetProfile pattern):**
  - [ ] Add line: `v1.GET("/status", s.BearerAuth(enum.PermissionUserViewSelf), s.v1_GetStatus())`
  - [ ] Function signature: `func (s kyc) v1_GetStatus() gin.HandlerFunc {`
  - [ ] Get session: `session := s.Session(c)`
  - [ ] Find member: `member, err := s.store.Db.Member.FindByUserID(c.Request.Context(), session.UserId)`
  - [ ] Build KYCStatusDto response
  - [ ] Calculate required vs uploaded documents
  - [ ] Return: `c.JSON(http.StatusOK, status)`
  - [ ] Use exact error handling from profile.go

- [ ] **Implement POST `/kyc/v1/submit` (following v1_UpdateProfile pattern):**
  - [ ] Add line: `v1.POST("/submit", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_SubmitForVerification())`
  - [ ] Function signature: `func (s kyc) v1_SubmitForVerification() gin.HandlerFunc {`
  - [ ] Parse request body using ShouldBind exactly like profile.go
  - [ ] Validate all required documents uploaded
  - [ ] Update KYC status to "submitted" using member.UpdateKYCStatus
  - [ ] Send confirmation email using mail service
  - [ ] Create audit log using s.AuditLog exactly like profile.go
  - [ ] Return submission confirmation JSON

- [ ] **Implement DELETE `/kyc/v1/documents/:document_type` (following delete patterns):**
  - [ ] Add line: `v1.DELETE("/documents/:document_type", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_DeleteDocument())`
  - [ ] Function signature: `func (s kyc) v1_DeleteDocument() gin.HandlerFunc {`
  - [ ] Parse document_type parameter: `docType := c.Param("document_type")`
  - [ ] Get member and validate document exists
  - [ ] Delete from storage using DeleteKYCDocument
  - [ ] Update database to remove reference
  - [ ] Return deletion confirmation JSON

#### 4.3 Admin Endpoints (following existing admin/CMS patterns from user management)
- [ ] **Implement GET `/kyc/v1/admin/pending` (following pagination patterns from user.go):**
  - [ ] Add line: `admin.GET("/pending", s.BearerAuth(enum.PermissionKYCView), s.v1_GetPendingVerifications())`
  - [ ] Parse pagination parameters following exact user.go pattern
  - [ ] Query using member.GetPendingKYCVerifications
  - [ ] Set pagination headers exactly like existing endpoints
  - [ ] Return array of pending members

- [ ] **Implement GET `/kyc/v1/admin/members/:member_id`:**
  - [ ] Add line: `admin.GET("/members/:member_id", s.BearerAuth(enum.PermissionKYCView), s.v1_GetMemberKYC())`
  - [ ] Parse member_id parameter
  - [ ] Query member with KYC details
  - [ ] Generate document view URLs using GetKYCDocumentURL
  - [ ] Return comprehensive member info

- [ ] **Implement POST `/kyc/v1/admin/verify/:member_id`:**
  - [ ] Add line: `admin.POST("/verify/:member_id", s.BearerAuth(enum.PermissionKYCVerify), s.v1_VerifyMember())`
  - [ ] Parse member_id and verification data using ShouldBind
  - [ ] Validate action (approve/reject)
  - [ ] Update member KYC status using member.UpdateKYCStatus
  - [ ] Send notification email using mail service
  - [ ] Create audit log using s.AuditLog
  - [ ] Return verification result

- [ ] **Implement GET `/kyc/v1/admin/documents/:member_id/:filename`:**
  - [ ] Add line: `admin.GET("/documents/:member_id/:filename", s.BearerAuth(enum.PermissionKYCView), s.v1_DownloadDocument())`
  - [ ] Parse member_id and filename parameters
  - [ ] Validate member owns document
  - [ ] Generate secure download URL using GetKYCDocumentURL
  - [ ] Log document access for audit
  - [ ] Redirect to document URL using `c.Redirect(http.StatusTemporaryRedirect, url)`

### Phase 5: Permission System Extensions 📋 PLANNED

#### 5.1 Extend `pkg/enum/index.go` (following exact permission patterns)
- [ ] **Add new permission constants (after line 53, maintaining alphabetical order):**
  - [ ] Add line: `PermissionKYCView Permission = "kyc_view"`
  - [ ] Add line: `PermissionKYCVerify Permission = "kyc_verify"`
  - [ ] Insert in correct alphabetical position among existing permissions

- [ ] **Update PermissionTenantValues function (around line 56):**
  - [ ] Add `PermissionKYCView,` to permissions array in alphabetical order
  - [ ] Add `PermissionKYCVerify,` to permissions array in alphabetical order
  - [ ] Maintain existing array structure and formatting

- [ ] **Update PermissionRootValues function (around line 77):**
  - [ ] Add `PermissionKYCView,` to permissions array in alphabetical order
  - [ ] Add `PermissionKYCVerify,` to permissions array in alphabetical order
  - [ ] Maintain existing array structure and formatting

### Phase 6: Email Integration 📋 PLANNED

#### 6.1 Extend `pkg/mail/index.go` (following existing mail method patterns)
- [ ] **Add SendKYCSubmissionConfirmation method (following existing mail methods):**
  - [ ] Function signature: `func (s *Mail) SendKYCSubmissionConfirmation(ctx context.Context, email, name string) error {`
  - [ ] Create email subject: `subject := "KYC Documents Submitted - Under Review"`
  - [ ] Create email body with submission confirmation message
  - [ ] Include expected processing timeline (2-3 business days)
  - [ ] Call: `return s.sendMail(email, subject, body)` (following existing pattern)
  - [ ] Add error handling and logging exactly like existing methods

- [ ] **Add SendKYCApprovalNotification method:**
  - [ ] Function signature: `func (s *Mail) SendKYCApprovalNotification(ctx context.Context, email, name string) error {`
  - [ ] Create email subject: `subject := "KYC Verification Approved - Account Activated"`
  - [ ] Create email body with approval confirmation
  - [ ] Include information about activated features
  - [ ] Call: `return s.sendMail(email, subject, body)`
  - [ ] Add error handling and logging

- [ ] **Add SendKYCRejectionNotification method:**
  - [ ] Function signature: `func (s *Mail) SendKYCRejectionNotification(ctx context.Context, email, name, reason string) error {`
  - [ ] Create email subject: `subject := "KYC Verification - Additional Information Required"`
  - [ ] Create email body with rejection details including specific reason
  - [ ] Include resubmission instructions and support contact
  - [ ] Call: `return s.sendMail(email, subject, body)`
  - [ ] Add error handling and logging

### Phase 7: Database Operations 📋 PLANNED

#### 7.1 Extend `store/db/member.go` (following exact repository patterns)
- [ ] **Add UpdateKYCDocuments method (after existing member methods, around line 275):**
  - [ ] Function signature: `func (s *member) UpdateKYCDocuments(ctx context.Context, memberID string, documents map[string]string) error {`
  - [ ] Create update object: `update := M{"$set": M{...}}` (following exact UpdateStatus pattern)
  - [ ] Add updated_at timestamp using same pattern as existing methods
  - [ ] Call: `return s.repo.UpdateOne(ctx, M{"_id": OID(memberID)}, update)`
  - [ ] Add error handling following exact pattern from existing methods

- [ ] **Add UpdateKYCStatus method:**
  - [ ] Function signature: `func (s *member) UpdateKYCStatus(ctx context.Context, memberID, status, verifiedBy string, verification *KYCVerification) error {`
  - [ ] Create update object with status and verification data
  - [ ] Set appropriate timestamps based on status transition
  - [ ] Call: `return s.repo.UpdateOne(ctx, M{"_id": OID(memberID)}, update)`
  - [ ] Add status transition validation
  - [ ] Return appropriate errors for invalid transitions

- [ ] **Add GetPendingKYCVerifications method (following FindByStatus pattern exactly):**
  - [ ] Function signature: `func (s *member) GetPendingKYCVerifications(ctx context.Context, tenantID enum.Tenant, page, limit int64) ([]*MemberDomain, int64, error) {`
  - [ ] Create filter: `M{"tenant_id": tenantID, "kyc_status": "submitted"}`
  - [ ] Add sorting: sort by kyc_documents.submitted_at ascending
  - [ ] Apply pagination with skip and limit following existing patterns
  - [ ] Get total count using CountDocuments
  - [ ] Call: `return domains, total, s.repo.FindAll(ctx, query, &domains)`

- [ ] **Add CountKYCByStatus method:**
  - [ ] Function signature: `func (s *member) CountKYCByStatus(ctx context.Context, tenantID enum.Tenant, status string) int64 {`
  - [ ] Create filter: `M{"tenant_id": tenantID, "kyc_status": status}`
  - [ ] Call: `return s.repo.CountDocuments(ctx, Query{Filter: filter})`
  - [ ] Follow exact pattern from existing Count method

### Phase 8: Testing 📋 PLANNED

#### 8.1 Create comprehensive mock test suite `route/kyc_test.go` (following route/auth_test.go patterns)
- [ ] **Test infrastructure setup (following existing test patterns):**
  - [ ] Import all required testing packages: `"testing"`, `"net/http/httptest"`, `"bytes"`, `"encoding/json"`
  - [ ] Import gin testing: `"github.com/gin-gonic/gin"`
  - [ ] Import mock frameworks: `"github.com/stretchr/testify/mock"`, `"github.com/stretchr/testify/assert"`
  - [ ] Create test setup function with test database following existing patterns
  - [ ] Create mock storage service: `type MockStorage struct { mock.Mock }`
  - [ ] Create mock email service: `type MockMail struct { mock.Mock }`
  - [ ] Create test member accounts and JWT tokens for authentication
  - [ ] Set up test file uploads with valid and invalid documents

- [ ] **Mock storage methods (required for TDD):**
  - [ ] Mock `UploadKYCDocument` method with success and failure scenarios
  - [ ] Mock `DeleteKYCDocument` method with success and failure scenarios
  - [ ] Mock `GetKYCDocumentURL` method with valid URL generation
  - [ ] Mock `ValidateKYCFile` helper with file validation scenarios

- [ ] **Mock database methods (required for TDD):**
  - [ ] Mock `UpdateKYCDocuments` with success and member not found scenarios
  - [ ] Mock `UpdateKYCStatus` with valid and invalid status transitions
  - [ ] Mock `GetPendingKYCVerifications` with pagination scenarios
  - [ ] Mock `CountKYCByStatus` with various status counts
  - [ ] Mock `FindByUserID` for member lookup scenarios

- [ ] **Mock email methods (required for TDD):**
  - [ ] Mock `SendKYCSubmissionConfirmation` with success and failure
  - [ ] Mock `SendKYCApprovalNotification` with success and failure  
  - [ ] Mock `SendKYCRejectionNotification` with success and failure

#### 8.2 Member endpoint tests (TDD - write before implementing endpoints)
- [ ] **Test POST `/kyc/v1/documents/upload` (complete test scenarios):**
  - [ ] `TestUploadDocument_Success` - Valid file upload with all parameters
  - [ ] `TestUploadDocument_InvalidFileType` - Unsupported file format (e.g., .txt)
  - [ ] `TestUploadDocument_FileTooLarge` - File exceeding 10MB limit
  - [ ] `TestUploadDocument_MissingFileType` - Missing file_type parameter
  - [ ] `TestUploadDocument_MissingDocumentType` - Missing document_type parameter
  - [ ] `TestUploadDocument_StorageFailure` - Mock storage service failure
  - [ ] `TestUploadDocument_DatabaseFailure` - Mock database update failure
  - [ ] `TestUploadDocument_Unauthorized` - Missing or invalid JWT token
  - [ ] `TestUploadDocument_InsufficientPermissions` - Wrong permission level

- [ ] **Test GET `/kyc/v1/status` (comprehensive status scenarios):**
  - [ ] `TestGetStatus_PendingKYC` - New member with no documents
  - [ ] `TestGetStatus_WithPartialDocuments` - Member with some documents uploaded
  - [ ] `TestGetStatus_WithAllDocuments` - Member with all required documents
  - [ ] `TestGetStatus_Submitted` - Member with submitted verification
  - [ ] `TestGetStatus_Approved` - Member with approved KYC
  - [ ] `TestGetStatus_Rejected` - Member with rejected KYC and reason
  - [ ] `TestGetStatus_MemberNotFound` - Invalid member lookup
  - [ ] `TestGetStatus_Unauthorized` - Missing authentication

- [ ] **Test POST `/kyc/v1/submit` (submission workflow scenarios):**
  - [ ] `TestSubmitForVerification_Success` - Complete documents submission
  - [ ] `TestSubmitForVerification_IncompleteDocuments` - Missing required documents
  - [ ] `TestSubmitForVerification_AlreadySubmitted` - Duplicate submission attempt
  - [ ] `TestSubmitForVerification_DatabaseFailure` - Status update failure
  - [ ] `TestSubmitForVerification_EmailFailure` - Email notification failure
  - [ ] `TestSubmitForVerification_InvalidRequestBody` - Malformed JSON
  - [ ] `TestSubmitForVerification_Unauthorized` - Authentication failure

- [ ] **Test DELETE `/kyc/v1/documents/:document_type` (deletion scenarios):**
  - [ ] `TestDeleteDocument_Success` - Valid document deletion
  - [ ] `TestDeleteDocument_DocumentNotFound` - Non-existent document
  - [ ] `TestDeleteDocument_InvalidDocumentType` - Invalid document type parameter
  - [ ] `TestDeleteDocument_StorageFailure` - Storage deletion failure
  - [ ] `TestDeleteDocument_DatabaseFailure` - Database update failure
  - [ ] `TestDeleteDocument_Unauthorized` - Authentication failure

#### 8.3 Admin endpoint tests (TDD - write before implementing admin endpoints)
- [ ] **Test GET `/kyc/v1/admin/pending` (admin workflow scenarios):**
  - [ ] `TestGetPendingVerifications_EmptyList` - No pending verifications
  - [ ] `TestGetPendingVerifications_WithPagination` - Multiple pending entries with pagination
  - [ ] `TestGetPendingVerifications_FilterByTenant` - Multi-tenant data isolation
  - [ ] `TestGetPendingVerifications_SortingByDate` - Correct chronological ordering
  - [ ] `TestGetPendingVerifications_PaginationHeaders` - Correct pagination metadata
  - [ ] `TestGetPendingVerifications_Unauthorized` - Missing admin permissions
  - [ ] `TestGetPendingVerifications_InsufficientPermissions` - Wrong permission level

- [ ] **Test GET `/kyc/v1/admin/members/:member_id` (member detail scenarios):**
  - [ ] `TestGetMemberKYC_Success` - Valid member KYC details with documents
  - [ ] `TestGetMemberKYC_NotFound` - Invalid member ID
  - [ ] `TestGetMemberKYC_WithDocumentURLs` - Proper document URL generation
  - [ ] `TestGetMemberKYC_CrossTenantAccess` - Tenant isolation validation
  - [ ] `TestGetMemberKYC_Unauthorized` - Missing admin permissions

- [ ] **Test POST `/kyc/v1/admin/verify/:member_id` (verification scenarios):**
  - [ ] `TestVerifyMember_ApprovalSuccess` - Successful KYC approval
  - [ ] `TestVerifyMember_RejectionWithReason` - KYC rejection with reason
  - [ ] `TestVerifyMember_RejectionWithoutReason` - Missing reason validation
  - [ ] `TestVerifyMember_InvalidAction` - Invalid verification action
  - [ ] `TestVerifyMember_MemberNotFound` - Non-existent member
  - [ ] `TestVerifyMember_EmailNotificationSuccess` - Email sending validation
  - [ ] `TestVerifyMember_EmailNotificationFailure` - Email failure handling
  - [ ] `TestVerifyMember_AuditLogCreation` - Audit trail verification
  - [ ] `TestVerifyMember_Unauthorized` - Missing verification permissions

- [ ] **Test GET `/kyc/v1/admin/documents/:member_id/:filename` (document access scenarios):**
  - [ ] `TestDownloadDocument_Success` - Valid document download URL
  - [ ] `TestDownloadDocument_MemberNotFound` - Invalid member ID
  - [ ] `TestDownloadDocument_DocumentNotFound` - Missing document file
  - [ ] `TestDownloadDocument_CrossMemberAccess` - Document ownership validation
  - [ ] `TestDownloadDocument_AccessLogging` - Audit trail for document access
  - [ ] `TestDownloadDocument_URLExpiration` - Presigned URL expiration validation
  - [ ] `TestDownloadDocument_Unauthorized` - Missing view permissions

#### 8.4 Database operation tests `store/db/member_kyc_test.go` (TDD for database layer)
- [ ] **Test UpdateKYCDocuments method:**
  - [ ] `TestUpdateKYCDocuments_Success` - Valid document updates
  - [ ] `TestUpdateKYCDocuments_InvalidMemberID` - Non-existent member
  - [ ] `TestUpdateKYCDocuments_PartialUpdate` - Update specific document fields
  - [ ] `TestUpdateKYCDocuments_TimestampUpdate` - Verify updated_at field

- [ ] **Test UpdateKYCStatus method:**
  - [ ] `TestUpdateKYCStatus_ValidTransitions` - All allowed status changes
  - [ ] `TestUpdateKYCStatus_InvalidTransitions` - Blocked status changes
  - [ ] `TestUpdateKYCStatus_ApprovalData` - Approval with staff details
  - [ ] `TestUpdateKYCStatus_RejectionData` - Rejection with reason and staff
  - [ ] `TestUpdateKYCStatus_AttemptCounter` - Increment attempt count

- [ ] **Test GetPendingKYCVerifications method:**
  - [ ] `TestGetPendingKYCVerifications_FilterByTenant` - Multi-tenant isolation
  - [ ] `TestGetPendingKYCVerifications_FilterByStatus` - Status-specific filtering
  - [ ] `TestGetPendingKYCVerifications_Pagination` - Correct pagination behavior
  - [ ] `TestGetPendingKYCVerifications_SortOrder` - Chronological sorting
  - [ ] `TestGetPendingKYCVerifications_TotalCount` - Accurate count for pagination

- [ ] **Test CountKYCByStatus method:**
  - [ ] `TestCountKYCByStatus_AllStatuses` - Count for each KYC status
  - [ ] `TestCountKYCByStatus_TenantIsolation` - Tenant-specific counts
  - [ ] `TestCountKYCByStatus_EmptyResults` - Zero counts for empty data

#### 8.5 Integration tests (TDD for end-to-end workflows)
- [ ] **Test complete KYC workflow:**
  - [ ] `TestCompleteKYCWorkflow_SuccessPath` - Full workflow from upload to approval
  - [ ] `TestCompleteKYCWorkflow_RejectionPath` - Full workflow with rejection and resubmission
  - [ ] `TestCompleteKYCWorkflow_MultipleAttempts` - Multiple rejection and resubmission cycles

- [ ] **Test file upload and storage integration:**
  - [ ] `TestFileUploadIntegration_Success` - End-to-end file storage
  - [ ] `TestFileUploadIntegration_StorageFailure` - Storage service failure scenarios
  - [ ] `TestFileUploadIntegration_ValidationFailure` - File validation failures

- [ ] **Test email notification integration:**
  - [ ] `TestEmailIntegration_AllNotificationTypes` - All email scenarios
  - [ ] `TestEmailIntegration_ServiceFailure` - Email service failure handling

#### 8.6 Authentication and authorization tests (TDD for security)
- [ ] **Test endpoint security:**
  - [ ] `TestSecurityAllEndpoints_WithoutAuth` - All endpoints reject unauthenticated requests
  - [ ] `TestSecurityAllEndpoints_WithInvalidAuth` - All endpoints reject invalid tokens
  - [ ] `TestSecurityMemberEndpoints_WithMemberAuth` - Member endpoints accept member tokens
  - [ ] `TestSecurityAdminEndpoints_WithAdminAuth` - Admin endpoints require admin permissions
  - [ ] `TestSecurityAdminEndpoints_WithMemberAuth` - Admin endpoints reject member tokens

- [ ] **Test data isolation:**
  - [ ] `TestDataIsolation_CrossTenantAccess` - Prevent cross-tenant data access
  - [ ] `TestDataIsolation_CrossMemberAccess` - Prevent cross-member data access

### Phase 9: Documentation 📋 PLANNED

#### 9.1 Update `docs/swagger.yaml`
- [ ] **Add KYC endpoint definitions:**
  - [ ] Add `/kyc/v1/documents/upload` POST endpoint
  - [ ] Add `/kyc/v1/status` GET endpoint
  - [ ] Add `/kyc/v1/submit` POST endpoint
  - [ ] Add `/kyc/v1/documents/{document_type}` DELETE endpoint
  - [ ] Add `/kyc/v1/admin/pending` GET endpoint
  - [ ] Add `/kyc/v1/admin/members/{member_id}` GET endpoint
  - [ ] Add `/kyc/v1/admin/verify/{member_id}` POST endpoint
  - [ ] Add `/kyc/v1/admin/documents/{member_id}/{filename}` GET endpoint

- [ ] **Add data model definitions:**
  - [ ] Add `KYCDocumentUploadData` schema
  - [ ] Add `KYCSubmissionData` schema
  - [ ] Add `KYCVerificationData` schema
  - [ ] Add `KYCStatusDto` schema
  - [ ] Add security definitions for new permissions

#### 9.2 Create `docs/api-kyc-management.md`
- [ ] **Create comprehensive API documentation:**
  - [ ] Overview and authentication requirements
  - [ ] Member workflow documentation
  - [ ] Admin workflow documentation
  - [ ] Error codes and troubleshooting
  - [ ] Code examples and use cases
  - [ ] Security considerations

#### 9.3 Update `docs/architecture.md`
- [ ] **Add KYC system documentation:**
  - [ ] KYC workflow diagrams
  - [ ] Integration points with existing systems
  - [ ] Security and compliance notes
  - [ ] Future enhancement considerations

## Implementation Priority & Time Estimates (FOLLOWING TDD)

**TDD Approach**: Write Tests First, Then Implement to Pass Tests

1. **Phase 8** - Testing (40 mins) - **FIRST STEP - Write all tests**
2. **Phase 2** - Database schema extensions (30 mins) - Implement to pass tests
3. **Phase 3** - Storage layer extensions (20 mins) - Implement to pass tests  
4. **Phase 5** - Permission additions (10 mins) - Implement to pass tests
5. **Phase 4** - API endpoints (60 mins) - Implement to pass tests
6. **Phase 7** - Database operations (30 mins) - Implement to pass tests
7. **Phase 6** - Email integration (20 mins) - Implement to pass tests
8. **Phase 9** - Documentation (30 mins) - Final documentation

## Acceptance Criteria

### Functional Requirements
- [ ] Members can upload ID documents (front/back) and selfie photos
- [ ] Support for passport, driver's license, and national ID documents
- [ ] Administrative interface for document review and verification
- [ ] Email notifications for all KYC status changes
- [ ] Integration with existing membership and authentication systems
- [ ] Complete audit trail for compliance requirements

### Technical Requirements  
- [ ] Secure document storage with proper access controls
- [ ] File validation for type, size, and format
- [ ] API documentation and comprehensive testing
- [ ] Error handling following established patterns
- [ ] Performance optimization for file uploads

### Security Requirements
- [ ] Document access restricted to authorized administrators only
- [ ] Secure file upload with validation and scanning
- [ ] Comprehensive audit logging for all KYC operations
- [ ] Integration with existing permission and authentication systems

## Future Considerations

### Automated eKYC Integration 🔮
- API endpoints designed for future automated provider integration
- Document format standardization for machine processing
- Webhook support for real-time verification status updates

### Enhanced Document Support 🔮
- Additional document types and international formats
- Document expiration tracking and renewal notifications
- Integration with government databases for verification

This implementation provides a solid foundation for manual eKYC verification while maintaining the flexibility to integrate automated providers in future phases.
