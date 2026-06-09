package crawler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"marketing-platform/internal/model"
)

// WorkerPool 协程池 - 每个账号独立协程池，最多5个worker
type WorkerPool struct {
	accountID  string
	maxWorkers int
	taskQueue  chan model.CrawlTaskItem
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mu         sync.RWMutex

	// 依赖注入
	taskManager *TaskManager
}

// NewWorkerPool 创建协程池
func NewWorkerPool(accountID string, maxWorkers int, taskManager *TaskManager) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		accountID:   accountID,
		maxWorkers:  maxWorkers,
		taskQueue:   make(chan model.CrawlTaskItem, 100),
		ctx:         ctx,
		cancel:      cancel,
		taskManager: taskManager,
	}
}

// Start 启动协程池
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return
	}

	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	p.running = true
	log.Printf("[WorkerPool] 账号 %s 启动 %d 个 worker", p.accountID, p.maxWorkers)
}

// worker 工作协程
func (p *WorkerPool) worker(workerID int) {
	defer p.wg.Done()

	log.Printf("[Worker-%d] 账号 %s worker 启动", workerID, p.accountID)

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				log.Printf("[Worker-%d] 账号 %s 任务队列已关闭", workerID, p.accountID)
				return
			}
			p.processTask(workerID, task)
		case <-p.ctx.Done():
			log.Printf("[Worker-%d] 账号 %s 收到停止信号", workerID, p.accountID)
			return
		}
	}
}

// processTask 处理单个任务
func (p *WorkerPool) processTask(workerID int, task model.CrawlTaskItem) {
	log.Printf("[Worker-%d] 账号 %s 开始处理任务: %s, 数据类型: %s", workerID, p.accountID, task.TaskID, task.DataType)

	// 更新任务状态为运行中
	if err := p.taskManager.UpdateTaskStatus(task.TaskID, model.TaskStatusRunning); err != nil {
		log.Printf("[Worker-%d] 更新任务状态失败: %v", workerID, err)
	}

	// 更新账号在线状态
	p.taskManager.UpdateAccountOnline(p.accountID)

	// 根据数据类型执行抓取
	var crawlErr error
	switch task.DataType {
	case model.DataTypeHourlyReport:
		crawlErr = p.taskManager.CrawlHourlyReport(task)
	case model.DataTypeDailyReport:
		crawlErr = p.taskManager.CrawlDailyReport(task)
	case model.DataTypeCampaign:
		crawlErr = p.taskManager.CrawlCampaign(task)
	case model.DataTypeAdGroup:
		crawlErr = p.taskManager.CrawlAdGroup(task)
	case model.DataTypeAd:
		crawlErr = p.taskManager.CrawlAd(task)
	case model.DataTypeCreative:
		crawlErr = p.taskManager.CrawlCreative(task)
	case model.DataTypeMaterial:
		crawlErr = p.taskManager.CrawlMaterial(task)
	default:
		crawlErr = fmt.Errorf("未知的数据类型: %s", task.DataType)
	}

	// 处理结果
	if crawlErr != nil {
		log.Printf("[Worker-%d] 账号 %s 任务 %s 执行失败: %v", workerID, p.accountID, task.TaskID, crawlErr)
		if err := p.taskManager.HandleTaskError(task.TaskID, crawlErr.Error()); err != nil {
			log.Printf("[Worker-%d] 更新任务错误状态失败: %v", workerID, err)
		}
	} else {
		log.Printf("[Worker-%d] 账号 %s 任务 %s 执行成功", workerID, p.accountID, task.TaskID)
		if err := p.taskManager.UpdateTaskStatus(task.TaskID, model.TaskStatusCompleted); err != nil {
			log.Printf("[Worker-%d] 更新任务完成状态失败: %v", workerID, err)
		}
	}

	// 更新账号在线状态
	p.taskManager.UpdateAccountLastCrawlTime(p.accountID)
}

// AddTask 添加任务到队列
func (p *WorkerPool) AddTask(task model.CrawlTaskItem) bool {
	p.mu.RLock()
	running := p.running
	p.mu.RUnlock()

	if !running {
		log.Printf("[WorkerPool] 账号 %s 协程池未启动，任务 %s 加入队列失败", p.accountID, task.TaskID)
		return false
	}

	select {
	case p.taskQueue <- task:
		log.Printf("[WorkerPool] 任务 %s 加入账号 %s 队列成功", task.TaskID, p.accountID)
		return true
	case <-time.After(5 * time.Second):
		log.Printf("[WorkerPool] 任务 %s 加入账号 %s 队列超时", task.TaskID, p.accountID)
		return false
	}
}

// Stop 停止协程池
func (p *WorkerPool) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	log.Printf("[WorkerPool] 账号 %s 协程池停止中...", p.accountID)
	p.cancel()

	// 等待所有 worker 结束
	p.wg.Wait()
	close(p.taskQueue)

	p.mu.Lock()
	p.running = false
	p.mu.Unlock()

	log.Printf("[WorkerPool] 账号 %s 协程池已停止", p.accountID)
}

// IsRunning 检查协程池是否运行中
func (p *WorkerPool) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// QueueLength 获取队列长度
func (p *WorkerPool) QueueLength() int {
	return len(p.taskQueue)
}
