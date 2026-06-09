package crawler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

"marketing-platform/internal/model"
"marketing-platform/internal/service"
)

// TaskManager 任务管理器 - 管理所有账号的协程池
type TaskManager struct {
	pools map[string]*WorkerPool
	mutex sync.RWMutex
	db    *gorm.DB

	// 依赖服务
	reportService *service.ReportService
	crawlService  *service.CrawlService
	apiClient     *APIClient

	// 配置
	maxWorkersPerAccount int

	// 抓取器
	hourlyReportCrawler *HourlyReportCrawler
	dailyReportCrawler   *DailyReportCrawler
	campaignCrawler      *CampaignCrawler
	adGroupCrawler       *AdGroupCrawler
	adCrawler            *AdCrawler
	creativeCrawler      *CreativeCrawler
	materialCrawler      *MaterialCrawler

	ctx    context.Context
	cancel context.CancelFunc
}

// NewTaskManager 创建任务管理器
func NewTaskManager(db *gorm.DB, maxWorkers int) *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		pools:               make(map[string]*WorkerPool),
		db:                  db,
		maxWorkersPerAccount: maxWorkers,
		ctx:                  ctx,
		cancel:               cancel,
	}
}

// Init 初始化任务管理器
func (m *TaskManager) Init() {
	m.reportService = service.NewReportService()
	m.crawlService = service.NewCrawlService()
	m.apiClient = NewAPIClient(m.db)

	// 初始化抓取器
	m.hourlyReportCrawler = NewHourlyReportCrawler(m.db, m.apiClient)
	m.dailyReportCrawler = NewDailyReportCrawler(m.db, m.apiClient)
	m.campaignCrawler = NewCampaignCrawler(m.db, m.apiClient)
	m.adGroupCrawler = NewAdGroupCrawler(m.db, m.apiClient)
	m.adCrawler = NewAdCrawler(m.db, m.apiClient)
	m.creativeCrawler = NewCreativeCrawler(m.db, m.apiClient)
	m.materialCrawler = NewMaterialCrawler(m.db, m.apiClient)

	log.Println("[TaskManager] 初始化完成")
}

// Start 启动任务管理器
func (m *TaskManager) Start() {
	if m.apiClient == nil {
		m.Init()
	}
	log.Println("[TaskManager] 启动任务管理器")
}

// Stop 停止任务管理器
func (m *TaskManager) Stop() {
	log.Println("[TaskManager] 停止任务管理器...")

	// 停止所有协程池
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for accountID, pool := range m.pools {
		pool.Stop()
		log.Printf("[TaskManager] 账号 %s 协程池已停止", accountID)
	}

	m.cancel()
	log.Println("[TaskManager] 任务管理器已停止")
}

// StartCrawl 启动抓取任务
func (m *TaskManager) StartCrawl(req *model.CrawlRequest) (*model.CrawlStatistics, error) {
	stats := &model.CrawlStatistics{}

	for _, accountID := range req.AccountIDs {
		// 获取或创建协程池
		pool := m.getOrCreatePool(accountID)

		for _, dataType := range req.DataTypes {
			// 生成任务ID
			taskID := m.generateTaskID(accountID, dataType)

			// 创建数据库任务记录
			task := &model.CrawlTask{
				TaskID:    taskID,
				AccountID: accountID,
				DataType:  dataType,
				StartDate: req.StartDate,
				EndDate:   req.EndDate,
				Status:    model.TaskStatusPending,
			}

			// 保存任务到数据库
			if err := m.db.Create(task).Error; err != nil {
				log.Printf("[TaskManager] 创建任务记录失败: %v", err)
				continue
			}

			stats.TotalTasks++

			// 添加到协程池队列
			taskItem := model.CrawlTaskItem{
				TaskID:    taskID,
				AccountID: accountID,
				DataType:  dataType,
				StartDate: req.StartDate,
				EndDate:   req.EndDate,
			}

			if pool.AddTask(taskItem) {
				stats.PendingTasks++
			}
		}
	}

	return stats, nil
}

