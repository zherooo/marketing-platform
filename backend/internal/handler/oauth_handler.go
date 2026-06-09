package handler

import (
	"fmt"

	"marketing-platform/internal/config"
	"marketing-platform/internal/middleware"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// OAuthHandler OAuth处理器
type OAuthHandler struct {
	oauthService *service.OAuthService
}

// NewOAuthHandler 创建OAuth处理器
func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{
		oauthService: service.NewOAuthService(),
	}
}

// GetAuthURL 获取授权URL
// @Summary 获取腾讯广告授权URL
// @Tags oauth
// @Router /api/v1/oauth/url [get]
func (h *OAuthHandler) GetAuthURL(c *gin.Context) {
	cfg := config.Config.TencentAd

	// 构建授权URL
	authURL := fmt.Sprintf(
		"https://developers.e.qq.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=ads_management",
		cfg.AppID,
		cfg.RedirectURI,
	)

	middleware.Success(c, gin.H{
		"url": authURL,
	})
}

// Callback 授权回调
// @Summary 授权回调
// @Tags oauth
// @Param code query string true "授权码"
// @Router /api/v1/oauth/callback [get]
func (h *OAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		middleware.BadRequest(c, "code is required")
		return
	}

	// TODO: 调用腾讯广告API交换access_token
	// 这里需要使用腾讯广告Go SDK实现

	middleware.Success(c, gin.H{
		"message": "OAuth callback received, token exchange not implemented",
	})
}

// GetTokenList 获取Token列表
func (h *OAuthHandler) GetTokenList(c *gin.Context) {
	tokens, err := h.oauthService.GetAllTokens()
	if err != nil {
		middleware.InternalError(c, "Failed to get tokens")
		return
	}

	middleware.Success(c, gin.H{
		"list": tokens,
	})
}

// DeleteToken 删除Token
func (h *OAuthHandler) DeleteToken(c *gin.Context) {
	accountID := c.Param("account_id")
	if accountID == "" {
		middleware.BadRequest(c, "account_id is required")
		return
	}

	if err := h.oauthService.DeleteToken(accountID); err != nil {
		middleware.InternalError(c, "Failed to delete token")
		return
	}

	middleware.Success(c, nil)
}
