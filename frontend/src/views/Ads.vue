<template>
  <div class="page">
    <div class="page-card">
      <div class="page-header">
        <h3 class="page-title">广告</h3>
      </div>
      
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="ad_id" label="广告ID" width="200" />
        <el-table-column prop="ad_name" label="广告名称" min-width="200" />
        <el-table-column prop="adgroup_id" label="所属组" width="200" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.ad_status)" size="small">
              {{ getStatusName(row.ad_status) }}
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
            <el-button type="success" link @click="handleCascadeCrawl(row)" :loading="crawlingId === row.ad_id">
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
    <el-dialog v-model="detailVisible" title="广告详情" width="600px">
      <el-descriptions :column="2" border v-if="currentRow">
        <el-descriptions-item label="广告ID">{{ currentRow.ad_id }}</el-descriptions-item>
        <el-descriptions-item label="账号ID">{{ currentRow.account_id }}</el-descriptions-item>
        <el-descriptions-item label="广告名称" :span="2">{{ currentRow.ad_name }}</el-descriptions-item>
        <el-descriptions-item label="所属组">{{ currentRow.adgroup_id }}</el-descriptions-item>
        <el-descriptions-item label="所属系列">{{ currentRow.campaign_id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentRow.ad_status)" size="small">
            {{ getStatusName(currentRow.ad_status) }}
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
import { getAds, triggerAdCascade } from '@/api'
import type { Ad } from '@/types'

const loading = ref(false)
const tableData = ref<Ad[]>([])
const detailVisible = ref(false)
const currentRow = ref<Ad | null>(null)
const crawlingId = ref('')

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getAds({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total
  } catch (error) {
    console.error('获取广告失败', error)
  } finally {
    loading.value = false
  }
}

const viewDetail = (row: Ad) => {
  currentRow.value = row
  detailVisible.value = true
}

const handleCascadeCrawl = async (row: Ad) => {
  crawlingId.value = row.ad_id
  try {
    await triggerAdCascade(row.ad_id, row.account_id)
    ElMessage.success('级联抓取已触发，正在抓取该广告下的创意、素材')
  } catch (error) {
    ElMessage.error('触发失败')
    console.error(error)
  } finally {
    crawlingId.value = ''
  }
}

const getStatusName = (status: number): string => {
  const statuses: Record<number, string> = {
    0: '无效',
    1: '有效',
    2: '暂停'
  }
  return statuses[status] || '未知'
}

const getStatusType = (status: number): string => {
  const types: Record<number, string> = {
    0: 'info',
    1: 'success',
    2: 'warning'
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
