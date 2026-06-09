<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <img src="/vite.svg" alt="logo" v-if="!isCollapse" />
        <span v-if="!isCollapse">广告洞察</span>
        <el-icon v-else><DataLine /></el-icon>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="false"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataLine /></el-icon>
          <template #title>数据看板</template>
        </el-menu-item>
        
        <el-sub-menu index="reports">
          <template #title>
            <el-icon><TrendCharts /></el-icon>
            <span>报表数据</span>
          </template>
          <el-menu-item index="/reports/daily">日报表</el-menu-item>
          <el-menu-item index="/reports/hourly">小时报表</el-menu-item>
          <el-menu-item index="/reports/target">定向报表</el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="ads">
          <template #title>
            <el-icon><Goods /></el-icon>
            <span>广告结构</span>
          </template>
          <el-menu-item index="/campaigns">广告系列</el-menu-item>
          <el-menu-item index="/adgroups">广告组</el-menu-item>
          <el-menu-item index="/ads">广告</el-menu-item>
          <el-menu-item index="/materials">广告素材</el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="crawler">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>数据抓取</span>
          </template>
          <el-menu-item index="/crawler">抓取管理</el-menu-item>
          <el-menu-item index="/oauth">账号授权</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="hosting">
          <template #title>
            <el-icon><SetUp /></el-icon>
            <span>智能托管</span>
          </template>
          <el-menu-item index="/hosting/dashboard">托管看板</el-menu-item>
          <el-menu-item index="/hosting/rules">托管规则</el-menu-item>
          <el-menu-item index="/hosting/executions">执行记录</el-menu-item>
          <el-menu-item index="/hosting/alerts">告警通知</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>
    
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.meta.title !== '数据看板'">
              {{ route.meta.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        
        <div class="header-right">
          <!-- 账号选择 -->
          <el-select
            v-model="currentAccountId"
            placeholder="选择账号"
            clearable
            class="account-select"
            @change="handleAccountChange"
          >
            <el-option
              v-for="account in accounts"
              :key="account.account_id"
              :label="account.account_name"
              :value="account.account_id"
            >
              <span>{{ account.account_name }}</span>
              <el-tag v-if="account.is_online" size="small" type="success" class="ml-2">在线</el-tag>
            </el-option>
          </el-select>
          
          <!-- 用户信息 -->
          <el-dropdown @command="handleUserCommand">
            <span class="user-info">
              <el-avatar :size="32" icon="UserFilled" />
              <span class="username">{{ user?.username }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      
      <!-- 主内容区 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore, useCrawlerStore } from '@/stores'
import { getTokenList } from '@/api'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const crawlerStore = useCrawlerStore()

const isCollapse = ref(false)
const accounts = ref<any[]>([])
const currentAccountId = ref('')

const user = computed(() => authStore.user)
const activeMenu = computed(() => route.path)

// 获取账号列表
const fetchAccounts = async () => {
  try {
    const res = await getTokenList()
    accounts.value = res.data || []
    crawlerStore.setAccounts(accounts.value)
    
    // 设置当前账号
    if (accounts.value.length > 0 && !currentAccountId.value) {
      const online = accounts.value.find(a => a.is_online)
      currentAccountId.value = online?.account_id || accounts.value[0].account_id
      crawlerStore.setCurrentAccount(currentAccountId.value)
    }
  } catch (error) {
    console.error('获取账号列表失败', error)
  }
}

// 账号切换
const handleAccountChange = (accountId: string) => {
  crawlerStore.setCurrentAccount(accountId)
  ElMessage.success('已切换账号')
}

// 用户菜单
const handleUserCommand = async (command: string) => {
  if (command === 'logout') {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    authStore.logout()
    router.push('/login')
    ElMessage.success('已退出登录')
  } else if (command === 'password') {
    ElMessage.info('修改密码功能开发中')
  }
}

onMounted(() => {
  authStore.fetchUser()
  fetchAccounts()
})

watch(() => crawlerStore.accounts, (newAccounts) => {
  accounts.value = newAccounts
}, { deep: true })
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background: #304156;
  transition: width 0.3s;
  overflow-x: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  background: #263445;
}

.logo img {
  width: 32px;
  height: 32px;
  margin-right: 8px;
}

.sidebar-menu {
  border-right: none;
  background: transparent;
}

/* 菜单文字与图标颜色 - 提升可读性 */
.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  color: #bfcbd9 !important;
}

.sidebar-menu :deep(.el-menu-item .el-icon),
.sidebar-menu :deep(.el-sub-menu__title .el-icon) {
  color: #bfcbd9 !important;
}

/* hover 状态 */
.sidebar-menu :deep(.el-menu-item:hover),
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background-color: #263445 !important;
  color: #fff !important;
}

.sidebar-menu :deep(.el-menu-item:hover .el-icon),
.sidebar-menu :deep(.el-sub-menu__title:hover .el-icon) {
  color: #fff !important;
}

/* 选中状态 */
.sidebar-menu :deep(.el-menu-item.is-active) {
  background-color: #1890ff !important;
  color: #fff !important;
}

.sidebar-menu :deep(.el-menu-item.is-active .el-icon) {
  color: #fff !important;
}

/* 子菜单展开项 */
.sidebar-menu :deep(.el-menu--inline) {
  background-color: #1f2d3d !important;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 220px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
}

.collapse-btn:hover {
  color: #409eff;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.account-select {
  width: 200px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.username {
  font-size: 14px;
  color: #606266;
}

.main-content {
  background: #f5f7fa;
  padding: 16px;
}

.ml-2 {
  margin-left: 8px;
}
</style>
