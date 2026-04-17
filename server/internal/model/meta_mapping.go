package model

import "time"

// MetaMapping 元数据映射配置模型（用于导入/导出 Markdown meta 数据）
type MetaMapping struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	TemplateKey   string    `gorm:"size:50;not null;index" json:"template_key"`     // 模版 key（历史沿用 platform）
	TemplateName  string    `gorm:"size:100;not null" json:"template_name"`         // 模版显示名称（历史沿用 platform_name）
	SourceField   string    `gorm:"size:100;not null" json:"source_field"`          // 源字段名（meta key）
	TargetField   string    `gorm:"size:100;not null" json:"target_field"`          // 目标字段名
	FieldType     string    `gorm:"size:20;not null" json:"field_type"`             // string, boolean, date, array
	TransformRule string    `gorm:"type:text" json:"transform_rule"`                // JSON 规则
	IsActive      bool      `gorm:"default:true;index" json:"is_active"`
	IsSystem      bool      `gorm:"default:false" json:"is_system"`
	SortOrder     int       `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 兼容历史表名（不做破坏性迁移）
func (MetaMapping) TableName() string { return "meta_mappings" }

const (
	FieldTypeString  = "string"
	FieldTypeBoolean = "boolean"
	FieldTypeDate    = "date"
	FieldTypeArray   = "array"
)

