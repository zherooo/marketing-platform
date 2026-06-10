# 广告数据洞察平台

基于腾讯广告 API 的广告数据采集、分析与展示平台。支持多账号 OAuth 授权、定时数据抓取、多维度报表分析与可视化看板。

> 📸 项目截图请查看 [SCREENSHOTS.md](./SCREENSHOTS.md)

---

## 功能概览

### 核心功能

| 模块 | 功能 | 说明 |
|------|------|------|
| **用户认证** | 登录 / 注册 / 修改密码 | JWT Token + bcrypt 密码加密 |
| **OAuth 授权** | 腾讯广告账号授权 | 获取授权 URL → 回调 → Token 管理 |
| **数据看板** | 核心指标卡片 + ECharts 图表 | 趋势图、饼图、漏斗图 |
| **报表数据** | 日报 / 小时报 / 定向报表 | 分页查询、多维度筛选 |
| **广告结构** | 广告系列 → 广告组 → 广告 → 创意 → 素材 | 五层结构，支持详情查看 |
| **数据抓取** | 手动触发 + 定时调度 | 多协程并发抓取，任务管理与重试 |
| **定时任务** | Cron 表达式调度 | 每小时采集广告结构，每日生成报表 |
| **智能托管** | IF-THEN 规则引擎 + 自动执行 | 8 大托管场景，成本/预算/效果/风险全覆盖 |

---

## 技术架构

### 技术栈总览

| 层级 | 技术 | 版本 |
|------|------|------|
| **前端** | Vue 3 + TypeScript + Vite | Vue 3.4 / TS 5.4 / Vite 5 |
| **UI 框架** | Element Plus + ECharts | Element Plus 2.6 / ECharts 5.5 |
| **状态管理** | Pinia + Vue Router 4 | Pinia 2.1 |
| **后端** | Go + Gin | Go 1.21+ / Gin 1.11 |
| **数据库** | MySQL + GORM | MySQL 8.0+ / GORM 2.0 |
| **定时调度** | robfig/cron v3 | 秒级精度 |
| **认证** | JWT + bcrypt | golang-jwt v5 |
| **日志** | Zap + Lumberjack | 文件轮转 + 控制台双写 |
| **配置** | Viper | YAML 配置文件 |

---

## 后端架构

### 目录结构

```
backend/
├── main.go                          # 入口：11 步启动流程
├── config.yaml                      # 配置文件（YAML）
├── internal/
│   ├── config/                      # 配置管理（Viper）
│   │   └── config.go
│   ├── database/                    # 数据库初始化
│   │   └── database.go              # GORM 连接 + AutoMigrate（16 张表）
│   ├── model/                       # 数据模型（GORM）
│   │   ├── user.go                  # 用户表
│   │   ├── oauth.go                 # OAuth + 广告结构（9 张表）
│   │   ├── report.go               # 报表（日/小时/定向 + 报表任务）
│   │   └── crawl_task.go           # 抓取任务 + 任务日志
│   ├── handler/                     # HTTP 处理器（Controller 层）
│   │   ├── auth_handler.go          # 登录 / 注册 / 用户信息 / 改密
│   │   ├── report_handler.go        # 日报 / 小时报 / 定向报表 / 趋势
│   │   ├── campaign_handler.go      # 广告系列 / 广告组
│   │   ├── ad_handler.go            # 广告 / 创意
│   │   ├── material_handler.go      # 素材 + 统计
│   │   ├── oauth_handler.go         # OAuth 授权 URL / 回调 / Token 管理
│   │   └── crawler_handler.go       # 抓取触发 / 任务管理 / 统计
│   ├── service/                     # 业务逻辑层
│   │   ├── user_service.go          # 用户 CRUD
│   │   ├── report_service.go        # 报表查询 / 批量保存
│   │   ├── campaign_service.go      # 广告系列 + 广告组查询
│   │   ├── ad_service.go            # 广告 + 创意查询
│   │   ├── material_service.go      # 素材查询 + 统计
│   │   ├── crawl_service.go         # 抓取任务 CRUD + 清理
│   │   └── oauth_service.go         # Token 管理
│   ├── crawler/                     # 数据抓取引擎
│   │   ├── task_manager.go          # 任务管理器（核心调度）
│   │   ├── worker_pool.go           # 每账号独立协程池
│   │   ├── api_client.go            # 腾讯广告 API 客户端（Token/限流/重试）
│   │   ├── campaign_crawler.go      # 广告系列抓取器
│   │   ├── ad_crawler.go            # 广告 / 广告组 / 创意抓取器
│   │   ├── material_crawler.go      # 素材抓取器
│   │   └── hourly_report_crawler.go # 小时报表抓取器
│   ├── scheduler/                   # 定时任务调度
│   │   └── scheduler.go             # 10 个定时任务 + 手动触发
│   ├── middleware/                   # 中间件
│   │   ├── auth.go                  # JWT 认证 / CORS / Panic 恢复
│   │   └── response.go              # 统一响应格式
│   └── logger/                      # 日志
│       ├── logger.go                # Zap 初始化
│       └── request_logger.go        # 请求日志中间件
```

