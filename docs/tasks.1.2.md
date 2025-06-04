# Task 1.2: Authentication & Authorization Implementation Plan for Members

## Pre-Implementation Checklist

### Environment Prerequisites
- [ ] Core infrastructure (Task 1.1) completed
    - [ ] Verify MongoDB connection
    - [ ] Verify Redis connection for session management

### Access Requirements
- [ ] Access to existing OAuth system (`pkg/oauth/index.go`)
- [ ] Access to User (`store/db/user.go`) and Member (`store/db/member.go`) database collections
- [ ] Access to Redis for token-related storage (e.g., JTI blacklist)

## Implementation Checklist

### 1.2.1 Extend Existing OAuth System for Member Authentication

#### 1.2.1.1 Member-Specific Authentication Logic
- [ ] Analyze existing OAuth implementation in `pkg/oauth/index.go` for user authentication.
- [ ] Extend OAuth system to specifically handle member authentication:
    - [ ] Define or extend `store/db/member.go` repository functions:
        - [ ] `FindByCredentials` (or similar) for member login, potentially linking to `User` credentials.
        - [ ] `VerifyMemberStatus` to check KYC, membership validity, and age during authentication.
    - [ ] Implement or adapt member-specific authentication flows in `route/auth.go`:
        - [ ] Member registration endpoint (clarify if distinct from user registration or an extension).
        - [ ] Member login endpoint.
        - [ ] Member-related email verification processes.
- [ ] Integrate member status validation (KYC, membership, age 18+) directly into the authentication decision process.

#### 1.2.1.2 Member Token Management (Leveraging Existing Mechanisms)
- [ ] Design token structure for member authentication:
    - [ ] Define member-specific claims to be included in JWT (e.g., `membership_id`, `membership_status`, `kyc_status`).
    - [ ] Determine appropriate token expiration times for members.
- [ ] Ensure member token revocation works with existing mechanisms:
    - [ ] Confirm `/logout` endpoint (`route/auth.go`) and `oauth.RevokeTokenByUser` (using `VersionToken` in `User` model) effectively invalidate member sessions.
    - [ ] Verify existing Redis-based JTI blacklist for refresh tokens covers member refresh tokens.
- [ ] Review token introspection capabilities:
    - [ ] Leverage existing `BearerAuth` middleware for all token validation.
    - [ ] Ensure `/me` endpoint (`route/auth.go`) provides necessary member-specific session/token information or extend as needed.

### 1.2.2 Implement JWT Token Handling for Members

#### 1.2.2.1 Member-Specific Token Generation Logic
- [ ] Extend existing JWT implementation:
    - [ ] Update `pkg/oauth/claims.go` to include new member-specific claims structure (e.g., `MemberAccessClaims`).
    - [ ] Modify `oauth.GenerateToken` in `pkg/oauth/index.go` (or create a member-specific variant) to fetch member details and include member-specific claims.
    - [ ] Ensure member roles/permissions are correctly sourced and injected into claims.
- [ ] Review and enhance token signing security:
    - [ ] Investigate and implement a key rotation system for JWT signing keys (if not already robust).
    - [ ] Consider upgrading signing algorithm from HS256 to RS256 for enhanced security.
- [ ] Create/adapt token generation function specifically for the member password-based login flow.

#### 1.2.2.2 Refresh Token Mechanism for Members
- [ ] Ensure existing refresh token system (`oauth.RefreshToken` and `v1_RefreshToken` endpoint) seamlessly supports member tokens:
    - [ ] Verify member-specific claims do not negatively impact refresh logic.
    - [ ] Confirm refresh token rotation (JTI blacklisting in Redis) applies correctly.
- [ ] Implement advanced security: `Maintain refresh token family to detect theft` (new enhancement).

