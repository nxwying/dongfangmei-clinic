package followup

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CustomerID uint   `json:"customer_id" binding:"required"`
			Type       string `json:"type" binding:"required"`
			DueDate    string `json:"due_date" binding:"required"`
			Note       string `json:"note"`
			AssignedTo *uint  `json:"assigned_to"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		userID := c.GetUint("user_id")
		task := model.FollowUpTask{
			CustomerID:  req.CustomerID,
			Type:        req.Type,
			DueDate:     req.DueDate,
			Note:        req.Note,
			Status:      "pending",
			CreatedBy:   userID,
			AssignedTo:  req.AssignedTo,
		}
		if err := db.Create(&task).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		system.WriteAuditLog(db, c, "create", "followup", task.ID, "创建回访任务")
		db.Preload("Customer").First(&task, task.ID)
		c.JSON(http.StatusCreated, task)
	}
}

func ListTasks(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Preload("Customer").Order("due_date ASC, created_at DESC")
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}
		if t := c.Query("type"); t != "" {
			query = query.Where("type = ?", t)
		}
		if a := c.Query("assigned_to"); a != "" {
			query = query.Where("assigned_to = ?", a)
		}
		var tasks []model.FollowUpTask
		query.Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	}
}

func CompleteTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct {
			Result       string `json:"result"`
			ResultStatus string `json:"result_status"`
		}
		c.ShouldBindJSON(&req)
		now := time.Now()
		updates := map[string]interface{}{
			"status": "completed", "completed_at": &now,
		}
		if req.Result != "" {
			updates["result"] = req.Result
		}
		if req.ResultStatus != "" {
			updates["result_status"] = req.ResultStatus
		}
		db.Model(&model.FollowUpTask{}).Where("id = ?", id).Updates(updates)
		system.WriteAuditLog(db, c, "complete", "followup", uint(id), "完成回访")
		c.JSON(http.StatusOK, gin.H{"message": "已完成"})
	}
}

func AutoGenerate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dormantCustomers []struct{ ID uint; Name string }
		db.Raw(`SELECT c.id, c.name FROM customers c
			LEFT JOIN orders o ON c.id=o.customer_id AND o.status='paid'
			GROUP BY c.id HAVING COALESCE(MAX(o.created_at),'2000-01-01') < NOW() - interval '90 days'`).Scan(&dormantCustomers)
		today := time.Now().Format("2006-01-02")
		count := 0
		userID := c.GetUint("user_id")
		for _, dc := range dormantCustomers {
			var existing int64
			db.Model(&model.FollowUpTask{}).Where("customer_id=? AND type='dormant' AND status='pending'", dc.ID).Count(&existing)
			if existing == 0 {
				db.Create(&model.FollowUpTask{
					CustomerID: dc.ID, Type: "dormant", DueDate: today,
					Note: "超过90天未到店", Status: "pending", CreatedBy: userID,
				})
				count++
			}
		}
		c.JSON(http.StatusOK, gin.H{"generated": count, "message": "已生成" + strconv.Itoa(count) + "条沉睡客户回访任务"})
	}
}

// AutoGeneratePostTreatment generates D1/D3/D7/D30 follow-up tasks
func AutoGeneratePostTreatment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CustomerID  uint   `json:"customer_id" binding:"required"`
			TreatmentID uint   `json:"treatment_id"`
			TreatDate   string `json:"treat_date"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		treatDate := req.TreatDate
		if treatDate == "" {
			treatDate = time.Now().Format("2006-01-02")
		}
		date, _ := time.Parse("2006-01-02", treatDate)
		intervals := []struct{ days int; note string }{
			{1, "术后第1天回访：关心恢复情况，提醒注意事项"},
			{3, "术后第3天回访：询问恢复进展，解答疑问"},
			{7, "术后第7天回访：了解初步效果，提醒复诊"},
			{30, "术后第30天回访：了解最终效果，邀请反馈"},
		}
		userID := c.GetUint("user_id")
		count := 0
		for _, iv := range intervals {
			due := date.AddDate(0, 0, iv.days)
			dueStr := due.Format("2006-01-02")
			var existing int64
			db.Model(&model.FollowUpTask{}).Where("customer_id=? AND type='post_treatment' AND due_date=?", req.CustomerID, dueStr).Count(&existing)
			if existing == 0 {
				task := model.FollowUpTask{
					CustomerID: req.CustomerID,
					Type:       "post_treatment",
					DueDate:    dueStr,
					Note:       iv.note,
					Status:     "pending",
					CreatedBy:  userID,
				}
				if req.TreatmentID > 0 {
					tid := req.TreatmentID
					task.MedicalRecordID = &tid
				}
				db.Create(&task)
				count++
			}
		}
		c.JSON(http.StatusOK, gin.H{"generated": count, "message": "已生成" + strconv.Itoa(count) + "条治疗后回访任务"})
	}
}

// Stats returns conversion and completion statistics
func Stats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type TypeStats struct {
			Type      string  `json:"type"`
			Total     int64   `json:"total"`
			Completed int64   `json:"completed"`
			Rebooked  int64   `json:"rebooked"`
			Rate      float64 `json:"rate"`
			ConvRate  float64 `json:"conv_rate"`
		}
		var rows []TypeStats
		db.Raw(`SELECT type,
			COUNT(*) as total,
			SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN result_status='rebooked' THEN 1 ELSE 0 END) as rebooked
			FROM follow_up_tasks GROUP BY type`).Scan(&rows)
		for i, r := range rows {
			if r.Total > 0 {
				rows[i].Rate = float64(r.Completed) / float64(r.Total) * 100
			}
			if r.Completed > 0 {
				rows[i].ConvRate = float64(r.Rebooked) / float64(r.Completed) * 100
			}
		}
		type Overall struct {
			Total     int64   `json:"total"`
			Completed int64   `json:"completed"`
			Rebooked  int64   `json:"rebooked"`
			Rate      float64 `json:"rate"`
			ConvRate  float64 `json:"conv_rate"`
		}
		var overall Overall
		db.Raw(`SELECT COUNT(*) as total,
			SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN result_status='rebooked' THEN 1 ELSE 0 END) as rebooked
			FROM follow_up_tasks`).Scan(&overall)
		if overall.Total > 0 {
			overall.Rate = float64(overall.Completed) / float64(overall.Total) * 100
		}
		if overall.Completed > 0 {
			overall.ConvRate = float64(overall.Rebooked) / float64(overall.Completed) * 100
		}
		c.JSON(http.StatusOK, gin.H{"by_type": rows, "overall": overall})
	}
}
