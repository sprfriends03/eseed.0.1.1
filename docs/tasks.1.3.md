# Task 1.3: Member Management (as an Extension of User System)

## Overview
This task focuses on developing backend functionalities for **user (referred to as Member for eG Platform context) registration and self-service profile management, extending the existing User system** outlined in `docs/base.md`. It includes enhancing APIs for user sign-up with necessary validations (input, duplication, email, age verification for members) and providing APIs for users (members) to manage their own profiles (CRUD operations, data validation, privacy controls). This system is crucial for user onboarding, self-service capabilities, and data management within their tenant context.

## Key Implementation Areas

### 1. User (Member) Registration API (Enhancements to User System)
   - **1.1 Input Validation (for Member-Specific Fields):**
     - Implement comprehensive validation for all registration fields, including **any new member-specific fields** (e.g., date of birth) added to the `user` model.
     - Ensure data types, formats, and required fields are correctly handled using existing validation mechanisms (`pkg/validate/index.go`).
   - **1.2 Duplicate Detection (for Users/Members):**
     - Ensure existing duplicate detection for user emails (and other unique identifiers like username, phone per tenant, as per `docs/base.md`) effectively serves member registration.
     - Provide appropriate feedback to users attempting to register with existing credentials.
   - **1.3 Email Verification (for Users/Members):**
     - Integrate with the existing email service (`pkg/mail/index.go`) to send verification links/codes upon user/member registration.
     - Implement or ensure an endpoint exists to confirm email verification.
     - Manage user/member status based on email verification (e.g., `data_status` field or a new field in the `user` model like `email_verified_at`).
   - **1.4 Age Verification (18+) (for Users/Members):**
     - Implement logic to verify that the registering user/member is 18 years or older based on a **new date of birth field in the `user` model**.
     - Ensure compliance with legal requirements for age restriction.

### 2. User (Member) Profile Management API (Self-Service Enhancements)
   - **2.1 Self-Service CRUD Operations:**
     - Develop or enhance endpoints (e.g., in `route/user.go`, or a new `route/account.go` or `route/profile.go`) for users/members to Read, Update, and potentially Delete their own profiles.
     - Ensure that users/members can only manage their own profiles, using appropriate permissions (e.g., `PermissionUserViewSelf`, `PermissionUserUpdateSelf`).
   - **2.2 Data Validation (for User/Member Profile Updates):**
     - Implement validation for all fields that can be updated by a user/member in their profile, using `pkg/validate/index.go`.
     - Ensure data integrity and consistency.
   - **2.3 Privacy Controls (New Feature for Users/Members):**
     - Design and implement mechanisms for users/members to control the visibility of their profile information.
     - This will likely involve adding **new privacy setting fields to the `user` model** in `store/db/user.go`.
     - Ensure that privacy settings are respected when displaying user/member data.

## Code Reuse & Alignment with `docs/base.md` (Focus on Extending User System)

To ensure consistency, maintainability, and speed of development, the implementation of Member Management will heavily reuse and extend existing User system components, patterns, and guidelines outlined in `docs/base.md`. **No new `member.go` files will be created for routes or DB; modifications will be within existing `user.go` files or new user-centric files if deemed more organized for self-service routes (e.g. `account.go` or `profile.go`).**

### General Backend Structure (Extending User Module):
-   **Routes (e.g., `route/user.go`, or new `route/account.go` / `route/profile.go` for self-service):**
    -   New self-service endpoints will follow the structure of existing routes in `route/user.go`.
    -   Utilize `BearerAuth` middleware from `route/index.go` for authenticated endpoints.
    -   Define appropriate **user self-service permissions** in `pkg/enum/index.go` if new ones are needed beyond general user permissions (e.g., `PermissionUserViewSelf`, `PermissionUserUpdateSelf`, `PermissionUserDeleteSelf` if not already present or suitable).
    -   Employ standard middleware chain from `route/index.go`.
-   **Data Storage (`store/db/user.go` - Modifications):**
    -   The existing `user` collection model in `store/db/user.go` will be **extended with new member-specific fields** (e.g., `date_of_birth`, `kyc_status`, `privacy_preferences`, `email_verified_at`).
    -   It already incorporates `BaseDomain` fields and is tenant-scoped.
    -   Existing unique indexes for email will apply. New indexes might be needed for new query patterns.
