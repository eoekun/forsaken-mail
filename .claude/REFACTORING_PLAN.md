# Forsaken-Mail 重构实现计划

> 本文档是面向开发者的重构提示词，用于指导从 Node.js 迁移到 Go + 现代前端的完整实现。

---

## 项目上下文

Forsaken-Mail 是一个自托管临时邮箱服务。当前版本使用 Node.js + Express + Socket.IO + jQuery，存在安全缺陷（XSS、无认证）、无数据持久化、依赖已停维库等问题。本次重构目标是用 Go 重写后端，用 React + Vite + Tailwind 重写前端，保持核心功能不变，同时修复所有已知问题。

当前代码位于项目根目录，重构后的新代码建议在新分支上开发。

---

## 技术栈约束

后端必须使用：
- Go 1.24+
- `net/http` 标准库（不使用 gin/echo 等框架）
- `github.com/emersion/go-smtp` — SMTP 服务器
- `github.com/jhillyerd/enmime` — MIME 邮件解析
- `github.com/coder/websocket` — WebSocket
- `github.com/mattn/go-sqlite3` — SQLite 驱动（需要 CGO_ENABLED=1）
- `golang.org/x/oauth2` — OAuth2 客户端
- `golang.org/x/time/rate` — 速率限制
- `gopkg.in/natefinish/lumberjack.v2` — 日志轮转

前端必须使用：
- Vite 6 构建
- React 19 + React Router 7
- Tailwind CSS 4 + DaisyUI 5
- DOMPurify 3.x（邮件 HTML 消毒）
- 原生 WebSocket API（不用 Socket.IO）

前端不允许引入状态管理库（Redux/Zustand 等），用 React 自带的 useState/useContext 足够。

---

## 目录结构

```
forsaken-mail/
├── cmd/server/main.go              # 入口
├── internal/
│   ├── config/config.go            # 环境变量 → Config struct
│   ├── smtp/
│   │   ├── server.go               # go-smtp Backend + Session 实现
│   │   └── ratelimit.go            # 按 IP 速率限制
│   ├── mail/
│   │   ├── store.go                # SQLite CRUD（mails 表）
│   │   ├── router.go               # 收件 → 存库 → WS 推送 → Webhook
│   │   └── cleanup.go              # 定时清理（按时间 + 按数量）
│   ├── ws/hub.go                   # WebSocket 连接管理 + 消息广播
│   ├── auth/
│   │   ├── oauth.go                # GitHub/Google OAuth2 流程
│   │   ├── session.go              # cookie session 加密/解密
│   │   └── middleware.go           # 认证中间件 + 白名单校验
│   ├── webhook/dingtalk.go         # 钉钉 Webhook 通知
│   ├── audit/store.go              # 审计日志 CRUD（audit_logs 表）
│   ├── settings/store.go           # B 类配置 CRUD（settings 表）
│   ├── api/
│   │   ├── router.go               # 路由注册（公开 + 受保护）
│   │   ├── mail.go                 # GET /api/mails
│   │   ├── config.go               # GET /api/config, GET /api/domain-test
│   │   ├── auth.go                 # login/callback/logout handler
│   │   ├── admin.go                # admin/audit-logs, admin/settings, admin/status
│   │   ├── webhook.go              # POST /api/webhook/test
│   │   ├── ws.go                   # GET /ws（WebSocket 升级）
│   │   └── health.go               # GET /api/health
│   └── logger/logger.go            # slog + lumberjack 初始化
├── web/                            # 前端 Vite + React 项目
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   └── src/
│       ├── main.jsx                # React 入口（createRoot）
│       ├── App.jsx                 # 根组件（路由 + 认证状态）
│       ├── hooks/
│       │   └── useWebSocket.js     # WebSocket 连接 + shortId + 邮件状态
│       ├── pages/
│       │   ├── LoginPage.jsx       # 登录页
│       │   ├── MainPage.jsx        # 主页面（邮箱 + 邮件列表 + 详情）
│       │   └── AdminPage.jsx       # Admin 页面（三个 Tab）
│       ├── components/
│       │   ├── Navbar.jsx          # 顶部导航栏
│       │   ├── MailboxAddress.jsx  # 临时邮箱地址（显示/编辑/复制）
│       │   ├── MailList.jsx        # 邮件列表表格
│       │   ├── MailDetail.jsx      # 邮件详情卡片
│       │   ├── MailHistory.jsx     # 最近邮箱历史
│       │   ├── HelpModal.jsx       # 帮助模态框（DNS/Webhook 测试）
│       │   ├── AuditLogTab.jsx     # Admin: 审计日志
│       │   ├── SettingsTab.jsx     # Admin: 系统配置
│       │   └── StatusTab.jsx       # Admin: 系统状态
│       ├── lib/
│       │   └── api.js              # fetch 封装（含 401 重定向）
│       └── style.css               # Tailwind 入口
├── embed/.gitkeep                  # Vite 构建产物（Go embed）
├── go.mod
├── Dockerfile                      # 三阶段构建
├── docker-compose.yml
├── .env.example
└── doc/                            # 分析文档（参考用，不参与构建）
```