#### 1.2.2.3 Member-Aware Token Validation Middleware
- [ ] Extend existing `mdw.BearerAuth()` middleware in `route/index.go` or add complementary middleware to:
    - [ ] Validate member-specific claims within the JWT.
    - [ ] Perform member-specific permission checks (e.g., based on membership tier, KYC status).
    - [ ] Enrich request context with validated member information (extending `s.Session(c)`).
- [ ] Clarify revocation checks: `BearerAuth` should primarily rely on `VersionToken` (from `User` model, linked to member) for access token validity and the JTI blacklist for refresh token validity.

### 1.2.3 Tune Security Enhancements for Members

#### 1.2.3.1 Security Headers
- [ ] Verify and configure appropriate security headers (Strict-Transport-Security, Content-Security-Policy, etc.) on member authentication-related endpoints.

#### 1.2.3.2 Rate Limiting
- [ ] Review and tune existing rate limiting configurations (`route/index.go`) for member-specific authentication endpoints (login, registration, refresh).
    - [ ] Apply stricter limits for login attempts if necessary.
    - [ ] Consider progressive backoff for repeated failed member login attempts.

#### 1.2.3.3 Audit Logging
- [ ] Apply existing audit logging mechanisms (`mdw.AuditLog`) to capture all critical member authentication and authorization events:
    - [ ] Member login attempts (success/failure).
    - [ ] Member token generation and refresh events.
    - [ ] Member-related revocation events.

### Testing Framework for Member Authentication

#### Authentication Flow Testing
- [ ] Create test cases for member authentication flows:
    - [ ] Member registration and login tests.
    - [ ] Member email verification tests.
- [ ] Implement token validation tests for member tokens:
    - [ ] Valid, expired, and malformed member token scenarios.
    - [ ] Correctness of member-specific claims.
    - [ ] Member token revocation test cases.

#### Security Testing
- [ ] Implement security-focused tests for member authentication:
    - [ ] Brute force protection on member login.
    - [ ] CSRF protection on member-related state-changing actions (if any in auth flow).
    - [ ] Rate limiting effectiveness for member endpoints.
- [ ] Perform security scanning focused on:
    - [ ] JWT configuration for member tokens.
    - [ ] Password handling policies for members.

### Documentation for Member Authentication

#### API Documentation
- [ ] Document all member-specific authentication endpoints and modifications to existing ones:
    - [ ] Request/response formats with member-specific fields.
    - [ ] Error codes related to member authentication (e.g., KYC required, membership expired).
    - [ ] Member authentication flow diagrams.
    - [ ] Examples of member JWTs and their claims.
- [ ] Update security documentation:
    - [ ] Token handling best practices for member tokens.
    - [ ] Overview of member-specific security features.

#### Developer Guide
- [ ] Create or update authentication integration guide for developers using member authentication:
    - [ ] How to authenticate as a member.
    - [ ] Handling member-specific claims and errors.

## Validation Checklist

### Member Authentication Flow Validation
- [ ] Verify member registration flow works end-to-end.
- [ ] Test member login with username/password (and linkage to user account).
- [ ] Confirm member email verification process.
- [ ] Test member-related password reset functionality (if different from user).

### Member Token Management Validation
- [ ] Verify member token generation includes correct member-specific claims.
- [ ] Test member token validation middleware, including member-specific permission checks.
- [ ] Confirm refresh token mechanism works for member tokens.
- [ ] Validate member token revocation process.
- [ ] Test any member-specific token introspection capabilities (e.g., extended `/me` endpoint).

### Security Validation for Members
- [ ] Test rate limiting on member authentication endpoints.
- [ ] Confirm CORS settings for member-related API interactions.
- [ ] Validate security headers on member auth endpoints.
- [ ] Check audit logging for member authentication events.

## Next Steps
- [ ] Review implementation with security team, focusing on member-specific aspects.
- [ ] Plan integration with Member Management system (Task 1.3).
- [ ] Configure monitoring for member authentication events.
- [ ] Set up alerts for security incidents related to member authentication. 