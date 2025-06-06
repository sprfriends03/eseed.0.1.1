package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store/db"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
	"github.com/sirupsen/logrus"
)

type profile struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := profile{m}

		v1 := r.Group("/profile/v1")
		v1.GET("", s.BearerAuth(enum.PermissionUserViewSelf), s.v1_GetProfile())
		v1.PUT("", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_UpdateProfile())
		v1.GET("/privacy", s.BearerAuth(enum.PermissionUserViewSelf), s.v1_GetPrivacy())
		v1.PUT("/privacy", s.BearerAuth(enum.PermissionUserPrivacySelf), s.v1_UpdatePrivacy())

		// Profile picture management
		v1.POST("/picture", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_UploadProfilePicture())
		v1.DELETE("/picture", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_DeleteProfilePicture())

		// Account deletion
		v1.DELETE("/account", s.BearerAuth(enum.PermissionUserDeleteSelf), s.v1_DeleteAccount())

		// Public profile endpoint - no auth required
		v1.GET("/public/:user_id", s.NoAuth(), s.v1_GetPublicProfile())
	})
}

// @Tags Profile
// @Summary Get User Profile
// @Description Retrieves the authenticated user's profile information
// @Security BearerAuth
// @Success 200 {object} db.UserProfileDto "User profile information"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Router /profile/v1 [get]
func (s profile) v1_GetProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		user, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		// Return profile data with privacy settings not applied (self view)
		c.JSON(http.StatusOK, user.ProfileDto(false))
	}
}

// @Tags Profile
// @Summary Update User Profile
// @Description Updates the authenticated user's profile information
// @Security BearerAuth
// @Param body body db.UserProfileUpdateData true "Updated profile data"
// @Success 200 {object} db.UserProfileDto "Updated user profile"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input data"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Failure 409 {object} ecode.Error "Conflict - Email or phone already in use"
// @Router /profile/v1 [put]
func (s profile) v1_UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.UserProfileUpdateData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		user, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		// Update user domain with new profile data
		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: user.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update = data.Domain(update)

		// Save updated user
		updatedUser, err := s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.UserConflict.Stack(err))
			return
		}

		// Clear user cache
		s.store.DelUser(c.Request.Context(), db.SID(updatedUser.ID))

		// Log the update
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionUpdate, data, updatedUser, db.SID(updatedUser.ID))

		c.JSON(http.StatusOK, updatedUser.ProfileDto(false))
	}
}

// @Tags Profile
// @Summary Get User Privacy Settings
// @Description Retrieves the authenticated user's privacy settings
// @Security BearerAuth
// @Success 200 {object} db.PrivacySettings "User privacy settings"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Router /profile/v1/privacy [get]
func (s profile) v1_GetPrivacy() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		user, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		// If privacy settings don't exist yet, return default settings
		if user.PrivacyPreferences == nil {
			user.PrivacyPreferences = &db.PrivacySettings{
				ShowEmail:          gopkg.Pointer(false),
				ShowPhone:          gopkg.Pointer(false),
				ShowName:           gopkg.Pointer(false),
				ShowDateOfBirth:    gopkg.Pointer(false),
				ShowProfilePicture: gopkg.Pointer(false),
				IsPublic:           gopkg.Pointer(false),
			}
		}

		c.JSON(http.StatusOK, user.PrivacyPreferences)
	}
}

// @Tags Profile
// @Summary Update User Privacy Settings
// @Description Updates the authenticated user's privacy settings
// @Security BearerAuth
// @Param body body db.UserPrivacyUpdateData true "Updated privacy settings"
// @Success 200 {object} db.PrivacySettings "Updated privacy settings"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input data"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Router /profile/v1/privacy [put]
func (s profile) v1_UpdatePrivacy() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.UserPrivacyUpdateData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		user, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		// Update user domain with new privacy settings
		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: user.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update = data.Domain(update)

		// Save updated user
		updatedUser, err := s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			logrus.Errorf("Failed to update user privacy settings: %v", err)
			c.Error(ecode.InternalServerError.Stack(err))
			return
		}

		// Clear user cache
		s.store.DelUser(c.Request.Context(), db.SID(updatedUser.ID))

		// Log the update
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionUpdate, data, updatedUser, db.SID(updatedUser.ID))

		c.JSON(http.StatusOK, updatedUser.PrivacyPreferences)
	}
}

// @Tags Profile
// @Summary Get Public User Profile
// @Description Retrieves a user's public profile information based on their privacy settings
// @Param user_id path string true "User ID"
// @Success 200 {object} db.UserProfileDto "User public profile information"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Failure 403 {object} ecode.Error "Forbidden - Profile is not public"
// @Router /profile/v1/public/{user_id} [get]
func (s profile) v1_GetPublicProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		user, err := s.store.Db.User.FindOneById(c.Request.Context(), userID)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		// Check if profile is public
		if user.PrivacyPreferences == nil || user.PrivacyPreferences.IsPublic == nil || !*user.PrivacyPreferences.IsPublic {
			c.Error(ecode.New(http.StatusForbidden, "profile_not_public").Desc(fmt.Errorf("This profile is not public")))
			return
		}

		// Return profile with privacy settings applied
		c.JSON(http.StatusOK, user.ProfileDto(true))
	}
}

