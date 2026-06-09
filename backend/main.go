package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"marketing-platform/internal/config"
	"marketing-platform/internal/crawler"
	"marketing-platform/internal/database"
	"marketing-platform/internal/logger"
	"marketing-platform/internal/middleware"
	"marketing-platform/internal/router"
	"marketing-platform/internal/scheduler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// 3. 初始化数据库
	if err := database.Init(&cfg.Database, cfg.Server.Mode); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 4. 自动迁移数据库表
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 5. 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 6. 创建Gin引擎
	r := gin.New()

	// 7. 注册中间件
	r.Use(logger.InitRequestLogger())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 8. 初始化数据抓取管理器
	crawlManager := crawler.NewTaskManager(database.GetDB(), cfg.Crawler.MaxWorkersPerAccount)
	crawlManager.Init()
	crawlManager.Start()

	// 9. 初始化定时任务调度器
	sched := scheduler.NewScheduler(database.GetDB())
	sched.SetCrawlManager(crawlManager)
	if err := sched.Start(); err != nil {
		log.Printf("Warning: Failed to start scheduler: %v", err)
	}

	// 10. 设置路由
	router.Setup(r, sched)

	// 11. 启动服务器
	addr := cfg.Server.GetAddress()
	fmt.Printf("Server starting on %s\n", addr)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	fmt.Println("\nShutting down server...")

	// 停止抓取管理器
	crawlManager.Stop()

	// 停止调度器
	sched.Stop()
}
