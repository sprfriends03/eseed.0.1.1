# Membership Management API

## Overview

The Membership Management API provides comprehensive functionality for handling member subscriptions, tier-based pricing, and administrative oversight of membership operations. It enables secure membership purchases, renewals, status tracking, and administrative management while maintaining strict security and compliance standards.

## Features

- **Tier-based Membership Plans**: Basic, Premium, and VIP tiers with different slot allocations
- **Purchase & Renewal System**: Complete membership lifecycle management
- **KYC Integration**: Requires approved KYC status for membership purchases
- **Administrative Controls**: Comprehensive admin management for membership oversight
- **Payment Processing**: Integration points ready for Stripe/PayPal implementation
- **Membership Analytics**: Basic analytics with expansion capabilities
- **Security Controls**: Permission-based access control and tenant isolation
- **Status Management**: Complete membership status lifecycle tracking

## Authentication & Authorization

All membership endpoints require authentication via Bearer token. Different endpoints require different permission levels:

- **Member Endpoints**: `membership_view`, `membership_create`, `membership_renew`, `membership_delete`
- **Admin Endpoints**: `membership_manage`

## Membership Tiers & Pricing

### Tier Structure

| Tier | Price | Duration | Plant Slots | Features |
|------|-------|----------|-------------|----------|
| Basic | $29.99 | 1 Year | 2 | Basic features |
| Premium | $99.99 | 1 Year | 5 | Enhanced features |
| VIP | $199.99 | 1 Year | 10 | Premium features |

### Membership Status Values

- `pending_payment`: Membership created but payment pending
- `active`: Active membership with full access
- `expired`: Membership has expired
- `canceled`: Membership canceled by member or admin
- `suspended`: Membership suspended by admin

## Member API Endpoints

### 1. Purchase Membership

**Endpoint**: `POST /membership/v1/purchase`

**Description**: Purchase a new membership plan

**Request Body**:
```json
{
  "membership_type": "premium",
  "auto_renew": false
}
```

**Parameters**:
- `membership_type` (string, required): Membership tier
  - Values: `basic`, `premium`, `vip`
- `auto_renew` (boolean, optional): Enable automatic renewal

**Response**:
```json
{
  "message": "Membership purchased successfully",
  "membership": {
    "id": "654db9eca1f1b1bdbf3d4617",
    "member_id": "654db9eca1f1b1bdbf3d4618",
    "membership_type": "premium",
    "start_date": "2023-12-01T10:30:00Z",
    "expiration_date": "2024-12-01T10:30:00Z",
    "status": "active",
    "allocated_slots": 5,
    "used_slots": 0,
    "payment_amount": 99.99,
    "payment_status": "paid",
    "auto_renew": false
  },
  "member_type": "premium"
}
```

**Requirements**:
- KYC status must be "approved"
- No existing active membership
- Valid membership type selection

**Error Codes**:
- `400`: Invalid membership type or request format
- `409`: Member already has active membership
- `403`: KYC verification required

### 2. Get Membership Status

**Endpoint**: `GET /membership/v1/status`

**Description**: Retrieve current membership status and details

**Response**:
```json
{
  "membership": {
    "id": "654db9eca1f1b1bdbf3d4617",
    "member_id": "654db9eca1f1b1bdbf3d4618",
    "membership_type": "premium",
    "start_date": "2023-12-01T10:30:00Z",
    "expiration_date": "2024-12-01T10:30:00Z",
    "status": "active",
    "allocated_slots": 5,
    "used_slots": 2,
    "payment_amount": 99.99,
    "payment_status": "paid",
    "auto_renew": false,
    "days_remaining": 325
  }
}
```

**Error Codes**:
- `404`: No active membership found
- `401`: Unauthorized access

### 3. Renew Membership

**Endpoint**: `POST /membership/v1/renew`

**Description**: Renew existing membership with optional tier upgrade

**Request Body**:
```json
{
  "membership_type": "vip"
}
```

**Parameters**:
- `membership_type` (string, required): New or same membership tier
  - Values: `basic`, `premium`, `vip`

**Response**:
```json
{
  "message": "Membership renewed successfully",
  "new_expiration": "2025-12-01T10:30:00Z",
  "membership_type": "vip",
  "allocated_slots": 10
}
```

**Business Logic**:
- Extends expiration date by 1 year from current expiration or today (whichever is later)
- Allows tier upgrades during renewal
- Updates slot allocation based on new tier

**Error Codes**:
- `404`: No membership found to renew
- `400`: Invalid membership type

### 4. Get Membership History

**Endpoint**: `GET /membership/v1/history`

**Description**: Retrieve complete membership transaction history

