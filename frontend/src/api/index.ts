import { request } from './axios'
import type {
  LoginRequest,
  LoginResponse,
  OAuthToken,
  OAuthAccount,
  DailyReport,
  HourlyReport,
  TargetReport,
  Campaign,
  AdGroup,
  Ad,
  AdCreative,
  AdMaterial,
  CrawlTask,
  CrawlStatistics,
  CrawlRequest,
  PageResponse,
  User,
  HostingRule,
  HostingExecution,
  HostingAlert,
  HostingDashboard,
  TriggerResult
} from '@/types'

// ==================== 认证相关 ====================

// 用户登录
export const login = (data: LoginRequest) => {
  return request.post<LoginResponse>('/auth/login', data)
}

// 用户注册
export const register = (data: LoginRequest & { email: string }) => {
  return request.post<User>('/auth/register', data)
}

// 获取当前用户信息
export const getCurrentUser = () => {
  return request.get<User>('/user/info')
}

// 修改密码
export const changePassword = (oldPassword: string, newPassword: string) => {
  return request.post('/user/password', { old_password: oldPassword, new_password: newPassword })
}

// ==================== OAuth 相关 ====================

// 获取授权 URL
export const getOAuthUrl = () => {
  return request.get<{ url: string }>('/oauth/url')
}

// 获取 Token 列表
export const getTokenList = () => {
  return request.get<OAuthToken[]>('/oauth/tokens')
}

// 删除 Token
export const deleteToken = (accountId: string) => {
  return request.delete(`/oauth/tokens/${accountId}`)
}

// ==================== 报表相关 ====================

