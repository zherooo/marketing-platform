<template>
  <div class="hosting-alert-page">
    <el-card class="filter-card">
      <el-row :gutter="16" align="middle">
        <el-col :span="3">
          <el-input v-model="filterAccountId" placeholder="账号ID" clearable />
        </el-col>
        <el-col :span="3">
          <el-select v-model="filterAlertType" placeholder="告警类型" clearable>
            <el-option label="广告拒审" value="ad_rejected" />
            <el-option label="广告异常" value="ad_abnormal" />
            <el-option label="赔付变动" value="compensation" />
            <el-option label="成本飙升" value="cost_spike" />
            <el-option label="预算触顶" value="budget_limit" />
          </el-select>
        </el-col>
        <el-col :span="3">
          <el-select v-model="filterSeverity" placeholder="严重级别" clearable>
            <el-option label="低" :value="1" />
            <el-option label="中" :value="2" />
            <el-option label="高" :value="3" />
            <el-option label="紧急" :value="4" />
          </el-select>
        </el-col>
        <el-col :span="3">
          <el-select v-model="filterStatus" placeholder="状态" clearable>
            <el-option label="未读" :value="1" />
            <el-option label="已读" :value="2" />
            <el-option label="已处理" :value="3" />
            <el-option label="已忽略" :value="4" />
          </el-select>
        </el-col>
        <el-col :span="3">
          <el-button type="primary" @click="fetchList">查询</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-card style="margin-top: 16px">
      <el-table :data="alerts" stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="alert_title" label="告警标题" min-width="180" />
        <el-table-column prop="alert_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ getAlertTypeLabel(row.alert_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="account_id" label="账号ID" width="160" />
        <el-table-column prop="severity" label="级别" width="70">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)" size="small">{{ getSeverityLabel(row.severity) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <span :class="row.status === 1 ? 'unread-dot' : ''">
              {{ getAlertStatusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="alert_content" label="内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="created_at" label="发生时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 1" size="small" link type="primary" @click="handleMarkRead(row)">标记已读</el-button>
            <el-button v-if="row.status <= 2" size="small" link type="warning" @click="handleIgnore(row)">忽略</el-button>
            <el-button v-if="row.status <= 2" size="small" link type="success" @click="openHandleDialog(row)">处理</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @change="fetchList"
        style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>

    <!-- 处理弹窗 -->
    <el-dialog v-model="handleDialogVisible" title="处理告警" width="500px">
      <el-form :model="handleForm" label-width="80px">
        <el-form-item label="处理人">
          <el-input v-model="handleForm.handler" placeholder="请输入处理人" />
        </el-form-item>
        <el-form-item label="处理结果">
          <el-input v-model="handleForm.result" type="textarea" :rows="3" placeholder="请输入处理结果" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="handleLoading" @click="submitHandle">确认处理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getHostingAlerts, markAlertRead, handleAlert, ignoreAlert } from '@/api'
import type { HostingAlert } from '@/types'

const alerts = ref<HostingAlert[]>([])
const loading = ref(false)
const filterAccountId = ref('')
const filterAlertType = ref('')
const filterSeverity = ref<number | undefined>()
const filterStatus = ref<number | undefined>()
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const handleDialogVisible = ref(false)
const handleLoading = ref(false)
const currentAlertId = ref<number>(0)
const handleForm = ref({ handler: '', result: '' })

const alertTypeLabels: Record<string, string> = {
  ad_rejected: '广告拒审', ad_abnormal: '广告异常', compensation: '赔付变动',
  cost_spike: '成本飙升', budget_limit: '预算触顶'
}
const severityLabels: Record<number, string> = { 1: '低', 2: '中', 3: '高', 4: '紧急' }
const alertStatusLabels: Record<number, string> = { 1: '未读', 2: '已读', 3: '已处理', 4: '已忽略' }

const getAlertTypeLabel = (t: string) => alertTypeLabels[t] || t
const getSeverityLabel = (s: number) => severityLabels[s] || '未知'
const getSeverityType = (s: number) => s >= 3 ? 'danger' : s === 2 ? 'warning' : 'info'
const getAlertStatusLabel = (s: number) => alertStatusLabels[s] || '未知'
const formatTime = (t: string) => new Date(t).toLocaleString('zh-CN')

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getHostingAlerts({
      account_id: filterAccountId.value,
      alert_type: filterAlertType.value,
      severity: filterSeverity.value,
      status: filterStatus.value,
      page: page.value,
      page_size: pageSize.value
    })
    alerts.value = res.data.list || []
    total.value = res.data.total
  } catch {
    ElMessage.error('获取告警列表失败')
  } finally {
    loading.value = false
  }
}

const handleMarkRead = async (row: HostingAlert) => {
  try {
    await markAlertRead(row.id)
    ElMessage.success('已标记为已读')
    fetchList()
  } catch { ElMessage.error('操作失败') }
}

const handleIgnore = async (row: HostingAlert) => {
  try {
    await ignoreAlert(row.id)
    ElMessage.success('已忽略')
    fetchList()
  } catch { ElMessage.error('操作失败') }
}

const openHandleDialog = (row: HostingAlert) => {
  currentAlertId.value = row.id
  handleForm.value = { handler: '', result: '' }
  handleDialogVisible.value = true
}

const submitHandle = async () => {
  handleLoading.value = true
  try {
    await handleAlert(currentAlertId.value, handleForm.value.handler, handleForm.value.result)
    ElMessage.success('处理成功')
    handleDialogVisible.value = false
    fetchList()
  } catch { ElMessage.error('处理失败') }
  finally { handleLoading.value = false }
}

onMounted(fetchList)
</script>

<style scoped>
.filter-card { margin-bottom: 0 }
.unread-dot { color: #e6a23c; font-weight: bold }
.unread-dot::before { content: '● '; font-size: 8px }
</style>
