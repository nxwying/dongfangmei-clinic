package analysis

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LTV(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type ltvRow struct {
			Source        string  `json:"source"`
			CustomerCount int64   `json:"customer_count"`
			TotalRevenue  float64 `json:"total_revenue"`
			AvgRevenue    float64 `json:"avg_revenue"`
			AvgOrders     float64 `json:"avg_orders"`
		}
		var rows []ltvRow
		db.Raw(`SELECT COALESCE(NULLIF(TRIM(c.source),''),'其他') as source,
			COUNT(DISTINCT c.id) as customer_count,
			COALESCE(SUM(o.final_amount),0) as total_revenue,
			CASE WHEN COUNT(DISTINCT c.id)>0 THEN COALESCE(SUM(o.final_amount),0)/COUNT(DISTINCT c.id) ELSE 0 END as avg_revenue,
			CASE WHEN COUNT(DISTINCT c.id)>0 THEN COUNT(o.id)*1.0/COUNT(DISTINCT c.id) ELSE 0 END as avg_orders
			FROM customers c LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
			GROUP BY 1 ORDER BY total_revenue DESC`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

func CrossSell(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type pairRow struct {
			ItemA string `json:"item_a"`
			ItemB string `json:"item_b"`
			Count int64  `json:"count"`
		}
		var rows []pairRow
		db.Raw(`SELECT a.item_name as item_a, b.item_name as item_b, COUNT(*) as count
			FROM order_items a JOIN order_items b ON a.order_id=b.order_id AND a.id<b.id
			GROUP BY a.item_name, b.item_name ORDER BY count DESC LIMIT 20`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

func NoShow(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type nsRow struct {
			TimeSlot string `json:"time_slot"`
			Total    int64  `json:"total"`
			NoShow   int64  `json:"no_show"`
			Rate     float64 `json:"rate"`
		}
		var rows []nsRow
		db.Raw(`SELECT time_slot, COUNT(*) as total,
			SUM(CASE WHEN status='cancelled' OR status='no_show' THEN 1 ELSE 0 END) as no_show,
			AVG(CASE WHEN status='cancelled' OR status='no_show' THEN 100.0 ELSE 0 END) as rate
			FROM appointments WHERE date>=? GROUP BY time_slot ORDER BY total DESC`,
			time.Now().AddDate(0, -3, 0).Format("2006-01-02")).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

func Churn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type churnRow struct {
			ID       uint   `json:"id"`
			Name     string `json:"name"`
			Phone    string `json:"phone"`
			LastVisit string `json:"last_visit"`
			DaysGone int    `json:"days_gone"`
			Revenue  float64 `json:"revenue"`
		}
		var rows []churnRow
		db.Raw(`SELECT c.id, c.name, c.phone, MAX(o.created_at)::date as last_visit,
			EXTRACT(DAY FROM NOW()-MAX(o.created_at))::int as days_gone,
			COALESCE(SUM(o.final_amount),0) as revenue
			FROM customers c LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
			GROUP BY c.id HAVING EXTRACT(DAY FROM NOW()-MAX(o.created_at))>60
			ORDER BY days_gone DESC LIMIT 50`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

// ProcedureProfit calculates revenue and estimated cost per procedure.
func ProcedureProfit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type row struct {
			Procedure  string  `json:"procedure"`
			Revenue    float64 `json:"revenue"`
			Gross      float64 `json:"gross"`
			OrderCount int64   `json:"order_count"`
			EstCost    float64 `json:"est_cost"`
			Margin     float64 `json:"margin"`
			MarginPct  float64 `json:"margin_pct"`
		}
		var rows []row
		db.Raw(`SELECT oi.item_name as procedure,
			COALESCE(SUM(oi.subtotal),0) as revenue,
			COALESCE(SUM(oi.subtotal - (COALESCE(o.commission_amount,0) * oi.subtotal / NULLIF(o.final_amount,0))),0) as gross,
			COUNT(DISTINCT oi.order_id) as order_count,
			COALESCE(SUM(e.amount),0) as est_cost,
			COALESCE(SUM(oi.subtotal),0)-COALESCE(SUM(e.amount),0) as margin,
			CASE WHEN SUM(oi.subtotal)>0 THEN (SUM(oi.subtotal)-COALESCE(SUM(e.amount),0))/SUM(oi.subtotal)*100 ELSE 0 END as margin_pct
			FROM order_items oi
			JOIN orders o ON oi.order_id=o.id AND o.status='paid'
			LEFT JOIN expenses e ON e.category=oi.item_name AND e.date>=DATE(o.created_at)-30
			GROUP BY oi.item_name ORDER BY revenue DESC`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

// MonthlyTrend returns 12-month trend data.
func MonthlyTrend(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type row struct {
			Month         string  `json:"month"`
			Revenue       float64 `json:"revenue"`
			CustomerCount int64   `json:"customer_count"`
			OrderCount    int64   `json:"order_count"`
			AvgOrder      float64 `json:"avg_order"`
		}
		var rows []row
		db.Raw(`SELECT TO_CHAR(DATE_TRUNC('month',o.created_at),'YYYY-MM') as month,
			COALESCE(SUM(o.final_amount),0) as revenue,
			COUNT(DISTINCT o.customer_id) as customer_count,
			COUNT(DISTINCT o.id) as order_count,
			CASE WHEN COUNT(DISTINCT o.id)>0 THEN COALESCE(SUM(o.final_amount),0)/COUNT(DISTINCT o.id) ELSE 0 END as avg_order
			FROM orders o WHERE o.status='paid' AND o.created_at>=NOW()-INTERVAL'12 months'
			GROUP BY 1 ORDER BY 1`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

// RevenueStructure returns revenue breakdown by procedure and payment.
func RevenueStructure(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type item struct{ Name string `json:"name"`; Revenue float64 `json:"revenue"`; Count int64 `json:"count"`; Pct float64 `json:"pct"` }
		var byProc, byPayment []item
		var totalRev float64
		db.Raw(`SELECT COALESCE(SUM(oi.subtotal),0) FROM order_items oi JOIN orders o ON oi.order_id=o.id AND o.status='paid'`).Scan(&totalRev)
		db.Raw(`SELECT oi.item_name as name, SUM(oi.subtotal) as revenue, COUNT(*) as count
			FROM order_items oi JOIN orders o ON oi.order_id=o.id AND o.status='paid'
			GROUP BY oi.item_name ORDER BY revenue DESC`).Scan(&byProc)
		db.Raw(`SELECT pay_method as name, SUM(amount) as revenue, COUNT(*) as count
			FROM payments GROUP BY pay_method ORDER BY revenue DESC`).Scan(&byPayment)
		for i := range byProc { if totalRev > 0 { byProc[i].Pct = byProc[i].Revenue / totalRev * 100 } }
		for i := range byPayment { if totalRev > 0 { byPayment[i].Pct = byPayment[i].Revenue / totalRev * 100 } }
		c.JSON(http.StatusOK, gin.H{"by_procedure": byProc, "by_payment": byPayment, "total_revenue": totalRev})
	}
}

// Repurchase returns repurchase rate analysis.
func Repurchase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type overall struct{ TotalCustomers int64 `json:"total_customers"`; Repeat int64 `json:"repeat"`; Rate float64 `json:"rate"` }
		var ov overall
		db.Raw(`SELECT COUNT(*) as total_customers,
			COUNT(CASE WHEN order_count>=2 THEN 1 END) as repeat,
			CASE WHEN COUNT(*)>0 THEN COUNT(CASE WHEN order_count>=2 THEN 1 END)*100.0/COUNT(*) ELSE 0 END as rate
			FROM (SELECT customer_id,COUNT(*) as order_count FROM orders WHERE status='paid' GROUP BY customer_id) sub`).Scan(&ov)
		type proc struct{ Name string `json:"name"`; Orders int64 `json:"orders"`; Repeat int64 `json:"repeat"`; Rate float64 `json:"rate"` }
		var procs []proc
		db.Raw(`SELECT oi.item_name as name, COUNT(DISTINCT o.customer_id) as orders,
			COUNT(DISTINCT CASE WHEN sub.cnt>=2 THEN o.customer_id END) as repeat,
			CASE WHEN COUNT(DISTINCT o.customer_id)>0 THEN COUNT(DISTINCT CASE WHEN sub.cnt>=2 THEN o.customer_id END)*100.0/COUNT(DISTINCT o.customer_id) ELSE 0 END as rate
			FROM order_items oi JOIN orders o ON oi.order_id=o.id AND o.status='paid'
			LEFT JOIN (SELECT customer_id,COUNT(*) as cnt FROM orders WHERE status='paid' GROUP BY customer_id) sub ON o.customer_id=sub.customer_id
			GROUP BY oi.item_name ORDER BY orders DESC`).Scan(&procs)
		c.JSON(http.StatusOK, gin.H{"overall": ov, "by_procedure": procs})
	}
}

// Utilization returns doctor utilization and no-show by time slot.
func Utilization(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type doc struct{ Name string `json:"name"`; Total int64 `json:"total"`; Done int64 `json:"done"`; Rate float64 `json:"rate"` }
		var docs []doc
		db.Raw(`SELECT COALESCE(doctor_name,'未知') as name, COUNT(*) as total,
			SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END) as done,
			CASE WHEN COUNT(*)>0 THEN SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END)*100.0/COUNT(*) ELSE 0 END as rate
			FROM appointments WHERE date>=NOW()-INTERVAL'30 days' GROUP BY doctor_name ORDER BY total DESC`).Scan(&docs)
		type slot struct{ TimeSlot string `json:"time_slot"`; Total int64 `json:"total"`; Noshow int64 `json:"noshow"`; Rate float64 `json:"rate"` }
		var slots []slot
		db.Raw(`SELECT time_slot, COUNT(*) as total,
			SUM(CASE WHEN status='cancelled' THEN 1 ELSE 0 END) as noshow,
			AVG(CASE WHEN status='cancelled' THEN 100.0 ELSE 0 END) as rate
			FROM appointments GROUP BY time_slot ORDER BY total DESC`).Scan(&slots)
		c.JSON(http.StatusOK, gin.H{"doctors": docs, "time_slots": slots})
	}
}
