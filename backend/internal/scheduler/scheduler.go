package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

"marketing-platform/internal/config"
"marketing-platform/internal/crawler"
"marketing-platform/internal/logger"
"marketing-platform/internal/model"
"marketing-platform/internal/service"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron           *cron.Cron
	db             *gorm.DB
	crawlManager   *crawler.TaskManager
	reportService  *service.ReportService
	crawlService   *service.CrawlService
	hostingService *service.HostingExecutorService
	config         *config.SchedulerConfig
}

// NewScheduler 创建调度器
func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{
		cron:           cron.New(cron.WithSeconds()),
		db:             db,
		reportService:  service.NewReportService(),
		crawlService:   service.NewCrawlService(),
		hostingService: service.NewHostingExecutorService(),
		config:         &config.Config.Scheduler,
	}
}

// SetCrawlManager 设置抓取管理器
func (s *Scheduler) SetCrawlManager(manager *crawler.TaskManager) {
	s.crawlManager = manager
}

// SetDB 设置数据库实例
func (s *Scheduler) SetDB(db *gorm.DB) {
	s.db = db
	s.crawlService.SetDB(db)
}

// Start 启动定时任务
func (s *Scheduler) Start() error {
	logger.Logger.Info("启动定时任务调度器")

	// 添加定时任务
	if err := s.registerCronJobs(); err != nil {
		return fmt.Errorf("注册定时任务失败: %w", err)
	}

	s.cron.Start()
	logger.Logger.Info("定时任务调度器已启动")

	return nil
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	logger.Logger.Info("停止定时任务调度器...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Logger.Info("定时任务调度器已停止")
}

// registerCronJobs 注册定时任务
func (s *Scheduler) registerCronJobs() error {
	// 1. 小时报表采集 - 每小时10分执行
	if _, err := s.cron.AddFunc(s.config.HourlyReportCron, func() {
		s.runWithRecover("小时报表采集", s.CollectHourlyReports)
	}); err != nil {
		return fmt.Errorf("注册小时报表任务失败: %w", err)
	}

	// 2. 日报表采集 - 每天凌晨2点执行
	if _, err := s.cron.AddFunc(s.config.DailyReportCron, func() {
		s.runWithRecover("日报表采集", s.CollectDailyReports)
	}); err != nil {
		return fmt.Errorf("注册日报表任务失败: %w", err)
	}

	// 3. 广告系列采集 - 每6小时执行
	if _, err := s.cron.AddFunc(s.config.CampaignCron, func() {
		s.runWithRecover("广告系列采集", s.CollectCampaigns)
	}); err != nil {
		return fmt.Errorf("注册广告系列任务失败: %w", err)
	}

	// 4. 广告组采集 - 每6小时执行
	if _, err := s.cron.AddFunc(s.config.AdGroupCron, func() {
		s.runWithRecover("广告组采集", s.CollectAdGroups)
	}); err != nil {
		return fmt.Errorf("注册广告组任务失败: %w", err)
	}

	// 5. 广告采集 - 每6小时执行
	if _, err := s.cron.AddFunc(s.config.AdCron, func() {
		s.runWithRecover("广告采集", s.CollectAds)
	}); err != nil {
		return fmt.Errorf("注册广告任务失败: %w", err)
	}

	// 6. 创意采集 - 每6小时执行
	if _, err := s.cron.AddFunc(s.config.CreativeCron, func() {
		s.runWithRecover("广告创意采集", s.CollectCreatives)
	}); err != nil {
		return fmt.Errorf("注册广告创意任务失败: %w", err)
	}

	// 7. 素材采集 - 每6小时执行
	if _, err := s.cron.AddFunc(s.config.MaterialCron, func() {
		s.runWithRecover("广告素材采集", s.CollectMaterials)
	}); err != nil {
		return fmt.Errorf("注册广告素材任务失败: %w", err)
	}

	// 8. Token刷新 - 每12小时执行
	if _, err := s.cron.AddFunc("0 0 */12 * * *", func() {
		s.runWithRecover("Token刷新", s.RefreshTokens)
	}); err != nil {
		return fmt.Errorf("注册Token刷新任务失败: %w", err)
	}

	// 9. 清理旧任务 - 每天凌晨1点执行
	if _, err := s.cron.AddFunc("0 0 1 * * *", func() {
		s.runWithRecover("清理旧任务", s.CleanOldTasks)
	}); err != nil {
		return fmt.Errorf("注册清理任务失败: %w", err)
	}

	// 10. 重试失败任务 - 每5分钟执行
	if _, err := s.cron.AddFunc("0 */5 * * * *", func() {
		s.runWithRecover("重试失败任务", s.RetryFailedTasks)
	}); err != nil {
		return fmt.Errorf("注册重试任务失败: %w", err)
	}

	// 11. 智能托管 - 性能快照采集 - 每5分钟执行
	if _, err := s.cron.AddFunc("0 */5 * * * *", func() {
		s.runWithRecover("广告性能快照采集", s.hostingService.CollectPerformanceSnapshots)
	}); err != nil {
		return fmt.Errorf("注册快照采集任务失败: %w", err)
	}

	// 12. 智能托管 - 规则评估与执行 - 每5分钟执行
	if _, err := s.cron.AddFunc("30 */5 * * * *", func() {
		s.runWithRecover("智能托管规则评估", s.hostingService.ExecuteAllActiveRules)
	}); err != nil {
		return fmt.Errorf("注册托管评估任务失败: %w", err)
	}

	// 13. 智能托管 - 广告状态检查 - 每10分钟执行
	if _, err := s.cron.AddFunc("0 */10 * * * *", func() {
		s.runWithRecover("广告状态检查", s.hostingService.CheckAdStatuses)
	}); err != nil {
		return fmt.Errorf("注册广告状态检查任务失败: %w", err)
	}

	// 14. 智能托管 - 每日重置计数器 - 每天零点执行
	if _, err := s.cron.AddFunc("0 0 0 * * *", func() {
		s.runWithRecover("重置托管每日计数", service.NewHostingRuleService().ResetDailyCount)
	}); err != nil {
		return fmt.Errorf("注册重置计数任务失败: %w", err)
	}

	// 15. 智能托管 - 清理旧记录 - 每天凌晨3点执行
	if _, err := s.cron.AddFunc("0 0 3 * * *", func() {
		s.runWithRecover("清理托管旧记录", func() error {
			return s.hostingService.CleanupOldRecords(30)
		})
	}); err != nil {
		return fmt.Errorf("注册清理托管记录任务失败: %w", err)
	}

	// 16. 智能托管 - 清理旧性能快照 - 每天凌晨3:30执行
	if _, err := s.cron.AddFunc("0 30 3 * * *", func() {
		s.runWithRecover("清理旧性能快照", s.hostingService.CleanOldSnapshots)
	}); err != nil {
		return fmt.Errorf("注册清理快照任务失败: %w", err)
	}

	return nil
}

// runWithRecover 带错误恢复的运行
func (s *Scheduler) runWithRecover(name string, fn func() error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error(fmt.Sprintf("%s发生panic", name),
				zap.Any("panic", r))
		}
	}()

	start := time.Now()
	logger.Logger.Info(fmt.Sprintf("开始执行: %s", name))

	if err := fn(); err != nil {
		logger.Logger.Error(fmt.Sprintf("%s执行失败", name),
			zap.Error(err),
			zap.Duration("耗时", time.Since(start)))
		return
	}

	logger.Logger.Info(fmt.Sprintf("%s执行成功", name),
		zap.Duration("耗时", time.Since(start)))
}

