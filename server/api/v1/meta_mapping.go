package v1

import (
	"fmt"

	"flec_blog/internal/dto"
	"flec_blog/internal/service"
	"flec_blog/pkg/errcode"
	"flec_blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// MetaMappingController 元数据映射控制器
type MetaMappingController struct {
	metaSvc *service.MetaMappingService
}

func NewMetaMappingController(metaSvc *service.MetaMappingService) *MetaMappingController {
	return &MetaMappingController{metaSvc: metaSvc}
}

// ListTemplates
//
//	@Summary		获取元数据映射模版列表
//	@Tags			Meta映射
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=[]dto.MetaMappingTemplateItem}
//	@Router			/admin/meta-mappings/templates [get]
func (c *MetaMappingController) ListTemplates(ctx *gin.Context) {
	items, err := c.metaSvc.ListTemplates()
	if err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails("获取模版列表失败"))
		return
	}
	response.Success(ctx, items)
}

// CreateTemplate
//
//	@Summary		创建元数据映射模版
//	@Tags			Meta映射
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateMetaMappingTemplateRequest	true	"模版信息"
//	@Success		200		{object}	response.Response{data=model.MetaMappingTemplate}
//	@Router			/admin/meta-mappings/templates [post]
func (c *MetaMappingController) CreateTemplate(ctx *gin.Context) {
	var req dto.CreateMetaMappingTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidateFailed(ctx, "请求参数错误")
		return
	}
	t, err := c.metaSvc.CreateTemplate(&req)
	if err != nil {
		response.Error(ctx, errcode.NewError(500, err.Error()))
		return
	}
	response.Success(ctx, t)
}

// UpdateTemplate
//
//	@Summary		更新元数据映射模版
//	@Tags			Meta映射
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		uint								true	"模版ID"
//	@Param			request	body		dto.UpdateMetaMappingTemplateRequest	true	"模版信息"
//	@Success		200		{object}	response.Response{data=model.MetaMappingTemplate}
//	@Router			/admin/meta-mappings/templates/{id} [put]
func (c *MetaMappingController) UpdateTemplate(ctx *gin.Context) {
	id := ctx.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		response.ValidateFailed(ctx, "无效的模版ID")
		return
	}
	var req dto.UpdateMetaMappingTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidateFailed(ctx, "请求参数错误")
		return
	}
	t, err := c.metaSvc.UpdateTemplate(idUint, &req)
	if err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails(err.Error()))
		return
	}
	response.Success(ctx, t)
}

// DeleteTemplate
//
//	@Summary		删除元数据映射模版
//	@Tags			Meta映射
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		uint	true	"模版ID"
//	@Success		200	{object}	response.Response
//	@Router			/admin/meta-mappings/templates/{id} [delete]
func (c *MetaMappingController) DeleteTemplate(ctx *gin.Context) {
	id := ctx.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		response.ValidateFailed(ctx, "无效的模版ID")
		return
	}
	if err := c.metaSvc.DeleteTemplate(idUint); err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails(err.Error()))
		return
	}
	response.Success(ctx, nil)
}

// GetMappingsByTemplate
//
//	@Summary		获取模版的字段映射
//	@Tags			Meta映射
//	@Produce		json
//	@Security		BearerAuth
//	@Param			templateKey	path		string	true	"模版Key"
//	@Success		200			{object}	response.Response{data=dto.MetaMappingListResponse}
//	@Router			/admin/meta-mappings/{templateKey} [get]
func (c *MetaMappingController) GetMappingsByTemplate(ctx *gin.Context) {
	key := ctx.Param("templateKey")
	if key == "" {
		response.ValidateFailed(ctx, "模版Key不能为空")
		return
	}
	mappings, err := c.metaSvc.GetMappingsByTemplateKey(key)
	if err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails("获取映射失败"))
		return
	}

	items := make([]dto.MetaMappingItem, len(mappings))
	for i, m := range mappings {
		items[i] = dto.MetaMappingItem{
			ID:            m.ID,
			SourceField:   m.SourceField,
			TargetField:   m.TargetField,
			FieldType:     m.FieldType,
			TransformRule: m.TransformRule,
			IsActive:      m.IsActive,
			IsSystem:      m.IsSystem,
			SortOrder:     m.SortOrder,
		}
	}

	templateName := key
	if len(mappings) > 0 {
		templateName = mappings[0].TemplateName
	}
	response.Success(ctx, dto.MetaMappingListResponse{
		TemplateKey:  key,
		TemplateName: templateName,
		Mappings:     items,
	})
}

// CreateMapping
//
//	@Summary		创建字段映射
//	@Tags			Meta映射
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateMetaMappingRequest	true	"字段映射"
//	@Success		200		{object}	response.Response{data=model.MetaMapping}
//	@Router			/admin/meta-mappings [post]
func (c *MetaMappingController) CreateMapping(ctx *gin.Context) {
	var req dto.CreateMetaMappingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidateFailed(ctx, "请求参数错误")
		return
	}
	mapping, err := c.metaSvc.CreateMapping(&req)
	if err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails(err.Error()))
		return
	}
	response.Success(ctx, mapping)
}

// UpdateMapping
//
//	@Summary		更新字段映射
//	@Tags			Meta映射
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		uint						true	"映射ID"
//	@Param			request	body		dto.UpdateMetaMappingRequest	true	"字段映射"
//	@Success		200		{object}	response.Response{data=model.MetaMapping}
//	@Router			/admin/meta-mappings/{id} [put]
func (c *MetaMappingController) UpdateMapping(ctx *gin.Context) {
	id := ctx.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		response.ValidateFailed(ctx, "无效的映射ID")
		return
	}
	var req dto.UpdateMetaMappingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidateFailed(ctx, "请求参数错误")
		return
	}
	mapping, err := c.metaSvc.UpdateMapping(idUint, &req)
	if err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails(err.Error()))
		return
	}
	response.Success(ctx, mapping)
}

// DeleteMapping
//
//	@Summary		删除字段映射
//	@Tags			Meta映射
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		uint	true	"映射ID"
//	@Success		200	{object}	response.Response
//	@Router			/admin/meta-mappings/{id} [delete]
func (c *MetaMappingController) DeleteMapping(ctx *gin.Context) {
	id := ctx.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		response.ValidateFailed(ctx, "无效的映射ID")
		return
	}
	if err := c.metaSvc.DeleteMapping(idUint); err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails(err.Error()))
		return
	}
	response.Success(ctx, nil)
}

// ToggleMappingStatus
//
//	@Summary		切换字段映射状态
//	@Tags			Meta映射
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		uint	true	"映射ID"
//	@Success		200	{object}	response.Response{data=map[string]bool}
//	@Router			/admin/meta-mappings/{id}/toggle [put]
func (c *MetaMappingController) ToggleMappingStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		response.ValidateFailed(ctx, "无效的映射ID")
		return
	}
	isActive, err := c.metaSvc.ToggleMappingStatus(idUint)
	if err != nil {
		response.Error(ctx, errcode.ServerError.WithDetails(err.Error()))
		return
	}
	response.Success(ctx, map[string]bool{"is_active": isActive})
}