### 架构分层

```
 ┌──────────────────────────────────────────────────┐
 │                  HTTP 请求入口                     │
 │         Gin Router → Middleware 链                 │
 ├────────┬────────┬────────┬─────────┬──────────────┤
 │ Auth   │ Report │Campaign│Material │ Crawler      │  ← Handler 层
 │Handler │Handler │Handler │Handler  │ Handler       │
 ├────────┼────────┼────────┼─────────┼──────────────┤
 │ User   │ Report │Campaign│Material │ OAuth/Crawl  │  ← Service 层
 │Service │Service │Service │Service  │ Service       │
 ├────────┴────────┴────────┴─────────┴──────────────┤
 │                    GORM (Model 层)                  │
 ├────────────────────────────────────────────────────┤
 │              MySQL 数据库（16 张表）                  │
 └────────────────────────────────────────────────────┘

 ┌────────────────────────────────────────────────────┐
 │         数据抓取 & 调度子系统                         │
 ├──────────────────┬─────────────────────────────────┤
 │   Scheduler      │   Crawler.TaskManager           │
 │   (robfig/cron)  │   → 每账号 WorkerPool (channel)  │
 │   ↓ 定时触发      │   → 7 种 Crawler                │
 │   调用 TaskMgr   │   → API Client (限流/重试)        │
 └──────────────────┴─────────────────────────────────┘
```

### 数据表设计

| 表名 | 模型 | 说明 |
|------|------|------|
| `users` | User | 用户表（JWT 认证） |
| `oauth_tokens` | OAuthToken | OAuth 授权令牌 |
| `oauth_accounts` | OAuthAccount | OAuth 授权账号（含敏感信息） |
| `ad_accounts` | AdAccount | 广告主账号 |
| `campaigns` | Campaign | 推广计划（广告系列） |
| `ad_groups` | AdGroup | 广告组 |
| `ads` | Ad | 广告 |
| `ad_creatives` | AdCreative | 广告创意 |
| `ad_materials` | AdMaterial | 广告素材 |
| `daily_reports` | DailyReport | 日报表（曝光/点击/消耗/转化） |
| `hourly_reports` | HourlyReport | 小时报表 |
| `target_reports` | TargetReport | 定向标签报表 |
| `report_tasks` | ReportTask | 异步报表任务 |
| `crawl_tasks` | CrawlTask | 数据抓取任务 |
| `crawl_task_logs` | CrawlTaskLog | 抓取任务执行日志 |
| `hosting_rules` | HostingRule | 智能托管规则 |
| `hosting_executions` | HostingExecution | 托管执行记录 |
| `hosting_alerts` | HostingAlert | 托管告警通知 |
| `ad_performance_snapshots` | AdPerformanceSnapshot | 广告性能快照 |

### 抓取系统工作流程

```
定时调度 (Cron)                手动触发 (API)
      │                              │
      ▼                              ▼
   TaskManager.CreateTasks()
      │
      ▼ 按账号分发
 ┌────────────────┐  ┌────────────────┐
 │ 账号 A 协程池  │  │ 账号 B 协程池  │  ← 每账号独立 WorkerPool
 │ (5 workers)    │  │ (5 workers)    │
 │ ┌────┐ ┌────┐ │  │ ┌────┐ ┌────┐ │
 │ │W1  │ │W2  │ │  │ │W1  │ │W2  │ │
 │ └──┬─┘ └──┬─┘ │  │ └──┬─┘ └──┬─┘ │
 │    │      │   │  │    │      │   │
 │   ▼      ▼    │  │   ▼      ▼    │
 │  channel(100) │  │  channel(100) │
 └───────┬───────┘  └───────┬───────┘
         │                  │
         ▼                  ▼
    API Client (限流10/s, 重试3次, 指数退避)
         │
         ▼
   腾讯广告 API → 解析 → 存储 (GORM)
```

