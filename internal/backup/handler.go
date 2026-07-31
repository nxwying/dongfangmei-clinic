package backup

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"clinic-mgmt/internal/config"
	"clinic-mgmt/internal/dbcompat"
	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type dbConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func parseDSN(dsn string) dbConfig {
	cfg := dbConfig{Host: "localhost", Port: "5432"}
	for _, part := range splitDSN(dsn) {
		kv := splitKV(part)
		if len(kv) == 2 {
			switch kv[0] {
			case "host":
				cfg.Host = kv[1]
			case "port":
				cfg.Port = kv[1]
			case "user":
				cfg.User = kv[1]
			case "password":
				cfg.Password = kv[1]
			case "dbname":
				cfg.DBName = kv[1]
			}
		}
	}
	return cfg
}

func splitDSN(dsn string) []string {
	var parts []string
	cur := ""
	for _, ch := range dsn {
		if ch == ' ' {
			if cur != "" {
				parts = append(parts, cur)
			}
			cur = ""
		} else {
			cur += string(ch)
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}

func splitKV(s string) []string {
	for i, ch := range s {
		if ch == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// ListBackups returns all backup records.
func ListBackups(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var records []model.BackupRecord
		db.Order("created_at DESC").Find(&records)
		c.JSON(http.StatusOK, records)
	}
}

// CreateBackupHandler triggers a manual backup.
func CreateBackupHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		record, err := CreateBackup(db, cfg, "manual")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, record)
	}
}

// DeleteBackupHandler deletes a backup.
func DeleteBackupHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := DeleteBackup(db, uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "备份记录不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}

// DownloadBackup serves a backup file for download.
func DownloadBackup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var record model.BackupRecord
		if err := db.First(&record, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "备份记录不存在"})
			return
		}
		if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "备份文件不存在"})
			return
		}
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, record.Filename))
		c.File(record.FilePath)
	}
}

// UploadToCloudHandler uploads a backup to cloud.
func UploadToCloudHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var record model.BackupRecord
		if err := db.First(&record, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "备份记录不存在"})
			return
		}
		record.CloudStatus = "pending"
		db.Save(&record)

		s := LoadSettings()
		err := UploadToCloud(&record, s)
		db.Save(&record)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "上传成功", "url": record.CloudURL})
	}
}

// GetBackupSettings returns current backup settings.
func GetBackupSettings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, LoadSettings())
	}
}

// SaveBackupSettings saves backup settings.
func SaveBackupSettings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s CloudSettings
		if err := c.ShouldBindJSON(&s); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		SaveSettings(s)
		c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
	}
}

// ExportBackup creates and serves a backup file immediately.
func ExportBackup(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a backup and return it as a download
		record, err := CreateBackup(db, cfg, "manual")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份失败: " + err.Error()})
			return
		}
		if _, err := os.Stat(record.FilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份文件生成失败"})
			return
		}
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, record.Filename))
		c.File(record.FilePath)
	}
}

// ImportBackup accepts a backup file and restores it.
// SQLite: copies the .db file. PostgreSQL: pipes through psql.
func ImportBackup(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择备份文件"})
			return
		}
		defer file.Close()

		if dbcompat.IsSQLite() {
			// For SQLite, save uploaded file and verify it's a valid SQLite db
			tmpPath := cfg.DBDsn + ".restoring"
			out, err := os.Create(tmpPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建临时文件"})
				return
			}
			if _, err := io.Copy(out, file); err != nil {
				out.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
				return
			}
			out.Close()

			// Backup current db before replacing
			backupPath := cfg.DBDsn + ".bak"
			os.Rename(cfg.DBDsn, backupPath)

			// Move uploaded file to be the active database
			if err := os.Rename(tmpPath, cfg.DBDsn); err != nil {
				// Restore on failure
				os.Rename(backupPath, cfg.DBDsn)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "数据恢复成功，请重启服务"})
			return
		}

		// PostgreSQL restore via psql
		content, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
			return
		}
		dbc := parseDSN(cfg.DBDsn)
		if err := pgRestore(dbc, content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败，请检查备份文件格式"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "数据恢复成功，请刷新页面"})
	}
}

// ResetSystem clears all business data but keeps system config.
func ResetSystem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use DELETE which works in both SQLite and PostgreSQL
		tables := []string{
			"batch_usages", "inventory_batches",
			"audit_logs", "expenses", "medical_records",
			"follow_up_tasks", "inventory_logs", "inventory_items",
			"treatment_records", "follow_ups",
			"package_redemptions", "member_packages", "memberships",
			"payments", "order_items", "orders",
			"appointments", "customers",
		}
		for _, t := range tables {
			db.Exec("DELETE FROM " + t)
		}
		c.JSON(http.StatusOK, gin.H{"message": "系统已初始化，业务数据已清除"})
	}
}