---

## 数据库设计

使用单个 SQLite 文件，启动时自动建表。三张表：

**mails** — 邮件存储
- id INTEGER PRIMARY KEY AUTOINCREMENT
- short_id TEXT NOT NULL
- from_addr TEXT NOT NULL
- to_addr TEXT NOT NULL
- subject TEXT NOT NULL DEFAULT ''
- text_body TEXT NOT NULL DEFAULT ''
- html_body TEXT NOT NULL DEFAULT ''
- raw_size INTEGER NOT NULL DEFAULT 0
- created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
- 索引：short_id, created_at

**settings** — B 类配置（运行时可改）
- key TEXT PRIMARY KEY
- value TEXT NOT NULL
- updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP

**audit_logs** — 审计日志
- id INTEGER PRIMARY KEY AUTOINCREMENT
- event TEXT NOT NULL（LOGIN / LOGIN_DENIED / LOGOUT / MAILBOX_CREATE / MAIL_RECEIVED / MAIL_DROPPED / WEBHOOK_SENT / CONFIG_CHANGED）
- email TEXT NOT NULL DEFAULT ''
- detail TEXT NOT NULL DEFAULT '{}'（JSON）
- ip TEXT NOT NULL DEFAULT ''
- created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
- 索引：event, created_at

---

## 配置分层

### A 类：环境变量（启动时读取，运行时不可改）

| 变量 | 默认值 | 必填 |
|---|---|---|
| PORT | 3000 | 否 |
| MAILIN_HOST | 0.0.0.0 | 否 |
| MAILIN_PORT | 25 | 否 |
| MAIL_HOST | — | 是 |
| SITE_TITLE | Forsaken Mail | 否 |
| DB_PATH | ./data/forsaken-mail.db | 否 |
| OAUTH_PROVIDER | github | 否（github 或 google） |
| OAUTH_CLIENT_ID | — | 是 |
| OAUTH_CLIENT_SECRET | — | 是 |
| SESSION_SECRET | — | 是（openssl rand -hex 32） |
| LOG_LEVEL | info | 否 |
| LOG_FILE | ""（stdout） | 否 |
| LOG_MAX_SIZE_MB | 10 | 否 |
| LOG_MAX_BACKUPS | 3 | 否 |

### B 类：DB settings 表（首次用环境变量 seed，运行时通过 Admin API 改）

| key | 环境变量 seed | 默认值 |
|---|---|---|
| mail_host | MAIL_HOST | 必填 |
| site_title | SITE_TITLE | Forsaken Mail |
| allowed_emails | ALLOWED_EMAILS | ""（允许所有） |
| keyword_blacklist | KEYWORD_BLACKLIST | admin,postmaster,system,webmaster,administrator,hostmaster,service,server,root |
| dingtalk_webhook_token | DINGTALK_WEBHOOK_TOKEN | "" |
| dingtalk_webhook_message | DINGTALK_WEBHOOK_MESSAGE | new email received. |
| mail_retention_hours | MAIL_RETENTION_HOURS | 1 |
| mail_max_count | MAIL_MAX_COUNT | 100 |
| max_mail_size_bytes | MAX_MAIL_SIZE_BYTES | 1048576 |
| audit_retention_days | AUDIT_RETENTION_DAYS | 7 |
| audit_max_count | AUDIT_MAX_COUNT | 5000 |

读取顺序：DB 优先。DB 中不存在的 key 用环境变量 seed 写入。

注意：`mail_host` 和 `site_title` 同时出现在 A 类和 B 类中。A 类环境变量仅用于首次 seed，运行时读取以 DB 为准。这意味着一旦 Admin 页面修改了这两个值，环境变量中的值将被忽略（直到 DB 被删除或 key 被清除）。

---

## 核心模块实现要点

### SMTP 服务器

