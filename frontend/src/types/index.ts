// 用户类型
export interface User {
  id: number
  username: string
  email: string
  created_at: string
}

// OAuth Token 类型
export interface OAuthToken {
  id: number
  account_id: string
  account_name: string
  authorized: boolean
  is_online: boolean
  last_crawl_time?: string
  created_at: string
}

// OAuth Account 类型
export interface OAuthAccount {
  id: number
  account_id: string
  account_name: string
  authorized: boolean
  access_token: string
  refresh_token: string
  token_expires_at: string
  is_online: boolean
  last_crawl_time?: string
  created_at: string
}

// 日报表数据类型
export interface DailyReport {
  id: number
  account_id: string
  date: string
  view_count: number
  click_count: number
  spend: number
  ctr: number
  cpc: number
  cpm: number
  conversion_count: number
  cost_per_conversion: number
  created_at: string
}

// 小时报表数据类型
export interface HourlyReport {
  id: number
  account_id: string
  date: string
  hour: number
  view_count: number
  click_count: number
  spend: number
  ctr: number
  cpc: number
  created_at: string
}

// 定向报表类型
export interface TargetReport {
  id: number
  account_id: string
  date: string
  target_id: string
  target_name: string
  view_count: number
  click_count: number
  spend: number
  ctr: number
  conversion_count: number
  created_at: string
}

// 广告系列类型
export interface Campaign {
  id: number
  campaign_id: string
  account_id: string
  campaign_name: string
  campaign_type: number
  daily_budget: number
  status: number
  created_at: string
}

// 广告组类型
export interface AdGroup {
  id: number
  adgroup_id: string
  campaign_id: string
  account_id: string
  adgroup_name: string
  bid_amount: number
  status: number
  created_at: string
}

// 广告类型
export interface Ad {
  id: number
  ad_id: string
  adgroup_id: string
  account_id: string
  ad_name: string
  ad_status: number
  created_at: string
}

// 广告创意类型
export interface AdCreative {
  id: number
  creative_id: string
  ad_id: string
  account_id: string
  creative_name: string
  creative_type: number
  created_at: string
}

// 广告素材类型
export interface AdMaterial {
  id: number
  material_id: string
  account_id: string
  material_type: number
  material_url: string
  width: number
  height: number
  file_size: number
  status: number
  created_at: string
}

// 抓取任务类型
export interface CrawlTask {
  id: number
  task_id: string
  account_id: string
  task_type: string
  status: number
  progress: number
  total: number
  success_count: number
  fail_count: number
  error_message?: string
  started_at?: string
  completed_at?: string
  created_at: string
}

// 抓取统计类型
export interface CrawlStatistics {
  total_tasks: number
  running_tasks: number
  success_tasks: number
  fail_tasks: number
  pending_tasks: number
  total_crawl_count: number
  today_crawl_count: number
}

// 分页响应类型
export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// API 响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 登录请求类型
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应类型
export interface LoginResponse {
  token: string
  user: User
}

// 手动抓取请求类型
export interface CrawlRequest {
  task_type: string
  account_ids?: string[]
  date_range?: string
}

// ==================== 智能托管相关 ====================

// 触发条件
export interface TriggerCondition {
  type: string
  metric: string
  operator: string
  threshold: number
  duration: number
  time_range?: string
  ad_min_days?: number
  max_conv_count?: number
  status_field?: string
  status_value?: string
}

// 执行动作
export interface ExecutionAction {
  type: string
  target_id?: string
  target_type?: string
  bid_adjust_ratio?: number
  budget_raise_ratio?: number
  budget_raise_amount?: number
  notify_channels?: string[]
  notify_users?: string[]
  message?: string
}

// 托管规则
export interface HostingRule {
  id: number
  rule_name: string
  category: string
  scene: string
  description: string
  status: number
  priority: number
  trigger_condition: TriggerCondition
  execution_action: ExecutionAction
  account_ids: string
  campaign_ids: string
  ad_group_ids: string
  ad_ids: string
  cooldown_minutes: number
  max_executions_per_day: number
  today_exec_count: number
  total_exec_count: number
  created_at: string
  updated_at: string
}

// 托管执行记录
export interface HostingExecution {
  id: number
  rule_id: number
  rule_name: string
  account_id: string
  target_id: string
  target_type: string
  action_type: string
  trigger_snapshot: string
  action_params: string
  before_value: string
  after_value: string
  status: number
  error_msg: string
  api_raw_resp: string
  rollback_at?: string
  executed_at?: string
  created_at: string
  updated_at: string
}

// 托管告警
export interface HostingAlert {
  id: number
  rule_id: number
  execution_id: number
  account_id: string
  alert_type: string
  alert_title: string
  alert_content: string
  severity: number
  status: number
  notify_channel: string
  notify_user: string
  handler: string
  handle_result: string
  handled_at?: string
  read_at?: string
  created_at: string
  updated_at: string
}

// 触发结果
export interface TriggerResult {
  account_id: string
  target_id: string
  target_type: string
  reason: string
  matched: boolean
  snapshot?: any
  rule?: HostingRule
}

// 看板统计
export interface HostingDashboardStats {
  active_rules: number
  total_rules: number
  today_exec: number
  total_exec: number
  success_exec: number
  failed_exec: number
  unread_alerts: number
  total_alerts: number
}

// 趋势数据
export interface TrendData {
  date: string
  count: number
}

// 看板数据
export interface HostingDashboard {
  stats: HostingDashboardStats
  trend: TrendData[]
}
