Forsaken-Mail
==============

一个可自托管的临时邮箱服务，基于 Go + React 构建。

[English README](./README.md)

## 功能特性

- 在自有域名上接收随机或自定义地址的邮件
- 通过 WebSocket 实时推送新邮件
- 支持 OAuth2（GitHub/Google）或本地用户名/密码认证
- 邮箱地址白名单访问控制
- 钉钉 Webhook 新邮件通知
- 管理后台：审计日志、运行时配置、系统状态
- 国际化支持（中文 / 英文）
- SQLite 存储，自动清理过期邮件

## 快速开始

### DNS 配置

假设你要接收 `*@subdomain.domain.com` 的邮件，需添加两条 DNS 记录：

- **MX 记录**：`subdomain.domain.com  MX  10  mxsubdomain.domain.com`
- **A 记录**：`mxsubdomain.domain.com  A  <你的服务器IP>`

可使用 [SMTP 测试工具](http://mxtoolbox.com/diagnostic.aspx) 验证配置。

### Docker Compose（推荐）

```bash
cp .env.example .env   # 编辑配置
docker compose up -d --build
```

服务暴露端口：
- **25** — SMTP（接收邮件）
- **3000** — Web UI（可通过 `PORT` 环境变量修改）

浏览器访问 `http://localhost:3000`。

### 环境变量

完整列表见 `.env.example`。必填项：

| 变量 | 说明 |
|---|---|
| `MAIL_HOST` | 页面展示的邮箱域名（如 `mail.example.com`） |
| `SESSION_SECRET` | 会话加密密钥（`openssl rand -hex 32`） |
| `AUTH_MODE` | `oauth`（默认）或 `local` |
| `OAUTH_CLIENT_ID` / `OAUTH_CLIENT_SECRET` | `AUTH_MODE=oauth` 时必填 |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | `AUTH_MODE=local` 时必填（密码至少 8 位） |

可选：`DINGTALK_WEBHOOK_TOKEN`、`KEYWORD_BLACKLIST`、`SITE_TITLE`、`LOG_LEVEL`、`COOKIE_SECURE`。

### 数据持久化

SQLite 数据库存储在 `./data/` 目录（Docker volume 挂载）。首次部署：

```bash
mkdir -p data && sudo chown -R 100:101 data
```

容器以 `appuser`（UID 100）运行。

## 开发环境

前端需要 Node.js 24+。Go 后端通过 Docker 构建（本地无需安装 Go）。

```bash
# 终端 1：Go 后端（API 监听 :3000）
go run ./cmd/server

# 终端 2：Vite 开发服务器（前端 :5173，代理 API/WS/auth 到 :3000）
cd web && npm install && npm run dev
```

构建生产版本前端（输出到 `embed/`）：

```bash
cd web && npm run build
```

## 架构

```
外部邮件 ──(SMTP :25)──► go-smtp 服务器
  → enmime 解析 → mail.Router
    → SQLite 存储
    → WebSocket 推送到已订阅的客户端
    → 钉钉 Webhook（异步）

浏览器 ◄──(WebSocket /ws)──► ws.Hub
  → 客户端订阅 shortId
  → 服务端广播新邮件

浏览器 ◄──(HTTP /api/*)──► http.NewServeMux
  → REST API + React SPA（Go embed）
  → OAuth/本地认证中间件
```

**后端**（`internal/`）：config、smtp、mail（store/router/cleanup）、ws、auth（OAuth2 + 本地认证）、api、settings、audit、webhook、i18n、logger。

**前端**（`web/src/`）：React 19 + React Router 7 + Tailwind 4 + DaisyUI 5 + Vite 6。i18next 国际化。

## 许可证

GPL-2.0
