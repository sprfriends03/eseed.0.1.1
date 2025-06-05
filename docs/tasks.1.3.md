# Task 1.3: Member Management

## Overview
This task focuses on developing the backend functionalities for member registration and profile management. It includes creating APIs for member sign-up with necessary validations (input, duplication, email, age) and APIs for members to manage their profiles (CRUD operations, data validation, privacy controls). This system is crucial for user onboarding and data management.

## Key Implementation Areas

### 1. Member Registration API
   - **1.1 Input Validation:**
     - Implement comprehensive validation for all registration fields (e.g., name, email, password, date of birth).
     - Ensure data types, formats, and required fields are correctly handled.
   - **1.2 Duplicate Detection:**
     - Implement logic to check for existing members based on unique identifiers (e.g., email address).
     - Provide appropriate feedback to users attempting to register with existing credentials.
   - **1.3 Email Verification:**
     - Integrate with an email service to send verification links/codes upon registration.
     - Implement an endpoint to confirm email verification.
     - Manage member status based on email verification (e.g., pending verification, verified).
   - **1.4 Age Verification (18+):**
     - Implement logic to verify that the registering member is 18 years or older based on the provided date of birth.
     - Ensure compliance with legal requirements for age restriction.

### 2. Profile Management API
   - **2.1 CRUD Operations:**
     - Develop endpoints for Creating (handled by registration), Reading, Updating, and Deleting member profiles.
     - Ensure that members can only manage their own profiles.
   - **2.2 Data Validation:**
     - Implement validation for all fields that can be updated in the member profile.
     - Ensure data integrity and consistency.
   - **2.3 Privacy Controls:**
     - Design and implement mechanisms for members to control the visibility of their profile information.
     - Ensure that privacy settings are respected across the platform.

## Code Reuse & Alignment with `docs/base.md`

To ensure consistency, maintainability, and speed of development, the implementation of Member Management will heavily reuse existing components, patterns, and guidelines outlined in `docs/base.md`.

### General Backend Structure:
-   **Routes (`route/member.go` - new):**
    -   Structure similar to `route/user.go`.
    -   Utilize `BearerAuth` middleware from `route/index.go` for authenticated endpoints.
    -   Define appropriate permissions in `pkg/enum/index.go` (e.g., `PermissionMemberViewSelf`, `PermissionMemberUpdateSelf`, `PermissionMemberDeleteSelf`).
    -   Employ standard middleware chain from `route/index.go` (Cors, Compress, Trace, Logger, Recover, Error).
-   **Data Storage (`store/db/member.go` - new):**
    -   MongoDB model and queries for the `Members` collection, patterned after `store/db/user.go` and `store/db/role.go`.
    -   The `Members` collection schema (MVP Task 1.1.2.1) should incorporate `BaseDomain` fields (`_id`, `created_at`, `updated_at`, `created_by`, `updated_by`) as specified in `docs/base.md` under "Common Features Across Collections".
    -   Consider unique indexes for email, similar to user collection in `docs/base.md`.
-   **Error Handling:**
    -   Use predefined error codes and structures from `pkg/ecode/index.go`.
-   **API Documentation:**
    -   All new endpoints in `route/member.go` must have Swagger annotations, as per existing conventions.
-   **Password Hashing:**
    -   If members have passwords managed directly by this system (and not solely via OAuth providers which is the primary method for this task), use `pkg/util/password.go` (`HashPassword`, `VerifyPassword`) for password management, as shown in `docs/base.md` (User Management > Password Management). The registration API for Task 1.3 will primarily rely on OAuth but should consider password field requirements if a direct registration path is also implied by 'Member registration API'.

### 1. Member Registration API (`route/member.go`)
-   **1.1 Input Validation:**
    -   Leverage `pkg/validate/index.go` for validating all registration fields (name, email, date of birth, etc.).
    -   Refer to "Input validation" under "Security Implementation" in `docs/base.md`.
