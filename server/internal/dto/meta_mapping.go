package dto

// MetaMappingListResponse 元数据映射列表响应
type MetaMappingListResponse struct {
	TemplateKey  string            `json:"template_key"`
	TemplateName string            `json:"template_name"`
	Mappings     []MetaMappingItem `json:"mappings"`
}

// MetaMappingItem 元数据映射项
type MetaMappingItem struct {
	ID            uint   `json:"id"`
	SourceField   string `json:"source_field"`
	TargetField   string `json:"target_field"`
	FieldType     string `json:"field_type"`
	TransformRule string `json:"transform_rule"`
	IsActive      bool   `json:"is_active"`
	IsSystem      bool   `json:"is_system"`
	SortOrder     int    `json:"sort_order"`
}

type CreateMetaMappingRequest struct {
	TemplateKey   string `json:"template_key" binding:"required"`
	TemplateName  string `json:"template_name"` // 兼容旧前端传值，可忽略
	SourceField   string `json:"source_field" binding:"required"`
	TargetField   string `json:"target_field" binding:"required"`
	FieldType     string `json:"field_type" binding:"required"`
	TransformRule string `json:"transform_rule"`
	SortOrder     int    `json:"sort_order"`
}

type UpdateMetaMappingRequest struct {
	SourceField   string `json:"source_field"`
	TargetField   string `json:"target_field"`
	FieldType     string `json:"field_type"`
	TransformRule string `json:"transform_rule"`
	IsActive      *bool  `json:"is_active"`
	SortOrder     *int   `json:"sort_order"`
}

