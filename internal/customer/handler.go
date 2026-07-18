package customer

import (
	"net/http"
	"strconv"
	"time"

	"clinic-mgmt/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"clinic-mgmt/internal/system"
	"gorm.io/gorm/clause"
)

// ---------- Request structs ----------

type CreateCustomerRequest struct {
	Name      string     `json:"name" binding:"required"`
	Phone     string     `json:"phone" binding:"required"`
	Gender    int8       `json:"gender"`
	Birthday string `json:"birthday"`
	IDCard    string `json:"id_card"`
	Source    string     `json:"source"`
	WechatID  string     `json:"wechat_id,omitempty"`
	Tags      string     `json:"tags,omitempty"`
	Remark    string     `json:"remark"`
}

type UpdateCustomerRequest struct {
	Name      string     `json:"name"`
	Phone     string     `json:"phone"`
	Gender    int8       `json:"gender"`
	IDCard    string     `json:"id_card"`
	Birthday string `json:"birthday"`
	Source    string     `json:"source"`
	Remark    string     `json:"remark"`
}

type CreateFollowUpRequest struct {
	Content string     `json:"content" binding:"required"`
	Method  string     `json:"method"`
	NextAt  string `json:"next_at"`
}

// ---------- Customer handlers ----------

func ListCustomers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		keyword := c.Query("keyword")
		status := c.Query("status")
		source := c.Query("source")

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		query := db.Model(&model.Customer{})

		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("name ILIKE ? OR phone ILIKE ?", like, like)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if source != "" {
			query = query.Where("source = ?", source)
		}
		if startDate := c.Query("start_date"); startDate != "" {
			if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
				query = query.Where("created_at >= ?", parsed)
			}
		}
		if endDate := c.Query("end_date"); endDate != "" {
			if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
				query = query.Where("created_at < ?", parsed.Add(24*time.Hour))
			}
		}

		var total int64
		query.Count(&total)

		var customers []model.Customer
		result := query.Preload("Membership").
			Order("created_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&customers)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询客户列表失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":      customers,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func CreateCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateCustomerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入姓名和手机号"})
			return
		}

		customer := model.Customer{
			Name:     req.Name,
			Phone:    req.Phone,
			Gender:   req.Gender,
			Birthday: req.Birthday,
			IDCard:   req.IDCard,
			Source:   req.Source,
			Remark:   req.Remark,
			WechatID: req.WechatID,
			Tags:     req.Tags,
		}

		if err := db.Create(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建客户失败"})
			return
		}

		system.WriteAuditLog(db, c, "create", "customer", customer.ID, "创建客户 "+customer.Name+" "+customer.Phone)
		c.JSON(http.StatusCreated, customer)
	}
}

func GetCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var customer model.Customer
		result := db.Preload("Membership").First(&customer, id)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
			return
		}

		system.WriteAuditLog(db, c, "update", "customer", uint(id), "更新客户信息")
		c.JSON(http.StatusOK, customer)
	}
}

func UpdateCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var req UpdateCustomerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求数据格式错误"})
			return
		}

		updates := map[string]interface{}{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.Gender != 0 {
			updates["gender"] = req.Gender
		}
		if req.IDCard != "" {
			updates["id_card"] = req.IDCard
		}
		if req.Birthday != "" {
			updates["birthday"] = req.Birthday
		}
		if req.Source != "" {
			updates["source"] = req.Source
		}
		if req.Remark != "" {
			updates["remark"] = req.Remark
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "没有需要更新的字段"})
			return
		}

		var customer model.Customer
		if err := db.First(&customer, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
			return
		}

		if err := db.Model(&customer).Clauses(clause.Returning{}).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新客户失败"})
			return
		}

		system.WriteAuditLog(db, c, "update", "customer", uint(id), "更新客户信息")
		c.JSON(http.StatusOK, customer)
	}
}

func DeleteCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		db.Delete(&model.Customer{}, id)
		system.WriteAuditLog(db, c, "delete", "customer", uint(id), "删除客户")
		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

// ---------- FollowUp handlers ----------

func ListFollowUps(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var followUps []model.FollowUp
		db.Where("customer_id = ?", customerID).
			Order("created_at DESC").
			Find(&followUps)

		c.JSON(http.StatusOK, followUps)
	}
}

func CreateFollowUp(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
			return
		}

		var req CreateFollowUpRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入跟进内容"})
			return
		}

		createdBy := c.GetUint("user_id")

		followUp := model.FollowUp{
			CustomerID: uint(customerID),
			Content:    req.Content,
			Method:     req.Method,
			NextAt:     req.NextAt,
			CreatedBy:  createdBy,
			CreatedName: c.GetString("real_name"),
		}

		if err := db.Create(&followUp).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建跟进记录失败"})
			return
		}

		c.JSON(http.StatusCreated, followUp)
	}
}