---

## 前端架构

### 目录结构

```
frontend/
├── src/
│   ├── api/                     # API 接口层
│   │   ├── axios.ts             # Axios 实例（拦截器/Token 注入）
│   │   └── index.ts             # 29 个 API 方法
│   ├── router/
│   │   └── index.ts             # 路由配置 + 导航守卫
│   ├── stores/                  # Pinia 状态管理
│   │   ├── auth.ts              # 用户认证状态
│   │   └── crawler.ts           # 抓取管理状态
│   ├── views/                   # 页面组件
│   │   ├── Login.vue            # 登录页
│   │   ├── Register.vue         # 注册页
│   │   ├── Dashboard.vue        # 数据看板（ECharts 图表）
│   │   ├── Campaigns.vue        # 广告系列列表
│   │   ├── AdGroups.vue         # 广告组列表
│   │   ├── Ads.vue              # 广告列表
│   │   ├── Materials.vue        # 素材管理 + 统计
│   │   ├── Crawler.vue          # 数据抓取管理
│   │   ├── OAuth.vue            # 账号授权
│   │   └── Reports/             # 报表页面
│   │       ├── DailyReport.vue  # 日报表
│   │       ├── HourlyReport.vue # 小时报表
│   │       └── TargetReport.vue # 定向报表
│   │   └── hosting/             # 智能托管页面
│   │       ├── HostingDashboard.vue  # 托管看板
│   │       ├── HostingRule.vue       # 规则列表
│   │       ├── HostingRuleCreate.vue # 创建/编辑规则
│   │       ├── HostingExecution.vue  # 执行记录
│   │       └── HostingAlert.vue      # 告警通知
│   ├── components/
│   │   └── Layout.vue           # 主布局（侧边栏 + 头部）
│   ├── types/
│   │   └── index.ts             # 15 个 TS 接口定义
│   ├── styles/
│   │   └── index.css            # 全局样式
│   ├── App.vue                  # 根组件
│   └── main.ts                  # 应用入口
├── vite.config.ts               # Vite 配置（代理 + 别名）
├── tsconfig.json
└── package.json
```

### 路由设计

```
/login                    → Login.vue          (公开)
/register                 → Register.vue       (公开)
/                         → Layout.vue          (需登录)
  ├── /dashboard          → Dashboard.vue       (数据看板)
  ├── /reports/daily      → DailyReport.vue     (日报表)
  ├── /reports/hourly     → HourlyReport.vue    (小时报表)
  ├── /reports/target     → TargetReport.vue    (定向报表)
  ├── /campaigns          → Campaigns.vue       (广告系列)
  ├── /adgroups           → AdGroups.vue        (广告组)
  ├── /ads                → Ads.vue             (广告)
  ├── /materials          → Materials.vue       (素材管理)
  ├── /crawler            → Crawler.vue         (抓取管理)
  ├── /oauth              → OAuth.vue           (账号授权)
  └── /hosting/           → hosting/             (智能托管)
      ├── /dashboard      → HostingDashboard    (托管看板)
      ├── /rules          → HostingRule         (规则列表)
      ├── /rules/create   → HostingRuleCreate   (创建规则)
      ├── /rules/:id/edit → HostingRuleCreate   (编辑规则)
      ├── /executions     → HostingExecution    (执行记录)
      └── /alerts         → HostingAlert        (告警通知)
```

### 组件状态流转

```
                 ┌─────────┐
       未登录 → │  Login   │ ← /login, /register
                 └────┬────┘
                      │ JWT Token
                      ▼
                 ┌─────────┐
       已登录 → │ Layout   │  ← 侧边栏导航 + 头部用户信息
                 └────┬────┘
          ┌───────────┼───────────┬──────────┬──────────┐
          ▼           ▼           ▼          ▼          ▼
      Dashboard   Reports     Campaigns  Materials  Crawler
      (ECharts)   (表格)       (表格)     (表格+统计) (任务管理)

      API 请求 → Axios 拦截器 → 注入 Token → 代理到后端 :8080
```