// StartCrawlForAllAccounts 为所有账号启动抓取任务
func (m *TaskManager) StartCrawlForAllAccounts(dataType string, startDate, endDate time.Time) (*model.CrawlStatistics, error) {
	// 查询所有授权账号
	var accounts []model.OAuthAccount
	if err := m.db.Where("authorized = ?", true).Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("查询账号失败: %w", err)
	}

	req := &model.CrawlRequest{
		AccountIDs: make([]string, len(accounts)),
		DataTypes:  []string{dataType},
		StartDate:  startDate,
		EndDate:    endDate,
		Manual:     false,
	}

	for i, acc := range accounts {
		req.AccountIDs[i] = acc.AccountID
	}

	return m.StartCrawl(req)
}

// getOrCreatePool 获取或创建协程池
func (m *TaskManager) getOrCreatePool(accountID string) *WorkerPool {
	m.mutex.RLock()
	pool, exists := m.pools[accountID]
	m.mutex.RUnlock()

	if exists {
		return pool
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Double check
	pool, exists = m.pools[accountID]
	if exists {
		return pool
	}

	// 创建新协程池
	pool = NewWorkerPool(accountID, m.maxWorkersPerAccount, m)
	pool.Start()
	m.pools[accountID] = pool

	log.Printf("[TaskManager] 为账号 %s 创建新协程池", accountID)
	return pool
}

// generateTaskID 生成唯一任务ID
func (m *TaskManager) generateTaskID(accountID, dataType string) string {
	return fmt.Sprintf("%s_%s_%s", accountID, dataType, uuid.New().String()[:8])
}

// UpdateTaskStatus 更新任务状态
func (m *TaskManager) UpdateTaskStatus(taskID string, status int) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case model.TaskStatusRunning:
		updates["started_at"] = now
	case model.TaskStatusCompleted:
		updates["completed_at"] = now
		updates["progress"] = 100
	}

	return m.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// HandleTaskError 处理任务错误
func (m *TaskManager) HandleTaskError(taskID string, errorMsg string) error {
	// 查询当前重试次数
	var task model.CrawlTask
	if err := m.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"error_msg": errorMsg,
		"retry_count": task.RetryCount + 1,
	}

	// 如果重试次数小于最大重试次数，设置为待处理
	if task.RetryCount < 3 {
		updates["status"] = model.TaskStatusPending
		nextRetry := time.Now().Add(5 * time.Second)
		updates["next_retry_at"] = nextRetry
	} else {
		// 超过最大重试次数，标记为失败
		updates["status"] = model.TaskStatusFailed
	}

	return m.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// GetTaskProgress 获取任务进度
func (m *TaskManager) GetTaskProgress(taskID string) (*model.CrawlTask, error) {
	var task model.CrawlTask
	if err := m.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTaskProgress 更新任务进度
func (m *TaskManager) UpdateTaskProgress(taskID string, progress, successCount, failCount int) error {
	return m.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"progress":      progress,
		"success_count": successCount,
		"fail_count":    failCount,
	}).Error
}

// UpdateAccountOnline 更新账号在线状态
func (m *TaskManager) UpdateAccountOnline(accountID string) {
	m.db.Model(&model.OAuthAccount{}).Where("account_id = ?", accountID).Updates(map[string]interface{}{
		"is_online": true,
	})
}

// UpdateAccountLastCrawlTime 更新账号最后抓取时间
func (m *TaskManager) UpdateAccountLastCrawlTime(accountID string) {
	m.db.Model(&model.OAuthAccount{}).Where("account_id = ?", accountID).Updates(map[string]interface{}{
		"last_crawl_time": time.Now(),
	})
}

