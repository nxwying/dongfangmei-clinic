package system

import (
	"encoding/json"
	"net/http"
	"strconv"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

)

type CreateUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	RealName string `json:"real_name" binding:"required"`
	Phone    string `json:"phone"`
	RoleID   uint   `json:"role_id" binding:"required"`
}

type UpdateUserReq struct {
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	RoleID   uint   `json:"role_id"`
	Password string `json:"password,omitempty"`
}





func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []model.User
		db.Preload("Role").Find(&users)
		c.JSON(http.StatusOK, users)
	}
}

func CreateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		user := model.User{
			Username:     req.Username,
			PasswordHash: string(hash),
			RealName:     req.RealName,
			Phone:        req.Phone,
			RoleID:       req.RoleID,
			Status:       "active",
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
			return
		}
		WriteAuditLog(db, c, "create", "user", user.ID, "创建员工 "+user.RealName)
		c.JSON(http.StatusCreated, gin.H{"id": user.ID, "message": "创建成功"})
	}
}

func UpdateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req UpdateUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		updates := map[string]interface{}{}
		if req.RealName != "" {
			updates["real_name"] = req.RealName
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.RoleID > 0 {
			updates["role_id"] = req.RoleID
		}
		if req.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
				return
			}
			updates["password_hash"] = string(hash)
		}
		if err := db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "更新失败"})
			return
		}
		WriteAuditLog(db, c, "update", "user", uint(id), "修改员工")
		WriteAuditLog(db, c, "update", "role", uint(id), "修改角色")
		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

func UpdateUserStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		result := db.Model(&model.User{}).Where("id = ?", id).Update("status", req.Status)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "更新失败"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
	}
}


func ListAuditLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 50
		}
		var total int64
		db.Model(&model.AuditLog{}).Count(&total)
		var logs []model.AuditLog
		db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)
		c.JSON(http.StatusOK, gin.H{
			"data":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// WriteAuditLog records an operation to the audit log.
func WriteAuditLog(db *gorm.DB, c *gin.Context, action, target string, targetID uint, detail string) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		return
	}
	log := model.AuditLog{
		UserID:   userID,
		Action:   action,
		Target:   target,
		TargetID: &targetID,
		Detail:   detail,
	}
	db.Create(&log)
}


// ======================== 角色管理 ========================

func ListRoles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var roles []model.Role
		db.Order("name ASC").Find(&roles)
		type RoleResp struct {
			ID          uint     `json:"id"`
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Permissions []string `json:"permissions"`
		}
		var resp []RoleResp
		for _, r := range roles {
			var perms []string
			json.Unmarshal(r.Permissions, &perms)
			if perms == nil { perms = []string{} }
			resp = append(resp, RoleResp{
				ID: r.ID, Name: r.Name,
				Description: r.Description,
				Permissions: perms,
			})
		}
		c.JSON(http.StatusOK, resp)
	}
}

func CreateRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string   `json:"name" binding:"required"`
			Description string   `json:"description"`
			Permissions []string `json:"permissions"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.Permissions == nil { req.Permissions = []string{} }
		permData, _ := json.Marshal(req.Permissions)
		role := model.Role{
			Name: req.Name, Description: req.Description,
			Permissions: permData,
		}
		if err := db.Create(&role).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败，角色名可能已存在"})
			return
		}
		WriteAuditLog(db, c, "create", "role", role.ID, "创建角色")
		c.JSON(http.StatusCreated, gin.H{"id": role.ID, "name": role.Name})
	}
}

func UpdateRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Permissions []string `json:"permissions"`
		}
		c.ShouldBindJSON(&req)
		updates := map[string]interface{}{}
		if req.Name != "" { updates["name"] = req.Name }
		if req.Description != "" { updates["description"] = req.Description }
		if req.Permissions != nil {
			permData, _ := json.Marshal(req.Permissions)
			updates["permissions"] = permData
		}
		db.Model(&model.Role{}).Where("id = ?", id).Updates(updates)
		WriteAuditLog(db, c, "update", "role", uint(id), "更新角色")
		c.JSON(http.StatusOK, gin.H{"message": "已更新"})
	}
}


func DeleteRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		// Prevent deleting admin role
		var role model.Role
		if err := db.First(&role, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
			return
		}
		if role.Name == "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除超级管理员角色"})
			return
		}
		// Remove role from users
		db.Model(&model.User{}).Where("role_id = ?", id).Update("role_id", 0)
		db.Delete(&role)
		WriteAuditLog(db, c, "delete", "role", uint(id), "删除角色")
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}

