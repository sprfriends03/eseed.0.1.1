# Plant Management API

## Overview

The Plant Management API provides comprehensive functionality for managing the complete cannabis plant lifecycle within the cultivation facility. It enables members to create, track, and manage their plants while providing administrators with full oversight capabilities for plant health monitoring, analytics, and harvest management.

## Features

- **Complete Plant Lifecycle Management**: Track plants from seedling through harvest
- **Health & Care Monitoring**: Record care activities and track plant health improvements
- **Status Transition System**: Manage plant development stages with validation
- **Harvest Management**: Automated harvest readiness detection and processing
- **Image Upload Support**: Document plant growth with photos
- **Administrative Analytics**: Comprehensive reporting and health alert system
- **Integration Points**: Seamless integration with plant slots, plant types, and member systems

## Endpoints Summary

### Member Endpoints
- `GET /plants/v1/my-plants` - Get current member's plants with filtering
- `POST /plants/v1/create` - Create new plant in available slot
- `GET /plants/v1/{id}` - Get detailed plant information
- `PUT /plants/v1/{id}/status` - Update plant lifecycle status
- `PUT /plants/v1/{id}/care` - Record care activities and measurements
- `POST /plants/v1/{id}/images` - Upload plant growth images
- `POST /plants/v1/{id}/harvest` - Process plant harvest

### Administrative Endpoints
- `GET /plants/v1/admin/all` - Get all plants with advanced filtering
- `GET /plants/v1/admin/analytics` - Get comprehensive plant analytics
- `GET /plants/v1/admin/health-alerts` - Get plants requiring attention
- `PUT /plants/v1/admin/{id}/force-status` - Force status change (admin override)
- `GET /plants/v1/admin/harvest-ready` - Get plants ready for harvest

## Plant Lifecycle Statuses

The plant system uses the following status lifecycle with strict transition rules:

| Status | Description | Allowed Transitions | Typical Duration |
|--------|-------------|-------------------|------------------|
| `seedling` | Young plant developing first leaves | → `vegetative`, `dead` | 2-3 weeks |
| `vegetative` | Active growth phase | → `flowering`, `dead` | 4-8 weeks |
| `flowering` | Flowering/budding phase | → `harvested`, `dead` | 8-12 weeks |
| `harvested` | Successfully harvested | → `seedling` (new cycle) | Final state |
| `dead` | Plant died or removed | → `seedling` (replacement) | Final state |

## Authentication & Permissions

All endpoints require bearer token authentication. The following permissions are required:

| Permission | Description | Endpoints |
|------------|-------------|-----------|
| `plant_view` | View own plant information | GET my-plants, GET plant details |
| `plant_create` | Create new plants in owned slots | POST create |
| `plant_update` | Update own plant status and data | PUT status, POST images |
| `plant_care` | Record care activities | PUT care |
| `plant_harvest` | Harvest own plants | POST harvest |
| `plant_manage` | Admin: Full plant management access | All admin endpoints |

## Business Rules

### Plant Creation Rules
1. **Slot Ownership**: Members can only create plants in their allocated slots
2. **Slot Availability**: Plant slots must be in "allocated" status
3. **Plant Type Validation**: Selected plant type must be available in catalog
4. **Single Plant Per Slot**: Each slot can only contain one active plant
5. **Membership Verification**: Members must have active membership

### Status Transition Rules
1. **Sequential Progression**: Plants must follow natural lifecycle progression
2. **No Backward Transitions**: Cannot move backward in lifecycle (except admin override)
3. **Death Handling**: Plants can die at any stage and release their slot
4. **Harvest Validation**: Plants must be in flowering stage and past expected harvest date

### Care & Health Rules
1. **Ownership Verification**: Only plant owners can record care activities
2. **Health Improvement**: Different care types provide varying health benefits
3. **Measurement Validation**: Care measurements must be within realistic ranges
4. **Active Plant Only**: Cannot record care for harvested or dead plants

