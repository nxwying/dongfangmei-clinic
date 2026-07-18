package backup

import (
	"fmt"
	"os/exec"
	"strings"

	"clinic-mgmt/internal/model"
)

// CloudSettings holds cloud storage configuration
type CloudSettings struct {
	CloudUploadEnabled bool   `json:"cloud_upload_enabled"`
	Provider           string `json:"provider"` // "oss" or "cos"
	Endpoint           string `json:"endpoint"`
	Bucket             string `json:"bucket"`
	Region             string `json:"region"`
	AccessKey          string `json:"access_key"`
	SecretKey          string `json:"secret_key"`
	RetentionDays      int    `json:"retention_days"`
	AutoBackupEnabled  bool   `json:"auto_backup_enabled"`
	BackupInterval     string `json:"backup_interval"` // "daily","weekly","monthly"
}

// LoadSettings reads cloud settings from config file
func LoadSettings() CloudSettings {
	return settings
}

var settings CloudSettings

// SaveSettings saves cloud settings
func SaveSettings(s CloudSettings) {
	settings = s
}

// UploadToCloud uploads a backup file to cloud storage using curl with AWS SigV4
func UploadToCloud(record *model.BackupRecord, cfg CloudSettings) error {
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		record.CloudStatus = "failed"
		record.ErrorMessage = "云存储未配置完整"
		return fmt.Errorf("cloud storage not configured")
	}

	filename := record.Filename
	url := fmt.Sprintf("https://%s.%s/%s", cfg.Bucket, cfg.Endpoint, filename)

	// Build curl command with AWS Signature V4
	args := []string{
		"-X", "PUT",
		"-T", record.FilePath,
		"--aws-sigv4", fmt.Sprintf("aws:amz:%s", cfg.Region),
		"--user", fmt.Sprintf("%s:%s", cfg.AccessKey, cfg.SecretKey),
		"--fail", "--silent", "--show-error",
		url,
	}

	cmd := exec.Command("curl", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		record.CloudStatus = "failed"
		record.ErrorMessage = fmt.Sprintf("上传失败: %s", strings.TrimSpace(string(output)))
		return fmt.Errorf("cloud upload failed: %s", string(output))
	}

	record.CloudStatus = "uploaded"
	record.CloudURL = url
	return nil
}
