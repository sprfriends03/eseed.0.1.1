package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store/db"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
)

type membership struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := membership{m}

		v1 := r.Group("/membership/v1")
		{
			// Member endpoints
			v1.POST("/purchase", s.BearerAuth(enum.PermissionMembershipCreate), s.v1_purchaseMembership())
			v1.GET("/status", s.BearerAuth(enum.PermissionMembershipView), s.v1_getMembershipStatus())
			v1.POST("/renew", s.BearerAuth(enum.PermissionMembershipRenew), s.v1_renewMembership())
			v1.GET("/history", s.BearerAuth(enum.PermissionMembershipView), s.v1_getMembershipHistory())
			v1.DELETE("/:id", s.BearerAuth(enum.PermissionMembershipDelete), s.v1_cancelMembership())

			// Admin endpoints
			admin := v1.Group("/admin")
			{
				admin.GET("/pending", s.BearerAuth(enum.PermissionMembershipManage), s.v1_getPendingMemberships())
				admin.GET("/expiring", s.BearerAuth(enum.PermissionMembershipManage), s.v1_getExpiringMemberships())
				admin.PUT("/:id/status", s.BearerAuth(enum.PermissionMembershipManage), s.v1_updateMembershipStatus())
				admin.GET("/analytics", s.BearerAuth(enum.PermissionMembershipManage), s.v1_getMembershipAnalytics())
			}
		}
	})
}

// Membership request/response structures
type PurchaseMembershipRequest struct {
	MembershipType string `json:"membership_type" binding:"required" validate:"oneof=basic premium vip"`
	AutoRenew      *bool  `json:"auto_renew" validate:"omitempty"`
}

type RenewMembershipRequest struct {
	MembershipType string `json:"membership_type" binding:"required" validate:"oneof=basic premium vip"`
}

type UpdateMembershipStatusRequest struct {
	Status string  `json:"status" binding:"required" validate:"oneof=pending_payment active expired canceled suspended"`
	Reason *string `json:"reason" validate:"omitempty"`
}

// Membership pricing configuration
var membershipPricing = map[string]struct {
	Price          float64
	Duration       time.Duration
	SlotAllocation int
}{
	"basic": {
		Price:          29.99,
		Duration:       365 * 24 * time.Hour, // 1 year
		SlotAllocation: 2,
	},
	"premium": {
		Price:          99.99,
		Duration:       365 * 24 * time.Hour, // 1 year
		SlotAllocation: 5,
	},
	"vip": {
		Price:          199.99,
		Duration:       365 * 24 * time.Hour, // 1 year
		SlotAllocation: 10,
	},
}

// v1_purchaseMembership handles POST /membership/v1/purchase
func (s membership) v1_purchaseMembership() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		var req PurchaseMembershipRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Verify KYC is approved
		if member.KYCStatus == nil || *member.KYCStatus != "approved" {
			c.Error(ecode.KYCVerificationRequired.Desc(fmt.Errorf("KYC verification required to purchase membership")))
			return
		}

		// Check if member already has an active membership
		activeMembership, err := s.store.Db.Membership.FindActiveByMemberID(ctx, db.SID(member.ID))
		if err == nil && activeMembership != nil {
			c.Error(ecode.MembershipAlreadyActive.Desc(fmt.Errorf("Member already has an active membership")))
			return
		}

		// Get membership configuration
		config, exists := membershipPricing[req.MembershipType]
		if !exists {
			c.Error(ecode.InvalidMembershipType.Desc(fmt.Errorf("Invalid membership type: %s", req.MembershipType)))
			return
		}

		// Create membership record
		startDate := time.Now()
		expirationDate := startDate.Add(config.Duration)
		autoRenew := false
		if req.AutoRenew != nil {
			autoRenew = *req.AutoRenew
		}

		membership := &db.MembershipDomain{
			MemberID:       gopkg.Pointer(db.SID(member.ID)),
			MembershipType: gopkg.Pointer(req.MembershipType),
			StartDate:      gopkg.Pointer(startDate),
			ExpirationDate: gopkg.Pointer(expirationDate),
			Status:         gopkg.Pointer("active"), // Simplified for MVP
			AllocatedSlots: gopkg.Pointer(config.SlotAllocation),
			UsedSlots:      gopkg.Pointer(0),
			PaymentAmount:  gopkg.Pointer(config.Price),
			PaymentStatus:  gopkg.Pointer("paid"), // Simplified for MVP
			AutoRenew:      gopkg.Pointer(autoRenew),
			TenantId:       gopkg.Pointer(session.TenantId),
		}

		// Save membership
		savedMembership, err := s.store.Db.Membership.Save(ctx, membership)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Update member's current membership info
		err = s.store.Db.Member.UpdateMembershipType(ctx, db.SID(member.ID), req.MembershipType)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":     "Membership purchased successfully",
			"membership":  savedMembership,
			"member_type": req.MembershipType,
		})
	}
}

