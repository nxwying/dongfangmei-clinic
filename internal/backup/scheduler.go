package backup

import (
	"log"
	"time"

	"clinic-mgmt/internal/config"

	"gorm.io/gorm"
)

// StartScheduler starts the background backup scheduler
func StartScheduler(db *gorm.DB, cfg *config.Config) {
	log.Println("[备份调度器] 已启动")

	// Check immediately if auto backup is needed
	checkAndRun(db, cfg)

	// Check every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		checkAndRun(db, cfg)
	}
}

func checkAndRun(db *gorm.DB, cfg *config.Config) {
	if !settings.AutoBackupEnabled {
		return
	}

	// Determine if we should run based on interval
	now := time.Now()
	var lastBackup time.Time
	var lastRecord struct {
		CreatedAt time.Time
	}
	db.Model(&struct{}{}).Table("backup_records").
		Select("created_at").
		Where("backup_type = 'auto' AND status = 'success'").
		Order("created_at DESC").
		Limit(1).
		Scan(&lastRecord)
	lastBackup = lastRecord.CreatedAt

	shouldRun := false
	switch settings.BackupInterval {
	case "daily":
		shouldRun = lastBackup.IsZero() || now.Sub(lastBackup) >= 24*time.Hour
	case "weekly":
		shouldRun = lastBackup.IsZero() || now.Sub(lastBackup) >= 7*24*time.Hour
	case "monthly":
		shouldRun = lastBackup.IsZero() || now.Sub(lastBackup) >= 30*24*time.Hour
	default:
		return
	}

	if shouldRun {
		log.Println("[备份调度器] 开始自动备份...")
		record, err := CreateBackup(db, cfg, "auto")
		if err != nil {
			log.Printf("[备份调度器] 备份失败: %v", err)
		} else {
			log.Printf("[备份调度器] 备份成功: %s (%.1f MB)", record.Filename, float64(record.FileSize)/1024/1024)
		}

		// Cleanup old backups
		if settings.RetentionDays > 0 {
			CleanupOldBackups(db, settings.RetentionDays)
			log.Printf("[备份调度器] 已清理 %d 天前的旧备份", settings.RetentionDays)
		}
	}
}
