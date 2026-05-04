# 双模式认证方案

## 目标

支持两种登录模式，通过环境变量切换，同时只启用一种：
- **OAuth 模式**（现有）：GitHub / Google OAuth 登录
- **账号模式**（新增）：用户名 + 密码登录

## 当前架构

```
认证流程：
LoginPage → OAuth 按钮 → /auth/{provider}/login → OAuth 回调 → Session Cookie → 受保护路由

关键文件：
- internal/config/config.go      — 环境变量加载
- internal/auth/oauth.go         — OAuth Provider 接口
- internal/auth/session.go       — Session 加密/解密（AES-GCM）
- internal/auth/middleware.go    — 认证中间件
- internal/api/auth.go           — OAuth 路由 handler
- internal/api/router.go         — 路由注册
- web/src/pages/LoginPage.jsx    — 登录页面
- web/src/App.jsx                — AuthContext
```

## 设计方案

### 1. 环境变量

新增环境变量 `AUTH_MODE`，可选值：
- `oauth`（默认）— OAuth 登录，需要 OAUTH_PROVIDER、OAUTH_CLIENT_ID、OAUTH_CLIENT_SECRET
- `local` — 本地账号登录，需要 ADMIN_USERNAME、ADMIN_PASSWORD

```env
# 认证模式：oauth 或 local
AUTH_MODE=oauth

# OAuth 模式配置
OAUTH_PROVIDER=github
OAUTH_CLIENT_ID=***
OAUTH_CLIENT_SECRET=***

# 本地账号模式配置
ADMIN_USERNAME=admin
ADMIN_PASSWORD=***

# 通用
SESSION_SECRET=***
```

**验证逻辑**：
- `AUTH_MODE=oauth` 时，必须配置 OAUTH_CLIENT_ID 和 OAUTH_CLIENT_SECRET
- `AUTH_MODE=local` 时，必须配置 ADMIN_USERNAME 和 ADMIN_PASSWORD
- 启动时校验，缺失必要配置则报错退出

### 2. 后端变更

#### 2.1 config/config.go

```go
type Config struct {
    // ... 现有字段 ...
    AuthMode       string  // "oauth" 或 "local"
    AdminUsername  string
    AdminPassword  string
}
```

加载逻辑：
- 读取 `AUTH_MODE` 环境变量，默认 `oauth`
- 如果 `AUTH_MODE=local`，读取 `ADMIN_USERNAME` 和 `ADMIN_PASSWORD`
- 根据模式进行不同的 validate 逻辑

#### 2.2 auth/local.go（新建）

```go
package auth

// LocalAuth 本地账号认证
type LocalAuth struct {
    username string
    password string
}

func NewLocalAuth(username, password string) *LocalAuth

// Verify 验证用户名密码
func (la *LocalAuth) Verify(username, password string) bool
```

密码存储：环境变量明文配置（个人临时邮箱场景，不需要 bcrypt）

#### 2.3 api/auth.go

新增 handler：
```go
// POST /auth/login — 本地账号登录
func (rt *Router) handleLocalLogin(w http.ResponseWriter, r *http.Request)

// POST /auth/logout — 登出（已有，复用）
```

`handleLocalLogin` 逻辑：
1. 解析 JSON body `{ "username": "...", "password": "..." }`
2. 调用 `LocalAuth.Verify` 验证
3. 成功则创建 Session Cookie（复用现有 SessionManager）
4. 返回 `{ "status": "ok" }`
5. 失败返回 401

#### 2.4 api/config.go

`/api/config` 响应增加 `auth_mode` 字段：
```json
{
  "host": "tmail.o2.eoekun.top",
  "site_title": "TMail",
  "auth_mode": "local"
}
```

前端根据 `auth_mode` 决定显示 OAuth 按钮还是登录表单。

#### 2.5 api/router.go

路由注册变更：
```go
// 根据 cfg.AuthMode 注册不同的认证路由
if rt.cfg.AuthMode == "local" {
    mux.HandleFunc("/auth/login", rt.handleLocalLogin)
} else {
    mux.HandleFunc("/auth/", rt.routeAuth)  // OAuth 路由
}
mux.HandleFunc("/auth/logout", rt.handleLogout)  // 登出两种模式通用
```

#### 2.6 middleware.go

中间件不需要修改 — 它只检查 Session Cookie 是否有效，不关心登录方式。

### 3. 前端变更

#### 3.1 LoginPage.jsx

根据 `config.auth_mode` 切换显示：

**OAuth 模式**（现有）：
- 显示 GitHub / Google 按钮

**账号模式**（新增）：
- 显示用户名 + 密码表单
- 登录按钮调用 `POST /auth/login`
- 登录失败显示错误提示

```jsx
export default function LoginPage() {
  const { config } = useAuth()
  
  if (config?.auth_mode === 'local') {
    return <LocalLoginForm />
  }
  
  return <OAuthLoginForm />
}
```

#### 3.2 App.jsx

`/api/config` 响应中读取 `auth_mode`，传递给 LoginPage。

### 4. 安全考虑

- `ADMIN_PASSWORD` 至少 8 位，启动时校验
- 登录接口增加速率限制（复用 SMTP 的 RateLimiter 思路，或简单计数器）
- 登录失败记录审计日志
- Session Cookie 设置 HttpOnly、Secure（HTTPS）、SameSite=Lax

### 5. 数据库变更

无。账号信息通过环境变量配置，不存数据库。

## 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/config/config.go` | 修改 | 增加 AuthMode、AdminUsername、AdminPassword |
| `internal/auth/local.go` | 新建 | 本地账号认证逻辑 |
| `internal/api/auth.go` | 修改 | 增加 handleLocalLogin |
| `internal/api/config.go` | 修改 | 响应增加 auth_mode |
| `internal/api/router.go` | 修改 | 根据模式注册不同路由 |
| `web/src/pages/LoginPage.jsx` | 修改 | 双模式登录 UI |
| `web/src/App.jsx` | 修改 | 传递 auth_mode |
| `web/src/locales/zh.json` | 修改 | 增加登录表单翻译 |
| `web/src/locales/en.json` | 修改 | 增加登录表单翻译 |
| `.env.example` | 修改 | 增加 AUTH_MODE、ADMIN_USERNAME、ADMIN_PASSWORD |

## 验证

```bash
# OAuth 模式（现有）
AUTH_MODE=oauth OAUTH_CLIENT_ID=xxx OAUTH_CLIENT_SECRET=yyy ...
# 登录页显示 GitHub/Google 按钮

# 本地账号模式
AUTH_MODE=local ADMIN_USERNAME=admin ADMIN_PASSWORD=secret123 ...
# 登录页显示用户名密码表单
```
