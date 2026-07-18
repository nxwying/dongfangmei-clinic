package commission

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---- Rule CRUD ----
func ListRules(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rules []model.CommissionRule
		db.Order("role, procedure").Find(&rules)
		c.JSON(http.StatusOK, rules)
	}
}
func CreateRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r model.CommissionRule
		if err := c.ShouldBindJSON(&r); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"参数错误"}); return }
		db.Create(&r)
		system.WriteAuditLog(db, c, "create", "commission_rule", r.ID, "创建提成规则: "+r.Name)
		c.JSON(http.StatusCreated, r)
	}
}
func UpdateRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req model.CommissionRule
		c.ShouldBindJSON(&req)
		updates := map[string]interface{}{}
		if req.Name != "" { updates["name"] = req.Name }
		if req.Role != "" { updates["role"] = req.Role }
		if req.Procedure != "" { updates["procedure"] = req.Procedure }
		updates["rate"] = req.Rate; updates["tier_min"] = req.TierMin; updates["tier_max"] = req.TierMax
		updates["rule_type"] = req.RuleType; updates["is_active"] = req.IsActive
		db.Model(&model.CommissionRule{}).Where("id=?", id).Updates(updates)
		system.WriteAuditLog(db, c, "update", "commission_rule", uint(id), "更新提成规则")
		c.JSON(http.StatusOK, gin.H{"message":"更新成功"})
	}
}
func DeleteRule(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.CommissionRule{}, id)
		system.WriteAuditLog(db, c, "delete", "commission_rule", uint(id), "删除提成规则")
		c.JSON(http.StatusOK, gin.H{"message":"已删除"})
	}
}

// ---- Calculate ----
func Calculate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ym := c.DefaultQuery("year_month", time.Now().Format("2006-01"))

		var rules []model.CommissionRule
		db.Where("is_active = ?", true).Find(&rules)
		if len(rules) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error":"没有活跃的提成规则"})
			return
		}

		// Fetch all paid orders in the month
		type OrderRow struct {
			ID        uint
			FinalAmount float64
			CreatedBy uint
		}
		var orders []OrderRow
		startDate := ym + "-01"
		endDate := ym + "-31"
		db.Raw(`SELECT id, final_amount, created_by FROM orders WHERE status='paid' 
			AND DATE(created_at) >= ? AND DATE(created_at) <= ?`, startDate, endDate).Scan(&orders)

		// Fetch treatment records for doctor commissions
		type TreatmentRow struct {
			OrderID   *uint
			DoctorID  *uint
			DoctorName string
			Items     string
		}
		var treatments []TreatmentRow
		db.Raw(`SELECT order_id, doctor_id, doctor_name, items FROM treatment_records 
			WHERE DATE(created_at) >= ? AND DATE(created_at) <= ?`, startDate, endDate).Scan(&treatments)
		txByOrder := map[uint]TreatmentRow{}
		for _, t := range treatments {
			if t.OrderID != nil { txByOrder[*t.OrderID] = t }
		}

		type DetailItem struct {
			OrderID    uint    `json:"order_id"`
			Amount     float64 `json:"amount"`
			Rule       string  `json:"rule"`
			Commission float64 `json:"commission"`
		}
		commByUser := map[uint]struct {
			name  string
			rev   float64
			comm  float64
			items []DetailItem
		}{}

		for _, o := range orders {
			// Consultant commission (created_by)
			uid := o.CreatedBy
			rev := o.FinalAmount
			for _, rule := range rules {
				if rule.Role == "consultant" && uid == uid {
					rate := rule.Rate
					if rule.RuleType == "tiered" {
						if rev >= rule.TierMin && (rule.TierMax == 0 || rev <= rule.TierMax) {
							// matched tier
						} else { continue }
					}
					comm := rev * rate / 100
					entry := commByUser[uid]
					entry.rev += rev
					entry.comm += comm
					entry.items = append(entry.items, DetailItem{OrderID: o.ID, Amount: rev, Rule: rule.Name, Commission: comm})
					commByUser[uid] = entry
				}
			}
			// Doctor commission (from treatment records)
			if tx, ok := txByOrder[o.ID]; ok && tx.DoctorID != nil {
				did := *tx.DoctorID
				for _, rule := range rules {
					if rule.Role == "doctor" && (rule.Procedure == "" || rule.Procedure == tx.Items) {
						comm := rev * rule.Rate / 100
						entry := commByUser[did]
						entry.rev += rev
						entry.comm += comm
						entry.items = append(entry.items, DetailItem{OrderID: o.ID, Amount: rev, Rule: rule.Name + " - " + tx.DoctorName, Commission: comm})
						commByUser[did] = entry
					}
				}
			}
		}

		// Store results
		var userNames = map[uint]string{}
		var userRows []struct{ ID uint; RealName string }
		db.Raw("SELECT id, real_name FROM users").Scan(&userRows)
		for _, u := range userRows { userNames[u.ID] = u.RealName }

		for uid, entry := range commByUser {
			detailBytes, _ := json.Marshal(entry.items)
			db.Where("user_id=? AND year_month=?", uid, ym).Delete(&model.CommissionResult{})
			db.Create(&model.CommissionResult{
				UserID: uid, RealName: userNames[uid], YearMonth: ym,
				TotalRevenue: entry.rev, TotalCommission: entry.comm,
				Details: string(detailBytes), Status: "draft",
			})
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("计算完成，共 %d 人", len(commByUser))})
	}
}

func ListResults(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ym := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
		var results []model.CommissionResult
		db.Where("year_month=?", ym).Order("total_commission DESC").Find(&results)
		c.JSON(http.StatusOK, results)
	}
}

func ConfirmResult(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Model(&model.CommissionResult{}).Where("id=?", id).Update("status", "confirmed")
		c.JSON(http.StatusOK, gin.H{"message":"已确认"})
	}
}
