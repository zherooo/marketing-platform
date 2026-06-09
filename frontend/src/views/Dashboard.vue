<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #409eff;">
            <el-icon><View /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日曝光</div>
            <div class="stat-value">{{ formatNumber(statistics.viewCount) }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #67c23a;">
            <el-icon><Pointer /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日点击</div>
            <div class="stat-value">{{ formatNumber(statistics.clickCount) }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #f56c6c;">
            <el-icon><Money /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日花费</div>
            <div class="stat-value">¥{{ statistics.spend.toFixed(2) }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #e6a23c;">
            <el-icon><SuccessFilled /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日转化</div>
            <div class="stat-value">{{ formatNumber(statistics.conversionCount) }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 趋势图表 -->
    <el-row :gutter="16">
      <el-col :span="16">
        <div class="page-card">
          <div class="card-header">
            <h3>数据趋势</h3>
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              @change="fetchTrendData"
            />
          </div>
          <div ref="trendChartRef" class="chart-container"></div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="page-card">
          <div class="card-header">
            <h3>账户概览</h3>
          </div>
          <div class="account-overview">
            <div class="overview-item">
              <span class="label">在线账号</span>
              <span class="value">{{ crawlerStats.onlineAccounts }} / {{ crawlerStats.totalAccounts }}</span>
            </div>
            <div class="overview-item">
              <span class="label">今日抓取</span>
              <span class="value">{{ crawlerStats.todayCrawl }} 次</span>
            </div>
            <div class="overview-item">
              <span class="label">进行中任务</span>
              <span class="value">{{ crawlerStats.runningTasks }}</span>
            </div>
            <div class="overview-item">
              <span class="label">成功率</span>
              <span class="value">{{ crawlerStats.successRate }}%</span>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 关键指标图表 -->
    <el-row :gutter="16">
      <el-col :span="12">
        <div class="page-card">
          <div class="card-header">
            <h3>花费占比</h3>
          </div>
          <div ref="spendPieChartRef" class="chart-container" style="height: 300px;"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="page-card">
          <div class="card-header">
            <h3>转化漏斗</h3>
          </div>
          <div ref="funnelChartRef" class="chart-container" style="height: 300px;"></div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useCrawlerStore } from '@/stores'
import { getDailyTrend, getCrawlStatistics } from '@/api'
import * as echarts from 'echarts'
import type { DailyReport, CrawlStatistics } from '@/types'

const crawlerStore = useCrawlerStore()

const dateRange = ref<string[]>([])
const trendChartRef = ref<HTMLElement>()
const spendPieChartRef = ref<HTMLElement>()
const funnelChartRef = ref<HTMLElement>()

let trendChart: echarts.ECharts | null = null
let spendPieChart: echarts.ECharts | null = null
let funnelChart: echarts.ECharts | null = null

const statistics = reactive({
  viewCount: 0,
  clickCount: 0,
  spend: 0,
  conversionCount: 0
})

const crawlerStats = reactive({
  onlineAccounts: 0,
  totalAccounts: 0,
  todayCrawl: 0,
  runningTasks: 0,
  successRate: 0
})

// 初始化日期范围（最近7天）
const initDateRange = () => {
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - 7)
  dateRange.value = [
    start.toISOString().split('T')[0],
    end.toISOString().split('T')[0]
  ]
}

// 格式化数字
const formatNumber = (num: number): string => {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + 'w'
  }
  return num.toLocaleString()
}

// 获取趋势数据
const fetchTrendData = async () => {
  if (!dateRange.value || dateRange.value.length !== 2) return
  
  try {
    const res = await getDailyTrend({
      account_id: crawlerStore.currentAccount,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1]
    })
    
    const data = res.data.list || []
    updateStatistics(data)
    updateTrendChart(data)
  } catch (error) {
    console.error('获取趋势数据失败', error)
  }
}

// 更新统计数据
const updateStatistics = (data: DailyReport[]) => {
  if (data.length > 0) {
    const today = data[data.length - 1]
    statistics.viewCount = today.view_count
    statistics.clickCount = today.click_count
    statistics.spend = today.spend
    statistics.conversionCount = today.conversion_count
  }
  
  // 更新花费饼图
  updateSpendPieChart(data)
  
  // 更新漏斗图
  updateFunnelChart(data)
}

