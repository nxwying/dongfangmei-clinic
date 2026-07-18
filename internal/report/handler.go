package report

import (
	"net/http"
	"time"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DailySalesResponse struct {
	Date          string  `json:"date"`
	OrderCount    int64   `json:"order_count"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalDiscount float64 `json:"total_discount"`
	RechargeTotal float64 `json:"recharge_total"`
}

func DailySales(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
		var resp DailySalesResponse
		resp.Date = date

		db.Raw(`SELECT COUNT(*) as order_count, COALESCE(SUM(final_amount),0) as total_revenue, COALESCE(SUM(discount_amount),0) as total_discount FROM orders WHERE status = 'paid' AND DATE(created_at) = ?`, date).Scan(&resp)
		db.Raw(`SELECT COALESCE(SUM(amount),0) as recharge_total FROM payments WHERE pay_method = 'recharge' AND DATE(created_at) = ?`, date).Scan(&resp)

		c.JSON(http.StatusOK, resp)
	}
}

type MemberSummary struct {
	TotalMembers   int64   `json:"total_members"`
	NewThisMonth   int64   `json:"new_this_month"`
	TotalRecharged float64 `json:"total_recharged"`
	TotalConsumed  float64 `json:"total_consumed"`
}

func ProfitReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

		var totalRevenue float64
		db.Raw("SELECT COALESCE(SUM(final_amount),0) FROM orders WHERE status='paid' AND DATE(created_at)=?", date).Scan(&totalRevenue)

		var highValue float64
		db.Raw("SELECT COALESCE(SUM(amount),0) FROM expenses WHERE type='high_value' AND date=?", date).Scan(&highValue)
		var orderCost float64
		db.Raw("SELECT COALESCE(SUM(cost_amount),0) FROM orders WHERE status='paid' AND DATE(created_at)=?", date).Scan(&orderCost)
		highValue += orderCost

		var general float64
		db.Raw("SELECT COALESCE(SUM(amount),0) FROM expenses WHERE type='general' AND date=?", date).Scan(&general)

		var commissions float64
		db.Raw("SELECT COALESCE(SUM(amount),0) FROM expenses WHERE type='commission' AND date=?", date).Scan(&commissions)
		var orderComm float64
		db.Raw("SELECT COALESCE(SUM(commission_amount),0) FROM orders WHERE status='paid' AND DATE(created_at)=?", date).Scan(&orderComm)
		commissions += orderComm

		totalExpense := highValue + general
		grossRevenue := totalRevenue - commissions
		netProfit := grossRevenue - totalExpense

		c.JSON(http.StatusOK, gin.H{
			"date":          date,
			"total_revenue": totalRevenue,
			"commissions":   commissions,
			"gross_revenue": grossRevenue,
			"high_value":    highValue,
			"general":       general,
			"total_expense": totalExpense,
			"net_profit":    netProfit,
		})
	}
}

func GetMemberSummary(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var summary MemberSummary
		db.Model(&model.Membership{}).Count(&summary.TotalMembers)
		monthStart := time.Now().Format("2006-01") + "-01"
		db.Model(&model.Membership{}).Where("created_at >= ?", monthStart).Count(&summary.NewThisMonth)
		db.Model(&model.Membership{}).Select("COALESCE(SUM(total_recharged),0) as total_recharged, COALESCE(SUM(total_consumed),0) as total_consumed").Scan(&summary)
		c.JSON(http.StatusOK, summary)
	}
}

// BusinessAnalytics returns comprehensive analytics data for all 7 reports.
func BusinessAnalytics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Monthly trend (last 12 months)
		type MonthRow struct {
			Month       string  `json:"month"`
			Revenue     float64 `json:"revenue"`
			HighValue   float64 `json:"high_value"`
			General     float64 `json:"general"`
			Commission  float64 `json:"commission"`
			NetProfit   float64 `json:"net_profit"`
		}
		var monthly []MonthRow
		db.Raw(`SELECT m.month,
			COALESCE(SUM(o.final_amount),0) as revenue,
			0 as high_value, 0 as general, 0 as commission, 0 as net_profit
			FROM generate_series(date_trunc('month',now())-interval'11 months',date_trunc('month',now()),interval'1 month') m(month)
			LEFT JOIN orders o ON date_trunc('month',o.created_at)=m.month AND o.status='paid'
			GROUP BY m.month ORDER BY m.month`).Scan(&monthly)

		// Fill expense data into monthly
		type ExpRow struct { Month string; HighValue float64; General float64; Commission float64 }
		var expRows []ExpRow
		db.Raw(`SELECT to_char(date_trunc('month',date::date),'YYYY-MM') as month,
			COALESCE(SUM(amount) FILTER(WHERE type='high_value'),0) as high_value,
			COALESCE(SUM(amount) FILTER(WHERE type='general'),0) as general,
			COALESCE(SUM(amount) FILTER(WHERE type='commission'),0) as commission
			FROM expenses WHERE date>=to_char(date_trunc('month',now())-interval'11 months','YYYY-MM-DD')
			GROUP BY month ORDER BY month`).Scan(&expRows)
		expMap := map[string]ExpRow{}
		for _, r := range expRows { expMap[r.Month] = r }
		for i, m := range monthly {
			if e, ok := expMap[m.Month]; ok {
				monthly[i].HighValue = e.HighValue
				monthly[i].General = e.General
				monthly[i].Commission = e.Commission
			}
			monthly[i].NetProfit = m.Revenue - monthly[i].HighValue - monthly[i].General - monthly[i].Commission
		}

		// 2. Top selling items
		var topItems []struct {
			ItemName string  `json:"item_name"`
			Count    int64   `json:"count"`
			Total    float64 `json:"total"`
		}
		db.Raw(`SELECT oi.item_name, SUM(oi.quantity) as count, COALESCE(SUM(oi.subtotal),0) as total
			FROM order_items oi JOIN orders o ON oi.order_id=o.id
			WHERE o.status='paid'
			GROUP BY oi.item_name ORDER BY total DESC LIMIT 10`).Scan(&topItems)

		// 3. Customer sources
		var sources []struct {
			Source string `json:"source"`
			Count  int64  `json:"count"`
		}
		db.Raw(`SELECT COALESCE(NULLIF(source,''),'其他') as source, COUNT(*) as count
			FROM customers GROUP BY source ORDER BY count DESC`).Scan(&sources)

		// 4. Staff performance
		var staffPerf []struct {
			UserID    uint    `json:"user_id"`
			RealName  string  `json:"real_name"`
			Orders    int64   `json:"orders"`
			Total     float64 `json:"total"`
		}
		db.Raw(`SELECT o.created_by as user_id, COALESCE(u.real_name,'未知') as real_name,
			COUNT(*) as orders, COALESCE(SUM(o.final_amount),0) as total
			FROM orders o LEFT JOIN users u ON o.created_by=u.id
			WHERE o.status='paid' GROUP BY o.created_by, u.real_name ORDER BY total DESC`).Scan(&staffPerf)

		// 5. Package usage
		var pkgStats []struct {
			Status string `json:"status"`
			Count  int64  `json:"count"`
		}
		db.Model(&model.MemberPackage{}).
			Select("status, COUNT(*) as count").Group("status").Scan(&pkgStats)

		// 6. Average order value & customer lifetime
		var avgData struct {
			AvgOrder    float64 `json:"avg_order"`
			TotalOrders int64   `json:"total_orders"`
			TotalRev    float64 `json:"total_revenue"`
		}
		db.Raw(`SELECT COUNT(*) as total_orders, COALESCE(AVG(final_amount),0) as avg_order,
			COALESCE(SUM(final_amount),0) as total_revenue FROM orders WHERE status='paid'`).Scan(&avgData)

		// 7. Appointment funnel
		var apptFunnel []struct {
			Status string `json:"status"`
			Count  int64  `json:"count"`
		}
		db.Model(&model.Appointment{}).
			Select("status, COUNT(*) as count").Group("status").Scan(&apptFunnel)

		c.JSON(http.StatusOK, gin.H{
			"monthly_trend":    monthly,
			"top_items":        topItems,
			"customer_sources": sources,
			"staff_perf":       staffPerf,
			"package_stats":    pkgStats,
			"avg_order":        avgData,
			"appointment_funnel": apptFunnel,
		})
	}
}

// PeriodReport returns aggregated data for a custom date range.
func PeriodReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
		endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

		type periodTotals struct {
			TotalRevenue float64
			OrderCount   int64
		}
		var pt periodTotals
		db.Raw("SELECT COALESCE(SUM(final_amount),0) as total_revenue, COUNT(*) as order_count FROM orders WHERE status='paid' AND DATE(created_at) >= ? AND DATE(created_at) <= ?", startDate, endDate).Scan(&pt)
		totalRevenue := pt.TotalRevenue
		orderCount := pt.OrderCount

		type expTotals struct {
			Commissions float64
			HighValue   float64
			General     float64
		}
		var et expTotals
		db.Raw("SELECT COALESCE(SUM(amount) FILTER(WHERE type='commission'),0) as commissions, COALESCE(SUM(amount) FILTER(WHERE type='high_value'),0) as high_value, COALESCE(SUM(amount) FILTER(WHERE type='general'),0) as general FROM expenses WHERE date >= ? AND date <= ?", startDate, endDate).Scan(&et)
		commissions := et.Commissions
		highValue := et.HighValue
		general := et.General

		// Add order commission_amount to commissions
		var orderComm float64
		db.Raw("SELECT COALESCE(SUM(commission_amount),0) FROM orders WHERE status='paid' AND DATE(created_at) >= ? AND DATE(created_at) <= ?", startDate, endDate).Scan(&orderComm)
		commissions += orderComm

		// Add order cost_amount to high_value
		var orderCost float64
		db.Raw("SELECT COALESCE(SUM(cost_amount),0) FROM orders WHERE status='paid' AND DATE(created_at) >= ? AND DATE(created_at) <= ?", startDate, endDate).Scan(&orderCost)
		highValue += orderCost

		grossRevenue := totalRevenue - commissions
		netProfit := grossRevenue - highValue - general

		type DailyRow struct {
			Date        string  `json:"date"`
			Revenue     float64 `json:"revenue"`
			Discount    float64 `json:"discount"`
			Gross       float64 `json:"gross"`
			Commissions float64 `json:"commissions"`
			Cost        float64 `json:"cost"`
			NetProfit   float64 `json:"net_profit"`
			Orders      int64   `json:"orders"`
		}
		var daily []DailyRow
		db.Raw(`SELECT d::date as date,
			COALESCE(o.revenue,0) as revenue, COALESCE(o.discount,0) as discount,
			COALESCE(o.revenue,0)-COALESCE(o.commissions,0)-COALESCE(e.commissions,0) as gross,
			COALESCE(o.commissions,0)+COALESCE(e.commissions,0) as commissions,
			COALESCE(o.cost,0)+COALESCE(e.cost,0) as cost,
			COALESCE(o.revenue,0)-COALESCE(o.commissions,0)-COALESCE(o.cost,0)-COALESCE(e.commissions,0)-COALESCE(e.cost,0) as net_profit,
			COALESCE(o.orders,0) as orders
			FROM generate_series(?::date,?::date,'1 day') d
			LEFT JOIN (SELECT DATE(created_at) dt,COALESCE(SUM(final_amount),0) revenue,COALESCE(SUM(discount_amount),0) discount,COALESCE(SUM(commission_amount),0) commissions,COALESCE(SUM(cost_amount),0) cost,COUNT(*) orders FROM orders WHERE status='paid' GROUP BY DATE(created_at)) o ON d=o.dt
			LEFT JOIN (SELECT date,COALESCE(SUM(amount) FILTER(WHERE type='commission'),0) commissions,COALESCE(SUM(amount) FILTER(WHERE type='high_value' OR type='general'),0) cost FROM expenses WHERE date>=? AND date<=? GROUP BY date) e ON d::text=e.date
			ORDER BY d`, startDate, endDate, startDate, endDate).Scan(&daily)

		type TopItem struct {
			Name  string  `json:"name"`
			Count int64   `json:"count"`
			Total float64 `json:"total"`
			Gross float64 `json:"gross"`
		}
		var topItems []TopItem
		db.Raw(`SELECT oi.item_name as name, SUM(oi.quantity) as count, COALESCE(SUM(oi.subtotal),0) as total,
			COALESCE(SUM(oi.subtotal - (COALESCE(o.commission_amount,0) * oi.subtotal / NULLIF(o.final_amount,0))),0) as gross
			FROM order_items oi JOIN orders o ON oi.order_id=o.id
			WHERE o.status='paid' AND oi.created_at>=? AND oi.created_at<=?
			GROUP BY oi.item_name ORDER BY gross DESC LIMIT 10`, startDate+" 00:00:00", endDate+" 23:59:59").Scan(&topItems)

		c.JSON(http.StatusOK, gin.H{
			"date_from":      startDate,
			"date_to":        endDate,
			"total_revenue":  totalRevenue,
			"commissions":    commissions,
			"high_value":     highValue,
			"general":        general,
			"gross_revenue":  grossRevenue,
			"net_profit":     netProfit,
			"order_count":    orderCount,
			"daily_breakdown": daily,
			"top_items":      topItems,
		})
	}
}

// SourceROI returns customer source analysis.
func SourceROI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
		endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

		type SourceRow struct {
			Source         string  `json:"source"`
			CustomerCount  int64   `json:"customer_count"`
			OrderCount     int64   `json:"order_count"`
			TotalRevenue   float64 `json:"total_revenue"`
			AvgOrderValue  float64 `json:"avg_order_value"`
		}
		var rows []SourceRow
		db.Raw(`SELECT
			CASE
				WHEN LOWER(TRIM(c.source)) IN ('xiaohongshu','小红书','小红書') THEN 'xiaohongshu'
				WHEN LOWER(TRIM(c.source)) IN ('wechat','微信','wx') THEN 'wechat'
				WHEN LOWER(TRIM(c.source)) IN ('referral','转介绍','介绍') THEN 'referral'
				WHEN LOWER(TRIM(c.source)) IN ('walk_in','到店','上门','自然到店') THEN 'walk_in'
				WHEN LOWER(TRIM(c.source)) IN ('douyin','抖音','dy','tik tok') THEN 'douyin'
				WHEN LOWER(TRIM(c.source)) IN ('dianping','大众点评','点评') THEN 'dianping'
				WHEN LOWER(TRIM(c.source)) IN ('美团','meituan') THEN 'meituan'
				WHEN LOWER(TRIM(c.source)) IN ('direct','直客','zhike') THEN 'direct'
				WHEN LOWER(TRIM(c.source)) IN ('referral','转介绍','介绍') THEN 'referral'
				WHEN LOWER(TRIM(c.source)) IN ('老带新','朋友介绍','old_customer_ref') THEN 'old_customer_ref'
				ELSE COALESCE(NULLIF(TRIM(c.source),''),'其他')
			END as source,
			COUNT(DISTINCT c.id) as customer_count,
			COUNT(o.id) as order_count,
			COALESCE(SUM(o.final_amount),0) as total_revenue,
			CASE WHEN COUNT(o.id)>0 THEN COALESCE(SUM(o.final_amount),0)/COUNT(o.id) ELSE 0 END as avg_order_value
			FROM customers c
			LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
				AND DATE(o.created_at)>=? AND DATE(o.created_at)<=?
			GROUP BY 1
			ORDER BY total_revenue DESC`, startDate, endDate).Scan(&rows)

		c.JSON(http.StatusOK, rows)
	}
}

// Insights returns repurchase, staff perf, and package vs single analysis.
func Insights(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
		endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

		// 1. Repurchase
		type RepData struct {
			TotalCustomers  int64   `json:"total_customers"`
			RepeatCustomers int64   `json:"repeat_customers"`
			RepeatRate   float64 `json:"repeat_rate"`
			SingleOrder  int64   `json:"single_order"`
			TwoOrders    int64   `json:"two_orders"`
			ThreePlusOrd int64   `json:"three_plus_orders"`
		}
		var rep RepData
		db.Raw(`SELECT
			COUNT(*) as total_customers,
			SUM(CASE WHEN sub.order_count>=2 THEN 1 ELSE 0 END) as repeat_customers,
			SUM(CASE WHEN sub.order_count=1 THEN 1 ELSE 0 END) as single_order,
			SUM(CASE WHEN sub.order_count=2 THEN 1 ELSE 0 END) as two_orders,
			SUM(CASE WHEN sub.order_count>=3 THEN 1 ELSE 0 END) as three_plus_orders
			FROM (SELECT c.id,COUNT(o.id) as order_count FROM customers c
				LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
				AND DATE(o.created_at)>=? AND DATE(o.created_at)<=?
				GROUP BY c.id) sub`, startDate, endDate).Scan(&rep)
		if rep.TotalCustomers > 0 {
			rep.RepeatRate = float64(rep.RepeatCustomers) / float64(rep.TotalCustomers) * 100
		}

		// 2. Staff perf
		type StaffRow struct {
			RealName string  `json:"real_name"`
			Orders   int64   `json:"orders"`
			Total    float64 `json:"total"`
			Gross    float64 `json:"gross"`
		}
		var staff []StaffRow
		db.Raw(`SELECT COALESCE(u.real_name,'未知') as real_name,
			COUNT(*) as orders, COALESCE(SUM(o.final_amount),0) as total,
			COALESCE(SUM(o.final_amount - COALESCE(o.commission_amount,0)),0) as gross
			FROM orders o LEFT JOIN users u ON o.created_by=u.id
			WHERE o.status='paid' AND DATE(o.created_at)>=? AND DATE(o.created_at)<=?
			GROUP BY o.created_by,u.real_name ORDER BY total DESC`, startDate, endDate).Scan(&staff)

		// 3. Package vs single
		type PSRow struct {
			Type       string  `json:"type"`
			CustCount  int64   `json:"customer_count"`
			OrdCount   int64   `json:"order_count"`
			TotRev     float64 `json:"total_revenue"`
			AvgCust    float64 `json:"avg_per_customer"`
		}
		var results []PSRow

		type pkRes struct { CustomerCount int64; OrderCount int64; TotalRevenue float64 }
		var pkg pkRes
		db.Raw(`SELECT COALESCE(COUNT(DISTINCT mp.customer_id),0) as customer_count,COUNT(o.id) as order_count,COALESCE(SUM(o.final_amount),0) as total_revenue
			FROM member_packages mp LEFT JOIN orders o ON mp.customer_id=o.customer_id
			AND o.status='paid' AND DATE(o.created_at)>=? AND DATE(o.created_at)<=?`, startDate, endDate).Scan(&pkg)
		var ps1 PSRow
		ps1.Type = "package"; ps1.CustCount = pkg.CustomerCount; ps1.OrdCount = pkg.OrderCount; ps1.TotRev = pkg.TotalRevenue
		if ps1.CustCount > 0 { ps1.AvgCust = ps1.TotRev / float64(ps1.CustCount) }
		results = append(results, ps1)

		type skRes struct { CustomerCount int64; OrderCount int64; TotalRevenue float64 }
		var single skRes
		db.Raw(`SELECT COALESCE(COUNT(DISTINCT c.id),0) as customer_count,COUNT(o.id) as order_count,COALESCE(SUM(o.final_amount),0) as total_revenue
			FROM customers c INNER JOIN orders o ON c.id=o.customer_id AND o.status='paid'
			AND DATE(o.created_at)>=? AND DATE(o.created_at)<=?
			WHERE c.id NOT IN (SELECT DISTINCT customer_id FROM member_packages)`, startDate, endDate).Scan(&single)
		var ps2 PSRow
		ps2.Type = "single"; ps2.CustCount = single.CustomerCount; ps2.OrdCount = single.OrderCount; ps2.TotRev = single.TotalRevenue
		if ps2.CustCount > 0 { ps2.AvgCust = ps2.TotRev / float64(ps2.CustCount) }
		results = append(results, ps2)

		c.JSON(http.StatusOK, gin.H{"repurchase": rep, "staff_perf": staff, "package_vs_single": results})
	}
}

// PaymentMethods returns payment method breakdown.
func PaymentMethods(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
		endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

		type PMRow struct {
			Method string  `json:"method"`
			Count  int64   `json:"count"`
			Total  float64 `json:"total"`
			Gross  float64 `json:"gross"`
		}
		var rows []PMRow
		db.Raw(`SELECT p.pay_method as method, COUNT(*) as count,
			COALESCE(SUM(p.amount),0) as total,
			COALESCE(SUM(p.amount - (COALESCE(o.commission_amount,0) * p.amount / NULLIF(o.final_amount,0))),0) as gross
			FROM payments p LEFT JOIN orders o ON p.order_id=o.id
			WHERE DATE(p.created_at)>=? AND DATE(p.created_at)<=?
			GROUP BY p.pay_method
			ORDER BY total DESC`, startDate, endDate).Scan(&rows)

		c.JSON(http.StatusOK, rows)
	}
}

// CustomerTrend returns monthly new customer and conversion data.
func CustomerTrend(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type Row struct {
			Month             string  `json:"month"`
			NewCustomers      int64   `json:"new_customers"`
			ConvertedCustomers int64  `json:"converted_customers"`
		}
		var rows []Row
		db.Raw(`SELECT to_char(DATE_TRUNC('month',c.created_at),'YYYY-MM') as month,
			COUNT(DISTINCT c.id) as new_customers,
			COUNT(DISTINCT o.customer_id) as converted_customers
			FROM customers c
			LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
			WHERE c.created_at >= DATE_TRUNC('month',NOW())-INTERVAL '11 months'
			GROUP BY DATE_TRUNC('month',c.created_at)
			ORDER BY month`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}
