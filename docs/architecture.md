# Seed eG Platform Architecture Document

## Overview

This document outlines the technical architecture for the Seed eG Plant Slot Management and Membership Platform MVP. The platform leverages the existing Go-based backend infrastructure with MongoDB, Redis, and MinIO to enable cannabis club members to register, complete verification, manage memberships, and track plant slots with NFT integration while ensuring compliance with German cannabis regulations.

## System Architecture

### High-Level Architecture

The Seed eG Platform follows a modern, scalable architecture built on proven foundations:

```
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│               │      │               │      │               │
│  Client Apps  │◄────►│   API Layer   │◄────►│  Data Layer   │
│   (Vue.js)    │      │  (Go + Gin)   │      │ (MongoDB +    │
│               │      │               │      │  Redis +      │
│               │      │               │      │  MinIO)       │
└───────────────┘      └───────┬───────┘      └───────────────┘
                              │                       ▲
                              ▼                       │
                      ┌───────────────┐      ┌───────────────┐
                      │               │      │               │
                      │ Service Layer │◄────►│  Blockchain   │
                      │  (Business    │      │  Integration  │
                      │   Logic)      │      │  & External   │
                      │               │      │  Services     │
                      └───────────────┘      └───────────────┘
```

### Client Applications

- **Web Application**: 
  - Vue.js SPA with TypeScript and Vuetify
  - Google Material Design for consistent UI/UX
  - Progressive Web App capabilities for mobile use
  - Responsive design optimized for both desktop and mobile
  - WebSocket integration for real-time updates
  - Offline capabilities for field usage

- **Mobile Optimization**:
  - Service worker implementation for offline functionality
  - LocalStorage/IndexedDB for local data storage
  - Background sync for delayed operations
  - Push notifications via WebSocket

### API Layer (Built on Existing Base)

**Core Infrastructure** (from `route/index.go`):
```go
// Middleware Chain - Standard setup from base
app.Use(
    mdw.Cors(),           // CORS handling
    mdw.Compress(),       // Response compression
    mdw.Trace(),          // Request tracing
    mdw.Logger(),         // Request logging
    mdw.Recover(),        // Panic recovery
    mdw.Error(),          // Error handling
    mdw.RateLimit()       // Rate limiting
)

// Authentication Middleware - Reusing existing OAuth system
func (s middleware) BearerAuth(permissions ...enum.Permission) gin.HandlerFunc
func (s middleware) BasicAuth() gin.HandlerFunc
func (s middleware) NoAuth() gin.HandlerFunc
```

**Enhanced API Features**:
- RESTful API with Gin framework (existing)
- JWT authentication with refresh tokens (existing `pkg/oauth/index.go`)
- Permission-based authorization using existing role system
- Rate limiting for security (existing implementation)
- Comprehensive error handling with standard error codes (`pkg/ecode/index.go`)
- WebSocket support for real-time updates (existing `pkg/ws/index.go`)
- File upload/download capabilities (existing `route/storage.go`)
- Audit logging for all operations (existing audit system)

**Error Handling** (using existing `pkg/ecode/index.go`):
```go
// Standard Error Structure
type Error struct {
    Status   int    `json:"-"`
    ErrCode  string `json:"error"`
    ErrDesc  string `json:"error_description"`
    ErrStack string `json:"-"`
}

// Cannabis Club Specific Errors
var (
    KYCVerificationRequired = New(http.StatusForbidden, "kyc_verification_required")
    MembershipExpired      = New(http.StatusForbidden, "membership_expired")
    PlantSlotUnavailable   = New(http.StatusConflict, "plant_slot_unavailable")
    HarvestNotReady        = New(http.StatusConflict, "harvest_not_ready")
    NFTMintingFailed       = New(http.StatusInternalServerError, "nft_minting_failed")
    PaymentProcessingError = New(http.StatusPaymentRequired, "payment_processing_error")
)
```

### Service Layer (Extended from Base)

**Core Services** (leveraging existing patterns from `store/` and `pkg/`):

1. **AuthService** (extends existing `pkg/oauth/index.go`):
```go
type AuthService struct {
    oauth *oauth.Oauth
    store *store.Store
    mail  *mail.Mail
}

// Enhanced token generation for cannabis club members
func (s *AuthService) GenerateMemberToken(ctx context.Context, memberID string) (*db.AuthTokenDto, error) {
    // Verify member KYC status
    member, err := s.store.GetMember(ctx, memberID)
    if err != nil {
        return nil, err
    }
    
    if member.KYCStatus != "verified" {
        return nil, ecode.KYCVerificationRequired
    }
    
    // Check membership validity
    if member.CurrentMembershipID != "" {
        membership, err := s.store.GetMembership(ctx, member.CurrentMembershipID)
        if err != nil || membership.Status != "active" || membership.ExpirationDate.Before(time.Now()) {
            return nil, ecode.MembershipExpired
        }
    }
    
    return s.oauth.GenerateToken(ctx, memberID)
}
```

2. **MembershipService** (new, follows existing store patterns):
```go
type MembershipService struct {
    store   *store.Store
    payment *PaymentService
    nft     *NFTService
}

func (s *MembershipService) CreateMembership(ctx context.Context, req *CreateMembershipRequest) (*Membership, error) {
    // Validation using existing pattern
    if err := s.validateMembershipRequest(req); err != nil {
        return nil, ecode.ValidationError.Desc(err)
    }

    // Business logic
    membership := &Membership{
        MemberID:        req.MemberID,
        MembershipType:  req.MembershipType,
        Status:          "pending_payment",
        AllocatedSlots:  req.CatalogSlots,
        PaymentAmount:   req.PaymentAmount,
        TenantID:        req.TenantID,
        CreatedAt:       time.Now(),
    }

    // Store in database using existing patterns
    if err := s.store.Db.Membership.Create(ctx, membership); err != nil {
        return nil, ecode.InternalServerError.Desc(err)
    }

    // Cache invalidation using existing Redis patterns
    s.store.Rdb.Del(ctx, fmt.Sprintf("member:%s:memberships", req.MemberID))

    // Audit log using existing system
    s.auditLog(ctx, "membership", enum.DataActionCreate, membership, membership.ID)

    return membership, nil
}
```

3. **PlantSlotService** (new, following existing patterns):
```go
type PlantSlotService struct {
    store *store.Store
    nft   *NFTService
}

func (s *PlantSlotService) AllocateSlot(ctx context.Context, req *AllocateSlotRequest) (*PlantSlot, error) {
    // Use existing cache patterns
    key := fmt.Sprintf("available_slots:%s", req.TenantID)
    
    // Try to get available slots from cache first
    var availableSlots []string
    if err := s.store.Rdb.Get(ctx, key, &availableSlots); err != nil {
        // Load from database if not in cache
        slots, err := s.store.Db.PlantSlot.FindAvailable(ctx, req.TenantID)
        if err != nil {
            return nil, ecode.InternalServerError.Desc(err)
        }
        // Cache the results
        s.store.Rdb.Set(ctx, key, slots, time.Hour*1)
        availableSlots = slots
    }

    if len(availableSlots) == 0 {
        return nil, ecode.PlantSlotUnavailable
    }

    // Allocate slot and mint NFT
    slot := &PlantSlot{
        MembershipID:     req.MembershipID,
        SlotNumber:       availableSlots[0],
        Status:          "allocated",
        FarmLocation:    req.FarmLocation,
        AreaDesignation: req.AreaDesignation,
        TenantID:        req.TenantID,
        CreatedAt:       time.Now(),
    }

    // Mint NFT for the slot
    nftResult, err := s.nft.MintPlantSlotNFT(ctx, slot)
    if err != nil {
        return nil, ecode.NFTMintingFailed.Desc(err)
    }
    
    slot.NFTTokenID = nftResult.TokenID
    slot.NFTContractAddress = nftResult.ContractAddress

    // Store using existing patterns
    if err := s.store.Db.PlantSlot.Create(ctx, slot); err != nil {
        return nil, ecode.InternalServerError.Desc(err)
    }

    // Cache invalidation
    s.store.Rdb.Del(ctx, key)
    s.store.Rdb.Del(ctx, fmt.Sprintf("membership:%s:slots", req.MembershipID))

    return slot, nil
}
```

4. **PlantService** (new, using MongoDB document patterns):
```go
type PlantService struct {
    store *store.Store
}

func (s *PlantService) AddCareRecord(ctx context.Context, plantID string, care *CareRecord) error {
    // Get existing care history document
    careDoc, err := s.store.Db.PlantCareHistory.FindByPlantID(ctx, plantID)
    if err != nil {
        // Create new document if doesn't exist
        careDoc = &PlantCareHistory{
            PlantID:     plantID,
            TenantID:    care.TenantID,
            CareRecords: []CareRecord{},
            CreatedAt:   time.Now(),
        }
    }

    // Add new care record
    careDoc.CareRecords = append(careDoc.CareRecords, *care)
    careDoc.UpdatedAt = time.Now()

    // Store in MongoDB
    if err := s.store.Db.PlantCareHistory.Upsert(ctx, careDoc); err != nil {
        return ecode.InternalServerError.Desc(err)
    }

    // Cache invalidation
    s.store.Rdb.Del(ctx, fmt.Sprintf("plant:%s:care_history", plantID))

    return nil
}
```

5. **NotificationService** (extends existing `pkg/mail/index.go` and `pkg/ws/index.go`):
```go
type NotificationService struct {
    mail *mail.Mail
    ws   *ws.Ws
    store *store.Store
}

func (s *NotificationService) NotifyHarvestReady(ctx context.Context, plant *Plant) error {
    // Get member details
    member, err := s.store.GetMember(ctx, plant.MemberID)
    if err != nil {
        return err
    }

    // Send email notification using existing mail service
    emailData := map[string]interface{}{
        "MemberName": member.FirstName,
        "PlantCode":  plant.PlantCode,
        "PlantType":  plant.PlantTypeName,
    }
    
    if err := s.mail.Send(ctx, member.Email, "harvest_ready", emailData); err != nil {
        return err
    }

    // Send WebSocket notification using existing WS service
    wsMessage := map[string]interface{}{
        "type":    "harvest_ready",
        "plant_id": plant.ID,
        "message": "Your plant is ready for harvest!",
    }
    
    wsData, _ := json.Marshal(wsMessage)
    s.ws.EmitTo([]string{member.ID}, wsData)

    return nil
}
```

6. **PaymentService** (new, integrating with Stripe):
```go
type PaymentService struct {
    store       *store.Store
    stripeKey   string
}

func (s *PaymentService) CreatePaymentIntent(ctx context.Context, req *PaymentIntentRequest) (*PaymentIntent, error) {
    // Validate membership request
    membership, err := s.store.Db.Membership.FindByID(ctx, req.MembershipID)
    if err != nil {
        return nil, ecode.NotFound.Desc("membership not found")
    }

    // Create Stripe payment intent
    params := &stripe.PaymentIntentParams{
        Amount:   stripe.Int64(int64(membership.PaymentAmount * 100)), // Convert to cents
        Currency: stripe.String("eur"),
        Metadata: map[string]string{
            "membership_id": membership.ID,
            "tenant_id":     membership.TenantID,
        },
    }

    pi, err := paymentintent.New(params)
    if err != nil {
        return nil, ecode.PaymentProcessingError.Desc(err)
    }

    // Store payment record
    payment := &Payment{
        MembershipID:           membership.ID,
        StripePaymentIntentID:  pi.ID,
        Amount:                membership.PaymentAmount,
        Currency:              "EUR",
        Status:                "pending",
        TenantID:              membership.TenantID,
        CreatedAt:             time.Now(),
    }

    if err := s.store.Db.Payment.Create(ctx, payment); err != nil {
        return nil, ecode.InternalServerError.Desc(err)
    }

    return &PaymentIntent{
        ID:           pi.ID,
        ClientSecret: pi.ClientSecret,
        Amount:       membership.PaymentAmount,
        Status:       "pending",
    }, nil
}
```

### Data Layer (Extended from Existing Base)

**Core Infrastructure** (leveraging existing `store/` structure):

