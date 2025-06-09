# Harvest Management API Documentation

## Overview

The Harvest Management API provides comprehensive functionality for managing the cannabis harvest lifecycle, from initial harvest through processing, quality control, and final collection. The system implements a 7-stage processing workflow with admin quality verification and dual collection methods.

## Base URL
```
/harvest/v1
```

## Authentication
All endpoints require Bearer token authentication. Specific permission requirements are listed for each endpoint.

## Processing Workflow Stages

The harvest processing follows a 7-stage workflow:
1. **harvested** - Initial harvest completed
2. **initial_processing** - Processing started
3. **drying** - Drying stage
4. **curing** - Curing stage  
5. **quality_check** - Admin quality verification
6. **ready** - Ready for collection
7. **collected** - Final collection completed

## Collection Methods

- **pickup** - Member collects at facility
- **scheduled_delivery** - Delivery to specified address

---

# Member Endpoints (5)

## 1. Get Member's Harvests

**Endpoint**: `GET /harvest/v1/my-harvests`  
**Permission**: `harvest_view`  
**Description**: Retrieve all harvests belonging to the authenticated member with filtering and pagination support.

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `status` | string | No | Filter by harvest status |
| `strain` | string | No | Filter by strain name |
| `processing_stage` | string | No | Filter by processing stage |
| `date_from` | string | No | Filter from date (YYYY-MM-DD) |
| `date_to` | string | No | Filter to date (YYYY-MM-DD) |
| `page` | integer | No | Page number (default: 1) |
| `limit` | integer | No | Items per page (default: 10, max: 100) |

### Response

```json
{
  "harvests": [
    {
      "id": "64f123456789abcdef123456",
      "plant_id": "64f123456789abcdef123455",
      "member_id": "64f123456789abcdef123454",
      "harvest_date": "2024-06-10T10:30:00Z",
      "weight": 45.5,
      "quality": 8,
      "strain": "Purple Haze",
      "status": "processing",
      "processing_stage": "drying",
      "estimated_ready": "2024-06-20T10:30:00Z",
      "images": [
        "https://storage.example.com/harvest_images/harvest1.jpg"
      ],
      "notes": "High quality harvest",
      "created_at": "2024-06-10T10:30:00Z",
      "updated_at": "2024-06-15T14:20:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 10,
  "has_next": false
}
```

### Error Responses

| Code | Error | Description |
|------|--------|-------------|
| 401 | `unauthorized` | Missing or invalid authentication token |
| 400 | `invalid_parameters` | Invalid query parameters |

---

## 2. Get Harvest Details

**Endpoint**: `GET /harvest/v1/{id}`  
**Permission**: `harvest_view`  
**Description**: Retrieve detailed information about a specific harvest. Members can only access their own harvests.

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Harvest ID |

### Response

```json
{
  "harvest": {
    "id": "64f123456789abcdef123456",
    "plant_id": "64f123456789abcdef123455",
    "member_id": "64f123456789abcdef123454",
    "harvest_date": "2024-06-10T10:30:00Z",
    "weight": 45.5,
    "quality": 8,
    "strain": "Purple Haze",
    "status": "processing",
    "processing_stage": "drying",
    "processing_started": "2024-06-10T11:00:00Z",
    "drying_completed": null,
    "curing_completed": null,
    "estimated_ready": "2024-06-20T10:30:00Z",
    "processing_notes": "Standard processing protocol",
    "quality_checks": [
      {
        "checked_by": "admin_user_id",
        "checked_at": "2024-06-12T09:00:00Z",
        "visual_quality": 8,
        "moisture_content": 12.5,
        "density": 0.85,
        "approved": true,
        "notes": "Excellent quality, meets all standards"
      }
    ],
    "collection_method": null,
    "preferred_collection_date": null,
    "delivery_address": null,
    "collection_scheduled": null,
    "images": [
      "https://storage.example.com/harvest_images/harvest1.jpg",
      "https://storage.example.com/harvest_images/harvest1_process.jpg"
    ],
    "notes": "High quality harvest from slot 15",
    "nft_token_id": null,
    "nft_contract_address": null,
    "created_at": "2024-06-10T10:30:00Z",
    "updated_at": "2024-06-15T14:20:00Z"
  }
}
```

### Error Responses

| Code | Error | Description |
|------|--------|-------------|
| 401 | `unauthorized` | Missing or invalid authentication token |
| 403 | `forbidden` | Cannot access another member's harvest |
| 404 | `harvest_not_found` | Harvest does not exist |

---

## 3. Update Harvest Status

**Endpoint**: `PUT /harvest/v1/{id}/status`  
**Permission**: `harvest_update`  
**Description**: Update the processing status of a harvest. Members can update their own harvests with valid status transitions.

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Harvest ID |

### Request Body

