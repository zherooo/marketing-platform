<template>
  <div class="hosting-dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value success">{{ stats.active_rules }}</div>
            <div class="stat-label">启用规则</div>
            <div class="stat-sub">共 {{ stats.total_rules }} 条规则</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value primary">{{ stats.today_exec }}</div>
            <div class="stat-label">今日执行</div>
            <div class="stat-sub">累计 {{ stats.total_exec }} 次</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value warning">{{ stats.unread_alerts }}</div>
            <div class="stat-label">未读告警</div>
            <div class="stat-sub">共 {{ stats.total_alerts }} 条</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value info">{{ stats.success_exec }}/{{ stats.total_exec }}</div>
            <div class="stat-label">成功率</div>
            <div class="stat-sub">失败 {{ stats.failed_exec }} 次</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 操作按钮 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="24">
        <el-card>
          <div class="action-bar">
            <el-button type="primary" :loading="evaluating" @click="handleEvaluate">
              <el-icon><VideoPlay /></el-icon> 立即评估
            </el-button>
            <el-button type="success" :loading="collecting" @click="handleCollect">
              <el-icon><RefreshRight /></el-icon> 采集快照
            </el-button>
            <el-button @click="$router.push('/hosting/rules/create')">
              <el-icon><Plus /></el-icon> 创建规则
            </el-button>
            <el-button @click="$router.push('/hosting/rules')">
              <el-icon><List /></el-icon> 管理规则
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 趋势图 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span>执行趋势（近7天）</span>
          </template>
          <div ref="chartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近执行记录 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近执行记录</span>
              <el-button text @click="$router.push('/hosting/executions')">查看更多</el-button>
            </div>
          </template>
          <el-table :data="executions" stripe v-loading="loading" style="width: 100%">
            <el-table-column prop="rule_name" label="规则名称" min-width="140" />
            <el-table-column prop="account_id" label="账号ID" width="160" />
            <el-table-column prop="target_id" label="目标ID" width="160" />
            <el-table-column prop="action_type" label="动作类型" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ getActionLabel(row.action_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="error_msg" label="错误信息" min-width="160" show-overflow-tooltip />
            <el-table-column prop="executed_at" label="执行时间" width="170">
              <template #default="{ row }">
                {{ formatTime(row.executed_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { getHostingDashboard, getHostingExecutions, triggerEvaluate, triggerCollectSnapshots } from '@/api'
import type { HostingDashboardStats, HostingExecution } from '@/types'
import * as echarts from 'echarts'

const stats = ref<HostingDashboardStats>({
  active_rules: 0, total_rules: 0, today_exec: 0, total_exec: 0,
  success_exec: 0, failed_exec: 0, unread_alerts: 0, total_alerts: 0
})
const executions = ref<HostingExecution[]>([])
const loading = ref(false)
const evaluating = ref(false)
const collecting = ref(false)
const chartRef = ref<HTMLElement>()

let chartInstance: echarts.ECharts | null = null

const actionLabels: Record<string, string> = {
  pause_ad: '暂停广告', adjust_bid: '调整出价', raise_budget: '提升预算',
  notify: '发送通知', resume_ad: '启用广告', quick_start: '一键起量'
}
const statusLabels: Record<number, string> = { 1: '成功', 2: '失败', 3: '待执行', 4: '已回滚' }
const getActionLabel = (type: string) => actionLabels[type] || type
const getStatusLabel = (s: number) => statusLabels[s] || '未知'
const getStatusType = (s: number) => s === 1 ? 'success' : s === 2 ? 'danger' : s === 4 ? 'info' : 'warning'
const formatTime = (t?: string) => t ? new Date(t).toLocaleString('zh-CN') : '-'

const fetchData = async () => {
  loading.value = true
  try {
    const [dashRes, execRes] = await Promise.all([
      getHostingDashboard(),
      getHostingExecutions({ page: 1, page_size: 10 })
    ])
    stats.value = dashRes.data.stats
    executions.value = execRes.data.list || []
    await nextTick()
    renderChart(dashRes.data.trend || [])
  } catch {
    ElMessage.error('获取看板数据失败')
  } finally {
    loading.value = false
  }
}

const renderChart = (trend: { date: string; count: number }[]) => {
  if (!chartRef.value) return
  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value)
  }
  chartInstance.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: trend.map(t => t.date.substring(5)) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{
      data: trend.map(t => t.count), type: 'line', smooth: true,
      areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: 'rgba(24,144,255,0.3)' },
        { offset: 1, color: 'rgba(24,144,255,0.05)' }
      ])},
      itemStyle: { color: '#1890ff' }
    }]
  })
}

const handleEvaluate = async () => {
  evaluating.value = true
  try {
    await triggerEvaluate()
    ElMessage.success('评估已触发，请稍后查看执行记录')
    setTimeout(fetchData, 3000)
  } catch {
    ElMessage.error('评估失败')
  } finally {
    evaluating.value = false
  }
}

const handleCollect = async () => {
  collecting.value = true
  try {
    await triggerCollectSnapshots()
    ElMessage.success('快照采集已触发')
  } catch {
    ElMessage.error('快照采集失败')
  } finally {
    collecting.value = false
  }
}

onMounted(fetchData)
onUnmounted(() => { chartInstance?.dispose() })
</script>

<style scoped>
.stats-row { margin-bottom: 8px }
.stat-item { text-align: center; padding: 8px 0 }
.stat-value { font-size: 32px; font-weight: bold }
.stat-value.success { color: #67c23a }
.stat-value.primary { color: #409eff }
.stat-value.warning { color: #e6a23c }
.stat-value.info { color: #909399 }
.stat-label { font-size: 14px; color: #909399; margin-top: 4px }
.stat-sub { font-size: 12px; color: #c0c4cc; margin-top: 2px }
.action-bar { display: flex; gap: 12px }
.card-header { display: flex; justify-content: space-between; align-items: center }
</style>
