package medical

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
  CustomerID uint   `json:"customer_id" binding:"required"`
  RecordType string `json:"record_type" binding:"required"`
  RecordDate string `json:"record_date"`
  DoctorName string `json:"doctor_name"`
  Content    string `json:"content"`
}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.RecordDate == "" {
			req.RecordDate = time.Now().Format("2006-01-02")
		}
		userID := c.GetUint("user_id")
		rec := model.MedicalRecord{
			CustomerID: req.CustomerID, RecordType: req.RecordType,
			RecordDate: req.RecordDate, DoctorName: req.DoctorName,
		Content: req.Content, Status: "draft", CreatedBy: userID,
		}
		if err := db.Create(&rec).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		system.WriteAuditLog(db, c, "create", "medical_record", rec.ID, "创建病历")
		c.JSON(http.StatusCreated, rec)
	}
}

func ListRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cid := c.Param("id")
		var records []model.MedicalRecord
		db.Where("customer_id = ?", cid).Order("record_date DESC, created_at DESC").Find(&records)
		c.JSON(http.StatusOK, records)
	}
}

func ListAll(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Preload("Customer").Order("created_at DESC")
		if t := c.Query("type"); t != "" {
			query = query.Where("record_type = ?", t)
		}
		if q := c.Query("q"); q != "" {
			query = query.Where("customer_id IN (SELECT id FROM customers WHERE name ILIKE ? OR phone ILIKE ?)", "%"+q+"%", "%"+q+"%")
		}
		var records []model.MedicalRecord
		query.Find(&records)
		c.JSON(http.StatusOK, records)
	}
}

func GetRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var rec model.MedicalRecord
		if err := db.Preload("Customer").First(&rec, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		c.JSON(http.StatusOK, rec)
	}
}

func UpdateRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct {
			DoctorName string `json:"doctor_name"`
			TemplateID *uint   `json:"template_id"`
			Content    string `json:"content"`
			Status     string `json:"status"`
		}
		c.ShouldBindJSON(&req)
		updates := map[string]interface{}{}
		if req.DoctorName != "" { updates["doctor_name"] = req.DoctorName }
		if req.Content != "" { updates["content"] = req.Content }
		if req.Status != "" { updates["status"] = req.Status }
		db.Model(&model.MedicalRecord{}).Where("id = ?", id).Updates(updates)
		system.WriteAuditLog(db, c, "update", "medical_record", uint(id), "更新病历")
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

func DeleteRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.MedicalRecord{}, id)
		system.WriteAuditLog(db, c, "delete", "medical_record", uint(id), "删除病历")
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}

// SignRecord signs a medical record (electronic signature)
func SignRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var record model.MedicalRecord
		if err := db.First(&record, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "病历不存在"})
			return
		}
		if record.Status == "signed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "病历已签字，不可重复签字"})
			return
		}
		record.Status = "signed"
		db.Save(&record)
		c.JSON(http.StatusOK, record)
	}
}

// ======================== Template CRUD ========================

func ListTemplates(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var templates []model.MedicalRecordTemplate
		q := db.Order("category, sort_order, name")
		if cat := c.Query("category"); cat != "" {
			q = q.Where("category = ?", cat)
		}
		q.Find(&templates)
		c.JSON(http.StatusOK, templates)
	}
}

func CreateTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name          string `json:"name" binding:"required"`
			ProcedureName string `json:"procedure_name"`
			Category      string `json:"category" binding:"required"`
			Fields        string `json:"fields"`
			Description   string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：名称和分类必填"})
			return
		}
		tmpl := model.MedicalRecordTemplate{
			Name:          req.Name,
			ProcedureName: req.ProcedureName,
			Category:      req.Category,
			Fields:        req.Fields,
			Description:   req.Description,
			IsActive:      true,
		}
		if err := db.Create(&tmpl).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		c.JSON(http.StatusCreated, tmpl)
	}
}

func GetTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var tmpl model.MedicalRecordTemplate
		if err := db.First(&tmpl, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "模版不存在"})
			return
		}
		c.JSON(http.StatusOK, tmpl)
	}
}

func UpdateTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct {
			Name          *string `json:"name"`
			ProcedureName *string `json:"procedure_name"`
			Category      *string `json:"category"`
			Fields        *string `json:"fields"`
			Description   *string `json:"description"`
			IsActive      *bool   `json:"is_active"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		updates := map[string]interface{}{}
		if req.Name != nil { updates["name"] = *req.Name }
		if req.ProcedureName != nil { updates["procedure_name"] = *req.ProcedureName }
		if req.Category != nil { updates["category"] = *req.Category }
		if req.Fields != nil { updates["fields"] = *req.Fields }
		if req.Description != nil { updates["description"] = *req.Description }
		if req.IsActive != nil { updates["is_active"] = *req.IsActive }
		db.Model(&model.MedicalRecordTemplate{}).Where("id = ?", id).Updates(updates)
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

func DeleteTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.MedicalRecordTemplate{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}
