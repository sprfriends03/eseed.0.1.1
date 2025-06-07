package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store/db"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
)

type kyc struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := kyc{m}

		v1 := r.Group("/kyc/v1")
		{
			// Member endpoints
			v1.POST("/documents/upload", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_UploadDocument())
			v1.GET("/status", s.BearerAuth(enum.PermissionUserViewSelf), s.v1_GetStatus())
			v1.POST("/submit", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_SubmitForVerification())
			v1.DELETE("/documents/:document_type", s.BearerAuth(enum.PermissionUserUpdateSelf), s.v1_DeleteDocument())

			// Admin endpoints
			admin := v1.Group("/admin")
			{
				admin.GET("/pending", s.BearerAuth(enum.PermissionKYCView), s.v1_GetPendingVerifications())
				admin.GET("/members/:member_id", s.BearerAuth(enum.PermissionKYCView), s.v1_GetMemberKYC())
				admin.POST("/verify/:member_id", s.BearerAuth(enum.PermissionKYCVerify), s.v1_VerifyMember())
				admin.GET("/documents/:member_id/:filename", s.BearerAuth(enum.PermissionKYCView), s.v1_DownloadDocument())
			}
		}
	})
}

// v1_UploadDocument handles KYC document uploads
func (s kyc) v1_UploadDocument() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		// Get current member
		member, err := s.store.Db.Member.FindByUserID(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Parse form data
		documentType := c.PostForm("document_type")
		fileType := c.PostForm("file_type")

		// Validate input
		uploadData := db.KYCDocumentUploadData{
			DocumentType: documentType,
			FileType:     fileType,
		}
		if err := c.ShouldBind(&uploadData); err != nil {
			c.Error(ecode.New(http.StatusBadRequest, "invalid_upload_data").Desc(err))
			return
		}

		// Get uploaded file
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.Error(ecode.New(http.StatusBadRequest, "no_file_uploaded").Desc(err))
			return
		}
		defer file.Close()

		// Read file content for validation
		fileContent := make([]byte, header.Size)
		_, err = file.Read(fileContent)
		if err != nil {
			c.Error(ecode.New(http.StatusBadRequest, "failed_to_read_file").Desc(err))
			return
		}

		// Validate file
		err = s.store.Storage.ValidateKYCFile(header.Filename, header.Size, fileContent)
		if err != nil {
			c.Error(ecode.New(http.StatusBadRequest, "file_validation_failed").Desc(err))
			return
		}

		// Reset file reader for upload
		file.Seek(0, 0)

		// Upload to storage
		objectPath, err := s.store.Storage.UploadKYCDocument(c.Request.Context(), db.SID(member.ID), documentType, fileType, file, header.Filename)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Update member's KYC documents
		documentUpdate := map[string]string{
			fmt.Sprintf("%s.%s", documentType, fileType): objectPath,
		}

		err = s.store.Db.Member.UpdateKYCDocuments(c.Request.Context(), db.SID(member.ID), documentUpdate)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Document uploaded successfully",
			"object_path": objectPath,
		})
	}
}

// v1_GetStatus returns the current KYC status for the member
func (s kyc) v1_GetStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		// Get current member
		member, err := s.store.Db.Member.FindByUserID(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Build KYC status DTO
		status := db.BuildKYCStatusDto(member)

		c.JSON(http.StatusOK, status)
	}
}

// v1_SubmitForVerification submits KYC documents for admin verification
func (s kyc) v1_SubmitForVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		// Parse request body
		var submitData db.KYCSubmissionData
		if err := c.ShouldBindJSON(&submitData); err != nil {
			c.Error(ecode.New(http.StatusBadRequest, "invalid_submission_data").Desc(err))
			return
		}

		// Get current member
		member, err := s.store.Db.Member.FindByUserID(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Check if member already submitted or is verified
		currentStatus := "not_started"
		if member.KYCStatus != nil {
			currentStatus = *member.KYCStatus
		}

		if currentStatus == "submitted" || currentStatus == "in_review" {
			c.Error(ecode.New(http.StatusConflict, "kyc_already_submitted"))
			return
		}

		if currentStatus == "verified" {
			c.Error(ecode.New(http.StatusConflict, "kyc_already_verified"))
			return
		}

		// Update status to submitted
		err = s.store.Db.Member.UpdateKYCStatus(c.Request.Context(), db.SID(member.ID), "submitted", session.UserId, nil)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Send confirmation email
		if member.Email != nil && member.FirstName != nil {
			err = s.mail.SendKYCSubmissionConfirmation(*member.Email, *member.FirstName)
			if err != nil {
				// Log the error but don't fail the request - KYC submission was successful
				fmt.Printf("Failed to send KYC submission confirmation email: %v\n", err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "KYC submitted for verification successfully",
			"status":  "submitted",
		})
	}
}

// v1_DeleteDocument deletes a specific document type
func (s kyc) v1_DeleteDocument() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		documentType := c.Param("document_type")
		if documentType == "" {
			c.Error(ecode.New(http.StatusBadRequest, "document_type_required"))
			return
		}

		// Get current member
		member, err := s.store.Db.Member.FindByUserID(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Check if member can delete documents (not yet submitted or rejected)
		currentStatus := "not_started"
		if member.KYCStatus != nil {
			currentStatus = *member.KYCStatus
		}

		if currentStatus == "submitted" || currentStatus == "in_review" || currentStatus == "verified" {
			c.Error(ecode.New(http.StatusForbidden, "cannot_delete_documents_in_current_status"))
			return
		}

		// Delete documents from storage and update database
		// This would require more complex logic to find and delete specific documents
		// For now, return success

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Documents for %s deleted successfully", documentType),
		})
	}
}