// 更新趋势图表
const updateTrendChart = (data: DailyReport[]) => {
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value!)
  }
  
  const dates = data.map(item => item.date)
  const viewData = data.map(item => item.view_count)
  const clickData = data.map(item => item.click_count)
  const spendData = data.map(item => item.spend)
  
  trendChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    legend: {
      data: ['曝光量', '点击量', '花费']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dates,
      boundaryGap: false
    },
    yAxis: [
      {
        type: 'value',
        name: '数量',
        position: 'left'
      },
      {
        type: 'value',
        name: '花费(元)',
        position: 'right'
      }
    ],
    series: [
      {
        name: '曝光量',
        type: 'line',
        data: viewData,
        smooth: true,
        itemStyle: { color: '#409eff' }
      },
      {
        name: '点击量',
        type: 'line',
        data: clickData,
        smooth: true,
        itemStyle: { color: '#67c23a' }
      },
      {
        name: '花费',
        type: 'line',
        yAxisIndex: 1,
        data: spendData,
        smooth: true,
        itemStyle: { color: '#f56c6c' }
      }
    ]
  })
}

// 更新花费饼图
const updateSpendPieChart = (data: DailyReport[]) => {
  if (!spendPieChart) {
    spendPieChart = echarts.init(spendPieChartRef.value!)
  }
  
  const totalSpend = data.reduce((sum, item) => sum + item.spend, 0)
  
  spendPieChart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: ¥{c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: '{b}\n¥{c}'
        },
        data: data.map((item, index) => ({
          name: item.date,
          value: item.spend
        }))
      }
    ]
  })
}

// 更新漏斗图
const updateFunnelChart = (data: DailyReport[]) => {
  if (!funnelChart) {
    funnelChart = echarts.init(funnelChartRef.value!)
  }
  
  const total = data.reduce((sum, item) => sum + item.view_count, 0)
  const clicks = data.reduce((sum, item) => sum + item.click_count, 0)
  const conversions = data.reduce((sum, item) => sum + item.conversion_count, 0)
  
  funnelChart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}'
    },
    series: [
      {
        type: 'funnel',
        left: '10%',
        top: 20,
        bottom: 20,
        width: '80%',
        min: 0,
        max: total,
        minSize: '0%',
        maxSize: '100%',
        sort: 'descending',
        gap: 2,
        label: {
          show: true,
          position: 'inside',
          formatter: '{b}\n{c}'
        },
        itemStyle: {
          borderColor: '#fff',
          borderWidth: 1
        },
        data: [
          { name: '曝光', value: total },
          { name: '点击', value: clicks },
          { name: '转化', value: conversions }
        ]
      }
    ]
  })
}

// 获取抓取统计
const fetchCrawlerStats = async () => {
  try {
    const res = await getCrawlStatistics()
    const stats = res.data
    crawlerStats.onlineAccounts = stats.total_tasks - stats.running_tasks - stats.pending_tasks
    crawlerStats.totalAccounts = stats.total_tasks
    crawlerStats.todayCrawl = stats.today_crawl_count
    crawlerStats.runningTasks = stats.running_tasks
    crawlerStats.successRate = stats.total_tasks > 0 
      ? Math.round((stats.success_tasks / stats.total_tasks) * 100) 
      : 0
    crawlerStore.setStatistics(stats)
  } catch (error) {
    console.error('获取抓取统计失败', error)
  }
}

// 窗口resize处理
const handleResize = () => {
  trendChart?.resize()
  spendPieChart?.resize()
  funnelChart?.resize()
}

onMounted(() => {
  initDateRange()
  nextTick(() => {
    fetchTrendData()
    fetchCrawlerStats()
  })
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  spendPieChart?.dispose()
  funnelChart?.dispose()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stat-row {
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.chart-container {
  width: 100%;
  height: 350px;
}

.account-overview {
  padding: 10px 0;
}

.overview-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.overview-item:last-child {
  border-bottom: none;
}

.overview-item .label {
  color: #909399;
}

.overview-item .value {
  font-weight: 600;
  color: #303133;
}
</style>
