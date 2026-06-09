<template>
  <div class="hosting-rule-page">
    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-row :gutter="16" align="middle">
        <el-col :span="4">
          <el-select v-model="filterCategory" placeholder="规则分类" clearable @change="fetchList">
            <el-option label="成本控制" value="cost_control" />
            <el-option label="预算管理" value="budget_manage" />
            <el-option label="效果优化" value="effect_optimize" />
            <el-option label="风险预警" value="risk_alert" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filterStatus" placeholder="状态" clearable @change="fetchList">
            <el-option label="启用" value="1" />
            <el-option label="禁用" value="0" />
          </el-select>
        </el-col>
        <el-col :span="4" :offset="6" style="text-align: right">
          <el-button type="primary" @click="$router.push('/hosting/rules/create')">
            <el-icon><Plus /></el-icon> 创建规则
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 规则列表 -->
    <el-card style="margin-top: 16px">
      <el-table :data="rules" stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="rule_name" label="规则名称" min-width="160" />
        <el-table-column prop="category" label="分类" width="100">
          <template #default="{ row }">
            <el-tag :type="categoryTag(row.category)">{{ categoryLabel(row.category) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="scene" label="场景" min-width="140" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val: number) => handleToggle(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" />
        <el-table-column prop="today_exec_count" label="今日执行" width="80" />
        <el-table-column prop="total_exec_count" label="累计执行" width="80" />
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link @click="handleTest(row)">测试</el-button>
            <el-button size="small" link type="primary" @click="$router.push(`/hosting/rules/${row.id}/edit`)">编辑</el-button>
            <el-popconfirm title="确定删除此规则？" @confirm="handleDelete(row)">
              <template #reference>
                <el-button size="small" link type="danger">删除</el-button>
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

    <!-- 测试结果对话框 -->
    <el-dialog v-model="testDialogVisible" title="规则测试结果" width="700px">
      <div v-if="testLoading" style="text-align: center; padding: 40px">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <p>正在评估规则...</p>
      </div>
      <div v-else-if="testResults.length === 0" style="text-align: center; padding: 40px; color: #909399">
        没有匹配的触发条件
      </div>
      <div v-else>
        <p><strong>匹配结果：</strong>{{ testResults.length }} 条</p>
        <el-table :data="testResults" stripe>
          <el-table-column prop="account_id" label="账号ID" width="160" />
          <el-table-column prop="target_id" label="目标ID" width="160" />
          <el-table-column prop="target_type" label="目标类型" width="100" />
          <el-table-column prop="reason" label="触发原因" min-width="200" />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getHostingRules, toggleHostingRule, deleteHostingRule, testHostingRule } from '@/api'
import type { HostingRule, TriggerResult } from '@/types'

const rules = ref<HostingRule[]>([])
const loading = ref(false)
const filterCategory = ref('')
const filterStatus = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const testDialogVisible = ref(false)
const testLoading = ref(false)
const testResults = ref<TriggerResult[]>([])

const categoryLabels: Record<string, string> = {
  cost_control: '成本控制', budget_manage: '预算管理',
  effect_optimize: '效果优化', risk_alert: '风险预警'
}
const categoryLabel = (c: string) => categoryLabels[c] || c
const categoryTag = (c: string) => {
  const m: Record<string, string> = { cost_control: 'danger', budget_manage: 'warning', effect_optimize: 'success', risk_alert: 'info' }
  return m[c] || ''
}
const formatTime = (t: string) => new Date(t).toLocaleString('zh-CN')

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getHostingRules({
      category: filterCategory.value,
      status: filterStatus.value,
      page: page.value,
      page_size: pageSize.value
    })
    rules.value = res.data.list || []
    total.value = res.data.total
  } catch {
    ElMessage.error('获取规则列表失败')
  } finally {
    loading.value = false
  }
}

const handleToggle = async (row: HostingRule, val: number) => {
  try {
    await toggleHostingRule(row.id)
    ElMessage.success(val === 1 ? '规则已启用' : '规则已禁用')
  } catch {
    row.status = val === 1 ? 0 : 1
    ElMessage.error('操作失败')
  }
}

const handleDelete = async (row: HostingRule) => {
  try {
    await deleteHostingRule(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch {
    ElMessage.error('删除失败')
  }
}

const handleTest = async (row: HostingRule) => {
  testDialogVisible.value = true
  testLoading.value = true
  testResults.value = []
  try {
    const res = await testHostingRule(row.id)
    testResults.value = res.data.results || []
  } catch {
    ElMessage.error('测试失败')
  } finally {
    testLoading.value = false
  }
}

onMounted(fetchList)
</script>

<style scoped>
.filter-card { margin-bottom: 0 }
</style>