// CollectHourlyReports 采集小时报表
func (s *Scheduler) CollectHourlyReports() error {
	if s.crawlManager == nil {
		logger.Logger.Warn("抓取管理器未初始化，跳过小时报表采集")
		return nil
	}

	// 抓取昨天的小时报
	yesterday := time.Now().AddDate(0, 0, -1)
	stats, err := s.crawlManager.StartCrawlForAllAccounts(model.DataTypeHourlyReport, yesterday, yesterday)
	if err != nil {
		return fmt.Errorf("采集小时报表失败: %w", err)
	}

	logger.Logger.Info("小时报表采集任务已创建",
		zap.Int64("总任务数", stats.TotalTasks),
		zap.Int64("待处理", stats.PendingTasks))

	return nil
}

// CollectDailyReports 采集日报表
func (s *Scheduler) CollectDailyReports() error {
	if s.crawlManager == nil {
		logger.Logger.Warn("抓取管理器未初始化，跳过日报表采集")
		return nil
	}

	// 抓取昨天的日报
	yesterday := time.Now().AddDate(0, 0, -1)
	stats, err := s.crawlManager.StartCrawlForAllAccounts(model.DataTypeDailyReport, yesterday, yesterday)
	if err != nil {
		return fmt.Errorf("采集日报表失败: %w", err)
	}

	logger.Logger.Info("日报表采集任务已创建",
		zap.Int64("总任务数", stats.TotalTasks),
		zap.Int64("待处理", stats.PendingTasks))

	return nil
}