// GetStatistics 获取抓取统计
func (m *TaskManager) GetStatistics() (*model.CrawlStatistics, error) {
	stats := &model.CrawlStatistics{}

	var total int64
	m.db.Model(&model.CrawlTask{}).Count(&total)
	stats.TotalTasks = total

	m.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusPending).Count(&stats.PendingTasks)
	m.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusRunning).Count(&stats.RunningTasks)
	m.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusCompleted).Count(&stats.CompletedTasks)
	m.db.Model(&model.CrawlTask{}).Where("status = ?", model.TaskStatusFailed).Count(&stats.FailedTasks)

	return stats, nil
}

// GetRunningAccounts 获取正在运行的账号
func (m *TaskManager) GetRunningAccounts() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	accounts := make([]string, 0, len(m.pools))
	for accountID, pool := range m.pools {
		if pool.IsRunning() {
			accounts = append(accounts, accountID)
		}
	}
	return accounts
}

// CancelTask 取消任务
func (m *TaskManager) CancelTask(taskID string) error {
	return m.db.Model(&model.CrawlTask{}).Where("task_id = ?", taskID).Update("status", model.TaskStatusCancelled).Error
}

// GetTasksByStatus 按状态获取任务列表
func (m *TaskManager) GetTasksByStatus(status int, limit int) ([]model.CrawlTask, error) {
	var tasks []model.CrawlTask
	query := m.db.Where("status = ?", status).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetPool 获取指定账号的协程池
func (m *TaskManager) GetPool(accountID string) *WorkerPool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.pools[accountID]
}

// CrawlHourlyReport 抓取小时报表
func (m *TaskManager) CrawlHourlyReport(task model.CrawlTaskItem) error {
	if m.hourlyReportCrawler == nil {
		return fmt.Errorf("小时报表抓取器未初始化")
	}
	return m.hourlyReportCrawler.Crawl(m.ctx, task)
}

// CrawlDailyReport 抓取日报表
func (m *TaskManager) CrawlDailyReport(task model.CrawlTaskItem) error {
	if m.dailyReportCrawler == nil {
		return fmt.Errorf("日报表抓取器未初始化")
	}
	return m.dailyReportCrawler.Crawl(m.ctx, task)
}

// CrawlCampaign 抓取广告系列
func (m *TaskManager) CrawlCampaign(task model.CrawlTaskItem) error {
	if m.campaignCrawler == nil {
		return fmt.Errorf("广告系列抓取器未初始化")
	}
	return m.campaignCrawler.Crawl(m.ctx, task)
}

// CrawlAdGroup 抓取广告组
func (m *TaskManager) CrawlAdGroup(task model.CrawlTaskItem) error {
	if m.adGroupCrawler == nil {
		return fmt.Errorf("广告组抓取器未初始化")
	}
	return m.adGroupCrawler.Crawl(m.ctx, task)
}

// CrawlAd 抓取广告
func (m *TaskManager) CrawlAd(task model.CrawlTaskItem) error {
	if m.adCrawler == nil {
		return fmt.Errorf("广告抓取器未初始化")
	}
	return m.adCrawler.Crawl(m.ctx, task)
}

// CrawlCreative 抓取广告创意
func (m *TaskManager) CrawlCreative(task model.CrawlTaskItem) error {
	if m.creativeCrawler == nil {
		return fmt.Errorf("广告创意抓取器未初始化")
	}
	return m.creativeCrawler.Crawl(m.ctx, task)
}

// CrawlMaterial 抓取广告素材
func (m *TaskManager) CrawlMaterial(task model.CrawlTaskItem) error {
	if m.materialCrawler == nil {
		return fmt.Errorf("广告素材抓取器未初始化")
	}
	return m.materialCrawler.Crawl(m.ctx, task)
}

