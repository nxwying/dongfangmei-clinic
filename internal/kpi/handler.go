package kpi

import (
	"net/http"
	"time"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListTargets(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ym := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
		var targets []model.KpiTarget
		db.Where("year_month=?", ym).Find(&targets)
		c.JSON(http.StatusOK, targets)
	}
}
func SaveTarget(db *gorm.DB) gin.HandlerFunc { // upsert
	return func(c *gin.Context) {
		var t model.KpiTarget
		c.ShouldBindJSON(&t)
		var existing model.KpiTarget
		db.Where("user_id=? AND year_month=?", t.UserID, t.YearMonth).First(&existing)
		if existing.ID > 0 {
			db.Model(&existing).Updates(map[string]interface{}{
				"revenue_target": t.RevenueTarget, "order_target": t.OrderTarget,
				"visit_target": t.VisitTarget, "followup_target": t.FollowupTarget,
				"new_customer_target": t.NewCustomerTarget,
			})
		} else {
			db.Create(&t)
		}
		c.JSON(http.StatusOK, gin.H{"message":"保存成功"})
	}
}

type RankRow struct {
	UserID       uint    `json:"user_id"`
	RealName     string  `json:"real_name"`
	Revenue      float64 `json:"revenue"`
	Gross        float64 `json:"gross"`
	Orders       int64   `json:"orders"`
	Followups    int64   `json:"followups"`
	NewCustomers int64   `json:"new_customers"`
}

func Leaderboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ym := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
		start := ym + "-01"
		end := ym + "-31"

		var rows []RankRow
		db.Raw(`SELECT u.id as user_id, u.real_name,
			COALESCE(SUM(o.final_amount),0) as revenue,
			COALESCE(SUM(o.final_amount - COALESCE(o.commission_amount,0)),0) as gross,
			COUNT(DISTINCT o.id) as orders,
			(SELECT COUNT(*) FROM follow_up_tasks WHERE created_by=u.id AND due_date>=? AND due_date<=? AND status='completed') as followups,
			(SELECT COUNT(*) FROM customers WHERE created_by=u.id AND DATE(created_at)>=? AND DATE(created_at)<=?) as new_customers
			FROM users u LEFT JOIN orders o ON o.created_by=u.id AND o.status='paid'
			AND DATE(o.created_at)>=? AND DATE(o.created_at)<=?
			GROUP BY u.id ORDER BY revenue DESC`, start, end, start, end, start, end).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}
