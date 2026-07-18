package settings

import (
	"encoding/json"
	"net/http"
	"strconv"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"clinic-mgmt/internal/system"
)

// ------- Request structs -------

type CreateTreatmentItemReq struct {
	Name     string  `json:"name" binding:"required"`
	Category string  `json:"category"`
	Price    float64 `json:"price" binding:"required"`
	Duration int     `json:"duration"`
}

type UpdateTreatmentItemReq struct {
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Price    *float64 `json:"price"`
	Duration *int     `json:"duration"`
	Status   *bool    `json:"status"`
}

type CreatePackageTemplateReq struct {
	Name         string          `json:"name" binding:"required"`
	Items        json.RawMessage `json:"items"`
	TotalSessions int            `json:"total_sessions" binding:"required"`
	Price        float64         `json:"price" binding:"required"`
}

type UpdatePackageTemplateReq struct {
	Name          string          `json:"name"`
	Items         json.RawMessage `json:"items"`
	TotalSessions *int            `json:"total_sessions"`
	Price         *float64        `json:"price"`
	Status        *string         `json:"status"`
}

// ------- Treatment Item Handlers -------

func treatmentItemToResponse(t model.TreatmentItem) gin.H {
	status := "active"
	if !t.IsActive {
		status = "inactive"
	}
	return gin.H{
		"id":         t.ID,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
		"name":       t.Name,
		"category":   t.Category,
		"price":      t.Price,
		"duration":   t.DurationMin,
		"status":     status,
	}
}

func treatmentItemsToResponse(items []model.TreatmentItem) []gin.H {
	result := make([]gin.H, len(items))
	for i, t := range items {
		result[i] = treatmentItemToResponse(t)
	}
	return result
}

func ListItems(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []model.TreatmentItem
		db.Find(&items)
		c.JSON(http.StatusOK, treatmentItemsToResponse(items))
	}
}

func CreateItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateTreatmentItemReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		item := model.TreatmentItem{
			Name:        req.Name,
			Category:    req.Category,
			Price:       req.Price,
			DurationMin: req.Duration,
			IsActive:    true,
		}

		if err := db.Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}

		system.WriteAuditLog(db, c, "create", "treatment_item", item.ID, "创建项目 "+item.Name)
		c.JSON(http.StatusCreated, treatmentItemToResponse(item))
	}
}

func UpdateItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
			return
		}

		var req UpdateTreatmentItemReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		updates := map[string]interface{}{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Category != "" {
			updates["category"] = req.Category
		}
		if req.Price != nil {
			updates["price"] = *req.Price
		}
		if req.Duration != nil {
			updates["duration_min"] = *req.Duration
		}
		if req.Status != nil {
			updates["is_active"] = *req.Status
		}

		if err := db.Model(&model.TreatmentItem{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}

		var item model.TreatmentItem
		db.First(&item, id)
		system.WriteAuditLog(db, c, "update", "treatment_item", uint(id), "修改项目")
		c.JSON(http.StatusOK, treatmentItemToResponse(item))
	}
}

// ------- Package Template Handlers -------

func packageTemplateToResponse(t *model.PackageTemplate) gin.H {
	status := "active"
	if !t.IsActive {
		status = "inactive"
	}
	return gin.H{
		"id":              t.ID,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
		"name":            t.Name,
		"description":     t.Description,
		"items":           t.Items,
		"price":           t.TotalPrice,
		"duration_months": t.DurationMonths,
		"total_sessions":  t.SessionCount,
		"status":          status,
	}
}

func packageTemplatesToResponse(templates []model.PackageTemplate) []gin.H {
	result := make([]gin.H, len(templates))
	for i, t := range templates {
		result[i] = packageTemplateToResponse(&t)
	}
	return result
}

func ListPackageTemplates(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var templates []model.PackageTemplate
		db.Find(&templates)
		c.JSON(http.StatusOK, packageTemplatesToResponse(templates))
	}
}

func CreatePackageTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePackageTemplateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		itemsStr := ""
		if len(req.Items) > 0 {
			itemsStr = string(req.Items)
		}
		template := model.PackageTemplate{
			Name:         req.Name,
			SessionCount: req.TotalSessions,
			TotalPrice:   req.Price,
			Items:        itemsStr,
			IsActive:     true,
		}

		if err := db.Create(&template).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}

		system.WriteAuditLog(db, c, "create", "package_template", template.ID, "创建套餐模板 "+template.Name)
		c.JSON(http.StatusCreated, packageTemplateToResponse(&template))
	}
}

func UpdatePackageTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
			return
		}

		var req UpdatePackageTemplateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		updates := map[string]interface{}{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if len(req.Items) > 0 {
			updates["items"] = string(req.Items)
		}
		if req.TotalSessions != nil {
			updates["session_count"] = *req.TotalSessions
		}
		if req.Price != nil {
			updates["total_price"] = *req.Price
		}
		if req.Status != nil {
			updates["is_active"] = *req.Status == "active"
		}

		if err := db.Model(&model.PackageTemplate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}

		var template model.PackageTemplate
		db.First(&template, id)
		c.JSON(http.StatusOK, packageTemplateToResponse(&template))
	}
}