```go
// Enhanced Store Structure (store/index.go)
type Store struct {
    Db      *db.Db      // MongoDB connection and models
    Rdb     *rdb.Rdb    // Redis operations
    Storage *storage.Storage // MinIO file storage
}

// Enhanced Database Structure (store/db/index.go)
type Db struct {
    // Existing collections
    User     *mongo.Collection
    Role     *mongo.Collection
    Tenant   *mongo.Collection
    Client   *mongo.Collection
    AuditLog *mongo.Collection
    
    // Cannabis Club collections
    Member              *mongo.Collection
    Membership          *mongo.Collection
    PlantSlot           *mongo.Collection
    Plant               *mongo.Collection
    SeasonalCatalog     *mongo.Collection
    PlantType           *mongo.Collection
    Payment             *mongo.Collection
    
    // Document collections for complex data
    PlantCareHistory    *mongo.Collection
    HarvestDetails      *mongo.Collection
    NFTRecord           *mongo.Collection
}
```

**MongoDB Collections** (following existing base patterns):

#### 1. Member Collection (extends existing User pattern):
```javascript
{
    _id: ObjectId,
    // Base user fields (following existing pattern)
    username: String,       // Unique per tenant, lowercase
    email: String,         // Unique per tenant, lowercase  
    password: String,      // Hashed using existing bcrypt implementation
    phone: String,         // lowercase
    data_status: String,   // "enable", "disable" (existing pattern)
    role_ids: [String],    // Array of role ObjectIds (existing pattern)
    tenant_id: String,     // Cannabis club reference (existing pattern)
    version_token: Number, // For token versioning (existing pattern)
    
    // Cannabis-specific fields
    first_name: String,
    last_name: String,
    date_of_birth: DateTime,
    address: {
        street: String,
        city: String,
        postal_code: String,
        state: String,
        country: String
    },
    
    // eKYC fields
    kyc_status: String,    // "pending", "verified", "rejected"
    kyc_documents: {
        id_document_url: String,
        selfie_url: String,
        uploaded_at: DateTime
    },
    kyc_verified_at: DateTime,
    kyc_verified_by: String,
    kyc_notes: String,
    
    // Membership tracking
    current_membership_id: String,
    membership_history: [String],
    
    // Following existing audit pattern
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

// Indexes (following existing pattern)
Indexes:
- {tenant_id: 1, email: 1} unique
- {tenant_id: 1, username: 1} unique
- {tenant_id: 1, phone: 1} unique
- {kyc_status: 1}
- {current_membership_id: 1}
```

#### 2. Membership Collection (new, following base patterns):
```javascript
{
    _id: ObjectId,
    member_id: String,          // Reference to member
    tenant_id: String,          // Cannabis club reference (base pattern)
    
    // Membership details
    membership_type: String,    // "annual", "seasonal"
    status: String,            // "pending_payment", "active", "expired", "cancelled", "transferred"
    activation_date: DateTime,
    expiration_date: DateTime,
    grace_end_date: DateTime,
    
    // Payment tracking
    payment_amount: Number,
    payment_currency: String,
    payment_status: String,    // "pending", "completed", "failed", "refunded"
    stripe_payment_intent_id: String,
    
    // Plant slots
    allocated_slots: Number,
    used_slots: Number,
    
    // Transfer tracking
    transferred_from: String,   // Previous member ID
    transferred_to: String,     // New member ID
    transfer_date: DateTime,
    transfer_reason: String,
    
    // Renewal tracking
    renewal_count: Number,
    auto_renewal: Boolean,
    
    // Catalog reference
    catalog_id: String,        // Reference to seasonal catalog
    
    // Following existing base pattern
    data_status: String,       // "enable", "disable"
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
- {member_id: 1}
- {status: 1}
- {expiration_date: 1}
- {stripe_payment_intent_id: 1} unique
```

#### 3. PlantSlot Collection (new):
```javascript
{
    _id: ObjectId,
    membership_id: String,      // Reference to membership
    tenant_id: String,          // Cannabis club reference (base pattern)
    
    // Slot identification
    slot_number: String,        // Unique within farm/area
    farm_location: String,
    area_designation: String,
    
    // Slot status
    status: String,            // "available", "allocated", "occupied", "harvesting", "maintenance"
    
    // Plant assignment
    current_plant_id: String,   // Reference to current plant
    plant_history: [String],    // Array of plant IDs
    
    // NFT integration
    nft_token_id: String,       // Blockchain token ID
    nft_contract_address: String,
    nft_metadata_uri: String,
    nft_mint_transaction: String,
    nft_mint_date: DateTime,
    
    // Catalog assignment
    catalog_id: String,         // Reference to seasonal catalog
    plant_type_id: String,      // Selected plant variety
    
    // Following existing base pattern
    data_status: String,        // "enable", "disable"
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
- {membership_id: 1}
- {status: 1}
- {nft_token_id: 1} unique
- {slot_number: 1, farm_location: 1, area_designation: 1} unique
- {current_plant_id: 1}
```

#### 4. Plant Collection (new):
```javascript
{
    _id: ObjectId,
    plant_slot_id: String,      // Reference to plant slot
    member_id: String,          // Reference to member (for quick access)
    tenant_id: String,          // Cannabis club reference (base pattern)
    
    // Plant identification
    plant_type_id: String,      // Reference to plant type
    plant_code: String,         // Unique identifier (auto-generated)
    
    // Lifecycle tracking
    state: String,             // "seedling", "vegetative", "flowering", "harvesting", "harvested", "destroyed"
    cycle_number: Number,      // Growth cycle count
    start_date: DateTime,      // Planting date
    
    // Growth tracking
    estimated_harvest_date: DateTime,
    actual_harvest_date: DateTime,
    days_to_harvest: Number,   // Calculated field
    
    // Yield tracking
    estimated_yield: Number,    // Grams
    actual_yield: Number,       // Grams
    
    // Care history reference (MongoDB document pattern)
    care_history_doc_id: String, // Reference to PlantCareHistory document
    
    // Quality metrics
    quality_score: Number,      // 1-10 scale
    thc_content: Number,        // Percentage
    cbd_content: Number,        // Percentage
    
    // Health monitoring
    health_status: String,      // "healthy", "sick", "pest_issues", "nutrient_deficiency"
    last_care_date: DateTime,
    next_care_due: DateTime,
    
    // Following existing base pattern
    data_status: String,        // "enable", "disable"
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
- {plant_slot_id: 1}
- {member_id: 1}
- {state: 1}
- {plant_code: 1} unique
- {actual_harvest_date: 1}
- {next_care_due: 1}
```

#### 5. SeasonalCatalog Collection (new):
```javascript
{
    _id: ObjectId,
    tenant_id: String,          // Cannabis club reference (base pattern)
    
    // Catalog details
    season_name: String,        // "Spring 2025", "Summer 2025"
    description: String,
    image_url: String,
    
    // Availability
    start_date: DateTime,       // When catalog becomes available
    end_date: DateTime,         // When catalog closes
    registration_deadline: DateTime, // Last day to register
    status: String,            // "upcoming", "active", "registration_closed", "closed"
    
    // Plant types available in this catalog
    available_plant_types: [String], // Array of plant type IDs
    
    // Pricing
    membership_price: Number,
    currency: String,
    processing_fee: Number,     // Per gram processing fee
    
    // Capacity management
    max_members: Number,
    current_members: Number,
    max_slots: Number,
    allocated_slots: Number,
    
    // Following existing base pattern
    data_status: String,        // "enable", "disable"
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
- {status: 1}
- {start_date: 1, end_date: 1}
- {registration_deadline: 1}
```

#### 6. PlantType Collection (new):
```javascript
{
    _id: ObjectId,
    tenant_id: String,          // Cannabis club reference (base pattern)
    
    // Plant variety details
    name: String,
    species: String,           // "indica", "sativa", "hybrid"
    genetics: String,          // Genetic lineage
    description: String,
    image_url: String,
    
    // Growing characteristics
    growth_period_days: Number,
    seed_to_harvest_days: Number,
    difficulty_level: String,   // "beginner", "intermediate", "advanced"
    flowering_time_weeks: Number,
    
    // Expected yields
    expected_yield_min: Number, // Grams
    expected_yield_max: Number, // Grams
    yield_indoor: Number,       // Average indoor yield
    yield_outdoor: Number,      // Average outdoor yield
    
    // Cannabinoid profiles
    thc_content_min: Number,    // Percentage
    thc_content_max: Number,    // Percentage
    cbd_content_min: Number,    // Percentage
    cbd_content_max: Number,    // Percentage
    
    // Terpene profile
    dominant_terpenes: [String],
    aroma_profile: String,
    flavor_profile: String,
    
    // Growing requirements
    light_requirements: String, // "low", "medium", "high"
    water_requirements: String, // "low", "medium", "high"
    nutrient_requirements: String, // "low", "medium", "high"
    temperature_range: {
        min: Number,            // Celsius
        max: Number             // Celsius
    },
    humidity_range: {
        min: Number,            // Percentage
        max: Number             // Percentage
    },
    
    // Availability
    available_seasons: [String], // Array of catalog IDs where this type is available
    
    // Following existing base pattern
    data_status: String,        // "enable", "disable"
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
- {species: 1}
- {name: 1}
- {difficulty_level: 1}
```

#### 7. Payment Collection (new, following base patterns):
```javascript
{
    _id: ObjectId,
    membership_id: String,      // Reference to membership
    member_id: String,          // Reference to member
    tenant_id: String,          // Cannabis club reference (base pattern)
    
    // Payment details
    amount: Number,
    currency: String,
    payment_type: String,       // "membership", "processing_fee", "renewal"
    
    // Stripe integration
    stripe_payment_intent_id: String,
    stripe_charge_id: String,
    stripe_customer_id: String,
    
    // Payment status
    status: String,            // "pending", "processing", "completed", "failed", "cancelled", "refunded"
    payment_method: String,    // "card", "bank_transfer", "sepa"
    
    // Timestamps
    initiated_at: DateTime,
    completed_at: DateTime,
    failed_at: DateTime,
    
    // Error handling
    failure_reason: String,
    failure_code: String,
    
    // Refund tracking
    refund_amount: Number,
    refund_reason: String,
    refunded_at: DateTime,
    
    // Following existing base pattern
    data_status: String,        // "enable", "disable"
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
- {member_id: 1}
- {membership_id: 1}
- {stripe_payment_intent_id: 1} unique
- {status: 1}
- {completed_at: 1}
```

**MongoDB Document Collections** (for complex data, following base document patterns):

#### Plant Care History Documents (following base MongoDB document pattern):
```javascript
{
    "_id": ObjectId,
    "plant_id": "12345",        // Reference to Plant collection
    "tenant_id": "club_001",    // Following base tenant pattern
    "care_records": [
        {
            "timestamp": "2025-06-15T14:30:00Z",
            "type": "watering",     // "watering", "fertilizer", "pruning", "training", "observation"
            "amount": "250",        // ml as string for watering
            "ph_level": "6.2",
            "ec_level": "1.4",      // Electrical conductivity
            "nutrients": {
                "nitrogen": "20",
                "phosphorus": "10",
                "potassium": "20",
                "calcium": "15",
                "magnesium": "5"
            },
            "environmental": {
                "temperature": "24",  // Celsius
                "humidity": "65",     // Percentage
                "light_hours": "18"
            },
            "notes": "Plant showing excellent growth, new leaves developing",
            "recorded_by": "member_123",
            "staff_verified": true,
            "images": [
                {
                    "url": "https://storage.seedeg.com/plants/12345/care_001.jpg",
                    "timestamp": "2025-06-15T14:30:00Z",
                    "description": "Overall plant health"
                }
            ]
        }
    ],
    "growth_metrics": {
        "height": [
            {"value": "5", "unit": "cm", "date": "2025-06-15", "measured_by": "member_123"},
            {"value": "8", "unit": "cm", "date": "2025-06-22", "measured_by": "member_123"}
        ],
        "leaf_count": [
            {"value": "4", "date": "2025-06-15"},
            {"value": "6", "date": "2025-06-22"}
        ],
        "stem_diameter": [
            {"value": "3", "unit": "mm", "date": "2025-06-15"}
        ],
        "node_count": [
            {"value": "3", "date": "2025-06-15"}
        ]
    },
    "health_assessments": [
        {
            "date": "2025-06-15",
            "overall_health": "excellent",
            "pest_issues": "none",
            "disease_signs": "none",
            "nutrient_status": "healthy",
            "leaf_color": "vibrant_green",
            "assessed_by": "staff_456"
        }
    ],
    "created_at": "2025-06-01T00:00:00Z",
    "updated_at": "2025-06-22T14:30:00Z"
}
```

