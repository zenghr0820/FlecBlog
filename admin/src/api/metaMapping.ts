import request from '@/utils/request'

// 获取映射模版列表
export const getTemplates = () => request.get('/admin/meta-mappings/templates')

// 创建映射模版
export const createTemplate = (data: any) =>
  request.post('/admin/meta-mappings/templates', data)

// 更新映射模版
export const updateTemplate = (id: number, data: any) =>
  request.put(`/admin/meta-mappings/templates/${id}`, data)

// 删除映射模版
export const deleteTemplate = (id: number) =>
  request.delete(`/admin/meta-mappings/templates/${id}`)

// 获取指定模版的映射配置
export const getMappingsByTemplate = (templateKey: string) =>
  request.get(`/admin/meta-mappings/${templateKey}`)

// 创建映射配置
export const createMapping = (data: any) => request.post('/admin/meta-mappings', data)

// 更新映射配置
export const updateMapping = (id: number, data: any) =>
  request.put(`/admin/meta-mappings/${id}`, data)

// 删除映射配置
export const deleteMapping = (id: number) => request.delete(`/admin/meta-mappings/${id}`)

// 切换映射状态
export const toggleMappingStatus = (id: number) =>
  request.put(`/admin/meta-mappings/${id}/toggle`)


