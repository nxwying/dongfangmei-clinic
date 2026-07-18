package marketing

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DormantCustomers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type Result struct {
			ID            uint    `json:"id"`
			Name          string  `json:"name"`
			Phone         string  `json:"phone"`
			LastVisit     string  `json:"last_visit"`
			DaysSinceLast int     `json:"days_since_last"`
			TotalSpent    float64 `json:"total_spent"`
		}
		var results []Result
		db.Raw(`SELECT c.id,c.name,c.phone,
			COALESCE(MAX(o.created_at)::text,'从未消费') as last_visit,
			COALESCE(EXTRACT(DAY FROM NOW()-MAX(o.created_at)),999)::int as days_since_last,
			COALESCE(SUM(o.final_amount),0) as total_spent
			FROM customers c LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
			GROUP BY c.id
			HAVING COALESCE(MAX(o.created_at),'2000-01-01') < NOW() - interval '90 days'
			ORDER BY days_since_last DESC`).Scan(&results)
		if results == nil { results = []Result{} }
		if results == nil { results = []Result{} }
		if results == nil { results = []Result{} }
		c.JSON(http.StatusOK, results)
	}
}

func BirthdayCustomers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type Result struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Phone string `json:"phone"`
			Birthday string `json:"birthday"`
		}
		var results []Result
		month := time.Now().Format("01")
		db.Raw(`SELECT id,name,phone,birthday FROM customers
			WHERE birthday != '' AND SUBSTRING(birthday,6,2)=?`, month).Scan(&results)
		if results == nil { results = []Result{} }
		c.JSON(http.StatusOK, results)
	}
}

func NearDormantCustomers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type Result struct {
			ID            uint    `json:"id"`
			Name          string  `json:"name"`
			Phone         string  `json:"phone"`
			LastVisit     string  `json:"last_visit"`
			DaysSinceLast int     `json:"days_since_last"`
			TotalSpent    float64 `json:"total_spent"`
		}
		var results []Result
		db.Raw(`SELECT c.id,c.name,c.phone,
			COALESCE(MAX(o.created_at)::text,'从未消费') as last_visit,
			COALESCE(EXTRACT(DAY FROM NOW()-MAX(o.created_at)),999)::int as days_since_last,
			COALESCE(SUM(o.final_amount),0) as total_spent
			FROM customers c LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
			GROUP BY c.id
			HAVING COALESCE(MAX(o.created_at),'2000-01-01') >= NOW() - interval '90 days'
			AND COALESCE(MAX(o.created_at),'2000-01-01') < NOW() - interval '60 days'
			ORDER BY total_spent DESC`).Scan(&results)
		if results == nil { results = []Result{} }
		c.JSON(http.StatusOK, results)
	}
}
