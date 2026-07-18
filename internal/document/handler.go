package document

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

// documentsBasePath is relative to the working directory.
const documentsBasePath = "data/documents"

var allowedTypes = map[string]string{
	"application/pdf":       "pdf",
	"image/jpeg":            "jpg",
	"image/png":             "png",
}

var docTypeNames = map[string]string{
	"invoice":     "invoice",
	"delivery":    "delivery",
	"mfr_qual":    "mfr_qual",
	"dist_qual":   "dist_qual",
	"inspection":  "inspection",
}

func docTypeDir(docType string) string {
	return filepath.Join(documentsBasePath, docType)
}

// sanitizeFilename removes characters that could cause path traversal or file issues.
func sanitizeFilename(name string) string {
	// Replace path separators and null bytes
	name = strings.NewReplacer(
		"/", "_", "\\", "_", "\x00", "", ":", "_",
		" ", "_", "　", "_",
	).Replace(name)
	// Trim leading dots and spaces
	name = strings.TrimLeft(name, "._")
	if name == "" {
		name = "unnamed"
	}
	// Limit length
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

// List returns paginated documents with filters.
func List(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&model.Document{}).Order("created_at DESC")

		if docType := c.Query("doc_type"); docType != "" {
			query = query.Where("doc_type = ?", docType)
		}
		if product := c.Query("product_name"); product != "" {
			query = query.Where("product_name ILIKE ?", "%"+product+"%")
		}
		if supplier := c.Query("supplier"); supplier != "" {
			query = query.Where("supplier ILIKE ?", "%"+supplier+"%")
		}
		if keyword := c.Query("keyword"); keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("title ILIKE ? OR product_name ILIKE ? OR supplier ILIKE ? OR serial_no ILIKE ?",
				like, like, like, like)
		}
		if start := c.Query("start_date"); start != "" {
			query = query.Where("issue_date >= ?", start)
		}
		if end := c.Query("end_date"); end != "" {
			query = query.Where("issue_date <= ?", end)
		}

		today := time.Now().Format("2006-01-02")
		thirtyDaysLater := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

		if c.Query("expiring_soon") == "true" {
			query = query.Where("expiry_date != '' AND expiry_date >= ? AND expiry_date <= ?", today, thirtyDaysLater)
		}
		if c.Query("expired") == "true" {
			query = query.Where("expiry_date != '' AND expiry_date < ?", today)
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		var total int64
		query.Count(&total)

		var items []model.Document
		query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&items)

		c.JSON(http.StatusOK, gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"items":     items,
		})
	}
}

// Create handles file upload with metadata.
func Create(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse multipart form (max 50MB)
		if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件太大或格式错误，最大50MB"})
			return
		}

		docType := c.PostForm("doc_type")
		if docType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择证件类型"})
			return
		}
		if _, ok := docTypeNames[docType]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的证件类型"})
			return
		}

		title := c.PostForm("title")
		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入文件标题"})
			return
		}

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择上传文件"})
			return
		}
		defer file.Close()

		// Validate file type
		fileType := header.Header.Get("Content-Type")
		if _, ok := allowedTypes[fileType]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 PDF、JPG、PNG 格式"})
			return
		}

		// Validate file size (extra check)
		if header.Size > 50<<20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过50MB"})
			return
		}

		// Build storage path
		dir := docTypeDir(docType)
		if err := os.MkdirAll(dir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建存储目录失败"})
			return
		}

		// Generate unique filename: date_title_sanitized.ext
		ext := filepath.Ext(header.Filename)
		baseName := fmt.Sprintf("%s_%s",
			time.Now().Format("20060102"),
			sanitizeFilename(title),
		)
		if len(baseName) > 200 {
			baseName = baseName[:200]
		}
		storageName := baseName + ext
		storagePath := filepath.Join(dir, storageName)

		// Handle filename conflicts by appending a counter
		for i := 1; fileExists(storagePath); i++ {
			storageName = fmt.Sprintf("%s_%d%s", baseName, i, ext)
			storagePath = filepath.Join(dir, storageName)
		}

		// Save file to disk
		out, err := os.Create(storagePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			os.Remove(storagePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
			return
		}

		// Store relative path in DB
		relPath := filepath.Join(docType, storageName)

		// Parse fields
		amount, _ := strconv.ParseFloat(c.PostForm("amount"), 64)

		doc := model.Document{
			DocType:     docType,
			Title:       title,
			FileName:    header.Filename,
			FilePath:    relPath,
			FileSize:    header.Size,
			FileType:    fileType,
			ProductName: c.PostForm("product_name"),
			Supplier:    c.PostForm("supplier"),
			SerialNo:    c.PostForm("serial_no"),
			IssueDate:   c.PostForm("issue_date"),
			ExpiryDate:  c.PostForm("expiry_date"),
			Amount:      amount,
			Remark:      c.PostForm("remark"),
			CreatedBy:   c.GetUint("user_id"),
		}

		if err := db.Create(&doc).Error; err != nil {
			os.Remove(storagePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存记录失败"})
			return
		}

		system.WriteAuditLog(db, c, "create", "document", doc.ID, "上传证件: "+doc.Title)
		c.JSON(http.StatusCreated, doc)
	}
}

// Get returns a single document by ID.
func Get(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var doc model.Document
		if err := db.First(&doc, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "证件不存在"})
			return
		}
		c.JSON(http.StatusOK, doc)
	}
}

// Delete soft-deletes a document.
func Delete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var doc model.Document
		if err := db.First(&doc, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "证件不存在"})
			return
		}
		db.Delete(&doc)
		system.WriteAuditLog(db, c, "delete", "document", doc.ID, "删除证件: "+doc.Title)
		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

// Download serves the file for preview or download.
func Download(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var doc model.Document
		if err := db.First(&doc, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "证件不存在"})
			return
		}

		fullPath := filepath.Join(documentsBasePath, doc.FilePath)
		// Security: prevent path traversal
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "文件路径错误"})
			return
		}
		absBase, _ := filepath.Abs(documentsBasePath)
		if !strings.HasPrefix(absPath, absBase) {
			c.JSON(http.StatusForbidden, gin.H{"error": "非法文件路径"})
			return
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}

		c.Header("Content-Type", doc.FileType)
		if c.Query("download") == "1" {
			c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.FileName))
		} else {
			c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, doc.FileName))
		}
		c.File(absPath)
	}
}

// Expiring returns documents that expire within 30 days or are already expired.
func Expiring(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		today := time.Now().Format("2006-01-02")
		thirtyDaysLater := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

		type ExpiringDoc struct {
			model.Document
			Status string `json:"status"`
		}

		var docs []model.Document
		db.Where("expiry_date != '' AND (expiry_date < ? OR expiry_date <= ?)", today, thirtyDaysLater).
			Order(gorm.Expr("CASE WHEN expiry_date < ? THEN 0 ELSE 1 END, expiry_date ASC", today)).
			Limit(20).
			Find(&docs)

		result := make([]ExpiringDoc, len(docs))
		for i, d := range docs {
			status := "expiring"
			if d.ExpiryDate < today {
				status = "expired"
			}
			result[i] = ExpiringDoc{Document: d, Status: status}
		}

		c.JSON(http.StatusOK, result)
	}
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