#### Harvest Details Documents:
```javascript
{
    "_id": ObjectId,
    "plant_id": "12345",
    "member_id": "member_123",
    "tenant_id": "club_001",
    "harvest_date": "2025-09-01T00:00:00Z",
    
    "pre_harvest": {
        "final_height": "120",      // cm
        "final_node_count": "12",
        "trichome_status": "80_cloudy_20_amber",
        "harvest_decision_reason": "optimal_trichome_development",
        "last_feeding_date": "2025-08-25T00:00:00Z",
        "flush_duration_days": "7"
    },
    
    "harvest_process": {
        "harvested_by": "staff_456",
        "harvest_method": "wet_trim",  // "wet_trim", "dry_trim"
        "harvest_start_time": "2025-09-01T08:00:00Z",
        "harvest_end_time": "2025-09-01T11:30:00Z",
        "weather_conditions": "sunny_low_humidity",
        "tools_used": ["pruning_shears", "trimming_scissors", "collection_bags"]
    },
    
    "yield_details": {
        "total_wet_weight": "156.8",    // grams (including stems and leaves)
        "bud_wet_weight": "89.4",       // grams (buds only)
        "trim_wet_weight": "35.2",      // grams (sugar leaves)
        "stem_weight": "32.2",          // grams
        "final_dry_weight": "28.5",     // grams (final smokable product)
        "dry_trim_weight": "8.1",       // grams (dry trim for processing)
        "waste_weight": "8.6",          // grams (unusable material)
        "moisture_content_final": "12.5", // percentage
        "drying_loss_percentage": "68.1" // percentage weight loss during drying
    },
    
    "quality_assessment": {
        "visual_inspection": {
            "bud_structure": "dense",    // "dense", "loose", "airy"
            "color": "deep_green_purple_highlights",
            "trichome_coverage": "excellent", // "poor", "good", "excellent"
            "pistil_color": "orange_brown",
            "overall_appearance": "9.0"  // 1-10 scale
        },
        "aroma_profile": {
            "intensity": "8.5",         // 1-10 scale
            "primary_notes": ["citrus", "pine", "earthy"],
            "secondary_notes": ["sweet", "spicy"],
            "terpene_dominance": "limonene"
        },
        "bud_density": "8.0",           // 1-10 scale
        "trim_quality": "7.5",          // 1-10 scale
        "overall_quality": "8.2"       // 1-10 scale
    },
    
    "lab_testing": {
        "tested": true,
        "test_date": "2025-09-05T00:00:00Z",
        "lab_name": "Cannabis Analytics Lab GmbH",
        "lab_certification": "ISO_17025",
        "sample_id": "CAL_2025_090501",
        "cannabinoids": {
            "thc_total": "22.3",        // percentage
            "thc_a": "24.8",            // percentage
            "thc_d9": "0.8",            // percentage
            "cbd_total": "0.8",         // percentage
            "cbd_a": "0.9",             // percentage
            "cbd": "0.1",               // percentage
            "cbg": "1.2",               // percentage
            "cbn": "0.3"                // percentage
        },
        "terpenes": {
            "total_terpenes": "2.8",    // percentage
            "myrcene": "1.2",           // percentage
            "limonene": "0.8",          // percentage
            "pinene": "0.6",            // percentage
            "linalool": "0.2"           // percentage
        },
        "contaminants": {
            "pesticides": "none_detected",
            "heavy_metals": {
                "lead": "below_detection_limit",
                "cadmium": "below_detection_limit",
                "mercury": "below_detection_limit",
                "arsenic": "below_detection_limit"
            },
            "microbials": {
                "total_yeast_mold": "pass",
                "e_coli": "not_detected",
                "salmonella": "not_detected"
            },
            "residual_solvents": "not_applicable"
        }
    },
    
    "processing": {
        "drying": {
            "method": "hang_dry",       // "hang_dry", "rack_dry", "freeze_dry"
            "start_date": "2025-09-01T12:00:00Z",
            "completion_date": "2025-09-08T00:00:00Z",
            "duration_days": "7",
            "temperature_range": "18-21", // Celsius
            "humidity_range": "45-55",   // Percentage
            "air_circulation": "gentle_fan",
            "light_exposure": "complete_darkness"
        },
        "curing": {
            "start_date": "2025-09-08T00:00:00Z",
            "method": "glass_jar_cure",
            "duration_planned_days": "21",
            "current_duration_days": "21",
            "container_type": "amber_glass_jars",
            "burping_schedule": {
                "week_1": "twice_daily_15_minutes",
                "week_2": "daily_10_minutes",
                "week_3": "every_other_day_5_minutes"
            },
            "humidity_packs_used": true,
            "target_humidity": "62"     // Percentage
        }
    },
    
    "packaging": {
        "package_date": "2025-09-29T00:00:00Z",
        "packaging_method": "vacuum_sealed_glass",
        "container_details": {
            "type": "amber_glass_jar",
            "size": "50ml",
            "seal_type": "airtight_lid",
            "desiccant_pack": true
        },
        "net_weight": "25.2",           // grams (final product weight)
        "package_weight": "1.8",        // grams (packaging weight)
        "total_weight": "27.0",         // grams
        "label_info": {
            "strain_name": "Blue Dream Phenotype #3",
            "harvest_date": "2025-09-01",
            "package_date": "2025-09-29",
            "thc_content": "22.3%",
            "cbd_content": "0.8%",
            "terpene_content": "2.8%",
            "best_by_date": "2026-09-01",
            "storage_instructions": "store_cool_dark_dry",
            "lot_number": "BD2025090101",
            "qr_code": "https://verify.seedeg.com/lot/BD2025090101"
        }
    },
    
    "compliance": {
        "german_cannabis_law_compliant": true,
        "tracking_id": "DE_SEED_2025_090101",
        "cultivation_license": "CULT_2025_001",
        "processing_license": "PROC_2025_001",
        "batch_documentation": "complete",
        "chain_of_custody": [
            {
                "stage": "cultivation",
                "responsible": "staff_456",
                "date": "2025-06-01T00:00:00Z"
            },
            {
                "stage": "harvest",
                "responsible": "staff_456",
                "date": "2025-09-01T00:00:00Z"
            },
            {
                "stage": "processing",
                "responsible": "staff_789",
                "date": "2025-09-08T00:00:00Z"
            },
            {
                "stage": "packaging",
                "responsible": "staff_789",
                "date": "2025-09-29T00:00:00Z"
            }
        ]
    },
    
    "distribution": {
        "member_allocation": "25.2",    // grams
        "processing_fee_per_gram": "0.10", // euros
        "total_processing_fee": "2.52", // euros
        "pickup_scheduled": "2025-10-01T14:00:00Z",
        "pickup_location": "Seed eG Facility A",
        "pickup_contact": "staff_789",
        "pickup_status": "completed",
        "pickup_actual": "2025-10-01T14:15:00Z",
        "member_signature": "digital_signature_hash",
        "id_verification": "completed"
    },
    
    "images": [
        {
            "url": "https://storage.seedeg.com/harvests/67890/pre_harvest_full.jpg",
            "type": "pre_harvest_full_plant",
            "timestamp": "2025-09-01T07:45:00Z",
            "description": "Full plant before harvest"
        },
        {
            "url": "https://storage.seedeg.com/harvests/67890/wet_buds.jpg",
            "type": "wet_harvest",
            "timestamp": "2025-09-01T11:30:00Z",
            "description": "Freshly harvested wet buds"
        },
        {
            "url": "https://storage.seedeg.com/harvests/67890/drying.jpg",
            "type": "drying_process",
            "timestamp": "2025-09-04T00:00:00Z",
            "description": "Buds hanging to dry"
        },
        {
            "url": "https://storage.seedeg.com/harvests/67890/final_product.jpg",
            "type": "final_product",
            "timestamp": "2025-09-29T00:00:00Z",
            "description": "Final cured and packaged product"
        }
    ],
    
    "created_at": "2025-09-01T00:00:00Z",
    "updated_at": "2025-10-01T14:15:00Z"
}
```

#### NFT Record Collection (new, following base patterns):
```javascript
{
    _id: ObjectId,
    plant_slot_id: String,      // Reference to PlantSlot
    member_id: String,          // Reference to Member
    tenant_id: String,          // Cannabis club reference (base pattern)
    
    // NFT Details
    token_id: String,           // Unique blockchain token ID
    contract_address: String,   // Smart contract address
    chain_id: Number,           // Blockchain network ID
    
    // Metadata
    metadata_uri: String,       // IPFS or centralized metadata URI
    metadata: {
        name: String,           // "Seed eG Plant Slot #123"
        description: String,    // Detailed description
        image: String,          // NFT image URI
        external_url: String,   // Link to plant slot details
        attributes: [
            {
                trait_type: String, // "Farm Location", "Slot Number", "Season"
                value: String
            }
        ]
    },
    
    // Blockchain Tracking
    mint_transaction_hash: String,
    mint_block_number: Number,
    mint_gas_used: Number,
    mint_gas_price: String,
    
    // Transfer History
    transfer_history: [
        {
            from_address: String,
            to_address: String,
            transaction_hash: String,
            block_number: Number,
            timestamp: DateTime
        }
    ],
    
    // Status
    status: String,             // "minting", "minted", "transferred", "burned"
    owner_address: String,      // Current owner wallet address
    
    // Following existing base pattern
    data_status: String,        // "enable", "disable"
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
- {plant_slot_id: 1}
- {member_id: 1}
- {token_id: 1} unique
- {contract_address: 1, token_id: 1} unique
- {owner_address: 1}
```

**Redis Cache Implementation** (extending existing `store/rdb/index.go` patterns):