- 实现 `go-smtp` 的 `Backend` 接口（`NewSession`）和 `Session` 接口（`Login` 返回 nil、`Mail`、`Rcpt`、`Data`、`Reset`、`Logout`）
- `Data` 方法中用 `io.LimitReader` 限制大小，然后 `enmime.ReadEnvelope` 解析
- `Login` 直接返回 nil（无 SMTP 认证，这是临时邮箱的设计）
- 在 `NewSession` 中按 IP 做速率限制（`rate.Limiter`），超出返回错误
- 解析完成后调用 `mail.Router.Handle`

### 邮件路由

`Router.Handle(from, toList, envelope)` 逻辑：
1. 遍历 toList，从每个地址提取 shortId（@前面的部分，需要校验域名匹配 MAIL_HOST）
2. 对每个 shortId：存库 → 通过 Hub.SendTo 推送到 WebSocket → 异步发钉钉通知
3. 存库前检查 shortId 是否在 keyword_blacklist 中（可选，取决于业务需求）

### WebSocket Hub

- 核心结构：`map[string]map[*Client]bool`（shortId → 客户端集合）
- 通过 channel 串行化 register/unregister/broadcast 三个操作，避免 mutex 竞争
- Client 结构体包含 conn、send channel、shortId
- 每个 Client 启动一个 readPump（接收客户端消息）和一个 writePump（发送服务端消息）
- 断开连接时自动 unregister 并清理

### WebSocket 消息协议

统一使用 JSON，`type` 字段区分消息类型，全部 snake_case。

客户端 → 服务端：
- `{"type":"request_shortid"}`
- `{"type":"set_shortid","short_id":"alexsun789"}`

服务端 → 客户端：
- `{"type":"shortid","short_id":"alexsun789"}`
- `{"type":"mail","data":{"id":1,"from":"...","to":"...","subject":"...","html":"...","created_at":"..."}}`
- `{"type":"error","message":"short id in blacklist"}`

### OAuth 认证

- 定义 Provider 接口：`AuthCodeURL(state)`, `Exchange(ctx, code)`, `GetEmail(ctx, token)`
- 实现 GitHubProvider 和 GoogleProvider
- GitHub 获取邮箱：调用 `GET /user/emails` API，取 primary 且 verified 的邮箱
- Google 获取邮箱：调用 `GET /userinfo` API，取 email 字段
- Session 加密：用 AES-GCM 加密 JSON 序列化的 session 数据，密钥来自 SESSION_SECRET
- Cookie：HttpOnly=true, SameSite=Lax, Path=/, MaxAge=86400

### 认证中间件

- 从 cookie 解密 session → 校验过期时间 → 从 settings 读取 allowed_emails → 校验邮箱是否在白名单
- 白名单为空时允许所有已认证用户
- 白名单支持精确邮箱匹配
- 未认证 → 302 到 /login；白名单拒绝 → 403

### 邮件清理

启动一个 goroutine，每 5 分钟执行：
1. `DELETE FROM mails WHERE created_at < now - mail_retention_hours`
2. 如果 count > mail_max_count，`DELETE FROM mails WHERE id NOT IN (SELECT id FROM mails ORDER BY created_at DESC LIMIT mail_max_count)`
3. 同样清理 audit_logs（按 audit_retention_days 和 audit_max_count）

### 审计日志

在以下位置调用 `audit.Record(event, email, detail, ip)`：
- auth/callback 成功 → LOGIN
- auth/callback 白名单拒绝 → LOGIN_DENIED
- auth/logout → LOGOUT
- ws Hub 分配 shortId → MAILBOX_CREATE
- mail Router 收到邮件 → MAIL_RECEIVED
- SMTP Data 邮件超大拒绝 → MAIL_DROPPED
- webhook 通知发送后 → WEBHOOK_SENT
- admin settings 更新 → CONFIG_CHANGED

---

## API 路由表

### 公开路由（无认证）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/login` | 返回登录页 HTML（内嵌在前端 SPA 中） |
| GET | `/auth/{provider}/login` | 构造 OAuth URL 并 302 重定向 |
| GET | `/auth/{provider}/callback` | 处理 OAuth 回调，设置 session cookie |
| GET | `/api/health` | 返回 200 + `{"status":"ok"}` |

