package crawler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"marketing-platform/internal/model"
)

// APIConfig API配置
type APIConfig struct {
	BaseURL        string
	MaxRetry       int
	RequestTimeout time.Duration
	RateLimit      rate.Limit // 每秒请求数限制
}

// DefaultAPIConfig 默认配置
var DefaultAPIConfig = APIConfig{
	BaseURL:        "https://api.e.qq.com/v1.3",
	MaxRetry:       3,
	RequestTimeout: 30 * time.Second,
	RateLimit:      10, // 每秒10个请求
}

// APIResponse API响应结构
type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// APIError API错误
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API错误 [%d]: %s", e.Code, e.Message)
}

// Limiter 限流器 - 每个账号独立的限流器
type Limiter struct {
	limiter  *rate.Limiter
	lastUsed time.Time
	mu       sync.Mutex
}

// APIClient API客户端
type APIClient struct {
	config   APIConfig
	httpClient *http.Client
	db       *gorm.DB

	// 账号限流器映射
	limiters map[string]*Limiter
	limiterMu sync.RWMutex

	// Token缓存
	tokenCache *cache.Cache

	// 请求锁（防止同一账号并发刷新token）
	tokenMu sync.Map
}

// NewAPIClient 创建API客户端
func NewAPIClient(db *gorm.DB) *APIClient {
	return &APIClient{
		config:   DefaultAPIConfig,
		httpClient: &http.Client{
			Timeout: DefaultAPIConfig.RequestTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		db:         db,
		limiters:   make(map[string]*Limiter),
		tokenCache: cache.New(55*time.Minute, 10*time.Minute),
	}
}

// getLimiter 获取或创建账号限流器
func (c *APIClient) getLimiter(accountID string) *rate.Limiter {
	c.limiterMu.Lock()
	defer c.limiterMu.Unlock()

	if l, exists := c.limiters[accountID]; exists {
		return l.limiter
	}

	limiter := rate.NewLimiter(c.config.RateLimit, 5) // burst为5
	c.limiters[accountID] = &Limiter{
		limiter:  limiter,
		lastUsed: time.Now(),
	}
	return limiter
}

// GetAccessToken 获取账号的AccessToken
func (c *APIClient) GetAccessToken(accountID string) (string, error) {
	// 尝试从缓存获取
	if token, found := c.tokenCache.Get(accountID); found {
		return token.(string), nil
	}

	// 从数据库查询
	var oauth model.OAuthAccount
	if err := c.db.Where("account_id = ?", accountID).First(&oauth).Error; err != nil {
		return "", fmt.Errorf("查询账号失败: %w", err)
	}

	// 检查token是否过期
	if oauth.TokenExpiresAt.Before(time.Now()) {
		// 需要刷新token
		newToken, err := c.RefreshToken(accountID, oauth.RefreshToken)
		if err != nil {
			return "", err
		}
		return newToken, nil
	}

	// 缓存token
	c.tokenCache.Set(accountID, oauth.AccessToken, time.Until(oauth.TokenExpiresAt))
	return oauth.AccessToken, nil
}

// RefreshToken 刷新AccessToken
func (c *APIClient) RefreshToken(accountID, refreshToken string) (string, error) {
	// 使用锁防止并发刷新
	lock, _ := c.tokenMu.LoadOrStore(accountID, &sync.Mutex{})
	mu := lock.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	// 再次检查缓存（可能其他协程已刷新）
	if token, found := c.tokenCache.Get(accountID); found {
		return token.(string), nil
	}

	// 调用刷新接口
	// 注意：这里需要根据腾讯广告的实际刷新接口实现
	refreshURL := fmt.Sprintf("%s/oauth/refresh_access_token", c.config.BaseURL)
	
	reqBody := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"app_id":        "", // 从配置获取
	}

	body, err := c.doRequest(http.MethodPost, refreshURL, accountID, reqBody)
	if err != nil {
		return "", err
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("解析刷新响应失败: %w", err)
	}

	// 更新数据库
	expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	if err := c.db.Model(&model.OAuthAccount{}).Where("account_id = ?", accountID).Updates(map[string]interface{}{
		"access_token":    resp.AccessToken,
		"refresh_token":   resp.RefreshToken,
		"token_expires_at": expiresAt,
	}).Error; err != nil {
		return "", fmt.Errorf("更新token失败: %w", err)
	}

	// 更新缓存
	c.tokenCache.Set(accountID, resp.AccessToken, time.Until(expiresAt))
	
	return resp.AccessToken, nil
}

// Request 发起API请求
func (c *APIClient) Request(ctx context.Context, method, endpoint string, accountID string, params map[string]interface{}) ([]byte, error) {
	// 获取限流器
	limiter := c.getLimiter(accountID)

	// 限流等待
	if err := limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("限流等待失败: %w", err)
	}

	// 获取token
	token, err := c.GetAccessToken(accountID)
	if err != nil {
		return nil, fmt.Errorf("获取token失败: %w", err)
	}

	// 构建URL
	url := fmt.Sprintf("%s%s", c.config.BaseURL, endpoint)
	if method == http.MethodGet && len(params) > 0 {
		url = url + "?" + buildQueryParams(params)
	}

	// 构建请求
	var reqBody []byte
	if method != http.MethodGet && params != nil {
		reqBody, _ = json.Marshal(params)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码异常: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查API错误码
	if apiResp.Code != 0 {
		return nil, &APIError{
			Code:    apiResp.Code,
			Message: apiResp.Message,
		}
	}

	return apiResp.Data, nil
}