```go
// Cache Keys (following existing pattern from base)
const (
    // Member-related cache keys
    MemberCacheKey           = "member:%s"
    MemberProfileKey         = "member:%s:profile"
    MemberMembershipsKey     = "member:%s:memberships"
    MemberPlantsKey          = "member:%s:plants"
    MemberSlotsKey           = "member:%s:slots"
    
    // Membership cache keys
    MembershipCacheKey       = "membership:%s"
    MembershipDetailsKey     = "membership:%s:details"
    ActiveMembershipsKey     = "memberships:active:%s" // tenant_id
    ExpiringMembershipsKey   = "memberships:expiring:%s" // tenant_id
    
    // Plant slot cache keys
    PlantSlotCacheKey        = "plantslot:%s"
    AvailableSlotsKey        = "slots:available:%s" // tenant_id
    SlotsByFarmKey           = "slots:farm:%s:%s" // tenant_id, farm_location
    
    // Plant cache keys
    PlantCacheKey            = "plant:%s"
    PlantCareHistoryKey      = "plant:%s:care_history"
    PlantsReadyHarvestKey    = "plants:ready_harvest:%s" // tenant_id
    
    // Catalog cache keys
    CatalogCacheKey          = "catalog:%s"
    ActiveCatalogsKey        = "catalogs:active:%s" // tenant_id
    CatalogPlantTypesKey     = "catalog:%s:plant_types"
    
    // Payment cache keys
    PaymentIntentKey         = "payment_intent:%s"
    PaymentStatusKey         = "payment:%s:status"
    
    // Session and rate limiting (existing pattern)
    SessionCacheKey          = "session:%s"
    RateLimitKey            = "ratelimit:%s:%s" // ip, endpoint
    
    // Real-time notifications
    NotificationQueueKey     = "notifications:%s" // member_id
    WebSocketSessionKey      = "ws_session:%s" // session_id
)

// Cache TTL Configuration (following existing pattern)
const (
    MemberCacheTTL           = time.Hour * 24
    MemberProfileTTL         = time.Hour * 12
    MembershipCacheTTL       = time.Hour * 12
    PlantSlotCacheTTL        = time.Hour * 6
    PlantCacheTTL            = time.Hour * 2
    PlantCareHistoryTTL      = time.Hour * 1
    CatalogCacheTTL          = time.Hour * 24
    PaymentIntentTTL         = time.Hour * 2
    SessionCacheTTL          = time.Hour * 1 // existing
    NotificationTTL          = time.Hour * 48
    AvailableSlotsTTL        = time.Minute * 30
)

// Enhanced Cache Operations (extending existing rdb patterns)
func (s *Store) GetMemberWithDetails(ctx context.Context, memberID string) (*MemberDetails, error) {
    key := fmt.Sprintf(MemberProfileKey, memberID)
    cache := &MemberDetails{}

    // Try cache first (existing pattern)
    if err := s.Rdb.GetWithRefresh(ctx, key, cache, func() (interface{}, error) {
        // Load from database if not in cache
        member, err := s.Db.Member.FindByID(ctx, memberID)
        if err != nil {
            return nil, err
        }

        // Get current membership
        var membership *Membership
        if member.CurrentMembershipID != "" {
            membership, _ = s.Db.Membership.FindByID(ctx, member.CurrentMembershipID)
        }

        // Get plant slots
        slots, _ := s.Db.PlantSlot.FindByMemberID(ctx, memberID)

        return &MemberDetails{
            Member:     member,
            Membership: membership,
            Slots:      slots,
        }, nil
    }, MemberProfileTTL); err != nil {
        return nil, err
    }

    return cache, nil
}

// Real-time Cache Updates (for WebSocket notifications)
func (s *Store) CacheNotification(ctx context.Context, memberID string, notification *Notification) error {
    key := fmt.Sprintf(NotificationQueueKey, memberID)
    
    // Get existing notifications
    var notifications []*Notification
    s.Rdb.Get(ctx, key, &notifications)
    
    // Add new notification
    notifications = append(notifications, notification)
    
    // Keep only last 50 notifications
    if len(notifications) > 50 {
        notifications = notifications[len(notifications)-50:]
    }
    
    return s.Rdb.Set(ctx, key, notifications, NotificationTTL)
}
```

**File Storage** (extending existing MinIO integration from `store/storage/index.go`):

```go
// Enhanced Storage Service (extending existing storage patterns)
type StorageService struct {
    client     *minio.Client
    bucketName string
}

// Storage Buckets (extending existing pattern)
const (
    // Existing buckets
    ImagesBucket    = "images"
    VideosBucket    = "videos"
    DocumentsBucket = "documents"
    
    // Cannabis club specific buckets
    KYCDocumentsBucket    = "kyc-documents"
    PlantImagesBucket     = "plant-images"
    HarvestImagesBucket   = "harvest-images"
    NFTMetadataBucket     = "nft-metadata"
    ComplianceBucket      = "compliance-docs"
    LabReportsBucket      = "lab-reports"
)

// Enhanced Upload Methods (following existing patterns)
func (s *StorageService) UploadKYCDocument(ctx context.Context, memberID string, fileType string, file io.Reader, size int64) (*UploadResult, error) {
    // Generate secure filename
    fileName := fmt.Sprintf("%s/%s_%d_%s", memberID, fileType, time.Now().Unix(), uuid.New().String())
    
    // Upload with encryption
    info, err := s.client.PutObject(ctx, KYCDocumentsBucket, fileName, file, size, minio.PutObjectOptions{
        ContentType:        "application/octet-stream",
        ServerSideEncryption: encrypt.NewSSE(),
        UserMetadata: map[string]string{
            "member-id": memberID,
            "file-type": fileType,
            "uploaded": time.Now().Format(time.RFC3339),
        },
    })
    
    if err != nil {
        return nil, err
    }
    
    return &UploadResult{
        URL:      fmt.Sprintf("https://%s/%s/%s", s.endpoint, KYCDocumentsBucket, fileName),
        FileName: fileName,
        Size:     info.Size,
        ETag:     info.ETag,
    }, nil
}
```

### API Endpoints (Extended from Base)

**Following existing route patterns from `route/` directory:**

```go
// Enhanced Route Registration (route/index.go pattern)
func RegisterCannabisRoutes(r *gin.Engine, store *store.Store) {
    // Create middleware instance (existing pattern)
    mdw := middleware{store: store}
    
    // API v1 group (existing pattern)
    v1 := r.Group("/api/v1")
    
    // Authentication routes (enhanced from existing auth.go)
    authGroup := v1.Group("/auth")
    {
        authGroup.POST("/register", mdw.NoAuth(), registerMember)
        authGroup.POST("/login", mdw.NoAuth(), loginMember)
        authGroup.POST("/refresh", mdw.NoAuth(), refreshToken)
        authGroup.POST("/logout", mdw.BearerAuth(), logoutMember)
        authGroup.POST("/verify-email", mdw.NoAuth(), verifyEmail)
        authGroup.POST("/reset-password", mdw.NoAuth(), resetPassword)
    }
    
    // Member management (new, following existing user pattern)
    memberGroup := v1.Group("/members")
    {
        memberGroup.GET("/me", mdw.BearerAuth(enum.PermissionMemberView), getMemberProfile)
        memberGroup.PUT("/me", mdw.BearerAuth(enum.PermissionMemberUpdate), updateMemberProfile)
        memberGroup.POST("/kyc/upload", mdw.BearerAuth(enum.PermissionMemberUpdate), uploadKYCDocuments)
        memberGroup.GET("/kyc/status", mdw.BearerAuth(enum.PermissionMemberView), getKYCStatus)
        memberGroup.POST("/kyc/submit", mdw.BearerAuth(enum.PermissionMemberUpdate), submitKYCVerification)
    }
    
    // Catalog management (new)
    catalogGroup := v1.Group("/catalogs")
    {
        catalogGroup.GET("", mdw.BearerAuth(enum.PermissionMemberView), listCatalogs)
        catalogGroup.GET("/:id", mdw.BearerAuth(enum.PermissionMemberView), getCatalogDetails)
        catalogGroup.GET("/:id/plant-types", mdw.BearerAuth(enum.PermissionMemberView), getCatalogPlantTypes)
    }
    
    // Membership management (new)
    membershipGroup := v1.Group("/memberships")
    {
        membershipGroup.POST("", mdw.BearerAuth(enum.PermissionMembershipCreate), createMembership)
        membershipGroup.GET("", mdw.BearerAuth(enum.PermissionMembershipView), getMemberMemberships)
        membershipGroup.GET("/:id", mdw.BearerAuth(enum.PermissionMembershipView), getMembershipDetails)
        membershipGroup.POST("/:id/renew", mdw.BearerAuth(enum.PermissionMembershipUpdate), renewMembership)
        membershipGroup.POST("/:id/cancel", mdw.BearerAuth(enum.PermissionMembershipUpdate), cancelMembership)
        membershipGroup.POST("/:id/transfer", mdw.BearerAuth(enum.PermissionMembershipUpdate), transferMembership)
    }
    
    // Plant slot management (new)
    slotGroup := v1.Group("/plantslots")
    {
        slotGroup.GET("", mdw.BearerAuth(enum.PermissionPlantSlotView), getMemberPlantSlots)
        slotGroup.GET("/:id", mdw.BearerAuth(enum.PermissionPlantSlotView), getPlantSlotDetails)
        slotGroup.PUT("/:id", mdw.BearerAuth(enum.PermissionPlantSlotUpdate), updatePlantSlot)
        slotGroup.GET("/:id/nft", mdw.BearerAuth(enum.PermissionPlantSlotView), getPlantSlotNFT)
        slotGroup.POST("/:id/assign-plant", mdw.BearerAuth(enum.PermissionPlantSlotUpdate), assignPlantToSlot)
    }
    
    // Plant management (new)
    plantGroup := v1.Group("/plants")
    {
        plantGroup.GET("/:id", mdw.BearerAuth(enum.PermissionPlantView), getPlantDetails)
        plantGroup.PUT("/:id", mdw.BearerAuth(enum.PermissionPlantUpdate), updatePlant)
        plantGroup.POST("/:id/care", mdw.BearerAuth(enum.PermissionPlantUpdate), addCareRecord)
        plantGroup.GET("/:id/care-history", mdw.BearerAuth(enum.PermissionPlantView), getPlantCareHistory)
        plantGroup.POST("/:id/images", mdw.BearerAuth(enum.PermissionPlantUpdate), uploadPlantImages)
        plantGroup.PUT("/:id/state", mdw.BearerAuth(enum.PermissionPlantUpdate), updatePlantState)
    }
    
    // Harvest management (new)
    harvestGroup := v1.Group("/harvests")
    {
        harvestGroup.GET("", mdw.BearerAuth(enum.PermissionHarvestView), getMemberHarvests)
        harvestGroup.GET("/:id", mdw.BearerAuth(enum.PermissionHarvestView), getHarvestDetails)
        harvestGroup.POST("/plants/:plant_id", mdw.BearerAuth(enum.PermissionHarvestUpdate), recordHarvest)
        harvestGroup.PUT("/:id", mdw.BearerAuth(enum.PermissionHarvestUpdate), updateHarvestDetails)
        harvestGroup.POST("/:id/process", mdw.BearerAuth(enum.PermissionHarvestUpdate), processHarvest)
        harvestGroup.POST("/:id/pickup", mdw.BearerAuth(enum.PermissionHarvestView), schedulePickup)
    }
    
    // Payment routes (new, following existing patterns)
    paymentGroup := v1.Group("/payments")
    {
        paymentGroup.POST("/create-intent", mdw.BearerAuth(enum.PermissionMembershipCreate), createPaymentIntent)
        paymentGroup.POST("/confirm", mdw.BearerAuth(enum.PermissionMembershipCreate), confirmPayment)
        paymentGroup.GET("/history", mdw.BearerAuth(enum.PermissionMemberView), getPaymentHistory)
        paymentGroup.POST("/webhooks/stripe", mdw.NoAuth(), stripeWebhookHandler)
    }
    
    // Storage routes (enhanced from existing storage.go)
    storageGroup := v1.Group("/storage")
    {
        // Existing routes
        storageGroup.GET("/images/:filename", mdw.NoAuth(), downloadImage)
        storageGroup.GET("/videos/:filename", mdw.NoAuth(), downloadVideo)
        storageGroup.POST("/images", mdw.BearerAuth(), uploadImage)
        storageGroup.POST("/videos", mdw.BearerAuth(), uploadVideo)
        
        // Cannabis-specific routes
        storageGroup.POST("/kyc/documents", mdw.BearerAuth(enum.PermissionMemberUpdate), uploadKYCDocument)
        storageGroup.POST("/plants/:plant_id/images", mdw.BearerAuth(enum.PermissionPlantUpdate), uploadPlantImage)
        storageGroup.POST("/harvests/:harvest_id/images", mdw.BearerAuth(enum.PermissionHarvestUpdate), uploadHarvestImage)
        storageGroup.GET("/kyc/:filename", mdw.BearerAuth(enum.PermissionKYCVerify), downloadKYCDocument)
    }
    
    // CMS routes (enhanced from existing CMS pattern)
    cmsGroup := v1.Group("/cms")
    {
        // Member management (staff/admin)
        membersCMS := cmsGroup.Group("/members")
        {
            membersCMS.GET("", mdw.BearerAuth(enum.PermissionUserView), listAllMembers)
            membersCMS.GET("/:id", mdw.BearerAuth(enum.PermissionUserView), getMemberDetailsCMS)
            membersCMS.PUT("/:id", mdw.BearerAuth(enum.PermissionUserUpdate), updateMemberCMS)
            membersCMS.DELETE("/:id", mdw.BearerAuth(enum.PermissionUserDelete), deleteMember)
        }
        
        // KYC verification (staff/admin)
        kycCMS := cmsGroup.Group("/kyc")
        {
            kycCMS.GET("/pending", mdw.BearerAuth(enum.PermissionKYCVerify), getPendingKYCVerifications)
            kycCMS.POST("/:id/approve", mdw.BearerAuth(enum.PermissionKYCVerify), approveKYCVerification)
            kycCMS.POST("/:id/reject", mdw.BearerAuth(enum.PermissionKYCVerify), rejectKYCVerification)
            kycCMS.GET("/:id/documents", mdw.BearerAuth(enum.PermissionKYCVerify), getKYCDocuments)
        }
        
        // Membership management (admin)
        membershipsCMS := cmsGroup.Group("/memberships")
        {
            membershipsCMS.GET("", mdw.BearerAuth(enum.PermissionMembershipView), listAllMemberships)
            membershipsCMS.GET("/:id", mdw.BearerAuth(enum.PermissionMembershipView), getMembershipDetailsCMS)
            membershipsCMS.PUT("/:id", mdw.BearerAuth(enum.PermissionMembershipUpdate), updateMembershipCMS)
            membershipsCMS.POST("/:id/extend", mdw.BearerAuth(enum.PermissionMembershipUpdate), extendMembership)
        }
        
        // Plant slot management (staff/admin)
        slotsCMS := cmsGroup.Group("/plantslots")
        {
            slotsCMS.GET("", mdw.BearerAuth(enum.PermissionPlantSlotView), listAllPlantSlots)
            slotsCMS.POST("", mdw.BearerAuth(enum.PermissionPlantSlotCreate), createPlantSlot)
            slotsCMS.PUT("/:id", mdw.BearerAuth(enum.PermissionPlantSlotUpdate), updatePlantSlotCMS)
            slotsCMS.GET("/:id/history", mdw.BearerAuth(enum.PermissionPlantSlotView), getPlantSlotHistory)
        }
        
        // Catalog management (admin)
        catalogsCMS := cmsGroup.Group("/catalogs")
        {
            catalogsCMS.GET("", mdw.BearerAuth(enum.PermissionCatalogManage), listAllCatalogs)
            catalogsCMS.POST("", mdw.BearerAuth(enum.PermissionCatalogManage), createCatalog)
            catalogsCMS.PUT("/:id", mdw.BearerAuth(enum.PermissionCatalogManage), updateCatalog)
            catalogsCMS.DELETE("/:id", mdw.BearerAuth(enum.PermissionCatalogManage), deleteCatalog)
        }
        
        // Plant type management (admin)
        plantTypesCMS := cmsGroup.Group("/plant-types")
        {
            plantTypesCMS.GET("", mdw.BearerAuth(enum.PermissionCatalogManage), listPlantTypes)
            plantTypesCMS.POST("", mdw.BearerAuth(enum.PermissionCatalogManage), createPlantType)
            plantTypesCMS.PUT("/:id", mdw.BearerAuth(enum.PermissionCatalogManage), updatePlantType)
            plantTypesCMS.DELETE("/:id", mdw.BearerAuth(enum.PermissionCatalogManage), deletePlantType)
        }
        
        // Analytics and reporting (admin)
        statsCMS := cmsGroup.Group("/stats")
        {
            statsCMS.GET("/dashboard", mdw.BearerAuth(enum.PermissionStatsView), getDashboardStats)
            statsCMS.GET("/members", mdw.BearerAuth(enum.PermissionStatsView), getMemberStats)
            statsCMS.GET("/plants", mdw.BearerAuth(enum.PermissionStatsView), getPlantStats)
            statsCMS.GET("/harvests", mdw.BearerAuth(enum.PermissionStatsView), getHarvestStats)
            statsCMS.GET("/revenue", mdw.BearerAuth(enum.PermissionStatsView), getRevenueStats)
        }
    }
    
    // WebSocket routes (enhanced from existing ws pattern)
    wsGroup := v1.Group("/ws")
    {
        wsGroup.GET("", mdw.BearerAuth(), handleWebSocketConnection)
        wsGroup.GET("/plant/:plant_id", mdw.BearerAuth(enum.PermissionPlantView), handlePlantWebSocket)
        wsGroup.GET("/member", mdw.BearerAuth(enum.PermissionMemberView), handleMemberWebSocket)
    }
}
```

