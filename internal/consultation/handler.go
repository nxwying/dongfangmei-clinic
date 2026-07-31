package consultation

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/middleware"
	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// List returns consultations with filters and pagination.
func List(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&model.Consultation{})
		if cid := c.Query("customer_id"); cid != "" {
			query = query.Where("customer_id = ?", cid)
		}
		if level := c.Query("intention_level"); level != "" {
			query = query.Where("intention_level = ?", level)
		}
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}
		if consultantID := c.Query("consultant_id"); consultantID != "" {
			query = query.Where("consultant_id = ?", consultantID)
		}
		if startDate := c.Query("start_date"); startDate != "" {
			query = query.Where("contact_date >= ?", startDate)
		}
		if endDate := c.Query("end_date"); endDate != "" {
			query = query.Where("contact_date <= ?", endDate)
		}

		p := middleware.ParsePagination(c)
		var total int64
		query.Count(&total)

		var list []model.Consultation
		query.Preload("Customer").Order("contact_date DESC, created_at DESC").
			Offset(p.Offset()).Limit(p.PageSize).Find(&list)

		c.JSON(http.StatusOK, middleware.PaginatedResult(list, total, p))
	}
}

// Create adds a new consultation record.
func Create(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.Consultation
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.CustomerID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择客户"})
			return
		}
		if req.ContactDate == "" {
			req.ContactDate = time.Now().Format("2006-01-02")
		}
		if req.IntentionLevel == "" {
			req.IntentionLevel = "C"
		}
		if req.Status == "" {
			req.Status = "pending"
		}
		req.CreatedBy = c.GetUint("user_id")

		if err := db.Create(&req).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}

		system.WriteAuditLog(db, c, "create", "consultation", req.ID, "新增咨询跟单")
		c.JSON(http.StatusCreated, req)
	}
}

// Update modifies a consultation record.
func Update(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req model.Consultation
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		req.ID = uint(id)
		if err := db.Model(&model.Consultation{}).Where("id = ?", id).Updates(map[string]interface{}{
			"contact_date":     req.ContactDate,
			"contact_method":   req.ContactMethod,
			"content":          req.Content,
			"customer_concern": req.CustomerConcern,
			"intention_level":  req.IntentionLevel,
			"interested_items": req.InterestedItems,
			"quoted_amount":    req.QuotedAmount,
			"estimated_amount": req.EstimatedAmount,
			"status":           req.Status,
			"next_contact_date": req.NextContactDate,
			"next_action":      req.NextAction,
			"remark":           req.Remark,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
		system.WriteAuditLog(db, c, "update", "consultation", uint(id), "更新咨询跟单")
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// Delete removes a consultation record.
func Delete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := db.Delete(&model.Consultation{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}

// Stats returns pipeline summary by intention level.
func Stats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type levelStat struct {
			Level       string  `json:"level"`
			Count       int64   `json:"count"`
			TotalEst    float64 `json:"total_estimated"`
		}
		var stats []levelStat
		db.Raw(`SELECT intention_level as level, COUNT(*) as count,
			COALESCE(SUM(estimated_amount),0) as total_est
			FROM consultations WHERE deleted_at IS NULL AND status = 'pending'
			GROUP BY intention_level ORDER BY level`).Scan(&stats)

		// Today's and upcoming follow-ups
		today := time.Now().Format("2006-01-02")
		var todayCount int64
		db.Model(&model.Consultation{}).Where("next_contact_date = ? AND status = 'pending'", today).Count(&todayCount)
		var overdueCount int64
		db.Model(&model.Consultation{}).Where("next_contact_date < ? AND status = 'pending' AND next_contact_date != ''", today).Count(&overdueCount)

		// Win rate
		var wonCount, totalCount int64
		db.Model(&model.Consultation{}).Where("status = 'won'").Count(&wonCount)
		db.Model(&model.Consultation{}).Where("status IN ('won','lost')").Count(&totalCount)
		var winRate float64
		if totalCount > 0 {
			winRate = float64(wonCount) / float64(totalCount) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"by_level":      stats,
			"today_followup": todayCount,
			"overdue":       overdueCount,
			"won":           wonCount,
			"win_rate":      winRate,
		})
	}
}
