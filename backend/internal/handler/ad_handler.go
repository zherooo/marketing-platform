package handler

import (
	"strconv"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// AdHandler 广告处理器
type AdHandler struct {
	adService       *service.AdService
	creativeService *service.CreativeService
}

// NewAdHandler 创建广告处理器
func NewAdHandler() *AdHandler {
	return &AdHandler{
		adService:       service.NewAdService(),
		creativeService: service.NewCreativeService(),
	}
}

// GetAds 获取广告列表
func (h *AdHandler) GetAds(c *gin.Context) {
	accountID := c.Query("account_id")
	adGroupID := c.Query("adgroup_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ads, total, err := h.adService.GetAds(accountID, adGroupID, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get ads")
		return
	}

	middleware.Success(c, gin.H{
		"list":      ads,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAd 获取广告详情
func (h *AdHandler) GetAd(c *gin.Context) {
	adID := c.Param("id")
	if adID == "" {
		middleware.BadRequest(c, "ad id is required")
		return
	}

	ad, err := h.adService.GetAdByID(adID)
	if err != nil {
		middleware.NotFound(c, "Ad not found")
		return
	}

	middleware.Success(c, ad)
}

// GetCreatives 获取广告创意列表
func (h *AdHandler) GetCreatives(c *gin.Context) {
	accountID := c.Query("account_id")
	adID := c.Query("ad_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	creatives, total, err := h.creativeService.GetCreatives(accountID, adID, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get creatives")
		return
	}

	middleware.Success(c, gin.H{
		"list":      creatives,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetCreative 获取广告创意详情
func (h *AdHandler) GetCreative(c *gin.Context) {
	creativeID := c.Param("id")
	if creativeID == "" {
		middleware.BadRequest(c, "creative id is required")
		return
	}

	creative, err := h.creativeService.GetCreativeByID(creativeID)
	if err != nil {
		middleware.NotFound(c, "Creative not found")
		return
	}

	middleware.Success(c, creative)
}
