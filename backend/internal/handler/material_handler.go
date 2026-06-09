package handler

import (
	"strconv"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// MaterialHandler 广告素材处理器
type MaterialHandler struct {
	materialService *service.MaterialService
}

// NewMaterialHandler 创建广告素材处理器
func NewMaterialHandler() *MaterialHandler {
	return &MaterialHandler{
		materialService: service.NewMaterialService(),
	}
}

// GetMaterials 获取广告素材列表
func (h *MaterialHandler) GetMaterials(c *gin.Context) {
	accountID := c.Query("account_id")
	materialType := c.Query("material_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	materials, total, err := h.materialService.GetMaterials(accountID, materialType, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get materials")
		return
	}

	middleware.Success(c, gin.H{
		"list":      materials,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetMaterial 获取广告素材详情
func (h *MaterialHandler) GetMaterial(c *gin.Context) {
	materialID := c.Param("id")
	if materialID == "" {
		middleware.BadRequest(c, "material id is required")
		return
	}

	material, err := h.materialService.GetMaterialByID(materialID)
	if err != nil {
		middleware.NotFound(c, "Material not found")
		return
	}

	middleware.Success(c, material)
}

// GetMaterialStats 获取素材统计
func (h *MaterialHandler) GetMaterialStats(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		middleware.BadRequest(c, "account_id is required")
		return
	}

	stats, err := h.materialService.GetMaterialStats(accountID)
	if err != nil {
		middleware.InternalError(c, "Failed to get material stats")
		return
	}

	middleware.Success(c, stats)
}