```json
{
  "status": "processing",
  "processing_stage": "curing",
  "notes": "Moving to curing stage ahead of schedule"
}
```

### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | New harvest status |
| `processing_stage` | string | No | New processing stage |
| `notes` | string | No | Update notes |

### Valid Status Transitions

| From | To | Description |
|------|----|-----------| 
| `harvested` | `processing` | Begin processing |
| `processing` | `processing` | Stage progression within processing |

### Valid Processing Stage Transitions

| From | To |
|------|----|
| `harvested` | `initial_processing` |
| `initial_processing` | `drying` |
| `drying` | `curing` |
| `curing` | `quality_check` |

### Response

```json
{
  "message": "Harvest status updated successfully",
  "harvest": {
    "id": "64f123456789abcdef123456",
    "status": "processing",
    "processing_stage": "curing",
    "updated_at": "2024-06-15T16:30:00Z"
  }
}
```

### Error Responses

| Code | Error | Description |
|------|--------|-------------|
| 401 | `unauthorized` | Missing or invalid authentication token |
| 403 | `forbidden` | Cannot update another member's harvest |
| 404 | `harvest_not_found` | Harvest does not exist |
| 409 | `invalid_transition` | Invalid status or stage transition |
| 400 | `validation_error` | Invalid request data |

---

## 4. Upload Harvest Image

**Endpoint**: `POST /harvest/v1/{id}/images`  
**Permission**: `harvest_update`  
**Description**: Upload and attach images to a harvest record for documentation purposes.

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Harvest ID |

### Request Body

```json
{
  "image_url": "https://storage.example.com/harvest_images/new_image.jpg",
  "description": "Final harvest photo showing quality",
  "stage": "drying"
}
```

### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image_url` | string | Yes | URL of the uploaded image |
| `description` | string | No | Image description |
| `stage` | string | No | Processing stage when image was taken |

### Response

```json
{
  "message": "Harvest image uploaded successfully",
  "image": {
    "url": "https://storage.example.com/harvest_images/new_image.jpg",
    "description": "Final harvest photo showing quality",
    "stage": "drying",
    "uploaded_at": "2024-06-15T17:00:00Z"
  }
}
```

### Error Responses

| Code | Error | Description |
|------|--------|-------------|
| 401 | `unauthorized` | Missing or invalid authentication token |
| 403 | `forbidden` | Cannot update another member's harvest |
| 404 | `harvest_not_found` | Harvest does not exist |
| 400 | `invalid_image_url` | Invalid or inaccessible image URL |

---

## 5. Collect Harvest

**Endpoint**: `POST /harvest/v1/{id}/collect`  
**Permission**: `harvest_collect`  
**Description**: Schedule or complete harvest collection. Only available for harvests in "ready" status.

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Harvest ID |

### Request Body

**For Pickup Collection:**
```json
{
  "collection_method": "pickup",
  "special_notes": "Will pick up during business hours"
}
```

**For Scheduled Delivery:**
```json
{
  "collection_method": "scheduled_delivery",
  "delivery_address": "123 Main Street, City, State 12345",
  "preferred_date": "2024-06-25T14:00:00Z",
  "special_notes": "Weekday delivery preferred"
}
```

### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `collection_method` | string | Yes | `pickup` or `scheduled_delivery` |
| `delivery_address` | string | Conditional | Required for scheduled delivery |
| `preferred_date` | string | No | ISO 8601 datetime for delivery preference |
| `special_notes` | string | No | Additional collection instructions |

### Response

```json
{
  "message": "Harvest collection scheduled successfully",
  "collection": {
    "method": "scheduled_delivery",
    "scheduled_date": "2024-06-25T14:00:00Z",
    "delivery_address": "123 Main Street, City, State 12345",
    "status": "scheduled",
    "tracking_id": "HC240615001"
  }
}
```

### Error Responses

| Code | Error | Description |
|------|--------|-------------|
| 401 | `unauthorized` | Missing or invalid authentication token |
| 403 | `forbidden` | Cannot collect another member's harvest |
| 404 | `harvest_not_found` | Harvest does not exist |
| 409 | `harvest_not_ready` | Harvest not in ready status |
| 400 | `invalid_collection_data` | Missing or invalid collection information |

---

# Admin Endpoints (5)

## 6. Get All Harvests (Admin)

**Endpoint**: `GET /harvest/v1/admin/all`  
**Permission**: `harvest_manage`  
**Description**: Retrieve all harvests across all members with comprehensive filtering and analytics data.

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `member_id` | string | No | Filter by specific member |
| `status` | string | No | Filter by harvest status |
| `processing_stage` | string | No | Filter by processing stage |
| `quality_min` | integer | No | Minimum quality rating filter |
| `quality_max` | integer | No | Maximum quality rating filter |
| `date_from` | string | No | Filter from date (YYYY-MM-DD) |
| `date_to` | string | No | Filter to date (YYYY-MM-DD) |
| `page` | integer | No | Page number (default: 1) |
| `limit` | integer | No | Items per page (default: 20, max: 100) |
| `sort_by` | string | No | Sort field (`harvest_date`, `quality`, `weight`) |
| `sort_order` | string | No | Sort direction (`asc`, `desc`) |