---

## API 接口总览

### 认证模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | 否 |
| POST | `/api/v1/auth/login` | 用户登录 | 否 |
| GET | `/api/v1/user/info` | 当前用户信息 | 是 |
| POST | `/api/v1/user/password` | 修改密码 | 是 |

### OAuth 模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/oauth/url` | 获取授权 URL | 否 |
| GET | `/api/v1/oauth/callback` | OAuth 回调 | 否 |
| GET | `/api/v1/oauth/tokens` | Token 列表 | 是 |
| DELETE | `/api/v1/oauth/tokens/:id` | 删除 Token | 是 |

### 报表模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/reports/daily` | 日报表 | 是 |
| GET | `/api/v1/reports/hourly` | 小时报表 | 是 |
| GET | `/api/v1/reports/target` | 定向报表 | 是 |
| GET | `/api/v1/reports/trend` | 趋势数据 | 是 |

### 广告结构模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/campaigns` | 广告系列列表 | 是 |
| GET | `/api/v1/campaigns/:id` | 广告系列详情 | 是 |
| GET | `/api/v1/adgroups` | 广告组列表 | 是 |
| GET | `/api/v1/adgroups/:id` | 广告组详情 | 是 |
| GET | `/api/v1/ads` | 广告列表 | 是 |
| GET | `/api/v1/ads/:id` | 广告详情 | 是 |
| GET | `/api/v1/creatives` | 创意列表 | 是 |
| GET | `/api/v1/creatives/:id` | 创意详情 | 是 |
| GET | `/api/v1/materials` | 素材列表 | 是 |
| GET | `/api/v1/materials/:id` | 素材详情 | 是 |
| GET | `/api/v1/materials/stats` | 素材统计 | 是 |

**级联抓取（前端页面操作）**：

各广告结构列表页的操作列均有「抓取」按钮，点击后触发级联抓取：

| 页面 | 按钮 | 级联链路 |
|------|------|----------|
| 广告系列 (`/campaigns`) | `抓取` | 系列 → 广告组 → 广告 → 创意 → 素材 |
| 广告组 (`/adgroups`) | `抓取` | 广告组 → 广告 → 创意 → 素材 |
| 广告 (`/ads`) | `抓取` | 广告 → 创意 → 素材 |

### 数据抓取模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/crawler/start` | 手动启动抓取 | 是 |
| GET | `/api/v1/crawler/statistics` | 抓取统计 | 是 |
| GET | `/api/v1/crawler/tasks` | 任务列表 | 是 |
| GET | `/api/v1/crawler/tasks/running` | 运行中任务 | 是 |
| POST | `/api/v1/crawler/tasks/:id/cancel` | 取消任务 | 是 |
| POST | `/api/v1/crawler/tasks/:id/retry` | 重试任务 | 是 |
| POST | `/api/v1/crawler/trigger/hourly` | 触发小时报抓取 | 是 |
| POST | `/api/v1/crawler/trigger/daily` | 触发日报抓取 | 是 |
| POST | `/api/v1/crawler/trigger/struct` | 触发广告结构抓取 | 是 |
| POST | `/api/v1/crawler/trigger/campaign` | 触发广告系列抓取 | 是 |
| POST | `/api/v1/crawler/trigger/campaign/:campaign_id/cascade` | 级联抓取广告系列（系列→组→广告→创意→素材） | 是 |
| POST | `/api/v1/crawler/trigger/adgroup/:adgroup_id/cascade` | 级联抓取广告组（组→广告→创意→素材） | 是 |
| POST | `/api/v1/crawler/trigger/ad/:ad_id/cascade` | 级联抓取广告（广告→创意→素材） | 是 |

---

## 智能托管系统

### 概述

智能托管系统基于 IF-THEN 规则引擎，自动监控广告投放数据，当触发条件满足时自动执行预设动作（暂停广告、调整出价、发送告警等），实现广告投放的智能化管理。

### 托管场景一览

系统内置以下 8 大托管场景，覆盖成本控制、预算管理、效果优化、风险预警四大类：

#### 📊 成本控制

