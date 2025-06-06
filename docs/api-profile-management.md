# Profile Management API Documentation

## Overview
The Profile Management API provides self-service capabilities for users to manage their profiles, privacy settings, profile pictures, and account deletion. This extends the existing User system with member-specific functionality.

## Authentication
All endpoints require Bearer token authentication using the `Authorization: Bearer <token>` header, except for public profile endpoints.

## Base URL
All endpoints are prefixed with `/profile/v1`

---

## Profile Management Endpoints

### Get User Profile
**Endpoint:** `GET /profile/v1`  
**Permission Required:** `PermissionUserViewSelf`  
**Description:** Retrieves the authenticated user's complete profile information.

#### Response
```json
{
  "user_id": "671db9eca1f1b1bdbf3d4618",
  "name": "John Doe",
  "phone": "1234567890",
  "email": "john@example.com",
  "username": "johndoe",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "email_verified": true,
  "email_verified_at": "2024-01-01T10:00:00Z",
  "profile_picture": "1706789123-avatar.jpg",
  "privacy_preferences": {
    "show_email": false,
    "show_phone": false,
    "show_name": true,
    "show_date_of_birth": false,
    "show_profile_picture": true,
    "is_public": false
  }
}
```

#### Error Responses
- `401 Unauthorized`: Invalid or expired token
- `404 Not Found`: User not found

---

### Update User Profile
**Endpoint:** `PUT /profile/v1`  
**Permission Required:** `PermissionUserUpdateSelf`  
**Description:** Updates the authenticated user's profile information.

#### Request Body
```json
{
  "name": "John Doe",
  "phone": "1234567890",
  "email": "john@example.com",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "profile_picture": "1706789123-new-avatar.jpg"
}
```

#### Field Validation
- `name`: Required, string
- `phone`: Required, lowercase string
- `email`: Required, lowercase, valid email format
- `date_of_birth`: Required, valid date
- `profile_picture`: Optional, string

#### Response
```json
{
  "user_id": "671db9eca1f1b1bdbf3d4618",
  "name": "John Doe",
  "phone": "1234567890",
  "email": "john@example.com",
  "username": "johndoe",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "email_verified": true,
  "email_verified_at": "2024-01-01T10:00:00Z",
  "profile_picture": "1706789123-new-avatar.jpg"
}
```

#### Error Responses
- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Invalid or expired token
- `404 Not Found`: User not found
- `409 Conflict`: Email or phone already in use

---

## Privacy Settings Endpoints

### Get Privacy Settings
**Endpoint:** `GET /profile/v1/privacy`  
**Permission Required:** `PermissionUserViewSelf`  
**Description:** Retrieves the user's privacy settings.

#### Response
```json
{
  "show_email": false,
  "show_phone": false,
  "show_name": true,
  "show_date_of_birth": false,
  "show_profile_picture": true,
  "is_public": false
}
```

#### Error Responses
- `401 Unauthorized`: Invalid or expired token
- `404 Not Found`: User not found

---

### Update Privacy Settings
**Endpoint:** `PUT /profile/v1/privacy`  
**Permission Required:** `PermissionUserPrivacySelf`  
**Description:** Updates the user's privacy settings.

#### Request Body
```json
{
  "show_email": false,
  "show_phone": false,
  "show_name": true,
  "show_date_of_birth": false,
  "show_profile_picture": true,
  "is_public": false
}
```

#### Field Validation
All fields are required boolean values:
- `show_email`: Controls email visibility in public profile
- `show_phone`: Controls phone visibility in public profile
- `show_name`: Controls name visibility in public profile
- `show_date_of_birth`: Controls birth date visibility in public profile
- `show_profile_picture`: Controls profile picture visibility in public profile
- `is_public`: Master switch for public profile accessibility

#### Response
```json
{
  "show_email": false,
  "show_phone": false,
  "show_name": true,
  "show_date_of_birth": false,
  "show_profile_picture": true,
  "is_public": false
}
```

#### Error Responses
- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Invalid or expired token
- `404 Not Found`: User not found

---

## Profile Picture Management

### Upload Profile Picture
**Endpoint:** `POST /profile/v1/picture`  
**Permission Required:** `PermissionUserUpdateSelf`  
**Description:** Uploads a new profile picture for the authenticated user.

#### Request
- **Content-Type:** `multipart/form-data`
- **Form Field:** `file` (image file)

#### Supported Formats
- JPEG (image/jpeg)
- PNG (image/png)
- GIF (image/gif)
- WebP (image/webp)

#### Response
```json
{
  "message": "Profile picture uploaded successfully",
  "profile_picture": "1706789123-avatar.jpg"
}
```

#### Error Responses
- `400 Bad Request`: Invalid file format or missing file
- `401 Unauthorized`: Invalid or expired token
- `500 Internal Server Error`: Upload failed

---

### Delete Profile Picture
**Endpoint:** `DELETE /profile/v1/picture`  
**Permission Required:** `PermissionUserUpdateSelf`  
**Description:** Removes the user's profile picture.

#### Response
```json
{
  "message": "Profile picture deleted successfully"
}
```

#### Error Responses
- `401 Unauthorized`: Invalid or expired token
- `404 Not Found`: User not found

