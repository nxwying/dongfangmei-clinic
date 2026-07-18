package treatment

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CustomerID    uint   `json:"customer_id" binding:"required"`
			OrderID       *uint  `json:"order_id"`
			Items         string `json:"items"`
			DoctorID      *uint  `json:"doctor_id"`
			DoctorName    string `json:"doctor_name"`
			Notes         string `json:"notes"`
			Dosage        string `json:"dosage"`
			Products      string `json:"products"`
			TreatmentDate string `json:"treatment_date"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.TreatmentDate == "" {
			req.TreatmentDate = time.Now().Format("2006-01-02")
		}
		rec := model.TreatmentRecord{
			CustomerID:    req.CustomerID,
			OrderID:       req.OrderID,
			Items:         req.Items,
			DoctorID:      req.DoctorID,
			DoctorName:    req.DoctorName,
			Notes:         req.Notes,
			Dosage:        req.Dosage,
			Products:      req.Products,
			TreatmentDate: req.TreatmentDate,
		}
		if err := db.Create(&rec).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		system.WriteAuditLog(db, c, "create", "treatment", rec.ID, "创建治疗记录")
		userID := c.GetUint("user_id")
		// Auto-generate post-treatment followups (D1/D3/D7/D30)
		go func() {
			schedules := []struct{ day int; label string }{
				{1, "D1 术后第1天回访"},
				{3, "D3 术后第3天回访"},
				{7, "D7 术后第7天回访"},
				{30, "D30 术后第30天回访"},
			}
			for _, s := range schedules {
				due := time.Now().AddDate(0, 0, s.day).Format("2006-01-02")
				db.Create(&model.FollowUpTask{
					CustomerID: rec.CustomerID,
					Type:       "post_treatment",
					DueDate:    due,
					Status:     "pending",
					Note:       s.label + " - " + rec.Items,
					CreatedBy:  userID,
				})
			}
		}()
		c.JSON(http.StatusCreated, rec)
	}
}

func ListRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("id")
		var records []model.TreatmentRecord
		query := db.Where("customer_id = ?", customerID).Order("treatment_date DESC")
		query.Find(&records)
		c.JSON(http.StatusOK, records)
	}
}

func UpdateRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct {
			Notes    string `json:"notes"`
			Dosage   string `json:"dosage"`
			Products string `json:"products"`
		}
		c.ShouldBindJSON(&req)
		updates := map[string]interface{}{}
		if req.Notes != "" { updates["notes"] = req.Notes }
		if req.Dosage != "" { updates["dosage"] = req.Dosage }
		if req.Products != "" { updates["products"] = req.Products }
		db.Model(&model.TreatmentRecord{}).Where("id = ?", id).Updates(updates)
		system.WriteAuditLog(db, c, "update", "treatment", uint(id), "更新治疗记录")
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

func GetRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var rec model.TreatmentRecord
		if err := db.First(&rec, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		c.JSON(http.StatusOK, rec)
	}
}
