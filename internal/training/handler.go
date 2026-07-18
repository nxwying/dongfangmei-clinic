package training

import (
	"net/http"
	"strconv"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== Training CRUD ==========

func List(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []model.Training
		q := db.Order("date DESC")
		if uid := c.Query("user_id"); uid != "" {
			q = q.Where("user_id=?", uid)
		}
		if cat := c.Query("category"); cat != "" {
			q = q.Where("category=?", cat)
		}
		if passed := c.Query("passed"); passed != "" {
			q = q.Where("passed=?", passed)
		}
		if mandatory := c.Query("is_mandatory"); mandatory != "" {
			q = q.Where("is_mandatory=?", mandatory == "true")
		}
		if certExpiring := c.Query("cert_expiring"); certExpiring == "true" {
			q = q.Where("cert_expiry != '' AND cert_expiry >= CURRENT_DATE AND cert_expiry <= CURRENT_DATE + INTERVAL '30 days'")
		}
		if certExpired := c.Query("cert_expired"); certExpired == "true" {
			q = q.Where("cert_expiry != '' AND cert_expiry < CURRENT_DATE")
		}
		q.Find(&list)
		c.JSON(http.StatusOK, list)
	}
}

func Create(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t model.Training
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
			return
		}
		if err := db.Create(&t).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}
		c.JSON(http.StatusCreated, t)
	}
}

func Update(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var t model.Training
		if err := db.First(&t, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		var input model.Training
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		// Don't allow changing ID
		input.ID = uint(id)
		if err := db.Model(&t).Updates(input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
		c.JSON(http.StatusOK, input)
	}
}

func Delete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.Training{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}

// ========== Stats ==========

func Stats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type statsResponse struct {
			Total    int64   `json:"total"`
			Hours    float64 `json:"hours"`
			Staff    int64   `json:"staff"`
			Expiring int64   `json:"expiring"`
			Expired  int64   `json:"expired"`
			Points   int64   `json:"points"`
			Cost     float64 `json:"cost"`
			PassRate float64 `json:"pass_rate"`
		}
		var resp statsResponse
		db.Model(&model.Training{}).Select("COUNT(*)").Scan(&resp.Total)
		db.Model(&model.Training{}).Select("COALESCE(SUM(hours),0)").Scan(&resp.Hours)
		db.Model(&model.Training{}).Select("COUNT(DISTINCT user_id)").Scan(&resp.Staff)
		db.Model(&model.Training{}).Where("cert_expiry != '' AND cert_expiry >= CURRENT_DATE AND cert_expiry <= CURRENT_DATE + INTERVAL '30 days'").Count(&resp.Expiring)
		db.Model(&model.Training{}).Where("cert_expiry != '' AND cert_expiry < CURRENT_DATE").Count(&resp.Expired)
		db.Model(&model.Training{}).Select("COALESCE(SUM(points),0)").Scan(&resp.Points)
		db.Model(&model.Training{}).Select("COALESCE(SUM(cost),0)").Scan(&resp.Cost)
		var passed, total int64
		db.Model(&model.Training{}).Where("passed = 'passed'").Count(&passed)
		db.Model(&model.Training{}).Where("passed IN ('passed','failed')").Count(&total)
		if total > 0 {
			resp.PassRate = float64(passed) / float64(total) * 100
		}
		c.JSON(http.StatusOK, resp)
	}
}

func PerStaffStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type stat struct {
			UserID        uint    `json:"user_id"`
			RealName      string  `json:"real_name"`
			TotalHours    float64 `json:"total_hours"`
			TotalSessions int64   `json:"total_sessions"`
			TotalPoints   int64   `json:"total_points"`
			TotalCost     float64 `json:"total_cost"`
			PassRate      float64 `json:"pass_rate"`
		}
		var rows []stat
		db.Raw(`SELECT t.user_id, u.real_name,
			COALESCE(SUM(t.hours),0) as total_hours,
			COUNT(*) as total_sessions,
			COALESCE(SUM(t.points),0) as total_points,
			COALESCE(SUM(t.cost),0) as total_cost
			FROM trainings t
			JOIN users u ON u.id=t.user_id
			GROUP BY t.user_id, u.real_name
			ORDER BY total_hours DESC`).Scan(&rows)
		// Calculate pass rate per staff
		for i := range rows {
			var passed, total int64
			db.Model(&model.Training{}).Where("user_id = ? AND passed = 'passed'", rows[i].UserID).Count(&passed)
			db.Model(&model.Training{}).Where("user_id = ? AND passed IN ('passed','failed')", rows[i].UserID).Count(&total)
			if total > 0 {
				rows[i].PassRate = float64(passed) / float64(total) * 100
			}
		}
		c.JSON(http.StatusOK, rows)
	}
}

// Category分布统计
func CategoryStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type catStat struct {
			Category string `json:"category"`
			Count    int64  `json:"count"`
			Hours    float64 `json:"hours"`
		}
		var rows []catStat
		db.Raw(`SELECT category, COUNT(*) as count, COALESCE(SUM(hours),0) as hours
			FROM trainings GROUP BY category ORDER BY count DESC`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

// 月度培训统计趋势
func MonthlyStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type monStat struct {
			Mon   string  `json:"mon"`
			Count int64   `json:"count"`
			Hours float64 `json:"hours"`
			Cost  float64 `json:"cost"`
		}
		var rows []monStat
		db.Raw(`SELECT SUBSTRING(date,1,7) as mon,
			COUNT(*) as count,
			COALESCE(SUM(hours),0) as hours,
			COALESCE(SUM(cost),0) as cost
			FROM trainings WHERE date != ''
			GROUP BY SUBSTRING(date,1,7)
			ORDER BY mon DESC LIMIT 12`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

// ========== TrainingPlan CRUD ==========

func ListPlans(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []model.TrainingPlan
		q := db.Order("created_at DESC")
		if status := c.Query("status"); status != "" {
			q = q.Where("status=?", status)
		}
		if role := c.Query("role_required"); role != "" {
			q = q.Where("role_required=?", role)
		}
		q.Find(&list)
		c.JSON(http.StatusOK, list)
	}
}

func CreatePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p model.TrainingPlan
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		db.Create(&p)
		c.JSON(http.StatusCreated, p)
	}
}

func UpdatePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var p model.TrainingPlan
		if err := db.First(&p, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "计划不存在"})
			return
		}
		var input model.TrainingPlan
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		input.ID = uint(id)
		db.Model(&p).Updates(input)
		c.JSON(http.StatusOK, input)
	}
}

func DeletePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Delete(&model.TrainingPlan{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}