### Enhanced Security Implementation

**Permission System** (extending existing `pkg/enum/index.go`):

```go
// Enhanced Permission System (pkg/enum/index.go)
type Permission string

const (
    // Existing base permissions
    PermissionUserView    Permission = "user.view"
    PermissionUserCreate  Permission = "user.create"
    PermissionUserUpdate  Permission = "user.update"
    PermissionUserDelete  Permission = "user.delete"
    PermissionRoleView    Permission = "role.view"
    PermissionRoleCreate  Permission = "role.create"
    PermissionRoleUpdate  Permission = "role.update"
    PermissionRoleDelete  Permission = "role.delete"
    
    // Cannabis Club Member Permissions
    PermissionMemberView        Permission = "member.view"
    PermissionMemberCreate      Permission = "member.create"
    PermissionMemberUpdate      Permission = "member.update"
    PermissionMemberDelete      Permission = "member.delete"
    
    // Membership Permissions
    PermissionMembershipView    Permission = "membership.view"
    PermissionMembershipCreate  Permission = "membership.create"
    PermissionMembershipUpdate  Permission = "membership.update"
    PermissionMembershipDelete  Permission = "membership.delete"
    PermissionMembershipTransfer Permission = "membership.transfer"
    
    // Plant Slot Permissions
    PermissionPlantSlotView     Permission = "plantslot.view"
    PermissionPlantSlotCreate   Permission = "plantslot.create"
    PermissionPlantSlotUpdate   Permission = "plantslot.update"
    PermissionPlantSlotDelete   Permission = "plantslot.delete"
    PermissionPlantSlotAssign   Permission = "plantslot.assign"
    
    // Plant Management Permissions
    PermissionPlantView         Permission = "plant.view"
    PermissionPlantCreate       Permission = "plant.create"
    PermissionPlantUpdate       Permission = "plant.update"
    PermissionPlantDelete       Permission = "plant.delete"
    PermissionPlantCare         Permission = "plant.care"
    
    // Harvest Permissions
    PermissionHarvestView       Permission = "harvest.view"
    PermissionHarvestCreate     Permission = "harvest.create"
    PermissionHarvestUpdate     Permission = "harvest.update"
    PermissionHarvestDelete     Permission = "harvest.delete"
    PermissionHarvestProcess    Permission = "harvest.process"
    
    // KYC Permissions
    PermissionKYCView          Permission = "kyc.view"
    PermissionKYCVerify        Permission = "kyc.verify"
    PermissionKYCReject        Permission = "kyc.reject"
    
    // Catalog Management Permissions
    PermissionCatalogView      Permission = "catalog.view"
    PermissionCatalogManage    Permission = "catalog.manage"
    PermissionPlantTypeManage  Permission = "planttype.manage"
    
    // Payment Permissions
    PermissionPaymentView      Permission = "payment.view"
    PermissionPaymentProcess   Permission = "payment.process"
    PermissionPaymentRefund    Permission = "payment.refund"
    
    // Analytics Permissions
    PermissionStatsView        Permission = "stats.view"
    PermissionReportsGenerate  Permission = "reports.generate"
    
    // Compliance Permissions
    PermissionComplianceView   Permission = "compliance.view"
    PermissionComplianceManage Permission = "compliance.manage"
    
    // NFT Permissions
    PermissionNFTView          Permission = "nft.view"
    PermissionNFTMint          Permission = "nft.mint"
    PermissionNFTTransfer      Permission = "nft.transfer"
)

// Role Definitions (enhanced from base patterns)
var (
    // Cannabis Club Member Role
    RoleMember = []Permission{
        PermissionMemberView,
        PermissionMemberUpdate,
        PermissionMembershipView,
        PermissionPlantSlotView,
        PermissionPlantView,
        PermissionPlantUpdate,
        PermissionPlantCare,
        PermissionHarvestView,
        PermissionPaymentView,
        PermissionNFTView,
    }
    
    // Cannabis Club Staff Role
    RoleStaff = append(RoleMember, []Permission{
        PermissionMembershipUpdate,
        PermissionPlantSlotUpdate,
        PermissionPlantSlotAssign,
        PermissionPlantCreate,
        PermissionPlantDelete,
        PermissionHarvestUpdate,
        PermissionHarvestProcess,
        PermissionKYCView,
    }...)
    
    // Cannabis Club Manager Role
    RoleManager = append(RoleStaff, []Permission{
        PermissionMemberCreate,
        PermissionMemberDelete,
        PermissionMembershipCreate,
        PermissionMembershipDelete,
        PermissionMembershipTransfer,
        PermissionPlantSlotCreate,
        PermissionPlantSlotDelete,
        PermissionHarvestCreate,
        PermissionHarvestDelete,
        PermissionKYCVerify,
        PermissionKYCReject,
        PermissionPaymentProcess,
        PermissionPaymentRefund,
        PermissionStatsView,
        PermissionNFTMint,
    }...)
    
    // Cannabis Club Admin Role
    RoleAdmin = append(RoleManager, []Permission{
        PermissionCatalogManage,
        PermissionPlantTypeManage,
        PermissionComplianceView,
        PermissionComplianceManage,
        PermissionReportsGenerate,
        PermissionNFTTransfer,
        PermissionUserView,
        PermissionUserCreate,
        PermissionUserUpdate,
        PermissionUserDelete,
        PermissionRoleView,
        PermissionRoleCreate,
        PermissionRoleUpdate,
        PermissionRoleDelete,
    }...)
)
```

**Enhanced Security Middleware** (extending existing `route/index.go`):

```go
// Enhanced Authentication Middleware (route/index.go)
func (s middleware) BearerAuth(permissions ...enum.Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get token from header (existing pattern)
        token := c.GetHeader("Authorization")
        if token == "" || !strings.HasPrefix(token, "Bearer ") {
            c.Error(ecode.Unauthorized)
            return
        }
        
        // Validate token (existing OAuth implementation)
        session, err := s.oauth.ValidateToken(c.Request.Context(), strings.TrimPrefix(token, "Bearer "))
        if err != nil {
            c.Error(ecode.Unauthorized.Desc(err))
            return
        }
        
        // Enhanced: Check if member (not just user)
        if isMemberRoute(c.Request.URL.Path) {
            member, err := s.store.GetMember(c.Request.Context(), session.UserID)
            if err != nil {
                c.Error(ecode.Forbidden.Desc("member account required"))
                return
            }
            
            // Check KYC status for sensitive operations
            if requiresKYC(c.Request.URL.Path) && member.KYCStatus != "verified" {
                c.Error(ecode.KYCVerificationRequired)
                return
            }
            
            // Check membership status for member-only operations
            if requiresActiveMembership(c.Request.URL.Path) {
                if member.CurrentMembershipID == "" {
                    c.Error(ecode.MembershipExpired.Desc("no active membership"))
                    return
                }
                
                membership, err := s.store.GetMembership(c.Request.Context(), member.CurrentMembershipID)
                if err != nil || membership.Status != "active" || membership.ExpirationDate.Before(time.Now()) {
                    c.Error(ecode.MembershipExpired)
                    return
                }
                
                c.Set("membership", membership)
            }
            
            c.Set("member", member)
        }
        
        // Permission checking (existing pattern enhanced)
        if len(permissions) > 0 {
            userPermissions := s.GetUserPermissions(session)
            for _, required := range permissions {
                if !slices.Contains(userPermissions, required) {
                    c.Error(ecode.Forbidden.Desc(fmt.Sprintf("missing permission: %s", required)))
                    return
                }
            }
        }
        
        // Rate limiting (existing pattern enhanced)
        if s.Limiter(c, session) {
            c.Error(ecode.TooManyRequests)
            return
        }
        
        c.Set("session", session)
        c.Next()
    }
}

// Helper functions for enhanced security checks
func isMemberRoute(path string) bool {
    memberPaths := []string{"/api/v1/members", "/api/v1/memberships", "/api/v1/plantslots", "/api/v1/plants", "/api/v1/harvests"}
    for _, memberPath := range memberPaths {
        if strings.HasPrefix(path, memberPath) {
            return true
        }
    }
    return false
}

func requiresKYC(path string) bool {
    kycRequiredPaths := []string{"/api/v1/memberships", "/api/v1/plantslots", "/api/v1/plants", "/api/v1/harvests", "/api/v1/payments"}
    for _, kycPath := range kycRequiredPaths {
        if strings.HasPrefix(path, kycPath) {
            return true
        }
    }
    return false
}

func requiresActiveMembership(path string) bool {
    membershipPaths := []string{"/api/v1/plantslots", "/api/v1/plants", "/api/v1/harvests"}
    for _, membershipPath := range membershipPaths {
        if strings.HasPrefix(path, membershipPath) {
            return true
        }
    }
    return false
}

// Enhanced Rate Limiting (extending existing pattern)
var (
    defaultLimit    = redis_rate.PerMinute(100)
    authLimit       = redis_rate.PerMinute(10)
    uploadLimit     = redis_rate.PerHour(50)
    kycLimit        = redis_rate.PerDay(5)      // KYC submissions
    paymentLimit    = redis_rate.PerHour(10)    // Payment attempts
    plantCareLimit  = redis_rate.PerHour(60)    // Plant care updates
    harvestLimit    = redis_rate.PerDay(20)     // Harvest operations
)

func (s middleware) getRateLimit(path string) redis_rate.Limit {
    switch {
    case strings.Contains(path, "/auth/"):
        return authLimit
    case strings.Contains(path, "/kyc/"):
        return kycLimit
    case strings.Contains(path, "/payments/"):
        return paymentLimit
    case strings.Contains(path, "/storage/"):
        return uploadLimit
    case strings.Contains(path, "/plants/") && strings.Contains(path, "/care"):
        return plantCareLimit
    case strings.Contains(path, "/harvests/"):
        return harvestLimit
    default:
        return defaultLimit
    }
}
```