// v1_getMembershipStatus handles GET /membership/v1/status
func (s membership) v1_getMembershipStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get active membership
		membership, err := s.store.Db.Membership.FindActiveByMemberID(ctx, db.SID(member.ID))
		if err != nil {
			c.Error(ecode.MembershipNotFound.Desc(fmt.Errorf("No active membership found")))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"membership": membership,
		})
	}
}

// v1_renewMembership handles POST /membership/v1/renew
func (s membership) v1_renewMembership() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		var req RenewMembershipRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get current membership
		currentMembership, err := s.store.Db.Membership.FindActiveByMemberID(ctx, db.SID(member.ID))
		if err != nil {
			c.Error(ecode.MembershipNotFound.Desc(fmt.Errorf("No membership found to renew")))
			return
		}

		// Get membership configuration
		config, exists := membershipPricing[req.MembershipType]
		if !exists {
			c.Error(ecode.InvalidMembershipType.Desc(fmt.Errorf("Invalid membership type: %s", req.MembershipType)))
			return
		}

		// Extend expiration date from current expiration or now (whichever is later)
		now := time.Now()
		var newStartDate time.Time
		if currentMembership.ExpirationDate != nil && currentMembership.ExpirationDate.After(now) {
			newStartDate = *currentMembership.ExpirationDate
		} else {
			newStartDate = now
		}
		newExpirationDate := newStartDate.Add(config.Duration)

		// Update current membership - update expiration and type
		currentMembership.ExpirationDate = gopkg.Pointer(newExpirationDate)
		currentMembership.MembershipType = gopkg.Pointer(req.MembershipType)
		currentMembership.AllocatedSlots = gopkg.Pointer(config.SlotAllocation)
		currentMembership.PaymentAmount = gopkg.Pointer(config.Price)

		_, err = s.store.Db.Membership.Save(ctx, currentMembership)

		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "Membership renewed successfully",
			"new_expiration":  newExpirationDate,
			"membership_type": req.MembershipType,
			"allocated_slots": config.SlotAllocation,
		})
	}
}

// v1_getMembershipHistory handles GET /membership/v1/history
func (s membership) v1_getMembershipHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get all memberships for this member
		memberships, err := s.store.Db.Membership.FindByMemberID(ctx, db.SID(member.ID))
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"memberships": memberships,
			"total":       len(memberships),
		})
	}
}

// v1_cancelMembership handles DELETE /membership/v1/:id
func (s membership) v1_cancelMembership() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		membershipID := c.Param("id")

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get membership and verify ownership
		membership, err := s.store.Db.Membership.FindByID(ctx, membershipID)
		if err != nil {
			c.Error(ecode.MembershipNotFound.Desc(fmt.Errorf("Membership not found")))
			return
		}

		if membership.MemberID == nil || *membership.MemberID != db.SID(member.ID) {
			c.Error(ecode.Forbidden.Desc(fmt.Errorf("Not authorized to cancel this membership")))
			return
		}

		// Update membership status to canceled
		err = s.store.Db.Membership.UpdateStatus(ctx, membershipID, "canceled")
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Membership canceled successfully",
		})
	}
}

// Admin endpoints

// v1_getPendingMemberships handles GET /membership/v1/admin/pending
func (s membership) v1_getPendingMemberships() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

		// Simple implementation for MVP
		count, _ := s.store.Db.Membership.Count(ctx, db.M{
			"tenant_id": session.TenantId,
			"status":    "pending_payment",
		})

		c.JSON(http.StatusOK, gin.H{
			"memberships": []interface{}{}, // Simplified for MVP
			"total":       count,
			"page":        page,
			"limit":       limit,
		})
	}
}

// v1_getExpiringMemberships handles GET /membership/v1/admin/expiring
func (s membership) v1_getExpiringMemberships() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		daysThreshold, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

		// Get expiring memberships
		memberships, err := s.store.Db.Membership.FindExpiringSoon(ctx, daysThreshold, session.TenantId)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"memberships":    memberships,
			"total":          len(memberships),
			"days_threshold": daysThreshold,
		})
	}
}

// v1_updateMembershipStatus handles PUT /membership/v1/admin/:id/status
func (s membership) v1_updateMembershipStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		membershipID := c.Param("id")

		var req UpdateMembershipStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Update membership status
		err := s.store.Db.Membership.UpdateStatus(ctx, membershipID, req.Status)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Membership status updated successfully",
			"membership_id": membershipID,
			"new_status":    req.Status,
		})
	}
}

// v1_getMembershipAnalytics handles GET /membership/v1/admin/analytics
func (s membership) v1_getMembershipAnalytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Basic analytics for MVP
		totalMemberships, _ := s.store.Db.Membership.Count(ctx, db.M{
			"tenant_id": session.TenantId,
		})

		activeMemberships, _ := s.store.Db.Membership.Count(ctx, db.M{
			"tenant_id": session.TenantId,
			"status":    "active",
		})

		c.JSON(http.StatusOK, gin.H{
			"total_memberships":  totalMemberships,
			"active_memberships": activeMemberships,
		})
	}
}
