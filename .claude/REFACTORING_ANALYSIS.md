# Forsaken-Mail 技术选型分析与重构建议

## 一、当前技术选型逐项分析

---

### 1. SMTP 服务器：`smtp-server` + `mailparser`

**优点：**

- 纯 Node.js 实现，零外部依赖（无需 Postfix/Exim 等 MTA），部署简单
- `smtp-server` 是 `nodemailer` 作者 Andris Reinman 维护的库，质量可靠
- `mailparser` 的 `simpleParser` 能自动处理 MIME 多部分编码、附件、编码转换
- 代码量极小（`mailin.js` 仅 93 行），逻辑清晰

**缺点：**

- **无 TLS 支持** — 当前配置 `secure: false`，明文传输 SMTP，现代邮件服务器可能拒绝投递
- **无认证** — `disabledCommands: ['AUTH']` 意味着任何人都能向任意地址投递邮件，存在被滥发垃圾邮件的风险
- **单进程瓶颈** — SMTP 服务器运行在主进程中，邮件解析（`simpleParser`）是 CPU 密集操作，高并发时会阻塞事件循环
- **无速率限制** — 无连接数限制、无 IP 黑名单，容易被恶意刷流量
- **无持久化** — 邮件完全在内存中，服务重启即丢失，用户无法查看历史邮件

---

### 2. Web 框架：Express.js 4.x

**优点：**

- 生态成熟，中间件丰富
- 学习曲线低，代码直观
- 对于当前 4 个 API 端点的规模完全够用

**缺点：**

- **Express 4.x 已进入维护模式**，Express 5.x 已发布正式版
- 错误处理不完善 — 当前 `app.js` 第 26 行 `app.use(err => debug(err))` 只是打日志，没有向客户端返回错误响应
- 无请求体大小限制配置（虽然 `express.json()` 有默认 100KB 限制，但未显式声明）
- 无 CORS 配置，如果前后端分离部署会有问题

---

### 3. 实时通信：Socket.IO 4.x

**优点：**

- 自动降级（WebSocket → 轮询），兼容性好
- 内置房间、命名空间等高级功能
- 自动重连机制
- 服务端自动提供客户端 JS 文件（`/socket.io/socket.io.js`）

**缺点：**

- **协议开销大** — Socket.IO 不是原生 WebSocket，有自己的协议层，每条消息都有额外封装
- **与 HTTP 服务器耦合** — 当前 `bin/www` 中 Socket.IO 挂载在同一 HTTP server 上，无法独立扩缩
- **内存泄漏风险** — `io.js` 第 11 行 `let onlines = new Map()` 仅在 `disconnect` 时清理，但如果 socket 异常断开未触发 disconnect 事件，条目会永久残留
- **无认证中间件** — 连接时未验证身份，任何人都能连接并占用短 ID

---

### 4. 前端：jQuery 2.1.4 + Semantic UI 2.1.7（CDN 加载）

**优点：**

- 零构建步骤，直接浏览器加载
- jQuery 对于当前 DOM 操作规模足够
- Semantic UI 组件丰富，开箱即用

**缺点：**

- **jQuery 2.x 已停止维护**（当前最新 3.x），存在已知安全漏洞
- **Semantic UI 已停止维护**（最后更新 2018 年），社区已迁移至 Fomantic UI
- **CDN 依赖** — 使用 `lib.baomitu.com`（国内 CDN），如果 CDN 不可用则整个前端瘫痪
- **无模块化** — 所有前端逻辑在一个 287 行的 `app.js` 文件中，无组件拆分
- **XSS 风险** — `public/js/app.js` 第 53 行 `$('#mailcard .content:last').html(mail.html)` 直接将邮件 HTML 注入 DOM，恶意邮件可执行任意 JS
- **无响应式设计** — 依赖 Semantic UI 的栅格系统，但未做移动端适配优化

---

### 5. 数据存储：纯内存（Map）+ 浏览器 localStorage

**优点：**

- 零运维成本，无需管理数据库
- 读写性能极高（纯内存操作）
- 部署简单，无外部依赖

**缺点：**

- **零持久化** — 服务重启后所有在线用户丢失，邮件永久丢失
- **无法水平扩展** — 每个实例有独立的 `onlines` Map，多实例部署时邮件无法路由到正确实例
- **localStorage 有容量限制**（通常 5-10MB），且同域共享
- **无邮件保留策略** — 没有过期清理机制（虽然邮件不存储，但如果未来加存储需要考虑）

---

### 6. 部署：Docker + Nginx

**优点：**