// doRequest 内部请求方法（用于刷新token等不需要token的场景）
func (c *APIClient) doRequest(method, url, accountID string, params interface{}) ([]byte, error) {
	body, _ := json.Marshal(params)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// RetryRequest 带重试的请求
func (c *APIClient) RetryRequest(ctx context.Context, method, endpoint, accountID string, params map[string]interface{}) ([]byte, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.MaxRetry; attempt++ {
		if attempt > 0 {
			// 指数退避
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("[APIClient] 请求失败，%v 后重试 (尝试 %d/%d)", backoff, attempt, c.config.MaxRetry)
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		body, err := c.Request(ctx, method, endpoint, accountID, params)
		if err == nil {
			return body, nil
		}

		lastErr = err

		// 判断错误类型，决定是否重试
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			switch apiErr.Code {
			case 401: // token过期，需要刷新后重试
			 c.tokenCache.Delete(accountID)
			case 429: // 限流，等待后重试
			 continue
			default: // 其他错误，不重试
			 return nil, err
			}
		}
	}

	return nil, fmt.Errorf("请求失败，已重试 %d 次: %w", c.config.MaxRetry, lastErr)
}

// GetCampaigns 获取广告系列列表
func (c *APIClient) GetCampaigns(ctx context.Context, accountID string, page, pageSize int) ([]byte, error) {
	params := map[string]interface{}{
		"page":       page,
		"page_size":  pageSize,
		"fields":     []string{"campaign_id", "campaign_name", "campaign_type", "status", "daily_budget", "created_time"},
	}
	return c.RetryRequest(ctx, http.MethodGet, "/campaigns/get", accountID, params)
}

// GetAdGroups 获取广告组列表
func (c *APIClient) GetAdGroups(ctx context.Context, accountID, campaignID string, page, pageSize int) ([]byte, error) {
	params := map[string]interface{}{
		"campaign_id": campaignID,
		"page":        page,
		"page_size":   pageSize,
		"fields":      []string{"adgroup_id", "adgroup_name", "status", "bid_amount", "created_time"},
	}
	return c.RetryRequest(ctx, http.MethodGet, "/adgroups/get", accountID, params)
}

// GetAds 获取广告列表
func (c *APIClient) GetAds(ctx context.Context, accountID, adgroupID string, page, pageSize int) ([]byte, error) {
	params := map[string]interface{}{
		"adgroup_id": adgroupID,
		"page":       page,
		"page_size":  pageSize,
		"fields":     []string{"ad_id", "ad_name", "status", "created_time"},
	}
	return c.RetryRequest(ctx, http.MethodGet, "/ads/get", accountID, params)
}

// GetCreatives 获取广告创意列表
func (c *APIClient) GetCreatives(ctx context.Context, accountID, adID string, page, pageSize int) ([]byte, error) {
	params := map[string]interface{}{
		"ad_id":      adID,
		"page":       page,
		"page_size":  pageSize,
		"fields":     []string{"creative_id", "creative_elements", "preview_url", "status"},
	}
	return c.RetryRequest(ctx, http.MethodGet, "/creatives/get", accountID, params)
}

// GetMaterials 获取广告素材列表
func (c *APIClient) GetMaterials(ctx context.Context, accountID string, materialType string, page, pageSize int) ([]byte, error) {
	params := map[string]interface{}{
		"material_type": materialType,
		"page":          page,
		"page_size":     pageSize,
	}
	return c.RetryRequest(ctx, http.MethodGet, "/materials/get", accountID, params)
}

// GetHourlyReport 获取小时报表
func (c *APIClient) GetHourlyReport(ctx context.Context, accountID string, date string, page, pageSize int) ([]byte, error) {
	params := map[string]interface{}{
		"date":      date,
		"page":      page,
		"page_size": pageSize,
		"fields": []string{
			"date", "hour", "campaign_id", "adgroup_id", "ad_id",
			"view_count", "click_count", "convert_count", "cost",
		},
	}
	return c.RetryRequest(ctx, http.MethodGet, "/reports/hourly_get", accountID, params)
}

// GetDailyReport 获取日报表
func (c *APIClient) GetDailyReport(ctx context.Context, accountID string, startDate, endDate string, page, pageSize int) ([]byte, error) {
	params := map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
		"page":       page,
		"page_size":  pageSize,
		"fields": []string{
			"date", "campaign_id", "adgroup_id", "ad_id",
			"view_count", "click_count", "convert_count", "cost",
		},
	}
	return c.RetryRequest(ctx, http.MethodGet, "/reports/daily_get", accountID, params)
}

// buildQueryParams 构建查询参数
func buildQueryParams(params map[string]interface{}) string {
	var buf bytes.Buffer
	for k, v := range params {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		switch val := v.(type) {
		case string:
			buf.WriteString(val)
		case int:
			buf.WriteString(fmt.Sprintf("%d", val))
		default:
			if b, err := json.Marshal(v); err == nil {
				buf.Write(b)
			}
		}
	}
	return buf.String()
}
