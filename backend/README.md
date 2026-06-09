# 广告数据洞察平台 - 后端

基于 Go + Gin 的广告数据洞察平台后端服务。

## 技术栈

- Go 1.21+
- Gin v1.9+ (Web框架)
- GORM v2.0+ (ORM)
- MySQL 8.0+
- JWT (认证)
- Zap + Lumberjack (日志)
- Cron (定时任务)

## 项目结构

```
backend/
├── cmd/
│   └── server/           # 程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/       # 数据库连接
│   ├── handler/        # HTTP处理器
│   ├── logger/         # 日志模块
│   ├── middleware/     # 中间件
│   ├── model/          # 数据模型
│   ├── router/         # 路由定义
│   ├── scheduler/      # 定时任务
│   └── service/       # 业务逻辑层
├── config.yaml         # 配置文件
├── go.mod
├── go.sum
└── main.go
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+

### 2. 配置

编辑 `config.yaml` 文件，配置数据库连接、JWT密钥、腾讯广告API凭证等。

### 3. 安装依赖

```bash
go mod download
```

### 4. 运行

```bash
go run main.go
```

### 5. API文档

启动服务后访问 `http://localhost:8080/health` 检查服务状态。

#### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/auth/login | 用户登录 |
| POST | /api/v1/auth/register | 用户注册 |

#### OAuth接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/oauth/url | 获取授权URL |
| GET | /api/v1/oauth/callback | 授权回调 |

#### 报表接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/reports/daily | 获取日报表 |
| GET | /api/v1/reports/hourly | 获取小时报表 |
| GET | /api/v1/reports/target | 获取定向报表 |
| GET | /api/v1/reports/trend | 获取趋势数据 |

#### 广告系列接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/campaigns | 获取广告系列列表 |
| GET | /api/v1/campaigns/:id | 获取广告系列详情 |

**参数说明：**
- `account_id` - 账号ID（可选）
- `page` - 页码，默认1
- `page_size` - 每页数量，默认20

#### 广告组接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/adgroups | 获取广告组列表 |
| GET | /api/v1/adgroups/:id | 获取广告组详情 |

**参数说明：**
- `account_id` - 账号ID（可选）
- `campaign_id` - 广告系列ID（可选）
- `page` - 页码，默认1
- `page_size` - 每页数量，默认20

#### 广告接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/ads | 获取广告列表 |
| GET | /api/v1/ads/:id | 获取广告详情 |

**参数说明：**
- `account_id` - 账号ID（可选）
- `adgroup_id` - 广告组ID（可选）
- `page` - 页码，默认1
- `page_size` - 每页数量，默认20

#### 广告创意接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/creatives | 获取广告创意列表 |
| GET | /api/v1/creatives/:id | 获取广告创意详情 |

**参数说明：**
- `account_id` - 账号ID（可选）
- `ad_id` - 广告ID（可选）
- `page` - 页码，默认1
- `page_size` - 每页数量，默认20

#### 广告素材接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/materials | 获取广告素材列表 |
| GET | /api/v1/materials/:id | 获取广告素材详情 |
| GET | /api/v1/materials/stats | 获取素材统计 |

**参数说明：**
- `account_id` - 账号ID（必填）
- `material_type` - 素材类型（可选）：1-图片, 2-视频, 3-文本, 4-卡片, 5-小程序
- `page` - 页码，默认1
- `page_size` - 每页数量，默认20

#### 数据抓取管理接口 (需认证)

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
| POST | /api/v1/crawler/trigger/campaign | 触发广告系列抓取 |

## 开发规范

遵循项目根目录下的 `总览.md` 文件中的规定。
