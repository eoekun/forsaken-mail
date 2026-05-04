# Forsaken-Mail 项目架构与技术栈文档

## 1. 项目概述

**Forsaken-Mail** 是一个**自托管的临时/一次性邮箱服务**，允许用户：

- 在自有域名上接收随机生成（或自定义）的临时邮箱地址（如 `alexmail123@mail.example.com`）
- 通过 Web UI 实时查看收到的邮件（无需刷新页面）
- 一键复制临时邮箱地址用于注册、测试等场景
- 在浏览器 `localStorage` 中维护最近使用的短 ID 历史记录
- 可选地通过**钉钉 Webhook**在收到新邮件时接收通知
- 在内置帮助对话框中测试 DNS/MX 配置和 Webhook 连通性

项目由 **Hongcai Deng**（2015）创建，采用 **GPL-2.0** 许可证。

---

## 2. 核心技术栈

| 层级 | 技术 | 版本/详情 |
|---|---|---|
| **语言** | JavaScript (Node.js) | `>=24`（`engines` 中指定） |
| **Web 框架** | Express.js | `^4.21.2` |
| **SMTP 服务器** | `smtp-server` (npm) | `^3.18.1` — 原生 SMTP 服务器，无需外部 MTA |
| **邮件解析** | `mailparser` (npm) | `^3.9.1` — 通过 `simpleParser` 解析原始邮件流 |
| **实时通信** | Socket.IO | `^4.8.3` — 基于 WebSocket 的实时推送 |
| **调试日志** | `debug` (npm) | `^4.4.1` |
| **前端 CSS 框架** | Semantic UI | `2.1.7`（通过 CDN 加载：`lib.baomitu.com`） |
| **前端 JS 库** | jQuery | `2.1.4`（通过 CDN 加载） |
| **前端工具库** | Clipboard.js `1.5.5`、Prism.js（原始邮件 JSON 语法高亮） |
| **容器化** | Docker (node:24-alpine)、Docker Compose |
| **反向代理** | Nginx `1.27-alpine`（Docker Compose 方案中） |

**无数据库**。所有状态均为临时性 — 服务端进程内存（已连接的 socket/短 ID）和浏览器端 `localStorage`。

---

## 3. 项目目录结构

```
forsaken-mail/
├── app.js                  # Express 应用配置（中间件、静态文件、API 挂载）
├── package.json            # npm 清单，依赖，脚本
├── Dockerfile              # 单容器 Docker 构建（node:24-alpine）
├── docker-compose.yml      # 多容器编排：app + nginx（含 Basic Auth）
├── .env.example            # 环境变量模板
├── .dockerignore           # Docker 构建排除文件
├── .gitignore              # Git 忽略规则
├── LICENSE                 # GPL-2.0
├── README.md               # 英文文档
├── README.zh-CN.md         # 中文文档
│
├── bin/
│   └── www                 # 入口文件 — 创建 HTTP 服务器，挂载 Socket.IO，启动监听
│
├── modules/
│   ├── config.js           # 集中配置（读取所有环境变量，含默认值/验证）
│   ├── mailin.js           # SMTP 服务器 — 监听收件，解析邮件，触发 'message' 事件
│   ├── io.js               # Socket.IO 处理器 — 管理连接客户端、短 ID 分配、邮件路由
│   └── dingtalk.js         # 钉钉 Webhook 通知器 — 收到新邮件时发送通知
│
├── routes/
│   └── api.js              # Express API 路由（REST 端点）
│
├── public/                 # 前端静态资源
│   ├── index.html          # 单页 HTML（整个 UI）
│   ├── css/
│   │   ├── app.css         # 自定义样式
│   │   └── prism.css       # Prism.js 语法高亮主题
│   └── js/
│       ├── app.js          # 主前端逻辑（jQuery + Socket.IO 客户端）
│       └── prism.js        # Prism.js 库（原始邮件 JSON 显示）
│
└── docker/
    └── nginx/
        ├── Dockerfile          # Nginx 镜像（含 htpasswd 支持）
        ├── default.conf        # Nginx 反向代理配置（Basic Auth + WebSocket 升级）
        └── docker-entrypoint.sh # 启动时根据环境变量生成 .htpasswd
```

---

## 4. 核心模块详解

### 4.1 入口文件：`bin/www`

- 创建 `http.Server`，包裹 Express 应用
- 将 `Socket.IO.Server` 挂载到 HTTP 服务器
- 初始化 Socket.IO 处理器（`modules/io.js`）
- 监听 `process.env.PORT`（默认 `3000`）

### 4.2 `app.js` — Express 应用

- 挂载 `express.json()` 中间件
- 以 1 小时缓存提供 `public/` 目录的静态文件
- 在 `/api` 路径挂载 `routes/api.js`
- 404 兜底和错误处理

### 4.3 `modules/mailin.js` — SMTP 服务器