// @Tags Profile
// @Summary Upload Profile Picture
// @Description Uploads a profile picture for the authenticated user
// @Security BearerAuth
// @Param file formData file true "Profile picture image file"
// @Success 200 {object} object{message=string,profile_picture=string} "Profile picture uploaded successfully"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid file format or size"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 500 {object} ecode.Error "Internal Server Error - Upload failed"
// @Router /profile/v1/picture [post]
func (s profile) v1_UploadProfilePicture() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}
		defer file.Close()

		// Validate file type (only images)
		contentType := header.Header.Get("Content-Type")
		if !contains([]string{"image/jpeg", "image/png", "image/gif", "image/webp"}, contentType) {
			c.Error(ecode.New(http.StatusBadRequest, "invalid_file_type").Desc(fmt.Errorf("Only image files are allowed")))
			return
		}

		// Upload to storage
		filename, err := s.store.Storage.UploadProfileImage(c.Request.Context(), file, header.Filename)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Update user profile with new picture URL
		user, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: user.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update.ProfilePicture = gopkg.Pointer(filename)

		updatedUser, err := s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.InternalServerError.Stack(err))
			return
		}

		// Clear user cache
		s.store.DelUser(c.Request.Context(), db.SID(updatedUser.ID))

		// Log the update
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionUpdate, gin.H{"profile_picture": filename}, updatedUser, db.SID(updatedUser.ID))

		c.JSON(http.StatusOK, gin.H{
			"message":         "Profile picture uploaded successfully",
			"profile_picture": filename,
		})
	}
}

// @Tags Profile
// @Summary Delete Profile Picture
// @Description Removes the profile picture for the authenticated user
// @Security BearerAuth
// @Success 200 {object} object{message=string} "Profile picture deleted successfully"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Router /profile/v1/picture [delete]
func (s profile) v1_DeleteProfilePicture() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		user, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		// If user has a profile picture, delete it from storage
		if user.ProfilePicture != nil && *user.ProfilePicture != "" {
			// Delete from storage (best effort, don't fail if this fails)
			if err := s.store.Storage.DeleteProfileImage(c.Request.Context(), *user.ProfilePicture); err != nil {
				logrus.Warnf("Failed to delete profile picture from storage: %v", err)
			}
		}

		// Update user profile to remove picture URL
		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: user.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update.ProfilePicture = gopkg.Pointer("")

		updatedUser, err := s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.InternalServerError.Stack(err))
			return
		}

		// Clear user cache
		s.store.DelUser(c.Request.Context(), db.SID(updatedUser.ID))

		// Log the update
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionUpdate, gin.H{"profile_picture": ""}, updatedUser, db.SID(updatedUser.ID))

		c.JSON(http.StatusOK, gin.H{"message": "Profile picture deleted successfully"})
	}
}

// @Tags Profile
// @Summary Delete User Account
// @Description Permanently deletes the authenticated user's account and all associated data
// @Security BearerAuth
// @Success 200 {object} object{message=string} "Account deleted successfully"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Failure 500 {object} ecode.Error "Internal Server Error - Deletion failed"
// @Router /profile/v1/account [delete]
func (s profile) v1_DeleteAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		user, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		// Check if user is root - root users cannot delete their account
		if user.IsRoot != nil && *user.IsRoot {
			c.Error(ecode.New(http.StatusForbidden, "root_account_deletion_not_allowed").Desc(fmt.Errorf("Root users cannot delete their account")))
			return
		}

		// Delete profile picture from storage if exists
		if user.ProfilePicture != nil && *user.ProfilePicture != "" {
			if err := s.store.Storage.DeleteProfileImage(c.Request.Context(), *user.ProfilePicture); err != nil {
				logrus.Warnf("Failed to delete profile picture during account deletion: %v", err)
			}
		}

		// Revoke all user tokens
		if err := s.oauth.RevokeTokenByUser(c.Request.Context(), db.SID(user.ID)); err != nil {
			logrus.Warnf("Failed to revoke user tokens during account deletion: %v", err)
		}

		// Delete user from database (soft delete by setting data_status to disable)
		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: user.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update.DataStatus = gopkg.Pointer(enum.DataStatusDisable)

		deletedUser, err := s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.InternalServerError.Stack(err))
			return
		}

		// Clear user cache
		s.store.DelUser(c.Request.Context(), db.SID(deletedUser.ID))

		// Log the deletion
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionDelete, gin.H{"reason": "self_deletion"}, deletedUser, db.SID(deletedUser.ID))

		c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
	}
}

// Helper function to check if a slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