-   **Error Handling:**
    -   Use predefined error codes and structures from `pkg/ecode/index.go`.
-   **API Documentation:**
    -   All new or modified user/member-facing endpoints must have updated Swagger annotations.
-   **Password Hashing:**
    -   Existing password management (`pkg/util/password.go`) will be used if direct user/member password registration/management is part of the flow (alongside OAuth).

### 1. User (Member) Registration API (Extending `route/auth.go` or relevant user creation endpoint)
-   **1.1 Input Validation (for new fields in `user` model):**
    -   Leverage `pkg/validate/index.go` for validating all registration fields, including any **new fields added to the `user` DTOs/structs** (e.g., `date_of_birth`).
-   **1.2 Duplicate Detection:**
    -   Existing mechanisms in `store/db/user.go` for checking email/username uniqueness will be utilized.
-   **1.3 Email Verification (Workflow for `user`):**
    -   Utilize `pkg/mail/index.go` for sending user verification emails.
    -   Ensure an endpoint (possibly existing or new) handles token/code validation from the email link.
    -   The `user` model in `store/db/user.go` will be updated with a field to track email verification status (e.g., `email_verified_at`, or use `data_status`).
-   **1.4 Age Verification (18+) (Logic within User Service):**
    -   A **new `date_of_birth` field** will be added to the `user` model.
    -   Input validation for DOB via `pkg/validate/index.go`.
    -   Custom logic within the user service layer (associated with `store/db/user.go`) will calculate age and enforce the 18+ rule during registration.

### 2. User (Member) Profile Management API (Self-Service Endpoints)
    (e.g. in `route/user.go` or new `route/account.go`/`route/profile.go`)
-   **2.1 Self-Service CRUD Operations:**
    -   **Read (e.g., GET `/v1/account/profile`):** Endpoint to fetch the authenticated user's own profile.
        -   Secure with `s.BearerAuth(enum.PermissionUserViewSelf)` (or similar existing/new permission).
        -   Service logic uses authenticated user ID to fetch data from `store/db/user.go`.
    -   **Update (e.g., PUT `/v1/account/profile`):** Endpoint to update the authenticated user's own profile.
        -   Secure with `s.BearerAuth(enum.PermissionUserUpdateSelf)` (or similar).
        -   Service logic validates input (including any new `user` fields) and updates the user's record.
    -   **Delete (e.g., DELETE `/v1/account/profile`):** Consider if self-service deletion is MVP. If so:
        -   Secure with `s.BearerAuth(enum.PermissionUserDeleteSelf)` (or similar).
        -   Service logic handles deletion (soft delete preferred, by updating `data_status` in `user` model).
-   **2.2 Data Validation:**
    -   Leverage `pkg/validate/index.go` for validating all updatable fields in the user's self-managed profile.
-   **2.3 Privacy Controls (New fields in `user` model):**
    -   **Schema:** Add fields to the `user` model in `store/db/user.go` for storing privacy preferences (e.g., `profile_visibility_settings`).
    -   **API Endpoints:** The profile update endpoint (e.g., PUT `/v1/account/profile`) will manage these settings.
    -   **Logic:** Data retrieval methods for user profiles must honor these settings.

### Cache Management (Leveraging Existing User Cache Strategy):
-   The existing user cache strategy (`GetUser` in `store/index.go` using `UserCacheKey` and `UserCacheTTL` from `docs/base.md`) should be reviewed and potentially extended if new, frequently accessed user/member fields are added that would benefit from caching. Cache invalidation for `user` updates should cover these new fields.

## Dependencies
- **Authentication System (Task 1.2):** Requires a fully functional authentication system for securing user/member-specific actions.
- **Email Service Integration (`pkg/mail/index.go`):** Necessary for sending verification emails.
- **Data Encryption Setup:** Member-sensitive data (e.g., PII in the `user` model) must be encrypted at rest and in transit.
- **Database Schema:** The `user` collection schema in `store/db/user.go` will be modified.

## Timeline
- **Estimated Duration:** 1-2 weeks (aligning with "Week 3-4" timeline, adjusted for extending existing system rather than building new).
- **Target Completion:** [TBD, to be updated based on project progress]

