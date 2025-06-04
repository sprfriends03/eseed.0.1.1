# Task 1.2: Authentication & Authorization for Members - Status

## Overall Status
- [ ] **Not Started**
- [X] **Planning Phase**
- [ ] In Progress
- [ ] Testing
- [ ] Completed
- [ ] Deployed

## Current Progress
- Refined detailed implementation plan (`docs/tasks.1.2.md`) to focus on member-specific extensions.
- Analyzed MVP requirements for member authentication and authorization.
- Identified existing components (`pkg/oauth`, `route/auth.go`, `store/db/user.go`) to be extended/leveraged.
- Clarified scope to exclude external OAuth providers and SSL certificate setup for this task.
- Detailed member-specific claims, validation logic, and necessary adaptations to token management.

## Key Milestones (from refined plan)
- [ ] 1.2.1 Extend Existing OAuth System for Member Authentication (0%)
    - [ ] 1.2.1.1 Member-Specific Authentication Logic (0%)
    - [ ] 1.2.1.2 Member Token Management (Leveraging Existing Mechanisms) (0%)
- [ ] 1.2.2 Implement JWT Token Handling for Members (0%)
    - [ ] 1.2.2.1 Member-Specific Token Generation Logic (0%)
    - [ ] 1.2.2.2 Refresh Token Mechanism for Members (0%)
    - [ ] 1.2.2.3 Member-Aware Token Validation Middleware (0%)
- [ ] 1.2.3 Tune Security Enhancements for Members (0%)
- [ ] Testing Framework for Member Authentication (0%)
- [ ] Documentation for Member Authentication (0%)

## Blockers/Issues
- Awaiting completion of Task 1.1 (Core Infrastructure Setup).
- Need to finalize member-specific claims structure for JWTs.
- Need to clearly define the relationship and authentication flow between `User` and `Member` entities (e.g., does a Member first register as a User?).

## Next Actions
1.  Finalize the design of the `Member` authentication flow, clarifying its relation to the existing `User` authentication.
2.  Define the precise list of member-specific claims for JWTs.
3.  Begin drafting `MemberAccessClaims` structure in `pkg/oauth/claims.go`.
4.  Outline repository functions required in `store/db/member.go` for fetching data needed during member authentication (e.g., KYC status, membership details).
5.  Start adapting `route/auth.go` for member registration/login endpoints based on the finalized flow.

## Recent Updates
- 2023-06-07 (Previous): Created initial implementation plan and status tracking; Analyzed existing code.
- 2023-06-07 (Current): Refined implementation plan in `docs/tasks.1.2.md` to remove external OAuth and SSL. Further refined plan to clearly delineate member-specific extensions versus leveraging existing user authentication mechanisms. Updated this status file accordingly. 