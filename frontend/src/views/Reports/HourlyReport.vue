<template>
  <div class="report-page">
    <div class="page-card">
      <div class="page-header">
        <h3 class="page-title">小时报表</h3>
        <div class="filter-area">
          <el-date-picker
            v-model="selectedDate"
            type="date"
            placeholder="选择日期"
            value-format="YYYY-MM-DD"
          />
          <el-button type="primary" @click="fetchData" :loading="loading">查询</el-button>
        </div>
      </div>
      
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column prop="hour" label="小时" width="100">
          <template #default="{ row }">
            {{ row.hour }}:00
          </template>
        </el-table-column>
        <el-table-column prop="view_count" label="曝光量" width="120" sortable>
          <template #default="{ row }">
            {{ row.view_count.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="click_count" label="点击量" width="120" sortable>
          <template #default="{ row }">
            {{ row.click_count.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="spend" label="花费(元)" width="120" sortable>
          <template #default="{ row }">
            {{ row.spend.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="ctr" label="CTR" width="100" sortable>
          <template #default="{ row }">
            {{ (row.ctr * 100).toFixed(2) }}%
          </template>
        </el-table-column>
        <el-table-column prop="cpc" label="CPC" width="100" sortable>
          <template #default="{ row }">
            {{ row.cpc.toFixed(2) }}
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
import { getHourlyReports } from '@/api'
import type { HourlyReport } from '@/types'

const crawlerStore = useCrawlerStore()
const loading = ref(false)

const selectedDate = ref('')
const tableData = ref<HourlyReport[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 24,
  total: 0
})

const initDate = () => {
  selectedDate.value = new Date().toISOString().split('T')[0]
}

const fetchData = async () => {
  if (!selectedDate.value) {
    ElMessage.warning('请选择日期')
    return
  }
  
  loading.value = true
  try {
    const res = await getHourlyReports({
      account_id: crawlerStore.currentAccount,
      date: selectedDate.value,
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total
  } catch (error) {
    console.error('获取小时报表失败', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  initDate()
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
