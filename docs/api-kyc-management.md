# KYC (Know Your Customer) API Management

## Overview

The KYC API provides comprehensive identity verification capabilities for members of the platform. It enables secure document upload, administrative verification workflows, and automated email notifications while maintaining strict security and compliance standards.

## Features

- **Document Upload**: Secure upload and validation of identity documents
- **Multi-format Support**: Supports PDF, JPG, PNG, TIFF formats
- **Administrative Verification**: Complete workflow for admin review and approval/rejection
- **Email Notifications**: Automated notifications for all KYC status changes
- **Audit Trail**: Complete history tracking for compliance requirements
- **Security Controls**: File validation, size limits, and magic number checks
- **Tenant Isolation**: Multi-tenant security with cross-tenant data protection

## Authentication & Authorization

All KYC endpoints require authentication via Bearer token. Different endpoints require different permission levels:

- **Member Endpoints**: `user_view_self`, `user_update_self`
- **Admin Endpoints**: `kyc_view`, `kyc_verify`

## Member API Endpoints

### 1. Upload KYC Document

**Endpoint**: `POST /kyc/v1/documents/upload`

**Description**: Upload identity documents for verification

**Content-Type**: `multipart/form-data`

**Parameters**:
- `file` (file, required): Document file (PDF, JPG, PNG, TIFF, max 10MB)
- `document_type` (string, required): Type of document
  - Values: `passport`, `drivers_license`, `national_id`, `proof_of_address`
- `file_type` (string, required): Side/type of document
  - Values: `front`, `back`, `document`

**Response**:
```json
{
  "message": "Document uploaded successfully",
  "object_path": "member_id/passport/front/1234567890-passport.jpg"
}
```

**Security Features**:
- File type validation via magic number checking
- Maximum file size enforcement (10MB)
- Secure storage with unique file paths
- Tenant isolation for document access

### 2. Get KYC Status

**Endpoint**: `GET /kyc/v1/status`

**Description**: Retrieve current KYC status and document information

**Response**:
```json
{
  "kyc_status": "submitted",
  "can_submit": false,
  "has_documents": true,
  "documents_status": {
    "passport": {
      "has_front": true,
      "has_back": true,
      "uploaded_at": "2023-12-01T10:30:00Z",
      "is_complete": true
    },
    "proof_of_address": {
      "has_document": true,
      "uploaded_at": "2023-12-01T10:35:00Z",
      "is_complete": true
    }
  },
  "verification": {
    "submitted_at": "2023-12-01T11:00:00Z",
    "admin_notes": "Documents under review"
  },
  "history": [
    {
      "action": "submitted",
      "action_at": "2023-12-01T11:00:00Z",
      "notes": "Initial submission with passport and proof of address"
    }
  ]
}
```

**Status Values**:
- `not_started`: No documents uploaded
- `pending_kyc`: Documents uploaded but not submitted
- `submitted`: Submitted for admin review
- `in_review`: Currently being reviewed by admin
- `verified`: Approved and verified
- `rejected`: Rejected, requires resubmission

### 3. Submit for Verification

**Endpoint**: `POST /kyc/v1/submit`

**Description**: Submit uploaded documents for admin verification

**Request Body**:
```json
{
  "document_type": "passport",
  "has_all_documents": true,
  "confirm_accuracy": true
}
```

**Response**:
```json
{
  "message": "KYC submitted for verification successfully",
  "status": "submitted"
}
```

**Automated Actions**:
- Sends confirmation email to member
- Updates status to "submitted"
- Creates audit trail entry
- Notifies admin team (if configured)

### 4. Delete Document

**Endpoint**: `DELETE /kyc/v1/documents/{document_type}`

**Description**: Delete specific document type (only allowed before submission)

**Parameters**:
- `document_type` (path, required): Type of document to delete

**Response**:
```json
{
  "message": "Documents for passport deleted successfully"
}
```

**Restrictions**:
- Only allowed for statuses: `not_started`, `pending_kyc`, `rejected`
- Cannot delete documents once submitted or verified

