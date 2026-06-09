package database

import (
	"fmt"
	"log"

	"marketing-platform/internal/config"
	"marketing-platform/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig, serverMode string) error {
	var logLevel logger.LogLevel
	if serverMode == "debug" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Silent
	}

	dsn := cfg.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.GetConnMaxLifetime())

	DB = db
	log.Println("Database connection established")
	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	log.Println("Starting database migration...")

	err := DB.AutoMigrate(
		&model.User{},
		&model.OAuthToken{},
		&model.OAuthAccount{},
		&model.AdAccount{},
		&model.Project{},
		&model.Campaign{},
		&model.AdGroup{},
		&model.Ad{},
		&model.AdCreative{},
		&model.AdMaterial{},
		&model.DailyReport{},
		&model.HourlyReport{},
		&model.TargetReport{},
		&model.ReportTask{},
		&model.CrawlTask{},
		&model.CrawlTaskLog{},
		// 智能托管
		&model.HostingRule{},
		&model.HostingExecution{},
		&model.HostingAlert{},
		&model.AdPerformanceSnapshot{},
	)

	if err != nil {
		log.Printf("Database migration failed: %v", err)
		return err
	}

	// 添加表注释
	if err := addTableComments(); err != nil {
		log.Printf("Warning: Failed to add table comments: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// addTableComments 添加表注释
func addTableComments() error {
	comments := []struct {
		Table string
		Comment string
	}{
		{"users", "用户表"},
		{"oauth_tokens", "OAuth授权令牌表"},
		{"oauth_accounts", "OAuth授权账号表"},
		{"ad_accounts", "广告账号表"},
		{"projects", "项目表-可选层级"},
		{"campaigns", "推广计划表"},
		{"ad_groups", "广告组表"},
		{"ads", "广告表-广告创意"},
		{"ad_creatives", "广告创意表"},
		{"ad_materials", "广告素材表"},
		{"daily_reports", "日报表-腾讯广告API"},
		{"hourly_reports", "小时报表-腾讯广告API"},
		{"target_reports", "定向标签报表-腾讯广告API"},
		{"report_tasks", "异步报表任务表"},
		{"crawl_tasks", "数据抓取任务表"},
		{"crawl_task_logs", "数据抓取任务日志表"},
		{"hosting_rules", "智能托管规则表"},
		{"hosting_executions", "智能托管执行记录表"},
		{"hosting_alerts", "智能托管告警通知表"},
		{"ad_performance_snapshots", "广告性能快照表"},
	}

	for _, c := range comments {
		sql := fmt.Sprintf("ALTER TABLE `%s` COMMENT = ?", c.Table)
		if err := DB.Exec(sql, c.Comment).Error; err != nil {
			return fmt.Errorf("failed to add comment for %s: %w", c.Table, err)
		}
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
