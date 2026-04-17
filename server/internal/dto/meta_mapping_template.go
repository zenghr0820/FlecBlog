package dto

// MetaMappingTemplateItem 元数据映射模版信息
type MetaMappingTemplateItem struct {
	ID           uint   `json:"id"`
	TemplateKey  string `json:"template_key"`
	TemplateName string `json:"template_name"`
	Description  string `json:"description"`
	MappingCount int    `json:"mapping_count"`
}

type CreateMetaMappingTemplateRequest struct {
	TemplateKey  string `json:"template_key" binding:"required"`
	TemplateName string `json:"template_name" binding:"required"`
	Description  string `json:"description"`
}

type UpdateMetaMappingTemplateRequest struct {
	TemplateName string `json:"template_name"`
	Description  string `json:"description"`
}

