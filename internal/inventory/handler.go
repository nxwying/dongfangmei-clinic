package inventory

import (
	"net/http"
	"strconv"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListItems(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []model.InventoryItem
		db.Order("name ASC").Find(&items)
		c.JSON(http.StatusOK, items)
	}
}

func CreateItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item model.InventoryItem
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if err := db.Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		system.WriteAuditLog(db, c, "create", "inventory", item.ID, "创建库存"+item.Name)
		c.JSON(http.StatusCreated, item)
	}
}

func UpdateItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req model.InventoryItem
		c.ShouldBindJSON(&req)
		updates := map[string]interface{}{}
		if req.Name != "" { updates["name"] = req.Name }
		if req.Category != "" { updates["category"] = req.Category }
		if req.Quantity != 0 { updates["quantity"] = req.Quantity }
		if req.Unit != "" { updates["unit"] = req.Unit }
		if req.MinStock != 0 { updates["min_stock"] = req.MinStock }
		if req.Price != 0 { updates["price"] = req.Price }
		if req.Supplier != "" { updates["supplier"] = req.Supplier }
		db.Model(&model.InventoryItem{}).Where("id = ?", id).Updates(updates)
		system.WriteAuditLog(db, c, "update", "inventory", uint(id), "更新库存")
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

func DeleteItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.InventoryItem{}, id)
		system.WriteAuditLog(db, c, "delete", "inventory", uint(id), "删除库存")
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}

func StockIn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct { Quantity float64 `json:"quantity" binding:"required"`; Note string `json:"note"` }
		if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"参数错误"});return }
		userID := c.GetUint("user_id")
		var item model.InventoryItem
		if err := db.First(&item, id).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error":"不存在"});return }
		newQty := item.Quantity + req.Quantity
		db.Model(&item).Update("quantity", newQty)
		db.Create(&model.InventoryLog{ItemID: uint(id), Type: "in", Quantity: req.Quantity, BalanceAfter: newQty, Note: req.Note, CreatedBy: userID})
		system.WriteAuditLog(db, c, "stock_in", "inventory", uint(id), "入库 "+strconv.FormatFloat(req.Quantity,'f',2,64))
		c.JSON(http.StatusOK, gin.H{"message": "入库成功", "quantity": newQty})
	}
}

func StockOut(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct { Quantity float64 `json:"quantity" binding:"required"`; Note string `json:"note"` }
		if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"参数错误"});return }
		userID := c.GetUint("user_id")
		var item model.InventoryItem
		if err := db.First(&item, id).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error":"不存在"});return }
		if item.Quantity < req.Quantity { c.JSON(http.StatusBadRequest, gin.H{"error":"库存不足"});return }
		newQty := item.Quantity - req.Quantity
		db.Model(&item).Update("quantity", newQty)
		db.Create(&model.InventoryLog{ItemID: uint(id), Type: "out", Quantity: req.Quantity, BalanceAfter: newQty, Note: req.Note, CreatedBy: userID})
		system.WriteAuditLog(db, c, "stock_out", "inventory", uint(id), "出库 "+strconv.FormatFloat(req.Quantity,'f',2,64))
		c.JSON(http.StatusOK, gin.H{"message": "出库成功", "quantity": newQty})
	}
}

func ListLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var logs []model.InventoryLog
		db.Where("item_id = ?", id).Preload("Item").Order("created_at DESC").Find(&logs)
		c.JSON(http.StatusOK, logs)
	}
}
