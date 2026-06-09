<template>
  <div class="report-page">
    <div class="page-card">
      <div class="page-header">
        <h3 class="page-title">定向报表</h3>
        <div class="filter-area">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
          <el-button type="primary" @click="fetchData" :loading="loading">查询</el-button>
        </div>
      </div>
      
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column prop="target_name" label="定向名称" min-width="150" />
        <el-table-column prop="target_id" label="定向ID" width="180" />
        <el-table-column prop="view_count" label="曝光量" width="120" sortable>
          <template #default="{ row }">
            {{ row.view_count.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="click_count" label="点击量" width="100" sortable>
          <template #default="{ row }">
            {{ row.click_count.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="spend" label="花费(元)" width="100" sortable>
          <template #default="{ row }">
            {{ row.spend.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="ctr" label="CTR" width="80" sortable>
          <template #default="{ row }">
            {{ (row.ctr * 100).toFixed(2) }}%
          </template>
        </el-table-column>
        <el-table-column prop="conversion_count" label="转化数" width="100" sortable>
          <template #default="{ row }">
            {{ row.conversion_count.toLocaleString() }}
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useCrawlerStore } from '@/stores'
import { getTargetReports } from '@/api'
import type { TargetReport } from '@/types'

const crawlerStore = useCrawlerStore()
const loading = ref(false)

const dateRange = ref<string[]>([])
const tableData = ref<TargetReport[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const initDateRange = () => {
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - 7)
  dateRange.value = [
    start.toISOString().split('T')[0],
    end.toISOString().split('T')[0]
  ]
}

const fetchData = async () => {
  if (!dateRange.value || dateRange.value.length !== 2) {
    ElMessage.warning('请选择日期范围')
    return
  }
  
  loading.value = true
  try {
    const res = await getTargetReports({
      account_id: crawlerStore.currentAccount,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1],
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total
  } catch (error) {
    console.error('获取定向报表失败', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  initDateRange()
  fetchData()
})
</script>

<style scoped>
.report-page {
  padding: 0;
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
</style>
