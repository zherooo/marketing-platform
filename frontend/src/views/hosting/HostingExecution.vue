<template>
  <div class="hosting-execution-page">
    <el-card class="filter-card">
      <el-row :gutter="16" align="middle">
        <el-col :span="4">
          <el-input v-model="filterAccountId" placeholder="账号ID" clearable />
        </el-col>
        <el-col :span="4">
          <el-select v-model="filterStatus" placeholder="执行状态" clearable>
            <el-option label="成功" :value="1" />
            <el-option label="失败" :value="2" />
            <el-option label="待执行" :value="3" />
            <el-option label="已回滚" :value="4" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="fetchList">查询</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-card style="margin-top: 16px">
      <el-table :data="executions" stripe v-loading="loading" style="width: 100%">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-content">
              <el-descriptions :column="2" border size="small">
                <el-descriptions-item label="触发快照">
                  <pre class="json-preview">{{ formatJson(row.trigger_snapshot) }}</pre>
                </el-descriptions-item>
                <el-descriptions-item label="执行参数">
                  <pre class="json-preview">{{ formatJson(row.action_params) }}</pre>
                </el-descriptions-item>
                <el-descriptions-item label="执行前值">
                  <pre class="json-preview">{{ formatJson(row.before_value) }}</pre>
                </el-descriptions-item>
                <el-descriptions-item label="执行后值">
                  <pre class="json-preview">{{ formatJson(row.after_value) }}</pre>
                </el-descriptions-item>
                <el-descriptions-item label="API响应" :span="2">
                  <pre class="json-preview">{{ formatJson(row.api_raw_resp) }}</pre>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="rule_name" label="规则名称" min-width="140" />
        <el-table-column prop="account_id" label="账号ID" width="160" />
        <el-table-column prop="target_id" label="目标ID" width="160" />
        <el-table-column prop="target_type" label="目标类型" width="90" />
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
          <template #default="{ row }">{{ formatTime(row.executed_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-popconfirm
              v-if="row.status === 1 && row.before_value"
              title="确定回滚此操作？"
              @confirm="handleRollback(row)"
            >
              <template #reference>
                <el-button size="small" link type="warning">回滚</el-button>
              </template>
            </el-popconfirm>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getHostingExecutions, rollbackExecution } from '@/api'
import type { HostingExecution } from '@/types'

const executions = ref<HostingExecution[]>([])
const loading = ref(false)
const filterAccountId = ref('')
const filterStatus = ref<number | undefined>()
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const actionLabels: Record<string, string> = {
  pause_ad: '暂停广告', adjust_bid: '调整出价', raise_budget: '提升预算',
  notify: '发送通知', resume_ad: '启用广告', quick_start: '一键起量'
}
const statusLabels: Record<number, string> = { 1: '成功', 2: '失败', 3: '待执行', 4: '已回滚' }
const getActionLabel = (t: string) => actionLabels[t] || t
const getStatusLabel = (s: number) => statusLabels[s] || '未知'
const getStatusType = (s: number) => s === 1 ? 'success' : s === 2 ? 'danger' : s === 4 ? 'info' : 'warning'
const formatTime = (t?: string) => t ? new Date(t).toLocaleString('zh-CN') : '-'

const formatJson = (str: string) => {
  if (!str) return '-'
  try { return JSON.stringify(JSON.parse(str), null, 2) } catch { return str }
}

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getHostingExecutions({
      account_id: filterAccountId.value,
      status: filterStatus.value,
      page: page.value,
      page_size: pageSize.value
    })
    executions.value = res.data.list || []
    total.value = res.data.total
  } catch {
    ElMessage.error('获取执行记录失败')
  } finally {
    loading.value = false
  }
}

const handleRollback = async (row: HostingExecution) => {
  try {
    await rollbackExecution(row.id)
    ElMessage.success('回滚成功')
    fetchList()
  } catch {
    ElMessage.error('回滚失败')
  }
}

onMounted(fetchList)
</script>

<style scoped>
.filter-card { margin-bottom: 0 }
.expand-content { padding: 8px 16px; max-height: 400px; overflow-y: auto }
.json-preview { font-size: 12px; white-space: pre-wrap; word-break: break-all; max-height: 200px; overflow-y: auto; margin: 0 }
</style>
