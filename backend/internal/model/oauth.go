package model

import (
	"time"
)

// OAuthToken OAuth授权令牌表
type OAuthToken struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	AccountID    string    `gorm:"uniqueIndex;size:64;not null;comment:广告主账号ID" json:"account_id"`
	AccessToken  string    `gorm:"type:text;not null;comment:访问令牌" json:"access_token"`
	RefreshToken string    `gorm:"type:text;not null;comment:刷新令牌" json:"refresh_token"`
	ExpiresAt    time.Time `gorm:"not null;comment:过期时间" json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (OAuthToken) TableName() string {
	return "oauth_tokens"
}

// IsExpired 检查令牌是否过期
func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// OAuthAccount OAuth授权账号表（包含账号信息和Token）
type OAuthAccount struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	AccountID    string    `gorm:"uniqueIndex;size:64;not null;comment:广告主账号ID" json:"account_id"`
	AccountName  string    `gorm:"size:255;comment:账号名称" json:"account_name"`
	Authorized   bool      `gorm:"default:false;comment:是否已授权" json:"authorized"`
	AccessToken  string    `gorm:"type:text;comment:访问令牌" json:"access_token"`
	RefreshToken string    `gorm:"type:text;comment:刷新令牌" json:"refresh_token"`
	TokenExpiresAt time.Time `gorm:"comment:Token过期时间" json:"token_expires_at"`
	IsOnline     bool      `gorm:"default:false;comment:是否在线" json:"is_online"`
	LastCrawlTime *time.Time `gorm:"comment:最后抓取时间" json:"last_crawl_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}

// AdAccount 广告账号表
type AdAccount struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	AccountID    string    `gorm:"uniqueIndex;size:64;not null;comment:账号ID" json:"account_id"`
	AccountName  string    `gorm:"size:255;comment:账号名称" json:"account_name"`
	AccountRole  int       `gorm:"default:0;comment:账号角色" json:"account_role"`
	Status       int       `gorm:"default:1;comment:状态(1正常,0禁用)" json:"status"`
	TokenID      uint      `gorm:"index;comment:关联TokenID" json:"token_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AdAccount) TableName() string {
	return "ad_accounts"
}

// Campaign 推广计划表
type Campaign struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CampaignID   string    `gorm:"uniqueIndex;size:64;not null;comment:推广计划ID" json:"campaign_id"`
	AccountID    string    `gorm:"index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	CampaignName string    `gorm:"size:500;comment:推广计划名称" json:"campaign_name"`
	CampaignType int       `gorm:"default:0;comment:推广计划类型" json:"campaign_type"`
	Status       int       `gorm:"default:0;comment:状态" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Campaign) TableName() string {
	return "campaigns"
}

// AdGroup 广告组表
type AdGroup struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	GroupID      string    `gorm:"uniqueIndex;size:64;not null;comment:广告组ID" json:"group_id"`
	CampaignID   string    `gorm:"index;size:64;not null;comment:推广计划ID" json:"campaign_id"`
	AccountID    string    `gorm:"index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	GroupName    string    `gorm:"size:500;comment:广告组名称" json:"group_name"`
	Status       int       `gorm:"default:0;comment:状态" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AdGroup) TableName() string {
	return "ad_groups"
}

// Ad 广告表（广告创意）
type Ad struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	AdID         string    `gorm:"uniqueIndex;size:64;not null;comment:广告ID" json:"ad_id"`
	GroupID      string    `gorm:"index;size:64;not null;comment:广告组ID" json:"group_id"`
	CampaignID   string    `gorm:"index;size:64;not null;comment:推广计划ID" json:"campaign_id"`
	AccountID    string    `gorm:"index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	AdName       string    `gorm:"size:500;comment:广告名称" json:"ad_name"`
	CreativeID   string    `gorm:"index;size:64;comment:创意ID" json:"creative_id"`
	AdType       int       `gorm:"default:0;comment:广告类型" json:"ad_type"`
	Status       int       `gorm:"default:0;comment:状态" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Ad) TableName() string {
	return "ads"
}

// AdCreative 广告创意表
type AdCreative struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreativeID  string    `gorm:"uniqueIndex;size:64;not null;comment:创意ID" json:"creative_id"`
	AccountID   string    `gorm:"index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	CreativeName string   `gorm:"size:500;comment:创意名称" json:"creative_name"`
	CreativeType int      `gorm:"default:0;comment:创意类型(图片/视频/图文)" json:"creative_type"`
	ImageIDs    string    `gorm:"type:text;comment:素材ID列表(JSON)" json:"image_ids"`
	Title       string    `gorm:"size:500;comment:创意标题" json:"title"`
	Description string    `gorm:"type:text;comment:创意描述" json:"description"`
	Status      int       `gorm:"default:0;comment:状态" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AdCreative) TableName() string {
	return "ad_creatives"
}

// AdMaterial 广告素材表
type AdMaterial struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	MaterialID   string    `gorm:"uniqueIndex;size:64;not null;comment:素材ID" json:"material_id"`
	AccountID    string    `gorm:"index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	MaterialType int       `gorm:"default:0;comment:素材类型(1图片,2视频,3音频)" json:"material_type"`
	MaterialURL  string    `gorm:"type:text;comment:素材URL" json:"material_url"`
	Width        int       `gorm:"default:0;comment:宽度" json:"width"`
	Height       int       `gorm:"default:0;comment:高度" json:"height"`
	FileSize     int64     `gorm:"default:0;comment:文件大小(字节)" json:"file_size"`
	Signature    string    `gorm:"size:128;comment:文件签名" json:"signature"`
	Status       int       `gorm:"default:0;comment:状态" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AdMaterial) TableName() string {
	return "ad_materials"
}

// Project 项目表（可选层级，腾讯广告部分账号有）
type Project struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ProjectID  string    `gorm:"uniqueIndex;size:64;not null;comment:项目ID" json:"project_id"`
	AccountID  string    `gorm:"index;size:64;not null;comment:广告主账号ID" json:"account_id"`
	ProjectName string   `gorm:"size:500;comment:项目名称" json:"project_name"`
	Status     int       `gorm:"default:0;comment:状态" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}
