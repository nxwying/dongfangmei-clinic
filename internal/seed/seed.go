package seed

import (
	"encoding/json"
	"log"

	"clinic-mgmt/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	// Only seed if no roles exist
	var count int64
	db.Model(&model.Role{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Seeding initial data...")

	roles := []struct {
		Name        string
		Description string
		Permissions []string
	}{
		{
			Name: "admin", Description: "管理员（全部权限）",
			Permissions: []string{"admin"},
		},
		{
			Name: "receptionist", Description: "前台（客户管理、收银、预约）",
			Permissions: []string{
				"customer:read", "customer:write",
				"membership:read", "membership:recharge",
				"order:read", "order:write", "order:pay",
				"appointment:read", "appointment:write",
				"settings:read", "settings:write",
			},
		},
		{
			Name: "consultant", Description: "咨询师（客户管理、开单）",
			Permissions: []string{
				"customer:read", "customer:write",
				"membership:read", "membership:write",
				"order:read", "order:write",
				"appointment:read", "appointment:write",
			},
		},
		{
			Name: "doctor", Description: "医生（查看客户、查看订单、预约）",
			Permissions: []string{
				"customer:read",
				"order:read",
				"appointment:read", "appointment:write",
			},
		},
		{
			Name: "inventory", Description: "库管",
			Permissions: []string{},
		},
	}

	for _, r := range roles {
		perms, err := json.Marshal(r.Permissions)
		if err != nil {
			log.Fatalf("failed to marshal permissions: %v", err)
		}
		db.Create(&model.Role{
			Name:        r.Name,
			Description: r.Description,
			Permissions: perms,
		})
	}

	// Create admin user (password: admin123)
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	var adminRole model.Role
	db.Where("name = ?", "admin").First(&adminRole)

	db.Create(&model.User{
		Username:     "admin",
		PasswordHash: string(hash),
		RealName:     "系统管理员",
		RoleID:       adminRole.ID,
		Status:       "active",
	})

	log.Println("Seed complete: 5 roles + admin user created")
}