## Admin API Endpoints

### 1. Get Pending Verifications

**Endpoint**: `GET /kyc/v1/admin/pending`

**Description**: Retrieve list of members with pending KYC verifications

**Query Parameters**:
- `page` (integer, optional): Page number (default: 1)
- `limit` (integer, optional): Items per page (default: 10, max: 100)

**Response**:
```json
{
  "members": [
    {
      "id": "654db9eca1f1b1bdbf3d4617",
      "email": "member@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "kyc_status": "submitted",
      "submitted_at": "2023-12-01T11:00:00Z",
      "document_type": "passport",
      "documents": {
        "passport": {
          "has_front": true,
          "has_back": true,
          "is_complete": true
        }
      }
    }
  ],
  "total_count": 25,
  "page": 1,
  "limit": 10
}
```

### 2. Get Member KYC Details

**Endpoint**: `GET /kyc/v1/admin/members/{member_id}`

**Description**: Get detailed KYC information for a specific member

**Parameters**:
- `member_id` (path, required): Member identifier

**Response**:
```json
{
  "id": "654db9eca1f1b1bdbf3d4617",
  "email": "member@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "kyc_status": {
    "kyc_status": "submitted",
    "has_documents": true,
    "documents_status": {
      // ... document details
    },
    "verification": {
      // ... verification details
    },
    "history": [
      // ... verification history
    ]
  }
}
```

### 3. Verify Member KYC

**Endpoint**: `POST /kyc/v1/admin/verify/{member_id}`

**Description**: Approve or reject member KYC verification

**Parameters**:
- `member_id` (path, required): Member identifier

**Request Body**:
```json
{
  "action": "approve",
  "reason": "All documents verified successfully",
  "notes": "Clear passport images, valid proof of address"
}
```

**Actions**:
- `approve`: Verify and approve KYC (status becomes "verified")
- `reject`: Reject KYC with reason (status becomes "rejected")

**Response**:
```json
{
  "message": "Member KYC approved successfully",
  "status": "verified"
}
```

**Automated Actions**:
- Sends approval/rejection email to member
- Updates verification status and timestamps
- Records admin action in audit trail
- Updates member's access permissions

### 4. Download Document

**Endpoint**: `GET /kyc/v1/admin/documents/{member_id}/{filename}`

**Description**: Generate secure download URL for KYC document

**Parameters**:
- `member_id` (path, required): Member identifier
- `filename` (path, required): Document filename

**Response**:
```json
{
  "download_url": "https://storage.example.com/presigned-url-with-auth",
  "expires_in": 3600
}
```

**Security Features**:
- Presigned URLs with 1-hour expiration
- Admin-only access with permission validation
- Tenant isolation enforcement
- Audit logging of document access

## KYC Status Workflow

```
not_started → pending_kyc → submitted → in_review → verified
                   ↓              ↓         ↓
                   └─────────── rejected ──┘
```

### Status Transitions

1. **not_started** → **pending_kyc**: First document uploaded
2. **pending_kyc** → **submitted**: Member submits for verification
3. **submitted** → **in_review**: Admin begins review process
4. **in_review** → **verified**: Admin approves verification
5. **in_review** → **rejected**: Admin rejects with reason
6. **rejected** → **pending_kyc**: Member uploads new documents

### Permissions by Status

| Status | Upload | Submit | Delete | Modify |
|--------|--------|--------|--------|--------|
| not_started | ✅ | ❌ | ✅ | ✅ |
| pending_kyc | ✅ | ✅ | ✅ | ✅ |
| submitted | ❌ | ❌ | ❌ | ❌ |
| in_review | ❌ | ❌ | ❌ | ❌ |
| verified | ❌ | ❌ | ❌ | ❌ |
| rejected | ✅ | ✅ | ✅ | ✅ |

## Document Types & Requirements

### Passport
- **Required Files**: Front page, back page (if applicable)
- **Formats**: PDF, JPG, PNG, TIFF
- **Requirements**: Clear, legible, not expired, matches member information