- Docker Compose 一键部署，环境一致性好
- Nginx 提供 Basic Auth 保护和 WebSocket 代理
- 资源限制合理（app 300MB，nginx 30MB）

**缺点：**

- **Nginx Dockerfile 安装了 `apache2-utils`** 仅为了 `htpasswd`，镜像体积浪费
- **无健康检查** — docker-compose.yml 中未配置 `healthcheck`
- **无日志持久化** — 容器重启后日志丢失
- **SMTP 端口直接暴露** — 未通过 Nginx 代理，无法统一管理流量

---

### 7. 代码质量与工程化

**优点：**

- 代码量小（总计约 700 行），易于理解
- 模块划分合理（config/mailin/io/dingtalk 各司其职）

**缺点：**

- **零测试** — 无任何测试框架和测试用例
- **零 Lint** — 无 ESLint/Prettier 配置
- **无类型检查** — 纯 JS，无 TypeScript/JSDoc
- **`var` 与 `let/const` 混用** — `app.js` 用 `let`，前端用 `var`，风格不统一
- **无 CI/CD** — 无 GitHub Actions 或类似配置
- **无环境区分** — 无 development/production 环境配置

---

## 二、安全问题汇总

| 严重级别 | 问题 | 位置 |
|---|---|---|
| **高** | XSS：邮件 HTML 直接注入 DOM | `public/js/app.js:53` |
| **高** | SMTP 无认证，可被滥发 | `modules/mailin.js:69` |
| **中** | SMTP 无 TLS，明文传输 | `modules/mailin.js:68` |
| **中** | Socket.IO 无连接认证 | `modules/io.js:106` |
| **中** | 无速率限制（API 和 SMTP 均无） | 全局 |
| **低** | 旧版 jQuery/Socket.IO 已知漏洞 | `package.json` |

---

## 三、重构建议

### 方案 A：渐进式重构（推荐，保持架构不变）

在现有技术栈基础上逐步改进，适合当前项目规模：

#### 3A.1 优先级 P0 — 安全修复

```
1. XSS 修复
   - 引入 DOMPurify 对邮件 HTML 做消毒后再注入
   - npm install dompurify
   - 将 .html(mail.html) 改为 .html(DOMPurify.sanitize(mail.html))

2. SMTP 基础防护
   - 添加连接速率限制（smtp-server 的 size 参数限制邮件大小）
   - 添加 onConnect 回调，按 IP 限制连接频率
   - 可选：启用 STARTTLS（需配置证书）

3. Socket.IO 认证
   - 添加 middleware 验证连接来源
   - 或至少添加 CORS 白名单限制
```

#### 3A.2 优先级 P1 — 邮件持久化

```
方案选择：
  a) SQLite（推荐）— 零配置、单文件、适合单实例
     npm install better-sqlite3
     - 邮件存储到 SQLite 文件
     - 添加过期清理（如 24 小时自动删除）
     - API 增加历史邮件查询端点

  b) Redis — 适合未来多实例扩展
     - SET 邮件，EX 设置 TTL 自动过期
     - 但增加外部依赖

  c) 文件系统 — 最简单
     - 每封邮件存为 .eml 文件
     - 按日期目录组织
     - 但查询不便
```

#### 3A.3 优先级 P2 — 依赖升级

```
1. Express 4 → 5
   - 路由语法变化不大，主要是中间件兼容性检查

2. jQuery 2 → 3（或去除 jQuery）
   - 当前代码量小，可直接用原生 fetch + querySelector 替代

3. Semantic UI → Fomantic UI（社区维护分支）
   - 或迁移到更现代的轻量 CSS 框架（如 Pico CSS、Open Props）

4. CDN 依赖本地化
   - 将 CSS/JS 库打包到 public/ 目录，消除外部 CDN 依赖
```

#### 3A.4 优先级 P3 — 工程化

```
1. 添加 ESLint + Prettier
2. 添加基础测试（至少覆盖 config 模块和邮件路由逻辑）
3. 添加 GitHub Actions CI（lint + test）
4. 统一代码风格（全部使用 const/let，消除 var）
```

---

### 方案 B：全面现代化重构（适合长期维护）

如果计划长期维护或扩展功能，建议全面重构：

#### 技术栈替换方案

```
后端：
  Express 4    →  Hono 或 Fastify（性能更好，TypeScript 原生支持）
  Socket.IO    →  原生 WebSocket（ws 库，协议开销小）或 Socket.IO 5
  smtp-server   →  保持（已是最佳选择）
  CommonJS     →  ESM（import/export）
  JavaScript   →  TypeScript

前端：
  jQuery       →  去除（原生 API 已足够）
  Semantic UI  →  Tailwind CSS + 少量组件库
  单文件 HTML   →  Vite 构建（但仍保持轻量，不引入 React/Vue）

存储：
  纯内存       →  SQLite（better-sqlite3）或 LiteFS（分布式 SQLite）

部署：
  保持 Docker Compose，但添加：
  - 多阶段构建减小镜像体积
  - 健康检查端点
  - 结构化日志（pino/winston）
  - Prometheus metrics 端点
```

