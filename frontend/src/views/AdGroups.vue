<template>
  <div class="page">
    <div class="page-card">
      <div class="page-header">
        <h3 class="page-title">广告组</h3>
      </div>
      
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="adgroup_id" label="组ID" width="200" />
        <el-table-column prop="adgroup_name" label="组名称" min-width="200" />
        <el-table-column prop="campaign_id" label="所属系列" width="200" />
        <el-table-column label="出价" width="120">
          <template #default="{ row }">
            ¥{{ row.bid_amount.toFixed(2) }}
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
            <el-button type="success" link @click="handleCascadeCrawl(row)" :loading="crawlingId === row.adgroup_id">
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
    <el-dialog v-model="detailVisible" title="广告组详情" width="600px">
      <el-descriptions :column="2" border v-if="currentRow">
        <el-descriptions-item label="组ID">{{ currentRow.adgroup_id }}</el-descriptions-item>
        <el-descriptions-item label="账号ID">{{ currentRow.account_id }}</el-descriptions-item>
        <el-descriptions-item label="组名称" :span="2">{{ currentRow.adgroup_name }}</el-descriptions-item>
        <el-descriptions-item label="所属系列">{{ currentRow.campaign_id }}</el-descriptions-item>
        <el-descriptions-item label="出价">¥{{ currentRow.bid_amount.toFixed(2) }}</el-descriptions-item>
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
import { getAdGroups, triggerAdGroupCascade } from '@/api'
import type { AdGroup } from '@/types'

const loading = ref(false)
const tableData = ref<AdGroup[]>([])
const detailVisible = ref(false)
const currentRow = ref<AdGroup | null>(null)
const crawlingId = ref('')

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getAdGroups({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total
  } catch (error) {
    console.error('获取广告组失败', error)
  } finally {
    loading.value = false
  }
}

const viewDetail = (row: AdGroup) => {
  currentRow.value = row
  detailVisible.value = true
}

const handleCascadeCrawl = async (row: AdGroup) => {
  crawlingId.value = row.adgroup_id
  try {
    await triggerAdGroupCascade(row.adgroup_id, row.account_id)
    ElMessage.success('级联抓取已触发，正在抓取该组下的广告、创意、素材')
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