### 受保护路由（认证中间件拦截）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/config` | 返回 `{"host":"...","siteTitle":"..."}` |
| GET | `/api/mails?shortId=xxx` | 返回指定 shortId 的邮件列表（JSON 数组） |
| GET | `/api/domain-test?domain=xxx` | DNS MX 记录诊断 |
| POST | `/api/webhook/test` | Body: `{"token":"...","message":"..."}` |
| GET | `/ws` | WebSocket 升级 |
| GET | `/api/admin/audit-logs?event=xxx&offset=0&limit=50` | 审计日志查询 |
| GET | `/api/admin/settings` | 返回所有 B 类配置 |
| PUT | `/api/admin/settings` | Body: `{"key":"value",...}` 批量更新 |
| GET | `/api/admin/status` | 返回系统状态 JSON |
| GET | `/auth/logout` | 清除 session cookie，302 到 /login |
| GET | `/` | 前端 SPA（Go embed 静态资源） |

---

## 前端实现要点

### Vite 配置

- 使用 `@vitejs/plugin-react`
- build.outDir 指向 `../embed`，供 Go `//go:embed` 嵌入
- build.emptyOutDir: true
- dev server 配置 proxy：`/api` → `http://localhost:3000`，`/ws` → `ws://localhost:3000`，`/auth` → `http://localhost:3000`

### 路由（React Router）

```
/          → MainPage（需认证，未认证重定向到 /login）
/login     → LoginPage
/admin     → AdminPage（需认证）
```

App.jsx 根组件：
- 启动时 fetch `/api/config` 获取站点配置和认证状态
- 未认证时将所有路由重定向到 `/login`
- 用 React Context 向子组件提供 config（host、siteTitle）和 user（email）

### 组件设计

**App.jsx** — 根组件。初始化时调用 `/api/config` 判断认证状态。提供 AuthContext（config + user + logout 函数）。渲染 Router。

**useWebSocket.js** — 自定义 hook。管理 WebSocket 连接生命周期：connect、reconnect（指数退避 1s→30s）。返回 `{ shortId, setShortId, mails, clearMails }`。onmessage 收到 mail 时追加到 mails 数组，同时触发 `new Notification()`。邮件状态（mails、selectedMail）全部在此 hook 中管理，不拆分。

**MainPage.jsx** — 主页面布局。组合 Navbar + MailboxAddress + MailHistory + MailList + MailDetail + HelpModal。通过 useWebSocket 获取实时数据和邮件状态。

**MailboxAddress.jsx** — 邮箱地址组件。显示 `shortId@host`，支持复制（navigator.clipboard.writeText）、刷新（发送 `{"type":"request_shortid"}`）、编辑（切换为 input，正则校验 `^[a-z0-9._\-+]{1,64}$`，提交 `{"type":"set_shortid","short_id":"xxx"}`）。

**MailList.jsx** — 邮件列表。接收 mails 数组，渲染为 DaisyUI table。点击行设置 selectedMail。

**MailDetail.jsx** — 邮件详情。接收 selectedMail。HTML 正文用 `DOMPurify.sanitize(mail.html)` 消毒后通过 dangerouslySetInnerHTML 渲染。纯文本内容直接展示。

**MailHistory.jsx** — 最近邮箱。从 localStorage 读取 shortId 历史（最多 6 条），渲染为 DaisyUI button group。点击切换 shortId（发送 `{"type":"set_shortid","short_id":"xxx"}`）。

**HelpModal.jsx** — 帮助模态框。DNS 测试（fetch GET /api/domain-test）和 Webhook 测试（fetch POST /api/webhook/test），结果展示在 pre 标签中。

**AdminPage.jsx** — Admin 页面。三个 DaisyUI tab：AuditLogTab、SettingsTab、StatusTab。

**AuditLogTab.jsx** — 审计日志。fetch GET /api/admin/audit-logs，支持 event 下拉筛选，表格展示，分页加载。

**SettingsTab.jsx** — 系统配置。fetch GET /api/admin/settings 渲染表单（仅 B 类可编辑配置），修改后 PUT 保存。保存成功后显示 toast 提示。不展示 A 类配置（环境变量）。

**StatusTab.jsx** — 系统状态。fetch GET /api/admin/status 展示各项指标。

### XSS 防护

- 所有邮件 HTML 渲染前必须经过 DOMPurify.sanitize()
- MailDetail 组件中：`<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(mail.html) }} />`
- 用户输入（shortId 编辑框）用正则校验后才发送到服务端
- API 返回的文本内容通过 React 的 `{text}` 自动转义（React 默认转义文本节点）

---

## 安全要求