## Testing Requirements
- **Registration Flow Tests (for Users/Members):**
  - Test successful registration with all required fields, including new member-specific ones (DOB).
  - Test input validation for new fields.
  - Test existing duplicate email/username registration attempts.
  - Test email verification link/code generation and validation workflow.
  - Test age verification logic (18+).
- **Profile Management Tests (Self-Service by User/Member):**
  - Test CRUD operations for a user managing their own profile (retrieve, update relevant fields).
  - Test data validation for profile updates, including new member-specific fields.
  - Test new privacy control settings and their enforcement.
  - Test access control (user can only manage their own profile).
- **Security Tests:**
  - Review new/modified endpoints for vulnerabilities.
  - Ensure password hashing (if direct password set) and secure storage for the `user` collection.
- **GDPR Compliance Tests:**
  - Verify data minimization principles for all `user` fields.
  - Test data access and portability features for users/members.
  - Test right to erasure (soft/hard deletion of `user` record).

## Documentation Needs
- **API Documentation:**
  - Update Swagger/OpenAPI specification for any modified or new user/member-facing endpoints.
  - Include request/response formats for DTOs involving new `user` fields.
- **Data Model Documentation:**
  - **Update documentation for the `user` collection** in `docs/base.md` or a dedicated data model document, detailing all fields including new member-specific ones, their data types, and constraints.
- **Privacy Policy Documentation:**
  - Contribution to the platform's privacy policy regarding how user/member data (now consolidated in the `user` collection) is collected, stored, used, and managed.
  - Explanation of user/member rights and privacy controls.

## Implementation Checklist (Revised for Extending User System)

### General Backend Structure & Setup (Extending User Module)
- [ ] **Routes (e.g., `route/user.go`, or new `route/account.go` / `route/profile.go`):**
    - [ ] Plan and create/modify route files for user (member) self-service registration and profile management.
    - [ ] Structure new routes reusing patterns from `route/user.go` (from `docs/base.md`).
    - [ ] Utilize existing `BearerAuth` middleware for all authenticated user (member) self-service endpoints.
    - [ ] Define/verify **user self-service permissions** in `pkg/enum/index.go` (e.g., `PermissionUserViewSelf`, `PermissionUserUpdateSelf`, `PermissionUserDeleteSelf`).
    - [ ] Employ the existing standard middleware chain.
- [ ] **Data Storage (`store/db/user.go` - Modifications):**
    - [ ] Identify and add **new member-specific fields** to the `user` model struct in `store/db/user.go` (e.g., `date_of_birth`, `kyc_status` (if MVP), `privacy_preferences`, `email_verified_at`).
    - [ ] Update MongoDB schema migrations/index definitions if new indexed fields are added to `user` collection.
- [ ] **Error Handling (for new/modified user/member endpoints):**
    - [ ] Consistently use predefined error codes/structures from `pkg/ecode/index.go`.
- [ ] **API Documentation (for new/modified user/member endpoints):**
    - [ ] Add/Update comprehensive Swagger annotations.
- [ ] **Password Handling for Users/Members:**
    - [ ] Confirm if direct password field in `user` model needs specific handling for member registration alongside primary OAuth flow. Ensure `pkg/util/password.go` is used.

### 1. User (Member) Registration API (Enhancements)
- [ ] **1.1 Input Validation (for new `user` fields):**
    - [ ] Define validation rules for new member-specific fields in user DTOs/structs.
    - [ ] Implement these validations leveraging `pkg/validate/index.go`.
- [ ] **1.2 Duplicate Detection:**
    - [ ] Verify existing user duplicate detection (email, username) meets member registration needs.
- [ ] **1.3 Email Verification (Workflow for `user`):**
    - [ ] Integrate `pkg/mail/index.go` for sending **user/member verification emails**.
    - [ ] Implement/ensure an endpoint handles **user/member email verification** (token/code validation).
    - [ ] Add/Utilize a field in `user` model (`store/db/user.go`) for email verification status.
    - [ ] Implement logic to manage user/member status based on email verification.