| 场景 | 触发条件 (IF) | 执行动作 (THEN) |
|------|--------------|----------------|
| **转化成本超出预期** | 转化成本/CPC 超过目标成本的阈值 | 暂停广告 / 调低出价 |
| **深夜成本失控** | 凌晨 0-6 点成本飙升超过阈值 | 降低出价 / 暂停广告 / 定时启用 |

#### 💰 预算管理

| 场景 | 触发条件 (IF) | 执行动作 (THEN) |
|------|--------------|----------------|
| **日消耗触顶后提额** | 日消耗达到日限额 80%/100% | 提升日限额 |
| **消耗数据异常** | 学习期内消耗/曝光未达标 | 暂停广告 |

#### 🎯 效果优化

| 场景 | 触发条件 (IF) | 执行动作 (THEN) |
|------|--------------|----------------|
| **低效广告自动关停** | 投放 >3 天，转化 <5，成本超标 | 暂停广告 |
| **优质广告倾斜预算** | A/B 测试判定更优 | 分配更多流量 |
| **新广告起量期扶持** | 新建 24-48 小时内 | 一键起量加速 |

#### ⚠️ 风险预警

| 场景 | 触发条件 (IF) | 执行动作 (THEN) |
|------|--------------|----------------|
| **广告拒审或异常** | 诊断/拒审状态变化 | 发送通知 |
| **赔付状态变动** | 赔付状态变化 | 发送告警 |

### 系统架构

```
┌──────────────────────────────────────────────────────────────┐
│                    智能托管系统架构                             │
├──────────────────────────────────────────────────────────────┤
│  定时调度 (Cron)                                              │
│  ├── 每5分钟：采集广告性能快照                                  │
│  ├── 每5分钟：规则评估 + 自动执行                              │
│  ├── 每10分钟：广告状态健康检查                                │
│  └── 每日零点：重置计数器 + 清理旧记录                         │
├──────────────────────────────────────────────────────────────┤
│  规则引擎 (RuleEngine)                                        │
│  ├── 加载启用规则                                              │
│  ├── 获取规则作用范围内的广告性能快照                           │
│  ├── 评估触发条件（成本/预算/效果/风险）                        │
│  └── 输出匹配的 TriggerResult                                │
├──────────────────────────────────────────────────────────────┤
│  动作执行器 (ActionExecutor)                                   │
│  ├── pause_ad：暂停广告                                       │
│  ├── adjust_bid：调整出价（按比例）                           │
│  ├── raise_budget：提升日限额                                 │
│  ├── notify：发送通知（邮件/短信/钉钉/飞书）                   │
│  ├── resume_ad：恢复广告投放                                  │
│  ├── quick_start：一键起量                                    │
│  └── rollback：执行回滚                                       │
├──────────────────────────────────────────────────────────────┤
│  防护机制                                                      │
│  ├── 冷却时间 (CooldownMinutes)：两次执行最小间隔              │
│  └── 每日上限 (MaxExecutionsPerDay)：单日最大执行次数          │
└──────────────────────────────────────────────────────────────┘
```

### 核心数据表

| 表名 | 说明 | 核心字段 |
|------|------|---------|
| `hosting_rules` | 托管规则表 | 规则名称、分类、触发条件(JSON)、执行动作(JSON)、作用范围、冷却时间、执行上限 |
| `hosting_executions` | 执行记录表 | 规则ID、目标信息、动作类型、触发快照、执行前后值、执行状态、回滚支持 |
| `hosting_alerts` | 告警通知表 | 告警类型、严重级别、通知渠道、处理人、处理结果、阅读/处理状态 |
| `ad_performance_snapshots` | 广告性能快照表 | 曝光/点击/转化/消耗/CTR/CVR/CPC/CPA、预算使用率、诊断状态、A/B分组 |

### 规则配置

#### 触发条件 (TriggerCondition)

```json
{
  "type": "cost_control",
  "metric": "conversion_cost",
  "operator": "gt",
  "value": 50.0,
  "duration": 5,
  "time_range": "0-6",
  "ad_min_days": 3,
  "max_conv_count": 5
}
```

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `type` | 条件类型 | cost_control / budget_manage / effect_optimize / risk_alert |
| `metric` | 监控指标 | conversion_cost / cpc / spend / budget_ratio / impressions / conversions / delivery_hours |
| `operator` | 比较运算符 | gt(>) / gte(>=) / lt(<) / lte(<=) / eq(=) |
| `value` | 阈值 | 50.0（转化成本 50 元） |
| `duration` | 持续分钟数 | 5（条件持续满足 5 分钟才触发） |
| `time_range` | 时间范围 | "0-6"（凌晨 0-6 点） |
| `ad_min_days` | 最小投放天数 | 3（投放超过 3 天才评估） |

