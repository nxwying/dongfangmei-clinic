package label

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListRules(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rules []model.TagRule
		db.Order("apply_order ASC").Find(&rules)
		c.JSON(http.StatusOK, rules)
	}
}

func CreateRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rule model.TagRule
		if err := c.ShouldBindJSON(&rule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if err := db.Create(&rule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		system.WriteAuditLog(db, c, "create", "tag_rule", rule.ID, "创建标签规则: "+rule.Name)
		c.JSON(http.StatusCreated, rule)
	}
}

func UpdateRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req model.TagRule
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		updates := map[string]interface{}{}
		if req.Name != "" { updates["name"] = req.Name }
		if req.Conditions != "" { updates["conditions"] = req.Conditions }
		if req.Tag != "" { updates["tag"] = req.Tag }
		updates["is_active"] = req.IsActive
		updates["apply_order"] = req.ApplyOrder
		db.Model(&model.TagRule{}).Where("id = ?", id).Updates(updates)
		system.WriteAuditLog(db, c, "update", "tag_rule", uint(id), "更新标签规则")
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

func DeleteRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.TagRule{}, id)
		system.WriteAuditLog(db, c, "delete", "tag_rule", uint(id), "删除标签规则")
		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

type condition struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value float64 `json:"value"`
}

// ApplyRules runs all active rules and updates customer tags.
func ApplyRules(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rules []model.TagRule
		db.Where("is_active = ?", true).Order("apply_order ASC").Find(&rules)

		type csResult struct {
			ID     uint   `json:"id"`
			Name   string `json:"name"`
			Orders int64  `json:"orders"`
			Total  float64 `json:"total"`
			Tags   string `json:"tags"`
			VisitCount    int64      `json:"visit_count"`
			LastVisitAt   *time.Time `json:"last_visit_at"`
			CreatedAt     time.Time  `json:"created_at"`
		}

		var customers []csResult
		db.Raw(`SELECT c.id, c.name, c.tags,
			COALESCE(COUNT(DISTINCT CASE WHEN o.status='paid' THEN o.id END),0) as orders,
			COALESCE(SUM(CASE WHEN o.status='paid' THEN o.final_amount ELSE 0 END),0) as total,
			COALESCE(COUNT(DISTINCT CASE WHEN a.status='completed' THEN a.id END),0) as visit_count,
			MAX(CASE WHEN a.status='completed' THEN a.created_at END) as last_visit_at
			FROM customers c
			LEFT JOIN orders o ON c.id=o.customer_id
			LEFT JOIN appointments a ON c.id=a.customer_id
			GROUP BY c.id`).Scan(&customers)

		updatedCount := 0
		for _, cust := range customers {
			var currentTags []string
			if cust.Tags != "" {
				json.Unmarshal([]byte(cust.Tags), &currentTags)
			}

			newTags := make([]string, len(currentTags))
			copy(newTags, currentTags)

			for _, rule := range rules {
				var conds []condition
				if err := json.Unmarshal([]byte(rule.Conditions), &conds); err != nil {
					continue
				}

				matched := true
				for _, cond := range conds {
					switch cond.Field {
					case "orders":
						v := int64(cond.Value)
						if cond.Op == ">=" && cust.Orders < v { matched = false }
						if cond.Op == ">" && cust.Orders <= v { matched = false }
						if cond.Op == "==" && cust.Orders != v { matched = false }
						if cond.Op == "<" && cust.Orders >= v { matched = false }
						if cond.Op == "<=" && cust.Orders > v { matched = false }
						if cond.Op == "!=" && cust.Orders == v { matched = false }
					case "total", "avg_order_amount":
						v := cond.Value
						compare := cust.Total
						if cond.Field == "avg_order_amount" && cust.Orders > 0 { compare = cust.Total / float64(cust.Orders) }
						if cond.Op == ">=" && compare < v { matched = false }
						if cond.Op == ">" && compare <= v { matched = false }
						if cond.Op == "==" && compare != v { matched = false }
						if cond.Op == "<" && compare >= v { matched = false }
						if cond.Op == "<=" && compare > v { matched = false }
						if cond.Op == "!=" && compare == v { matched = false }
					case "visit_count":
						v := int64(cond.Value)
						if cond.Op == ">=" && cust.VisitCount < v { matched = false }
						if cond.Op == ">" && cust.VisitCount <= v { matched = false }
						if cond.Op == "==" && cust.VisitCount != v { matched = false }
						if cond.Op == "<" && cust.VisitCount >= v { matched = false }
						if cond.Op == "<=" && cust.VisitCount > v { matched = false }
						if cond.Op == "!=" && cust.VisitCount == v { matched = false }
					case "days_since_last_visit":
						v := int64(cond.Value)
						days := int64(9999)
						if cust.LastVisitAt != nil { days = int64(time.Since(*cust.LastVisitAt).Hours() / 24) }
						if cond.Op == ">=" && days < v { matched = false }
						if cond.Op == ">" && days <= v { matched = false }
						if cond.Op == "==" && days != v { matched = false }
						if cond.Op == "<" && days >= v { matched = false }
						if cond.Op == "<=" && days > v { matched = false }
						if cond.Op == "!=" && days == v { matched = false }
					case "days_since_register":
						v := int64(cond.Value)
						days := int64(time.Since(cust.CreatedAt).Hours() / 24)
						if cond.Op == ">=" && days < v { matched = false }
						if cond.Op == ">" && days <= v { matched = false }
						if cond.Op == "==" && days != v { matched = false }
						if cond.Op == "<" && days >= v { matched = false }
						if cond.Op == "<=" && days > v { matched = false }
						if cond.Op == "!=" && days == v { matched = false }
					}
				}

				if matched {
					found := false
					for _, t := range newTags {
						if t == rule.Tag { found = true; break }
					}
					if !found {
						newTags = append(newTags, rule.Tag)
					}
				}
			}

			newTagsJSON, _ := json.Marshal(newTags)
			if string(newTagsJSON) != cust.Tags {
				db.Model(&model.Customer{}).Where("id = ?", cust.ID).Update("tags", newTagsJSON)
				updatedCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{"updated": updatedCount, "message": "更新了 " + strconv.Itoa(updatedCount) + " 位客户标签"})
	}
}
