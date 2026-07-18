package handler

import (
	authHandler "clinic-mgmt/internal/auth/handler"
	authService "clinic-mgmt/internal/auth/service"
	"clinic-mgmt/internal/appointment"
	"clinic-mgmt/internal/config"
	"clinic-mgmt/internal/customer"
	"clinic-mgmt/internal/expense"
	"clinic-mgmt/internal/membership"
	"clinic-mgmt/internal/middleware"
	"clinic-mgmt/internal/order"
	pkgHandler "clinic-mgmt/internal/package"
	"clinic-mgmt/internal/backup"
	"clinic-mgmt/internal/report"
	"clinic-mgmt/internal/settings"
	"clinic-mgmt/internal/system"
	"clinic-mgmt/internal/treatment"
	"clinic-mgmt/internal/followup"
	"clinic-mgmt/internal/inventory"
	"clinic-mgmt/internal/license"
	"clinic-mgmt/internal/photo"
	"clinic-mgmt/internal/commission"
	"clinic-mgmt/internal/kpi"
	"clinic-mgmt/internal/training"
	"clinic-mgmt/internal/analysis"
	"clinic-mgmt/internal/consent"
	"clinic-mgmt/internal/label"
	"clinic-mgmt/internal/document"
	"clinic-mgmt/internal/marketing"
	"clinic-mgmt/internal/medical"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	authService.InitJWT(cfg)

	api := r.Group("/api/v1")

	// Public routes
	api.POST("/auth/login", authHandler.Login(db))

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// Profile
		protected.GET("/auth/profile", authHandler.GetProfile(db))

		// Users & Roles (admin only)
	protected.GET("/users", system.ListUsers(db))
	protected.POST("/users", middleware.RequirePermission("admin"), system.CreateUser(db))
		protected.PUT("/users/:id", middleware.RequirePermission("admin"), system.UpdateUser(db))
		protected.PUT("/users/:id/status", middleware.RequirePermission("admin"), system.UpdateUserStatus(db))

		protected.GET("/roles", middleware.RequirePermission("admin"), system.ListRoles(db))
		protected.POST("/roles", middleware.RequirePermission("admin"), system.CreateRole(db))
		protected.PUT("/roles/:id", middleware.RequirePermission("admin"), system.UpdateRole(db))
		protected.DELETE("/roles/:id", middleware.RequirePermission("admin"), system.DeleteRole(db))

		// Customers
		protected.GET("/customers", customer.ListCustomers(db))
		protected.POST("/customers", customer.CreateCustomer(db))
		protected.GET("/customers/:id", customer.GetCustomer(db))
		protected.PUT("/customers/:id", customer.UpdateCustomer(db))
		protected.DELETE("/customers/:id", customer.DeleteCustomer(db))

		// Customer follow-ups
		protected.GET("/customers/:id/followups", customer.ListFollowUps(db))
		protected.POST("/customers/:id/followups", customer.CreateFollowUp(db))

		// Membership
		protected.POST("/customers/:id/membership", membership.CreateMembership(db))
		protected.GET("/customers/:id/membership", membership.GetMembership(db))

		// Recharge
		protected.POST("/customers/:id/recharge", membership.Recharge(db))

		// Customer packages
		protected.GET("/customers/:id/packages", pkgHandler.ListPackages(db))
		protected.POST("/customers/:id/packages", pkgHandler.CreatePackage(db))

		// Appointments
		protected.GET("/appointments", appointment.ListAppointments(db))
		protected.POST("/appointments", appointment.CreateAppointment(db))
		protected.PUT("/appointments/:id/checkin", appointment.CheckIn(db))
		protected.PUT("/appointments/:id/complete", appointment.Complete(db))
		protected.PUT("/appointments/:id/cancel", appointment.Cancel(db))

		// Orders
		protected.POST("/orders", order.CreateOrder(db))
		protected.GET("/orders", order.ListOrders(db))
		protected.GET("/orders/:id", order.GetOrder(db))
		protected.POST("/orders/:id/pay", order.PayOrder(db))
		protected.POST("/orders/:id/refund", order.RefundOrder(db))

		// Package redemption (POS payment via package)
		protected.POST("/orders/:id/pay/package", pkgHandler.RedeemPackage(db))

		// Settings
		protected.GET("/settings/items", settings.ListItems(db))
		protected.POST("/settings/items", settings.CreateItem(db))
		protected.PUT("/settings/items/:id", settings.UpdateItem(db))
		protected.GET("/settings/package-templates", settings.ListPackageTemplates(db))
		protected.POST("/settings/package-templates", settings.CreatePackageTemplate(db))
		protected.PUT("/settings/package-templates/:id", settings.UpdatePackageTemplate(db))

		// Reports
		protected.GET("/reports/profit", middleware.RequirePermission("admin"), report.ProfitReport(db))
		protected.GET("/reports/analytics", middleware.RequirePermission("admin"), report.BusinessAnalytics(db))
		protected.GET("/reports/daily-sales", middleware.RequirePermission("admin"), report.DailySales(db))
		protected.GET("/reports/member-summary", middleware.RequirePermission("admin"), report.GetMemberSummary(db))
		protected.GET("/reports/period", middleware.RequirePermission("admin"), report.PeriodReport(db))
		protected.GET("/reports/source-roi", middleware.RequirePermission("admin"), report.SourceROI(db))
		protected.GET("/reports/payment-methods", middleware.RequirePermission("admin"), report.PaymentMethods(db))
		protected.GET("/reports/insights", middleware.RequirePermission("admin"), report.Insights(db))
		protected.GET("/reports/customer-trend", middleware.RequirePermission("admin"), report.CustomerTrend(db))

		// Expenses
		protected.GET("/expenses", expense.ListExpenses(db))
		protected.POST("/expenses", expense.CreateExpense(db))
		protected.PUT("/expenses/:id", expense.UpdateExpense(db))
		protected.DELETE("/expenses/:id", expense.DeleteExpense(db))

		// Backup
		protected.GET("/backup/list", middleware.RequirePermission("admin"), backup.ListBackups(db))
		protected.POST("/backup/create", middleware.RequirePermission("admin"), backup.CreateBackupHandler(db, cfg))
		protected.DELETE("/backup/:id", middleware.RequirePermission("admin"), backup.DeleteBackupHandler(db))
		protected.GET("/backup/:id/download", middleware.RequirePermission("admin"), backup.DownloadBackup(db))
		protected.POST("/backup/:id/upload-cloud", middleware.RequirePermission("admin"), backup.UploadToCloudHandler(db))
		protected.GET("/backup/settings", middleware.RequirePermission("admin"), backup.GetBackupSettings(db))
		protected.PUT("/backup/settings", middleware.RequirePermission("admin"), backup.SaveBackupSettings(db))
		protected.GET("/backup/export", middleware.RequirePermission("admin"), backup.ExportBackup(db, cfg))
		protected.POST("/backup/reset", middleware.RequirePermission("admin"), backup.ResetSystem(db))
		protected.POST("/backup/import", middleware.RequirePermission("admin"), backup.ImportBackup(db, cfg))

		// Audit logs
		protected.GET("/audit-logs", middleware.RequirePermission("system:audit"), system.ListAuditLogs(db))

		// Treatment records
		protected.POST("/treatment/records", treatment.CreateRecord(db))
		protected.GET("/treatment/records/:id", treatment.GetRecord(db))
		protected.PUT("/treatment/records/:id", treatment.UpdateRecord(db))
		protected.GET("/customers/:id/treatments", treatment.ListRecords(db))

		// Follow-up
		protected.GET("/followup/tasks", followup.ListTasks(db))
		protected.POST("/followup/tasks", followup.CreateTask(db))
		protected.PUT("/followup/tasks/:id/complete", followup.CompleteTask(db))
		protected.POST("/followup/tasks/auto-generate", followup.AutoGenerate(db))
		protected.POST("/followup/tasks/auto-generate-post-treatment", followup.AutoGeneratePostTreatment(db))
		protected.GET("/followup/tasks/stats", followup.Stats(db))

		// Inventory
		protected.GET("/inventory/items", inventory.ListItems(db))
		protected.POST("/inventory/items", inventory.CreateItem(db))
		protected.PUT("/inventory/items/:id", inventory.UpdateItem(db))
		protected.DELETE("/inventory/items/:id", inventory.DeleteItem(db))
		protected.POST("/inventory/items/:id/stock-in", inventory.StockIn(db))
		protected.POST("/inventory/items/:id/stock-out", inventory.StockOut(db))
		protected.GET("/inventory/items/:id/logs", inventory.ListLogs(db))



		// Documents
		protected.GET("/documents", document.List(db))
		protected.POST("/documents", document.Create(db))
		protected.GET("/documents/expiring", document.Expiring(db))
		protected.GET("/documents/:id", document.Get(db))
		protected.DELETE("/documents/:id", document.Delete(db))
		protected.GET("/documents/:id/download", document.Download(db))

		// Photos
		protected.GET("/photos", photo.List(db))
		protected.POST("/photos", photo.Upload(db))
		api.GET("/photos/:id/download", photo.Download(db))
		protected.DELETE("/photos/:id", photo.Delete(db))

		// Tag Rules
		protected.GET("/tag-rules", label.ListRules(db))
		protected.POST("/tag-rules", label.CreateRule(db))
		protected.PUT("/tag-rules/:id", label.UpdateRule(db))
		protected.DELETE("/tag-rules/:id", label.DeleteRule(db))
		protected.POST("/tag-rules/apply", label.ApplyRules(db))
		// Marketing
		protected.GET("/marketing/dormant", marketing.DormantCustomers(db))
		protected.GET("/marketing/birthday", marketing.BirthdayCustomers(db))
		protected.GET("/marketing/near-dormant", marketing.NearDormantCustomers(db))

		// Medical templates
		protected.POST("/medical/records/:id/sign", medical.SignRecord(db))

		// Medical records
		protected.POST("/medical/records", medical.CreateRecord(db))
		protected.GET("/medical/records", medical.ListAll(db))
		protected.GET("/medical/records/:id", medical.GetRecord(db))
		protected.PUT("/medical/records/:id", medical.UpdateRecord(db))
		protected.DELETE("/medical/records/:id", medical.DeleteRecord(db))
		protected.GET("/customers/:id/medical-records", medical.ListRecords(db))
		// Commission
		protected.GET("/commission/rules", commission.ListRules(db))
		protected.POST("/commission/rules", commission.CreateRule(db))
		protected.PUT("/commission/rules/:id", commission.UpdateRule(db))
		protected.DELETE("/commission/rules/:id", commission.DeleteRule(db))
		protected.POST("/commission/calculate", commission.Calculate(db))
		protected.GET("/commission/results", commission.ListResults(db))
		protected.PUT("/commission/results/:id/confirm", commission.ConfirmResult(db))

		// KPI
		protected.GET("/kpi/targets", kpi.ListTargets(db))
		protected.POST("/kpi/targets", kpi.SaveTarget(db))
		protected.GET("/kpi/leaderboard", kpi.Leaderboard(db))

		// Training
		protected.GET("/training", training.List(db))
		protected.POST("/training", training.Create(db))
		protected.PUT("/training/:id", training.Update(db))
		protected.DELETE("/training/:id", training.Delete(db))
		protected.GET("/training/stats", training.Stats(db))
		protected.GET("/training/staff-stats", training.PerStaffStats(db))
		protected.GET("/training/category-stats", training.CategoryStats(db))
		protected.GET("/training/monthly-stats", training.MonthlyStats(db))
		protected.GET("/training/plans", training.ListPlans(db))
		protected.POST("/training/plans", training.CreatePlan(db))
		protected.PUT("/training/plans/:id", training.UpdatePlan(db))
		protected.DELETE("/training/plans/:id", training.DeletePlan(db))

		// Analysis
		protected.GET("/analysis/ltv", analysis.LTV(db))
		protected.GET("/analysis/cross-sell", analysis.CrossSell(db))
		protected.GET("/analysis/no-show", analysis.NoShow(db))
		protected.GET("/analysis/churn", analysis.Churn(db))

		// Consent Forms
		protected.GET("/consent", consent.List(db))
		protected.POST("/consent", consent.Create(db))
		protected.GET("/consent/:id", consent.Get(db))
		protected.PUT("/consent/:id", consent.Update(db))
		protected.DELETE("/consent/:id", consent.Delete(db))


		// Analysis - Profit
		protected.GET("/analysis/procedure-profit", analysis.ProcedureProfit(db))
		protected.GET("/analysis/monthly-trend", analysis.MonthlyTrend(db))
		protected.GET("/analysis/revenue-structure", analysis.RevenueStructure(db))
		protected.GET("/analysis/repurchase", analysis.Repurchase(db))
		protected.GET("/analysis/utilization", analysis.Utilization(db))


		// License
		protected.GET("/license/status", license.StatusHandler(db))
		protected.POST("/license/activate", license.ActivateHandler(db))

	}
}
