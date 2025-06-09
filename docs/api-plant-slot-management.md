# Plant Slot Management API

## Overview

The Plant Slot Management API provides comprehensive functionality for managing cannabis cultivation slots within the facility. It enables members to request, manage, and track their plant slots while providing administrators with full oversight capabilities for allocation, maintenance, and analytics.

## Features

- **Slot Allocation System**: Request and automatically allocate available plant slots to members
- **Status Management**: Track slot statuses from available through occupied to maintenance
- **Transfer Functionality**: Transfer slots between verified members
- **Maintenance Tracking**: Report and track maintenance activities with detailed logs
- **Administrative Oversight**: Comprehensive admin controls for slot management and analytics
- **Capacity Management**: Intelligent allocation based on membership tiers and availability

## Endpoints Summary

### Member Endpoints
- `GET /plant-slots/v1/my-slots` - Get current member's slots
- `POST /plant-slots/v1/request` - Request new plant slots
- `GET /plant-slots/v1/{id}` - Get detailed slot information
- `PUT /plant-slots/v1/{id}/status` - Update slot status
- `POST /plant-slots/v1/{id}/maintenance` - Report maintenance needs
- `POST /plant-slots/v1/transfer` - Transfer slots to another member

### Administrative Endpoints
- `GET /plant-slots/v1/admin/all` - Get all slots with filtering
- `POST /plant-slots/v1/admin/assign` - Assign slots to members
- `GET /plant-slots/v1/admin/maintenance` - Get slots requiring maintenance
- `GET /plant-slots/v1/admin/analytics` - Get slot utilization analytics
- `PUT /plant-slots/v1/admin/{id}/force-status` - Force status change (admin override)

## Plant Slot Statuses

The plant slot system uses the following status lifecycle:

| Status | Description | Allowed Transitions |
|--------|-------------|-------------------|
| `available` | Slot is open and ready for allocation | → `allocated` |
| `allocated` | Slot assigned to member but not in use | → `occupied`, `available` |
| `occupied` | Slot actively contains plants | → `maintenance`, `available` |
| `maintenance` | Slot undergoing maintenance/cleaning | → `available`, `out_of_service` |
| `out_of_service` | Slot temporarily unavailable | → `maintenance`, `available` |

## Authentication & Permissions

All endpoints require bearer token authentication. The following permissions are required:

| Permission | Description | Endpoints |
|------------|-------------|-----------|
| `plant_slot_view` | View own plant slot information | GET my-slots, GET slot details |
| `plant_slot_create` | Request new plant slots | POST request |
| `plant_slot_update` | Update own slot status, report maintenance | PUT status, POST maintenance |
| `plant_slot_transfer` | Transfer slots to other members | POST transfer |
| `plant_slot_assign` | Admin: Assign slots to members | POST admin/assign |
| `plant_slot_manage` | Admin: Full slot management access | All admin endpoints |

## Business Rules

### Slot Allocation Rules
1. **Membership Requirement**: Members must have an active membership to request slots
2. **Single Allocation**: Members can only have one active slot allocation at a time
3. **Capacity Limits**: Allocation respects membership tier slot limits
4. **Availability Check**: Only available slots can be allocated

### Transfer Rules
1. **Ownership Verification**: Only slot owners can initiate transfers
2. **Recipient Validation**: Recipients must have active memberships
3. **Status Restrictions**: Occupied slots cannot be transferred
4. **Audit Trail**: All transfers are logged with reasons

### Maintenance Rules
1. **Owner Reporting**: Slot owners can report maintenance needs
2. **Automatic Scheduling**: Slots requiring cleaning after 30+ days are flagged
3. **Status Management**: Maintenance automatically updates slot status
4. **Admin Override**: Administrators can force any status change

## API Documentation

### Get My Plant Slots

**Endpoint:** `GET /plant-slots/v1/my-slots`  
**Permission:** `plant_slot_view`

Returns all plant slots assigned to the authenticated member.