**Response**:
```json
{
  "memberships": [
    {
      "id": "654db9eca1f1b1bdbf3d4617",
      "membership_type": "premium",
      "start_date": "2023-12-01T10:30:00Z",
      "expiration_date": "2024-12-01T10:30:00Z",
      "status": "active",
      "payment_amount": 99.99,
      "payment_status": "paid"
    },
    {
      "id": "654db9eca1f1b1bdbf3d4618",
      "membership_type": "basic",
      "start_date": "2022-12-01T10:30:00Z",
      "expiration_date": "2023-11-30T10:30:00Z",
      "status": "expired",
      "payment_amount": 29.99,
      "payment_status": "paid"
    }
  ],
  "total": 2
}
```

### 5. Cancel Membership

**Endpoint**: `DELETE /membership/v1/{id}`

**Description**: Cancel specific membership (member ownership verified)

**Parameters**:
- `id` (path, required): Membership identifier

**Response**:
```json
{
  "message": "Membership canceled successfully"
}
```

**Security**:
- Verifies member ownership of the membership
- Updates status to "canceled"
- Maintains audit trail

**Error Codes**:
- `404`: Membership not found
- `403`: Not authorized to cancel this membership

## Admin API Endpoints

### 1. Get Pending Memberships

**Endpoint**: `GET /membership/v1/admin/pending`

**Description**: Retrieve list of memberships awaiting payment

**Query Parameters**:
- `page` (integer, optional): Page number (default: 1)
- `limit` (integer, optional): Items per page (default: 20, max: 100)

**Response**:
```json
{
  "memberships": [],
  "total": 0,
  "page": 1,
  "limit": 20
}
```

### 2. Get Expiring Memberships

**Endpoint**: `GET /membership/v1/admin/expiring`

**Description**: Retrieve list of memberships expiring soon

**Query Parameters**:
- `days` (integer, optional): Days threshold for expiration (default: 30)

**Response**:
```json
{
  "memberships": [
    {
      "id": "654db9eca1f1b1bdbf3d4617",
      "member_id": "654db9eca1f1b1bdbf3d4618",
      "member_email": "member@example.com",
      "membership_type": "premium",
      "expiration_date": "2023-12-15T10:30:00Z",
      "days_until_expiration": 14,
      "status": "active"
    }
  ],
  "total": 1,
  "days_threshold": 30
}
```

### 3. Update Membership Status

**Endpoint**: `PUT /membership/v1/admin/{id}/status`

**Description**: Admin override for membership status

**Parameters**:
- `id` (path, required): Membership identifier

**Request Body**:
```json
{
  "status": "suspended",
  "reason": "Payment dispute"
}
```

**Parameters**:
- `status` (string, required): New membership status
  - Values: `pending_payment`, `active`, `expired`, `canceled`, `suspended`
- `reason` (string, optional): Reason for status change

**Response**:
```json
{
  "message": "Membership status updated successfully",
  "membership_id": "654db9eca1f1b1bdbf3d4617",
  "new_status": "suspended"
}
```

### 4. Get Membership Analytics

**Endpoint**: `GET /membership/v1/admin/analytics`

**Description**: Basic membership analytics and statistics

**Response**:
```json
{
  "total_memberships": 150,
  "active_memberships": 125,
  "membership_by_tier": {
    "basic": 45,
    "premium": 60,
    "vip": 20
  },
  "revenue": {
    "total_revenue": 12499.75,
    "monthly_recurring": 2134.95
  },
  "expiring_soon": 12
}
```

## Membership Workflow

```
Purchase → Active → Renewal/Expiration
    ↓         ↓           ↓
  Payment   Usage    Extend/Upgrade
    ↓         ↓           ↓
  Active    Slots     New Term
    ↓         ↓           ↓
  Monitor   Track     Continue
```

### Status Transitions

1. **Purchase** → **Active**: Successful payment and KYC verification
2. **Active** → **Renewal**: Member renews before expiration
3. **Active** → **Expired**: Membership expires without renewal
4. **Active** → **Canceled**: Member or admin cancels membership
5. **Active** → **Suspended**: Admin suspends membership
6. **Expired** → **Active**: Renewal within grace period
7. **Suspended** → **Active**: Admin reactivates membership

### Business Rules

- **KYC Requirement**: All purchases require approved KYC status
- **Single Active Membership**: Only one active membership per member
- **Grace Period**: Expired memberships can be renewed (implementation ready)
- **Slot Management**: Track allocated vs. used plant slots
- **Tier Upgrades**: Allowed during renewal process
- **Payment Integration**: Ready for Stripe/PayPal implementation

## Integration Points

### Payment System Integration
```json
{
  "payment_provider": "stripe",
  "amount": 99.99,
  "currency": "USD",
  "member_id": "654db9eca1f1b1bdbf3d4618",
  "membership_type": "premium",
  "description": "Premium membership - 1 year"
}
```

