package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置实例
var Config *AppConfig

// AppConfig 应用配置
type AppConfig struct {
	Server    ServerConfig     `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	TencentAd TencentAdConfig `mapstructure:"tencent_ad"`
	Log       LogConfig       `mapstructure:"log"`
	Report    ReportConfig    `mapstructure:"report"`
	Scheduler SchedulerConfig  `mapstructure:"scheduler"`
	Crawler   CrawlerConfig    `mapstructure:"crawler"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

// TencentAdConfig 腾讯广告配置
type TencentAdConfig struct {
	AppID                  string `mapstructure:"app_id"`
	AppSecret              string `mapstructure:"app_secret"`
	RedirectURI            string `mapstructure:"redirect_uri"`
	TokenRefreshInterval   int    `mapstructure:"token_refresh_interval"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// ReportConfig 报表任务配置
type ReportConfig struct {
	DailyCron      string `mapstructure:"daily_cron"`
	HourlyCron     string `mapstructure:"hourly_cron"`
	AsyncThreshold int    `mapstructure:"async_threshold"`
	BatchSize      int    `mapstructure:"batch_size"`
}

// SchedulerConfig 定时任务调度配置
type SchedulerConfig struct {
	HourlyReportCron string `mapstructure:"hourly_report_cron"` // 小时报表采集
	DailyReportCron  string `mapstructure:"daily_report_cron"`  // 日报表采集
	CampaignCron     string `mapstructure:"campaign_cron"`      // 广告系列采集
	AdGroupCron      string `mapstructure:"adgroup_cron"`       // 广告组采集
	AdCron           string `mapstructure:"ad_cron"`            // 广告采集
	CreativeCron     string `mapstructure:"creative_cron"`      // 广告创意采集
	MaterialCron     string `mapstructure:"material_cron"`       // 广告素材采集
}

// CrawlerConfig 数据抓取配置
type CrawlerConfig struct {
	MaxWorkersPerAccount int `mapstructure:"max_workers_per_account"` // 每个账号最大并发数
	TaskQueueSize        int `mapstructure:"task_queue_size"`         // 任务队列大小
	MaxRetry             int `mapstructure:"max_retry"`              // 最大重试次数
	RetryDelay           int `mapstructure:"retry_delay"`            // 重试延迟（秒）
	RequestTimeout       int `mapstructure:"request_timeout"`        // 请求超时（秒）
	RateLimit            int `mapstructure:"rate_limit"`             // API限流（每秒请求数）
}

// Load 加载配置文件
func Load(configPath string) (*AppConfig, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 环境变量支持
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	Config = &cfg
	return &cfg, nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Name, c.Charset)
}

// GetAddress 获取服务器地址
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetConnMaxLifetime 获取连接最大生存时间
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}
