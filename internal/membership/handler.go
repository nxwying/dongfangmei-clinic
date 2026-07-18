package membership

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"

	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"clinic-mgmt/internal/system"
)

type createMembershipReq struct {
	Level string `json:"level"`
}

type rechargeReq struct {
	Amount     float64 `json:"amount" binding:"required"`
	GiftAmount float64 `json:"gift_amount"`
}

// CreateMembership creates a membership for an existing customer.
// POST /api/v1/customers/:id/membership
func CreateMembership(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var req createMembershipReq
		if err := c.ShouldBindJSON(&req); err != nil {
			req.Level = "regular"
		}
		if req.Level == "" {
			req.Level = "regular"
		}

		// Check customer exists
		var customer model.Customer
		if err := db.First(&customer, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
			return
		}

		// Check membership doesn't already exist
		var existing model.Membership
		if err := db.Where("customer_id = ?", id).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该客户已是会员"})
			return
		}

		membership := model.Membership{
			CustomerID: uint(id),
			Level:      req.Level,
			OpenedAt:   time.Now().Unix(),
		}

		if err := db.Create(&membership).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会员失败"})
			return
		}

		system.WriteAuditLog(db, c, "create", "membership", uint(id), "开卡: "+req.Level)
		c.JSON(http.StatusCreated, membership)
	}
}

// Recharge adds balance to a customer's membership.
// POST /api/v1/customers/:id/recharge
func Recharge(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var req rechargeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "充值金额必须大于0"})
			return
		}

		var membership model.Membership
		if err := db.Where("customer_id = ?", id).First(&membership).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "会员不存在"})
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			// Update membership balances
			if err := tx.Model(&membership).Updates(map[string]interface{}{
				"balance":         gorm.Expr("balance + ?", req.Amount),
				"gift_balance":    gorm.Expr("gift_balance + ?", req.GiftAmount),
				"total_recharged": gorm.Expr("total_recharged + ?", req.Amount),
			}).Error; err != nil {
				return err
			}


			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "充值失败: " + err.Error()})
			return
		}

		system.WriteAuditLog(db, c, "recharge", "membership", uint(id), fmt.Sprintf("充值: +¥%.2f(赠送¥%.2f)", req.Amount, req.GiftAmount))
		// Reload to return updated membership
		db.Where("customer_id = ?", id).First(&membership)
		c.JSON(http.StatusOK, membership)
	}
}

// GetMembership retrieves a customer's membership.
// GET /api/v1/customers/:id/membership
func GetMembership(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var membership model.Membership
		if err := db.Where("customer_id = ?", id).First(&membership).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "会员不存在"})
			return
		}

		c.JSON(http.StatusOK, membership)
	}
}
