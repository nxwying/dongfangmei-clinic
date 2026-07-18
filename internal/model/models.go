// internal/model/models.go
// Placeholder models — will be fully defined in subsequent tasks.
package model

import "time"

type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string `gorm:"size:256;not null" json:"-"`
	RealName     string `gorm:"size:64" json:"real_name"`
	RoleID       uint   `json:"role_id"`
	Role         Role   `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Phone        string `gorm:"size:20" json:"phone"`
	Status       string `gorm:"size:20;default:active" json:"status"`
	LastLoginAt  int64  `gorm:"default:0" json:"last_login_at"`
}

type Role struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:32;not null" json:"name"`
	Description string `gorm:"size:256" json:"description"`
	Permissions []byte `gorm:"type:jsonb" json:"permissions"`
}

type Customer struct {
	BaseModel
	Name       string      `gorm:"size:100;not null" json:"name"`
	Phone      string      `gorm:"uniqueIndex;size:20;not null" json:"phone"`
	Gender     int8        `gorm:"default:0" json:"gender"`
	Birthday   string      `gorm:"size:10" json:"birthday,omitempty"`
	IDCard     string      `gorm:"size:18" json:"id_card,omitempty"`
	Source     string      `gorm:"size:50" json:"source,omitempty"`
	WechatID   string      `gorm:"size:100" json:"wechat_id,omitempty"`
	Remark     string      `gorm:"type:text" json:"remark,omitempty"`
	Tags       string      `gorm:"type:jsonb;default:'[]'" json:"tags,omitempty"`
	Status     string      `gorm:"size:20;default:potential" json:"status"`
	Membership *Membership `gorm:"foreignKey:CustomerID" json:"membership,omitempty"`
}

type Membership struct {
	BaseModel
	CustomerID     uint    `gorm:"uniqueIndex;not null" json:"customer_id"`
	Level          string  `gorm:"size:20;default:regular" json:"level"`
	Balance        float64 `gorm:"type:decimal(12,2);default:0" json:"balance"`
	GiftBalance    float64 `gorm:"type:decimal(12,2);default:0" json:"gift_balance"`
	TotalRecharged float64 `gorm:"type:decimal(12,2);default:0" json:"total_recharged"`
	TotalConsumed  float64 `gorm:"type:decimal(12,2);default:0" json:"total_consumed"`
	OpenedAt       int64   `json:"opened_at"`
}

