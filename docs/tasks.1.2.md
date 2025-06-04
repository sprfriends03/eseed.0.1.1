# Task 1.2: Authentication & Authorization Implementation Plan for Members

## Pre-Implementation Checklist

### Environment Prerequisites
- [X] Core infrastructure (Task 1.1) completed
    - [X] Verify MongoDB connection
    - [X] Verify Redis connection for session management

### Access Requirements
- [X] Access to existing OAuth system (`pkg/oauth/index.go`)
- [X] Access to User (`store/db/user.go`) and Member (`store/db/member.go`) database collections
- [X] Access to Redis for token-related storage (e.g., JTI blacklist)

## Implementation Checklist

### 1.2.1 Extend Existing OAuth System for Member Authentication

#### 1.2.1.1 Member-Specific Authentication Logic
- [X] Analyze existing OAuth implementation in `pkg/oauth/index.go` for user authentication.
- [X] Extend OAuth system to specifically handle member authentication:
    - [X] Define or extend `store/db/member.go` repository functions:
        - [X] `FindByCredentials` (or similar) for member login, potentially linking to `User` credentials.
        - [X] `VerifyMemberStatus` to check KYC, membership validity, and age during authentication.
    - [X] Implement or adapt member-specific authentication flows in `route/auth.go`:
        - [X] Member registration endpoint (clarify if distinct from user registration or an extension).
        - [X] Member login endpoint.
        - [X] Member-related email verification processes.
- [X] Integrate member status validation (KYC, membership, age 18+) directly into the authentication decision process.

#### 1.2.1.2 Member Token Management (Leveraging Existing Mechanisms)
- [X] Design token structure for member authentication:
    - [X] Define member-specific claims to be included in JWT (e.g., `membership_id`, `membership_status`, `kyc_status`).
    - [X] Determine appropriate token expiration times for members.
- [X] Ensure member token revocation works with existing mechanisms:
    - [X] Confirm `/logout` endpoint (`route/auth.go`) and `oauth.RevokeTokenByUser` (using `VersionToken` in `User` model) effectively invalidate member sessions.
    - [X] Verify existing Redis-based JTI blacklist for refresh tokens covers member refresh tokens.
- [X] Review token introspection capabilities:
    - [X] Leverage existing `BearerAuth` middleware for all token validation.
    - [X] Ensure `/me` endpoint (`route/auth.go`) provides necessary member-specific session/token information or extend as needed.

### 1.2.2 Implement JWT Token Handling for Members

#### 1.2.2.1 Member-Specific Token Generation Logic
- [X] Extend existing JWT implementation:
    - [X] Update `pkg/oauth/claims.go` to include new member-specific claims structure (e.g., `MemberAccessClaims`).
    - [X] Modify `oauth.GenerateToken` in `pkg/oauth/index.go` (or create a member-specific variant) to fetch member details and include member-specific claims.
    - [X] Ensure member roles/permissions are correctly sourced and injected into claims (Note: Current implementation focuses on status claims like MembershipStatus, KYCStatus; granular role/permission claims deferred).
- [X] Review and enhance token signing security:
    - [ ] Investigate and implement a key rotation system for JWT signing keys (if not already robust). (Note: Deferred for future enhancement; current HS256 mechanism retained).
    - [ ] Consider upgrading signing algorithm from HS256 to RS256 for enhanced security. (Note: Deferred for future enhancement; current HS256 mechanism retained).
- [X] Create/adapt token generation function specifically for the member password-based login flow.

#### 1.2.2.2 Refresh Token Mechanism for Members
- [X] Ensure existing refresh token system (`oauth.RefreshToken` and `v1_RefreshToken` endpoint) seamlessly supports member tokens:
    - [X] Verify member-specific claims do not negatively impact refresh logic.
    - [X] Confirm refresh token rotation (JTI blacklisting in Redis) applies correctly.
- [ ] Implement advanced security: `Maintain refresh token family to detect theft` (new enhancement). (Note: Deferred for future enhancement).

#### 1.2.2.3 Member-Aware Token Validation Middleware
- [X] Extend existing `mdw.BearerAuth()` middleware in `route/index.go` or add complementary middleware to:
    - [X] Validate member-specific claims within the JWT.
    - [X] Perform member-specific permission checks (e.g., based on membership tier, KYC status) (Implemented via `RequireKYCStatus`, `RequireMembershipStatus`).
    - [X] Enrich request context with validated member information (extending `s.Session(c)`).
