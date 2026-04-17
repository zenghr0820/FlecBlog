import request from '@/utils/request'
import type {
  MetaMappingTemplate,
  MetaMappingListResponse,
  CreateMetaMappingTemplateRequest,
  UpdateMetaMappingTemplateRequest,
  CreateMetaMappingRequest,
  UpdateMetaMappingRequest
} from '@/types/metaMapping'

// 获取映射模版列表
export const getTemplates = (): Promise<MetaMappingTemplate[]> =>
  request.get('/admin/meta-mappings/templates')

// 创建映射模版
export const createTemplate = (data: CreateMetaMappingTemplateRequest): Promise<MetaMappingTemplate> =>
  request.post('/admin/meta-mappings/templates', data)

// 更新映射模版
export const updateTemplate = (id: number, data: UpdateMetaMappingTemplateRequest): Promise<MetaMappingTemplate> =>
  request.put(`/admin/meta-mappings/templates/${id}`, data)

// 删除映射模版
export const deleteTemplate = (id: number): Promise<void> =>
  request.delete(`/admin/meta-mappings/templates/${id}`)

// 获取指定模版的映射配置
export const getMappingsByTemplate = (templateKey: string): Promise<MetaMappingListResponse> =>
  request.get(`/admin/meta-mappings/${templateKey}`)

// 创建映射配置
export const createMapping = (data: CreateMetaMappingRequest): Promise<any> =>
  request.post('/admin/meta-mappings', data)

// 更新映射配置
export const updateMapping = (id: number, data: UpdateMetaMappingRequest): Promise<any> =>
  request.put(`/admin/meta-mappings/${id}`, data)

// 删除映射配置
export const deleteMapping = (id: number): Promise<void> =>
  request.delete(`/admin/meta-mappings/${id}`)

// 切换映射状态
export const toggleMappingStatus = (id: number): Promise<{ is_active: boolean }> =>
  request.put(`/admin/meta-mappings/${id}/toggle`)
