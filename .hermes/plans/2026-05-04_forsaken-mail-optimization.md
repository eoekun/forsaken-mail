# Forsaken-Mail 功能优化计划

## 目标

对 forsaken-mail 临时邮件服务器进行功能优化，提升安全性、可用性和开发者体验。

## 当前状态

- **技术栈**: Go 后端 + React 前端，SQLite 存储，Docker 部署
- **部署地址**: `ubuntu@130.162.245.79`，域名 `tmail.o2.eoekun.top`
- **代码分支**: `refactor/go-react`

## 实现规范

### 开发工具
- **全程使用 Claude Code** 实现，模型为 `mimo-v2.5-pro`
- **不指定 `--model` 参数**，让 Claude Code 使用 settings.json 中的模型映射
- 本地不进行 Go 编译，所有构建/验证通过远程服务器 Docker 完成

### 验证流程
```
本地修改代码 → git commit → git push → SSH 到服务器 → docker compose build && docker compose up -d → 验证功能
```

### 数据库策略
- **全新覆盖逻辑**：schema 变更时直接修改 CREATE TABLE 语句，不做 ALTER TABLE 迁移
- **历史数据不保留**：服务器上的 SQLite 数据库会在重建时清空
- **⚠️ 保留 .env 配置**：服务器上的 `.env` 文件包含 OAuth 密钥等配置，部署时必须保留

### Git 历史修正
- `doc/` 目录迁移到 `.claude/` 后，需要使用 `git filter-branch` 或 `git filter-repo` 从历史中移除
- 采用 `push -f` 覆盖远程仓库

---

## 需求清单

### Phase 0: 预处理

#### 0.1 修复 site_title 显示问题

**问题**: 页面中展示的 "Forsaken Mail" 没有正确使用 admin 设置的 `site_title` 值

**排查方向**:
- 检查前端 `Navbar.jsx`、`index.html` 中的 title 是否硬编码
- 检查 `App.jsx` 中的 `AuthContext` 是否正确传递 config.site_title
- 检查后端 `/api/config` 接口是否返回正确的 site_title

**预期修复点**:
- `web/index.html` — `<title>` 标签
- `web/src/components/Navbar.jsx` — 导航栏标题
- `web/src/pages/LoginPage.jsx` — 登录页面标题
- 浏览器标签页 title 动态更新

---

#### 0.2 doc/ 迁移到 .claude/ 并清理 Git 历史

**目标**: 将 `doc/` 目录下的文档移动到 `.claude/` 目录，从 git 历史中彻底移除

**步骤**:
1. 移动文件：
   ```
   doc/ARCHITECTURE.md → .claude/ARCHITECTURE.md
   doc/LANGUAGE_COMPARISON.md → .claude/LANGUAGE_COMPARISON.md
   doc/REFACTORING_ANALYSIS.md → .claude/REFACTORING_ANALYSIS.md
   doc/REFACTORING_PLAN.md → .claude/REFACTORING_PLAN.md
   ```
2. 删除 `doc/` 目录
3. 提交变更
4. 使用 `git filter-repo` 从历史中移除 `doc/` 路径：
   ```bash
   git filter-repo --path doc/ --invert-paths
   ```
5. `git push --force origin refactor/go-react`

**注意**: force push 会覆盖远程历史，确保本地是最新状态

---

#### 0.3 清理旧 Node.js 代码

**删除文件**:
- `app.js`
- `bin/` 目录
- `modules/` 目录
- `routes/` 目录
- `public/` 目录（旧前端）
- 根目录 `package.json`

---

### Phase 1: 基础优化

#### 1.1 RateLimiter 内存清理

**文件**: `internal/smtp/ratelimit.go`

**方案**: 为 limiter 条目增加 `lastAccess` 时间戳，定期清理过期条目
- 每 5 分钟检查一次
- 清理超过 30 分钟未访问的条目
- 超过 10000 条时淘汰最旧的

---

#### 1.2 禁用发件能力

**目标**: 防止被利用发送垃圾邮件

**实现**: 在 `session.MailFrom()` 阶段拒绝所有发件请求
```go
func (s *session) MailFrom(from string, opts *smtp.MailOptions) error {
    return fmt.Errorf("550 Sending mail is not allowed")
}
```

**审计**: 记录被拒绝的发件尝试

---

### Phase 2: 核心功能

#### 2.1 邮件 API

**新增路由**:
```
GET    /api/emails/{shortId}             # 邮件列表
GET    /api/emails/{shortId}/{mailId}    # 邮件详情
DELETE /api/emails/{shortId}             # 清空邮件
DELETE /api/emails/{shortId}/{mailId}    # 删除单封
```

**实现**: `internal/api/emails.go`，复用 OAuth 认证

---

#### 2.2 邮件已读状态

**数据库**: `CREATE TABLE` 中直接加入 `is_read INTEGER NOT NULL DEFAULT 0` 字段

**后端**:
- `internal/mail/store.go` — `MarkAsRead(id int64)` 方法
- `internal/api/mail.go` — `PUT /api/mails/{id}/read` 接口

**前端**:
- `MailList.jsx` — 已读/未读样式区分
- `MailDetail.jsx` — 查看时自动标记已读

---

### Phase 3: 智能功能

#### 3.1 验证码自动提取

**新增**: `internal/mail/extract.go`

**提取规则**:
- 验证码：正则匹配 4-8 位数字（verification code、验证码等）
- 链接：提取 http/https URL

**数据库**: `extracted_codes`、`extracted_links` TEXT 字段（JSON 数组）

**前端**: 验证码高亮卡片，点击复制

---

#### 3.2 多邮箱并行监听

**WebSocket 扩展**:
```json
{"type": "subscribe", "short_id": "abc"}
{"type": "unsubscribe", "short_id": "abc"}
{"type": "mail", "short_id": "abc", "data": {...}}
```

**前端**: `MailboxTabs.jsx` 标签栏 + 未读角标

**后端**: `hub.go` Client 支持多 shortId 订阅

---

## 实现顺序

```
Phase 0（预处理）
  ├── #0.1 修复 site_title 显示
  ├── #0.2 doc/ 迁移 + git 历史清理
  └── #0.3 清理旧 Node.js 代码

Phase 1（基础优化）
  ├── #1.1 RateLimiter 内存清理
  └── #1.2 禁用发件能力

Phase 2（核心功能）
  ├── #2.1 邮件 API
  └── #2.2 邮件已读状态

Phase 3（智能功能）
  ├── #3.1 验证码自动提取
  └── #3.2 多邮箱并行监听
```

## 部署检查清单

每次部署前：
- [ ] 确认服务器 `.env` 文件存在且内容正确
- [ ] `docker compose build` 成功
- [ ] `docker compose up -d` 启动正常
- [ ] 访问 `https://tmail.o2.eoekun.top` 验证页面加载
- [ ] GitHub OAuth 登录正常
- [ ] 发送测试邮件验证收件功能

## 风险与注意事项

1. **force push**: #0.2 会覆盖远程历史，确保协作者知晓
2. **数据库清空**: 全新 schema 意味着服务器上现有邮件数据会丢失
3. **.env 保留**: 部署时必须确保 `.env` 不被覆盖
4. **site_title**: 修复后需验证 admin 设置的值正确显示在所有页面
5. **WebSocket 协议**: #3.2 扩展协议需前后端同步更新
