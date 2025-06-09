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
)

type harvest struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := harvest{m}

		// Member endpoints
		v1 := r.Group("/harvest/v1")
		v1.GET("/my-harvests", s.BearerAuth(enum.PermissionHarvestView), s.v1_getMyHarvests())
		v1.GET("/:id", s.BearerAuth(enum.PermissionHarvestView), s.v1_getHarvestDetails())
		v1.PUT("/:id/status", s.BearerAuth(enum.PermissionHarvestUpdate), s.v1_updateHarvestStatus())
		v1.POST("/:id/images", s.BearerAuth(enum.PermissionHarvestUpdate), s.v1_uploadHarvestImage())
		v1.POST("/:id/collect", s.BearerAuth(enum.PermissionHarvestCollect), s.v1_collectHarvest())

		// Admin endpoints
		admin := v1.Group("/admin")
		admin.GET("/all", s.BearerAuth(enum.PermissionHarvestManage), s.v1_getAllHarvests())
		admin.GET("/processing", s.BearerAuth(enum.PermissionHarvestManage), s.v1_getProcessingHarvests())
		admin.GET("/analytics", s.BearerAuth(enum.PermissionHarvestManage), s.v1_getHarvestAnalytics())
		admin.POST("/:id/quality-check", s.BearerAuth(enum.PermissionHarvestManage), s.v1_qualityCheck())
		admin.PUT("/:id/force-status", s.BearerAuth(enum.PermissionHarvestManage), s.v1_forceStatusUpdate())
	})
}

// Request/Response structures following established patterns
type UpdateHarvestStatusRequest struct {
	Status          string  `json:"status" binding:"required" validate:"oneof=processing drying curing quality_check ready"`
	ProcessingStage *string `json:"processing_stage" validate:"omitempty,oneof=initial_processing drying curing quality_check ready"`
	Notes           *string `json:"notes" validate:"omitempty,max=500"`
}