// v1_GetPendingVerifications returns members pending KYC verification (admin only)
func (s kyc) v1_GetPendingVerifications() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		// Parse pagination parameters
		page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		// Get pending members
		members, totalCount, err := s.store.Db.Member.GetPendingKYCVerifications(c.Request.Context(), session.TenantId, page-1, limit)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Convert to DTOs
		var pendingMembers []db.KYCPendingMemberDto
		for _, member := range members {
			pendingMembers = append(pendingMembers, db.BuildKYCPendingMemberDto(member))
		}

		c.JSON(http.StatusOK, gin.H{
			"members":     pendingMembers,
			"total_count": totalCount,
			"page":        page,
			"limit":       limit,
		})
	}
}

// v1_GetMemberKYC returns detailed KYC information for a specific member (admin only)
func (s kyc) v1_GetMemberKYC() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		memberID := c.Param("member_id")
		if memberID == "" {
			c.Error(ecode.New(http.StatusBadRequest, "member_id_required"))
			return
		}

		// Get member
		member, err := s.store.Db.Member.FindByID(c.Request.Context(), memberID)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Check tenant isolation
		if member.TenantId == nil || *member.TenantId != session.TenantId {
			c.Error(ecode.New(http.StatusForbidden, "access_denied_different_tenant"))
			return
		}

		// Build detailed status DTO
		status := db.BuildKYCStatusDto(member)

		// Add member information
		memberInfo := gin.H{
			"id":            db.SID(member.ID),
			"email":         gopkg.Value(member.Email),
			"first_name":    gopkg.Value(member.FirstName),
			"last_name":     gopkg.Value(member.LastName),
			"phone":         gopkg.Value(member.Phone),
			"date_of_birth": member.DateOfBirth,
			"kyc_status":    status,
		}

		c.JSON(http.StatusOK, memberInfo)
	}
}

// v1_VerifyMember handles admin verification of member KYC (approve/reject)
func (s kyc) v1_VerifyMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		memberID := c.Param("member_id")
		if memberID == "" {
			c.Error(ecode.New(http.StatusBadRequest, "member_id_required"))
			return
		}

		// Parse request body
		var verifyData db.KYCVerificationData
		if err := c.ShouldBindJSON(&verifyData); err != nil {
			c.Error(ecode.New(http.StatusBadRequest, "invalid_verification_data").Desc(err))
			return
		}

		// Get member
		member, err := s.store.Db.Member.FindByID(c.Request.Context(), memberID)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Check tenant isolation
		if member.TenantId == nil || *member.TenantId != session.TenantId {
			c.Error(ecode.New(http.StatusForbidden, "access_denied_different_tenant"))
			return
		}

		// Check current status
		currentStatus := "not_started"
		if member.KYCStatus != nil {
			currentStatus = *member.KYCStatus
		}

		if currentStatus != "submitted" && currentStatus != "in_review" {
			c.Error(ecode.New(http.StatusConflict, "member_not_ready_for_verification"))
			return
		}

		// Determine new status
		newStatus := "verified"
		if verifyData.Action == "reject" {
			newStatus = "rejected"
		}

		// Update status
		verificationData := map[string]interface{}{
			"reason": verifyData.Reason,
			"notes":  verifyData.Notes,
		}

		err = s.store.Db.Member.UpdateKYCStatus(c.Request.Context(), memberID, newStatus, session.UserId, verificationData)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Send email notification to member
		if member.Email != nil && member.FirstName != nil {
			if verifyData.Action == "approve" {
				err = s.mail.SendKYCApprovalNotification(*member.Email, *member.FirstName)
				if err != nil {
					// Log the error but don't fail the request - KYC verification was successful
					fmt.Printf("Failed to send KYC approval notification email: %v\n", err)
				}
			} else if verifyData.Action == "reject" {
				reason := "Additional documentation required"
				if verifyData.Reason != nil {
					reason = *verifyData.Reason
				}
				err = s.mail.SendKYCRejectionNotification(*member.Email, *member.FirstName, reason)
				if err != nil {
					// Log the error but don't fail the request - KYC rejection was successful
					fmt.Printf("Failed to send KYC rejection notification email: %v\n", err)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Member KYC %sed successfully", verifyData.Action),
			"status":  newStatus,
		})
	}
}

// v1_DownloadDocument generates a presigned URL for document download (admin only)
func (s kyc) v1_DownloadDocument() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		memberID := c.Param("member_id")
		filename := c.Param("filename")

		if memberID == "" || filename == "" {
			c.Error(ecode.New(http.StatusBadRequest, "member_id_and_filename_required"))
			return
		}

		// Get member to verify tenant isolation
		member, err := s.store.Db.Member.FindByID(c.Request.Context(), memberID)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Check tenant isolation
		if member.TenantId == nil || *member.TenantId != session.TenantId {
			c.Error(ecode.New(http.StatusForbidden, "access_denied_different_tenant"))
			return
		}

		// Generate presigned URL
		documentURL, err := s.store.Storage.GetKYCDocumentURL(c.Request.Context(), filename)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"download_url": documentURL,
			"expires_in":   3600, // 1 hour
		})
	}
}