#### 执行动作 (ExecutionAction)

```json
{
  "type": "adjust_bid",
  "bid_adjust_ratio": 0.9,
  "notify_channels": ["email", "dingtalk"],
  "notify_users": ["admin@example.com"],
  "message": "成本已超出阈值，自动降出价 10%"
}
```

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `type` | 动作类型 | pause_ad / adjust_bid / raise_budget / notify / resume_ad / quick_start |
| `bid_adjust_ratio` | 出价调整比例 | 0.9（降低 10%） |
| `budget_raise_ratio` | 预算提升比例 | 1.2（提升 20%） |
| `budget_raise_amount` | 预算提升金额(分) | 10000（增加 100 元） |
| `notify_channels` | 通知渠道 | ["email", "sms", "dingtalk", "feishu"] |
| `notify_users` | 通知用户 | ["admin@example.com"] |
| `message` | 自定义消息 | 自定义通知内容 |

### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/hosting/dashboard` | 看板统计 + 趋势数据 |
| `GET` | `/api/v1/hosting/rules` | 规则列表 |
| `POST` | `/api/v1/hosting/rules` | 创建规则 |
| `GET` | `/api/v1/hosting/rules/:id` | 规则详情 |
| `PUT` | `/api/v1/hosting/rules/:id` | 更新规则 |
| `DELETE` | `/api/v1/hosting/rules/:id` | 删除规则 |
| `POST` | `/api/v1/hosting/rules/:id/toggle` | 切换启用/禁用 |
| `POST` | `/api/v1/hosting/rules/:id/test` | 测试规则（评估不执行） |
| `GET` | `/api/v1/hosting/executions` | 执行记录列表 |
| `GET` | `/api/v1/hosting/executions/:id` | 执行记录详情 |
| `POST` | `/api/v1/hosting/executions/:id/rollback` | 回滚执行 |
| `GET` | `/api/v1/hosting/alerts` | 告警列表 |
| `GET` | `/api/v1/hosting/alerts/:id` | 告警详情 |
| `POST` | `/api/v1/hosting/alerts/:id/read` | 标记已读 |
| `POST` | `/api/v1/hosting/alerts/:id/handle` | 处理告警 |
| `POST` | `/api/v1/hosting/alerts/:id/ignore` | 忽略告警 |
| `POST` | `/api/v1/hosting/trigger/evaluate` | 手动触发规则评估 |
| `POST` | `/api/v1/hosting/trigger/collect` | 手动触发快照采集 |

### 定时任务（托管相关）

| 任务 | Cron 表达式 | 执行频率 |
|------|------------|----------|
| 性能快照采集 | `*/5 * * * *` (每5分) | 从日报表聚合生成快照 |
| 规则评估执行 | `30 */5 * * * *` (每5分30秒) | 评估启用规则，执行匹配动作 |
| 广告状态检查 | `*/10 * * * *` (每10分) | 检查诊断/拒审状态 |
| 每日计数器重置 | `0 0 * * *` (每日零点) | 重置规则今日执行计数 |
| 旧记录清理 | `0 3 * * *` (每日凌晨3点) | 清理 30 天前的执行记录 |
| 旧快照清理 | `30 3 * * *` (每日凌晨3:30) | 清理 7 天前的快照 |

### 代码结构

