package pkg

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatePackageReq struct {
	PackageTemplateID uint `json:"package_template_id" binding:"required"`
}

type RedeemPackageReq struct {
	PackageID uint    `json:"package_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
}

func CreatePackage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var req CreatePackageReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		// Verify membership exists
		var membership model.Membership
		if err := db.Where("customer_id = ?", customerID).First(&membership).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "会员不存在"})
			return
		}
		_ = membership

		// Read package template
		var template model.PackageTemplate
		if err := db.First(&template, req.PackageTemplateID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "套餐模板不存在"})
			return
		}

		// Create member package
		now := time.Now()
		mp := model.MemberPackage{
			CustomerID:    uint(customerID),
			TemplateName:  template.Name,
			TotalSessions: template.SessionCount,
			UsedSessions:  0,
			TotalAmount:   template.TotalPrice,
			ActivatedAt:   &now,
			Status:        "active",
		}

		if template.DurationMonths > 0 {
			expiresAt := now.AddDate(0, template.DurationMonths, 0)
			mp.ExpiresAt = &expiresAt
		}

		if err := db.Create(&mp).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建套餐失败"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": mp.ID, "message": "套餐创建成功"})
	}
}

func ListPackages(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		// Verify membership exists
		var membership model.Membership
		if err := db.Where("customer_id = ?", customerID).First(&membership).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "会员不存在"})
			return
		}
		_ = membership

		var packages []model.MemberPackage
		db.Where("customer_id = ?", uint(customerID)).Order("created_at desc").Find(&packages)
		c.JSON(http.StatusOK, packages)
	}
}

func RedeemPackage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
			return
		}

		var req RedeemPackageReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		// Validate the order exists and is not already paid
		var order model.Order
		if err := db.First(&order, orderID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
			return
		}

		if order.Status == "paid" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "订单已支付"})
			return
		}

		// Transaction: redeem session + create redemption record + mark order paid
		tx := db.Begin()

		var mp model.MemberPackage
		if err := tx.First(&mp, req.PackageID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "套餐不存在"})
			return
		}

		if mp.UsedSessions >= mp.TotalSessions {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "套餐次数已用完"})
			return
		}

		newUsed := mp.UsedSessions + 1
		if err := tx.Model(&mp).Update("used_sessions", newUsed).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新套餐次数失败"})
			return
		}

		redemption := model.PackageRedemption{
			MemberPackageID: mp.ID,
			SessionIndex:    newUsed,
		}
		if err := tx.Create(&redemption).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建核销记录失败"})
			return
		}

		updates := map[string]interface{}{
			"status":      "paid",
			"paid_amount": req.Amount,
		}
		if err := tx.Model(&model.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新订单状态失败"})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "核销成功"})
	}
}