type MemberPackage struct {
	BaseModel
	CustomerID    uint       `gorm:"not null"`
	TemplateName  string     `gorm:"size:128"`
	TotalSessions int        `gorm:"default:0"`
	UsedSessions  int        `gorm:"default:0"`
	TotalAmount   float64    `gorm:"type:decimal(12,2)"`
	PaidAmount    float64    `gorm:"type:decimal(12,2);default:0"`
	ActivatedAt   *time.Time `json:"activated_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	Status        string     `gorm:"size:20;default:active"`
}

type PackageRedemption struct {
	BaseModel
	MemberPackageID uint   `gorm:"not null"`
	AppointmentID   *uint  `json:"appointment_id"`
	SessionIndex    int    `gorm:"default:1"`
	Remark          string `gorm:"size:256"`
}


// BackupRecord tracks backup files and cloud upload status
type BackupRecord struct {
	BaseModel
	Filename     string `gorm:"size:256" json:"filename"`
	FileSize     int64  `json:"file_size"`
	FilePath     string `gorm:"size:512" json:"file_path"`
	BackupType   string `gorm:"size:20;default:manual" json:"backup_type"`
	Status       string `gorm:"size:20;default:success" json:"status"`
	CloudStatus  string `gorm:"size:20;default:pending" json:"cloud_status"`
	CloudURL     string `gorm:"size:512" json:"cloud_url"`
	ErrorMessage string `gorm:"size:1024" json:"error_message"`
}

type Appointment struct {
	BaseModel
	CustomerID   uint       `gorm:"not null;index" json:"customer_id"`
	Customer     *Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ConsultantID *uint      `json:"consultant_id"`
	DoctorID     *uint      `json:"doctor_id"`
	Date         time.Time  `gorm:"not null;index" json:"date"`
	TimeSlot     string     `gorm:"size:20" json:"time_slot"`
	Items        string     `gorm:"type:text" json:"items"`
	Status       string     `gorm:"size:20;default:booked" json:"status"`
	Remark       string     `gorm:"size:512" json:"remark"`
	Duration     int        `gorm:"default:30" json:"duration"`
	CreatedBy    uint       `gorm:"default:0" json:"created_by"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type Order struct {
	BaseModel
	OrderNo        string      `gorm:"uniqueIndex;size:50;not null" json:"order_no"`
	CustomerID     uint        `gorm:"not null;index" json:"customer_id"`
	Customer       *Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	TotalAmount    float64     `gorm:"type:decimal(12,2);not null" json:"total_amount"`
	DiscountAmount float64     `gorm:"type:decimal(12,2);default:0" json:"discount_amount"`
	CommissionAmount float64    `gorm:"type:decimal(12,2);default:0" json:"commission_amount"`
	CostAmount float64          `gorm:"type:decimal(12,2);default:0" json:"cost_amount"`
	FinalAmount    float64     `gorm:"type:decimal(12,2);not null" json:"final_amount"`
	PaidAmount     float64     `gorm:"type:decimal(12,2);default:0" json:"paid_amount"`
	Status         string      `gorm:"size:20;default:pending" json:"status"`
	Remark         string      `gorm:"type:text" json:"remark,omitempty"`
	CreatedBy      uint        `json:"created_by"`
	PaidAt         *int64      `json:"paid_at,omitempty"`
	Items          []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	Payments       []Payment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
}

type OrderItem struct {
	BaseModel
	OrderID   uint    `gorm:"not null;index" json:"order_id"`
	ItemType  string  `gorm:"size:32" json:"item_type"`
	ItemID    uint    `gorm:"default:0" json:"item_id"`
	ItemName  string  `gorm:"size:128" json:"item_name"`
	Quantity  int     `gorm:"default:1" json:"quantity"`
	UnitPrice float64 `gorm:"type:decimal(12,2)" json:"unit_price"`
	Subtotal  float64 `gorm:"type:decimal(12,2)" json:"subtotal"`
}

type Payment struct {
	BaseModel
	OrderID   uint    `gorm:"not null;index" json:"order_id"`
	PayMethod string  `gorm:"size:20;not null" json:"pay_method"`
	Amount    float64 `gorm:"type:decimal(12,2);not null" json:"amount"`
	Reference string  `gorm:"size:100" json:"reference,omitempty"`
}

type AuditLog struct {
	BaseModel
	UserID   uint   `gorm:"index" json:"user_id"`
	Action   string `gorm:"size:64;not null" json:"action"`
	Target   string `gorm:"size:64" json:"target"`
	TargetID *uint  `json:"target_id"`
	Detail   string `gorm:"type:text" json:"detail"`
}

type FollowUp struct {
	BaseModel
	CustomerID  uint   `gorm:"not null;index" json:"customer_id"`
MedicalRecordID *uint `json:"medical_record_id,omitempty"`
	CreatedBy   uint   `json:"created_by"`
	CreatedName string `gorm:"size:100" json:"created_name,omitempty"`
	Content     string `gorm:"type:text;not null" json:"content"`
	Method      string `gorm:"size:20" json:"method,omitempty"`
	NextAt      string `gorm:"size:10" json:"next_at,omitempty"`
}

type TreatmentItem struct {
	BaseModel
	Name        string  `gorm:"size:128;not null" json:"name"`
	Category    string  `gorm:"size:64" json:"category"`
	Description string  `gorm:"type:text" json:"description"`
	Price       float64 `gorm:"type:decimal(12,2)" json:"price"`
	DurationMin int     `gorm:"default:60" json:"duration"`
	IsActive    bool    `gorm:"default:true" json:"-"`

}

type Expense struct {
	BaseModel
	Type     string  `gorm:"size:20;not null" json:"type"`
	Category string  `gorm:"size:50" json:"category"`
	Amount   float64 `gorm:"type:decimal(12,2);not null" json:"amount"`
	Note     string  `gorm:"type:text" json:"note"`
	Date     string  `gorm:"size:10" json:"date"`
}

type PackageTemplate struct {
	BaseModel
	Name           string  `gorm:"size:128;not null"`
	Description    string  `gorm:"type:text"`
	Items          string  `gorm:"type:jsonb"`
	TotalPrice     float64 `gorm:"type:decimal(12,2)" json:"price"`
	DurationMonths int     `gorm:"default:12" json:"duration_months"`
	SessionCount   int     `gorm:"default:1" json:"total_sessions"`
	IsActive       bool    `gorm:"default:true" json:"-"`
}

// TreatmentRecord stores patient treatment history.
type TreatmentRecord struct {
	BaseModel
	CustomerID    uint      `gorm:"not null;index" json:"customer_id"`
	Customer      *Customer `json:"customer,omitempty"`
	OrderID       *uint     `json:"order_id"`
	Items         string    `gorm:"type:text" json:"items"`
	DoctorID      *uint     `json:"doctor_id"`
	DoctorName    string    `gorm:"size:64" json:"doctor_name"`
	Notes         string    `gorm:"type:text" json:"notes"`
	Dosage        string    `gorm:"size:255" json:"dosage"`
	Products      string    `gorm:"type:text" json:"products"`
	TreatmentDate string    `gorm:"size:10" json:"treatment_date"`
}

// FollowUpTask tracks post-treatment and other follow-ups.
type FollowUpTask struct {
	AssignedTo  *uint      `json:"assigned_to,omitempty"`
	Result      string     `gorm:"type:text" json:"result,omitempty"`
	ResultStatus string    `gorm:"size:30" json:"result_status,omitempty"`
	BaseModel
	CustomerID  uint       `gorm:"not null;index" json:"customer_id"`
MedicalRecordID *uint `json:"medical_record_id,omitempty"`
	Customer    *Customer  `json:"customer,omitempty"`
	Type        string     `gorm:"size:30" json:"type"`       // post_treatment, birthday, dormant, custom
	DueDate     string     `gorm:"size:10;not null" json:"due_date"`
	Status      string     `gorm:"size:20;default:pending" json:"status"` // pending, completed, skipped
	Note        string     `gorm:"type:text" json:"note"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedBy   uint       `json:"created_by"`
}

// InventoryItem tracks stock of drugs and consumables.
type InventoryItem struct {
	BaseModel
	Name       string  `gorm:"size:128;not null" json:"name"`
	Category   string  `gorm:"size:50" json:"category"`
	Quantity   float64 `gorm:"type:decimal(12,2);default:0" json:"quantity"`
	Unit       string  `gorm:"size:20" json:"unit"`
	MinStock   float64 `gorm:"type:decimal(12,2);default:0" json:"min_stock"`
	Price      float64 `gorm:"type:decimal(12,2);default:0" json:"price"`
	ExpiryDate string  `gorm:"size:10" json:"expiry_date"`
	Supplier   string  `gorm:"size:100" json:"supplier"`
	Remark     string  `gorm:"type:text" json:"remark"`
}

// InventoryLog records stock movements.
type InventoryLog struct {
	BaseModel
	ItemID       uint           `gorm:"not null;index" json:"item_id"`
	Item         *InventoryItem `json:"item,omitempty"`
	Type         string         `gorm:"size:10" json:"type"` // in, out
	Quantity     float64        `gorm:"type:decimal(12,2)" json:"quantity"`
	BalanceAfter float64        `gorm:"type:decimal(12,2)" json:"balance_after"`
	Reference    string         `gorm:"size:100" json:"reference"`
	Note         string         `gorm:"type:text" json:"note"`
	CreatedBy    uint           `json:"created_by"`
}

// MedicalRecord stores structured medical documents per patient.
type MedicalRecord struct {
	BaseModel
	CustomerID uint      `gorm:"not null;index" json:"customer_id"`
	Customer   *Customer `json:"customer,omitempty"`
	RecordType string    `gorm:"size:50;not null" json:"record_type"`
	RecordDate string    `gorm:"size:10;not null" json:"record_date"`
	CreatedBy  uint      `json:"created_by"`
	DoctorName string    `gorm:"size:64" json:"doctor_name"`
	TemplateID  *uint      `json:"template_id,omitempty"`
	Content    string    `gorm:"type:jsonb" json:"content"`
	Status     string    `gorm:"size:20;default:draft" json:"status"`
}

// MedicalRecordTemplate — 电子病历模板（预设项目字段）
type MedicalRecordTemplate struct {
	BaseModel
	Name          string `gorm:"size:200;not null" json:"name"`
	ProcedureName string `gorm:"size:200" json:"procedure_name"`
	Category      string `gorm:"size:50" json:"category"`
	Fields        string `gorm:"type:jsonb" json:"fields"`
	Description   string `gorm:"type:text" json:"description"`
	IsActive      bool   `gorm:"default:true" json:"is_active"`
	SortOrder     int    `gorm:"default:0" json:"sort_order"`
}

// Document stores uploaded regulatory documents
type Document struct {
	BaseModel
	DocType     string  `gorm:"size:20;not null;index" json:"doc_type"`
	Title       string  `gorm:"size:200;not null" json:"title"`
	FileName    string  `gorm:"size:255;not null" json:"file_name"`
	FilePath    string  `gorm:"size:500;not null" json:"-"`
	FileSize    int64   `gorm:"not null;default:0" json:"file_size"`
	FileType    string  `gorm:"size:100;not null" json:"file_type"`
	ProductName string  `gorm:"size:200;default:''" json:"product_name"`
	Supplier    string  `gorm:"size:200;default:''" json:"supplier"`
	SerialNo    string  `gorm:"size:100;default:''" json:"serial_no"`
	IssueDate   string  `gorm:"size:10;default:''" json:"issue_date"`
	ExpiryDate  string  `gorm:"size:10;default:''" json:"expiry_date"`
	Amount      float64 `gorm:"type:decimal(12,2);default:0" json:"amount"`
	Remark      string  `gorm:"type:text" json:"remark"`
	CreatedBy   uint    `json:"created_by"`
}

// Photo stores before/after treatment photos for medical aesthetics.
type Photo struct {
	BaseModel
	CustomerID  uint   `gorm:"not null;index" json:"customer_id"`
MedicalRecordID *uint `json:"medical_record_id,omitempty"`
	TreatmentID *uint  `json:"treatment_id,omitempty"`
	PhotoType   string `gorm:"size:10;not null" json:"photo_type"`   // before / after
	BodyPart    string `gorm:"size:50" json:"body_part"`              // face, nose, eyes, etc.
	FilePath    string `gorm:"size:500;not null" json:"-"`
	FileName    string `gorm:"size:255;not null" json:"file_name"`
	FileSize    int64  `gorm:"not null;default:0" json:"file_size"`
	FileType    string `gorm:"size:100;not null" json:"file_type"`
	CreatedBy   uint   `json:"created_by"`
}

// TagRule stores auto-tagging rules based on customer behavior.
type TagRule struct {
	BaseModel
	Name       string `gorm:"size:100;not null" json:"name"`
	Conditions string `gorm:"type:text;not null" json:"conditions"`  // JSON array of conditions
	Tag        string `gorm:"size:50;not null" json:"tag"`
	IsActive   bool   `gorm:"default:true" json:"is_active"`
	ApplyOrder int    `gorm:"default:0" json:"apply_order"`
}

// ========== 绩效提成引擎 ==========

// CommissionRule stores commission configuration per position/procedure.
type CommissionRule struct {
	BaseModel
	Name         string  `gorm:"size:100;not null" json:"name"`
	Role         string  `gorm:"size:30;not null" json:"role"`       // consultant / doctor / nurse
	RuleType     string  `gorm:"size:30;not null" json:"rule_type"`   // percentage / fixed / tiered
	Procedure    string  `gorm:"size:100" json:"procedure"`            // 适用项目，空=全部
	Rate         float64 `gorm:"type:decimal(5,2)" json:"rate"`        // 百分比或固定金额
	TierMin      float64 `gorm:"type:decimal(12,2)" json:"tier_min"`   // 阶梯下限
	TierMax      float64 `gorm:"type:decimal(12,2)" json:"tier_max"`   // 阶梯上限
	IsActive     bool    `gorm:"default:true" json:"is_active"`
}

// CommissionResult stores monthly commission calculation results.
type CommissionResult struct {
	BaseModel
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	RealName    string  `gorm:"size:64" json:"real_name"`
	YearMonth   string  `gorm:"size:7;not null;index" json:"year_month"` // 2026-07
	TotalRevenue float64 `gorm:"type:decimal(12,2)" json:"total_revenue"`
	TotalCommission float64 `gorm:"type:decimal(12,2)" json:"total_commission"`
	Details     string  `gorm:"type:jsonb" json:"details"`
	Status      string  `gorm:"size:20;default:draft" json:"status"` // draft / confirmed / paid
}

// ========== KPI目标管理 ==========

// KpiTarget stores monthly KPI targets per user.
type KpiTarget struct {
	BaseModel
	UserID        uint    `gorm:"not null;index" json:"user_id"`
	YearMonth     string  `gorm:"size:7;not null;index" json:"year_month"`
	RevenueTarget float64 `gorm:"type:decimal(12,2)" json:"revenue_target"`
	OrderTarget   int     `gorm:"default:0" json:"order_target"`
	VisitTarget   int     `gorm:"default:0" json:"visit_target"`
	FollowupTarget int    `gorm:"default:0" json:"followup_target"`
	NewCustomerTarget int `gorm:"default:0" json:"new_customer_target"`
}

// ========== 培训认证 ==========

// Training records staff training and certification.
type Training struct {
	BaseModel
	UserID          uint    `gorm:"not null;index" json:"user_id"`
	Title           string  `gorm:"size:200;not null" json:"title"`
	Trainer         string  `gorm:"size:100" json:"trainer"`
	Date            string  `gorm:"size:10" json:"date"`
	Hours           int     `gorm:"default:1" json:"hours"`
	CertExpiry      string  `gorm:"size:10" json:"cert_expiry"`
	Notes           string  `gorm:"type:text" json:"notes"`
	Category        string  `gorm:"size:30;default:internal" json:"category"`
	Location        string  `gorm:"size:200" json:"location"`
	Cost            float64 `gorm:"default:0" json:"cost"`
	ExamScore       float64 `gorm:"default:0" json:"exam_score"`
	PracticalScore  float64 `gorm:"default:0" json:"practical_score"`
	Passed          string  `gorm:"size:20;default:pending" json:"passed"`
	CertNumber      string  `gorm:"size:100" json:"cert_number"`
	CertIssuer      string  `gorm:"size:200" json:"cert_issuer"`
	CertImage       string  `gorm:"size:500" json:"cert_image"`
	MaterialURL     string  `gorm:"size:500" json:"material_url"`
	Satisfaction    int     `gorm:"default:0" json:"satisfaction"`
	IsMandatory     bool    `gorm:"default:false" json:"is_mandatory"`
	RoleRequired    string  `gorm:"size:100" json:"role_required"`
	Points          int     `gorm:"default:0" json:"points"`
}

// TrainingPlan — 培训计划/必修课
type TrainingPlan struct {
	BaseModel
	Title        string `gorm:"size:200;not null" json:"title"`
	Category     string `gorm:"size:30;default:internal" json:"category"`
	RoleRequired string `gorm:"size:100" json:"role_required"`
	TargetHours  int    `gorm:"default:0" json:"target_hours"`
	Required     bool   `gorm:"default:false" json:"required"`
	Deadline     string `gorm:"size:10" json:"deadline"`
	Description  string `gorm:"type:text" json:"description"`
	Status       string `gorm:"size:20;default:active" json:"status"`
}

// ========== 电子知情同意书 ==========
type ConsentForm struct {
	BaseModel
	CustomerID    uint   `gorm:"not null;index" json:"customer_id"`
	TreatmentID   *uint  `json:"treatment_id,omitempty"`
	ProcedureName string `gorm:"size:200;not null" json:"procedure_name"`
	DoctorName    string `gorm:"size:100" json:"doctor_name"`
	Content       string `gorm:"type:text;not null" json:"content"`
	PatientSign   string `gorm:"type:text" json:"patient_sign"`
	DoctorSign    string `gorm:"type:text" json:"doctor_sign"`
	SignDate      string `gorm:"size:10" json:"sign_date"`
	Status        string `gorm:"size:20;default:draft" json:"status"`
	CreatedBy     uint   `json:"created_by"`
}

// ========== 高值耗材批号管理 ==========
type InventoryBatch struct {
	BaseModel
	ItemID       uint    `gorm:"not null;index" json:"item_id"`
	BatchNo      string  `gorm:"size:100;not null" json:"batch_no"`
	Supplier     string  `gorm:"size:200" json:"supplier"`
	Quantity     float64 `gorm:"type:decimal(12,2)" json:"quantity"`
	UsedQty      float64 `gorm:"type:decimal(12,2);default:0" json:"used_qty"`
	ExpiryDate   string  `gorm:"size:10" json:"expiry_date"`
	PurchaseDate string  `gorm:"size:10" json:"purchase_date"`
	Unit         string  `gorm:"size:20" json:"unit"`
	UnitPrice    float64 `gorm:"type:decimal(12,2)" json:"unit_price"`
	Remark       string  `gorm:"type:text" json:"remark"`
}

type BatchUsage struct {
	BaseModel
	BatchID     uint    `gorm:"not null;index" json:"batch_id"`
	TreatmentID uint    `json:"treatment_id"`
	CustomerID  uint    `json:"customer_id"`
MedicalRecordID *uint `json:"medical_record_id,omitempty"`
	Quantity    float64 `gorm:"type:decimal(12,2)" json:"quantity"`
	DoctorName  string  `gorm:"size:100" json:"doctor_name"`
	Note        string  `gorm:"type:text" json:"note"`
}
