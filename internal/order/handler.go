package order

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"clinic-mgmt/internal/system"
)

type CreateOrderReq struct {
	CustomerID     uint              `json:"customer_id" binding:"required"`
	CommissionAmount float64         `json:"commission_amount"`
	CostAmount float64               `json:"cost_amount"`
	DiscountAmount float64           `json:"discount_amount"`
	Remark         string            `json:"remark"`
	Items          []CreateOrderItem `json:"items" binding:"required"`
}

type CreateOrderItem struct {
	ItemType  string  `json:"item_type"`
	ItemID    uint    `json:"item_id"`
	ItemName  string  `json:"item_name" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	UnitPrice float64 `json:"unit_price" binding:"required"`
}

type PayOrderReq struct {
	Payments []PayOrderPayment `json:"payments" binding:"required"`
}

type PayOrderPayment struct {
	PayMethod string  `json:"pay_method" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
}

func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("%s%06d", now.Format("20060102"), rand.Intn(1000000))
}

// CreateOrder handles POST /api/v1/orders
func CreateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateOrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "订单项不能为空"})
			return
		}

		var items []model.OrderItem
		var totalAmount float64
		for _, item := range req.Items {
			subtotal := float64(item.Quantity) * item.UnitPrice
			totalAmount += subtotal
			items = append(items, model.OrderItem{
				ItemType:  item.ItemType,
				ItemID:    item.ItemID,
				ItemName:  item.ItemName,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Subtotal:  subtotal,
			})
		}

		finalAmount := totalAmount - req.DiscountAmount
		if finalAmount < 0 {
			finalAmount = 0
		}

		userID := c.GetUint("user_id")

		order := model.Order{
			OrderNo:           generateOrderNo(),
			CustomerID:        req.CustomerID,
			TotalAmount:       totalAmount,
			DiscountAmount:    req.DiscountAmount,
			CommissionAmount:  req.CommissionAmount,
			CostAmount:        req.CostAmount,
			FinalAmount:       finalAmount,
			Status:            "pending",
			Remark:            req.Remark,
			CreatedBy:         userID,
			Items:             items,
		}

		if err := db.Create(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
			return
		}

		// Reload with associations
		db.Preload("Items").Preload("Customer").First(&order, order.ID)

		system.WriteAuditLog(db, c, "create", "order", order.ID, fmt.Sprintf("创建订单 ¥%.2f", order.FinalAmount))
		c.JSON(http.StatusCreated, order)
	}
}

// GetOrder handles GET /api/v1/orders/:id
func GetOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
			return
		}

		var order model.Order
		if err := db.Preload("Items").Preload("Payments").Preload("Customer").First(&order, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
			return
		}

		system.WriteAuditLog(db, c, "update", "order", uint(id), "支付订单")
		c.JSON(http.StatusOK, order)
	}
}

// ListOrders handles GET /api/v1/orders
func ListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&model.Order{})

		if dateStr := c.Query("date"); dateStr != "" {
			query = query.Where("DATE(created_at) = ?", dateStr)
		}
		if startDate := c.Query("start_date"); startDate != "" {
			query = query.Where("DATE(created_at) >= ?", startDate)
		}
		if endDate := c.Query("end_date"); endDate != "" {
			query = query.Where("DATE(created_at) <= ?", endDate)
		}
		if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
			if customerID, err := strconv.ParseUint(customerIDStr, 10, 64); err == nil {
				query = query.Where("customer_id = ?", customerID)
			}
		}
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}
		if q := c.Query("q"); q != "" {
			query = query.Where("customer_id IN (SELECT id FROM customers WHERE name ILIKE ? OR phone ILIKE ?)", "%"+q+"%", "%"+q+"%")
		}

		var orders []model.Order
		query.Preload("Customer").Preload("Items").Preload("Payments").Order("created_at DESC").Find(&orders)

		c.JSON(http.StatusOK, orders)
	}
}

// PayOrder handles POST /api/v1/orders/:id/pay
func PayOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
			return
		}

		var req PayOrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		if len(req.Payments) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "支付项不能为空"})
			return
		}

		var order model.Order
		if err := db.First(&order, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
			return
		}
		if order.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不是待支付"})
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			var payments []model.Payment
			var totalPaid float64

			for _, p := range req.Payments {
				// Handle balance deduction
				switch p.PayMethod {
				case "balance":
					result := tx.Model(&model.Membership{}).
						Where("customer_id = ? AND balance >= ?", order.CustomerID, p.Amount).
						Update("balance", gorm.Expr("balance - ?", p.Amount))
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected == 0 {
						return fmt.Errorf("余额不足")
					}
				case "gift_balance":
					result := tx.Model(&model.Membership{}).
						Where("customer_id = ? AND gift_balance >= ?", order.CustomerID, p.Amount).
						Update("gift_balance", gorm.Expr("gift_balance - ?", p.Amount))
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected == 0 {
						return fmt.Errorf("赠送余额不足")
					}
				}

				totalPaid += p.Amount
				payments = append(payments, model.Payment{
					OrderID:   uint(id),
					Amount:    p.Amount,
					PayMethod: p.PayMethod,
				})
			}

			// Create payment records
			if err := tx.Create(&payments).Error; err != nil {
				return err
			}

			// Update order status
			if err := tx.Model(&model.Order{}).Where("id = ?", id).Updates(map[string]interface{}{
				"status":      "paid",
				"paid_amount": totalPaid,
				"paid_at":     time.Now().Unix(),
			}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Return updated order
		db.Preload("Items").Preload("Payments").Preload("Customer").First(&order, id)
		c.JSON(http.StatusOK, order)
	}
}

// RefundOrder handles POST /api/v1/orders/:id/refund
func RefundOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
			return
		}

		var order model.Order
		if err := db.Preload("Payments").First(&order, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
			return
		}
		if order.Status != "paid" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不是已支付"})
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			for _, payment := range order.Payments {
				switch payment.PayMethod {
				case "balance":
					if err := tx.Model(&model.Membership{}).
						Where("customer_id = ?", order.CustomerID).
						Update("balance", gorm.Expr("balance + ?", payment.Amount)).Error; err != nil {
						return err
					}
				case "gift_balance":
					if err := tx.Model(&model.Membership{}).
						Where("customer_id = ?", order.CustomerID).
						Update("gift_balance", gorm.Expr("gift_balance + ?", payment.Amount)).Error; err != nil {
						return err
					}
				}
			}

			// Update order status
			if err := tx.Model(&model.Order{}).Where("id = ?", id).Update("status", "refunded").Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "退费失败"})
			return
		}

		system.WriteAuditLog(db, c, "refund", "order", uint(id), "退费")
		c.JSON(http.StatusOK, gin.H{"message": "退费成功"})
	}
}
