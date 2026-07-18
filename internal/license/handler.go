package license

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Status returns current license info (no auth required)
func StatusHandler(_ *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := GetStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, status)
	}
}

// Activate handles license file upload and activation
func ActivateHandler(_ *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("license")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择授权文件"})
			return
		}
		defer file.Close()

		data := make([]byte, 10<<20) // 10MB max
		n, err := file.Read(data)
		if err != nil && err.Error() != "EOF" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取文件失败"})
			return
		}

		lic, err := SaveLicense(data[:n])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "授权成功",
			"customer": lic.Customer,
			"expires":  lic.ExpiresAt,
		})
	}
}