```
backend/internal/
├── model/hosting.go                    # 数据模型（4 张表 + JSON 字段）
├── engine/
│   ├── rule_engine.go                  # 规则引擎（条件评估）
│   └── action_executor.go              # 动作执行器（API 调用 + 回滚）
├── service/
│   ├── hosting_rule_service.go         # 规则 CRUD 服务
│   ├── hosting_executor_service.go     # 托管执行核心（评估、快照采集、统计）
│   └── hosting_alert_service.go        # 告警管理服务
├── handler/
│   ├── hosting_rule_handler.go         # 规则 API
│   ├── hosting_execution_handler.go    # 执行记录 + 看板 API
│   └── hosting_alert_handler.go        # 告警 API
└── scheduler/scheduler.go              # 注册 6 个托管定时任务

frontend/src/
├── types/index.ts                      # TypeScript 类型定义（托管部分）
├── api/index.ts                        # API 接口（托管部分）
├── router/index.ts                     # 路由配置（托管部分）
├── components/Layout.vue               # 侧边栏（托管菜单）
└── views/hosting/
    ├── HostingDashboard.vue            # 托管看板
    ├── HostingRule.vue                 # 规则列表
    ├── HostingRuleCreate.vue           # 创建/编辑规则
    ├── HostingExecution.vue            # 执行记录
    └── HostingAlert.vue                # 告警通知
```

---

## 快速开始

### 环境要求

- **Go** 1.21+
- **Node.js** 18+
- **MySQL** 8.0+

### 1. 数据库初始化

```sql
CREATE DATABASE IF NOT EXISTS marketing CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 后端启动

```bash
cd backend

# 修改 config.yaml 中的数据库配置（用户名/密码/库名）
# 默认: root / admin1234 / marketing

# 启动
go run main.go

# 服务运行在 http://127.0.0.1:8080
# 首次启动会自动创建所有数据表（AutoMigrate）
```

### 3. 前端启动

```bash
cd frontend

npm install
npm run dev

# 开发服务器运行在 http://localhost:3000
# /api 请求自动代理到 http://127.0.0.1:8080
```

### 4. 首次使用

1. 访问 `http://localhost:3000/register` 注册账号
2. 登录后进入数据看板
3. 前往「账号授权」页面完成腾讯广告 OAuth 授权
4. 前往「数据抓取管理」手动触发抓取或等待定时任务自动执行
5. 查看各报表和广告结构数据

---

## 定时任务调度

| 任务 | Cron 表达式 | 执行频率 |
|------|------------|----------|
| 小时报表采集 | `10 * * * *` | 每小时 10 分 |
| 日报表采集 | `0 2 * * *` | 每天凌晨 2 点 |
| 广告系列采集 | `0 * * * *` | 每小时 |
| 广告组采集 | `0 * * * *` | 每小时 |
| 广告采集 | `0 * * * *` | 每小时 |
| 广告创意采集 | `0 * * * *` | 每小时 |
| 广告素材采集 | `0 * * * *` | 每小时 |
| Token 刷新 | `0 */12 * * *` | 每 12 小时 |
| 清理旧任务 | `0 1 * * *` | 每天凌晨 1 点 |
| 重试失败任务 | `*/5 * * * *` | 每 5 分钟 |

---

## 配置文件

主配置文件：`backend/config.yaml`，包含以下配置项：

| 配置块 | 说明 |
|--------|------|
| `server` | 服务端口、运行模式（debug/release） |
| `database` | MySQL 连接信息、连接池参数 |
| `jwt` | JWT 密钥、过期时间 |
| `tencent_ad` | 腾讯广告 API 密钥和回调地址 |
| `log` | 日志级别、文件路径、轮转策略 |
| `report` | 报表生成参数 |
| `scheduler` | 定时任务 Cron 表达式 |
| `crawler` | 抓取并发数、重试、限流参数 |

---

## 项目结构总览

```
marketing-platform/
├── README.md                  # 本文件
├── backend/                   # Go 后端
│   ├── main.go                # 入口
│   ├── config.yaml            # 配置文件
│   └── internal/              # 核心代码
│       ├── config/            # 配置管理
│       ├── database/          # 数据库
│       ├── model/             # 数据模型
│       ├── handler/           # HTTP 处理器
│       ├── service/           # 业务逻辑
│       ├── crawler/           # 数据抓取引擎
│       ├── scheduler/         # 定时调度
│       ├── middleware/        # 中间件
│       ├── router/            # 路由
│       └── logger/            # 日志
└── frontend/                  # Vue 3 前端
    ├── src/
    │   ├── api/               # API 接口
    │   ├── router/            # 路由
    │   ├── stores/            # 状态管理
    │   ├── views/             # 页面组件
    │   ├── components/        # 公共组件
    │   ├── types/             # 类型定义
    │   └── styles/            # 样式
    ├── vite.config.ts         # Vite 配置
    └── package.json           # 依赖配置
```