type UploadHarvestImageRequest struct {
	ImageURL    string `json:"image_url" binding:"required"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type CollectHarvestRequest struct {
	CollectionMethod string     `json:"collection_method" binding:"required" validate:"oneof=pickup scheduled_delivery"`
	PreferredDate    *time.Time `json:"preferred_date" validate:"omitempty"`
	DeliveryAddress  *string    `json:"delivery_address" validate:"required_if=CollectionMethod scheduled_delivery"`
	Notes            *string    `json:"notes" validate:"omitempty,max=500"`
}

type QualityCheckRequest struct {
	VisualQuality int      `json:"visual_quality" binding:"required,gte=1,lte=10"`
	Moisture      *float64 `json:"moisture" validate:"omitempty,gte=0,lte=100"`
	Density       *float64 `json:"density" validate:"omitempty,gte=0"`
	Notes         *string  `json:"notes" validate:"omitempty,max=500"`
	Approved      bool     `json:"approved"`
}

type ForceStatusUpdateRequest struct {
	Status          string  `json:"status" binding:"required" validate:"oneof=processing drying curing quality_check ready collected"`
	ProcessingStage *string `json:"processing_stage" validate:"omitempty"`
	Reason          string  `json:"reason" binding:"required,max=255"`
}

// v1_getMyHarvests handles GET /harvest/v1/my-harvests
func (s harvest) v1_getMyHarvests() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Parse query parameters
		status := c.Query("status")
		strainFilter := c.Query("strain")
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")

		page, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit < 1 || limit > 100 {
			limit = 20
		}

		offset := (page - 1) * limit

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Build filter
		filter := db.M{
			"member_id": db.SID(member.ID),
			"tenant_id": session.TenantId,
		}

		if status != "" {
			filter["status"] = status
		}

		if strainFilter != "" {
			filter["strain"] = strainFilter
		}

		// Get harvests with pagination
		harvests, err := s.store.Db.Harvest.FindByMemberID(ctx, db.SID(member.ID), offset, limit)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Get total count
		totalCount, err := s.store.Db.Harvest.Count(ctx, filter)
		if err != nil {
			totalCount = int64(len(harvests))
		}

		c.JSON(http.StatusOK, gin.H{
			"harvests": harvests,
			"total":    len(harvests),
			"count":    totalCount,
			"page":     page,
			"limit":    limit,
		})
	}
}

// v1_getHarvestDetails handles GET /harvest/v1/:id
func (s harvest) v1_getHarvestDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		id := c.Param("id")

		// Validate ID format
		if len(id) != 24 {
			c.Error(ecode.BadRequest.Desc(fmt.Errorf("invalid harvest ID format")))
			return
		}

		// Get harvest
		harvestDomain, err := s.store.Db.Harvest.FindByID(ctx, id)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "harvest_not_found").Desc(err))
			return
		}

		// Get member info to verify ownership
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Verify ownership (members can only see their own harvests)
		if harvestDomain.MemberID == nil || *harvestDomain.MemberID != db.SID(member.ID) {
			c.Error(ecode.New(http.StatusForbidden, "harvest_not_owned").Desc(fmt.Errorf("Harvest not owned by member")))
			return
		}

		// Verify tenant
		if harvestDomain.TenantId == nil || *harvestDomain.TenantId != session.TenantId {
			c.Error(ecode.New(http.StatusForbidden, "harvest_unauthorized").Desc(fmt.Errorf("Harvest access denied")))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"harvest": harvestDomain,
		})
	}
}

// v1_updateHarvestStatus handles PUT /harvest/v1/:id/status
func (s harvest) v1_updateHarvestStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		id := c.Param("id")

		var req UpdateHarvestStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Validate ID format
		if len(id) != 24 {
			c.Error(ecode.BadRequest.Desc(fmt.Errorf("invalid harvest ID format")))
			return
		}

		// Get harvest and verify ownership
		harvestDomain, err := s.store.Db.Harvest.FindByID(ctx, id)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "harvest_not_found").Desc(err))
			return
		}

		// Get member info to verify ownership
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Verify ownership
		if harvestDomain.MemberID == nil || *harvestDomain.MemberID != db.SID(member.ID) {
			c.Error(ecode.New(http.StatusForbidden, "harvest_not_owned").Desc(fmt.Errorf("Harvest not owned by member")))
			return
		}

		// Members can only update their own harvests and only certain status transitions
		allowedStatuses := []string{"processing", "drying", "curing"}
		allowed := false
		for _, status := range allowedStatuses {
			if req.Status == status {
				allowed = true
				break
			}
		}

		if !allowed {
			c.Error(ecode.New(http.StatusForbidden, "status_update_not_allowed").Desc(fmt.Errorf("Members cannot set status to %s", req.Status)))
			return
		}

		// Update status
		if req.ProcessingStage != nil {
			err = s.store.Db.Harvest.UpdateProcessingStatus(ctx, id, *req.ProcessingStage, req.Notes)
		} else {
			err = s.store.Db.Harvest.UpdateStatus(ctx, id, req.Status)
		}

		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Harvest status updated successfully",
			"status":  req.Status,
		})
	}
}

// v1_uploadHarvestImage handles POST /harvest/v1/:id/images
func (s harvest) v1_uploadHarvestImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		id := c.Param("id")

		var req UploadHarvestImageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Validate ID format
		if len(id) != 24 {
			c.Error(ecode.BadRequest.Desc(fmt.Errorf("invalid harvest ID format")))
			return
		}

		// Get harvest and verify ownership
		harvestDomain, err := s.store.Db.Harvest.FindByID(ctx, id)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "harvest_not_found").Desc(err))
			return
		}

		// Get member info to verify ownership
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Verify ownership
		if harvestDomain.MemberID == nil || *harvestDomain.MemberID != db.SID(member.ID) {
			c.Error(ecode.New(http.StatusForbidden, "harvest_not_owned").Desc(fmt.Errorf("Harvest not owned by member")))
			return
		}

		// Add image
		err = s.store.Db.Harvest.AddImage(ctx, id, req.ImageURL)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Image uploaded successfully",
			"image_url":   req.ImageURL,
			"description": req.Description,
		})
	}
}

// v1_collectHarvest handles POST /harvest/v1/:id/collect
func (s harvest) v1_collectHarvest() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		id := c.Param("id")

		var req CollectHarvestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Validate ID format
		if len(id) != 24 {
			c.Error(ecode.BadRequest.Desc(fmt.Errorf("invalid harvest ID format")))
			return
		}

		// Get harvest and verify ownership
		harvestDomain, err := s.store.Db.Harvest.FindByID(ctx, id)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "harvest_not_found").Desc(err))
			return
		}

		// Get member info to verify ownership
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Verify ownership
		if harvestDomain.MemberID == nil || *harvestDomain.MemberID != db.SID(member.ID) {
			c.Error(ecode.New(http.StatusForbidden, "harvest_not_owned").Desc(fmt.Errorf("Harvest not owned by member")))
			return
		}

		// Verify harvest is ready for collection
		if harvestDomain.Status == nil || *harvestDomain.Status != "ready" {
			c.Error(ecode.New(http.StatusBadRequest, "harvest_not_ready").Desc(fmt.Errorf("Harvest is not ready for collection")))
			return
		}

		// Schedule collection or complete immediately for pickup
		if req.CollectionMethod == "pickup" {
			// Complete collection immediately for pickup
			err = s.store.Db.Harvest.CompleteCollection(ctx, id)
			if err != nil {
				c.Error(ecode.InternalServerError.Desc(err))
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":           "Harvest collected successfully",
				"collection_method": req.CollectionMethod,
				"collection_date":   time.Now(),
			})
		} else {
			// Schedule delivery
			err = s.store.Db.Harvest.ScheduleCollection(ctx, id, req.CollectionMethod, req.PreferredDate, req.DeliveryAddress)
			if err != nil {
				c.Error(ecode.InternalServerError.Desc(err))
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":           "Collection scheduled successfully",
				"collection_method": req.CollectionMethod,
				"preferred_date":    req.PreferredDate,
				"delivery_address":  req.DeliveryAddress,
			})
		}
	}
}

// v1_getAllHarvests handles GET /harvest/v1/admin/all
func (s harvest) v1_getAllHarvests() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Parse query parameters
		status := c.Query("status")
		strain := c.Query("strain")
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "50")

		page, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit < 1 || limit > 100 {
			limit = 50
		}

		// Build filter
		filter := db.M{
			"tenant_id": session.TenantId,
		}

		if status != "" {
			filter["status"] = status
		}

		if strain != "" {
			filter["strain"] = strain
		}

		// Get harvests
		// Note: This is a simplified version - you'd want to implement pagination properly
		var harvests []*db.HarvestDomain

		if status != "" {
			harvests, err = s.store.Db.Harvest.FindByStatus(ctx, status, session.TenantId)
		} else {
			// For admin, we need a method to get all harvests - this would need to be added to harvest domain
			totalCount, _ := s.store.Db.Harvest.Count(ctx, filter)
			c.JSON(http.StatusOK, gin.H{
				"harvests": []interface{}{}, // Placeholder - would implement FindAll with pagination
				"total":    0,
				"count":    totalCount,
				"page":     page,
				"limit":    limit,
				"message":  "Admin harvest listing (placeholder implementation)",
			})
			return
		}

		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Get total count
		totalCount, _ := s.store.Db.Harvest.Count(ctx, filter)

		c.JSON(http.StatusOK, gin.H{
			"harvests": harvests,
			"total":    len(harvests),
			"count":    totalCount,
			"page":     page,
			"limit":    limit,
		})
	}
}

// v1_getProcessingHarvests handles GET /harvest/v1/admin/processing
func (s harvest) v1_getProcessingHarvests() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		stage := c.DefaultQuery("stage", "drying")

		// Get processing harvests
		harvests, err := s.store.Db.Harvest.FindByProcessingStage(ctx, stage, session.TenantId)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"harvests": harvests,
			"total":    len(harvests),
			"stage":    stage,
		})
	}
}

// v1_getHarvestAnalytics handles GET /harvest/v1/admin/analytics
func (s harvest) v1_getHarvestAnalytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		timeRange := c.DefaultQuery("time_range", "month")

		// Get analytics
		metrics, err := s.store.Db.Harvest.GetProcessingMetrics(ctx, session.TenantId, timeRange)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"analytics":  metrics,
			"time_range": timeRange,
		})
	}
}

// v1_qualityCheck handles POST /harvest/v1/admin/:id/quality-check
func (s harvest) v1_qualityCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		id := c.Param("id")

		var req QualityCheckRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Validate ID format
		if len(id) != 24 {
			c.Error(ecode.BadRequest.Desc(fmt.Errorf("invalid harvest ID format")))
			return
		}

		// Get harvest
		harvestDomain, err := s.store.Db.Harvest.FindByID(ctx, id)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "harvest_not_found").Desc(err))
			return
		}

		// Verify tenant
		if harvestDomain.TenantId == nil || *harvestDomain.TenantId != session.TenantId {
			c.Error(ecode.New(http.StatusForbidden, "harvest_unauthorized").Desc(fmt.Errorf("Harvest access denied")))
			return
		}

		// Create quality check data
		qualityData := db.QualityCheckData{
			CheckedBy:     session.UserId,
			CheckDate:     time.Now(),
			VisualQuality: req.VisualQuality,
			Moisture:      req.Moisture,
			Density:       req.Density,
			Notes:         req.Notes,
			Approved:      req.Approved,
		}

		// Record quality check
		err = s.store.Db.Harvest.RecordQualityCheck(ctx, id, qualityData)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		status := "quality_check"
		if req.Approved {
			status = "ready"
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Quality check recorded successfully",
			"approved":     req.Approved,
			"new_status":   status,
			"quality_data": qualityData,
		})
	}
}

// v1_forceStatusUpdate handles PUT /harvest/v1/admin/:id/force-status
func (s harvest) v1_forceStatusUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		id := c.Param("id")

		var req ForceStatusUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Validate ID format
		if len(id) != 24 {
			c.Error(ecode.BadRequest.Desc(fmt.Errorf("invalid harvest ID format")))
			return
		}

		// Get harvest
		harvestDomain, err := s.store.Db.Harvest.FindByID(ctx, id)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "harvest_not_found").Desc(err))
			return
		}

		// Verify tenant
		if harvestDomain.TenantId == nil || *harvestDomain.TenantId != session.TenantId {
			c.Error(ecode.New(http.StatusForbidden, "harvest_unauthorized").Desc(fmt.Errorf("Harvest access denied")))
			return
		}

		// Force update status (admin can override any status)
		if req.ProcessingStage != nil {
			err = s.store.Db.Harvest.UpdateProcessingStatus(ctx, id, *req.ProcessingStage, &req.Reason)
		} else {
			err = s.store.Db.Harvest.UpdateStatus(ctx, id, req.Status)
		}

		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Status updated successfully",
			"new_status": req.Status,
			"reason":     req.Reason,
			"updated_by": session.UserId,
		})
	}
}