// 获取日报表
export const getDailyReports = (params: {
  account_id: string
  start_date: string
  end_date: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<DailyReport>>('/reports/daily', { params })
}

// 获取小时报表
export const getHourlyReports = (params: {
  account_id: string
  date: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<HourlyReport>>('/reports/hourly', { params })
}

// 获取定向报表
export const getTargetReports = (params: {
  account_id: string
  start_date: string
  end_date: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<TargetReport>>('/reports/target', { params })
}

// 获取趋势数据
export const getDailyTrend = (params: {
  account_id: string
  start_date: string
  end_date: string
}) => {
  return request.get<{ list: DailyReport[] }>('/reports/trend', { params })
}

// ==================== 广告系列相关 ====================

// 获取广告系列列表
export const getCampaigns = (params?: {
  account_id?: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<Campaign>>('/campaigns', { params })
}

// 获取广告系列详情
export const getCampaign = (id: string) => {
  return request.get<Campaign>(`/campaigns/${id}`)
}

// ==================== 广告组相关 ====================

// 获取广告组列表
export const getAdGroups = (params?: {
  account_id?: string
  campaign_id?: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<AdGroup>>('/adgroups', { params })
}

// 获取广告组详情
export const getAdGroup = (id: string) => {
  return request.get<AdGroup>(`/adgroups/${id}`)
}

// ==================== 广告相关 ====================

// 获取广告列表
export const getAds = (params?: {
  account_id?: string
  adgroup_id?: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<Ad>>('/ads', { params })
}

// 获取广告详情
export const getAd = (id: string) => {
  return request.get<Ad>(`/ads/${id}`)
}

// ==================== 广告创意相关 ====================

// 获取广告创意列表
export const getCreatives = (params?: {
  account_id?: string
  ad_id?: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<AdCreative>>('/creatives', { params })
}

// 获取广告创意详情
export const getCreative = (id: string) => {
  return request.get<AdCreative>(`/creatives/${id}`)
}

// ==================== 广告素材相关 ====================

// 获取广告素材列表
export const getMaterials = (params?: {
  account_id?: string
  material_type?: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<AdMaterial>>('/materials', { params })
}

// 获取广告素材详情
export const getMaterial = (id: string) => {
  return request.get<AdMaterial>(`/materials/${id}`)
}

// 获取素材统计
export const getMaterialStats = (accountId: string) => {
  return request.get<{
    total: number
    image: number
    video: number
    text: number
    card: number
    mini_app: number
  }>('/materials/stats', { params: { account_id: accountId } })
}

// ==================== 数据抓取相关 ====================

// 手动启动抓取
export const startCrawl = (data: CrawlRequest) => {
  return request.post('/crawler/start', data)
}

// 获取抓取统计
export const getCrawlStatistics = () => {
  return request.get<CrawlStatistics>('/crawler/statistics')
}

// 获取任务列表
export const getCrawlTasks = (params?: {
  task_type?: string
  status?: number
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<CrawlTask>>('/crawler/tasks', { params })
}

// 获取正在运行的任务
export const getRunningTasks = () => {
  return request.get<CrawlTask[]>('/crawler/tasks/running')
}

// 取消任务
export const cancelTask = (taskId: string) => {
  return request.post(`/crawler/tasks/${taskId}/cancel`)
}

// 重试任务
export const retryTask = (taskId: string) => {
  return request.post(`/crawler/tasks/${taskId}/retry`)
}

// 触发小时报表抓取
export const triggerHourlyReport = () => {
  return request.post('/crawler/trigger/hourly')
}

// 触发动报表抓取
export const triggerDailyReport = () => {
  return request.post('/crawler/trigger/daily')
}

// 触发广告结构抓取
export const triggerAllStruct = () => {
  return request.post('/crawler/trigger/struct')
}

// 触发广告系列抓取
export const triggerCampaign = () => {
  return request.post('/crawler/trigger/campaign')
}

// 触发广告系列级联抓取（系列 → 组 → 广告 → 创意 → 素材）
export const triggerCampaignCascade = (campaignId: string, accountId: string) => {
  return request.post(`/crawler/trigger/campaign/${campaignId}/cascade`, null, {
    params: { account_id: accountId }
  })
}

// 触发广告组级联抓取（组 → 广告 → 创意 → 素材）
export const triggerAdGroupCascade = (adgroupId: string, accountId: string) => {
  return request.post(`/crawler/trigger/adgroup/${adgroupId}/cascade`, null, {
    params: { account_id: accountId }
  })
}

// 触发广告级联抓取（广告 → 创意 → 素材）
export const triggerAdCascade = (adId: string, accountId: string) => {
  return request.post(`/crawler/trigger/ad/${adId}/cascade`, null, {
    params: { account_id: accountId }
  })
}

// ==================== 智能托管相关 ====================

// 规则管理
export const getHostingRules = (params?: {
  category?: string
  status?: string
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<HostingRule>>('/hosting/rules', { params })
}

export const getHostingRule = (id: number) => {
  return request.get<HostingRule>(`/hosting/rules/${id}`)
}

export const createHostingRule = (data: Partial<HostingRule>) => {
  return request.post<HostingRule>('/hosting/rules', data)
}

export const updateHostingRule = (id: number, data: Partial<HostingRule>) => {
  return request.put<HostingRule>(`/hosting/rules/${id}`, data)
}

export const deleteHostingRule = (id: number) => {
  return request.delete(`/hosting/rules/${id}`)
}

export const toggleHostingRule = (id: number) => {
  return request.post<HostingRule>(`/hosting/rules/${id}/toggle`)
}

export const testHostingRule = (id: number) => {
  return request.post<{ matched_count: number; results: TriggerResult[] }>(`/hosting/rules/${id}/test`)
}

// 执行记录
export const getHostingExecutions = (params?: {
  account_id?: string
  status?: number
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<HostingExecution>>('/hosting/executions', { params })
}

export const getHostingExecution = (id: number) => {
  return request.get<HostingExecution>(`/hosting/executions/${id}`)
}

export const rollbackExecution = (id: number) => {
  return request.post(`/hosting/executions/${id}/rollback`)
}

// 告警管理
export const getHostingAlerts = (params?: {
  account_id?: string
  alert_type?: string
  severity?: number
  status?: number
  page?: number
  page_size?: number
}) => {
  return request.get<PageResponse<HostingAlert>>('/hosting/alerts', { params })
}

export const getHostingAlert = (id: number) => {
  return request.get<HostingAlert>(`/hosting/alerts/${id}`)
}

export const markAlertRead = (id: number) => {
  return request.post(`/hosting/alerts/${id}/read`)
}

export const handleAlert = (id: number, handler: string, result: string) => {
  return request.post(`/hosting/alerts/${id}/handle`, { handler, result })
}

export const ignoreAlert = (id: number) => {
  return request.post(`/hosting/alerts/${id}/ignore`)
}

// 看板 & 触发
export const getHostingDashboard = () => {
  return request.get<HostingDashboard>('/hosting/dashboard')
}

export const triggerEvaluate = () => {
  return request.post('/hosting/trigger/evaluate')
}

export const triggerCollectSnapshots = () => {
  return request.post('/hosting/trigger/collect')
}