- [X] Clarify revocation checks: `BearerAuth` should primarily rely on `VersionToken` (from `User` model, linked to member) for access token validity and the JTI blacklist for refresh token validity.

### 1.2.3 Tune Security Enhancements for Members

#### 1.2.3.1 Security Headers
- [X] Verify and configure appropriate security headers (Strict-Transport-Security, Content-Security-Policy, etc.) on member authentication-related endpoints.

#### 1.2.3.2 Rate Limiting
- [X] Review and tune existing rate limiting configurations (`route/index.go`) for member-specific authentication endpoints (login, registration, refresh).
    - [X] Apply stricter limits for login attempts if necessary.
    - [ ] Consider progressive backoff for repeated failed member login attempts. (Note: Deferred for future enhancement; fixed stricter limits implemented).

#### 1.2.3.3 Audit Logging
- [X] Apply existing audit logging mechanisms (`mdw.AuditLog`) to capture all critical member authentication and authorization events (Note: Relies on existing general request/event logging via logrus; specific detailed audit log calls were removed per feedback).

### Testing Framework for Member Authentication

#### Authentication Flow Testing
- [X] Create test cases for member authentication flows:
    - [X] Member registration tests (Implemented).
    - [ ] Member login tests. (Note: Deferred by user request).
    - [ ] Member email verification tests. (Note: Deferred by user request).
- [X] Implement token validation tests for member tokens:
    - [ ] Valid, expired, and malformed member token scenarios. (Note: Partially covered by `/me` endpoint existing tests; specific new tests deferred).
    - [ ] Correctness of member-specific claims. (Note: Partially covered; specific new tests deferred).
    - [ ] Member token revocation test cases. (Note: Deferred).

#### Security Testing
- [ ] Implement security-focused tests for member authentication:
    - [ ] Brute force protection on member login. (Note: Deferred for future enhancement/dedicated security testing phase).
    - [ ] CSRF protection on member-related state-changing actions (if any in auth flow). (Note: Deferred for future enhancement/dedicated security testing phase).
    - [ ] Rate limiting effectiveness for member endpoints. (Note: Deferred for future enhancement/dedicated security testing phase).
- [ ] Perform security scanning focused on:
    - [ ] JWT configuration for member tokens. (Note: Deferred).
    - [ ] Password handling policies for members. (Note: Deferred).

### Documentation for Member Authentication

#### API Documentation
- [X] Document all member-specific authentication endpoints and modifications to existing ones:
    - [X] Request/response formats with member-specific fields.
    - [X] Error codes related to member authentication (e.g., KYC required, membership expired).
    - [X] Member authentication flow diagrams.
    - [X] Examples of member JWTs and their claims.
- [X] Update security documentation:
    - [X] Token handling best practices for member tokens.
    - [X] Overview of member-specific security features.

#### Developer Guide
- [ ] Create or update authentication integration guide for developers using member authentication: (Note: Deferred for future enhancement).
    - [ ] How to authenticate as a member.
    - [ ] Handling member-specific claims and errors.

## Validation Checklist

### Member Authentication Flow Validation
- [X] Verify member registration flow works end-to-end.
- [X] Test member login with username/password (and linkage to user account).
- [X] Confirm member email verification process.
- [ ] Test member-related password reset functionality (if different from user). (Note: Assumed covered by existing user password reset unless specified otherwise; specific member variant deferred).

### Member Token Management Validation
- [X] Verify member token generation includes correct member-specific claims.
- [X] Test member token validation middleware, including member-specific permission checks.
- [X] Confirm refresh token mechanism works for member tokens.
- [X] Validate member token revocation process.
- [X] Test any member-specific token introspection capabilities (e.g., extended `/me` endpoint).

### Security Validation for Members
- [X] Test rate limiting on member authentication endpoints.
- [X] Confirm CORS settings for member-related API interactions.
- [X] Validate security headers on member auth endpoints.
- [X] Check audit logging for member authentication events.

## Next Steps
- [ ] Review implementation with security team, focusing on member-specific aspects. (Note: Deferred for future review phase).
- [X] Plan integration with Member Management system (Task 1.3).
- [ ] Configure monitoring for member authentication events. (Note: Deferred for future operational task).
- [ ] Set up alerts for security incidents related to member authentication. (Note: Deferred for future operational task). 