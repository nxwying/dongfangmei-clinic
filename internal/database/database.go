// internal/database/database.go
package database

import (
	"log"

	"clinic-mgmt/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Customer{},
		&model.Membership{},
		&model.MemberPackage{},
		&model.PackageRedemption{},
		&model.Appointment{},
		&model.Order{},
		&model.OrderItem{},
		&model.Payment{},
		&model.AuditLog{},
		&model.FollowUp{},
		&model.TreatmentItem{},
		&model.PackageTemplate{},
		&model.Expense{},
		&model.TreatmentRecord{},
		&model.FollowUpTask{},
		&model.InventoryItem{},
		&model.InventoryLog{},
		&model.MedicalRecord{},
		&model.MedicalRecordTemplate{},
		&model.Document{},
		&model.BackupRecord{},
		&model.TagRule{},
		&model.Photo{},
		&model.Training{},
		&model.TrainingPlan{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}


	return db
}
