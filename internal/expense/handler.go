package expense

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateExpenseReq struct {
	Type     string  `json:"type" binding:"required"`   // "commission" or "cost"
	Category string  `json:"category" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Note     string  `json:"note"`
	Date     string  `json:"date"` // YYYY-MM-DD, defaults to today
}

type UpdateExpenseReq struct {
	Type     string   `json:"type"`
	Category string   `json:"category"`
	Amount   *float64 `json:"amount"`
	Note     string   `json:"note"`
	Date     string   `json:"date"`
}

func ListExpenses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&model.Expense{}).Order("date DESC, created_at DESC")

		if t := c.Query("type"); t != "" {
			query = query.Where("type = ?", t)
		}
		if date := c.Query("date"); date != "" {
			query = query.Where("date = ?", date)
		}
		if start := c.Query("start_date"); start != "" {
			query = query.Where("date >= ?", start)
		}
		if end := c.Query("end_date"); end != "" {
			query = query.Where("date <= ?", end)
		}

		var expenses []model.Expense
		query.Find(&expenses)
		c.JSON(http.StatusOK, expenses)
	}
}

func CreateExpense(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateExpenseReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.Date == "" {
			req.Date = time.Now().Format("2006-01-02")
		}

		exp := model.Expense{
			Type:     req.Type,
			Category: req.Category,
			Amount:   req.Amount,
			Note:     req.Note,
			Date:     req.Date,
		}
		if err := db.Create(&exp).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		system.WriteAuditLog(db, c, "create", "expense", exp.ID, "创建支出 "+exp.Note)
		c.JSON(http.StatusCreated, exp)
	}
}

func UpdateExpense(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req UpdateExpenseReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		updates := map[string]interface{}{}
		if req.Type != "" {
			updates["type"] = req.Type
		}
		if req.Category != "" {
			updates["category"] = req.Category
		}
		if req.Amount != nil {
			updates["amount"] = *req.Amount
		}
		if req.Date != "" {
			updates["date"] = req.Date
		}
		if req.Note != "" {
			updates["note"] = req.Note
		}
		if len(updates) > 0 {
			db.Model(&model.Expense{}).Where("id = ?", id).Updates(updates)
		}
		system.WriteAuditLog(db, c, "update", "expense", uint(id), "修改支出")
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

func DeleteExpense(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.Expense{}, id)
		system.WriteAuditLog(db, c, "delete", "expense", uint(id), "删除支出")
		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}
