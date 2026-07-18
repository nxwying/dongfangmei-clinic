package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"clinic-mgmt/internal/auth/service"
	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	RealName string `json:"real_name"`
	RoleID   uint   `json:"role_id"`
}

func Login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
			return
		}

		var user model.User
		if err := db.Where("username = ? AND status = ?", req.Username, "active").Preload("Role").First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		var permissions []string
		if user.Role.Permissions != nil {
			if err := json.Unmarshal(user.Role.Permissions, &permissions); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "权限数据解析失败"})
				return
			}
		}

		token, err := service.GenerateToken(user.ID, user.Username, user.RealName, user.RoleID, permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成Token失败"})
			return
		}

		// Update last login time
		now := time.Now().Unix()
		db.Model(&user).Update("last_login_at", now)

		c.JSON(http.StatusOK, LoginResponse{
			Token:    token,
			RealName: user.RealName,
			RoleID:   user.RoleID,
		})
	}
}

func GetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		userID, ok := userIDVal.(uint)
		if !exists || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID无效"})
			return
		}
		var user model.User
		if err := db.Preload("Role").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"real_name": user.RealName,
			"phone":     user.Phone,
			"role":      user.Role,
		})
	}
}
