<template>
  <div class="page">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #409eff;">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">总任务数</div>
            <div class="stat-value">{{ statistics?.total_tasks || 0 }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #67c23a;">
            <el-icon><SuccessFilled /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">成功</div>
            <div class="stat-value">{{ statistics?.success_tasks || 0 }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #f56c6c;">
            <el-icon><CircleCloseFilled /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">失败</div>
            <div class="stat-value">{{ statistics?.fail_tasks || 0 }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #e6a23c;">
            <el-icon><Loading /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">进行中</div>
            <div class="stat-value">{{ statistics?.running_tasks || 0 }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 操作按钮 -->
    <div class="page-card">
      <div class="action-buttons">
        <el-button type="primary" @click="handleStartCrawl" :loading="startLoading">
          <el-icon><VideoPlay /></el-icon>
          启动抓取
        </el-button>
        <el-button @click="handleTriggerHourly" :loading="triggerLoading">
          <el-icon><Clock /></el-icon>
          触发小时报表
        </el-button>
        <el-button @click="handleTriggerDaily" :loading="triggerLoading">
          <el-icon><Calendar /></el-icon>
          触发动报表
        </el-button>
        <el-button @click="handleTriggerStruct" :loading="triggerLoading">
          <el-icon><Grid /></el-icon>
          触发广告结构
        </el-button>
        <el-button @click="handleTriggerCampaign" :loading="triggerLoading">
          <el-icon><TrendCharts /></el-icon>
          触发广告系列
        </el-button>
        <el-button @click="fetchStatistics">
          <el-icon><Refresh /></el-icon>
          刷新统计
        </el-button>
      </div>
    </div>

    <!-- 任务列表 -->
    <div class="page-card">
      <div class="page-header">
        <h3 class="page-title">任务列表</h3>
        <div class="filter-area">
          <el-select v-model="filterStatus" placeholder="任务状态" clearable>
            <el-option label="进行中" :value="1" />
            <el-option label="已完成" :value="2" />
            <el-option label="已失败" :value="3" />
            <el-option label="已取消" :value="4" />
          </el-select>
          <el-button type="primary" @click="fetchTasks">筛选</el-button>
        </div>
      </div>
      
      <el-table :data="tasks" v-loading="loading" stripe>
        <el-table-column prop="task_id" label="任务ID" width="150" />
        <el-table-column label="任务类型" width="120">
          <template #default="{ row }">
            {{ getTaskTypeName(row.task_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="account_id" label="账号ID" width="150" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="进度" width="200">
          <template #default="{ row }">
            <el-progress
              :percentage="row.progress > 0 ? Math.round((row.progress / row.total) * 100) : 0"
              :status="getProgressStatus(row.status)"
            />
            <span class="progress-text">{{ row.progress }} / {{ row.total }}</span>
          </template>
        </el-table-column>
        <el-table-column label="结果" width="150">
          <template #default="{ row }">
            <span class="success-count">{{ row.success_count }} 成功</span>
            <span class="fail-count">{{ row.fail_count }} 失败</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              type="danger"
              size="small"
              text
              @click="handleCancel(row)"
            >
              取消
            </el-button>
            <el-button
              v-if="row.status === 3"
              type="primary"
              size="small"
              text
              @click="handleRetry(row)"
            >
              重试
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchTasks"
        @current-change="fetchTasks"
      />
    </div>

    <!-- 启动抓取弹窗 -->
    <el-dialog v-model="startDialogVisible" title="启动抓取" width="500px">
      <el-form :model="startForm" label-width="100px">
        <el-form-item label="任务类型" required>
          <el-select v-model="startForm.task_type" placeholder="请选择任务类型">
            <el-option label="小时报表" value="hourly_report" />
            <el-option label="日报表" value="daily_report" />
            <el-option label="广告系列" value="campaign" />
            <el-option label="广告组" value="adgroup" />
            <el-option label="广告" value="ad" />
            <el-option label="广告创意" value="creative" />
            <el-option label="广告素材" value="material" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期范围">
          <el-select v-model="startForm.date_range" placeholder="请选择">
            <el-option label="今天" value="today" />
            <el-option label="昨天" value="yesterday" />
            <el-option label="最近7天" value="last_7_days" />
            <el-option label="最近30天" value="last_30_days" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="startDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmStartCrawl" :loading="startLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useCrawlerStore } from '@/stores'
import {
  getCrawlStatistics,
  getCrawlTasks,
  startCrawl,
  triggerHourlyReport,
  triggerDailyReport,
  triggerAllStruct,
  triggerCampaign,
  cancelTask,
  retryTask
} from '@/api'
import type { CrawlTask, CrawlStatistics } from '@/types'

const crawlerStore = useCrawlerStore()
const loading = ref(false)
const startLoading = ref(false)
const triggerLoading = ref(false)
const statistics = ref<CrawlStatistics | null>(null)
const tasks = ref<CrawlTask[]>([])
const filterStatus = ref<number | undefined>()
const startDialogVisible = ref(false)

const startForm = reactive({
  task_type: 'hourly_report',
  date_range: 'today'
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

let refreshTimer: number | null = null

const fetchStatistics = async () => {
  try {
    const res = await getCrawlStatistics()
    statistics.value = res.data
    crawlerStore.setStatistics(res.data)
  } catch (error) {
    console.error('获取统计失败', error)
  }
}

const fetchTasks = async () => {
  loading.value = true
  try {
    const res = await getCrawlTasks({
      status: filterStatus.value,
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tasks.value = res.data.list || []
    pagination.total = res.data.total
    crawlerStore.setTasks(res.data.list)
  } catch (error) {
    console.error('获取任务列表失败', error)
  } finally {
    loading.value = false
  }
}

const handleStartCrawl = () => {
  startDialogVisible.value = true
}

const confirmStartCrawl = async () => {
  startLoading.value = true
  try {
    await startCrawl({
      task_type: startForm.task_type,
      date_range: startForm.date_range
    })
    ElMessage.success('任务已启动')
    startDialogVisible.value = false
    fetchTasks()
    fetchStatistics()
  } catch (error) {
    console.error('启动抓取失败', error)
  } finally {
    startLoading.value = false
  }
}

const handleTriggerHourly = async () => {
  triggerLoading.value = true
  try {
    await triggerHourlyReport()
    ElMessage.success('小时报表抓取已触发')
    fetchTasks()
  } catch (error) {
    console.error('触发失败', error)
  } finally {
    triggerLoading.value = false
  }
}

const handleTriggerDaily = async () => {
  triggerLoading.value = true
  try {
    await triggerDailyReport()
    ElMessage.success('日报表抓取已触发')
    fetchTasks()
  } catch (error) {
    console.error('触发失败', error)
  } finally {
    triggerLoading.value = false
  }
}

const handleTriggerStruct = async () => {
  triggerLoading.value = true
  try {
    await triggerAllStruct()
    ElMessage.success('广告结构抓取已触发')
    fetchTasks()
  } catch (error) {
    console.error('触发失败', error)
  } finally {
    triggerLoading.value = false
  }
}

const handleTriggerCampaign = async () => {
  triggerLoading.value = true
  try {
    await triggerCampaign()
    ElMessage.success('广告系列抓取已触发')
    fetchTasks()
  } catch (error) {
    console.error('触发失败', error)
  } finally {
    triggerLoading.value = false
  }
}

const handleCancel = async (row: CrawlTask) => {
  try {
    await ElMessageBox.confirm('确定要取消该任务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await cancelTask(row.task_id)
    ElMessage.success('任务已取消')
    fetchTasks()
    fetchStatistics()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('取消任务失败', error)
    }
  }
}

const handleRetry = async (row: CrawlTask) => {
  try {
    await retryTask(row.task_id)
    ElMessage.success('任务已重试')
    fetchTasks()
    fetchStatistics()
  } catch (error) {
    console.error('重试任务失败', error)
  }
}

const getTaskTypeName = (type: string): string => {
  const types: Record<string, string> = {
    hourly_report: '小时报表',
    daily_report: '日报表',
    campaign: '广告系列',
    adgroup: '广告组',
    ad: '广告',
    creative: '广告创意',
    material: '广告素材'
  }
  return types[type] || type
}

const getStatusName = (status: number): string => {
  const statuses: Record<number, string> = {
    0: '等待中',
    1: '进行中',
    2: '已完成',
    3: '已失败',
    4: '已取消'
  }
  return statuses[status] || '未知'
}

const getStatusType = (status: number): string => {
  const types: Record<number, string> = {
    0: 'info',
    1: 'warning',
    2: 'success',
    3: 'danger',
    4: 'info'
  }
  return types[status] || 'info'
}

const getProgressStatus = (status: number): string => {
  if (status === 3) return 'exception'
  if (status === 2) return 'success'
  return ''
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchStatistics()
  fetchTasks()
  
  // 每30秒自动刷新
  refreshTimer = window.setInterval(() => {
    fetchStatistics()
  }, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.page {
  padding: 0;
}

.stats-row {
  margin-bottom: 16px;
}

.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 24px;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filter-area {
  display: flex;
  gap: 12px;
}

.progress-text {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

.success-count {
  color: #67c23a;
  margin-right: 8px;
}

.fail-count {
  color: #f56c6c;
}
</style>