-   **1.2 Duplicate Detection:**
    -   Implement database queries in `store/db/member.go` to check for uniqueness of fields like email. This will be similar to how `store/db/user.go` handles checks for `username` or `email` uniqueness using `FindOneBy...` methods against indexed fields.
-   **1.3 Email Verification:**
    -   Utilize `pkg/mail/index.go` for sending verification emails.
    -   A new endpoint in `route/member.go` will handle token/code validation from the email link.
    -   The `Members` collection in `store/db/member.go` will need a field to track email verification status.
-   **1.4 Age Verification (18+):**
    -   Input for date of birth will be validated using `pkg/validate/index.go`.
    -   Custom logic within the registration service layer will calculate age and enforce the 18+ rule.

### 2. Profile Management API (`route/member.go`)
-   **2.1 CRUD Operations:**
    -   **Read (GET `/v1/members/me` or similar):** Endpoint to fetch the authenticated member's profile.
        -   Use `s.BearerAuth(enum.PermissionMemberViewSelf)` (new permission to be defined).
        -   Service logic will use the authenticated member's ID from the session (obtained via `BearerAuth`) to fetch data from `store/db/member.go`.
    -   **Update (PUT `/v1/members/me` or similar):** Endpoint to update the authenticated member's profile.
        -   Use `s.BearerAuth(enum.PermissionMemberUpdateSelf)` (new permission to be defined).
        -   Service logic validates input and updates the record in `store/db/member.go` scoped to the authenticated member.
    -   **Delete (DELETE `/v1/members/me` or similar):** Endpoint for a member to delete their own profile.
        -   Use `s.BearerAuth(enum.PermissionMemberDeleteSelf)` (new permission to be defined).
        -   Service logic to handle deletion, considering data retention policies and GDPR compliance (e.g., soft delete vs. hard delete, anonymization).
-   **2.2 Data Validation:**
    -   Leverage `pkg/validate/index.go` for validating all updatable profile fields.
-   **2.3 Privacy Controls:**
    -   This is largely a new feature area for member-facing controls.
    -   **Schema:** Add fields to the `Members` model in `store/db/member.go` for storing privacy preferences (e.g., visibility of certain profile fields).
    -   **API Endpoints:** New endpoints might be needed (e.g., `PUT /v1/members/me/privacy`) or integrated into the general profile update endpoint to manage these settings.
    -   **Logic:** Data retrieval methods in `store/db/member.go` and service layers must honor these settings when preparing data for responses. This extends the principle of permission checking to field-level visibility based on user settings.

### Cache Management (Consideration for future optimization):
-   While Member Management (Task 1.3) doesn't explicitly list caching in `mvp_tasks.md`, consider applying caching strategies from `store/rdb/index.go` and patterns like `GetUser` in `store/index.go` from `docs/base.md` for frequently accessed, relatively static member profile data if performance becomes a concern. This would involve:
    -   Defining cache keys (e.g., `member:%s`).
    -   Setting appropriate TTLs (e.g., `MemberCacheTTL`).
    -   Implementing cache-aside patterns (`GetWithRefresh` or similar from `store/rdb/index.go`).

## Dependencies
- **Authentication System (Task 1.2):** Requires a fully functional authentication system for securing member-specific actions and managing sessions.
- **Email Service Integration:** Necessary for sending verification emails and other notifications. This should be part of the core infrastructure or a dedicated service.
- **Data Encryption Setup:** Member-sensitive data (e.g., PII) must be encrypted at rest and in transit, requiring appropriate encryption mechanisms and key management from the core infrastructure.
- **Database Schema:** Requires `Members` collection to be defined and accessible as per `docs/mvp_tasks.md` (1.1.2.1).

## Timeline
- **Estimated Duration:** 1-2 weeks (aligning with "Week 3-4" timeline, assuming 1 week per major section from mvp_tasks)
- **Target Completion:** [TBD, to be updated based on project progress]

## Testing Requirements
- **Registration Flow Tests:**
  - Test successful registration with valid data.
  - Test registration with invalid input for each field.
  - Test duplicate email/username registration attempts.
  - Test email verification link/code generation and validation.
  - Test age verification logic (below 18, exactly 18, above 18).