- [ ] **1.4 Age Verification (18+) (Logic in User Service):**
    - [ ] Add `date_of_birth` field to `user` model and relevant DTOs.
    - [ ] Validate DOB input using `pkg/validate/index.go`.
    - [ ] Implement business logic in user service layer to calculate age and enforce 18+ rule.

### 2. User (Member) Profile Management API (Self-Service Endpoints)
- [ ] **2.1 Self-Service CRUD Operations (Endpoints and Logic for User's Own Profile):**
    - [ ] Develop/enhance endpoints for users/members to manage their own profiles, reusing patterns from `docs/base.md`.
    - [ ] **Read (e.g., GET `/v1/account/profile`):**
        - [ ] Develop endpoint to fetch authenticated user's own profile.
        - [ ] Secure with `PermissionUserViewSelf` (or equivalent).
        - [ ] Implement service logic to fetch specific user data from `store/db/user.go`.
    - [ ] **Update (e.g., PUT `/v1/account/profile`):**
        - [ ] Develop endpoint to update authenticated user's own profile.
        - [ ] Secure with `PermissionUserUpdateSelf` (or equivalent).
        - [ ] Implement service logic to validate input (for all user-updatable fields including new ones) and update the user's record.
    - [ ] **Delete (e.g., DELETE `/v1/account/profile` - if MVP):**
        - [ ] Develop endpoint for user to delete their own profile.
        - [ ] Secure with `PermissionUserDeleteSelf` (or equivalent).
        - [ ] Implement service logic for deletion (e.g., soft delete by updating `data_status` in `user` model).
    - [ ] Ensure access control: user manages only their own profile.
- [ ] **2.2 Data Validation (for User-Updatable Profile Fields):**
    - [ ] Define validation rules for all user-updatable profile fields (including new member-specific ones).
    - [ ] Implement validations leveraging `pkg/validate/index.go` in DTOs/structs.
- [ ] **2.3 Privacy Controls (New Feature for Users/Members):**
    - [ ] **Schema:** Add necessary fields to `user` model in `store/db/user.go` for privacy preferences.
    - [ ] **API Endpoints:** Develop/integrate endpoints to manage these settings (e.g., via profile update).
    - [ ] **Logic:** Ensure data retrieval methods for `user` profiles honor these privacy settings.

### Dependencies Verification (for Extending User System for Members)
- [ ] Verify Authentication System (Task 1.2) is functional for user context.
- [ ] Confirm `pkg/mail/index.go` is ready for user/member verification emails.
- [ ] Confirm Data Encryption Setup is adequate for all sensitive PII in the (extended) `user` model.
- [ ] Ensure `user` collection schema in `store/db/user.go` is updated and accessible.

### Testing Requirements (for Extended User/Member Features)
- [ ] **Registration Flow Tests:** (Focus on user/member registration including new fields)
    - [ ] Test successful registration with new member-specific fields (DOB).
    - [ ] Test validation for these new fields.
    - [ ] Test existing duplicate email/username checks.
    - [ ] Test user/member email verification workflow.
    - [ ] Test age verification logic.
- [ ] **Profile Management Tests:** (Focus on user self-service for their profile)
    - [ ] Test self-service CRUD for user's own profile, including new fields.
    - [ ] Test data validation for these updates.
    - [ ] Test new privacy control settings for users/members.
    - [ ] Test access control (self-management only).
- [ ] **Security Tests:**
    - [ ] Review modified/new user endpoints for vulnerabilities.
- [ ] **GDPR Compliance Tests:** (Applied to extended `user` data handling)
    - [ ] Verify data minimization for all `user` fields.
    - [ ] Test data access/portability for users/members.
    - [ ] Test right to erasure for `user` profiles.

### Documentation Needs (for Extended User/Member Module)
- [ ] **API Documentation:**
    - [ ] Update Swagger for modified/new user/member-facing endpoints and DTOs.
- [ ] **Data Model Documentation:**
    - [ ] **Update `user` collection documentation** to include all new member-specific fields.
- [ ] **Privacy Policy Documentation:**
    - [ ] Update privacy policy regarding user/member data collection and controls.

### Cache Management (Review for Extended User Data)
- [ ] Review existing `UserCache` strategy.
- [ ] If new `user` fields are frequently accessed and benefit from caching, extend `UserCache` DTO and cache update logic. Ensure proper cache invalidation. 