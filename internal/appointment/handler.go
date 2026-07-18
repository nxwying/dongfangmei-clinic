package appointment

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"clinic-mgmt/internal/system"
)

// CreateAppointmentReq is the request body for creating an appointment.
type CreateAppointmentReq struct {
	CustomerID   uint   `json:"customer_id" binding:"required"`
	ConsultantID *uint  `json:"consultant_id"`
	DoctorID     *uint  `json:"doctor_id"`
	Date         string `json:"date" binding:"required"`
	TimeSlot     string `json:"time_slot"`
	Items        string `json:"items"`
	Remark       string `json:"remark"`
	Duration     int    `json:"duration"`
}

// ListAppointments returns appointments with optional filters.
func ListAppointments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Preload("Customer").Order("date ASC, time_slot ASC")

		if date := c.Query("date"); date != "" {
			if parsed, err := time.Parse("2006-01-02", date); err == nil {
				query = query.Where("date = ?", parsed)
			}
		}
		if startDate := c.Query("start_date"); startDate != "" {
			if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
				query = query.Where("date >= ?", parsed)
			}
		}
		if endDate := c.Query("end_date"); endDate != "" {
			if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
				query = query.Where("date <= ?", parsed)
			}
		}
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}

		var appointments []model.Appointment
		if err := query.Find(&appointments).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		c.JSON(http.StatusOK, appointments)
	}
}

// CreateAppointment creates a new appointment.
func CreateAppointment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateAppointmentReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.Duration <= 0 {
			req.Duration = 30
		}

		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
			return
		}

		userID := c.GetUint("user_id")

		// Conflict check: same doctor, same date, same time slot
		if req.DoctorID != nil && *req.DoctorID > 0 && req.TimeSlot != "" {
			var conflictCount int64
			db.Model(&model.Appointment{}).
				Where("doctor_id = ? AND date = ? AND time_slot = ? AND status NOT IN ('cancelled') AND id != 0",
					*req.DoctorID, date, req.TimeSlot).
				Count(&conflictCount)
			if conflictCount > 0 {
				c.JSON(http.StatusConflict, gin.H{"warning": "该医生在此时间段已有预约", "conflict": true})
				return
			}
		}

		appointment := model.Appointment{
			CustomerID:   req.CustomerID,
			ConsultantID: req.ConsultantID,
			DoctorID:     req.DoctorID,
			Date:         date,
			TimeSlot:     req.TimeSlot,
			Items:        req.Items,
			Status:       "booked",
			Remark:       req.Remark,
			CreatedBy:    userID,
			Duration:     req.Duration,
		}

		if err := db.Create(&appointment).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "创建预约失败"})
			return
		}

		system.WriteAuditLog(db, c, "create", "appointment", appointment.ID, "创建预约")
		db.Preload("Customer").First(&appointment, appointment.ID)
		c.JSON(http.StatusCreated, appointment)
	}
}

// CheckIn marks an appointment as checked in.
func CheckIn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var appointment model.Appointment
		if err := db.First(&appointment, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
			return
		}
		if err := db.Model(&appointment).Update("status", "checked_in").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "签到失败"})
			return
		}
		appointment.Status = "checked_in"
		db.Preload("Customer").First(&appointment, id)
		c.JSON(http.StatusOK, appointment)
	}
}

// Complete marks an appointment as completed.
func Complete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var appointment model.Appointment
		if err := db.First(&appointment, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
			return
		}
		now := time.Now()
		if err := db.Model(&appointment).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "完成失败"})
			return
		}
		appointment.Status = "completed"
		appointment.CompletedAt = &now
		db.Preload("Customer").First(&appointment, id)
		c.JSON(http.StatusOK, appointment)
	}
}

// Cancel cancels an appointment.
func Cancel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var appointment model.Appointment
		if err := db.First(&appointment, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
			return
		}
		if err := db.Model(&appointment).Update("status", "cancelled").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消失败"})
			return
		}
		appointment.Status = "cancelled"
		db.Preload("Customer").First(&appointment, id)
		c.JSON(http.StatusOK, appointment)
	}
}
