# 广告数据洞察平台 - 前端

基于 Vue 3 + TypeScript + Element Plus 的广告数据可视化平台前端应用。

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - JavaScript 的超集，提供类型检查
- **Vite** - 下一代前端构建工具
- **Element Plus** - 基于 Vue 3 的组件库
- **ECharts 5** - 数据可视化图表库
- **Axios** - HTTP 请求库
- **Pinia** - Vue 3 状态管理
- **Vue Router 4** - Vue 3 官方路由管理器

## 项目结构

```
frontend/
├── src/
│   ├── api/               # API 接口封装
│   │   ├── axios.ts        # Axios 实例和拦截器
│   │   └── index.ts       # API 接口列表
│   ├── components/        # 公共组件
│   │   └── Layout.vue     # 主布局组件
│   ├── router/            # 路由配置
│   │   └── index.ts      # 路由定义
│   ├── stores/            # Pinia 状态管理
│   │   ├── auth.ts        # 认证状态
│   │   └── crawler.ts     # 抓取状态
│   ├── styles/            # 全局样式
│   │   └── index.css      # 全局 CSS
│   ├── types/             # TypeScript 类型定义
│   │   └── index.ts       # 类型声明
│   ├── utils/             # 工具函数
│   └── views/             # 页面组件
│       ├── Dashboard.vue  # 数据看板
│       ├── Login.vue       # 登录页
│       ├── OAuth.vue       # 账号授权
│       ├── Crawler.vue     # 数据抓取管理
│       ├── Campaigns.vue    # 广告系列
│       ├── AdGroups.vue    # 广告组
│       ├── Ads.vue         # 广告
│       ├── Materials.vue   # 广告素材
│       └── Reports/        # 报表模块
│           ├── DailyReport.vue   # 日报表
│           ├── HourlyReport.vue  # 小时报表
│           └── TargetReport.vue  # 定向报表
├── index.html             # HTML 入口
├── package.json           # 项目依赖
├── vite.config.ts        # Vite 配置
└── tsconfig.json         # TypeScript 配置
```

## 功能模块

### 1. 数据看板
- 核心指标统计卡片（曝光、点击、花费、转化）
- 数据趋势图表（折线图）
- 花费占比饼图
- 转化漏斗图
- 账户概览

### 2. 报表数据
- **日报表**：按日期范围查询广告日报数据
- **小时报表**：查看特定日期的每小时数据
- **定向报表**：查看不同定向标签的投放效果

### 3. 广告结构
- **广告系列**：管理广告推广计划
- **广告组**：管理广告组层级
- **广告**：管理具体广告
- **广告素材**：查看和管理广告素材

### 4. 数据抓取管理
- 抓取任务统计
- 手动触发抓取（小时报表、日报表、广告结构）
- 任务列表和进度监控
- 任务取消和重试

### 5. 账号授权
- OAuth 2.0 授权管理
- 已授权账号列表
- 账号状态监控

## 快速开始

### 1. 安装依赖

```bash
npm install
# 或
yarn install
```

### 2. 开发模式

```bash
npm run dev
```

访问 http://localhost:3000

### 3. 构建生产版本

```bash
npm run build
```

### 4. 预览生产版本

```bash
npm run preview
```

## 接口说明

前端通过 `/api/v1` 前缀访问后端接口，已配置代理。

### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/auth/login | 用户登录 |
| POST | /api/v1/auth/register | 用户注册 |

### 报表接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/reports/daily | 获取日报表 |
| GET | /api/v1/reports/hourly | 获取小时报表 |
| GET | /api/v1/reports/target | 获取定向报表 |
| GET | /api/v1/reports/trend | 获取趋势数据 |

### 广告系列接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/campaigns | 获取广告系列列表 |
| GET | /api/v1/campaigns/:id | 获取广告系列详情 |

### 广告组接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/adgroups | 获取广告组列表 |
| GET | /api/v1/adgroups/:id | 获取广告组详情 |

### 广告接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/ads | 获取广告列表 |
| GET | /api/v1/ads/:id | 获取广告详情 |

### 广告创意接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/creatives | 获取广告创意列表 |
| GET | /api/v1/creatives/:id | 获取广告创意详情 |

### 广告素材接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/materials | 获取广告素材列表 |
| GET | /api/v1/materials/:id | 获取广告素材详情 |
| GET | /api/v1/materials/stats | 获取素材统计 |

### 数据抓取管理接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/crawler/start | 手动启动抓取 |
| GET | /api/v1/crawler/statistics | 获取抓取统计 |
| GET | /api/v1/crawler/tasks | 查询任务列表 |
| GET | /api/v1/crawler/tasks/running | 获取正在运行的任务 |
| POST | /api/v1/crawler/tasks/:task_id/cancel | 取消任务 |
| POST | /api/v1/crawler/tasks/:task_id/retry | 重试任务 |
| POST | /api/v1/crawler/trigger/hourly | 触发小时报表抓取 |
| POST | /api/v1/crawler/trigger/daily | 触发动报表抓取 |
| POST | /api/v1/crawler/trigger/struct | 触发广告结构抓取 |

## 状态管理

使用 Pinia 进行状态管理，主要 store：

- `useAuthStore` - 用户认证状态（token、用户信息）
- `useCrawlerStore` - 数据抓取状态（账号列表、任务列表、统计）

## 路由守卫

所有路由（除 `/login` 外）都需要登录后才能访问，未登录会自动跳转到登录页。

## 环境变量

如需修改 API 地址，可修改 `vite.config.ts` 中的代理配置：

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',  // 后端地址
      changeOrigin: true
    }
  }
}
```
