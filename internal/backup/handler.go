package backup

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"clinic-mgmt/internal/config"
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
	for _, part := range strings.Fields(dsn) {
		kv := strings.SplitN(part, "=", 2)
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

// ListBackups returns all backup records
func ListBackups(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var records []model.BackupRecord
		db.Order("created_at DESC").Find(&records)
		c.JSON(http.StatusOK, records)
	}
}

// CreateBackupHandler triggers a manual backup
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

// DeleteBackupHandler deletes a backup
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

// DownloadBackup serves a backup file for download
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

// UploadToCloudHandler uploads a backup to cloud
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

		err := UploadToCloud(&record, settings)
		db.Save(&record)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "上传成功", "url": record.CloudURL})
	}
}

// GetBackupSettings returns current backup settings
func GetBackupSettings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, settings)
	}
}

// SaveBackupSettings saves backup settings
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

// ExportBackup runs pg_dump and returns the SQL file as a download (legacy support)
func ExportBackup(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbc := parseDSN(cfg.DBDsn)
		cmd := exec.Command(findExe("pg_dump"),
			"-h", dbc.Host,
			"-p", dbc.Port,
			"-U", dbc.User,
			"-d", dbc.DBName,
			"--no-owner", "--no-acl",
		)
		cmd.Env = append(os.Environ(), "PGPASSWORD="+dbc.Password, "PGSSLMODE=disable")

		output, err := cmd.Output()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份失败：" + err.Error()})
			return
		}

		filename := fmt.Sprintf("clinic_backup_%s.sql", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		c.Data(http.StatusOK, "application/octet-stream", output)
	}
}

// ImportBackup accepts a .sql file and restores it via psql.
func ImportBackup(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择备份文件"})
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
			return
		}

		dbc := parseDSN(cfg.DBDsn)
		cmd := exec.Command(findExe("psql"),
			"-h", dbc.Host,
			"-p", dbc.Port,
			"-U", dbc.User,
			"-d", dbc.DBName,
		)
		cmd.Env = append(os.Environ(), "PGPASSWORD="+dbc.Password, "PGSSLMODE=disable")
		cmd.Stdin = bytes.NewReader(content)

		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败，请检查备份文件格式"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "数据恢复成功，请刷新页面"})
	}
}

// ResetSystem clears all business data but keeps system config.
func ResetSystem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tables := []string{
			"audit_logs", "expenses", "medical_records",
			"follow_up_tasks", "inventory_logs", "inventory_items",
			"treatment_records", "follow_ups",
			"package_redemptions", "member_packages", "memberships",
			"payments", "order_items", "orders",
			"appointments", "customers",
		}
		for _, t := range tables {
			db.Exec("TRUNCATE TABLE " + t + " CASCADE")
		}
		c.JSON(http.StatusOK, gin.H{"message": "系统已初始化，业务数据已清除"})
	}
}