**Enhanced Audit Logging** (extending existing audit system):

```go
// Enhanced Audit Logging (route/index.go)
func (s middleware) AuditLog(c *gin.Context, module string, action enum.DataAction, data interface{}, entityID string) {
    session, exists := c.Get("session")
    if !exists {
        return
    }
    
    authSession := session.(*db.AuthSessionDto)
    
    // Enhanced audit log with cannabis-specific fields
    log := &db.AuditLog{
        Module:      module,
        URL:         c.Request.URL.Path,
        Method:      c.Request.Method,
        Action:      action,
        EntityID:    entityID,
        UserID:      authSession.UserID,
        TenantID:    authSession.TenantID,
        IPAddress:   c.ClientIP(),
        UserAgent:   c.Request.UserAgent(),
        RequestID:   c.GetHeader("X-Request-ID"),
        Data:        data,
        CreatedAt:   time.Now(),
    }
    
    // Add member-specific context if available
    if member, exists := c.Get("member"); exists {
        memberData := member.(*Member)
        log.MemberID = memberData.ID
        log.KYCStatus = memberData.KYCStatus
    }
    
    // Add membership context if available
    if membership, exists := c.Get("membership"); exists {
        membershipData := membership.(*Membership)
        log.MembershipID = membershipData.ID
        log.MembershipStatus = membershipData.Status
    }
    
    // Store audit log asynchronously
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        s.store.Db.AuditLog.Create(ctx, log)
    }()
}

// Cannabis-specific audit events
const (
    AuditModuleMember      = "member"
    AuditModuleMembership  = "membership"
    AuditModulePlantSlot   = "plant_slot"
    AuditModulePlant       = "plant"
    AuditModuleHarvest     = "harvest"
    AuditModuleKYC         = "kyc"
    AuditModulePayment     = "payment"
    AuditModuleNFT         = "nft"
    AuditModuleCompliance  = "compliance"
)
```

### WebSocket Implementation (Enhanced from Base)

**Enhanced WebSocket Service** (extending existing `pkg/ws/index.go`):

```go
// Enhanced WebSocket Service (pkg/ws/index.go)
type CannabisWs struct {
    *ws.Ws
    store *store.Store
}

// Cannabis-specific WebSocket message types
type WSMessageType string

const (
    WSMessagePlantStateChange    WSMessageType = "plant_state_change"
    WSMessageHarvestReady       WSMessageType = "harvest_ready"
    WSMessageMembershipExpiring WSMessageType = "membership_expiring"
    WSMessageKYCStatusUpdate    WSMessageType = "kyc_status_update"
    WSMessagePaymentUpdate      WSMessageType = "payment_update"
    WSMessageCareReminder       WSMessageType = "care_reminder"
    WSMessageSlotAvailable      WSMessageType = "slot_available"
    WSMessageHarvestScheduled   WSMessageType = "harvest_scheduled"
)

type WSMessage struct {
    Type      WSMessageType `json:"type"`
    Data      interface{}   `json:"data"`
    Timestamp time.Time     `json:"timestamp"`
    MessageID string        `json:"message_id"`
}

// Enhanced notification methods
func (cws *CannabisWs) NotifyPlantStateChange(ctx context.Context, plant *Plant, oldState, newState string) {
    member, err := cws.store.GetMember(ctx, plant.MemberID)
    if err != nil {
        return
    }
    
    message := WSMessage{
        Type: WSMessagePlantStateChange,
        Data: map[string]interface{}{
            "plant_id":     plant.ID,
            "plant_code":   plant.PlantCode,
            "old_state":    oldState,
            "new_state":    newState,
            "message":      fmt.Sprintf("Your plant %s has moved from %s to %s stage", plant.PlantCode, oldState, newState),
        },
        Timestamp: time.Now(),
        MessageID: uuid.New().String(),
    }
    
    data, _ := json.Marshal(message)
    cws.EmitTo([]string{member.ID}, data)
    
    // Cache notification for offline users
    cws.store.CacheNotification(ctx, member.ID, &Notification{
        Type:      string(WSMessagePlantStateChange),
        Message:   message.Data.(map[string]interface{})["message"].(string),
        Data:      message.Data,
        CreatedAt: time.Now(),
    })
}

func (cws *CannabisWs) NotifyHarvestReady(ctx context.Context, plant *Plant) {
    member, err := cws.store.GetMember(ctx, plant.MemberID)
    if err != nil {
        return
    }
    
    message := WSMessage{
        Type: WSMessageHarvestReady,
        Data: map[string]interface{}{
            "plant_id":            plant.ID,
            "plant_code":          plant.PlantCode,
            "estimated_yield":     plant.EstimatedYield,
            "harvest_window_days": 7,
            "message":            fmt.Sprintf("🌿 Your plant %s is ready for harvest! Estimated yield: %.1fg", plant.PlantCode, plant.EstimatedYield),
        },
        Timestamp: time.Now(),
        MessageID: uuid.New().String(),
    }
    
    data, _ := json.Marshal(message)
    cws.EmitTo([]string{member.ID}, data)
}

func (cws *CannabisWs) NotifyMembershipExpiring(ctx context.Context, membership *Membership, daysUntilExpiry int) {
    member, err := cws.store.GetMember(ctx, membership.MemberID)
    if err != nil {
        return
    }
    
    message := WSMessage{
        Type: WSMessageMembershipExpiring,
        Data: map[string]interface{}{
            "membership_id":       membership.ID,
            "expiration_date":     membership.ExpirationDate,
            "days_until_expiry":   daysUntilExpiry,
            "auto_renewal":        membership.AutoRenewal,
            "message":            fmt.Sprintf("⚠️ Your membership expires in %d days. Renew now to avoid interruption.", daysUntilExpiry),
        },
        Timestamp: time.Now(),
        MessageID: uuid.New().String(),
    }
    
    data, _ := json.Marshal(message)
    cws.EmitTo([]string{member.ID}, data)
}
```

### Blockchain Integration

**NFT Service Implementation**:

```go
// NFT Service (pkg/blockchain/nft.go)
type NFTService struct {
    store         *store.Store
    client        *ethclient.Client
    contractAddr  common.Address
    privateKey    *ecdsa.PrivateKey
    chainID       *big.Int
}

type PlantSlotNFT struct {
    TokenID         string                 `json:"token_id"`
    ContractAddress string                 `json:"contract_address"`
    OwnerAddress    string                 `json:"owner_address"`
    Metadata        PlantSlotNFTMetadata   `json:"metadata"`
    MintTransaction string                 `json:"mint_transaction"`
    MintedAt        time.Time              `json:"minted_at"`
}

type PlantSlotNFTMetadata struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Image       string         `json:"image"`
    ExternalURL string         `json:"external_url"`
    Attributes  []NFTAttribute `json:"attributes"`
}

type NFTAttribute struct {
    TraitType   string      `json:"trait_type"`
    Value       interface{} `json:"value"`
    DisplayType string      `json:"display_type,omitempty"`
}

func (nft *NFTService) MintPlantSlotNFT(ctx context.Context, slot *PlantSlot) (*PlantSlotNFT, error) {
    // Generate unique token ID
    tokenID := big.NewInt(time.Now().Unix())
    
    // Create metadata
    metadata := PlantSlotNFTMetadata{
        Name:        fmt.Sprintf("Seed eG Plant Slot #%s", slot.SlotNumber),
        Description: fmt.Sprintf("Cannabis cultivation slot at %s, Area %s. This NFT represents your right to cultivate cannabis in this designated slot.", slot.FarmLocation, slot.AreaDesignation),
        Image:       fmt.Sprintf("https://nft.seedeg.com/images/slot_%s.png", slot.SlotNumber),
        ExternalURL: fmt.Sprintf("https://app.seedeg.com/slots/%s", slot.ID),
        Attributes: []NFTAttribute{
            {TraitType: "Farm Location", Value: slot.FarmLocation},
            {TraitType: "Area", Value: slot.AreaDesignation},
            {TraitType: "Slot Number", Value: slot.SlotNumber},
            {TraitType: "Season", Value: slot.CatalogID},
            {TraitType: "Allocation Date", Value: slot.CreatedAt.Format("2006-01-02"), DisplayType: "date"},
            {TraitType: "Status", Value: slot.Status},
        },
    }
    
    // Upload metadata to IPFS or centralized storage
    metadataURI, err := nft.uploadMetadata(ctx, metadata)
    if err != nil {
        return nil, err
    }
    
    // Get member's wallet address
    member, err := nft.store.GetMember(ctx, slot.MemberID)
    if err != nil {
        return nil, err
    }
    
    if member.WalletAddress == "" {
        return nil, fmt.Errorf("member has no wallet address")
    }
    
    // Prepare transaction
    auth, err := bind.NewKeyedTransactorWithChainID(nft.privateKey, nft.chainID)
    if err != nil {
        return nil, err
    }
    
    // Mint NFT on blockchain
    tx, err := nft.contract.Mint(auth, common.HexToAddress(member.WalletAddress), tokenID, big.NewInt(1), []byte(metadataURI))
    if err != nil {
        return nil, err
    }
    
    // Wait for transaction confirmation
    receipt, err := bind.WaitMined(ctx, nft.client, tx)
    if err != nil {
        return nil, err
    }
    
    // Create NFT record
    nftRecord := &PlantSlotNFT{
        TokenID:         tokenID.String(),
        ContractAddress: nft.contractAddr.Hex(),
        OwnerAddress:    member.WalletAddress,
        Metadata:        metadata,
        MintTransaction: tx.Hash().Hex(),
        MintedAt:        time.Now(),
    }
    
    // Store NFT record in database
    dbRecord := &NFTRecord{
        PlantSlotID:      slot.ID,
        MemberID:         member.ID,
        TenantID:         slot.TenantID,
        TokenID:          tokenID.String(),
        ContractAddress:  nft.contractAddr.Hex(),
        ChainID:          nft.chainID.Int64(),
        MetadataURI:      metadataURI,
        Metadata:         metadata,
        MintTransactionHash: tx.Hash().Hex(),
        MintBlockNumber:     receipt.BlockNumber.Int64(),
        MintGasUsed:         receipt.GasUsed,
        Status:           "minted",
        OwnerAddress:     member.WalletAddress,
        CreatedAt:        time.Now(),
    }
    
    if err := nft.store.Db.NFTRecord.Create(ctx, dbRecord); err != nil {
        return nil, err
    }
    
    return nftRecord, nil
}

func (nft *NFTService) uploadMetadata(ctx context.Context, metadata PlantSlotNFTMetadata) (string, error) {
    // Convert metadata to JSON
    metadataJSON, err := json.Marshal(metadata)
    if err != nil {
        return "", err
    }
    
    // Upload to MinIO storage
    fileName := fmt.Sprintf("nft_metadata_%d.json", time.Now().Unix())
    
    info, err := nft.store.Storage.Client.PutObject(
        ctx,
        NFTMetadataBucket,
        fileName,
        bytes.NewReader(metadataJSON),
        int64(len(metadataJSON)),
        minio.PutObjectOptions{
            ContentType: "application/json",
        },
    )
    
    if err != nil {
        return "", err
    }
    
    // Return public URL
    return fmt.Sprintf("https://storage.seedeg.com/%s/%s", NFTMetadataBucket, info.Key), nil
}
```