- 使用 `smtp-server` 创建原生 SMTP 监听器（默认端口 25）
- **禁用认证**（`disabledCommands: ['AUTH']`）— 接收所有来信
- 使用 `mailparser` 的 `simpleParser` 解析传入的邮件流
- 构建标准化的邮件载荷，包含 `headers`（from, to, subject, date）、`text` 和 `html`
- 在 `mailin` EventEmitter 上触发 `'message'` 事件

### 4.4 `modules/io.js` — Socket.IO 处理器（核心业务逻辑）

- 在内存中维护已连接客户端的 `Map`：`shortId → socket`
- **短 ID 生成**：通过组合随机名字 + 标签 + 3 位数字创建可读 ID（如 `alexmail123`、`tomcloud456`），并验证是否在关键词黑名单中
- **Socket 事件**：
  - `request shortid` → 服务端生成并分配一个随机短 ID
  - `set shortid` → 客户端请求特定短 ID（经过验证和黑名单检查）
  - `disconnect` → 从在线 Map 中移除短 ID
- **邮件路由**：当 `mailin.js` 触发 `'message'` 事件时，提取收件人地址，在 `onlines` Map 中查找短 ID，若找到则通过 `socket.emit('mail', data)` 推送邮件数据
- 每封收到的邮件都会触发钉钉通知

### 4.5 `modules/config.js` — 配置中心

所有配置由**环境变量**驱动，附带合理默认值：

| 环境变量 | 默认值 | 说明 |
|---|---|---|
| `MAIL_HOST` | `disposable.dhc-app.com` | UI 中显示的邮件域名 |
| `SITE_TITLE` | `Forsaken Mail` | 站点标题 |
| `MAILIN_HOST` | `0.0.0.0` | SMTP 绑定地址 |
| `MAILIN_PORT` | `25` | SMTP 端口 |
| `PORT` | `3000` | Web UI 端口 |
| `KEYWORD_BLACKLIST` | — | 逗号分隔的禁止短 ID 关键词列表 |
| `DINGTALK_WEBHOOK_TOKEN` | — | 钉钉 Webhook Token |
| `DINGTALK_WEBHOOK_MESSAGE` | — | 钉钉通知消息模板 |

### 4.6 `modules/dingtalk.js` — 钉钉 Webhook

- 向钉钉机器人 Webhook API（`oapi.dingtalk.com/robot/send`）发送 HTTP POST
- 支持原始 `access_token` 值和完整 Webhook URL
- 构建格式化文本消息，包含发件人、收件人、主题、日期和文本预览（最多 200 字符）
- 导出 `sendTestNotification()` 函数供 API 测试端点使用

---

## 5. API 接口

### REST API（挂载在 `/api`）

| 方法 | 路径 | 功能 |
|---|---|---|
| `GET` | `/api/` | 健康检查（返回空 200 响应） |
| `GET` | `/api/config` | 返回 `{ host, siteTitle }` 供前端配置 |
| `GET` | `/api/domain-test?domain=X` | DNS 诊断 — 解析给定域名的 MX 记录，再解析每个 MX 主机的 A/AAAA 记录 |
| `POST` | `/api/webhook/test` | 接受 `{ token, message }` 请求体，向指定钉钉 Webhook 发送测试消息 |

### Socket.IO 事件（实时通信）

| 方向 | 事件 | 载荷 | 功能 |
|---|---|---|---|
| 客户端 → 服务端 | `request shortid` | — | 服务端生成并分配随机短 ID |
| 客户端 → 服务端 | `set shortid` | `string` | 客户端请求特定短 ID |
| 服务端 → 客户端 | `shortid` | `string` | 确认分配的短 ID |
| 服务端 → 客户端 | `mail` | `{ headers, text, html, ... }` | 实时推送新收到的邮件 |

---

## 6. 核心用户流程

```
用户打开 Web UI → 建立 Socket.IO 连接
        ↓
服务端分配随机短 ID（如 alexsun789）→ 用户看到 alexsun789@mail.example.com
        ↓
用户可通过编辑图标自定义短 ID（经过黑名单验证）
        ↓
用户复制邮箱地址用于注册/测试
        ↓
邮件到达该地址 → SMTP 服务器解析 → io.js 模块推送到对应 socket
        ↓
前端在表格中渲染邮件（发件人、主题、时间）并在卡片中显示 HTML 正文
        ↓
用户点击行查看完整邮件，或点击代码图标查看原始 JSON（Prism.js 语法高亮）
        ↓
浏览器通知显示新邮件（如已授权）
        ↓
最近使用的短 ID 存储在 localStorage 中，作为"最近邮箱"历史（最多 6 个）
```

---

## 7. 部署方案

### 方案 A：直接运行 Node.js

```bash
npm install && npm start    # 在端口 3000 启动（可通过 PORT 环境变量配置）
```

### 方案 B：Docker 单容器

```bash
docker build -t denghongcai/forsaken-mail .
docker run -d -p 25:25 -p 3000:3000 denghongcai/forsaken-mail
```

- 基于 `node:24-alpine`
- 暴露端口 25（SMTP）和 3000（Web UI）

### 方案 C：Docker Compose（推荐用于生产环境）

