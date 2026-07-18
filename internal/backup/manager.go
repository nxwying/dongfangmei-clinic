package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"clinic-mgmt/internal/config"
	"clinic-mgmt/internal/model"

	"gorm.io/gorm"
)

// findExe searches for an executable in common paths, used for pg_dump/psql
func findExe(name string) string {
	paths := []string{
		"/opt/homebrew/bin/" + name,
		"/usr/local/bin/" + name,
		"/usr/bin/" + name,
		"/usr/lib/postgresql/*/bin/" + name,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Fallback to PATH lookup
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return name
}

// BackupDir returns the directory where backup files are stored
func BackupDir(cfg *config.Config) string {
	dir := filepath.Join(filepath.Dir(os.Args[0]), "..", "backups")
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	os.MkdirAll(dir, 0755)
	return dir
}

// CreateBackup runs pg_dump and saves the file, returns the record
func CreateBackup(db *gorm.DB, cfg *config.Config, backupType string) (*model.BackupRecord, error) {
	dir := BackupDir(cfg)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("clinic_backup_%s.sql", timestamp)
	filepath := filepath.Join(dir, filename)

	// Run pg_dump
	dbc := parseDSN(cfg.DBDsn)
	cmd := exec.Command(findExe("pg_dump"),
		"-h", dbc.Host,
		"-p", dbc.Port,
		"-U", dbc.User,
		"-d", dbc.DBName,
		"--no-owner", "--no-acl",
		"-f", filepath,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbc.Password, "PGSSLMODE=disable")

	output, err := cmd.CombinedOutput()
	record := &model.BackupRecord{
		Filename:   filename,
		FileSize:   0,
		FilePath:   filepath,
		BackupType: backupType,
		Status:     "success",
		CloudStatus: "pending",
	}

	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("pg_dump error: %s", string(output))
		// Still save the record so user can see the failure
		db.Create(record)
		return record, fmt.Errorf("备份失败: %s", string(output))
	}

	// Get file size
	if fi, err := os.Stat(filepath); err == nil {
		record.FileSize = fi.Size()
	}

	// Save record
	db.Create(record)

	// Try auto cloud upload if configured
	settings := LoadSettings()
	if settings.CloudUploadEnabled && settings.Endpoint != "" {
		go UploadToCloud(record, settings)
	}

	return record, nil
}

// DeleteBackup deletes a backup file and its database record
func DeleteBackup(db *gorm.DB, id uint) error {
	var record model.BackupRecord
	if err := db.First(&record, id).Error; err != nil {
		return err
	}
	// Delete file
	os.Remove(record.FilePath)
	// Delete record
	db.Delete(&record)
	return nil
}

// CleanupOldBackups removes backups older than retentionDays
func CleanupOldBackups(db *gorm.DB, retentionDays int) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var records []model.BackupRecord
	db.Where("created_at < ?", cutoff).Find(&records)
	for _, r := range records {
		os.Remove(r.FilePath)
		db.Delete(&r)
	}
}
