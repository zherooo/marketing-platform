<template>
  <div class="page">
    <div class="page-card">
      <div class="page-header">
        <h3 class="page-title">广告系列</h3>
      </div>
      
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="campaign_id" label="系列ID" width="200" />
        <el-table-column prop="campaign_name" label="系列名称" min-width="200" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            {{ getCampaignTypeName(row.campaign_type) }}
          </template>
        </el-table-column>
        <el-table-column label="日预算" width="120">
          <template #default="{ row }">
            ¥{{ row.daily_budget.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="viewDetail(row)">详情</el-button>
            <el-button type="success" link @click="handleCascadeCrawl(row)" :loading="crawlingId === row.campaign_id">
              抓取
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
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </div>
    
    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="广告系列详情" width="600px">
      <el-descriptions :column="2" border v-if="currentRow">
        <el-descriptions-item label="系列ID">{{ currentRow.campaign_id }}</el-descriptions-item>
        <el-descriptions-item label="账号ID">{{ currentRow.account_id }}</el-descriptions-item>
        <el-descriptions-item label="系列名称" :span="2">{{ currentRow.campaign_name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ getCampaignTypeName(currentRow.campaign_type) }}</el-descriptions-item>
        <el-descriptions-item label="日预算">¥{{ currentRow.daily_budget.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentRow.status)" size="small">
            {{ getStatusName(currentRow.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(currentRow.created_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getCampaigns, triggerCampaignCascade } from '@/api'
import type { Campaign } from '@/types'

const loading = ref(false)
const tableData = ref<Campaign[]>([])
const detailVisible = ref(false)
const currentRow = ref<Campaign | null>(null)
const crawlingId = ref('')

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getCampaigns({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total
  } catch (error) {
    console.error('获取广告系列失败', error)
  } finally {
    loading.value = false
  }
}

const viewDetail = (row: Campaign) => {
  currentRow.value = row
  detailVisible.value = true
}

const handleCascadeCrawl = async (row: Campaign) => {
  crawlingId.value = row.campaign_id
  try {
    await triggerCampaignCascade(row.campaign_id, row.account_id)
    ElMessage.success('级联抓取已触发，正在抓取该系列下的广告组、广告、创意、素材')
  } catch (error) {
    ElMessage.error('触发失败')
    console.error(error)
  } finally {
    crawlingId.value = ''
  }
}

const getCampaignTypeName = (type: number): string => {
  const types: Record<number, string> = {
    1: '搜索广告',
    2: '展示广告',
    3: '视频广告'
  }
  return types[type] || '未知'
}

const getStatusName = (status: number): string => {
  const statuses: Record<number, string> = {
    0: '无效',
    1: '有效',
    2: '暂停',
    3: '已删除'
  }
  return statuses[status] || '未知'
}

const getStatusType = (status: number): string => {
  const types: Record<number, string> = {
    0: 'info',
    1: 'success',
    2: 'warning',
    3: 'danger'
  }
  return types[status] || 'info'
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
</style>
