package photo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"clinic-mgmt/internal/model"
	"clinic-mgmt/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const photoBasePath = "data/photos"

var photoAllowedTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

func List(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Query("customer_id")
		query := db.Model(&model.Photo{}).Order("created_at DESC")
		if customerID != "" {
			query = query.Where("customer_id = ?", customerID)
		}
		if t := c.Query("photo_type"); t != "" {
			query = query.Where("photo_type = ?", t)
		}
		if mr := c.Query("medical_record_id"); mr != "" {
			query = query.Where("medical_record_id = ?", mr)
		}
		var photos []model.Photo
		query.Find(&photos)
		c.JSON(http.StatusOK, photos)
	}
}

func Upload(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件太大，最大20MB"})
			return
		}
		customerID, _ := strconv.ParseUint(c.PostForm("customer_id"), 10, 64)
		if customerID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择客户"})
			return
		}
		photoType := c.PostForm("photo_type")
if photoType != "before" && photoType != "after" && photoType != "during" && photoType != "followup" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "照片类型必须为 before/during/after/followup"})
			return
		}

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
			return
		}
		defer file.Close()

		fileType := header.Header.Get("Content-Type")
		if _, ok := photoAllowedTypes[fileType]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 JPG/PNG/WebP 格式"})
			return
		}

		dir := filepath.Join(photoBasePath, fmt.Sprintf("customer_%d", customerID))
		os.MkdirAll(dir, 0755)

		ext := filepath.Ext(header.Filename)
		name := fmt.Sprintf("%s_%s_%s%s", time.Now().Format("20060102_150405"), photoType, strconv.FormatUint(customerID, 10), ext)
		path := filepath.Join(dir, name)

		out, err := os.Create(path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}
		defer out.Close()
		io.Copy(out, file)

		relPath := filepath.Join(fmt.Sprintf("customer_%d", customerID), name)

		var treatmentID *uint
		var medRecID *uint
		if medicalRecordID := c.PostForm("medical_record_id"); medicalRecordID != "" {
			if mid, err := strconv.ParseUint(medicalRecordID, 10, 64); err == nil {
				mr := uint(mid)
				medRecID = &mr
			}
		}
		if t := c.PostForm("treatment_id"); t != "" {
			if id, err := strconv.ParseUint(t, 10, 64); err == nil {
				tid := uint(id)
				treatmentID = &tid
			}
		}

		photo := model.Photo{
			CustomerID:  uint(customerID),
			TreatmentID: treatmentID,
			MedicalRecordID: medRecID,
			PhotoType:   photoType,
			BodyPart:    c.PostForm("body_part"),
			FilePath:    relPath,
			FileName:    header.Filename,
			FileSize:    header.Size,
			FileType:    fileType,
			CreatedBy:   c.GetUint("user_id"),
		}
		if err := db.Create(&photo).Error; err != nil {
			os.Remove(path)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存记录失败"})
			return
		}
		system.WriteAuditLog(db, c, "create", "photo", photo.ID, "上传照片")
		c.JSON(http.StatusCreated, photo)
	}
}

func Delete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var p model.Photo
		if err := db.First(&p, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
			return
		}
		fullPath := filepath.Join(photoBasePath, p.FilePath)
		os.Remove(fullPath)
		db.Delete(&p)
		system.WriteAuditLog(db, c, "delete", "photo", p.ID, "删除照片")
		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

func Download(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var p model.Photo
		if err := db.First(&p, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
			return
		}
		fullPath := filepath.Join(photoBasePath, p.FilePath)
		absPath, _ := filepath.Abs(fullPath)
		absBase, _ := filepath.Abs(photoBasePath)
		if !strings.HasPrefix(absPath, absBase) {
			c.JSON(http.StatusForbidden, gin.H{"error": "非法路径"})
			return
		}
		c.Header("Content-Type", p.FileType)
		c.File(absPath)
	}
}