### Response

```json
{
  "harvests": [
    {
      "id": "64f123456789abcdef123456",
      "plant_id": "64f123456789abcdef123455",
      "member_id": "64f123456789abcdef123454",
      "member_name": "John Doe",
      "harvest_date": "2024-06-10T10:30:00Z",
      "weight": 45.5,
      "quality": 8,
      "strain": "Purple Haze",
      "status": "processing",
      "processing_stage": "drying",
      "processing_days": 5,
      "estimated_ready": "2024-06-20T10:30:00Z",
      "quality_approved": true,
      "collection_scheduled": false,
      "created_at": "2024-06-10T10:30:00Z",
      "updated_at": "2024-06-15T14:20:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 20,
  "has_next": true,
  "summary": {
    "total_weight": 2875.5,
    "average_quality": 7.8,
    "status_breakdown": {
      "processing": 45,
      "ready": 12,
      "collected": 93
    },
    "stage_breakdown": {
      "drying": 20,
      "curing": 15,
      "quality_check": 10
    }
  }
}
```

---

## 7. Get Processing Harvests (Admin)

**Endpoint**: `GET /harvest/v1/admin/processing`  
**Permission**: `harvest_manage`  
**Description**: Get harvests currently in processing workflow with stage-specific information and timing.

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `stage` | string | No | Filter by processing stage |
| `overdue` | boolean | No | Show overdue items only |
| `days_in_stage` | integer | No | Filter by days in current stage |

### Response

```json
{
  "processing_harvests": [
    {
      "id": "64f123456789abcdef123456",
      "member_name": "John Doe",
      "strain": "Purple Haze",
      "weight": 45.5,
      "processing_stage": "drying",
      "days_in_stage": 5,
      "processing_started": "2024-06-10T11:00:00Z",
      "stage_started": "2024-06-12T09:00:00Z",
      "estimated_completion": "2024-06-18T09:00:00Z",
      "is_overdue": false,
      "next_action": "Check moisture levels",
      "quality_checked": false
    }
  ],
  "stage_summary": {
    "initial_processing": 5,
    "drying": 20,
    "curing": 15,
    "quality_check": 8
  },
  "overdue_count": 3,
  "ready_for_quality_check": 8
}
```

---

## 8. Get Harvest Analytics (Admin)

**Endpoint**: `GET /harvest/v1/admin/analytics`  
**Permission**: `harvest_manage`  
**Description**: Comprehensive harvest analytics and reporting data.

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `time_range` | string | No | `week`, `month`, `quarter`, `year` (default: month) |
| `member_id` | string | No | Analytics for specific member |
| `strain` | string | No | Analytics for specific strain |

### Response

```json
{
  "analytics": {
    "yield_analytics": {
      "total_harvests": 45,
      "total_weight": 1875.5,
      "average_weight": 41.7,
      "weight_trend": "+12.5%",
      "best_performing_strain": "Purple Haze",
      "yield_by_month": [
        {
          "month": "2024-06",
          "harvests": 15,
          "weight": 625.5
        }
      ]
    },
    "quality_analytics": {
      "average_quality": 7.8,
      "quality_trend": "+0.5",
      "quality_distribution": {
        "1-3": 2,
        "4-6": 8,
        "7-8": 25,
        "9-10": 10
      },
      "quality_by_strain": [
        {
          "strain": "Purple Haze",
          "average_quality": 8.2,
          "count": 12
        }
      ]
    },
    "processing_analytics": {
      "average_processing_time": 14.5,
      "processing_efficiency": 92.5,
      "stage_durations": {
        "initial_processing": 1.5,
        "drying": 6.0,
        "curing": 7.0
      },
      "quality_check_approval_rate": 95.5
    },
    "collection_analytics": {
      "total_collected": 38,
      "collection_methods": {
        "pickup": 28,
        "scheduled_delivery": 10
      },
      "average_collection_time": 2.3,
      "pending_collections": 7
    }
  },
  "time_range": "month",
  "generated_at": "2024-06-15T18:00:00Z"
}
```

---

## 9. Quality Check (Admin)

**Endpoint**: `POST /harvest/v1/admin/{id}/quality-check`  
**Permission**: `harvest_manage`  
**Description**: Record quality verification for a harvest in quality_check stage.

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Harvest ID |

### Request Body

```json
{
  "visual_quality": 8,
  "moisture_content": 12.5,
  "density": 0.85,
  "approved": true,
  "notes": "Excellent quality, meets all standards"
}
```

### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `visual_quality` | integer | Yes | Visual quality rating (1-10) |
| `moisture_content` | number | No | Moisture content percentage |
| `density` | number | No | Density measurement |
| `approved` | boolean | Yes | Quality approval status |
| `notes` | string | No | Quality check notes |

### Response

```json
{
  "message": "Quality check recorded successfully",
  "quality_check": {
    "checked_by": "admin_user_id",
    "checked_at": "2024-06-15T18:30:00Z",
    "visual_quality": 8,
    "moisture_content": 12.5,
    "density": 0.85,
    "approved": true,
    "notes": "Excellent quality, meets all standards"
  },
  "next_stage": "ready"
}
```

### Error Responses

| Code | Error | Description |
|------|--------|-------------|
| 401 | `unauthorized` | Missing or invalid authentication token |
| 403 | `forbidden` | Insufficient admin permissions |
| 404 | `harvest_not_found` | Harvest does not exist |
| 409 | `invalid_stage` | Harvest not in quality_check stage |
| 400 | `validation_error` | Invalid quality check data |

---

## 10. Force Status Update (Admin)

**Endpoint**: `PUT /harvest/v1/admin/{id}/force-status`  
**Permission**: `harvest_manage`  
**Description**: Force update harvest status and stage, bypassing normal validation rules. Use with caution.

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Harvest ID |

### Request Body

```json
{
  "status": "ready",
  "processing_stage": "ready",
  "force_reason": "Emergency quality approval override",
  "admin_notes": "Approved by facility manager due to urgent member request"
}
```

### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | Yes | New harvest status |
| `processing_stage` | string | No | New processing stage |
| `force_reason` | string | Yes | Reason for forcing status |
| `admin_notes` | string | No | Additional admin notes |

### Response

```json
{
  "message": "Harvest status force updated successfully",
  "harvest": {
    "id": "64f123456789abcdef123456",
    "status": "ready",
    "processing_stage": "ready",
    "force_updated_by": "admin_user_id",
    "force_updated_at": "2024-06-15T19:00:00Z",
    "force_reason": "Emergency quality approval override"
  },
  "warning": "Status was force updated, bypassing normal validation"
}
```

### Error Responses

| Code | Error | Description |
|------|--------|-------------|
| 401 | `unauthorized` | Missing or invalid authentication token |
| 403 | `forbidden` | Insufficient admin permissions |
| 404 | `harvest_not_found` | Harvest does not exist |
| 400 | `validation_error` | Invalid status or missing force reason |

---

# Error Codes

## Harvest-Specific Error Codes

| Code | Description | HTTP Status |
|------|-------------|------------|
| `harvest_not_found` | Harvest does not exist | 404 |
| `harvest_not_ready` | Harvest not ready for collection | 409 |
| `invalid_transition` | Invalid status or stage transition | 409 |
| `invalid_stage` | Operation not valid for current stage | 409 |
| `invalid_collection_data` | Invalid collection information | 400 |
| `invalid_image_url` | Invalid or inaccessible image URL | 400 |
| `quality_check_required` | Quality check must be completed first | 409 |
| `collection_already_scheduled` | Collection already scheduled | 409 |

## General Error Response Format

```json
{
  "error": "harvest_not_found",
  "error_description": "The specified harvest does not exist or you don't have permission to access it",
  "details": {
    "harvest_id": "64f123456789abcdef123456",
    "member_id": "64f123456789abcdef123454"
  }
}
```

---

# Rate Limits

- **Member endpoints**: 100 requests per minute
- **Admin endpoints**: 200 requests per minute  
- **Image upload**: 10 requests per minute

# Data Models

## HarvestDomain

```json
{
  "id": "string",
  "plant_id": "string",
  "member_id": "string", 
  "harvest_date": "string (ISO 8601)",
  "weight": "number",
  "quality": "integer (1-10)",
  "images": ["string"],
  "strain": "string",
  "status": "string",
  "processing_stage": "string",
  "processing_started": "string (ISO 8601)",
  "drying_completed": "string (ISO 8601)",
  "curing_completed": "string (ISO 8601)",
  "quality_checks": [
    {
      "checked_by": "string",
      "checked_at": "string (ISO 8601)",
      "visual_quality": "integer",
      "moisture_content": "number",
      "density": "number", 
      "approved": "boolean",
      "notes": "string"
    }
  ],
  "processing_notes": "string",
  "estimated_ready": "string (ISO 8601)",
  "collection_method": "string",
  "preferred_collection_date": "string (ISO 8601)",
  "delivery_address": "string",
  "collection_scheduled": "string (ISO 8601)",
  "notes": "string",
  "nft_token_id": "string",
  "nft_contract_address": "string",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
``` 