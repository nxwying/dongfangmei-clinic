package consent

import (
	"net/http"
	"strconv"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func List(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cid := c.Query("customer_id")
		q := db.Model(&model.ConsentForm{}).Order("created_at DESC")
		if cid != "" { q = q.Where("customer_id=?", cid) }
		var list []model.ConsentForm
		q.Find(&list)
		c.JSON(http.StatusOK, list)
	}
}

func Create(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var f model.ConsentForm
		if err := c.ShouldBindJSON(&f); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"参数错误"}); return }
		f.CreatedBy = c.GetUint("user_id")
		if err := db.Create(&f).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"创建失败"}); return }
		system.WriteAuditLog(db, c, "create", "consent", f.ID, "创建知情同意书")
		c.JSON(http.StatusCreated, f)
	}
}

func Update(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req model.ConsentForm
		c.ShouldBindJSON(&req)
		updates := map[string]interface{}{}
		if req.PatientSign != "" { updates["patient_sign"] = req.PatientSign }
		if req.DoctorSign != "" { updates["doctor_sign"] = req.DoctorSign }
		if req.Status != "" { updates["status"] = req.Status }
		if req.SignDate != "" { updates["sign_date"] = req.SignDate }
		if req.Content != "" { updates["content"] = req.Content }
		if req.ProcedureName != "" { updates["procedure_name"] = req.ProcedureName }
		if len(updates) > 0 { db.Model(&model.ConsentForm{}).Where("id=?", id).Updates(updates) }
		system.WriteAuditLog(db, c, "update", "consent", uint(id), "更新知情同意书")
		c.JSON(http.StatusOK, gin.H{"message":"更新成功"})
	}
}

func Get(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var f model.ConsentForm
		if err := db.First(&f, id).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error":"不存在"}); return }
		c.JSON(http.StatusOK, f)
	}
}

func Delete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.ConsentForm{}, id)
		system.WriteAuditLog(db, c, "delete", "consent", uint(id), "删除知情同意书")
		c.JSON(http.StatusOK, gin.H{"message":"已删除"})
	}
}