### Development and Deployment

**Environment Configuration** (extending existing `env/index.go`):

```go
// Enhanced Environment Variables (env/index.go)
var (
    // Existing variables
    Port         = getEnv("PORT", "3000")
    Environment  = getEnv("ENVIRONMENT", "development")
    MongoUri     = getEnv("MONGO_URI", "mongodb://localhost:27017")
    RedisUri     = getEnv("REDIS_URI", "redis://localhost:6379")
    MinioUri     = getEnv("MINIO_URI", "localhost:9000")
    
    // Cannabis Club specific variables
    ClubName             = getEnv("CLUB_NAME", "Seed eG")
    ClubLicense          = getEnv("CLUB_LICENSE", "CULT_2025_001")
    
    // Stripe Configuration
    StripeSecretKey      = getEnv("STRIPE_SECRET_KEY", "")
    StripeWebhookSecret  = getEnv("STRIPE_WEBHOOK_SECRET", "")
    StripePublishableKey = getEnv("STRIPE_PUBLISHABLE_KEY", "")
    
    // Blockchain Configuration
    EthereumRpcUrl       = getEnv("ETHEREUM_RPC_URL", "https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    NFTContractAddress   = getEnv("NFT_CONTRACT_ADDRESS", "")
    BlockchainPrivateKey = getEnv("BLOCKCHAIN_PRIVATE_KEY", "")
    ChainID              = getEnvInt("CHAIN_ID", 1)
    
    // Email Configuration (enhanced)
    SMTPHost             = getEnv("SMTP_HOST", "localhost")
    SMTPPort             = getEnvInt("SMTP_PORT", 587)
    SMTPUsername         = getEnv("SMTP_USERNAME", "")
    SMTPPassword         = getEnv("SMTP_PASSWORD", "")
    EmailFromAddress     = getEnv("EMAIL_FROM_ADDRESS", "noreply@seedeg.com")
    EmailFromName        = getEnv("EMAIL_FROM_NAME", "Seed eG")
    
    // File Storage Configuration
    StorageEndpoint      = getEnv("STORAGE_ENDPOINT", "localhost:9000")
    StorageAccessKey     = getEnv("STORAGE_ACCESS_KEY", "minioadmin")
    StorageSecretKey     = getEnv("STORAGE_SECRET_KEY", "minioadmin")
    StorageUseSSL        = getEnvBool("STORAGE_USE_SSL", false)
    StorageRegion        = getEnv("STORAGE_REGION", "us-east-1")
    
    // Security Configuration
    JWTSecret            = getEnv("JWT_SECRET", "your-secret-key")
    EncryptionKey        = getEnv("ENCRYPTION_KEY", "your-encryption-key")
    
    // Feature Flags
    EnableNFTMinting     = getEnvBool("ENABLE_NFT_MINTING", true)
    EnableKYCVerification = getEnvBool("ENABLE_KYC_VERIFICATION", true)
    EnableEmailNotifications = getEnvBool("ENABLE_EMAIL_NOTIFICATIONS", true)
    EnableWebSocketNotifications = getEnvBool("ENABLE_WEBSOCKET_NOTIFICATIONS", true)
    
    // Compliance Configuration
    MaxPlantsPerMember   = getEnvInt("MAX_PLANTS_PER_MEMBER", 3)
    MaxYieldPerPlant     = getEnvFloat("MAX_YIELD_PER_PLANT", 50.0) // grams
    MinimumMemberAge     = getEnvInt("MINIMUM_MEMBER_AGE", 18)
    
    // Business Configuration
    ProcessingFeePerGram = getEnvFloat("PROCESSING_FEE_PER_GRAM", 0.10) // euros
    MembershipPrice      = getEnvFloat("MEMBERSHIP_PRICE", 120.0)       // euros
    GracePeriodDays      = getEnvInt("GRACE_PERIOD_DAYS", 30)
)

func getEnvInt(key string, fallback int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
    if value := os.Getenv(key); value != "" {
        if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
            return floatValue
        }
    }
    return fallback
}

func getEnvBool(key string, fallback bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return fallback
}
```

**Enhanced Docker Configuration**:

```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      # Database connections
      - MONGO_URI=mongodb://mongo:27017
      - REDIS_URI=redis://redis:6379
      
      # Storage
      - STORAGE_ENDPOINT=minio:9000
      - STORAGE_ACCESS_KEY=minioadmin
      - STORAGE_SECRET_KEY=minioadmin
      - STORAGE_USE_SSL=false
      
      # Cannabis Club Configuration
      - CLUB_NAME=Seed eG Demo
      - CLUB_LICENSE=DEMO_2025_001
      
      # Payment Processing
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
      - STRIPE_PUBLISHABLE_KEY=${STRIPE_PUBLISHABLE_KEY}
      
      # Blockchain
      - ETHEREUM_RPC_URL=${ETHEREUM_RPC_URL}
      - NFT_CONTRACT_ADDRESS=${NFT_CONTRACT_ADDRESS}
      - BLOCKCHAIN_PRIVATE_KEY=${BLOCKCHAIN_PRIVATE_KEY}
      - CHAIN_ID=1
      
      # Email Configuration
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=587
      - SMTP_USERNAME=${SMTP_USERNAME}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
      - EMAIL_FROM_ADDRESS=noreply@seedeg.com
      - EMAIL_FROM_NAME=Seed eG
      
      # Security
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
      
      # Feature Flags
      - ENABLE_NFT_MINTING=true
      - ENABLE_KYC_VERIFICATION=true
      - ENABLE_EMAIL_NOTIFICATIONS=true
      
      # Business Rules
      - MAX_PLANTS_PER_MEMBER=3
      - MAX_YIELD_PER_PLANT=50.0
      - PROCESSING_FEE_PER_GRAM=0.10
      - MEMBERSHIP_PRICE=120.0
      
    depends_on:
      - mongo
      - redis
      - minio
    volumes:
      - ./logs:/app/logs
    restart: unless-stopped

  mongo:
    image: mongo:7-jammy
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
      - MONGO_INITDB_DATABASE=seedeg
    volumes:
      - mongo_data:/data/db
      - ./scripts/mongo-init.js:/docker-entrypoint-initdb.d/mongo-init.js:ro
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --requirepass ${REDIS_PASSWORD:-password}
    volumes:
      - redis_data:/data
    restart: unless-stopped

  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=${MINIO_ROOT_USER:-minioadmin}
      - MINIO_ROOT_PASSWORD=${MINIO_ROOT_PASSWORD:-minioadmin}
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    restart: unless-stopped

  # Blockchain node (optional for local development)
  ganache:
    image: trufflesuite/ganache:latest
    ports:
      - "8545:8545"
    command: >
      ganache
      --host 0.0.0.0
      --accounts 10
      --deterministic
      --mnemonic "candy maple cake sugar pudding cream honey rich smooth crumble sweet treat"
    profiles:
      - blockchain

  # Monitoring and observability
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    profiles:
      - monitoring

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    profiles:
      - monitoring

volumes:
  mongo_data:
  redis_data:
  minio_data:
  prometheus_data:
  grafana_data:
```

**Production Kubernetes Configuration**:

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: seedeg-production

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: seedeg-config
  namespace: seedeg-production
data:
  ENVIRONMENT: "production"
  CLUB_NAME: "Seed eG"
  CLUB_LICENSE: "CULT_2025_001"
  MAX_PLANTS_PER_MEMBER: "3"
  MAX_YIELD_PER_PLANT: "50.0"
  PROCESSING_FEE_PER_GRAM: "0.10"
  MEMBERSHIP_PRICE: "120.0"
  GRACE_PERIOD_DAYS: "30"
  MINIMUM_MEMBER_AGE: "18"
  ENABLE_NFT_MINTING: "true"
  ENABLE_KYC_VERIFICATION: "true"
  ENABLE_EMAIL_NOTIFICATIONS: "true"

---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: seedeg-secrets
  namespace: seedeg-production
type: Opaque
stringData:
  MONGO_URI: "mongodb://mongo-service:27017/seedeg"
  REDIS_URI: "redis://redis-service:6379"
  JWT_SECRET: "your-production-jwt-secret"
  ENCRYPTION_KEY: "your-production-encryption-key"
  STRIPE_SECRET_KEY: "sk_live_..."
  STRIPE_WEBHOOK_SECRET: "whsec_..."
  ETHEREUM_RPC_URL: "https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
  NFT_CONTRACT_ADDRESS: "0x..."
  BLOCKCHAIN_PRIVATE_KEY: "0x..."
  SMTP_USERNAME: "your-smtp-username"
  SMTP_PASSWORD: "your-smtp-password"
  STORAGE_ACCESS_KEY: "your-minio-access-key"
  STORAGE_SECRET_KEY: "your-minio-secret-key"

---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: seedeg-app
  namespace: seedeg-production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: seedeg-app
  template:
    metadata:
      labels:
        app: seedeg-app
    spec:
      containers:
      - name: seedeg-app
        image: seedeg/app:latest
        ports:
        - containerPort: 3000
        env:
        - name: PORT
          value: "3000"
        envFrom:
        - configMapRef:
            name: seedeg-config
        - secretRef:
            name: seedeg-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5

---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: seedeg-app-service
  namespace: seedeg-production
spec:
  selector:
    app: seedeg-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: ClusterIP
```

### Monitoring and Observability

**Enhanced Logging** (extending existing logger):

```go
// Enhanced Logging Configuration (pkg/logger/index.go)
type CannabisLogger struct {
    *logrus.Logger
    store *store.Store
}

func NewCannabisLogger(store *store.Store) *CannabisLogger {
    logger := logrus.New()
    
    // Configure logger based on environment
    if env.Environment == "production" {
        logger.SetFormatter(&logrus.JSONFormatter{})
        logger.SetLevel(logrus.InfoLevel)
    } else {
        logger.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
            ForceColors:   true,
        })
        logger.SetLevel(logrus.DebugLevel)
    }
    
    return &CannabisLogger{
        Logger: logger,
        store:  store,
    }
}

// Cannabis-specific logging methods
func (cl *CannabisLogger) LogMemberActivity(ctx context.Context, memberID string, activity string, data interface{}) {
    cl.WithFields(logrus.Fields{
        "member_id": memberID,
        "activity":  activity,
        "data":      data,
        "timestamp": time.Now(),
    }).Info("Member activity")
}

func (cl *CannabisLogger) LogPlantStateChange(ctx context.Context, plantID string, oldState, newState string) {
    cl.WithFields(logrus.Fields{
        "plant_id":  plantID,
        "old_state": oldState,
        "new_state": newState,
        "timestamp": time.Now(),
    }).Info("Plant state changed")
}

func (cl *CannabisLogger) LogHarvestEvent(ctx context.Context, harvestID string, event string, yield float64) {
    cl.WithFields(logrus.Fields{
        "harvest_id": harvestID,
        "event":      event,
        "yield":      yield,
        "timestamp":  time.Now(),
    }).Info("Harvest event")
}

