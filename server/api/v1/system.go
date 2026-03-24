package v1

import (
	"flec_blog/internal/service"
	"flec_blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// SystemController 系统信息控制器。
type SystemController struct {
	systemService *service.SystemService
}

// NewSystemController 创建系统信息控制器。
func NewSystemController(systemService *service.SystemService) *SystemController {
	return &SystemController{systemService: systemService}
}

// GetSystemStatic 获取系统静态信息。
func (h *SystemController) GetSystemStatic(c *gin.Context) {
	response.Success(c, h.systemService.GetStaticInfo())
}

// GetSystemDynamic 获取系统动态信息。
func (h *SystemController) GetSystemDynamic(c *gin.Context) {
	response.Success(c, h.systemService.GetDynamicInfo())
}