---

## Public Profile Access

### Get Public Profile
**Endpoint:** `GET /profile/v1/public/:user_id`  
**Authentication:** Not required  
**Description:** Retrieves a user's public profile information based on their privacy settings.

#### Parameters
- `user_id`: The ID of the user whose public profile to retrieve

#### Response
The response content depends on the user's privacy settings. Only fields marked as public will be included:

```json
{
  "user_id": "671db9eca1f1b1bdbf3d4618",
  "username": "johndoe",
  "name": "John Doe",
  "email_verified": true,
  "profile_picture": "1706789123-avatar.jpg"
}
```

**Note:** Fields like `email`, `phone`, `date_of_birth` will only be included if the user has enabled their visibility in privacy settings.

#### Error Responses
- `404 Not Found`: User not found
- `403 Forbidden`: Profile is not public (user has `is_public: false`)

---

## Account Management

### Delete User Account
**Endpoint:** `DELETE /profile/v1/account`  
**Permission Required:** `PermissionUserDeleteSelf`  
**Description:** Permanently deletes the authenticated user's account and all associated data.

#### Response
```json
{
  "message": "Account deleted successfully"
}
```

#### Account Deletion Process
1. Validates user permissions (root users cannot delete their accounts)
2. Deletes profile picture from storage
3. Revokes all user authentication tokens
4. Performs soft delete by setting `data_status` to `disable`
5. Clears user cache
6. Logs the deletion action

#### Error Responses
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: Root users cannot delete their accounts
- `404 Not Found`: User not found
- `500 Internal Server Error`: Deletion failed

---

## Data Models

### UserProfileDto
```go
type UserProfileDto struct {
    ID              string           `json:"user_id"`
    Name            string           `json:"name,omitempty"`
    Phone           string           `json:"phone,omitempty"`
    Email           string           `json:"email,omitempty"`
    Username        string           `json:"username"`
    DateOfBirth     *time.Time       `json:"date_of_birth,omitempty"`
    EmailVerified   bool             `json:"email_verified"`
    EmailVerifiedAt *time.Time       `json:"email_verified_at,omitempty"`
    PrivacySettings *PrivacySettings `json:"privacy_preferences,omitempty"`
    ProfilePicture  string           `json:"profile_picture,omitempty"`
}
```

### PrivacySettings
```go
type PrivacySettings struct {
    ShowEmail          *bool `json:"show_email,omitempty"`
    ShowPhone          *bool `json:"show_phone,omitempty"`
    ShowDateOfBirth    *bool `json:"show_date_of_birth,omitempty"`
    ShowName           *bool `json:"show_name,omitempty"`
    ShowProfilePicture *bool `json:"show_profile_picture,omitempty"`
    IsPublic           *bool `json:"is_public,omitempty"`
}
```

### UserProfileUpdateData
```go
type UserProfileUpdateData struct {
    Name           string     `json:"name" validate:"required"`
    Phone          string     `json:"phone" validate:"required,lowercase"`
    Email          string     `json:"email" validate:"required,lowercase,email"`
    DateOfBirth    *time.Time `json:"date_of_birth" validate:"required"`
    ProfilePicture string     `json:"profile_picture,omitempty"`
}
```

### UserPrivacyUpdateData
```go
type UserPrivacyUpdateData struct {
    ShowEmail          bool `json:"show_email" validate:"required"`
    ShowPhone          bool `json:"show_phone" validate:"required"`
    ShowDateOfBirth    bool `json:"show_date_of_birth" validate:"required"`
    ShowName           bool `json:"show_name" validate:"required"`
    ShowProfilePicture bool `json:"show_profile_picture" validate:"required"`
    IsPublic           bool `json:"is_public" validate:"required"`
}
```

---

## Error Handling

All endpoints use standardized error responses from the `ecode` package:

```json
{
  "status": 400,
  "code": "validation_error",
  "message": "Invalid input data",
  "error": "Field 'email' must be a valid email address"
}
```

Common error codes:
- `400 Bad Request`: Input validation failures
- `401 Unauthorized`: Authentication required or invalid token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Data conflicts (e.g., duplicate email)
- `500 Internal Server Error`: Server-side errors

---

## Security Considerations

1. **Authentication**: All endpoints (except public profile) require valid Bearer tokens
2. **Authorization**: Permission-based access control using enum permissions
3. **Data Validation**: Comprehensive input validation using the validate package
4. **Privacy Controls**: Granular control over profile visibility
5. **File Upload Security**: 
   - Content-type validation for profile pictures
   - File size limitations through storage layer
   - Secure file naming with timestamps
6. **Account Deletion**: 
   - Soft deletion preserves audit trails
   - Automatic cleanup of associated resources
   - Prevention of root account deletion

---

## Rate Limiting

All endpoints are subject to the standard rate limiting configured in the middleware:
- Authenticated users: Higher rate limits based on username
- Unauthenticated requests: Rate limited by IP address

---

## Caching

User profile data is cached using the existing user cache strategy:
- Cache key: Based on user ID
- TTL: Configured system-wide
- Invalidation: Automatic on profile updates, privacy changes, and account operations 