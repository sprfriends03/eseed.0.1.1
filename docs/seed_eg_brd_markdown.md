# Business Requirements Document (BRD)

**Project**: Seed eG Plant Slot Management and Membership Platform  
**Document Version**: 1.1  
**Last Updated**: March 30, 2025  

## Version History

| Version | Date | Author | Description of Changes |
|---------|------|--------|------------------------|
| 1.0 | March 20, 2025 | Original Author | Initial document creation |
| 1.1 | March 30, 2025 | Business Analyst | Enhanced requirements, improved structure, added user stories and acceptance criteria |

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Market Analysis](#market-analysis)
3. [Terminology and Glossary](#terminology-and-glossary)
4. [Project Overview](#project-overview)
5. [Phase 1: MVP Requirements](#phase-1-mvp-requirements)
6. [Phase 2 - Advanced: Marketplace Integration](#phase-2---advanced-marketplace-integration)
7. [Phase 3 - Social Network: Community Features](#phase-3---social-network-community-features)
8. [Technical Requirements](#technical-requirements)
9. [Risk Management](#risk-management)
10. [Implementation Timeline](#implementation-timeline)

## 1. Executive Summary

The Seed eG Plant Slot Management and Membership Platform will enable cannabis club members to register, complete verification, purchase memberships, and manage plant slots through NFT integration. The platform will track complete plant lifecycles, handle harvests, and enable trading within the Seed eG ecosystem, all while ensuring compliance with German cannabis regulations.

### Key Business Benefits:
- Streamlined membership management compliant with German cannabis club regulations
- Transparent plant tracking from seed to harvest using blockchain technology
- New revenue opportunities through subscription memberships and trading functionality
- Enhanced member experience with intuitive interfaces and potential social features

### Implementation Timeline:
- **Phase 1 (MVP)**: May - September 2025 (4 months)
- **Phase 2 (Marketplace)**: September 2025 - December 2025 (4 months)
- **Phase 3 (Social Network)**: January - April 2026 (4 months)

### Resource Requirements:
- Development team: 3-5 full-time engineers
- Design team: 1-2 UX/UI designers
- QA team: 1-2 testers
- Project management: 1 project manager
- Business analyst: 1 analyst

## 2. Market Analysis

### 2.1 Cannabis Market in Germany

- Legalized for recreational use as of April 2024, the cannabis market in Germany is projected to be worth €1.8 billion by 2028.
- Under German law (effective April 1, 2024), adults are permitted to:
  - Possess 25 grams (7⁄8 oz) or less of cannabis in public
  - Possess up to 50 grams (1¾ oz) of dried cannabis in private
  - Grow up to three flowering cannabis plants at home

- **"Pillar 2"** of German cannabis law allows adults aged 18 and older to join cannabis clubs with the following regulations:
  - **Membership**: No more than 500 members per club
  - **Location**: Must be at least 200 meters away from schools, children's and youth facilities, and playgrounds
  - **Advertising**: Cannabis social clubs cannot advertise the club or their products in any way
  - **Sponsorships**: Clubs cannot accept sponsorships
  - **On-Site Consumption**: Cannabis consumption is prohibited inside the clubs or within 200 meters of the club
  - **Limited Cultivation**: Clubs can only grow the amount of cannabis needed for members, with excess cannabis to be destroyed
  - **Packaging and Labeling**: All packaging must be neutral, including a leaflet detailing weight in grams, harvest date, best before date, variety, average THC and CBD content, and average percentage of CBD

- Consumer segments include recreational users and wellness-focused individuals interested in cannabis-derived products such as oils, edibles, and supplements.
- The platform must implement compliance features to ensure all activities adhere to current German cannabis regulations.
- Seed eG will operate within the framework of a cannabis club under "Pillar 2" regulations.

### 2.2 Blockchain in Agriculture and Cannabis

- Blockchain technology offers transparency, security, and traceability, which is highly valuable in industries such as cannabis, where legal and regulatory oversight is critical.
- By using a token-based system, members can engage directly with the cultivation process, providing a unique value proposition.

### 2.3 Competitive Analysis

| Competitor | Strengths | Weaknesses | Our Advantage |
|------------|-----------|------------|---------------|
| Traditional Cannabis Clubs | Established membership base, physical locations | Limited technology integration, manual tracking | Transparent blockchain tracking, digital-first approach |
| Standard NFT Platforms | Advanced blockchain technology | Not cannabis-specific, lack of regulatory compliance | Industry-specific solution with built-in compliance |
| Cannabis Tracking Software | Industry-specific features | Typically B2B focused, not consumer-friendly | Consumer-focused platform with intuitive UX and ownership model |
| Social Cannabis Networks | Community engagement features | Limited cultivation management, no ownership model | Complete platform with cultivation management and NFT ownership |

### 2.4 Trading Plant Slots and Memberships

- The platform must allow for the trading of plant slots and memberships within The Seed eG ecosystem.
- This creates a secondary market that increases the value proposition for members and promotes engagement with the platform.
- Trading functionality provides flexibility for members while maintaining the regulatory compliance required in the cannabis industry.

## 3. Terminology and Glossary

### 3.1 Core Terminology

| Term | Definition |
|------|------------|
| Plant Slot | A unique, identifiable position that represents the right to cultivate one plant. Each plant slot has a unique ID string and can be in various states ("Draft", "Active", "Disabled"). |
| Membership | A subscription that grants a user the right to cultivate plants in assigned plant slots for a specific time period. Initial memberships last for variable durations (typically X months, e.g., "3" months) aligned with plant cycles. |
| Season | A defined period during which a specific set of plant slots is available for allocation. Each season will have "1000" unique plant slots represented as NFTs. |
| eKYC | Electronic Know Your Customer - the process of verifying the identity of users through digital means, requiring passport/ID verification and portrait matching. |
| Harvest | The process of collecting mature plants at the end of a growth cycle, coinciding with membership expiration. |
| Yield | The output produced by a plant at harvest time, which can be processed, sold to Seed eG, or handled through other options. |
| Cycle | A growth period for plants. Full growth cycle is "3" months for initial maturity, with subsequent harvests possible every "2" months. |
| Catalog | A collection of plant slots organized by season, containing information about plant species and other attributes. |
| Grace Period | A "5" day period after membership expiration during which a user can still renew their membership. |
| NFT | Non-Fungible Token - a unique digital asset on a blockchain that represents plant slots. Each season has "1000" NFTs pre-minted and then assigned to members. |
| Farm | A production facility where plants are cultivated, identified by a unique string number in the slot ID. |
| Area | A specific zone within a farm where plants are grouped, identified by a unique string number in the slot ID. |
| Plant Type | The specific variety of plant being cultivated, identified by a unique string number in the slot ID. |
| Smart Contract | Self-executing contract with the terms directly written into code that runs on the blockchain, automating execution and enforcement. |
| Token | Digital asset issued on a blockchain that can represent ownership rights or access privileges. |
| Gas Fee | Cost required to perform a transaction on the blockchain network. |
| Blockchain | Distributed digital ledger that records transactions across many computers to ensure security and transparency. |
| Hash | Unique identifier generated from digital data, used to verify data integrity. |
| Wallet | Digital tool that stores private keys and allows users to interact with blockchain applications. |

### 3.2 Plant Slot States

| State | Definition |
|-------|------------|
| "Draft" | A newly created plant slot that has not yet been assigned to a member. |
| "Active" | A plant slot that is assigned to a member with an active membership, available for cultivation. |
| "Deactivated" | A plant slot that is no longer active and cannot be used for cultivation, typically due to membership expiration or cancellation. |

### 3.3 Plant States

| State | Duration | Actions |
|-------|----------|---------|
| "Seed" | No specific time requirement | - Initial assignment to plant slot<br>- System records initial plant data |
| "Growth" | Variable (e.g., "3" months depending on plant type) | - Record cultivation information through application<br>- Track water, fertilizer, and care activities<br>- Monitor plant development |
| "Mature" | Variable (e.g., "1" week depending on plant type) | - Harvest and record yield in kilograms<br>- Perform post-harvest care<br>- Prepare plant for new growth cycle |
| "Deceased" | No specific time requirement | - Option 1: Remove and mark plant as permanently dead<br>- Option 2: Remove and use sub-branches to start a new cycle |

### 3.4 Membership States

| State | Description | Actions |
|-------|-------------|---------|
| "Normal" | Without Membership | - User has completed registration<br>- eKYC verification may or may not be completed<br>- No plant slots assigned<br>- Can browse platform and purchase membership<br>- Can view available plant slots but cannot cultivate<br>- Can participate in community features with limitations<br>- Can upgrade to membership status at any time if verified |
| "New" | First subscription membership | - Member is assigned "3" plant slots<br>- Member can cultivate "3" plants for X months (X matching first cycle of plant)<br>- System initiates tracking of membership period |
| "Renewal" | Subsequent subscription membership | - Member continues care of existing plants<br>- Membership period is extended<br>- If not renewed, it will be automatically canceled after the grace period ends |
| "Cancelled" | Terminated membership | - Plant slots are deactivated<br>- Plants are marked as deceased<br>- Member loses access to cultivation rights |
| "Transferred" | Membership moved to another member | - New membership is marked as transferred, while the old membership is marked as disabled<br>- Member will have access to cultivation rights<br>- Plant slots are also transferred with the membership |
| "Disabled" | Temporarily inactive membership | - Occurs when a member fails to meet platform requirements<br>- Can result from incomplete eKYC, policy violations, or administrative hold<br>- Plant slots remain assigned but cannot be actively cultivated<br>- Member cannot trade or renew membership while disabled<br>- Can be reactivated upon resolution of the triggering issue<br>- After "30" days in disabled state, system prompts admin for resolution decision |

### 3.5 Plant Slot Identification

Plant slots are uniquely identified using a structured ID format that provides information about the plant's location and type:

| Component | Digits | Possible Values | Description |
|-----------|--------|-----------------|-------------|
| Farm | "2" | "01"-"99" | Identifies the specific farm where the plant is cultivated |
| Area | "2" | "01"-"99" | Identifies the specific area within the farm |
| Plant Type | "2" | "01"-"99" | Identifies the type/variety of plant being cultivated |
| Slot Number | "4" | "0001"-"9999" | Sequential string number identifying the specific slot |

This structured ID system allows for efficient tracking and management of up to "10" farms, "100" areas per farm, "10" plant types, and "1000" slots per season.

**Example**: A slot ID of "09088090018" would represent:
- Farm: "09" (9th farm)
- Area: "88" (88th area)
- Plant Type: "09" (9th plant type)
- Slot Number: "18" (18th slot)

### 3.6 Relationship Between Plants, Plant Slots, and Membership

The platform's core operational model is based on the following relationships:

1. **Plant and Plant Slot Relationship**:
   - A plant is an attribute of a plant slot, not an independent string entity
   - Each plant slot contains exactly one plant at any given time
   - Plant slots persist across membership periods and seasons
   - The same plant can continue growing through multiple cycles as long as it remains alive (perpetual plant model)
   - Plant sub-branches are trimmed during growth to support development and overall plant health

2. **Plant Slot and Membership Relationship**:
   - Each member is allocated "3" plant slots upon membership activation
   - Plant slots remain assigned to a member as long as their membership remains active
   - If a membership expires or is cancelled, the associated plant slots are disabled for that member
   - Membership can be traded among users within The Seed eG ecosystem after completing eKYC. Plant slot ownership will also transfer.

3. **Plant Lifecycle and Membership Duration**:
   - First plant cycle (from seed to harvest): Approximately X months (typically "3" months)
   - Harvesting period: Approximately "1" week
   - Subsequent plant cycles: Approximately Y months (typically "2" months) each
   - Membership initial duration: Aligned with plant cycles (typically X months)
   - Members must maintain active membership to retain cultivation rights
   - Multiple membership renewals may be required to complete a complete plant cycle
   - The platform will notify members of upcoming renewals needed to complete plant cycles

4. **NFT Representation and Relationship**:
   - Plant slots are represented as NFTs ("1000" NFTs per season)
   - Each harvested product package can also be represented as an NFT
   - Plant slot NFTs are pre-minted and assigned to members
   - Harvest product NFTs are minted upon successful processing of a harvest
   - Ownership of a plant slot NFT grants rights to the future harvest product NFTs
   - Transfer of membership transfers all associated plant slot NFTs
   - Transfer of a harvest product NFT transfers ownership of the physical product

This interconnected model ensures continuity of plant cultivation while allowing flexibility in membership management and trading of plant slots and harvested products.

## 4. Project Overview

### 4.1 Project Background

Seed eG is developing a platform that allows users to cultivate plants through a membership-based model, operating within the framework of a cannabis club under Germany's "Pillar 2" regulations. Each member receives plant slots that represent ownership of physical plants being cultivated on their behalf. The platform will manage the entire lifecycle from membership registration to plant cultivation and harvesting, while ensuring compliance with German cannabis laws.

### 4.2 Project Objectives

1. Create a user-friendly platform for managing memberships and plant slots with NFT integration
2. Implement a robust eKYC verification process for member onboarding
3. Develop a system for tracking plant lifecycles and managing harvests
4. Build an administrative interface for Seed eG staff to manage the platform
5. Enable trading of plant slots and memberships within The Seed eG ecosystem
6. Establish a foundation for future marketplace and social features

### 4.3 Project Phases

The project will be implemented in three main phases:

#### Phase 1 - MVP (May - August 2025)
This phase consists of two sub-phases:
1. **Prototype**: Create a working prototype to demonstrate key concepts:
   - Membership functionality
   - NFT plant slot concept and integration
   - CMS for managing users and plants
2. **Full Flow MVP**: Complete implementation of core functionality

#### Phase 2 - Advanced (September 2025 - December 2025)
Marketplace development for NFTs and additional blockchain features

#### Phase 3 - Social Network (January - April 2026)
Social and community features

This BRD provides comprehensive requirements for all three phases of the project.

### 4.4 Requirements Traceability Matrix

| Business Objective | Related Requirements | Priority | Phase |
|-------------------|---------------------|----------|-------|
| Compliant membership management | - User Registration (5.2)<br>- eKYC Integration (5.2.2)<br>- Membership Management (5.3)<br>- Cannabis Club Compliance (5.10) | Must Have | "1" |
| Plant tracking with blockchain | - Plant Slot Management (5.4)<br>- Plant Management (5.5), NFT Integration | Must Have | "1" |
| Revenue through memberships | Payment Integration (5.7), Membership Management (5.3) | Must Have | "1" |
| Trading functionality | - Trading Plant Slots (5.3.4)<br>- Marketplace Integration (6.1) | Should Have | "1", "2" |
| Community engagement | - Social Network Features (7.1-7.4) | Could Have | "3" |

## 5. Phase 1: MVP Requirements

### 5.0 Requirements Approach

#### 5.0.1 Prioritization Framework

The following priorities have been assigned to requirements using the MoSCoW method:
- **Must Have**: Requirements critical to the legal operation and core functionality of the platform. These requirements must be included in the MVP release.
- **Should Have**: Important requirements that provide significant business value but are not critical for initial operation. These should be included in the MVP if possible.
- **Could Have**: Desirable requirements that would enhance the platform but can be deferred to later phases without compromising core functionality.
- **Won't Have**: Requirements that have been considered but explicitly decided against for this version of the platform.

#### 5.0.2 Dependencies and Relationships

The requirements in this document have interdependencies that must be considered during implementation planning. Key dependencies include:
- eKYC verification must be implemented before membership activation
- Plant slot management requires NFT integration to be completed
- Harvest management depends on complete plant lifecycle tracking
- Trading functionality requires both membership and plant slot management

### 5.1 Phase 1: MVP Requirements

#### 5.1.1 Prototype Sub-phase

The prototype will demonstrate the following core functionality:

**Membership Demonstration**
- Basic user registration and account creation process
- Simplified manual eKYC verification flow
- Membership purchase and activation interface
- Visual representation of membership status and expiration

**NFT Plant Slot Integration**
- Pre-minting of all "1000" NFT plant slots for the season
- Assignment of pre-minted NFTs to members upon membership activation
- Visual representation of NFTs in user interface
- Connection between NFT ownership and plant slot rights

**CMS Management**
- Basic administrative interface for user management
- Manual eKYC verification interface for administrators
- Plant management screens for administrators
- Simplified catalog management
- Basic reporting functionality
- NFT management capabilities

#### 5.1.2 Full Flow MVP

The complete MVP will expand upon the prototype to deliver the following:

### 5.2 User Registration and Onboarding

#### 5.2.1 Account Creation

**Priority**: Must Have

**User Stories**:
- As a potential member, I want to create an account so that I can apply for membership.
- As a system administrator, I want to verify that users are at least "18" years old so that we comply with German regulations.
- As a system administrator, I want to limit registrations to "500" members so that we comply with German cannabis club regulations.

**Functional Requirements**:
- Users must be able to create accounts using email/password or social login (Facebook, Google)
- System must capture essential information including name, contact details, and bank account information
- System must verify that users are at least "18" years old
- System must implement controls to ensure membership count does not exceed "500" members per German cannabis club regulations
- System must maintain waiting lists when membership capacity is reached

**Acceptance Criteria**:
1. Authentication system includes multi-factor options
2. Sensitive data is properly encrypted at rest and in transit
3. All features comply with GDPR requirements
4. API integrations implement secure authentication
5. Security audits and penetration testing are scheduled regularly
6. Backup and disaster recovery processes are verified
7. Session management includes appropriate timeouts
8. Security logging captures relevant events

**Security Controls**:

1. **Authentication Security**
   - Multi-factor authentication support
   - Password complexity requirements
   - Account lockout policies
   - Session token management

2. **Data Protection**
   - AES-256 encryption for data at rest
   - TLS 1.3 for data in transit
   - Personal data anonymization
   - Right to be forgotten implementation

3. **Infrastructure Security**
   - Network segmentation
   - Intrusion detection systems
   - Regular vulnerability scanning
   - Secure configuration management

4. **Compliance Controls**
   - GDPR compliance framework
   - Data retention policies
   - Audit logging
   - Privacy impact assessments

### 8.3 Integration Requirements

**Priority**: Must Have

**User Stories**:
- As a developer, I want well-defined integration points so that external systems connect smoothly.
- As an administrator, I want secure integrations so that data exchanges maintain privacy and integrity.

**Functional Requirements**:
- eKYC provider integration for identity verification
- Stripe payment integration for fiat transactions
- Email notification system for automated communications
- Blockchain integration for NFT functionality
- SMS gateway integration for important alerts
- Shipping provider integrations for product fulfillment
- Banking system integration for direct transfers

**Acceptance Criteria**:
- eKYC integration verifies identity documents securely
- Stripe integration processes payments reliably
- Email notification system delivers messages consistently
- Blockchain integration manages NFTs effectively
- SMS gateway sends time-sensitive alerts properly
- Shipping provider integrations track deliveries accurately
- Banking integration processes transfers securely

**Integration Specifications**:

1. **eKYC Integration**:
   - API: RESTful with OAuth 2.0 authentication
   - Document upload with encryption
   - Verification status webhooks
   - Response time SLA: "30" seconds
   - Fallback to manual verification
   - Error handling with specific codes

2. **Payment Integration**:
   - Stripe API v2023-10-16 or later
   - Payment methods: Credit/debit cards, SEPA, bank transfers
   - 3D Secure support
   - Webhook notifications for payment events
   - Refund and dispute handling
   - Reconciliation reports

3. **Blockchain Integration**:
   - Ethereum-compatible blockchain
   - Smart contract deployment and management
   - NFT minting and transfer capabilities
   - Gas optimization strategies
   - Metadata storage and retrieval
   - Event monitoring and handling

## 9. Risk Management

### 9.1 Integration Risks

**Priority**: High

**Risk Details**:

- **Manual eKYC Verification Challenges**: The manual verification process may create bottlenecks if volume increases significantly.
  - Probability: High
  - Impact: High
  - Severity: Critical

- **Payment Processing Complexities**: Stripe integration may require additional configuration for specific transaction types.
  - Probability: Medium
  - Impact: Medium
  - Severity: Moderate

- **NFT Implementation Risks**: Early integration of NFT functionality may face technical challenges or regulatory issues.
  - Probability: High
  - Impact: High
  - Severity: Critical

### 9.2 Operational Risks

**Priority**: Medium to High

**Risk Details**:

- **Scalability Concerns**: As user base grows, system performance and manual processes may be affected.
  - Probability: Medium
  - Impact: High
  - Severity: High

- **Regulatory Compliance**: Changes in regulations related to KYC or cannabis possession limits may impact implementation.
  - Probability: Medium
  - Impact: High
  - Severity: High

- **Blockchain Transaction Costs**: Fluctuating gas fees and transaction costs on the blockchain may affect operational costs.
  - Probability: High
  - Impact: Medium
  - Severity: High

- **Smart Contract Security**: Vulnerabilities in smart contracts could lead to security issues if not properly audited.
  - Probability: Medium
  - Impact: Critical
  - Severity: High

### 9.3 Integration Risk Mitigation

**Priority**: High

**Mitigation Strategies**:

1. **For Manual eKYC Verification Challenges**:
   - Implement an efficient workflow with clear verification guidelines
   - Develop training materials and SOPs for verification staff
   - Create dashboard for monitoring verification queue and performance
   - Establish escalation paths for complex verification cases
   - Design system for future transition to automated verification
   - Set up SLAs for verification completion (e.g., "24" hour maximum)

2. **For Payment Processing Complexities**:
   - Start with standard transactions first before complex scenarios
   - Implement sandbox testing for all payment flows
   - Develop comprehensive error handling for payment failures
   - Create reconciliation processes for payment verification
   - Document common issues and resolutions for support teams
   - Establish backup payment provider options

3. **For NFT Implementation Risks**:
   - Begin with basic NFT functionality in prototype phase
   - Conduct legal review of NFT implementation for compliance
   - Implement phased approach to blockchain integration
   - Develop off-chain fallback mechanisms
   - Engage blockchain security experts for smart contract audits
   - Create comprehensive testing environment for blockchain features

4. **For Scalability Concerns**:
   - Design system architecture with horizontal scaling capabilities
   - Implement performance monitoring and alerting
   - Plan for automated scaling based on usage patterns
   - Develop capacity planning procedures
   - Create load testing scenarios for peak usage

5. **For Regulatory Compliance**:
   - Establish regular compliance review schedule
   - Create flexible system design to accommodate regulatory changes
   - Maintain relationships with legal experts in cannabis regulation
   - Implement configuration-driven compliance controls
   - Document all compliance decisions and rationale

## 10. Implementation Timeline

### Sprint Plan Overview

**Phase 1 - MVP (May - September 2025)**

#### Months 1-2: Foundation & Prototype
- **Week 1-2**: Project setup, environment configuration
- **Week 3-4**: Basic user registration and authentication
- **Week 5-6**: Manual eKYC verification interface
- **Week 7-8**: Membership management prototype

#### Months 3-4: Core MVP Development
- **Week 9-12**: Plant slot management system
- **Week 13-16**: Plant lifecycle tracking
- **Week 17-20**: Payment integration (Stripe)
- **Week 21-24**: Basic NFT integration

#### Month 5: Testing & Launch Preparation
- **Week 25-28**: Comprehensive testing, bug fixes, documentation
- **Week 29-32**: MVP launch preparation and deployment

**Phase 2 - Advanced Features (September - December 2025)**

#### Months 6-7: Marketplace Foundation
- **Week 33-36**: Marketplace platform development
- **Week 37-40**: Listing management system
- **Week 41-44**: Transaction processing and escrow

#### Months 8-9: Product Marketplace
- **Week 45-48**: Product catalog and packaging
- **Week 49-52**: Order management and fulfillment
- **Week 53-56**: Advanced analytics and reporting

**Phase 3 - Social Features (January - April 2026)**

#### Months 10-11: Community Foundation
- **Week 57-60**: User profiles and connections
- **Week 61-64**: Content creation and sharing

#### Months 12-13: Advanced Social Features
- **Week 65-68**: Messaging and notifications
- **Week 69-72**: Community governance and moderation

### Key Milestones

| Milestone | Target Date | Description |
|-----------|-------------|-------------|
| Prototype Demo | July 2025 | Working prototype with core concepts |
| MVP Launch | September 2025 | Full MVP with all Phase 1 features |
| Marketplace Beta | November 2025 | Trading platform beta release |
| Phase 2 Complete | December 2025 | Full marketplace functionality |
| Social Features Beta | February 2026 | Community features beta |
| Full Platform Launch | April 2026 | Complete platform with all features |

### Resource Allocation

#### Development Team Structure
- **Technical Lead**: 1 full-time (all phases)
- **Frontend Developers**: 2 full-time (all phases)
- **Backend Developers**: 2 full-time (all phases)
- **Blockchain Developer**: 1 full-time (Phase 1-2), 0.5 full-time (Phase 3)
- **Mobile Developer**: 1 full-time (Phase 1-2), 0.5 full-time (Phase 3)

#### Support Team Structure
- **UX/UI Designer**: 1 full-time (all phases)
- **QA Engineer**: 1 full-time (all phases)
- **DevOps Engineer**: 1 full-time (all phases)
- **Project Manager**: 1 full-time (all phases)
- **Business Analyst**: 1 full-time (Phase 1), 0.5 full-time (Phase 2-3)

### Success Criteria

#### Phase 1 Success Metrics
- "500" member capacity reached within "3" months of launch
- "95"% uptime for core platform functionality
- Average eKYC verification time under "24" hours
- Zero security incidents
- "90"% user satisfaction score

#### Phase 2 Success Metrics
- "100" active marketplace listings within first month
- "€10,000" in marketplace transaction volume within "3" months
- "85"% order fulfillment success rate
- "95"% payment processing success rate

#### Phase 3 Success Metrics
- "70"% of members engage with social features monthly
- "500" posts created within first month
- "90"% community guideline compliance rate
- "95"% content moderation response time under "2" hours

### Risk Monitoring and Contingency Plans

#### Weekly Risk Assessment
- Technical risks evaluation
- Resource availability review
- Timeline adherence check
- Quality metrics monitoring

#### Contingency Triggers
- Development delays exceeding "2" weeks
- Resource unavailability for more than "1" week
- Quality metrics falling below "85"% targets
- Security incidents of any severity

#### Escalation Procedures
1. **Level 1**: Project Manager handles routine issues
2. **Level 2**: Technical Lead involves additional resources
3. **Level 3**: Stakeholder escalation for scope/timeline changes
4. **Level 4**: Executive decision for major pivots

---

## Conclusion

This Business Requirements Document provides a comprehensive framework for developing the Seed eG Plant Slot Management and Membership Platform with a string-based architecture approach. The platform will enable compliant cannabis club operations while leveraging blockchain technology for transparency and member engagement.

The three-phase implementation approach ensures incremental value delivery while managing technical and operational risks. All data types have been specified as strings to provide maximum flexibility and consistency across the platform.

Success depends on careful attention to German cannabis regulations, robust security measures, and seamless user experience across all touchpoints. Regular monitoring and adaptation will be essential as the platform scales and evolves to meet member needs and regulatory requirements.. Users can successfully register with email/password
2. Users can successfully register with social login options
3. System captures all required user information
4. System prevents registration of users under "18" years old
5. System limits active memberships to "500"
6. Users attempting to register when the club is at capacity are added to a waiting list
7. System notifies administrators when waiting list grows beyond "50" users

**Data Validation Rules**:
- Email must be in valid format and unique in the system
- Password must be at least "8" characters with minimum complexity requirements
- Name fields must contain only alphabetic characters
- Date of birth must indicate user is at least "18" years old
- Contact phone must be in valid international format

**API Integration Requirements**:
- Social login requires OAuth 2.0 authentication with Facebook and Google
- API endpoints must include: `/api/v1/users/register`, `/api/v1/users/login`, `/api/v1/users/social-login`
- Response formats must follow standard JSON structure with appropriate HTTP status codes

#### 5.2.2 eKYC Integration

**Priority**: Must Have

**User Stories**:
- As a registered user, I want to verify my identity so that I can become a full member.
- As a system administrator, I want to review verification documents so that I can approve or reject user verification.

**Functional Requirements**:
- In the MVP phase, system must implement a manual eKYC verification process
- Users must manually upload passport/ID images
- System must provide an administrative interface for staff to manually verify uploaded documents
- Staff must verify portrait matching with the provided ID
- System must record verification status and timestamp
- System must notify users of verification results
- eKYC verification must be completed before membership activation
- System architecture should support future integration with automated eKYC providers in later phases

**Acceptance Criteria**:
- Users can upload front and back images of government-issued ID
- Users can upload a portrait photo for verification
- Administrators can view submitted documents in a verification queue
- Administrators can approve or reject verification with comments
- System notifies users of verification outcome via email
- Verification status is recorded with timestamp and approver information
- Architecture includes API endpoints for future automated eKYC integration

**Error Handling**:
- If image quality is poor, system prompts user to re-upload with specific guidelines
- If verification is rejected, user receives specific reason and instructions for resubmission
- System logs all verification attempts, including failures, for compliance purposes

**Future Integration Architecture**:
- System will include standardized API endpoints for eKYC provider integration: `/api/v1/kyc/verify`, `/api/v1/kyc/status`
- Data format will follow industry standards for identity verification
- System will maintain dual-path capability (manual and automated) during transition

**International Document Support**:
- System must support verification of EU member state identification documents
- System must handle passport verification from all countries
- System must implement country-specific validation rules for ID formats
- System must provide guidance for users with non-German documentation
- System must maintain an audit trail of documents processed by country of origin

**Verification Failure Handling**:
- System must provide specific feedback for verification failures
- System must categorize rejection reasons (image quality, document validity, data mismatch)
- System must implement a structured resubmission process
- System must limit verification attempts to prevent abuse (maximum "3" attempts within "24" hours)
- System must escalate to manual review after multiple failures

#### 5.2.3 Membership Activation

**Priority**: Must Have

**User Stories**:
- As a verified user, I want to activate my membership so that I can start cultivating plants.
- As a system administrator, I want memberships to align with plant growth cycles so that members have a coherent experience.

**Functional Requirements**:
- Upon successful registration and eKYC verification, users must be able to activate a membership
- Initial membership period must be set for X months (for example, "3" months)
- System must align membership duration with plant growth cycle
- System must record membership activation date for lifecycle tracking

**Acceptance Criteria**:
1. Verified users can view and select membership options
2. Users can complete membership purchase process
3. System activates membership upon successful payment
4. System records precise activation timestamp
5. Membership duration correctly aligns with plant growth cycle
6. Users receive confirmation of successful activation via email

**Workflow Diagram**:
```
[User Registration] → [eKYC Verification] → [Payment Processing] → [Membership Activation] → [Plant Slot Allocation]
```

#### 5.2.4 Initial Plant Slot Allocation

**Priority**: Must Have

**User Stories**:
- As a new member, I want to receive my plant slots so that I can start growing plants.
- As a system administrator, I want to ensure each plant slot has a unique identifier so that we can track individual plants.

**Functional Requirements**:
- System must assign "3" distinct plant slots to new members after membership activation
- Each plant slot must have a unique identifier string
- Plant slots must be allocated from the available pool for the current season

**Acceptance Criteria**:
1. System automatically assigns exactly "3" plant slots to each new member
2. Each plant slot has a unique identifier following the specified format
3. Plant slots are selected from the available pool for the current season
4. Users can view their allocated plant slots immediately after membership activation
5. System records the allocation timestamp and membership association

**Plant Slot Allocation Algorithm**:
1. System identifies available slots in the current season
2. System selects "3" slots prioritizing diverse plant types when possible
3. System assigns slots to member and updates availability status
4. System mints or assigns corresponding NFTs to represent ownership

### 5.3 Membership Management

#### 5.3.1 Membership Lifecycle

**Priority**: Must Have

**User Stories**:
- As a member, I want to see my membership status so that I know when renewal is needed.
- As a system administrator, I want the system to automatically track membership states so that we maintain accurate records.

**Functional Requirements**:
- System must implement and track the following membership states:
  - "Normal": without membership
  - "New": First subscription for a member
  - "Renewal": Subsequent subscription periods
  - "Cancelled": Terminated membership
  - "Transferred": Membership moved to another member
  - "Disabled": Temporarily inactive membership
- "New" membership must allocate "3" plant slots to the member
- Initial membership period must align with first plant cycle (X months, typically "3" months)
- Renewal periods must maintain continuity of plant care
- System must auto-cancel membership after a week(s) if not renewed
- System must implement a "5" day grace period after expiration for renewal
- System must send automated reminders for upcoming membership expirations
- System must notify members when renewals are needed to complete plant cycles

**Acceptance Criteria**:
1. System accurately tracks and displays current membership state
2. System correctly allocates "3" plant slots for new memberships
3. Membership periods align with plant growth cycles
4. System sends reminder notifications "14", "7", and "3" days before expiration
5. System implements "5" day grace period for renewals after expiration
6. System automatically cancels membership if not renewed within grace period
7. System maintains complete history of state transitions with timestamps

**Notification Requirements**:
- Email notifications must be sent for all state changes
- Notification templates must be customizable via admin interface
- All notifications must be logged in the system for audit purposes

#### 5.3.2 Membership Status Management

**Priority**: Must Have

**User Stories**:
- As a member, I want to understand the implications of my membership status so that I can make informed decisions.
- As a system administrator, I want status changes to update related records so that data remains consistent.

**Functional Requirements**:
- System must track membership state changes with timestamps
- System must enforce rules for each state:
  - "Normal": Registered user without active membership
  - "New": Initial subscription with "3" plant slots allocated
  - "Renewal": Extension of existing membership with retained plant slots
  - "Cancelled": Deactivation of plant slots and disabling of plants
  - "Transferred": Movement of membership and associated plant slots to another member
  - "Disabled": Temporarily inactive membership with restrictions
- System must enforce the rule that an active membership ("New" or "Renewal" state) is required for plant cultivation
- When membership enters "Cancelled" state, system must deactivate plant slots but maintain plant data
- System must distinguish between plant disabled due to cancellation and plant deceased due to natural causes

**Acceptance Criteria**:
1. System records all status changes with timestamps and user information
2. System enforces business rules for each membership state
3. System maintains data integrity across related records during state changes
4. System prevents cultivation actions for members without active membership
5. System correctly handles plant and plant slot status during membership cancellation
6. System distinguishes between different causes of plant disabling

**State Transition Rules**:
- "Normal" → "New" (upon first membership purchase)
- "New" → "Renewal" (upon payment for next period)
- "New" → "Cancelled" (upon explicit cancellation or failure to renew after grace period)
- "New" → "Transferred" (upon approved transfer to another member)
- "New" → "Disabled" (upon rule violation or administrative action)
- "Renewal" → "Renewal" (upon payment for next period)
- "Renewal" → "Cancelled" (upon explicit cancellation or failure to renew after grace period)
- "Renewal" → "Transferred" (upon approved transfer to another member)
- "Renewal" → "Disabled" (upon rule violation or administrative action)
- "Disabled" → "New"/"Renewal" (upon resolution of issues)
- "Disabled" → "Cancelled" (after "30" days without resolution)

#### 5.3.3 Membership Transfer

**Priority**: Should Have

**User Stories**:
- As a member, I want to transfer my membership to another person so that they can continue growing my plants.
- As a receiver, I want to accept a membership transfer so that I can start growing without waiting.

**Functional Requirements**:
- System must allow transfer of membership to another eKYC-verified account
- System must support transfer of membership to Seed eG
- System must enforce that the entire membership with all associated plant slots is transferred together
- System must prohibit the transfer of individual plants separate from the membership
- System must maintain complete history of membership transfers with timestamps
- System must update all relevant records when a transfer occurs:
  - Membership ownership records
  - Plant slot allocation records
  - NFT ownership records
  - Billing and payment details
- System must ensure continuous care for plants during transfer process

**Acceptance Criteria**:
1. Members can initiate transfer to another verified member
2. Members can initiate transfer to Seed eG
3. System enforces bundled transfer of membership with all associated plant slots
4. System prevents transfer of individual plants
5. System maintains complete transfer history with timestamps
6. System updates all related records upon successful transfer
7. Plants continue to receive care during transfer process

**Security Requirements**:
- Both parties must complete 2-factor authentication for transfer
- Transfer requires explicit confirmation from receiver
- System must verify receiver has completed eKYC process
- Receiver cannot exceed maximum slot allocation (typically "3")

#### 5.3.4 Trading Plant Slots and Memberships

**Priority**: Should Have

**User Stories**:
- As a member, I want to list my plant slots for trade so that I can monetize my assets.
- As a member, I want to purchase plant slots from others so that I can expand my cultivation.

**Functional Requirements**:
- System must implement a trading platform for plant slots and memberships within The Seed eG ecosystem
- Trading must be restricted to eKYC-verified members only
- System must support different trading models (fixed price, auction, etc.)
- System must provide secure transaction processing for trades
- System must allow members to list their plant slots or memberships for trade
- System must capture listing details (price, duration, terms, etc.)
- System must verify eligibility for trading (e.g., not within grace period)
- System must support listing cancellation and modification
- System must handle secure transfer of plant slots and memberships between members
- System must update NFT ownership records upon successful trade
- System must manage escrow for payment security during trades
- System must maintain detailed transaction records for all trades
- System must enforce regulatory compliance for all trades
- System must implement KYC verification for all trading participants
- System must maintain audit trails for regulatory purposes
- System must support reporting for tax and compliance purposes

**Acceptance Criteria**:
1. Members can list plant slots or memberships for sale at fixed price
2. Members can browse and purchase available listings
3. System verifies eligibility of both parties before finalizing trades
4. System handles secure payment processing through escrow
5. System updates ownership records for plant slots and NFTs
6. System maintains detailed transaction records for all trades
7. System enforces regulatory compliance for all transactions

**Trading Process Flow**:
```
[Seller Lists Item] → [Buyer Places Order] → [Payment Held in Escrow] → [System Transfers Ownership] → [Escrow Released to Seller]
```

**Dispute Resolution Process**:
1. Buyer or seller can open dispute within "48" hours
2. Both parties submit evidence and statements
3. Administrator reviews case within "72" hours
4. Administrator can rule in favor of buyer or seller
5. System implements ruling by releasing escrow or reversing transfer

### 5.4 Plant Slot Management

#### 5.4.1 Plant Slot Lifecycle

**Priority**: Must Have

**User Stories**:
- As a member, I want to understand my plant slot's status so that I know what actions I can take.
- As an administrator, I want to track plant slot states so that I can manage inventory effectively.

**Functional Requirements**:
- System must manage plant slots in different states: "Draft", "Active", "Disabled"
  - "Draft": Newly created, unassigned slots
  - "Active": Assigned to a member and available for cultivation
  - "Disabled": No longer active, unavailable for cultivation
- System must enforce that active plant slots are tied to active memberships
- Each plant slot must be represented by an NFT on the blockchain
- Plants are attributes of slots, not independent entities
- System must use the structured ID format for plant slots (Farm-Area-PlantType-SlotNumber)

**Acceptance Criteria**:
1. System correctly tracks and displays plant slot states
2. System enforces relationship between active slots and active memberships
3. Each plant slot has corresponding NFT representation
4. Plant data is correctly associated with slot, not stored as separate entity
5. All plant slots follow the specified ID format structure

**Plant Slot State Diagram**:
```
[Creation] → "Draft" → "Active" → "Disabled"
                ↑         ↓
                └─────────┘
              (Reactivation)
```

#### 5.4.2 Plant Slot Activation and Disabling

**Priority**: Must Have

**User Stories**:
- As a system administrator, I want plant slots to activate automatically when a membership starts so that members can begin cultivation.
- As a system administrator, I want plant slots to be disabled when memberships end so that inactive members can't continue cultivation.

**Functional Requirements**:
- System must pre-mint all "1000" NFTs for each season before any member assignments
- System must assign pre-minted NFTs to members upon membership activation
- System must activate plant slots upon membership activation
- System must disable plant slots when membership expires or is cancelled
- System must return or transfer NFTs when plant slots are disabled
- System must enforce that disabled slots cannot be reactivated for the same member

**Acceptance Criteria**:
1. System successfully pre-mints "1000" NFTs for each season
2. System correctly assigns NFTs to members upon membership activation
3. Plant slots activate automatically upon membership activation
4. Plant slots disable automatically when membership expires or cancels
5. NFTs return to Seed eG pool or transfer appropriately when slots are disabled
6. System prevents reactivation of disabled slots for the same member

**NFT Technical Requirements**:
- NFTs must follow ERC-1155 standard on Ethereum or equivalent standard on selected blockchain
- Smart contracts must undergo independent security audit by reputable firm
- Implementation must include gas fee optimization strategies:
  - Batch minting for initial plant slot NFTs
  - Meta-transaction support to reduce user fees
  - Gas price monitoring with transaction scheduling
- Each NFT must include comprehensive metadata:
  - Plant slot unique identifier string
  - Season and catalog information
  - Creation timestamp and ownership history
  - Plant type details and attributes
- System must implement fallback mechanisms for blockchain disruptions:
  - Off-chain record keeping with blockchain reconciliation
  - Delayed transaction processing during network congestion
  - Alternative verification methods during outages
- Complete audit trail of all NFT transactions must be maintained
- Secondary market royalty mechanism ("5"-"10"%) for Seed eG sustainability

#### 5.4.3 Plant Slot Records

**Priority**: Must Have

**User Stories**:
- As a member, I want to view comprehensive information about my plant slots so that I can track my cultivation history.
- As an administrator, I want to maintain detailed records of all plant slots so that we have complete traceability.

**Functional Requirements**:
- System must maintain comprehensive records for each plant slot including:
  - Unique ID string
  - Owner information
  - Current status string
  - Plant information
  - Harvest history
  - Yield data string
  - NFT information (token ID, transaction history)

**Acceptance Criteria**:
1. System stores complete information for each plant slot
2. Members can view their own plant slot records via user interface
3. Administrators can access and manage all plant slot records
4. System maintains historical data even after ownership changes
5. System links NFT transactions with physical plant slot records
6. Data is retained for at least "7" years for compliance purposes

**Data Storage Requirements**:
- Plant slot records must be stored in both blockchain (for ownership) and traditional database (for details)
- Records must be immutable once created, with changes tracked as new versions
- Backup procedures must ensure data cannot be lost

### 5.5 Plant Management

#### 5.5.1 Plant Lifecycle

**Priority**: Must Have

**User Stories**:
- As a member, I want to track my plant's lifecycle so that I know when to expect harvests.
- As a cultivator, I want to record care activities so that I can optimize plant health and yield.

**Functional Requirements**:
- System must track plant lifecycle using the following states and durations:
  - "Seed": No specific time requirement (initial assignment state)
  - "Growth": Variable duration based on plant type (e.g., "3" months)
  - "Mature": Variable duration based on plant type (e.g., "1" week)
  - "Deceased": Variable duration based on plant type (e.g., "1" week)
- System must support the following actions for each state:
  - "Seed": Record initial plant data and assignment to plant slot
  - "Growth": Enable recording of cultivation information through application (water, fertilizer, care activities)
  - "Mature": Enable harvest process with yield recording in kilograms, post-harvest care, and preparation for new cycle
  - "Deceased": Support options to either mark plant as permanently dead or use sub-branches to start a new cycle
- System must track timing of each state based on plant type specifications
- System must notify users when state transitions are approaching or required
- System must maintain complete history of plant state transitions and actions taken

**Acceptance Criteria**:
1. System accurately tracks and displays current plant state
2. System allows appropriate actions based on current state
3. System records all care activities with timestamps
4. System notifies users of upcoming or required state transitions
5. System maintains complete history of plant lifecycle
6. System handles plant death scenarios appropriately

**Plant Lifecycle State Diagram**:
```
┌─────────┐
│    .    │
▼         │
"Seed" → "Growth" → "Mature" → New Growth Cycle
  ▲                     │
  │                     ▼
  └──────── "Deceased" ──┘
```

#### 5.5.2 Plant Information Tracking

**Priority**: Must Have

**User Stories**:
- As a member, I want to record detailed information about my plants so that I can improve my cultivation techniques.
- As a cultivator, I want a mobile-friendly interface to update plant status so that I can record information while inspecting plants.

**Functional Requirements**:
- System must track detailed plant information including:
  - Plant state string ("Seed", "Growth", "Mature", "Deceased")
  - State transition dates and durations
  - Cycle number string
  - Cultivation records (water, fertilizer, etc.) with timestamps
  - Harvest yield in kilograms string (when applicable)
  - Actions taken during each state
  - Plant type-specific metrics and requirements
  - Images of plant at different stages (optional)
- System must provide a mobile-responsive application for farmers to update plant status and cultivation information, with the following requirements:
  - Progressive Web App for cross-platform compatibility
  - Offline capability with data synchronization when connectivity returns
  - Performance optimization for low-bandwidth environments
  - Local storage encryption for cached data
  - Camera integration for plant documentation with automatic tagging
  - Location services for validation of maintenance activities
  - Push notifications for critical alerts and required actions
  - Battery optimization for extended field usage
  - Simplified UI for gloved operation in cultivation environments
  - Barcode/QR code scanning for plant identification
  - Voice input option for hands-free operation
  - Cross-device synchronization of member activities
- System must generate alerts based on expected state transitions
- System must provide reporting on plant performance across cycles

**Acceptance Criteria**:
1. System captures all required plant information
2. Mobile interface works effectively on various devices and screen sizes
3. Users can update plant information in offline mode with later synchronization
4. Photo upload with automatic tagging works correctly
5. Quick-entry forms reduce time required for common updates
6. System enforces data validation rules to ensure quality
7. System alerts users about upcoming required actions
8. System generates performance reports comparing plants across cycles

**Data Validation Rules**:
- Water and fertilizer amounts must be within predefined ranges for plant type
- Plant images must be properly categorized by growth stage
- State transition dates must follow logical sequence
- Harvest yield data must include weight in grams and quality rating

#### 5.5.3 Bulk Management

**Priority**: Should Have

**User Stories**:
- As a cultivator managing multiple plants, I want to perform batch updates so that I can save time.
- As an administrator, I want to ensure bulk operations have validation so that accidental mass changes are prevented.

**Functional Requirements**:
- System should support bulk editing of plant health and growth stage
- System should provide batch update capabilities for common maintenance tasks
- System should enable mass actions for similar plants across multiple slots
- System should include validation to prevent unintended mass changes

**Acceptance Criteria**:
1. Users can select multiple plants for batch updates
2. System provides appropriate bulk action options based on plant states
3. System implements confirmation steps for potentially destructive actions
4. System validates all bulk operations before execution
5. System provides clear feedback on bulk operation results
6. Batch operations are recorded with details of affected plants

**Safety Mechanisms**:
- Confirmation dialog for all bulk operations affecting more than "5" plants
- Preview of changes before final submission
- Ability to cancel bulk operation in progress
- Audit log of all bulk operations with user information

### 5.6 Harvest Management

#### 5.6.1 Harvest Scheduling

**Priority**: Must Have

**User Stories**:
- As a member, I want to know when my plants will be ready for harvest so that I can plan accordingly.
- As an administrator, I want harvests to align with membership cycles so that we maintain regulatory compliance.

**Functional Requirements**:
- System must schedule first harvest upon plant maturity (typically after X months, e.g., "3" months from initial planting)
- System must schedule subsequent harvests every Y months (typically "2" months) after the first harvest
- System must align harvest schedules with membership expiration dates
- System must notify members of upcoming harvests
- System must allow for plant-specific variations in harvest timing based on plant type and growth conditions

**Acceptance Criteria**:
1. System calculates and displays accurate harvest dates
2. Harvest schedules align appropriately with membership periods
3. System sends notifications "14", "7", and "3" days before expected harvest
4. System accommodates adjustments for plant-specific variations
5. Harvest scheduling algorithm accounts for different plant types

**Harvest Timing Calculation**:
- First harvest = Plant start date + Initial growth period (typically "3" months)
- Subsequent harvests = Previous harvest date + Regrowth period (typically "2" months)
- Adjustments based on plant type and growing conditions are applied

#### 5.6.2 Harvest Options

**Priority**: Must Have

**User Stories**:
- As a member, I want to choose how my harvest is processed so that I get the end product I desire.
- As an administrator, I want to record members' harvest choices so that we can process harvests accordingly.

**Functional Requirements**:
- System must support harvest and process option for "€100" per plant
- System must calculate processing fees based on membership ("3" plant slots = "€300")
- System must support yield sale option to Seed eG at set prices
- System must enforce that harvest options are only available at membership expiration or cancellation

**Acceptance Criteria**:
1. Members can select harvest and process option with clear pricing information
2. System correctly calculates processing fees based on number of plants
3. Members can request yield sale to Seed eG with price estimates
4. Harvest options are only available at appropriate times
5. System records member selections and processes accordingly

**Harvest Options Workflow**:
```
[Plant Reaches Maturity] → [System Presents Harvest Options] → [Member Selects Option] → [System Processes Selection and Calculates Fees] → [Harvest Execution]
```

#### 5.6.3 Post-Harvest Actions

**Priority**: Must Have

**User Stories**:
- As a member who renews my membership, I want my plants to continue to the next growth cycle after harvest so that I maintain continuous cultivation.
- As an administrator, I want clear procedures for handling plants after harvest so that we maintain compliance.

**Functional Requirements**:
- System must initiate a new plant cycle after harvest if membership is renewed
- System must disable plant slots if membership is cancelled after harvest
- System must revert plant slots to Seed eG if no action is taken during the grace period

**Acceptance Criteria**:
1. System automatically initiates new growth cycle when membership is renewed
2. System disables plant slots when membership is cancelled after harvest
3. System reverts plant slots to Seed eG pool if no action during grace period
4. All post-harvest actions are recorded with timestamps
5. Members receive appropriate notifications about post-harvest status

**Decision Tree for Post-Harvest Actions**:
```
Membership Status:
├── Renewed → Start new growth cycle
├── Cancelled → Disable plant slots, return to Seed eG
└── In Grace Period → Temporary hold, revert to Seed eG if not renewed
```

### 5.7 Payment Integration

#### 5.7.1 Stripe Integration

**Priority**: Must Have

**User Stories**:
- As a member, I want a secure payment process so that I can safely purchase or renew memberships.
- As an administrator, I want automated payment processing so that financial transactions are handled efficiently.

**Functional Requirements**:
- System must integrate with Stripe for fiat currency transactions
- System must handle payments for:
  - Initial membership purchase
  - Membership renewals
  - Harvesting and processing fees

**Acceptance Criteria**:
1. Stripe integration works correctly for all payment types
2. Members can securely provide payment information
3. System handles successful payment confirmations appropriately
4. System handles payment failures with proper error messaging
5. System maintains comprehensive payment records

**API Integration Specifications**:
- Integration must use Stripe API v2023-10-16 or later
- System must implement Stripe Elements for secure card collection
- Payment intent confirmation flow must be implemented
- Webhook endpoints must be established for asynchronous notifications
- Compliance with PCI-DSS standards must be maintained

#### 5.7.2 Yield Sale Payments

**Priority**: Should Have

**User Stories**:
- As a member, I want to sell my yield to Seed eG so that I can monetize my harvest.
- As an administrator, I want a structured process for yield purchases so that we maintain consistency and compliance.

**Functional Requirements**:
- System must implement a request and confirmation process for yield sales to Seed eG
- System must record confirmation from Seed eG
- Actual payments for yield sales will be handled via bank transfer

**Acceptance Criteria**:
1. Members can submit yield sale requests with quantity and quality information
2. Administrators can review and confirm yield sale requests
3. System generates confirmation documentation for both parties
4. System tracks status of yield sale from request to completion
5. Bank transfer details are securely recorded and provided

**Yield Sale Process Flow**:
```
[Member Submits Sale Request] → [Administrator Reviews Request] → [Sale Terms Confirmed] → [Yield Transferred] → [Payment Processed via Bank Transfer] → [Transaction Recorded]
```

**Required Data for Yield Sales**:
- Harvest date and plant slot ID string
- Quantity in grams string
- Quality rating string (if applicable)
- Agreed price per gram string
- Total sale amount string
- Member bank details for transfer

#### 5.7.3 Transaction History

**Priority**: Must Have

**User Stories**:
- As a member, I want to view my complete transaction history so that I can track my financial activity.
- As an administrator, I want comprehensive transaction records so that we maintain financial transparency.

**Functional Requirements**:
- System must maintain a comprehensive transaction history showing:
  - Membership purchases
  - Membership renewals
  - Harvesting and processing fees
  - Yield sales to Seed eG

**Acceptance Criteria**:
1. Members can view their complete transaction history
2. Administrators can access transaction records for all members
3. Transaction records include all required details (date, amount, type, status)
4. System provides filtering and search capabilities for transaction records
5. Transaction data can be exported for accounting purposes

**Reporting Requirements**:
- Monthly financial summaries by transaction type
- Quarterly revenue reports
- Annual financial statements
- Tax documentation preparation support
- Regulatory compliance reporting

#### 5.7.4 Payment Error Handling

**Priority**: Must Have

**User Stories**:
- As a member, I want clear information when payments fail so that I can resolve issues quickly.
- As an administrator, I want comprehensive payment failure data so that we can address systemic issues.

**Functional Requirements**:
- System must implement robust error handling for payment failures
- System must provide clear, actionable error messages to users
- System must support payment retry with suggested fixes
- System must implement payment timeout handling
- System must notify administrators of repeated payment failures
- System must maintain comprehensive payment attempt logs

**Acceptance Criteria**:
1. Users receive specific, actionable error messages for payment failures
2. Payment retry functionality works correctly after error resolution
3. System handles network timeouts gracefully
4. Administrators receive alerts for unusual payment failure patterns
5. Payment logs include complete error details for troubleshooting

**Refund Process**:
- System must support full and partial refunds
- Refund requests require administrative approval
- System must maintain audit trail of all refund transactions
- System must generate appropriate notifications for refund status

### 5.8 User Profile Management

#### 5.8.1 Account Information

**Priority**: Must Have

**User Stories**:
- As a member, I want to manage my profile information so that my details remain current.
- As a member, I want to see my verification status so that I know if additional steps are needed.

**Functional Requirements**:
- System must allow users to view and update profile information (within eKYC constraints)
- System must display eKYC verification status string
- System must maintain user contact information for notifications

**Acceptance Criteria**:
1. Members can view their complete profile information
2. Members can update allowed profile fields
3. System prevents updates to eKYC-verified information
4. System clearly displays verification status
5. Contact information changes trigger verification of new details

**Profile Data Categories**:
- Personal Information (name, date of birth, etc.) - View only after verification
- Contact Information (email, phone, address) - Updateable with verification
- Preferences (notification settings, display options) - Freely updateable
- Security Settings (password, 2FA) - Updateable with authentication

#### 5.8.2 Membership Information

**Priority**: Must Have

**User Stories**:
- As a member, I want to view my membership details so that I understand my current status and options.
- As a member, I want clear information about expiration dates so that I can plan renewals accordingly.

**Functional Requirements**:
- System must display membership details including:
  - Membership ID string
  - Plant slot information
  - Owner information
  - Expiration date string
  - Renewal options

**Acceptance Criteria**:
1. Members can view all membership details in a clear format
2. System prominently displays expiration date and renewal countdown
3. System presents applicable renewal options based on current status
4. Membership history is available for review
5. Members can access membership documentation and receipts

**Membership Dashboard Requirements**:
- Visual indicator of membership status string ("active", "grace period", "expired")
- Timeline showing key dates (activation, renewal, expiration)
- Quick access to renewal functionality
- Summary of associated plant slots and their status
- Notification center for membership-related alerts

#### 5.8.3 Plant Slot Information

**Priority**: Must Have

**User Stories**:
- As a member, I want detailed information about my plant slots so that I can track my cultivation assets.
- As a member, I want to see my plants' complete history so that I can understand their performance.

**Functional Requirements**:
- System must display detailed plant slot information including:
  - Catalog information (season, images, description, plant species)
  - Slot ID string
  - Owner information
  - Status string
  - Current plant information
  - Historical plant information
  - Grand yield of plant slot string

**Acceptance Criteria**:
1. Members can view comprehensive information for each plant slot
2. Plant slot displays include visual representations of plants
3. Plant history is available with timeline visualization
4. Yield data is summarized with comparisons to averages
5. Plant slot NFT information is linked and viewable

**Plant Slot Visualization Requirements**:
- Grid view of all plant slots with status indicators
- Detailed view of individual plant slots with complete information
- Timeline visualization of plant lifecycle events
- Performance charts comparing yields across cycles
- Integration with NFT visualization showing ownership certificate

### 5.9 Administrator Features

#### 5.9.1 User Roles and Permissions

**Priority**: Must Have

**User Stories**:
- As a system administrator, I want role-based access control so that users have appropriate permissions.
- As a system owner, I want audit logs of administrative actions so that we maintain accountability.

**Functional Requirements**:
- System must implement role-based access control with at least three roles: "Administrator", "Staff", "Member"
- System must include specific permissions for manual eKYC verification
- System must maintain audit logs of all administrative actions

**Acceptance Criteria**:
1. System supports creation and management of roles with granular permissions
2. "Administrator" role has complete system access
3. "Staff" role has limited access based on job responsibilities
4. eKYC verification requires specific permissions
5. All administrative actions are logged with user information, timestamp, and details
6. Audit logs cannot be modified or deleted

**Permission Matrix**:

| Feature Area | Administrator | Staff | Member |
|-------------|--------------|-------|--------|
| User Management | Full Access | View Only | Self Only |
| Membership Management | Full Access | View & Update | Self Only |
| Plant Slot Management | Full Access | View & Update | Self Only |
| eKYC Verification | Full Access | With Permission | None |
| Harvest Management | Full Access | View & Update | Request Only |
| System Configuration | Full Access | None | None |
| Reporting | Full Access | Limited Access | Self Only |

#### 5.9.2 Catalog Management

**Priority**: Must Have

**User Stories**:
- As an administrator, I want to create and manage plant slot catalogs so that we can organize slots by season.
- As an administrator, I want to associate plant species and attributes with catalogs so that members have complete information.

**Functional Requirements**:
- System must allow administrators to create and manage plant slot catalogs including:
  - Catalog ID string
  - Season name (with "1000" NFT slots per season)
  - Images
  - Description
  - Plant species
  - Slot status string

**Acceptance Criteria**:
1. Administrators can create new catalogs with all required information
2. Administrators can update existing catalog information
3. Administrators can add and manage plant species information
4. System enforces the "1000" NFT limit per season
5. Catalog information is properly displayed to members

**Catalog Structure**:
- Each catalog represents one growing season
- Catalogs contain "1000" uniquely identified plant slots
- Each plant slot is linked to specific plant species information
- Catalogs include descriptive information and visual assets
- Status tracking for each catalog ("planning", "active", "archived")

#### 5.9.3 Plant Slot Administration

**Priority**: Must Have

**User Stories**:
- As an administrator, I want to manage all aspects of plant slots so that we maintain accurate records and availability.
- As an administrator, I want to track the complete history of each plant slot so that we have full traceability.

**Functional Requirements**:
- System must enable administrators to:
  - Create new seasons with "1000" plant slots each
  - Pre-mint "1000" NFTs for each season
  - Assign slots to catalogs
  - Update slot information
  - Track ownership history
  - Manage plant information
  - Record harvest history

**Acceptance Criteria**:
1. Administrators can create new seasons with all required slots
2. System supports pre-minting of "1000" NFTs per season
3. Administrators can assign slots to appropriate catalogs
4. System maintains complete history for each plant slot
5. Administrators can update slot information as needed
6. Plant and harvest information is properly recorded and accessible

**Bulk Operations**:
- Ability to create multiple plant slots with sequential IDs
- Batch update capabilities for slot assignments
- Mass state change operations with appropriate safeguards
- Bulk NFT minting with gas optimization strategies
- Import/export functionality for plant slot data

#### 5.9.4 eKYC Verification Management

**Priority**: Must Have

**User Stories**:
- As a verification administrator, I want to efficiently review verification requests so that members can be approved quickly.
- As a compliance officer, I want comprehensive verification records so that we maintain regulatory compliance.

**Functional Requirements**:
- System must provide an interface for administrators to:
  - View pending eKYC verification requests
  - Review uploaded identification documents
  - Compare portrait photos with ID documents
  - Approve or reject verification requests
  - Add notes to verification records
  - Generate reports on verification activities
  - Track verification completion times

**Acceptance Criteria**:
1. Administrators can access a queue of pending verification requests
2. System provides secure viewing of identity documents
3. Comparison tools facilitate portrait matching with ID photos
4. Administrators can approve or reject with required notes
5. System tracks verification metrics (completion time, approval rate)
6. Complete verification records are maintained for compliance

**Verification Workflow**:
```
[Member Submits Documents] → [Request Enters Queue] → [Administrator Reviews] → [Decision (Approve/Reject)] → [Member Notification] → [Status Update] → [Reporting]
```

**Required Verification Metrics**:
- Average verification completion time
- Verification approval rate
- Rejection reasons categorized
- Verification backlog size
- Verification volume trends

#### 5.9.5 System Configuration

**Priority**: Should Have

**User Stories**:
- As an administrator, I want to configure system settings so that the platform operates according to business requirements.
- As an administrator, I want to manage notification templates so that communications are consistent and appropriate.

**Functional Requirements**:
- System must allow configuration of notification settings
- System must support email notification configuration for reminders

**Acceptance Criteria**:
1. Administrators can configure system-wide settings
2. Email notification templates can be customized
3. Notification scheduling can be configured
4. System retains configuration history
5. Configuration changes are properly logged

**Configurable Parameters**:
- Notification timing (days before expiration for reminders)
- Email templates for various system events
- Grace period duration string
- Default plant lifecycle durations by species
- System timeout and security settings
- Trading fee percentages and minimums

### 5.10 Cannabis Club Compliance

#### 5.10.1 Membership Management

**Priority**: Must Have

**User Stories**:
- As a compliance officer, I want to ensure the platform adheres to German cannabis club regulations so that we maintain legal operation.
- As an administrator, I want automated compliance controls so that we prevent regulatory violations.

**Functional Requirements**:
- System must enforce a maximum of "500" members as per German cannabis club regulations
- System must maintain a waiting list when membership capacity is reached
- System must implement age verification to ensure all members are at least "18" years old
- System must track member activity to ensure compliance with distribution limits
- System must maintain comprehensive member records as required by German regulations

**Acceptance Criteria**:
1. System prevents new memberships when "500" member limit is reached
2. Waiting list functionality activates automatically when capacity is reached
3. Age verification prevents registration of users under "18" years
4. System tracks all distribution activities with amounts and dates
5. Member records meet or exceed regulatory requirements
6. Compliance reports can be generated on demand

**Regulatory Data Retention Requirements**:
- Member identification records: "7" years
- Transaction records: "7" years
- Distribution records: "7" years
- Cultivation records: "5" years
- Verification documentation: "7" years

#### 5.10.2 Cultivation Compliance

**Priority**: Must Have

**User Stories**:
- As a compliance officer, I want to ensure cultivation activities comply with regulations so that we avoid penalties.
- As an administrator, I want to document any excess cannabis destruction so that we maintain transparent compliance.

**Functional Requirements**:
- System must track and limit cultivation to only what is necessary for member consumption
- System must implement inventory controls to prevent excess production
- System must document destruction of any excess cannabis
- System must ensure total plant count complies with licensing requirements
- System must maintain records of all cultivation activities for regulatory reporting

**Acceptance Criteria**:
1. System enforces cultivation limits based on membership numbers
2. Inventory controls prevent excess production
3. System provides documentation workflow for excess destruction
4. Plant count reporting is accurate and current
5. Cultivation records are comprehensive and compliant
6. Regulatory reports can be generated on demand

**Cultivation Compliance Calculations**:
- Maximum plants per member: "3"
- Maximum yield per plant: [Variable by plant type]
- Maximum monthly allocation per member: [As defined by regulations]
- Monitoring and alerts for approaching limits

#### 5.10.3 Distribution Compliance

**Priority**: Must Have

**User Stories**:
- As a compliance officer, I want to track all distribution to ensure it complies with personal possession limits so that members stay within legal boundaries.
- As an administrator, I want systems to prevent distribution beyond legal limits so that we maintain regulatory compliance.

**Functional Requirements**:
- System must track all distribution to ensure compliance with personal possession limits:
  - Maximum "25" grams per person in public
  - Maximum "50" grams per person in private
- System must enforce cooling-off periods between distributions if necessary
- System must maintain detailed distribution records
- System must prevent on-site consumption within the facility or within "200" meters

**Acceptance Criteria**:
1. System prevents distributions exceeding "25" grams for public possession
2. System prevents distributions exceeding "50" grams for private possession
3. System tracks cumulative distribution to prevent exceeding limits
4. Distribution records include date, amount, member information, and confirmation
5. System implements configurable cooling-off periods between distributions
6. Notices about prohibited on-site consumption are displayed prominently

**Distribution Tracking Features**:
- Running total of recent distributions per member
- Countdown to next available distribution date
- Automatic alerts for unusual distribution patterns
- Verification requirement for large distributions
- Compliance officer review for edge cases

#### 5.10.4 Packaging and Labeling

**Priority**: Must Have

**User Stories**:
- As a packaging manager, I want to generate compliant packaging specifications so that our products meet regulatory requirements.
- As a compliance officer, I want to ensure all products include required information so that we adhere to labeling regulations.

**Functional Requirements**:
- System must generate compliant neutral packaging specifications
- System must create standardized information leaflets including:
  - Weight in grams string
  - Harvest date string
  - Best before date string
  - Cannabis variety string
  - Average THC content string
  - Average CBD content string
  - Average percentage of CBD string
- System must track package information for each distribution
- System must maintain package design templates that comply with German regulations

**Acceptance Criteria**:
1. System generates neutral packaging specifications that comply with regulations
2. Information leaflets include all required details
3. Package information is tracked for each distribution
4. Package design templates are compliant with German regulations
5. System prevents use of non-compliant packaging designs

**Packaging Template Requirements**:
- Neutral design without brand elements
- Standardized information placement
- QR code linking to verification system
- Tamper-evident features
- Child-resistant packaging compliance
- Material specifications meeting environmental regulations

#### 5.10.5 Advertising and Marketing Restrictions

**Priority**: Must Have

**User Stories**:
- As a compliance officer, I want to ensure the platform doesn't include advertising functionality so that we adhere to German cannabis club regulations.
- As an administrator, I want clear guidelines on allowed communications so that we maintain compliant outreach.

**Functional Requirements**:
- System must not include any advertising functionality for the club or products
- System must implement neutral design elements throughout the platform
- System must not facilitate sponsorships or promotional activities
- System must ensure all communications are informational rather than promotional
- System must include appropriate disclaimers on all public-facing content

**Acceptance Criteria**:
1. Platform contains no advertising functionality
2. Design elements remain neutral and non-promotional
3. No sponsorship or promotional features exist
4. All communications are strictly informational
5. Required disclaimers appear on all public-facing content

**Content Guidelines**:
- Factual, educational information only
- No promotional language or imagery
- No celebrity endorsements or testimonials
- No discount or special offer promotions
- Clear separation between informational and commercial content

## 6. Phase 2 - Advanced: Marketplace Integration

### 6.1 Membership Trading Marketplace

#### 6.1.1 Marketplace Platform

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a member, I want to browse available memberships for purchase so that I can find options that meet my needs.
- As a member, I want to compare different membership offerings so that I can make an informed decision.

**Functional Requirements**:
- System must provide a dedicated marketplace interface for trading memberships
- System must implement a responsive design that works across devices
- System must include search and filtering capabilities for available memberships
- System must display detailed membership information to potential buyers
- System must provide comparison tools for different membership offerings
- System must implement a watchlist feature for tracking interesting listings
- System must offer market analytics to help users make informed decisions
- System must display associated NFTs for each membership offering
- System must show the connection between plant slot NFTs and potential harvest product NFTs

**Acceptance Criteria**:
1. Marketplace interface provides intuitive browsing experience
2. Search and filtering functions work effectively across various criteria
3. Detailed membership information is clearly displayed
4. Comparison tools allow side-by-side evaluation of offerings
5. Watchlist functionality works correctly for tracking listings
6. Market analytics provide valuable insights for decision-making
7. NFT information is properly displayed and explained

**User Experience Requirements**:
- Mobile-first responsive design
- Intuitive navigation and discovery
- Advanced search with multiple filters
- Saved search functionality
- Personalized recommendations based on preferences
- Interactive comparison tools with visual elements
- Accessibility compliance (WCAG 2.1 AA)

#### 6.1.2 Listing Management

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a seller, I want to create detailed listings for my membership so that I can attract potential buyers.
- As a seller, I want to modify my listings as needed so that I can adjust to market conditions.

**Functional Requirements**:
- System must allow members to create listings for their memberships
- System must capture comprehensive listing details:
  - Price and payment terms string
  - Membership details (activation date, duration, etc.)
  - Plant slot information
  - Current plant states and histories
  - Listing duration and expiration string
- System must support listing modification and cancellation
- System must implement verification rules to ensure only valid memberships can be listed
- System must provide listing templates and guidance for sellers
- System must include promotional features for highlighted listings
- System must implement a review system for both buyers and sellers

**Acceptance Criteria**:
1. Members can create comprehensive listings with all required information
2. System captures and displays all necessary details for informed purchase decisions
3. Sellers can modify or cancel listings before purchase
4. System verifies eligibility of memberships before allowing listing
5. Listing templates simplify the creation process
6. Optional promotional features enhance listing visibility
7. Review system builds trust between buyers and sellers

**Listing Data Requirements**:
- Membership details (activation date string, duration string, remaining time string)
- Plant slot information (IDs, location, plant types)
- Current plant states (growth stage string, health string, cycle number string)
- Historical performance (past yields string, care history)
- Pricing details (fixed price string or auction format)
- Payment terms and conditions
- Transfer timeline and process

#### 6.1.3 Transaction Processing

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a buyer, I want secure payment processing so that I can purchase memberships safely.
- As a seller, I want guaranteed payment through escrow so that I can transfer ownership confidently.

**Functional Requirements**:
- System must provide secure payment processing for membership trades
- System must implement an escrow mechanism to protect both buyers and sellers
- System must handle the complete transfer of membership, plant slots, and associated NFTs
- System must maintain detailed transaction records and history
- System must generate receipts and confirmation notifications
- System must enforce cooling-off periods for transaction finalization
- System must provide transaction status tracking for all parties
- System must implement secure messaging between transaction participants
- System must handle partial and installment payments when applicable

**Acceptance Criteria**:
1. Payment processing is secure and reliable
2. Escrow mechanism protects both buyer and seller interests
3. Ownership transfer process is comprehensive and accurate
4. Transaction records are complete and accessible
5. All parties receive appropriate notifications throughout the process
6. Cooling-off periods are properly enforced
7. Transaction status is clearly visible to all participants
8. Secure messaging facilitates communication between parties
9. Alternative payment structures work correctly when applicable

**Escrow Mechanism Details**:
- Funds held in escrow until transfer conditions are met
- Clear escrow terms presented before transaction
- Multi-step verification process for release conditions
- Dispute resolution procedure for escrow conflicts
- Automatic and manual release options
- Complete audit trail of escrow status changes
- Compliance with financial regulations

#### 6.1.4 Dispute Resolution

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a marketplace user, I want access to dispute resolution so that I can resolve conflicts fairly.
- As an administrator, I want structured dispute processes so that we can efficiently address issues.

**Functional Requirements**:
- System must include a mechanism for reporting and resolving disputes
- System must provide administrator tools for managing disputes
- System must implement a structured resolution process with defined timeframes
- System must maintain complete records of all dispute communications
- System must offer mediation services for complex disputes
- System must implement automated resolution for common dispute types
- System must provide templates for dispute documentation
- System must track dispute resolution metrics

**Acceptance Criteria**:
1. Users can report disputes through clear channels
2. Administrators have effective tools for dispute management
3. Resolution process follows structured timeline with appropriate notifications
4. All communications are recorded for reference
5. Mediation services effectively address complex disputes
6. Common disputes are resolved efficiently through automated processes
7. Documentation templates streamline the dispute process
8. System tracks and reports on dispute metrics

**Dispute Resolution Process**:
```
[Dispute Reported] → [Initial Assessment] → [Evidence Collection] → [Administrator Review] → [Proposed Resolution] → [Acceptance or Appeal] → [Implementation] → [Closure]
```

**Dispute Categories and SLAs**:
- Payment disputes: "48" hour initial response
- Membership transfer issues: "24" hour initial response
- Plant condition misrepresentation: "72" hour initial response
- Technical platform issues: "24" hour initial response
- Terms of service violations: "48" hour initial response

### 6.2 Product Marketplace

#### 6.2.1 Harvested Product to Marketplace Integration

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a member, I want a seamless connection between my harvests and the marketplace so that I can easily sell my products.
- As a buyer, I want transparent product provenance so that I can trust what I'm purchasing.

**Functional Requirements**:
- System must establish a direct connection between harvested products and marketplace listings
- System must implement a product transformation process:
  - Track harvested raw materials from specific plant slots
  - Record processing steps and quality checks
  - Generate unique product batch IDs linked to source plants
  - Maintain complete chain of custody documentation
- System must automatically calculate available inventory based on verified harvest records
- System must enforce regulatory compliance throughout the product journey
- System must implement quality grading standards for harvested products
- System must maintain detailed history of plant care that produced each product batch
- System must generate a unique NFT for each packaged harvested product with the following characteristics:
  - Each harvest package is represented by an NFT on the blockchain
  - NFTs contain verifiable information about the product's origin, strain, and quality
  - NFTs can be transferred between members within the platform
  - NFTs serve as certificates of authenticity for the product
  - NFTs maintain a complete history of the product from seed to sale
- System must ensure that "1000" NFT plant slots can generate a variable number of harvest product NFTs based on yield

**Acceptance Criteria**:
1. Harvested products are seamlessly connected to marketplace listings
2. Product transformation process maintains complete traceability
3. Inventory calculations accurately reflect verified harvest records
4. Regulatory compliance is enforced throughout the product journey
5. Quality grading standards are consistently applied
6. Plant care history is accessible and linked to products
7. NFT generation works correctly for each packaged product
8. Variable NFT generation based on yield functions properly

**Blockchain Integration Requirements**:
- Smart contract for harvest product NFT minting
- Metadata structure that includes:
  - Source plant slot ID string
  - Harvest date and batch string
  - THC/CBD content analysis string
  - Processing method string
  - Quality grade string
  - Chain of custody hash string
- Verifiable link between plant slot NFT and harvest product NFT
- Gas-optimized batch minting for efficiency

#### 6.2.2 Product Catalog and Packaging

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a seller, I want to create appealing product listings so that I can effectively market my harvest.
- As a buyer, I want detailed product information so that I can make informed purchase decisions.

**Functional Requirements**:
- System must enable members to list harvested products for sale
- System must capture detailed product information:
  - Product type and variety string
  - Source plant slot IDs and harvest dates
  - Cultivation method and conditions string
  - Processing method and date string
  - Quality metrics and certifications
  - Images and descriptions
- System must support flexible packaging configurations:
  - Ability to create standard package sizes ("1g", "3.5g", "7g", "14g", "25g")
  - Custom packaging options with minimum/maximum constraints
  - Maximum package size of "25g" for public transactions per German law
  - Maximum package size of "50g" for private delivery per German law
- System must implement product categories and tagging for improved discoverability
- System must support product review and rating features
- System must enforce neutral packaging requirements per German cannabis club regulations
- System must automatically generate compliant information leaflets containing:
  - Weight in grams string
  - Harvest date string
  - Best before date string
  - Cannabis variety string
  - Average THC and CBD content string
  - Average percentage of CBD string

**Acceptance Criteria**:
1. Members can create comprehensive product listings
2. Product details are accurately captured and displayed
3. Packaging configuration options work correctly with appropriate constraints
4. Product categorization and tagging improves discovery
5. Review and rating features work effectively
6. Neutral packaging requirements are properly enforced
7. Information leaflets are automatically generated with all required details

**Product Catalog Structure**:
- Hierarchical categories and subcategories
- Multiple tagging dimensions (effect, flavor, cultivation method)
- Advanced search with filtering by multiple attributes
- Featured listings and new arrival sections
- Personalized recommendations based on preferences
- Detailed product pages with comprehensive information
- Educational content related to product types

#### 6.2.3 Pricing and Inventory Management

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a seller, I want flexible pricing options so that I can maximize the value of my products.
- As an administrator, I want inventory controls so that we prevent overselling and maintain compliance.

**Functional Requirements**:
- System must provide multiple pricing models:
  - Fixed price per package with quantity discounts
  - Weight-based pricing with minimum purchase requirements
  - Subscription-based recurring purchases with preferred pricing
  - Limited-time promotional pricing
  - Wholesale and bulk discount tiers
- System must implement a pricing calculator with the following features:
  - Base price setting per gram or unit string
  - Quality grade multipliers string
  - Packaging cost inclusion string
  - Tax calculation string
  - Handling fee addition string
  - Promotional discount application string
- System must maintain real-time inventory tracking:
  - Available product by weight and package size string
  - Reserved product for pending orders string
  - Low inventory alerts and thresholds
  - Automatic listing deactivation when inventory depleted
- System must prevent overselling through inventory locks during checkout
- System must enforce German legal purchase limits:
  - Maximum "25" grams per public purchase transaction
  - Maximum "50" grams for private delivery
  - Cumulative purchase tracking across multiple orders
- System must implement cooling-off periods between large purchases
- System must alert administrators about suspicious purchasing patterns
- System must correlate NFT metadata with physical inventory to ensure accuracy
- System must track NFT transfers between members and update inventory accordingly

**Acceptance Criteria**:
1. Multiple pricing models function correctly
2. Pricing calculator accurately determines final prices
3. Real-time inventory tracking prevents overselling
4. Purchase limits are properly enforced
5. Cooling-off periods function as designed
6. Suspicious purchase patterns trigger appropriate alerts
7. NFT metadata and physical inventory remain synchronized
8. Inventory updates correctly after NFT transfers

**Inventory Management Process Flow**:
```
[Harvest Recorded] → [Products Created] → [Inventory Added] → [Listings Created] → [Orders Placed] → [Inventory Reserved] → [Order Completed] → [Inventory Deducted]
```

**Compliance Checks**:
- Real-time verification against purchase limits
- Cooling-off period enforcement between purchases
- Cumulative purchase tracking across time periods
- Verification of buyer age and membership status
- Logging of all inventory transactions for auditing
- Automatic reporting of suspicious activities

#### 6.2.4 Order Management

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a buyer, I want a smooth order process so that I can complete purchases easily.
- As a seller, I want comprehensive order management so that I can fulfill orders efficiently.

**Functional Requirements**:
- System must provide a complete order processing system
- System must handle various payment methods and currencies
- System must track order status from placement to fulfillment
- System must generate invoices and receipts
- System must support order modifications and cancellations
- System must implement inventory management to prevent overselling
- System must enforce purchase limits based on German regulatory requirements:
  - Maximum "25" grams per public transaction
  - Maximum "50" grams for private delivery

**Acceptance Criteria**:
1. Order processing works smoothly from placement to completion
2. Multiple payment methods are supported
3. Order status tracking is accurate and informative
4. Invoices and receipts are correctly generated
5. Order modifications and cancellations work as expected
6. Inventory management prevents overselling
7. Purchase limits are properly enforced

**Order Status Tracking**:
- "Order placed"
- "Payment processing"
- "Payment confirmed"
- "Order processing"
- "Ready for pickup/delivery"
- "In transit" (if applicable)
- "Delivered/picked up"
- "Completed"
- "Cancelled" (with reason)

**Order Processing Flow**:
```
[Cart Creation] → [Checkout] → [Address Verification] → [Payment Processing] → [Order Confirmation] → [Inventory Allocation] → [Fulfillment] → [Delivery/Pickup] → [Completion]
```

#### 6.2.5 Shipping and Fulfillment

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a buyer, I want flexible delivery options so that I can receive products conveniently.
- As a seller, I want efficient fulfillment processes so that I can deliver products promptly.

**Functional Requirements**:
- System must calculate shipping costs based on product weight and destination
- System must generate shipping labels and documentation
- System must provide order tracking capabilities
- System must support multiple fulfillment methods:
  - Standard shipping with tracking
  - Express shipping options
  - Local pickup at designated locations
  - Scheduled delivery windows
- System must implement delivery confirmation and proof of delivery
- System must ensure regulatory compliance for product transportation
- System must support discreet packaging options
- System must implement chain of custody tracking throughout shipping

**Acceptance Criteria**:
1. Shipping cost calculation is accurate
2. Shipping labels and documentation are correctly generated
3. Order tracking provides accurate status information
4. Multiple fulfillment methods are supported
5. Delivery confirmation works reliably
6. Transportation complies with all regulatory requirements
7. Discreet packaging options are available
8. Chain of custody is maintained throughout the shipping process

**Shipping Integration Requirements**:
- API integration with selected shipping providers
- Label generation capabilities
- Tracking number assignment and management
- Automated status updates from carriers
- Exception handling for delivery issues
- Address verification and correction
- Packaging material specifications
- Route optimization for local deliveries

#### 6.2.6 Reporting and Analytics

**Priority**: Should Have (Phase 2)

**User Stories**:
- As a seller, I want sales analytics so that I can optimize my product offerings.
- As an administrator, I want comprehensive marketplace data so that we can improve platform performance.

**Functional Requirements**:
- System must provide comprehensive sales analytics for sellers
- System must generate detailed reports on marketplace performance
- System must track key metrics such as conversion rates and customer acquisition costs
- System must support export of data for external analysis
- System must provide harvest-to-sale tracking for regulatory compliance
- System must generate product performance reports:
  - Best-selling products and packages
  - Customer preferences and feedback
  - Pricing optimization recommendations
  - Inventory turnover metrics
- System must implement forecasting tools for future harvests and sales

**Acceptance Criteria**:
1. Sales analytics provide valuable insights for sellers
2. Marketplace performance reports are comprehensive and accurate
3. Key metrics are tracked and displayed effectively
4. Data export works correctly in common formats
5. Harvest-to-sale tracking meets regulatory requirements
6. Product performance reports help optimize offerings
7. Forecasting tools provide useful projections

**Dashboard Requirements**:
- Real-time sales and revenue metrics
- Conversion funnels and drop-off analysis
- Inventory forecasting based on sales velocity
- Seasonal trend identification
- Customer segmentation analysis
- Competitive pricing analysis
- Product performance comparisons
- Custom report generation capabilities
- Scheduled report delivery options

**Data Export Formats**:
- CSV for spreadsheet analysis
- JSON for programmatic integration
- PDF for formal reporting
- Excel for advanced analysis

## 7. Phase 3 - Social Network: Community Features

### 7.1 User Profiles and Connections

#### 7.1.1 Profile Management

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want a customizable profile so that I can control my community presence.
- As a member, I want to showcase my cultivation achievements so that I can build reputation.

**Functional Requirements**:
- System must provide comprehensive user profile capabilities:
  - Public profile with customizable visibility settings
  - Profile photo and cover image support
  - Bio and personal information fields
  - Showcase for memberships and achievements
  - Activity history and statistics
- System must implement profile verification badges for trusted members
- System must support profile editing and content management

**Acceptance Criteria**:
1. Members can create and customize public profiles
2. Visibility settings function correctly to protect privacy
3. Profile elements display appropriately across devices
4. Membership and achievement showcases work correctly
5. Activity history accurately reflects user participation
6. Verification badges are properly assigned and displayed
7. Profile editing maintains user control over content

**Profile Element Requirements**:
- Required elements: Username string, avatar, join date string
- Optional elements: Bio string, location string (area only), interests string, expertise string
- Showcase elements: Cultivation badges, harvest showcases, reputation score string
- Activity elements: Contribution count string, helpful ratings string, community roles string
- Privacy options: "Public", "members-only", "connections-only", "private"

#### 7.1.2 Connection Management

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want to connect with other members so that I can build my cultivation network.
- As a member, I want to control who can see my activity so that I maintain appropriate privacy.

**Functional Requirements**:
- System must enable members to follow other members
- System must implement a connection request and approval process
- System must provide tools for managing connections and followers
- System must include privacy controls for connection visibility
- System must implement blocking and reporting features

**Acceptance Criteria**:
1. Members can follow others to receive updates
2. Connection requests and approvals function correctly
3. Connection management tools are intuitive and effective
4. Privacy controls work correctly for all visibility settings
5. Blocking features effectively prevent unwanted interactions
6. Reporting system allows flagging of inappropriate behavior

**Connection Types**:
- "Followers": One-way relationship, see public updates
- "Connections": Two-way relationship, see connection-level content
- "Mentors/Mentees": Specialized relationship for knowledge sharing
- "Blocked": Prevents all interaction and visibility

**Privacy Controls by Content Type**:
- Profile information: Controllable by field
- Posts: "Public", "connections", "groups", or "private"
- Questions and answers: "Public" or "connections"
- Showcases: "Public", "connections", or "private"
- Activity: "Public", "connections", or "private"

#### 7.1.3 Discovery Features

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want to discover like-minded cultivators so that I can expand my knowledge.
- As a new member, I want to find active community members so that I can learn from experienced users.

**Functional Requirements**:
- System must provide member discovery tools based on interests and activity
- System must suggest potential connections based on relevant criteria
- System must highlight active and influential community members
- System must implement a reputation system to recognize valuable contributors

**Acceptance Criteria**:
1. Discovery tools effectively connect members with similar interests
2. Connection suggestions are relevant and helpful
3. Active community members are appropriately highlighted
4. Reputation system accurately reflects valuable contributions
5. Discovery features respect privacy settings

**Discovery Algorithms**:
- Interest-based matching from profile information
- Activity pattern similarity analysis
- Geographic proximity (region level only)
- Expertise complementarity identification
- Cultivation style and preference matching

**Reputation System Components**:
- Knowledge sharing contributions
- Community support activities
- Content quality ratings
- Longevity and consistency of participation
- Verification level achievements
- Special recognitions and badges

### 7.2 Content Creation and Sharing

#### 7.2.1 Post Creation

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want to create and share content so that I can contribute to the community.
- As a cultivation expert, I want to share rich media posts so that I can better illustrate techniques.

**Functional Requirements**:
- System must provide a rich text editor for creating posts
- System must support various content types:
  - Text posts with formatting
  - Image galleries with captions
  - Video content with preview thumbnails
  - Links with rich previews
  - Polls and interactive content
- System must implement post scheduling capabilities
- System must support drafts and content revision history

**Acceptance Criteria**:
1. Rich text editor provides appropriate formatting options
2. Multiple content types are supported with proper rendering
3. Image galleries function correctly with captions
4. Video content includes proper preview thumbnails
5. Link sharing generates useful rich previews
6. Polls and interactive content work as expected
7. Post scheduling functions reliably
8. Draft saving and revision history work correctly

**Content Creation Features**:
- Text formatting (bold, italic, headings, lists, etc.)
- Image upload with compression and optimization
- Gallery creation with ordering and captioning
- Video upload with thumbnail generation
- Link preview generation
- Poll creation with multiple question types
- Draft saving with auto-recovery
- Content scheduling with time zone support

#### 7.2.2 Content Organization

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want to organize content with hashtags so that topics are easily discoverable.
- As a content creator, I want to create collections so that related content stays together.

**Functional Requirements**:
- System must implement hashtags for content categorization
- System must support user mentions and notifications
- System must provide content collections and saving features
- System must implement topic-based groups and communities
- System must support content series and sequences

**Acceptance Criteria**:
1. Hashtags function correctly for content categorization
2. User mentions trigger appropriate notifications
3. Content collections allow effective organization
4. Topic-based groups facilitate focused discussions
5. Content series enable sequential learning experiences

**Content Organization Methods**:
- Hashtag system with trending and followed tags
- User mention system with notifications
- Personal collections for saved content
- Group creation and management
- Series creation for sequential content
- Automated categorization suggestions

#### 7.2.3 Sharing and Distribution

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want to share valuable content within the community so that others can benefit.
- As a creator, I want to control who can reshare my content so that I maintain appropriate privacy.

**Functional Requirements**:
- System must enable sharing of posts within the platform
- System must implement resharing with added comments
- System must provide external sharing functionality
- System must offer content embedding capabilities
- System must support private sharing with specific members

**Acceptance Criteria**:
1. Content sharing works reliably within the platform
2. Resharing with comments functions correctly
3. External sharing provides appropriate links and previews
4. Content embedding works within supported contexts
5. Private sharing respects specified recipient limitations

**Sharing Control Options**:
- Allow/disallow resharing
- Allow/disallow external sharing
- Specify allowed audiences for resharing
- Attribution requirements
- Time limits on sharing availability

### 7.3 Engagement and Interaction

#### 7.3.1 Reactions and Feedback

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want to react to content in multiple ways so that I can express nuanced feedback.
- As a creator, I want to see detailed engagement metrics so that I can improve my content.

**Functional Requirements**:
- System must provide multiple reaction types (beyond simple likes)
- System must implement commenting on posts and other content
- System must support threaded comment discussions
- System must enable rich media in comments and reactions
- System must provide analytics on engagement metrics

**Acceptance Criteria**:
1. Multiple reaction types function correctly
2. Commenting system works reliably across content types
3. Threaded comments support in-depth discussions
4. Rich media can be included in comments
5. Engagement analytics provide useful insights for creators

**Reaction Types**:
- "Appreciation" (like/upvote)
- "Educational" (learned something)
- "Helpful" (practical value)
- "Inspiring" (motivational)
- "Question" (need clarification)
- Custom reactions for special events

**Comment Features**:
- Image and link inclusion
- Threaded replies (multiple levels)
- Mentions and notifications
- Sorting options ("newest", "top", "controversial")

#### 7.3.2 Messaging and Communication

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want private messaging capabilities so that I can have direct conversations.
- As a community leader, I want to create group conversations so that we can collaborate effectively.

**Functional Requirements**:
- System must implement private one-to-one messaging
- System must support group conversations
- System must provide message formatting and media sharing
- System must implement read receipts and typing indicators
- System must offer message search and archiving

**Acceptance Criteria**:
1. Private messaging works reliably between members
2. Group conversations support multiple participants effectively
3. Message formatting and media sharing function correctly
4. Read receipts and typing indicators provide appropriate feedback
5. Message search and archiving work efficiently

**Messaging Features**:
- Text chat with formatting
- Image and file sharing
- Read receipts (with privacy options)
- Typing indicators
- Message reactions
- Voice messages
- Message editing and deletion
- Archived chat recovery

#### 7.3.3 Notifications and Alerts

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want a comprehensive notification system so that I don't miss important updates.
- As a member, I want to control notification frequency so that I'm not overwhelmed.

**Functional Requirements**:
- System must provide a comprehensive notification system
- System must support multiple notification channels (in-app, email, push)
- System must implement notification preferences and filters
- System must include notification batching to prevent overload
- System must provide a notification history and management interface

**Acceptance Criteria**:
1. Notification system reliably delivers timely updates
2. Multiple notification channels function correctly
3. Notification preferences and filters work as expected
4. Batching prevents notification overload
5. Notification history and management interface is user-friendly

**Notification Types**:
- Content interactions (likes, comments, shares)
- Mentions and tags
- Direct messages
- Connection requests and updates
- System announcements
- Event reminders
- Membership status updates
- Plant lifecycle milestones

### 7.4 Community Governance

#### 7.4.1 Content Moderation

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a community member, I want appropriate content moderation so that the platform remains safe and valuable.
- As a community moderator, I need effective tools to manage reported content efficiently.

**Functional Requirements**:
- System must provide automated content filtering
- System must implement community reporting tools
- System must support manual review by moderators
- System must include content policy enforcement
- System must provide escalation paths for serious issues

**Acceptance Criteria**:
1. Automated content filtering detects potential violations
2. Community reporting tools function effectively
3. Moderator review interface supports efficient decisions
4. Content policy enforcement is consistent and transparent
5. Escalation paths function correctly for serious issues

**Moderation Workflow**:
```
[Content Created] → [Automated Filtering] → [Passes Filter / Flagged for Review] → [User Reports] → [Moderator Queue] → [Review Decision] → [Action Taken] → [Notification to Involved Parties]
```

**Moderation Tools**:
- Keyword and pattern filtering
- AI-assisted content analysis
- User reporting interface
- Moderator review dashboard
- Graduated response options
- Review history and consistency tracking
- Appeals management system

#### 7.4.2 Community Standards

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a member, I want clear community guidelines so that I understand expected behavior.
- As an administrator, I want transparent enforcement so that members trust the moderation process.

**Functional Requirements**:
- System must publish and enforce community guidelines
- System must implement graduated responses to violations
- System must provide educational resources on proper conduct
- System must maintain transparency in moderation actions
- System must support appeals processes for moderation decisions

**Acceptance Criteria**:
1. Community guidelines are clearly published and accessible
2. Graduated responses are consistently applied
3. Educational resources help members understand expectations
4. Moderation actions are transparent where appropriate
5. Appeals process functions fairly and efficiently

**Guidelines Documentation**:
- Written in clear, accessible language
- Includes specific examples of acceptable/unacceptable behavior
- Regularly updated based on emerging issues
- Available in multiple languages
- Includes visual elements for better comprehension
- Interactive training modules for new members

#### 7.4.3 Analytics and Insights

**Priority**: Could Have (Phase 3)

**User Stories**:
- As a community manager, I want comprehensive analytics so that I can monitor community health.
- As an administrator, I want to identify emerging issues so that we can address them proactively.

**Functional Requirements**:
- System must track community health metrics
- System must provide moderators with insight tools
- System must generate regular community status reports
- System must identify trends and emerging issues
- System must support data-driven community improvement initiatives

**Acceptance Criteria**:
1. Community health metrics are accurately tracked
2. Insight tools provide valuable information to moderators
3. Community status reports are comprehensive and timely
4. Trend analysis identifies emerging issues effectively
5. Data supports community improvement initiatives

**Community Health Metrics**:
- Active member participation rates
- New member onboarding completion
- Content quality scores
- Response times for questions
- Resolution rates for reported issues
- Member satisfaction surveys
- Retention and engagement trends
- Growth metrics by segment

## 8. Technical Requirements

### 8.1 Platform Architecture

**Priority**: Must Have

**User Stories**:
- As a developer, I want a well-designed architecture so that the system is maintainable and scalable.
- As a user, I want a responsive application so that I can access the platform from any device.

**Functional Requirements**:
- Web-based application with responsive design
- Secure authentication and authorization
- Database for storing user, membership, and plant information (all as strings)
- Integration with external services (eKYC, payment processing)
- Administrative CMS for backend management
- Blockchain integration for NFT functionality
- Mobile-optimized interfaces for field usage

**Acceptance Criteria**:
1. Architecture documentation clearly defines all components
2. Web application is responsive across all device types
3. Authentication and authorization function securely
4. Database design supports all required functionality with string-based storage
5. External service integrations function correctly
6. Administrative CMS provides required functionality
7. Blockchain integration supports NFT functionality
8. Mobile interfaces are optimized for field usage

**Architecture Components**:

1. **Frontend Layer**:
   - Responsive web application using React
   - Mobile-optimized views
   - Progressive Web App capabilities
   - Offline functionality for critical features

2. **API Layer**:
   - RESTful API with GraphQL for complex queries
   - API gateway for service orchestration
   - Rate limiting and security controls
   - Comprehensive documentation

3. **Service Layer**:
   - Microservices architecture for key domains
   - Event-driven communication
   - Caching infrastructure
   - Background processing for asynchronous tasks

4. **Data Layer**:
   - Relational database for transactional data (string-based columns)
   - Blockchain integration for NFT functionality
   - Search indexing for performance
   - Data warehousing for analytics

### 8.2 Security Requirements

**Priority**: Must Have

**User Stories**:
- As a member, I want my personal data to be secure so that I can trust the platform.
- As an administrator, I want comprehensive security measures so that we prevent unauthorized access.

**Functional Requirements**:
- Secure user authentication with multi-factor authentication support
- Data encryption for sensitive information (both at rest and in transit)
- Compliance with data protection regulations (GDPR)
- Secure API integrations with encryption and token-based authentication
- Regular security audits and penetration testing
- Comprehensive backup and disaster recovery processes
- Session management and automatic timeouts
- Detailed security logging and monitoring

**Acceptance Criteria**:
1