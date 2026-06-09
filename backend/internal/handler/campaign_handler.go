package handler

import (
	"strconv"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// CampaignHandler 广告系列处理器
type CampaignHandler struct {
	campaignService *service.CampaignService
	adGroupService  *service.AdGroupService
}

// NewCampaignHandler 创建广告系列处理器
func NewCampaignHandler() *CampaignHandler {
	return &CampaignHandler{
		campaignService: service.NewCampaignService(),
		adGroupService:  service.NewAdGroupService(),
	}
}

// GetCampaigns 获取广告系列列表
func (h *CampaignHandler) GetCampaigns(c *gin.Context) {
	accountID := c.Query("account_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	campaigns, total, err := h.campaignService.GetCampaigns(accountID, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get campaigns")
		return
	}

	middleware.Success(c, gin.H{
		"list":      campaigns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetCampaign 获取广告系列详情
func (h *CampaignHandler) GetCampaign(c *gin.Context) {
	campaignID := c.Param("id")
	if campaignID == "" {
		middleware.BadRequest(c, "campaign id is required")
		return
	}

	campaign, err := h.campaignService.GetCampaignByID(campaignID)
	if err != nil {
		middleware.NotFound(c, "Campaign not found")
		return
	}

	middleware.Success(c, campaign)
}

// GetAdGroups 获取广告组列表
func (h *CampaignHandler) GetAdGroups(c *gin.Context) {
	accountID := c.Query("account_id")
	campaignID := c.Query("campaign_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	adGroups, total, err := h.adGroupService.GetAdGroups(accountID, campaignID, page, pageSize)
	if err != nil {
		middleware.InternalError(c, "Failed to get ad groups")
		return
	}

	middleware.Success(c, gin.H{
		"list":      adGroups,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAdGroup 获取广告组详情
func (h *CampaignHandler) GetAdGroup(c *gin.Context) {
	adGroupID := c.Param("id")
	if adGroupID == "" {
		middleware.BadRequest(c, "ad group id is required")
		return
	}

	adGroup, err := h.adGroupService.GetAdGroupByID(adGroupID)
	if err != nil {
		middleware.NotFound(c, "Ad group not found")
		return
	}

	middleware.Success(c, adGroup)
}