### Email Notification Events
- **membership_purchased**: Welcome email with membership details
- **membership_renewed**: Renewal confirmation with new expiration
- **membership_expiring**: Reminder email before expiration
- **membership_canceled**: Cancellation confirmation
- **membership_suspended**: Suspension notification

### KYC System Integration
```json
{
  "kyc_check": {
    "required": true,
    "status": "approved",
    "verified_at": "2023-11-15T10:30:00Z"
  }
}
```

## Error Handling

### Custom Error Codes

- **membership_not_found**: No membership found for the operation
- **membership_already_active**: Member already has an active membership
- **invalid_membership_type**: Invalid tier selection
- **kyc_verification_required**: KYC approval required for purchase
- **payment_required**: Payment processing failed or pending
- **membership_expired**: Operation not allowed on expired membership

### Error Response Format

```json
{
  "error": "kyc_verification_required",
  "message": "KYC verification required to purchase membership",
  "status": 403,
  "timestamp": "2023-12-01T12:00:00Z"
}
```

## Security & Compliance

### Access Controls
- **Bearer Authentication**: Required for all endpoints
- **Permission-based Authorization**: Role-specific access levels
- **Tenant Isolation**: Multi-tenant data separation
- **Member Ownership Verification**: Ensures users can only access their own data

### Data Protection
- **Audit Trails**: Complete history of membership changes
- **Secure Payment Handling**: PCI DSS compliance ready
- **Privacy Controls**: Member data protection and anonymization options
- **GDPR Compliance**: Data retention and deletion capabilities

### Monitoring & Alerts
- **Failed Transactions**: Monitor payment failures
- **Suspicious Activities**: Unusual membership patterns
- **Expiration Alerts**: Proactive member retention
- **System Health**: API performance and availability

## Performance Considerations

### Database Optimization
- **Indexed Queries**: Optimized for member and status lookups
- **Pagination Support**: Efficient large dataset handling
- **Caching Strategy**: Redis integration for frequently accessed data
- **Connection Pooling**: Optimized database connections

### API Performance
- **Response Time**: < 200ms for status queries
- **Throughput**: Supports high concurrent user load
- **Rate Limiting**: Prevents abuse and ensures fair usage
- **Error Recovery**: Graceful handling of service disruptions

## Future Enhancements

### Payment Integration Expansion
- **Multiple Payment Methods**: Credit cards, PayPal, bank transfers
- **Subscription Management**: Automated recurring billing
- **Refund Processing**: Automated refund capabilities
- **Payment Analytics**: Detailed financial reporting

### Advanced Analytics
- **Churn Analysis**: Member retention insights
- **Revenue Forecasting**: Predictive analytics for business planning
- **Usage Patterns**: Member behavior analysis
- **A/B Testing**: Pricing and feature optimization

### Enhanced Features
- **Gift Memberships**: Allow members to purchase for others
- **Corporate Plans**: Business account management
- **Loyalty Programs**: Rewards for long-term members
- **Mobile App Integration**: Native mobile support

## API Integration Examples

### Frontend Purchase Flow

```javascript
const purchaseMembership = async (membershipType, autoRenew = false) => {
  const response = await fetch('/membership/v1/purchase', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      membership_type: membershipType,
      auto_renew: autoRenew
    })
  });

  return response.json();
};
```

### Admin Status Management

```javascript
const updateMembershipStatus = async (membershipId, status, reason) => {
  const response = await fetch(`/membership/v1/admin/${membershipId}/status`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${adminToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      status,
      reason
    })
  });

  return response.json();
};
```

### Membership Status Check

```javascript
const getMembershipStatus = async () => {
  const response = await fetch('/membership/v1/status', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });

  return response.json();
};
```

## Testing & Validation

### Test Coverage
- **Unit Tests**: Business logic validation
- **Integration Tests**: API endpoint functionality
- **Security Tests**: Authorization and authentication
- **Performance Tests**: Load and stress testing

### Validation Rules
- **Input Validation**: Request parameter verification
- **Business Rule Validation**: Membership tier and status rules
- **Security Validation**: Permission and ownership checks
- **Data Integrity**: Database constraint enforcement

## Support & Monitoring

### Observability
- **Logging**: Comprehensive audit trails
- **Metrics**: Business and technical KPIs
- **Alerting**: Proactive issue detection
- **Dashboards**: Real-time system health monitoring

### Support Information
- **API Status**: Real-time service availability
- **Documentation**: Comprehensive integration guides
- **Support Channels**: Technical assistance and troubleshooting
- **Community**: Developer resources and best practices

---

## Contact & Resources

For technical integration support or questions about the Membership Management API:
- **Technical Support**: support@example.com
- **API Documentation**: [Link to detailed API docs]
- **Status Page**: [Link to service status]
- **Developer Portal**: [Link to developer resources] 