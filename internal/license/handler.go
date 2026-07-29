package license

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"clinic-mgmt/internal/model"
)

type Status struct {
	Activated     bool   `json:"activated"`
	MachineCode   string `json:"machine_code"`
	CustomerCount int64  `json:"customer_count"`
	RecordCount   int64  `json:"record_count"`
	CustomerLimit int    `json:"customer_limit"`
	RecordLimit   int    `json:"record_limit"`
}

func GetStatus(db *gorm.DB) *Status {
	var custCount, recCount int64
	db.Model(&model.Customer{}).Count(&custCount)
	db.Model(&model.MedicalRecord{}).Count(&recCount)
	activated := IsActivated()
	s := &Status{
		Activated:     activated,
		MachineCode:   MachineCode(),
		CustomerCount: custCount,
		RecordCount:   recCount,
	}
	if !activated {
		s.CustomerLimit = 10
		s.RecordLimit = 30
	}
	return s
}

func StatusHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, GetStatus(db))
	}
}

func ActivateHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct{ Code string `json:"code" binding:"required"` }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入解锁码"})
			return
		}
		if err := Activate(req.Code); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "激活成功"})
	}
}