- **Profile Management Tests:**
  - Test CRUD operations for member profiles (e.g., retrieve profile, update profile fields, attempt to delete profile).
  - Test data validation for profile updates.
  - Test privacy control settings and their enforcement.
  - Test access control (e.g., a member cannot update another member's profile).
- **Security Tests:**
  - Test for common vulnerabilities (e.g., SQL injection, XSS) if applicable to API interactions.
  - Ensure password hashing and secure storage.
- **GDPR Compliance Tests:**
  - Verify data minimization principles.
  - Test data access and portability features if applicable.
  - Test right to erasure (deletion of profile).

## Documentation Needs
- **API Documentation:**
  - Detailed Swagger/OpenAPI specification for all member management endpoints.
  - Include request/response formats, authentication requirements, and error codes.
- **Data Model Documentation:**
  - Updated documentation for the `Members` collection, including all fields, data types, and constraints.
  - Relationship with other collections if any.
- **Privacy Policy Documentation:**
  - Contribution to the platform's privacy policy regarding how member data is collected, stored, used, and managed.
  - Explanation of member rights and privacy controls. 

## Implementation Checklist

### General Backend Structure & Setup
- [ ] **Routes (`route/member.go`):**
    - [ ] Create new file `route/member.go`.
    - [ ] Structure `route/member.go` similar to `route/user.go`.
    - [ ] Utilize `BearerAuth` middleware from `route/index.go` for authenticated endpoints.
    - [ ] Define permissions in `pkg/enum/index.go`: `PermissionMemberViewSelf`, `PermissionMemberUpdateSelf`, `PermissionMemberDeleteSelf`.
    - [ ] Employ standard middleware chain from `route/index.go` (Cors, Compress, Trace, Logger, Recover, Error).
- [ ] **Data Storage (`store/db/member.go`):**
    - [ ] Create new file `store/db/member.go`.
    - [ ] Implement MongoDB model and queries for `Members` collection, patterned after `store/db/user.go` and `store/db/role.go`.
    - [ ] Incorporate `BaseDomain` fields (`_id`, `created_at`, `updated_at`, `created_by`, `updated_by`) into `Members` schema.
    - [ ] Add unique index for email in `Members` collection.
- [ ] **Error Handling:**
    - [ ] Consistently use predefined error codes/structures from `pkg/ecode/index.go`.
- [ ] **API Documentation:**
    - [ ] Add Swagger annotations for all new endpoints in `route/member.go`.
- [ ] **Password Hashing (Conditional):**
    - [ ] If direct password management is implemented, use `pkg/util/password.go` for hashing.

### 1. Member Registration API (`route/member.go`)
- [ ] **1.1 Input Validation:**
    - [ ] Implement comprehensive validation for all registration fields (name, email, password, date of birth) using `pkg/validate/index.go`.
    - [ ] Ensure data types, formats, and required fields are correctly handled.
    - [ ] Refer to "Input validation" under "Security Implementation" in `docs/base.md`.
- [ ] **1.2 Duplicate Detection:**
    - [ ] Implement database queries in `store/db/member.go` to check for email uniqueness.
    - [ ] Provide appropriate feedback for duplicate registration attempts.
- [ ] **1.3 Email Verification:**
    - [ ] Integrate with `pkg/mail/index.go` to send verification emails.
    - [ ] Implement an endpoint in `route/member.go` to confirm email verification (handle token/code).
    - [ ] Add a field to `Members` model in `store/db/member.go` to track email verification status.
    - [ ] Manage member status based on email verification (e.g., pending, verified).
- [ ] **1.4 Age Verification (18+):**
    - [ ] Validate date of birth input using `pkg/validate/index.go`.
    - [ ] Implement custom logic in the registration service layer to calculate age and enforce the 18+ rule.
    - [ ] Ensure compliance with legal requirements for age restriction.

### 2. Profile Management API (`route/member.go`)
- [ ] **2.1 CRUD Operations:**
    - [ ] **Read (GET `/v1/members/me` or similar):**
        - [ ] Develop endpoint to fetch the authenticated member's profile.
        - [ ] Secure with `s.BearerAuth(enum.PermissionMemberViewSelf)`.
        - [ ] Service logic to use authenticated member's ID to fetch data from `store/db/member.go`.
    - [ ] **Update (PUT `/v1/members/me` or similar):**
        - [ ] Develop endpoint to update the authenticated member's profile.
        - [ ] Secure with `s.BearerAuth(enum.PermissionMemberUpdateSelf)`.
        - [ ] Service logic to validate input and update record in `store/db/member.go` (scoped to authenticated member).
    - [ ] **Delete (DELETE `/v1/members/me` or similar):**
        - [ ] Develop endpoint for a member to delete their own profile.
        - [ ] Secure with `s.BearerAuth(enum.PermissionMemberDeleteSelf)`.
        - [ ] Service logic to handle deletion (consider data retention, GDPR: soft/hard delete, anonymization).
    - [ ] Ensure members can only manage their own profiles.
- [ ] **2.2 Data Validation:**
    - [ ] Implement validation for all updatable profile fields using `pkg/validate/index.go`.
    - [ ] Ensure data integrity and consistency.
- [ ] **2.3 Privacy Controls:**
    - [ ] Design and implement mechanisms for members to control profile information visibility.
    - [ ] **Schema:** Add fields to `Members` model in `store/db/member.go` for privacy preferences.
    - [ ] **API Endpoints:** Develop/integrate endpoints (e.g., `PUT /v1/members/me/privacy` or part of profile update) to manage privacy settings.
    - [ ] **Logic:** Ensure data retrieval methods in `store/db/member.go` and service layers honor privacy settings.
    - [ ] Ensure privacy settings are respected across the platform.

### Dependencies Verification
- [ ] Confirm Authentication System (Task 1.2) is fully functional.
- [ ] Confirm Email Service Integration is available and configured.
- [ ] Confirm Data Encryption Setup (for PII) is in place.
- [ ] Confirm `Members` collection schema (Task 1.1.2.1) is defined and accessible.

### Testing Requirements
- [ ] **Registration Flow Tests:**
    - [ ] Test successful registration with valid data.
    - [ ] Test registration with invalid input for each field.
    - [ ] Test duplicate email/username registration attempts.
    - [ ] Test email verification link/code generation and validation.
    - [ ] Test age verification logic (below 18, exactly 18, above 18).
- [ ] **Profile Management Tests:**
    - [ ] Test CRUD operations for member profiles (retrieve, update, delete).
    - [ ] Test data validation for profile updates.
    - [ ] Test privacy control settings and their enforcement.
    - [ ] Test access control (member cannot update another member's profile).
- [ ] **Security Tests:**
    - [ ] Test for common vulnerabilities (e.g., SQL injection, XSS).
    - [ ] Ensure password hashing and secure storage (if applicable).
- [ ] **GDPR Compliance Tests:**
    - [ ] Verify data minimization principles.
    - [ ] Test data access and portability features.
    - [ ] Test right to erasure (deletion of profile).

### Documentation Needs
- [ ] **API Documentation:**
    - [ ] Create/Update detailed Swagger/OpenAPI specification for all member management endpoints.
    - [ ] Include request/response formats, authentication requirements, and error codes.
- [ ] **Data Model Documentation:**
    - [ ] Update documentation for the `Members` collection (fields, data types, constraints, relationships).
- [ ] **Privacy Policy Documentation:**
    - [ ] Contribute content for the platform's privacy policy regarding member data.
    - [ ] Explain member rights and privacy controls in documentation.

### Cache Management (Future Optimization - Optional for MVP unless performance dictates)
- [ ] Consider applying caching strategies from `store/rdb/index.go`.
- [ ] Define cache keys (e.g., `member:%s`).
- [ ] Set appropriate TTLs (e.g., `MemberCacheTTL`).
- [ ] Implement cache-aside patterns (`GetWithRefresh` or similar). 