### Harvest Rules
1. **Readiness Validation**: Plants must be in flowering stage and ready for harvest
2. **Weight Requirements**: Harvest weight must be positive and realistic
3. **Quality Rating**: Quality must be rated 1-10
4. **Slot Release**: Successful harvest automatically releases the plant slot

## API Documentation

### Get My Plants

**Endpoint:** `GET /plants/v1/my-plants`  
**Permission:** `plant_view`

Returns all plants belonging to the authenticated member with filtering and pagination support.

**Query Parameters:**
- `status` (optional): Filter by plant status (seedling, vegetative, flowering, harvested, dead)
- `strain` (optional): Filter by plant strain (partial match)
- `health_min` (optional): Filter by minimum health rating (1-10)
- `health_max` (optional): Filter by maximum health rating (1-10)
- `ready_for_harvest` (optional): Filter plants ready for harvest (true/false)
- `page` (optional): Page number for pagination (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response:**
```json
{
  "plants": [
    {
      "id": "plant_id",
      "name": "Purple Haze #1",
      "status": "flowering",
      "strain": "Purple Haze",
      "health": 8,
      "planted_date": "2024-01-01T10:00:00Z",
      "expected_harvest": "2024-04-01T10:00:00Z",
      "updated_at": "2024-03-15T10:30:00Z"
    }
  ],
  "total": 1,
  "count": 1,
  "page": 1,
  "limit": 20
}
```

### Create Plant

**Endpoint:** `POST /plants/v1/create`  
**Permission:** `plant_create`

Create a new plant in an available plant slot.

**Request Body:**
```json
{
  "plant_slot_id": "slot_id",
  "plant_type_id": "type_id", 
  "name": "Purple Haze #1",
  "notes": "Premium strain cultivation"
}
```

**Response:**
```json
{
  "message": "Plant created successfully",
  "plant": {
    "id": "plant_id",
    "plant_type_id": "type_id",
    "plant_slot_id": "slot_id",
    "member_id": "member_id",
    "status": "seedling",
    "planted_date": "2024-01-01T10:00:00Z",
    "expected_harvest": "2024-04-01T10:00:00Z",
    "name": "Purple Haze #1",
    "health": 8,
    "strain": "Purple Haze",
    "notes": "Premium strain cultivation",
    "created_at": "2024-01-01T10:00:00Z"
  }
}
```

### Get Plant Details

**Endpoint:** `GET /plants/v1/{id}`  
**Permission:** `plant_view`

Get detailed information about a specific plant.

**Response:**
```json
{
  "plant": {
    "id": "plant_id",
    "name": "Purple Haze #1",
    "status": "flowering",
    "strain": "Purple Haze",
    "health": 9,
    "planted_date": "2024-01-01T10:00:00Z",
    "expected_harvest": "2024-04-01T10:00:00Z",
    "plant_type_id": "type_id",
    "plant_slot_id": "slot_id",
    "member_id": "member_id",
    "height": 45.5,
    "images": ["image_url_1", "image_url_2"],
    "notes": "Premium strain cultivation",
    "created_at": "2024-01-01T10:00:00Z"
  }
}
```

### Update Plant Status

**Endpoint:** `PUT /plants/v1/{id}/status`  
**Permission:** `plant_update`

Update the lifecycle status of a plant with validation.

**Request Body:**
```json
{
  "status": "flowering",
  "reason": "Plant has developed first flowers"
}
```

**Response:**
```json
{
  "message": "Plant status updated successfully",
  "plant": {
    "id": "plant_id",
    "status": "flowering",
    "updated_at": "2024-03-15T10:30:00Z"
  }
}
```

### Record Care Activities

**Endpoint:** `PUT /plants/v1/{id}/care`  
**Permission:** `plant_care`

Record care activities and measurements for a plant.

**Request Body:**
```json
{
  "care_type": "watering",
  "notes": "Regular watering with nutrients",
  "measurements": {
    "temperature": 24.5,
    "humidity": 65.0,
    "soil_ph": 6.5,
    "water_amount": 500.0
  },
  "products": ["nutrient_solution_a", "ph_adjuster"]
}
```

**Care Types:**
- `watering`: Regular watering (+1 health)
- `fertilizing`: Nutrient application (+2 health)
- `pruning`: Plant trimming and shaping
- `inspection`: Health check and monitoring
- `pest_control`: Pest management (+3 health)

**Response:**
```json
{
  "message": "Care record added successfully",
  "care_record": {
    "id": "care_record_id",
    "plant_id": "plant_id",
    "member_id": "member_id",
    "care_type": "watering",
    "care_date": "2024-03-15T10:30:00Z",
    "notes": "Regular watering with nutrients"
  }
}
```

### Upload Plant Images

**Endpoint:** `POST /plants/v1/{id}/images`  
**Permission:** `plant_update`

Upload growth documentation images for a plant.

**Request Body:**
```json
{
  "image_url": "https://storage.example.com/plant_images/plant_id/image.jpg",
  "description": "Week 8 flowering stage"
}
```

**Response:**
```json
{
  "message": "Plant image uploaded successfully",
  "image_url": "https://storage.example.com/plant_images/plant_id/image.jpg"
}
```

### Harvest Plant

**Endpoint:** `POST /plants/v1/{id}/harvest`  
**Permission:** `plant_harvest`

Process plant harvest with validation and slot release.

**Request Body:**
```json
{
  "weight": 45.5,
  "quality": 8,
  "notes": "Excellent quality harvest",
  "processing_type": "self_process"
}
```

**Processing Types:**
- `self_process`: Member will process harvest themselves
- `sell_to_seedeg`: Sell harvest to facility

**Response:**
```json
{
  "message": "Plant harvested successfully",
  "harvest": {
    "id": "harvest_id",
    "plant_id": "plant_id",
    "weight": 45.5,
    "quality": 8,
    "harvest_date": "2024-04-01T10:00:00Z",
    "processing_type": "self_process"
  }
}
```

## Administrative Endpoints

### Get All Plants

**Endpoint:** `GET /plants/v1/admin/all`  
**Permission:** `plant_manage`

Get all plants in the system with advanced filtering options.

**Query Parameters:**
- All member endpoint parameters plus:
- `member_id` (optional): Filter by specific member
- `plant_slot_id` (optional): Filter by specific plant slot
- `plant_type_id` (optional): Filter by plant type

**Response:**
```json
{
  "plants": [
    {
      "id": "plant_id",
      "name": "Purple Haze #1",
      "status": "flowering", 
      "strain": "Purple Haze",
      "health": 8,
      "member_id": "member_id",
      "plant_slot_id": "slot_id",
      "planted_date": "2024-01-01T10:00:00Z",
      "expected_harvest": "2024-04-01T10:00:00Z"
    }
  ],
  "total": 1,
  "count": 1
}
```

### Get Plant Analytics

**Endpoint:** `GET /plants/v1/admin/analytics`  
**Permission:** `plant_manage`

Get comprehensive analytics about plant performance and statistics.

**Query Parameters:**
- `member_id` (optional): Analytics for specific member
- `time_range` (optional): week, month, quarter, year

**Response:**
```json
{
  "analytics": {
    "status_distribution": {
      "seedling": 15,
      "vegetative": 25,
      "flowering": 30,
      "harvested": 20,
      "dead": 5
    },
    "health_distribution": {
      "1-3": 5,
      "4-6": 15,
      "7-8": 45,
      "9-10": 30
    },
    "strain_popularity": [
      {
        "strain": "Purple Haze",
        "count": 25,
        "avg_health": 8.2
      }
    ],
    "growth_metrics": {
      "avg_cycle_length": 95,
      "success_rate": 0.85,
      "avg_yield": 42.3
    },
    "upcoming_harvests": [
      {
        "plant_id": "plant_id",
        "name": "Purple Haze #1",
        "expected_date": "2024-04-01T10:00:00Z",
        "days_remaining": 5
      }
    ]
  },
  "generated_at": "2024-03-15T10:30:00Z"
}
```

### Get Health Alerts

**Endpoint:** `GET /plants/v1/admin/health-alerts`  
**Permission:** `plant_manage`

Get plants requiring immediate attention based on health metrics.

**Response:**
```json
{
  "alerts": [
    {
      "plant_id": "plant_id",
      "member_id": "member_id",
      "alert_type": "critical_health",
      "severity": "high",
      "message": "Plant health critically low (3/10)",
      "created_at": "2024-03-15T10:30:00Z",
      "health_rating": 3
    },
    {
      "plant_id": "plant_id_2",
      "member_id": "member_id_2",
      "alert_type": "overdue_care",
      "severity": "medium",
      "message": "No care recorded in 10 days",
      "created_at": "2024-03-15T10:30:00Z",
      "days_overdue": 10
    }
  ]
}
```

### Force Status Update

**Endpoint:** `PUT /plants/v1/admin/{id}/force-status`  
**Permission:** `plant_manage`

Administrative override to force any status change.

**Request Body:**
```json
{
  "status": "dead",
  "reason": "Plant contamination - emergency removal",
  "admin_override": true
}
```

**Response:**
```json
{
  "message": "Plant status force updated successfully",
  "plant": {
    "id": "plant_id",
    "status": "dead",
    "updated_at": "2024-03-15T10:30:00Z"
  }
}
```

### Get Harvest Ready Plants

**Endpoint:** `GET /plants/v1/admin/harvest-ready`  
**Permission:** `plant_manage`

Get plants that are ready for harvest within specified timeframe.

**Query Parameters:**
- `days_ahead` (optional): Look-ahead period in days (default: 7)

**Response:**
```json
{
  "harvest_ready": [
    {
      "plant_id": "plant_id",
      "name": "Purple Haze #1",
      "member_id": "member_id",
      "strain": "Purple Haze",
      "status": "flowering",
      "expected_harvest": "2024-04-01T10:00:00Z",
      "days_remaining": 2,
      "health": 9
    }
  ],
  "total": 1
}
```

## Error Handling

The API returns appropriate HTTP status codes and error messages:

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `plant_slot_required` | Plant slot ID is required |
| 400 | `plant_care_record_invalid` | Invalid care record data |
| 400 | `plant_lifecycle_violation` | Invalid status transition |
| 403 | `plant_unauthorized_owner` | Not authorized to access this plant |
| 404 | `plant_not_found` | Plant not found |
| 409 | `plant_slot_occupied` | Plant slot already occupied |
| 409 | `plant_not_ready_for_harvest` | Plant not ready for harvest |
| 409 | `plant_health_critical` | Plant health too low for operation |
| 409 | `plant_type_not_available` | Selected plant type not available |

**Error Response Format:**
```json
{
  "error": "plant_unauthorized_owner",
  "message": "Not authorized to access this plant",
  "status": 403
}
```

## Integration Points

### Plant Slot Integration
- **Creation**: Plant creation automatically sets slot status to "occupied"
- **Harvest**: Plant harvest automatically releases slot (status → "available")
- **Death**: Plant death automatically releases slot
- **Transfer**: Occupied slots cannot be transferred between members

### Plant Type Integration
- **Harvest Schedule**: Expected harvest calculated from PlantType flowering time
- **Strain Information**: Plant strain inherited from PlantType
- **Availability**: PlantType must be available for plant creation

### Member Integration
- **Ownership**: Only slot owners can create plants
- **Access Control**: Plant owners control all plant operations
- **Membership**: Active membership required for plant operations

### Care Record Integration
- **Health Tracking**: Care activities automatically improve plant health
- **History**: Complete care history available for each plant
- **Analytics**: Care frequency and effectiveness tracking

## Rate Limiting

API endpoints are rate limited to ensure system stability:

- **Member Endpoints**: 100 requests per minute per user
- **Admin Endpoints**: 200 requests per minute per admin
- **Image Upload**: 10 uploads per hour per plant

## Changelog

### Version 1.0.0 (2024-03-15)
- Initial release with complete plant lifecycle management
- All 12 endpoints implemented and tested
- Full integration with plant slot and plant type systems
- Comprehensive analytics and health monitoring
- Production-ready with complete error handling 