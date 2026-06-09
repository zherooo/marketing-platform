<template>
  <div class="oauth-container">
    <div class="oauth-box">
      <div class="oauth-header">
        <h2>账号授权</h2>
        <p>授权腾讯广告账号以获取广告数据</p>
      </div>
      
      <div class="oauth-content">
        <el-button type="primary" size="large" :loading="loading" @click="handleAuthorize">
          <el-icon v-if="!loading"><Connection /></el-icon>
          授权新账号
        </el-button>
      </div>
      
      <div class="account-list">
        <h3>已授权账号</h3>
        <el-table :data="accounts" v-loading="tableLoading">
          <el-table-column prop="account_name" label="账号名称" />
          <el-table-column prop="account_id" label="账号ID" width="180" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_online ? 'success' : 'info'" size="small">
                {{ row.is_online ? '在线' : '离线' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="最后抓取" width="180">
            <template #default="{ row }">
              {{ row.last_crawl_time ? formatDate(row.last_crawl_time) : '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button type="danger" size="small" text @click="handleDelete(row)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTokenList, deleteToken, getOAuthUrl } from '@/api'
import type { OAuthToken } from '@/types'

const loading = ref(false)
const tableLoading = ref(false)
const accounts = ref<OAuthToken[]>([])

const fetchAccounts = async () => {
  tableLoading.value = true
  try {
    const res = await getTokenList()
    accounts.value = res.data || []
  } catch (error) {
    console.error('获取账号列表失败', error)
  } finally {
    tableLoading.value = false
  }
}

const handleAuthorize = async () => {
  loading.value = true
  try {
    const res = await getOAuthUrl()
    window.open(res.data.url, '_blank', 'width=600,height=700')
    ElMessage.success('请在新窗口中完成授权')
  } catch (error) {
    console.error('获取授权URL失败', error)
  } finally {
    loading.value = false
  }
}

const handleDelete = async (row: OAuthToken) => {
  try {
    await ElMessageBox.confirm(`确定要删除账号 "${row.account_name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteToken(row.account_id)
    ElMessage.success('删除成功')
    fetchAccounts()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchAccounts()
})
</script>

<style scoped>
.oauth-container {
  padding: 20px;
}

.oauth-box {
  max-width: 900px;
  margin: 0 auto;
}

.oauth-header {
  text-align: center;
  margin-bottom: 30px;
}

.oauth-header h2 {
  font-size: 24px;
  color: #303133;
  margin-bottom: 8px;
}

.oauth-header p {
  color: #909399;
}

.oauth-content {
  display: flex;
  justify-content: center;
  margin-bottom: 40px;
}

.account-list h3 {
  font-size: 16px;
  color: #303133;
  margin-bottom: 16px;
}
</style>
