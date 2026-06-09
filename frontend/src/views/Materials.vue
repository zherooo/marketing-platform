<template>
  <div class="page">
    <div class="page-card">
      <div class="page-header">
        <h3 class="page-title">广告素材</h3>
        <div class="filter-area">
          <el-select v-model="filterType" placeholder="素材类型" clearable>
            <el-option label="图片" :value="1" />
            <el-option label="视频" :value="2" />
            <el-option label="文本" :value="3" />
            <el-option label="卡片" :value="4" />
            <el-option label="小程序" :value="5" />
          </el-select>
          <el-button type="primary" @click="fetchData">筛选</el-button>
        </div>
      </div>
      
      <!-- 统计卡片 -->
      <el-row :gutter="16" class="stats-row">
        <el-col :span="4">
          <div class="mini-stat">
            <div class="value">{{ stats.total }}</div>
            <div class="label">总计</div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="mini-stat">
            <div class="value">{{ stats.image }}</div>
            <div class="label">图片</div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="mini-stat">
            <div class="value">{{ stats.video }}</div>
            <div class="label">视频</div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="mini-stat">
            <div class="value">{{ stats.text }}</div>
            <div class="label">文本</div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="mini-stat">
            <div class="value">{{ stats.card }}</div>
            <div class="label">卡片</div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="mini-stat">
            <div class="value">{{ stats.mini_app }}</div>
            <div class="label">小程序</div>
          </div>
        </el-col>
      </el-row>
      
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column label="预览" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.material_url && row.material_type === 1"
              :src="row.material_url"
              :preview-src-list="[row.material_url]"
              fit="cover"
              style="width: 60px; height: 60px"
            />
            <el-icon v-else-if="row.material_type === 2" size="32"><VideoCamera /></el-icon>
            <el-icon v-else size="32"><Document /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="material_id" label="素材ID" width="200" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            {{ getMaterialTypeName(row.material_type) }}
          </template>
        </el-table-column>
        <el-table-column label="尺寸" width="120">
          <template #default="{ row }">
            {{ row.width }} × {{ row.height }}
          </template>
        </el-table-column>
        <el-table-column label="大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '有效' : '无效' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="viewDetail(row)">详情</el-button>
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
    <el-dialog v-model="detailVisible" title="素材详情" width="600px">
      <el-descriptions :column="2" border v-if="currentRow">
        <el-descriptions-item label="素材ID">{{ currentRow.material_id }}</el-descriptions-item>
        <el-descriptions-item label="账号ID">{{ currentRow.account_id }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ getMaterialTypeName(currentRow.material_type) }}</el-descriptions-item>
        <el-descriptions-item label="尺寸">{{ currentRow.width }} × {{ currentRow.height }}</el-descriptions-item>
        <el-descriptions-item label="文件大小">{{ formatFileSize(currentRow.file_size) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentRow.status === 1 ? 'success' : 'info'" size="small">
            {{ currentRow.status === 1 ? '有效' : '无效' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="预览" :span="2">
          <el-image
            v-if="currentRow.material_url && currentRow.material_type === 1"
            :src="currentRow.material_url"
            :preview-src-list="[currentRow.material_url]"
            fit="contain"
            style="max-width: 300px; max-height: 200px"
          />
          <video
            v-else-if="currentRow.material_url && currentRow.material_type === 2"
            :src="currentRow.material_url"
            controls
            style="max-width: 300px; max-height: 200px"
          />
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ formatDate(currentRow.created_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useCrawlerStore } from '@/stores'
import { getMaterials, getMaterialStats } from '@/api'
import type { AdMaterial } from '@/types'

const crawlerStore = useCrawlerStore()
const loading = ref(false)
const filterType = ref<number | undefined>()
const tableData = ref<AdMaterial[]>([])
const detailVisible = ref(false)
const currentRow = ref<AdMaterial | null>(null)

const stats = reactive({
  total: 0,
  image: 0,
  video: 0,
  text: 0,
  card: 0,
  mini_app: 0
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getMaterials({
      account_id: crawlerStore.currentAccount,
      material_type: filterType.value?.toString(),
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total
  } catch (error) {
    console.error('获取广告素材失败', error)
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const res = await getMaterialStats(crawlerStore.currentAccount)
    Object.assign(stats, res.data)
  } catch (error) {
    console.error('获取素材统计失败', error)
  }
}

const viewDetail = (row: AdMaterial) => {
  currentRow.value = row
  detailVisible.value = true
}

const getMaterialTypeName = (type: number): string => {
  const types: Record<number, string> = {
    1: '图片',
    2: '视频',
    3: '文本',
    4: '卡片',
    5: '小程序'
  }
  return types[type] || '未知'
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '-'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchData()
  fetchStats()
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

.filter-area {
  display: flex;
  gap: 12px;
}

.stats-row {
  margin-bottom: 20px;
}

.mini-stat {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}

.mini-stat .value {
  font-size: 24px;
  font-weight: 600;
  color: #409eff;
}

.mini-stat .label {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