func (cl *CannabisLogger) LogComplianceEvent(ctx context.Context, eventType string, details interface{}) {
    cl.WithFields(logrus.Fields{
        "compliance_event": eventType,
        "details":          details,
        "timestamp":        time.Now(),
    }).Warn("Compliance event")
}
```

**Metrics Collection**:

```go
// Metrics Service (pkg/metrics/index.go)
type MetricsService struct {
    store *store.Store
}

type DashboardMetrics struct {
    TotalMembers        int64   `json:"total_members"`
    ActiveMemberships   int64   `json:"active_memberships"`
    TotalPlants         int64   `json:"total_plants"`
    PlantsInVeg         int64   `json:"plants_in_veg"`
    PlantsInFlower      int64   `json:"plants_in_flower"`
    PlantsReadyHarvest  int64   `json:"plants_ready_harvest"`
    TotalHarvests       int64   `json:"total_harvests"`
    TotalYield          float64 `json:"total_yield_kg"`
    AvgYieldPerPlant    float64 `json:"avg_yield_per_plant"`
    PendingKYC          int64   `json:"pending_kyc"`
    MonthlyRevenue      float64 `json:"monthly_revenue"`
    MemberRetentionRate float64 `json:"member_retention_rate"`
}

func (ms *MetricsService) GetDashboardMetrics(ctx context.Context, tenantID string) (*DashboardMetrics, error) {
    // Use existing aggregation patterns from MongoDB
    pipeline := []bson.M{
        {"$match": bson.M{"tenant_id": tenantID}},
        {"$group": bson.M{
            "_id": nil,
            "total_members": bson.M{"$sum": 1},
            "active_members": bson.M{
                "$sum": bson.M{
                    "$cond": []interface{}{
                        bson.M{"$eq": []interface{}{"$current_membership_id", ""}},
                        0, 1,
                    },
                },
            },
            "pending_kyc": bson.M{
                "$sum": bson.M{
                    "$cond": []interface{}{
                        bson.M{"$eq": []interface{}{"$kyc_status", "pending"}},
                        1, 0,
                    },
                },
            },
        }},
    }
    
    var memberStats []bson.M
    cursor, err := ms.store.Db.Member.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    if err := cursor.All(ctx, &memberStats); err != nil {
        return nil, err
    }
    
    // Similar aggregations for plants, harvests, etc.
    // ... (implementation continues with other metrics)
    
    return &DashboardMetrics{
        TotalMembers:      memberStats[0]["total_members"].(int64),
        ActiveMemberships: memberStats[0]["active_members"].(int64),
        PendingKYC:       memberStats[0]["pending_kyc"].(int64),
        // ... populate other fields
    }, nil
}
```

### Testing Strategy

**Unit Tests** (following existing patterns):

```go
// tests/member_service_test.go
package tests

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "seedeg/internal/service"
    "seedeg/pkg/ecode"
)

type MockStore struct {
    mock.Mock
}

func (m *MockStore) GetMember(ctx context.Context, id string) (*Member, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Member), args.Error(1)
}

func TestMemberService_CreateMembership(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateMembershipRequest
        setup   func(*MockStore)
        want    *Membership
        wantErr error
    }{
        {
            name: "successful membership creation",
            input: &CreateMembershipRequest{
                MemberID:       "member_123",
                MembershipType: "annual",
                CatalogID:      "catalog_456",
                PaymentAmount:  120.0,
                TenantID:       "tenant_789",
            },
            setup: func(store *MockStore) {
                member := &Member{
                    ID:        "member_123",
                    KYCStatus: "verified",
                }
                store.On("GetMember", mock.Anything, "member_123").Return(member, nil)
            },
            want: &Membership{
                MemberID:       "member_123",
                MembershipType: "annual",
                Status:         "pending_payment",
                PaymentAmount:  120.0,
            },
            wantErr: nil,
        },
        {
            name: "KYC not verified",
            input: &CreateMembershipRequest{
                MemberID: "member_456",
            },
            setup: func(store *MockStore) {
                member := &Member{
                    ID:        "member_456",
                    KYCStatus: "pending",
                }
                store.On("GetMember", mock.Anything, "member_456").Return(member, nil)
            },
            want:    nil,
            wantErr: ecode.KYCVerificationRequired,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockStore := &MockStore{}
            tt.setup(mockStore)
            
            service := &service.MembershipService{
                Store: mockStore,
            }
            
            got, err := service.CreateMembership(context.Background(), tt.input)
            
            if tt.wantErr != nil {
                assert.Error(t, err)
                assert.Equal(t, tt.wantErr, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want.MemberID, got.MemberID)
                assert.Equal(t, tt.want.Status, got.Status)
            }
            
            mockStore.AssertExpectations(t)
        })
    }
}
```

**Integration Tests**:

```go
// tests/integration/api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "seedeg/internal/route"
)

func TestMembershipAPI(t *testing.T) {
    // Setup test database and dependencies
    store := setupTestStore(t)
    defer teardownTestStore(t, store)
    
    // Create test member with verified KYC
    member := createTestMember(t, store, "verified")
    token := generateTestToken(t, member.ID)
    
    // Setup Gin router
    gin.SetMode(gin.TestMode)
    router := gin.New()
    route.RegisterCannabisRoutes(router, store)
    
    t.Run("Create Membership", func(t *testing.T) {
        payload := map[string]interface{}{
            "catalog_id":      "catalog_123",
            "membership_type": "annual",
            "payment_amount":  120.0,
        }
        
        body, _ := json.Marshal(payload)
        req := httptest.NewRequest("POST", "/api/v1/memberships", bytes.NewBuffer(body))
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, "pending_payment", response["status"])
    })
    
    t.Run("Get Member Memberships", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/memberships", nil)
        req.Header.Set("Authorization", "Bearer "+token)
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response []map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Greater(t, len(response), 0)
    })
}
```

### Database Migration Strategy

**Migration System** (following existing patterns):

```go
// migrations/migration.go
package migrations

import (
    "context"
    "time"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Migration struct {
    Version     int
    Description string
    Up          func(ctx context.Context, db *mongo.Database) error
    Down        func(ctx context.Context, db *mongo.Database) error
}

var migrations = []Migration{
    {
        Version:     1,
        Description: "Create initial cannabis club collections",
        Up:          migration_001_initial_collections,
        Down:        migration_001_down,
    },
    {
        Version:     2,
        Description: "Add indexes for cannabis collections",
        Up:          migration_002_add_indexes,
        Down:        migration_002_down,
    },
    {
        Version:     3,
        Description: "Add NFT tracking fields",
        Up:          migration_003_nft_fields,
        Down:        migration_003_down,
    },
}

func migration_001_initial_collections(ctx context.Context, db *mongo.Database) error {
    // Create collections with validation
    collections := []string{
        "member", "membership", "plant_slot", "plant", 
        "seasonal_catalog", "plant_type", "payment",
        "plant_care_history", "harvest_details", "nft_record",
    }
    
    for _, collName := range collections {
        opts := options.CreateCollection()
        if err := db.CreateCollection(ctx, collName, opts); err != nil {
            // Ignore "collection already exists" error
            if !mongo.IsDuplicateKeyError(err) {
                return err
            }
        }
    }
    
    return nil
}

func migration_002_add_indexes(ctx context.Context, db *mongo.Database) error {
    // Member collection indexes
    memberIndexes := []mongo.IndexModel{
        {Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "kyc_status", Value: 1}}},
        {Keys: bson.D{{Key: "current_membership_id", Value: 1}}},
    }
    
    if _, err := db.Collection("member").Indexes().CreateMany(ctx, memberIndexes); err != nil {
        return err
    }
    
    // Membership collection indexes
    membershipIndexes := []mongo.IndexModel{
        {Keys: bson.D{{Key: "tenant_id", Value: 1}}},
        {Keys: bson.D{{Key: "member_id", Value: 1}}},
        {Keys: bson.D{{Key: "status", Value: 1}}},
        {Keys: bson.D{{Key: "expiration_date", Value: 1}}},
        {Keys: bson.D{{Key: "stripe_payment_intent_id", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
    }
    
    if _, err := db.Collection("membership").Indexes().CreateMany(ctx, membershipIndexes); err != nil {
        return err
    }
    
    // Continue for other collections...
    return nil
}

// Migration runner
func RunMigrations(ctx context.Context, db *mongo.Database) error {
    // Get current version
    currentVersion := getCurrentMigrationVersion(ctx, db)
    
    for _, migration := range migrations {
        if migration.Version > currentVersion {
            if err := migration.Up(ctx, db); err != nil {
                return err
            }
            
            // Update migration version
            if err := setMigrationVersion(ctx, db, migration.Version); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

### Performance Optimization

**Database Query Optimization**:

```go
// Optimized query patterns using aggregation pipelines
func (s *Store) GetMemberDashboard(ctx context.Context, memberID string) (*MemberDashboard, error) {
    // Use single aggregation pipeline to get all member data
    pipeline := []bson.M{
        {"$match": bson.M{"_id": memberID}},
        {"$lookup": bson.M{
            "from":         "membership",
            "localField":   "current_membership_id",
            "foreignField": "_id",
            "as":           "membership",
        }},
        {"$lookup": bson.M{
            "from":         "plant_slot",
            "localField":   "current_membership_id",
            "foreignField": "membership_id",
            "as":           "slots",
        }},
        {"$lookup": bson.M{
            "from":         "plant",
            "localField":   "slots._id",
            "foreignField": "plant_slot_id",
            "as":           "plants",
        }},
        {"$project": bson.M{
            "member": "$ROOT",
            "membership": bson.M{"$arrayElemAt": []interface{}{"$membership", 0}},
            "total_slots": bson.M{"$size": "$slots"},
            "active_plants": bson.M{
                "$size": bson.M{
                    "$filter": bson.M{
                        "input": "$plants",
                        "cond":  bson.M{"$in": []interface{}{"$this.state", []string{"seedling", "vegetative", "flowering"}}},
                    },
                },
            },
            "ready_for_harvest": bson.M{
                "$size": bson.M{
                    "$filter": bson.M{
                        "input": "$plants",
                        "cond":  bson.M{"$eq": []interface{}{"$this.state", "ready_for_harvest"}},
                    },
                },
            },
        }},
    }
    
    var result []MemberDashboard
    cursor, err := s.Db.Member.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    if err := cursor.All(ctx, &result); err != nil {
        return nil, err
    }
    
    if len(result) == 0 {
        return nil, mongo.ErrNoDocuments
    }
    
    return &result[0], nil
}
```

### Security Best Practices

**Data Encryption**:

```go
// Enhanced encryption service (pkg/crypto/index.go)
type CryptoService struct {
    key []byte
}

func NewCryptoService(key string) *CryptoService {
    return &CryptoService{
        key: []byte(key),
    }
}

func (cs *CryptoService) EncryptSensitiveData(data string) (string, error) {
    block, err := aes.NewCipher(cs.key)
    if err != nil {
        return "", err
    }
    
    // Use GCM for authenticated encryption
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (cs *CryptoService) DecryptSensitiveData(encryptedData string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encryptedData)
    if err != nil {
        return "", err
    }
    
    block, err := aes.NewCipher(cs.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}
```

### Conclusion

This comprehensive architecture document provides a complete implementation guide for the Seed eG Cannabis Club Platform, built upon the solid foundation of the existing Go-based backend infrastructure. The architecture maximizes the use of established patterns from the base system while extending functionality to meet the specific needs of cannabis club management, including:

- **Member Management**: Enhanced user system with KYC verification
- **Membership System**: Subscription-based access with payment integration
- **Plant Tracking**: Complete lifecycle management from seed to harvest
- **NFT Integration**: Blockchain-based ownership tokens for plant slots
- **Compliance**: Built-in regulatory compliance and audit trails
- **Real-time Updates**: WebSocket-based notifications and updates

The system is designed to be scalable, secure, and compliant with German cannabis regulations while providing an excellent user experience for both members and club administrators. All components follow established patterns from the existing codebase, ensuring consistency and maintainability.