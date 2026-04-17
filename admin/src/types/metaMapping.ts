// MetaMapping 模块类型定义

/**
 * 映射模版信息
 */
export interface MetaMappingTemplate {
  id: number
  template_key: string
  template_name: string
  description: string
  mapping_count: number
}

/**
 * 创建模版请求
 */
export interface CreateMetaMappingTemplateRequest {
  template_key: string
  template_name: string
  description?: string
}

/**
 * 更新模版请求
 */
export interface UpdateMetaMappingTemplateRequest {
  template_name?: string
  description?: string
}

/**
 * 单个映射项
 */
export interface MetaMappingItem {
  id: number
  source_field: string
  target_field: string
  field_type: string
  transform_rule: string
  is_active: boolean
  is_system: boolean
  sort_order: number
}

/**
 * 映射列表响应（按模版分组）
 */
export interface MetaMappingListResponse {
  template_key: string
  template_name: string
  mappings: MetaMappingItem[]
}

/**
 * 创建映射请求
 */
export interface CreateMetaMappingRequest {
  template_key: string
  template_name: string
  source_field: string
  target_field: string
  field_type: string
  transform_rule?: string
  sort_order?: number
}

/**
 * 更新映射请求
 */
export interface UpdateMetaMappingRequest {
  source_field: string
  target_field: string
  field_type: string
  transform_rule?: string
  sort_order?: number
}