#### 架构改进

```
当前问题：
  邮件路由依赖内存 Map → 无法多实例

改进方案：
  1. 引入 Redis 作为 pub/sub 和 session 存储
     - SMTP 进程收到邮件 → PUBLISH 到 Redis channel
     - Web 进程 SUBSCRIBE → 推送到对应 socket
     - 支持水平扩展

  2. 或使用 SQLite + 轮询（适合单实例但需持久化）
     - 邮件存入 SQLite
     - 前端通过 SSE/轮询 查询新邮件
     - 解耦 SMTP 和 Web 进程
```

#### 项目结构重构

```
forsaken-mail/
├── src/
│   ├── server/
│   │   ├── index.ts              # 入口
│   │   ├── smtp/
│   │   │   ├── server.ts         # SMTP 服务器
│   │   │   └── parser.ts         # 邮件解析
│   │   ├── web/
│   │   │   ├── app.ts            # HTTP 服务
│   │   │   ├── routes/
│   │   │   │   ├── api.ts        # REST API
│   │   │   │   └── health.ts     # 健康检查
│   │   │   └── ws.ts             # WebSocket 处理
│   │   ├── services/
│   │   │   ├── mail-router.ts    # 邮件路由逻辑
│   │   │   ├── shortid.ts        # 短 ID 生成/管理
│   │   │   └── notification.ts   # 通知服务（可扩展）
│   │   ├── storage/
│   │   │   ├── index.ts          # 存储抽象层
│   │   │   ├── memory.ts         # 内存实现（开发用）
│   │   │   └── sqlite.ts         # SQLite 实现
│   │   └── config.ts             # 配置
│   └── client/
│       ├── index.html
│       ├── main.js               # 入口
│       ├── components/
│       │   ├── mail-list.js
│       │   ├── mail-detail.js
│       │   └── history.js
│       └── styles/
│           └── main.css
├── tests/
│   ├── smtp.test.ts
│   ├── router.test.ts
│   └── api.test.ts
├── tsconfig.json
├── Dockerfile
├── docker-compose.yml
└── package.json
```

---

## 四、重构优先级路线图

```
Phase 1（1-2 天）— 安全加固
  ├── [x] DOMPurify 消毒邮件 HTML
  ├── [x] SMTP 连接速率限制
  └── [x] Socket.IO CORS 限制

Phase 2（2-3 天）— 持久化 + 可靠性
  ├── [ ] SQLite 邮件存储
  ├── [ ] 过期自动清理
  ├── [ ] 历史邮件 API
  └── [ ] 内存 Map 定期清理（防泄漏）

Phase 3（1-2 天）— 依赖升级
  ├── [ ] Express 5
  ├── [ ] 去除 jQuery，用原生 API
  ├── [ ] CSS 框架替换
  └── [ ] CDN 依赖本地化

Phase 4（1-2 天）— 工程化
  ├── [ ] ESLint + Prettier
  ├── [ ] 基础测试
  ├── [ ] CI/CD
  └── [ ] 结构化日志

Phase 5（可选）— 现代化
  ├── [ ] TypeScript 迁移
  ├── [ ] ESM 模块化
  └── [ ] 前端构建工具
```

---

## 五、总结

| 维度 | 当前状态 | 重构目标 |
|---|---|---|
| **安全性** | 多处 XSS 和未授权访问风险 | DOMPurify + 认证 + 速率限制 |
| **持久化** | 零持久化，重启丢数据 | SQLite 存储邮件 + 过期清理 |
| **可扩展性** | 单实例，内存 Map 路由 | 抽象存储层，支持 Redis 扩展 |
| **依赖健康** | 多个已停维库 | 全部升级到活跃维护版本 |
| **代码质量** | 无测试、无 Lint、风格不一 | TypeScript + ESLint + 测试覆盖 |
| **部署** | 基础 Docker Compose | 健康检查 + 日志 + 监控 |

**核心建议**：当前项目代码量小、架构简单，**方案 A（渐进式重构）** 是最务实的选择。优先修复安全问题（P0），再添加 SQLite 持久化（P1），最后做依赖升级和工程化改造（P2/P3）。不建议过度工程化 — 这个项目的核心价值就是"轻量"，保持简单是正确的。