// CollectCampaigns 采集广告系列
func (s *Scheduler) CollectCampaigns() error {
	if s.crawlManager == nil {
		logger.Logger.Warn("抓取管理器未初始化，跳过广告系列采集")
		return nil
	}

	// 采集所有广告系列
	stats, err := s.crawlManager.StartCrawlForAllAccounts(model.DataTypeCampaign, time.Time{}, time.Time{})
	if err != nil {
		return fmt.Errorf("采集广告系列失败: %w", err)
	}

	logger.Logger.Info("广告系列采集任务已创建",
		zap.Int64("总任务数", stats.TotalTasks),
		zap.Int64("待处理", stats.PendingTasks))

	return nil
}

// CollectAdGroups 采集广告组
func (s *Scheduler) CollectAdGroups() error {
	if s.crawlManager == nil {
		logger.Logger.Warn("抓取管理器未初始化，跳过广告组采集")
		return nil
	}

	stats, err := s.crawlManager.StartCrawlForAllAccounts(model.DataTypeAdGroup, time.Time{}, time.Time{})
	if err != nil {
		return fmt.Errorf("采集广告组失败: %w", err)
	}

	logger.Logger.Info("广告组采集任务已创建",
		zap.Int64("总任务数", stats.TotalTasks),
		zap.Int64("待处理", stats.PendingTasks))

	return nil
}

// CollectAds 采集广告
func (s *Scheduler) CollectAds() error {
	if s.crawlManager == nil {
		logger.Logger.Warn("抓取管理器未初始化，跳过广告采集")
		return nil
	}

	stats, err := s.crawlManager.StartCrawlForAllAccounts(model.DataTypeAd, time.Time{}, time.Time{})
	if err != nil {
		return fmt.Errorf("采集广告失败: %w", err)
	}

	logger.Logger.Info("广告采集任务已创建",
		zap.Int64("总任务数", stats.TotalTasks),
		zap.Int64("待处理", stats.PendingTasks))

	return nil
}

// CollectCreatives 采集广告创意
func (s *Scheduler) CollectCreatives() error {
	if s.crawlManager == nil {
		logger.Logger.Warn("抓取管理器未初始化，跳过广告创意采集")
		return nil
	}

	stats, err := s.crawlManager.StartCrawlForAllAccounts(model.DataTypeCreative, time.Time{}, time.Time{})
	if err != nil {
		return fmt.Errorf("采集广告创意失败: %w", err)
	}

	logger.Logger.Info("广告创意采集任务已创建",
		zap.Int64("总任务数", stats.TotalTasks),
		zap.Int64("待处理", stats.PendingTasks))

	return nil
}

// CollectMaterials 采集广告素材
func (s *Scheduler) CollectMaterials() error {
	if s.crawlManager == nil {
		logger.Logger.Warn("抓取管理器未初始化，跳过广告素材采集")
		return nil
	}

	stats, err := s.crawlManager.StartCrawlForAllAccounts(model.DataTypeMaterial, time.Time{}, time.Time{})
	if err != nil {
		return fmt.Errorf("采集广告素材失败: %w", err)
	}

	logger.Logger.Info("广告素材采集任务已创建",
		zap.Int64("总任务数", stats.TotalTasks),
		zap.Int64("待处理", stats.PendingTasks))

	return nil
}