```bash
cp .env.example .env   # 编辑凭证
docker compose up -d --build
```

- **两个容器**：`app`（Node.js）+ `nginx`（反向代理）
- Nginx 在 Web UI 前添加 **HTTP Basic Authentication**
- SMTP（端口 25）直接从 app 容器暴露
- Web UI（默认端口 80，可通过 `NGINX_PORT` 配置）通过 nginx 暴露
- 资源限制：app 300MB 内存，nginx 30MB 内存

---

## 8. 外部依赖

### 数据库：无

- 所有状态均为**内存态**（`modules/io.js` 中的 `onlines` Map）
- 无持久化层 — 服务重启后所有连接会话丢失
- 浏览器端通过 `localStorage` 持久化（短 ID 历史、站点标题缓存）

### 外部服务

| 服务 | 用途 | 是否必需 |
|---|---|---|
| **DNS（MX 记录）** | 必须为邮件域名配置，以便外部 SMTP 服务器能投递到本服务器 | 是（接收邮件所需） |
| **钉钉 Webhook** | 收到邮件时可选推送到钉钉群 | 否（`DINGTALK_WEBHOOK_TOKEN` 为空时禁用） |
| **CDN（lib.baomitu.com）** | 加载 Semantic UI、jQuery、Clipboard.js | 是（前端运行所需） |

---

## 9. 前端架构

**无前端框架**（非 React/Vue/Angular），前端是一个**单页原生 HTML 应用**：

- **Semantic UI 2.1.7** — CSS 框架，用于布局、菜单、表格、卡片、模态框、按钮、表单、标签
- **jQuery 2.1.4** — DOM 操作和 AJAX 调用
- **Socket.IO 客户端** — 由 Socket.IO 服务器在 `/socket.io/socket.io.js` 自动提供
- **Clipboard.js 1.5.5** — 一键复制邮箱地址到剪贴板
- **Prism.js** — 原始邮件 JSON 视图的语法高亮

UI 语言为**中文**（zh-CN），标签如"发信人"、"主题"、"帮助说明"等。

主要 UI 组件：
- 顶部固定导航栏（Logo、帮助按钮、邮箱地址显示/输入）
- "最近邮箱"历史记录区域（切换/复制按钮）
- 邮件列表表格（按到达时间排序，最新在前）
- 邮件详情卡片（显示 HTML 正文）
- 帮助模态框（DNS 测试表单和钉钉 Webhook 测试表单）

---

## 10. 架构数据流图

```
┌─────────────────────────────────────────────────────────────┐
│                        外部世界                              │
│                                                             │
│   外部邮件服务器 ──(SMTP:25)──►  ┌──────────────────────┐    │
│                                 │   mailin.js          │    │
│                                 │   (SMTP Server)      │    │
│                                 └──────────┬───────────┘    │
│                                            │ message 事件    │
│                                            ▼                │
│                                 ┌──────────────────────┐    │
│                                 │   io.js              │    │
│                                 │   (Socket.IO 处理器)  │    │
│                                 │                      │    │
│                                 │  onlines Map:        │    │
│                                 │  shortId → socket    │    │
│                                 └──────────┬───────────┘    │
│                                            │ socket.emit    │
│                                            ▼                │
│   用户浏览器 ◄──(WebSocket)──►  ┌──────────────────────┐    │
│                                 │   Socket.IO Server   │    │
│                                 └──────────────────────┘    │
│                                                             │
│   用户浏览器 ◄──(HTTP)──────►  ┌──────────────────────┐    │
│                                 │   Express (app.js)   │    │
│                                 │   + routes/api.js    │    │
│                                 └──────────────────────┘    │
│                                                             │
│                                 ┌──────────────────────┐    │
│                                 │   dingtalk.js        │────► 钉钉
│                                 │   (Webhook 通知)      │    │
│                                 └──────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

---

## 11. 测试与 CI/CD

**无测试设施**。`package.json` 中的测试脚本为：

```json
"test": "echo \"Error: no test specified\" && exit 1"
```

- 无测试框架（无 Mocha、Jest、Jasmine 等）
- 仓库中无测试文件
- 无 CI/CD 配置（无 `.github/workflows`、`.travis.yml`、`Jenkinsfile` 等）
- 无代码检查配置（无 ESLint、Prettier 等）

---

## 12. 总结

Forsaken-Mail 是一个**轻量级、零数据库、自托管的临时邮箱服务**：

- **约 350 行后端 JavaScript**（4 个模块 + 1 个路由文件 + 1 个入口文件）
- **约 290 行前端 JavaScript**（1 个文件）
- **130 行 HTML**（1 个文件）
- **5 个 npm 依赖**（express、smtp-server、mailparser、socket.io、debug）
- **无构建步骤、无转译、无数据库、无测试**

架构简洁而高效：SMTP 服务器接收任何传入邮件，解析后通过 Socket.IO 即时推送到正在监听该地址的浏览器标签页。整个系统设计为以 Docker 容器运行，可选配 nginx 反向代理提供 Basic Auth 保护。
