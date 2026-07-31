package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clinic-mgmt/internal/config"
	"clinic-mgmt/internal/dbcompat"
	"clinic-mgmt/internal/model"

	"gorm.io/gorm"
)

// BackupDir returns the directory where backup files are stored.
func BackupDir(cfg *config.Config) string {
	var dir string
	if dbcompat.IsSQLite() {
		// For SQLite, store backups next to the db file or in a sensible default
		dbPath := cfg.DBDsn
		d := filepath.Dir(dbPath)
		dir = filepath.Join(d, "backups")
	} else {
		dir = filepath.Join(filepath.Dir(os.Args[0]), "..", "backups")
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	os.MkdirAll(dir, 0755)
	return dir
}

// CreateBackup backs up the database.
// SQLite: copies the .db file directly. PostgreSQL: uses pg_dump.
func CreateBackup(db *gorm.DB, cfg *config.Config, backupType string) (*model.BackupRecord, error) {
	dir := BackupDir(cfg)
	timestamp := time.Now().Format("20060102_150405")

	record := &model.BackupRecord{
		BackupType:  backupType,
		Status:      "success",
		CloudStatus: "pending",
	}

	if dbcompat.IsSQLite() {
		// SQLite backup: copy the database file
		filename := fmt.Sprintf("clinic_backup_%s.db", timestamp)
		fp := filepath.Join(dir, filename)
		record.Filename = filename
		record.FilePath = fp

		if err := copyFile(cfg.DBDsn, fp); err != nil {
			record.Status = "failed"
			record.ErrorMessage = err.Error()
			db.Create(record)
			return record, fmt.Errorf("备份失败: %s", err.Error())
		}

		// Also export a SQL dump for portability
		sqlFile := filepath.Join(dir, fmt.Sprintf("clinic_backup_%s.sql", timestamp))
		exportSQLiteDump(db, sqlFile)
	} else {
		// PostgreSQL backup: use pg_dump
		filename := fmt.Sprintf("clinic_backup_%s.sql", timestamp)
		fp := filepath.Join(dir, filename)
		record.Filename = filename
		record.FilePath = fp

		dbc := parseDSN(cfg.DBDsn)
		if err := pgDump(dbc, fp); err != nil {
			record.Status = "failed"
			record.ErrorMessage = err.Error()
			db.Create(record)
			return record, fmt.Errorf("备份失败: %s", err.Error())
		}
	}

	// Get file size
	if fi, err := os.Stat(record.FilePath); err == nil {
		record.FileSize = fi.Size()
	}

	db.Create(record)

	// Try auto cloud upload if configured
	s := LoadSettings()
	if s.CloudUploadEnabled && s.Endpoint != "" {
		go UploadToCloud(record, s)
	}

	return record, nil
}

// copyFile copies src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// exportSQLiteDump writes a SQL representation of key tables with proper escaping.
func exportSQLiteDump(db *gorm.DB, filepath string) {
	f, err := os.Create(filepath)
	if err != nil {
		return
	}
	defer f.Close()

	tables := []string{"customers", "memberships", "member_packages", "orders", "order_items", "payments", "appointments"}
	for _, t := range tables {
		var rows []map[string]interface{}
		db.Table(t).Find(&rows)
		for _, row := range rows {
			cols := make([]string, 0, len(row))
			vals := make([]string, 0, len(row))
			for k, v := range row {
				cols = append(cols, k)
				vals = append(vals, sqlEscape(v))
			}
			fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s);\n", t, strings.Join(cols, ","), strings.Join(vals, ","))
		}
	}
}

// sqlEscape converts a Go value to a properly escaped SQL literal.
func sqlEscape(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case string:
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// DeleteBackup deletes a backup file and its database record.
func DeleteBackup(db *gorm.DB, id uint) error {
	var record model.BackupRecord
	if err := db.First(&record, id).Error; err != nil {
		return err
	}
	os.Remove(record.FilePath)
	db.Delete(&record)
	return nil
}

// CleanupOldBackups removes backups older than retentionDays.
func CleanupOldBackups(db *gorm.DB, retentionDays int) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var records []model.BackupRecord
	db.Where("created_at < ?", cutoff).Find(&records)
	for _, r := range records {
		os.Remove(r.FilePath)
		db.Delete(&r)
	}
}