// RefreshTokens 刷新所有账号的Token
func (s *Scheduler) RefreshTokens() error {
	// TODO: 实现Token刷新逻辑
	// 使用腾讯广告SDK刷新过期的Token
	log.Println("Token刷新任务执行中...")
	return nil
}

// CleanOldTasks 清理旧任务
func (s *Scheduler) CleanOldTasks() error {
	// 清理30天前的已完成/失败/取消的任务
	deleted, err := s.crawlService.CleanOldTasks(30)
	if err != nil {
		return fmt.Errorf("清理旧任务失败: %w", err)
	}

	logger.Logger.Info("清理旧任务完成",
		zap.Int64("删除数量", deleted))

	return nil
}

// RetryFailedTasks 重试失败任务
func (s *Scheduler) RetryFailedTasks() error {
	tasks, err := s.crawlService.GetPendingTasks(100)
	if err != nil {
		return fmt.Errorf("获取待重试任务失败: %w", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	logger.Logger.Info("重试失败任务",
		zap.Int("任务数量", len(tasks)))

	// 将待重试任务添加到协程池
	for _, task := range tasks {
		if s.crawlManager != nil {
			pool := s.crawlManager.GetPool(task.AccountID)
			if pool != nil {
				taskItem := model.CrawlTaskItem{
					TaskID:    task.TaskID,
					AccountID: task.AccountID,
					DataType:  task.DataType,
					StartDate: task.StartDate,
					EndDate:   task.EndDate,
				}
				pool.AddTask(taskItem)
			}
		}
	}

	return nil
}

// TriggerManualCrawl 手动触发抓取
func (s *Scheduler) TriggerManualCrawl(req *model.CrawlRequest) (*model.CrawlStatistics, error) {
	if s.crawlManager == nil {
		return nil, fmt.Errorf("抓取管理器未初始化")
	}

	logger.Logger.Info("手动触发抓取",
		zap.Strings("账号列表", req.AccountIDs),
		zap.Strings("数据类型", req.DataTypes),
		zap.Time("开始日期", req.StartDate),
		zap.Time("结束日期", req.EndDate))

	return s.crawlManager.StartCrawl(req)
}

// GetStatistics 获取抓取统计
func (s *Scheduler) GetStatistics() (*model.CrawlStatistics, error) {
	return s.crawlService.GetStatistics()
}

// GetRunningTasks 获取正在运行的任务
func (s *Scheduler) GetRunningTasks() ([]model.CrawlTask, error) {
	return s.crawlService.GetRunningTasks()
}

// ListTasks 查询任务列表
func (s *Scheduler) ListTasks(accountID string, status int, page, pageSize int) ([]model.CrawlTask, int64, error) {
	return s.crawlService.ListTasks(accountID, status, page, pageSize)
}

// CancelTask 取消任务
func (s *Scheduler) CancelTask(taskID string) error {
	return s.crawlService.CancelTask(taskID)
}

// RetryTask 重试任务
func (s *Scheduler) RetryTask(taskID string) error {
	return s.crawlService.RetryTask(taskID)
}

// CrawlCampaignCascade 级联抓取广告系列及其下属数据
func (s *Scheduler) CrawlCampaignCascade(accountID, campaignID string) error {
	if s.crawlManager == nil {
		return fmt.Errorf("抓取管理器未初始化")
	}
	return s.crawlManager.CrawlCampaignCascade(context.Background(), accountID, campaignID)
}

// CrawlAdGroupCascade 级联抓取广告组及其下属数据
func (s *Scheduler) CrawlAdGroupCascade(accountID, adgroupID string) error {
	if s.crawlManager == nil {
		return fmt.Errorf("抓取管理器未初始化")
	}
	return s.crawlManager.CrawlAdGroupCascade(context.Background(), accountID, adgroupID)
}

// CrawlAdCascade 级联抓取广告及其下属数据
func (s *Scheduler) CrawlAdCascade(accountID, adID string) error {
	if s.crawlManager == nil {
		return fmt.Errorf("抓取管理器未初始化")
	}
	return s.crawlManager.CrawlAdCascade(context.Background(), accountID, adID)
}
