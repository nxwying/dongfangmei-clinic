package license

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func StatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, GetStatus())
	}
}

func ActivateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择授权文件"})
			return
		}
		defer file.Close()
		data := make([]byte, 1<<20)
		n, _ := file.Read(data)
		lic, err := SaveLicense(data[:n])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "激活成功",
			"customer": lic.Customer,
			"expires":  lic.ExpiresAt,
		})
	}
}
