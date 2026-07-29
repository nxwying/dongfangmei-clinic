package middleware

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"clinic-mgmt/internal/license"
	"clinic-mgmt/internal/model"
)

func LicenseCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" { c.Next(); return }
		if license.IsActivated() { c.Next(); return }
		path := c.Request.URL.Path
		if strings.HasSuffix(path, "/customers") {
			var count int64
			db.Model(&model.Customer{}).Count(&count)
			if count >= 10 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "试用版最多添加10个客户", "code": "LICENSE_LIMIT"})
				return
			}
		}
		if strings.HasPrefix(path, "/api/v1/medical/records") {
			var count int64
			db.Model(&model.MedicalRecord{}).Count(&count)
			if count >= 30 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "试用版最多产生30条病历", "code": "LICENSE_LIMIT"})
				return
			}
		}
		c.Next()
	}
}