// CrawlCampaignCascade 级联抓取：广告系列 → 广告组 → 广告 → 创意 → 素材
func (m *TaskManager) CrawlCampaignCascade(ctx context.Context, accountID, campaignID string) error {
	log.Printf("[TaskManager] 级联抓取广告系列 %s (账号: %s)", campaignID, accountID)

	// Step 1: 抓取该广告系列下的广告组
	adgroupIDs, err := m.campaignCrawler.CrawlAdGroupsForCampaign(ctx, accountID, campaignID)
	if err != nil {
		return fmt.Errorf("抓取广告组失败: %w", err)
	}
	log.Printf("[TaskManager] 获取到 %d 个广告组", len(adgroupIDs))

	// Step 2: 抓取每个广告组下的广告
	allAdIDs := make([]string, 0)
	for _, adgroupID := range adgroupIDs {
		adIDs, err := m.adCrawler.CrawlAdsForAdGroup(ctx, accountID, adgroupID)
		if err != nil {
			log.Printf("[TaskManager] 广告组 %s 抓取广告失败: %v", adgroupID, err)
			continue
		}
		allAdIDs = append(allAdIDs, adIDs...)
	}
	log.Printf("[TaskManager] 获取到 %d 个广告", len(allAdIDs))

	// Step 3: 抓取每个广告下的创意
	for _, adID := range allAdIDs {
		if err := m.creativeCrawler.CrawlCreativesForAd(ctx, accountID, adID); err != nil {
			log.Printf("[TaskManager] 广告 %s 抓取创意失败: %v", adID, err)
		}
	}

	// Step 4: 抓取该账号的素材
	if err := m.materialCrawler.Crawl(ctx, model.CrawlTaskItem{AccountID: accountID}); err != nil {
		log.Printf("[TaskManager] 抓取素材失败: %v", err)
	}

	log.Printf("[TaskManager] 广告系列 %s 级联抓取完成", campaignID)
	return nil
}

// CrawlAdGroupCascade 级联抓取：广告组 → 广告 → 创意 → 素材
func (m *TaskManager) CrawlAdGroupCascade(ctx context.Context, accountID, adgroupID string) error {
	log.Printf("[TaskManager] 级联抓取广告组 %s (账号: %s)", adgroupID, accountID)

	// Step 1: 抓取该广告组下的广告
	adIDs, err := m.adCrawler.CrawlAdsForAdGroup(ctx, accountID, adgroupID)
	if err != nil {
		return fmt.Errorf("抓取广告失败: %w", err)
	}
	log.Printf("[TaskManager] 获取到 %d 个广告", len(adIDs))

	// Step 2: 抓取每个广告下的创意
	for _, adID := range adIDs {
		if err := m.creativeCrawler.CrawlCreativesForAd(ctx, accountID, adID); err != nil {
			log.Printf("[TaskManager] 广告 %s 抓取创意失败: %v", adID, err)
		}
	}

	// Step 3: 抓取该账号的素材
	if err := m.materialCrawler.Crawl(ctx, model.CrawlTaskItem{AccountID: accountID}); err != nil {
		log.Printf("[TaskManager] 抓取素材失败: %v", err)
	}

	log.Printf("[TaskManager] 广告组 %s 级联抓取完成", adgroupID)
	return nil
}

// CrawlAdCascade 级联抓取：广告 → 创意 → 素材
func (m *TaskManager) CrawlAdCascade(ctx context.Context, accountID, adID string) error {
	log.Printf("[TaskManager] 级联抓取广告 %s (账号: %s)", adID, accountID)

	// Step 1: 抓取该广告下的创意
	if err := m.creativeCrawler.CrawlCreativesForAd(ctx, accountID, adID); err != nil {
		return fmt.Errorf("抓取创意失败: %w", err)
	}

	// Step 2: 抓取该账号的素材
	if err := m.materialCrawler.Crawl(ctx, model.CrawlTaskItem{AccountID: accountID}); err != nil {
		log.Printf("[TaskManager] 抓取素材失败: %v", err)
	}

	log.Printf("[TaskManager] 广告 %s 级联抓取完成", adID)
	return nil
}