- CSP Header：`default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' ws: wss:`
  - 注意：Vite 生产构建默认将 JS 内联到 HTML，会被 `script-src 'self'` 拦截。解决方式：vite.config.js 中设置 `build.rollupOptions.output.inlineDynamicImports: false`，或使用 `vite-plugin-csp` 注入 nonce。推荐前者（外链 JS），最简单。
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- Cookie: HttpOnly, SameSite=Lax, Secure（生产环境）
- SMTP 速率限制：每 IP 每秒最多 1 个连接
- SMTP 邮件大小限制：通过 max_mail_size_bytes 配置，超出直接拒绝
- WebSocket 连接必须先通过 HTTP 认证（复用 session cookie）
- 日志中不记录 SESSION_SECRET、OAUTH_CLIENT_SECRET、session cookie 值

---

## Docker 部署

### Dockerfile（三阶段）

1. `FROM node:24-alpine AS frontend` — 安装前端依赖，vite build 输出到 embed/
2. `FROM golang:1.24-alpine AS backend` — go mod download，复制源码和 embed，go build（CGO_ENABLED=1）
3. `FROM alpine:3.21` — apk add ca-certificates sqlite-libs，复制二进制，VOLUME /data，EXPOSE 25 3000

### docker-compose.yml

- 单容器，端口映射 25:25 和 3000:3000
- volume: forsaken-data:/data, forsaken-logs:/var/log/forsaken-mail
- 资源限制 128MB
- healthcheck: wget /api/health 每 30 秒
- 环境变量通过 ${VAR:-default} 引用 .env

---

## 实施顺序

### Phase 1：后端骨架（先跑通 SMTP → 存库 → WS 推送）

1. `go mod init`，创建目录结构
2. config — 解析环境变量，返回 Config struct
3. logger — slog 初始化 + lumberjack
4. settings — settings 表 CRUD + 环境变量 seed 逻辑
5. mail/store + audit/store — SQLite 初始化（建表 mails + audit_logs + settings，Save/ListByShortId/Record/Query）
6. ws/hub — Hub + Client + readPump + writePump
7. mail/router — 收件 → 存库 → Hub.SendTo（先不做 webhook）
8. mail/cleanup — 定时清理 goroutine（mails + audit_logs）
9. smtp — Backend + Session 实现，调用 mail/router
10. cmd/server/main.go — 串联所有模块，启动 SMTP + HTTP
11. 验证：用 telnet 或 swaks 发邮件到 :25，确认能存库

### Phase 2：认证 + API

1. auth/oauth — GitHub + Google Provider 实现
2. auth/session — AES-GCM 加密/解密 session cookie
3. auth/middleware — 认证中间件 + 白名单校验
4. api/auth.go — login/callback/logout handler
5. api/config.go — /api/config + /api/domain-test
6. api/mail.go — GET /api/mails
7. api/webhook.go — POST /api/webhook/test
8. api/admin.go — admin/audit-logs, admin/settings, admin/status
9. api/health.go — GET /api/health
10. api/router.go — 注册所有路由，区分公开/受保护
11. webhook/dingtalk.go — 钉钉通知（复用当前 Node.js 版本逻辑）
12. 串联：mail/router 中加入 webhook 调用 + audit 记录
13. 验证：curl 各 API 端点

### Phase 3：前端

1. `npm create vite`，安装 react、react-router-dom、tailwindcss、daisyui、dompurify
2. vite.config.js — 配置 @vitejs/plugin-react、build.outDir、dev proxy
3. App.jsx — Router + AuthContext + 认证状态初始化（fetch /api/config）
4. useWebSocket hook — 连接管理、重连、shortId 状态、mails 数组、selectedMail
5. LoginPage — OAuth 登录按钮（GitHub / Google）
6. MainPage — 组合 Navbar、MailboxAddress、MailHistory、MailList、MailDetail
7. MailDetail — DOMPurify.sanitize() 消毒邮件 HTML 后 dangerouslySetInnerHTML
8. HelpModal — DNS 测试 + Webhook 测试
9. AdminPage — AuditLogTab（表格+筛选）、SettingsTab（表单+保存）、StatusTab
10. 响应式适配
11. 验证：npm run dev，浏览器测试完整流程（登录 → 收邮件 → Admin 操作）

### Phase 4：部署

1. Dockerfile 三阶段构建
2. docker-compose.yml
3. .env.example
4. 安全 Header
5. README.md 更新
6. 验证：docker compose up，端到端测试

---

## 参考文档

- `doc/ARCHITECTURE.md` — 当前项目架构详解
- `doc/LANGUAGE_COMPARISON.md` — 语言选型分析
- `doc/REFACTORING_ANALYSIS.md` — 技术选型优缺点分析
