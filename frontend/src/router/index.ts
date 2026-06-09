import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录', requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { title: '注册', requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/components/Layout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '数据看板' }
      },
      {
        path: 'reports/daily',
        name: 'DailyReport',
        component: () => import('@/views/Reports/DailyReport.vue'),
        meta: { title: '日报表' }
      },
      {
        path: 'reports/hourly',
        name: 'HourlyReport',
        component: () => import('@/views/Reports/HourlyReport.vue'),
        meta: { title: '小时报表' }
      },
      {
        path: 'reports/target',
        name: 'TargetReport',
        component: () => import('@/views/Reports/TargetReport.vue'),
        meta: { title: '定向报表' }
      },
      {
        path: 'campaigns',
        name: 'Campaigns',
        component: () => import('@/views/Campaigns.vue'),
        meta: { title: '广告系列' }
      },
      {
        path: 'adgroups',
        name: 'AdGroups',
        component: () => import('@/views/AdGroups.vue'),
        meta: { title: '广告组' }
      },
      {
        path: 'ads',
        name: 'Ads',
        component: () => import('@/views/Ads.vue'),
        meta: { title: '广告' }
      },
      {
        path: 'materials',
        name: 'Materials',
        component: () => import('@/views/Materials.vue'),
        meta: { title: '广告素材' }
      },
      {
        path: 'crawler',
        name: 'Crawler',
        component: () => import('@/views/Crawler.vue'),
        meta: { title: '数据抓取管理' }
      },
      {
        path: 'oauth',
        name: 'OAuth',
        component: () => import('@/views/OAuth.vue'),
        meta: { title: '账号授权' }
      },
      {
        path: 'hosting/dashboard',
        name: 'HostingDashboard',
        component: () => import('@/views/hosting/HostingDashboard.vue'),
        meta: { title: '智能托管看板' }
      },
      {
        path: 'hosting/rules',
        name: 'HostingRules',
        component: () => import('@/views/hosting/HostingRule.vue'),
        meta: { title: '托管规则' }
      },
      {
        path: 'hosting/rules/create',
        name: 'HostingRuleCreate',
        component: () => import('@/views/hosting/HostingRuleCreate.vue'),
        meta: { title: '创建规则' }
      },
      {
        path: 'hosting/rules/:id/edit',
        name: 'HostingRuleEdit',
        component: () => import('@/views/hosting/HostingRuleCreate.vue'),
        meta: { title: '编辑规则' }
      },
      {
        path: 'hosting/executions',
        name: 'HostingExecutions',
        component: () => import('@/views/hosting/HostingExecution.vue'),
        meta: { title: '执行记录' }
      },
      {
        path: 'hosting/alerts',
        name: 'HostingAlerts',
        component: () => import('@/views/hosting/HostingAlert.vue'),
        meta: { title: '告警通知' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  document.title = `${to.meta.title || '广告数据洞察平台'} - 广告数据洞察平台`
  
  const authStore = useAuthStore()
  const requiresAuth = to.meta.requiresAuth !== false
  
  if (requiresAuth && !authStore.isLoggedIn()) {
    next('/login')
  } else if (to.path === '/login' && authStore.isLoggedIn()) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
