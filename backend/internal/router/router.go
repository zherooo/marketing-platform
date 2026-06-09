package router

import (
	"marketing-platform/internal/handler"
	"marketing-platform/internal/middleware"
	"marketing-platform/internal/scheduler"

	"github.com/gin-gonic/gin"
)

// 全局调度器引用
var Scheduler *scheduler.Scheduler

// Setup 路由设置
func Setup(r *gin.Engine, sched *scheduler.Scheduler) {
	Scheduler = sched
	// 健康检查（无需认证）
	r.GET("/health", handler.NewAuthHandler().HealthCheck)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 认证相关（无需认证）
		authGroup := v1.Group("/auth")
		{
			authHandler := handler.NewAuthHandler()
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
		}

		// OAuth相关（无需认证）
		oauthGroup := v1.Group("/oauth")
		{
			oauthHandler := handler.NewOAuthHandler()
			oauthGroup.GET("/url", oauthHandler.GetAuthURL)
			oauthGroup.GET("/callback", oauthHandler.Callback)
		}

		// 需要认证的接口
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			// 用户
			userHandler := handler.NewAuthHandler()
			protected.GET("/user/info", userHandler.GetCurrentUser)
			protected.POST("/user/password", userHandler.ChangePassword)

			// 报表
			reportHandler := handler.NewReportHandler()
			protected.GET("/reports/daily", reportHandler.GetDailyReports)
			protected.GET("/reports/hourly", reportHandler.GetHourlyReports)
			protected.GET("/reports/target", reportHandler.GetTargetReports)
			protected.GET("/reports/trend", reportHandler.GetDailyTrend)

			// 广告系列
			campaignHandler := handler.NewCampaignHandler()
			protected.GET("/campaigns", campaignHandler.GetCampaigns)
			protected.GET("/campaigns/:id", campaignHandler.GetCampaign)
			protected.GET("/adgroups", campaignHandler.GetAdGroups)
			protected.GET("/adgroups/:id", campaignHandler.GetAdGroup)

			// 广告
			adHandler := handler.NewAdHandler()
			protected.GET("/ads", adHandler.GetAds)
			protected.GET("/ads/:id", adHandler.GetAd)
			protected.GET("/creatives", adHandler.GetCreatives)
			protected.GET("/creatives/:id", adHandler.GetCreative)

			// 广告素材
			materialHandler := handler.NewMaterialHandler()
			protected.GET("/materials", materialHandler.GetMaterials)
			protected.GET("/materials/:id", materialHandler.GetMaterial)
			protected.GET("/materials/stats", materialHandler.GetMaterialStats)

			// OAuth管理
			oauthHandler := handler.NewOAuthHandler()
			protected.GET("/oauth/tokens", oauthHandler.GetTokenList)
			protected.DELETE("/oauth/tokens/:account_id", oauthHandler.DeleteToken)

			// 数据抓取管理
			crawlerHandler := handler.NewCrawlerHandler(sched)
			crawler := protected.Group("/crawler")
			{
				crawler.POST("/start", crawlerHandler.StartCrawl)                                   // 手动启动抓取
				crawler.GET("/statistics", crawlerHandler.GetStatistics)                          // 获取抓取统计
				crawler.GET("/tasks", crawlerHandler.ListTasks)                                    // 查询任务列表
				crawler.GET("/tasks/running", crawlerHandler.GetRunningTasks)                       // 获取正在运行的任务
				crawler.POST("/tasks/:task_id/cancel", crawlerHandler.CancelTask)                   // 取消任务
				crawler.POST("/tasks/:task_id/retry", crawlerHandler.RetryTask)                    // 重试任务
				crawler.POST("/trigger/hourly", crawlerHandler.TriggerHourlyReport)                // 触发小时报表抓取
				crawler.POST("/trigger/daily", crawlerHandler.TriggerDailyReport)                  // 触发动报表抓取
				crawler.POST("/trigger/struct", crawlerHandler.TriggerAllStruct)                   // 触发广告结构抓取
				crawler.POST("/trigger/campaign", crawlerHandler.TriggerCampaign)                 // 触发广告系列抓取
				crawler.POST("/trigger/campaign/:campaign_id/cascade", crawlerHandler.TriggerCampaignCascade) // 触发广告系列级联抓取
				crawler.POST("/trigger/adgroup/:adgroup_id/cascade", crawlerHandler.TriggerAdGroupCascade)     // 触发广告组级联抓取
				crawler.POST("/trigger/ad/:ad_id/cascade", crawlerHandler.TriggerAdCascade)                   // 触发广告级联抓取
			}

			// 智能托管
			hostingRuleHandler := handler.NewHostingRuleHandler()
			hostingExecHandler := handler.NewHostingExecutionHandler()
			hostingAlertHandler := handler.NewHostingAlertHandler()
			hosting := protected.Group("/hosting")
			{
				// 规则管理
				hostingRules := hosting.Group("/rules")
				{
					hostingRules.GET("", hostingRuleHandler.ListRules)                    // 规则列表
					hostingRules.POST("", hostingRuleHandler.CreateRule)                  // 创建规则
					hostingRules.GET("/:id", hostingRuleHandler.GetRule)                  // 规则详情
					hostingRules.PUT("/:id", hostingRuleHandler.UpdateRule)               // 更新规则
					hostingRules.DELETE("/:id", hostingRuleHandler.DeleteRule)            // 删除规则
					hostingRules.POST("/:id/toggle", hostingRuleHandler.ToggleRuleStatus) // 切换启用/禁用
					hostingRules.POST("/:id/test", hostingRuleHandler.TestRule)           // 测试规则
				}

				// 执行记录
				hostingExec := hosting.Group("/executions")
				{
					hostingExec.GET("", hostingExecHandler.ListExecutions)                 // 执行记录列表
					hostingExec.GET("/:id", hostingExecHandler.GetExecution)               // 执行记录详情
					hostingExec.POST("/:id/rollback", hostingExecHandler.RollbackExecution) // 回滚执行
				}

				// 告警通知
				hostingAlert := hosting.Group("/alerts")
				{
					hostingAlert.GET("", hostingAlertHandler.ListAlerts)                  // 告警列表
					hostingAlert.GET("/:id", hostingAlertHandler.GetAlert)                // 告警详情
					hostingAlert.POST("/:id/read", hostingAlertHandler.MarkAlertRead)     // 标记已读
					hostingAlert.POST("/:id/handle", hostingAlertHandler.HandleAlert)     // 处理告警
					hostingAlert.POST("/:id/ignore", hostingAlertHandler.IgnoreAlert)     // 忽略告警
				}

				// 看板 & 触发
				hosting.GET("/dashboard", hostingExecHandler.GetDashboard)               // 看板数据
				hosting.POST("/trigger/evaluate", hostingExecHandler.TriggerEvaluate)     // 手动触发评估
				hosting.POST("/trigger/collect", hostingExecHandler.TriggerCollect)       // 手动触发快照采集
			}
		}
	}
}
