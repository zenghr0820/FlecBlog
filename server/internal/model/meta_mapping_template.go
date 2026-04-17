package model

import "time"

// MetaMappingTemplate 元数据映射模版（一个模版包含多条 meta mappings）
//
// TemplateKey 用作模版唯一标识（导入/导出时选择的 key），例如：vuepress、hexo、my_legacy_blog
type MetaMappingTemplate struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	TemplateKey  string    `gorm:"size:50;not null;uniqueIndex" json:"template_key"` // 历史沿用 template
	TemplateName string    `gorm:"size:100;not null" json:"template_name"`           // 历史沿用 template_name
	Description  string    `gorm:"type:text" json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 兼容历史表名（不做破坏性迁移）
func (MetaMappingTemplate) TableName() string { return "meta_mapping_templates" }