**Response:**
```json
{
  "slots": [
    {
      "id": "slot_id",
      "slot_number": 1,
      "member_id": "member_id",
      "status": "allocated",
      "location": "greenhouse-1",
      "position": {
        "row": 1,
        "column": 5
      },
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

### Request Plant Slots

**Endpoint:** `POST /plant-slots/v1/request`  
**Permission:** `plant_slot_create`

Request allocation of new plant slots.

**Request Body:**
```json
{
  "quantity": 2,
  "preferred_location": "greenhouse-1"
}
```

**Response:**
```json
{
  "message": "Plant slots allocated successfully",
  "slots": [
    {
      "id": "slot_id",
      "slot_number": 1,
      "member_id": "member_id",
      "membership_id": "membership_id",
      "status": "allocated",
      "location": "greenhouse-1",
      "position": {
        "row": 1,
        "column": 5
      },
      "notes": "",
      "maintenance_log": [],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 2
}
```

### Get Slot Details

**Endpoint:** `GET /plant-slots/v1/{id}`  
**Permission:** `plant_slot_view`

Get detailed information about a specific plant slot.

**Response:**
```json
{
  "slot": {
    "id": "slot_id",
    "slot_number": 1,
    "member_id": "member_id",
    "membership_id": "membership_id",
    "status": "occupied",
    "location": "greenhouse-1",
    "position": {
      "row": 1,
      "column": 5
    },
    "notes": "Premium strain cultivation",
    "maintenance_log": [
      {
        "date": "2024-01-10T09:00:00Z",
        "description": "Deep cleaning and sterilization",
        "performed_by": "staff_member_id"
      }
    ],
    "last_clean_date": "2024-01-10T09:00:00Z",
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Update Slot Status

**Endpoint:** `PUT /plant-slots/v1/{id}/status`  
**Permission:** `plant_slot_update`

Update the status of a plant slot.

**Request Body:**
```json
{
  "status": "occupied",
  "reason": "Started new cultivation cycle"
}
```

**Response:**
```json
{
  "message": "Slot status updated successfully",
  "slot_id": "slot_id",
  "new_status": "occupied"
}
```

### Report Maintenance

**Endpoint:** `POST /plant-slots/v1/{id}/maintenance`  
**Permission:** `plant_slot_update`

Report maintenance needs for a plant slot.

**Request Body:**
```json
{
  "description": "Irrigation system needs adjustment",
  "priority": "normal"
}
```

**Response:**
```json
{
  "message": "Maintenance request recorded successfully",
  "slot_id": "slot_id"
}
```

### Transfer Slots

**Endpoint:** `POST /plant-slots/v1/transfer`  
**Permission:** `plant_slot_transfer`

Transfer plant slots to another verified member.

**Request Body:**
```json
{
  "to_member_id": "recipient_member_id",
  "slot_ids": ["slot_id_1", "slot_id_2"],
  "reason": "Member upgrade transfer"
}
```

**Response:**
```json
{
  "message": "Slots transferred successfully",
  "from_member_id": "source_member_id",
  "to_member_id": "recipient_member_id",
  "transferred_slots": 2,
  "reason": "Member upgrade transfer"
}
```

## Administrative Endpoints

### Get All Slots (Admin)

**Endpoint:** `GET /plant-slots/v1/admin/all`  
**Permission:** `plant_slot_manage`

Retrieve all plant slots with filtering options.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)
- `status` (optional): Filter by status
- `location` (optional): Filter by location

**Response:**
```json
{
  "slots": [],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

### Assign Slots (Admin)

**Endpoint:** `POST /plant-slots/v1/admin/assign`  
**Permission:** `plant_slot_assign`

Manually assign slots to members.

**Request Body:**
```json
{
  "member_id": "member_id",
  "membership_id": "membership_id",
  "slot_ids": ["slot_id_1", "slot_id_2"],
  "assigned_by": "admin_id"
}
```

### Get Maintenance Slots (Admin)

**Endpoint:** `GET /plant-slots/v1/admin/maintenance`  
**Permission:** `plant_slot_manage`

Get slots requiring maintenance attention.

**Query Parameters:**
- `days` (optional): Days threshold for maintenance check (default: 30)

### Get Slot Analytics (Admin)

**Endpoint:** `GET /plant-slots/v1/admin/analytics`  
**Permission:** `plant_slot_manage`

Get comprehensive slot utilization analytics.

**Response:**
```json
{
  "total_slots": 100,
  "allocated_slots": 60,
  "occupied_slots": 45,
  "maintenance_slots": 5,
  "available_slots": 30,
  "utilization_rate": 70.0
}
```

### Force Status Update (Admin)

**Endpoint:** `PUT /plant-slots/v1/admin/{id}/force-status`  
**Permission:** `plant_slot_manage`

Force status change without business rule validation.

**Request Body:**
```json
{
  "status": "out_of_service",
  "reason": "Equipment failure"
}
```

## Error Handling

The API uses standard HTTP status codes and returns structured error responses:

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `plant_slot_not_found` | 404 | Plant slot not found |
| `plant_slot_insufficient_slots` | 409 | Not enough available slots |
| `plant_slot_membership_required` | 403 | Active membership required |
| `plant_slot_already_allocated` | 409 | Member already has allocated slots |
| `plant_slot_occupied_cannot_transfer` | 409 | Cannot transfer occupied slots |
| `invalid_status_transition` | 400 | Invalid status transition attempted |

**Error Response Format:**
```json
{
  "error": {
    "code": "plant_slot_insufficient_slots",
    "message": "Not enough available slots to fulfill request"
  }
}
```

## Integration Points

### Membership System Integration
- Validates active membership before slot allocation
- Respects membership tier slot limits
- Automatically releases slots on membership expiry

### KYC Integration
- Requires verified KYC status for slot requests
- Validates identity for slot transfers

### Audit System Integration
- Logs all slot allocation and transfer activities
- Tracks maintenance activities and status changes
- Maintains comprehensive audit trail

## Usage Examples

### Member Workflow Example
```bash
# 1. Check current slots
curl -H "Authorization: Bearer <token>" \
  GET /plant-slots/v1/my-slots

# 2. Request new slots
curl -H "Authorization: Bearer <token>" \
  -d '{"quantity": 2, "preferred_location": "greenhouse-1"}' \
  POST /plant-slots/v1/request

# 3. Update slot status when planting
curl -H "Authorization: Bearer <token>" \
  -d '{"status": "occupied", "reason": "Started cultivation"}' \
  PUT /plant-slots/v1/{slot_id}/status

# 4. Report maintenance if needed
curl -H "Authorization: Bearer <token>" \
  -d '{"description": "Irrigation issue", "priority": "high"}' \
  POST /plant-slots/v1/{slot_id}/maintenance
```

### Admin Workflow Example
```bash
# 1. Check slot analytics
curl -H "Authorization: Bearer <admin_token>" \
  GET /plant-slots/v1/admin/analytics

# 2. Get maintenance queue
curl -H "Authorization: Bearer <admin_token>" \
  GET /plant-slots/v1/admin/maintenance?days=7

# 3. Force status update if needed
curl -H "Authorization: Bearer <admin_token>" \
  -d '{"status": "out_of_service", "reason": "Equipment repair"}' \
  PUT /plant-slots/v1/admin/{slot_id}/force-status
```

This API provides a complete solution for plant slot management within the cannabis cultivation facility, ensuring proper allocation, tracking, and maintenance of cultivation spaces while maintaining compliance and audit trails. 