### Driver's License
- **Required Files**: Front side, back side
- **Formats**: PDF, JPG, PNG, TIFF
- **Requirements**: Valid, clear images, state/country issued

### National ID
- **Required Files**: Front side, back side
- **Formats**: PDF, JPG, PNG, TIFF
- **Requirements**: Government issued, current, clear text

### Proof of Address
- **Required Files**: Single document
- **Formats**: PDF, JPG, PNG, TIFF
- **Requirements**: Recent (within 3 months), shows full name and address
- **Examples**: Utility bill, bank statement, government correspondence

## Email Notifications

### Submission Confirmation
- **Trigger**: Member submits KYC for verification
- **Content**: Confirmation of receipt, expected timeline, next steps
- **Template**: Professional, informative

### Approval Notification
- **Trigger**: Admin approves KYC verification
- **Content**: Congratulations, account access details, feature unlocks
- **Template**: Celebratory, welcoming

### Rejection Notification
- **Trigger**: Admin rejects KYC verification
- **Content**: Reason for rejection, resubmission instructions, support contact
- **Template**: Helpful, constructive

## Error Handling

### Common Error Codes

- **400 Bad Request**: Invalid file format, missing parameters, validation errors
- **401 Unauthorized**: Invalid or expired authentication token
- **403 Forbidden**: Insufficient permissions, wrong tenant, status restrictions
- **404 Not Found**: Member not found, document not found
- **409 Conflict**: Status conflicts (already submitted, already verified)
- **413 Payload Too Large**: File exceeds 10MB limit
- **415 Unsupported Media Type**: Invalid file format
- **500 Internal Server Error**: Server processing errors

### Error Response Format

```json
{
  "error": "file_validation_failed",
  "message": "File content does not match expected file type",
  "status": 400,
  "timestamp": "2023-12-01T12:00:00Z"
}
```

## Security & Compliance

### File Security
- Magic number validation to prevent file type spoofing
- Size limits to prevent abuse
- Secure storage with encrypted transmission
- Access controls with presigned URLs

### Data Protection
- Tenant isolation for multi-tenant environments
- Admin permission requirements for sensitive operations
- Audit trails for all verification actions
- Secure document storage with access controls

### Compliance Features
- Complete verification history tracking
- Admin action logging with timestamps
- Document retention policies (configurable)
- GDPR compliance support for data deletion

## Integration Examples

### Frontend Upload Component

```javascript
const uploadDocument = async (file, documentType, fileType) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('document_type', documentType);
  formData.append('file_type', fileType);

  const response = await fetch('/kyc/v1/documents/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  return response.json();
};
```

### Admin Verification

```javascript
const verifyMember = async (memberId, action, reason, notes) => {
  const response = await fetch(`/kyc/v1/admin/verify/${memberId}`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${adminToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      action,
      reason,
      notes
    })
  });

  return response.json();
};
```

## Performance Considerations

### File Upload Optimization
- Progressive upload for large files
- Client-side validation before upload
- Compression recommendations for images
- Concurrent upload limits

### Database Optimization
- Indexed queries for pending verifications
- Pagination for large datasets
- Efficient status filtering
- Optimized document metadata storage

### Storage Optimization
- CDN integration for document delivery
- Automated lifecycle policies
- Compression for archived documents
- Regional storage for compliance

## Monitoring & Analytics

### Key Metrics
- Upload success/failure rates
- Verification processing times
- Document rejection reasons
- Member completion rates

### Audit Events
- Document uploads and deletions
- Status changes and transitions
- Admin verification actions
- Email notification deliveries

### Alerts
- Failed uploads beyond threshold
- Pending verifications exceeding SLA
- Storage capacity warnings
- Security violation attempts

---

## Support & Contact

For technical integration support or questions about the KYC API, please contact:
- **Technical Support**: support@example.com
- **API Documentation**: [Link to detailed API docs]
- **Status Page**: [Link to service status] 