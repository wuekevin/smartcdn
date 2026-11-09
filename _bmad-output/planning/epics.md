---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - planning/prds/prd-SmartCDN-2026-11-09/prd.md
  - planning/architecture/architecture-SmartCDN-2026-11-09/ARCHITECTURE-SPINE.md
  - planning/ux-designs/ux-SmartCDN-2026-11-09/DESIGN.md
  - planning/ux-designs/ux-SmartCDN-2026-11-09/EXPERIENCE.md
  - PROJECT_REQUIREMENTS.md
---

# SmartCDN - Epic Breakdown

## Overview

SmartCDN 是可私有化部署的商业级 DNS/CDN/SSL 管理平台，使用 Go + React 技术栈。本文档将 PRD 全部 58 个功能需求（FR-1~FR-58）和架构决策（AD-1~AD-20）分解为 17 个 Epic 和 70+ 个可独立开发、可测试、可验收的 Story。

Epic 执行顺序由依赖关系决定：Epic 1-3 建立基础，Epic 4 完成 Provider 实现，Epic 5-7 实现完整 DNS 能力，Epic 8-9 实现 CDN，Epic 10-11 实现三网调度，Epic 12-13 实现 SSL 自动化，Epic 14 实现监控，Epic 15-17 完成系统管理和部署。

---

## Requirements Inventory

### Functional Requirements

FR-1: 安装向导环境检测（Go/MySQL/Redis/权限/端口/时钟）逐项 PASS/FAIL
FR-2: 安装向导初始化流程（DB配置→Migration→管理员账号→加密密钥→install.lock）
FR-3: 应用加密密钥管理（SMARTCDN_ENCRYPT_KEY 环境变量，AES-256-GCM）
FR-4: 结构化日志（JSON 格式，无明文凭证）
FR-5: 用户认证（用户名+密码 → JWT，连续失败锁定）
FR-6: RBAC 权限控制（预置角色+自定义角色，API endpoint 粒度）
FR-7: 审计日志（不可变，操作前后状态摘要）
FR-8: CSRF 和 XSS 防护
FR-9: Provider 统一接口（12+1 方法，UnsupportedCapabilityError）
FR-10: Provider 能力检测与缓存
FR-11: Provider API 限速控制（令牌桶）
FR-12: 凭证安全管理（加密存储，脱敏展示，轮换）
FR-13: 阿里云 DNS 适配器
FR-14: 腾讯云 DNSPod 适配器
FR-15: Cloudflare DNS 适配器
FR-16: 多账号管理（同 Provider 多账号，启用/禁用）
FR-17: 域名 DNS 配置（权威/备用 Provider、TTL、发布策略）
FR-18: DNS 记录 CRUD（A/AAAA/CNAME/TXT/MX/NS/CAA/SRV，批量操作）
FR-19: DNS 历史快照与回滚
FR-20: DNS 迁移工作流（模式 B，7 步，人工授权 NS 切换）
FR-21: DNS 记录发布引擎（13 步安全发布流程）
FR-22: 分布式锁保护 DNS 变更
FR-23: Multi-DNS 配置（主+备权威 DNS，自动同步）
FR-24: NS 切换授权检查（人工确认，不自动执行）
FR-25: DNS 迁移完整性校验
FR-26: 迁移回滚与异常恢复
FR-27: CDN Provider 框架（CDN Provider 接口）
FR-28: Cloudflare Zone 管理
FR-29: Cloudflare for SaaS Custom Hostname 管理
FR-30: Cloudflare SaaS DCV 验证（DNS-01 TXT 自动/手动）
FR-31: 阿里云 CDN 管理接入
FR-32: CDN 缓存规则与缓存清理
FR-33: 自有节点管理（IP、ISP 标识、HTTPS 检查）
FR-34: Node Agent 安装与管理（mTLS，白名单任务）
FR-35: 节点资源池管理
FR-36: CDN 缓存规则（节点级和全局级）
FR-37: Probe Agent 二进制部署（mTLS，ISP 标识验证）
FR-38: 探测任务与上报（延迟/TTFB/成功率，HTTP-based）
FR-39: IP 候选池管理（Candidate/Eligible/Published/Standby 状态）
FR-40: 三网调度评分算法（延迟40%+TTFB40%+成功率20%）
FR-41: 自动调度执行与熔断（30min 内≥3次切换触发熔断）
FR-42: Node Agent 状态上报与任务执行（CPU/内存/HTTPS 状态）
FR-43: ACME 账号管理（Let's Encrypt / ZeroSSL）
FR-44: SSL 证书申请（DNS-01 / HTTP-01 Challenge）
FR-45: 证书自动续期（到期前 N 天触发工作流）
FR-46: 证书工作流 DAG 引擎（节点执行，并行分支，条件节点）
FR-47: 工作流 DAG 可视化编辑器
FR-48: 证书部署到宝塔 Nginx（备份→写入→nginx -t→reload→验证）
FR-49: 证书部署到阿里云 CDN（调用 CDN 证书更新 API）
FR-50: 证书到期监控与告警
FR-51: 证书部署目标管理（BaotaNginx/NativeNginx/SSH/AliyunCDN）
FR-52: 证书执行日志（节点级日志，实时 SSE）
FR-53: 实时监控仪表盘（三网趋势/DNS 健康/CDN 健康/HTTPS 状态）
FR-54: 告警规则管理（触发条件/等级/静默）
FR-55: 告警记录与确认（CRITICAL/WARNING/INFO，确认/静默）
FR-56: 通知渠道管理（邮件/Webhook）
FR-57: SSE/WebSocket 实时推送（全局事件流、任务日志流）
FR-58: 部署打包（宝塔 Go 项目/systemd/Docker/Docker Compose）

### NonFunctional Requirements

NFR-1: 安全 — AES-256-GCM 加密所有 Provider 凭证和 SSL 私钥（AD-8）
NFR-2: 安全 — RBAC 粒度至 API Endpoint，权限变更实时生效（AD-9）
NFR-3: 安全 — 越权返回 403 不返回 404（AD-9）
NFR-4: 安全 — CSRF Token 验证，CSP/X-Frame-Options 响应头（FR-8）
NFR-5: 安全 — SSRF 防护，拒绝 RFC1918/localhost/0.0.0.0（AD-19）
NFR-6: 安全 — 审计日志不可删除（AD-15）
NFR-7: 性能 — API P95 响应时间 < 500ms（非 Provider 调用链路）
NFR-8: 性能 — 列表接口 cursor-based 分页（AD-7）
NFR-9: 可靠性 — DNS 变更幂等（AD-5）
NFR-10: 可靠性 — 分布式锁防止 DNS 并发写（AD-6）
NFR-11: 可观测性 — JSON 结构化日志，zap，无明文凭证（FR-4）
NFR-12: 合规 — ICP 备案边界，API 层强制拦截（AD-20）
NFR-13: 部署 — 单 binary 嵌入前端，宝塔 Go 项目兼容（AD-16）
NFR-14: 探针真实性 — Probe Agent 必须部署于真实 ISP 网络（AD-10）

### Additional Requirements（来自 Architecture Spine）

- 六边形架构：domain/application/infra/handler 四层，application 层只依赖 domain port 接口（AD-1）
- Provider Registry：所有适配器通过 Registry 注入，不直接 import（AD-1）
- 编译时依赖注入：使用 google/wire（Stack）
- 前端 embed.FS 嵌入：Vite build 产物 embed 进 Go binary（AD-16）
- 数据库迁移：migrations/ 目录，goose 管理，不手动建表
- install.lock 机制：安装完成后写入，启动时检查（AD-14）
- mTLS Agent 通信：独立路由组 /api/agent/v1/，CA 签发客户端证书（AD-10 AD-11）
- 状态机保护：publishing_jobs.status 只由 PublishingJobService 写（AD-13）
- 审计日志中间件：Gin AuditLogger middleware，异步写入（AD-15）
- cursor-based 分页：所有列表 API（AD-7，NFR-8）

### UX Design Requirements

UX-DR1: 安装向导（/setup）：9 步向导，居中布局，无侧导航，步骤条，环境检测逐项展示 PASS/FAIL
UX-DR2: 登录页（/login）：左侧品牌区域+右侧表单卡片，锁定提示，密码可见性切换
UX-DR3: 全局导航：左侧 200px 展开/48px 折叠，TopBar 48px，Cmd+K 全局搜索
UX-DR4: 仪表盘：4 统计卡片+DNS 状态+CDN 节点环形图+三网趋势折线图+SSL 到期+最近告警，SSE 实时刷新
UX-DR5: DNS 账号 Drawer：Provider 选择+凭证输入（password 类型）+验证连接按钮+能力报告，验证通过后才能保存
UX-DR6: DNS 迁移 7 步向导：Step 5 人工 NS 切换授权检查点，不自动执行
UX-DR7: Cloudflare SaaS 4 步向导：页面顶部 Banner 说明 CNAME-only 约束
UX-DR8: 高危操作二次确认（文本输入验证）：NS 切换/批量 DNS 删除/删除域名绑定/证书强制替换
UX-DR9: 证书工作流 DAG 编辑器：全页路由（/ssl/workflows/:id/edit），7 种节点类型视觉区分
UX-DR10: IP 池页面：每 IP 展示各探针延迟/TTFB/成功率热图，状态标签（Published/Eligible/Standby/Offline）
UX-DR11: 移动端（<768px）：汉堡导航+卡片列表，高危操作隐藏
UX-DR12: 深色/浅色双主题：Arco Design ConfigProvider theme 切换，localStorage 持久化
UX-DR13: 状态标签颜色+文字双载体（不仅靠颜色区分）
UX-DR14: 表格 hover 展示行操作按钮（不占列宽）
UX-DR15: 骨架屏加载（Skeleton Table/Card），< 500ms 不显示全页 Spin

---

### FR Coverage Map

FR-1 ~ FR-4: Epic 1 — 项目基础架构与安装向导
FR-5 ~ FR-8: Epic 2 — 用户认证与 RBAC
FR-9 ~ FR-12: Epic 3 — DNS Provider 插件框架
FR-13 ~ FR-15: Epic 4 — DNS Provider 适配器（阿里云/DNSPod/Cloudflare）
FR-16 ~ FR-19: Epic 5 — 多域名与 DNS 记录管理
FR-20 ~ FR-22: Epic 6 — DNS 安全发布引擎（13 步）
FR-23 ~ FR-26: Epic 7 — DNS 迁移与 Multi-DNS
FR-27 ~ FR-32: Epic 8 — CDN Provider 框架与 Cloudflare SaaS
FR-33 ~ FR-36: Epic 9 — 自有节点管理与缓存
FR-37 ~ FR-41: Epic 10 — 三网智能调度与 Probe Agent
FR-42: Epic 11 — Node Agent 与节点状态
FR-43 ~ FR-45: Epic 12 — SSL 证书管理与 ACME
FR-46 ~ FR-52: Epic 13 — 证书工作流 DAG 引擎
FR-53 ~ FR-56: Epic 14 — 监控、告警与通知
FR-57: Epic 15 — 实时推送（SSE / WebSocket）
FR-58: Epic 16 — 部署打包
UX-DR1~UX-DR15: Epic 17 — UX 完整性与前端基础

---

## Epic List

### Epic 1: 项目基础架构与安装向导
运营者可以完成 SmartCDN 首次安装配置，系统通过环境检测、数据库初始化和管理员创建后进入可用状态。
**FRs covered:** FR-1, FR-2, FR-3, FR-4

### Epic 2: 用户认证与 RBAC
管理员可以安全登录系统，角色权限控制确保每个用户只能操作其授权范围内的资源，所有变更写入不可变审计日志。
**FRs covered:** FR-5, FR-6, FR-7, FR-8

### Epic 3: DNS Provider 插件框架
开发者可以通过统一的 Provider Port 接口添加新的 DNS 服务商，不修改任何调度层代码。
**FRs covered:** FR-9, FR-10, FR-11, FR-12

### Epic 4: DNS Provider 适配器（阿里云 / DNSPod / Cloudflare）
运营者可以添加阿里云 DNS、DNSPod、Cloudflare DNS 账号，系统准确报告每个账号的套餐能力限制。
**FRs covered:** FR-13, FR-14, FR-15

### Epic 5: 多域名与 DNS 记录管理
运营者可以管理多个业务域名的 DNS 配置，查看、新增、编辑、删除 DNS 记录，并通过历史快照随时回滚。
**FRs covered:** FR-16, FR-17, FR-18, FR-19

### Epic 6: DNS 安全发布引擎（13 步）
系统对所有 DNS 写操作强制执行 13 步安全发布流程，含分布式锁、幂等检查、权威验证和失败自动回滚。
**FRs covered:** FR-20（部分）, FR-21, FR-22

### Epic 7: DNS 迁移与 Multi-DNS
运营者可以安全地将域名从一个 DNS Provider 迁移到另一个（含记录预同步和 NS 切换授权检查点），并可配置主+备双权威 DNS。
**FRs covered:** FR-20（完整）, FR-23, FR-24, FR-25, FR-26

### Epic 8: CDN Provider 框架与 Cloudflare SaaS / 阿里云 CDN
运营者可以管理 Cloudflare Zone、配置 Cloudflare for SaaS Custom Hostname（含 DCV 自动化）以及接入阿里云 CDN。
**FRs covered:** FR-27, FR-28, FR-29, FR-30, FR-31, FR-32

### Epic 9: 自有节点管理、Node Agent 与缓存
运营者可以添加自有家宽/服务器节点，安装 Node Agent，配置节点缓存规则，查看节点实时状态。
**FRs covered:** FR-33, FR-34, FR-35, FR-36

### Epic 10: 三网智能调度与 Probe Agent
系统通过真实 ISP 网络部署的 Probe Agent 自动收集三网性能数据，调度器定期计算最优 IP 并发布到 DNS，含熔断保护。
**FRs covered:** FR-37, FR-38, FR-39, FR-40, FR-41

### Epic 11: Node Agent 状态上报与管理
Node Agent 定期向管理中心上报节点 CPU/内存/HTTPS 状态，管理中心可下发授权任务（reload/证书部署/HTTPS 检查）。
**FRs covered:** FR-42

### Epic 12: SSL 证书管理与 ACME 自动化
运营者可以管理 ACME 账号，申请/导入/管理 SSL 证书，配置自动续期，系统在证书到期前自动完成 DNS-01 验证和证书获取。
**FRs covered:** FR-43, FR-44, FR-45, FR-50, FR-51

### Epic 13: 证书工作流 DAG 引擎与部署
系统通过可视化 DAG 工作流编排证书的申请、验证、部署全流程，支持并行分支部署到宝塔 Nginx 和阿里云 CDN。
**FRs covered:** FR-46, FR-47, FR-48, FR-49, FR-52

### Epic 14: 监控、告警与通知
运营者可以在统一仪表盘查看三网性能趋势、DNS/CDN 健康状态，配置告警规则，通过邮件/Webhook 接收通知。
**FRs covered:** FR-53, FR-54, FR-55, FR-56

### Epic 15: 实时推送（SSE / WebSocket）
管理控制台通过 SSE 实时推送任务状态、告警和指标更新，工作流执行进度通过 WebSocket 推送。
**FRs covered:** FR-57

### Epic 16: 部署打包（宝塔 Go / systemd / Docker）
SmartCDN binary 可通过宝塔 Go 项目管理器部署（P0），或通过 systemd 单元文件、Docker Compose 部署（P1）。
**FRs covered:** FR-58

### Epic 17: 前端基础与 UX 完整性
前端 React SPA 实现完整设计规范：双主题切换、全局搜索、响应式布局、骨架屏、高危操作二次确认、访问控制可见性。
**FRs covered:** UX-DR1~UX-DR15（配合所有 Epic 的前端 Story）

---

## Epic 1: 项目基础架构与安装向导

**Epic 目标：** 运营者完成 SmartCDN 首次安装，系统通过环境检测和初始化后进入可用状态。建立项目目录结构、数据库 schema、加密基础设施和日志体系。

### Story 1.1: 项目脚手架与目录结构

作为开发者，
我想要建立 SmartCDN 项目的完整目录结构（cmd/internal/frontend/migrations/deploy），
以便后续所有 Epic 在统一的代码组织规范下开发。

**Acceptance Criteria:**

**Given** 一个空的 Git 仓库
**When** 执行初始化脚本
**Then** 目录结构符合 Architecture Spine AD-1 六边形架构：cmd/{smartcdn,probe-agent,node-agent}，internal/{domain/port,application,infra,handler}，frontend/，migrations/，deploy/{docker,systemd,baota}
**And** go.mod 中 module 名为 `github.com/wuekevin/smartcdn`，Go version 1.23
**And** `wire` 编译时依赖注入框架已配置
**And** `.gitignore` 覆盖：.env，*.lock（除 install.lock），dist/，vendor/

---

### Story 1.2: 数据库迁移基础（goose + 核心表）

作为开发者，
我想要通过 goose 管理数据库迁移，
以便每次发布能以可重复的方式创建或升级数据库 schema。

**Acceptance Criteria:**

**Given** 配置好的 MySQL 8.0 实例
**When** 运行 `goose up`
**Then** 创建 `users`、`roles`、`user_roles`、`role_permissions`、`audit_logs`、`dns_providers`、`dns_accounts` 基础表（字段见 Architecture Spine ERD）
**And** 所有业务表含 `deleted_at DATETIME(3) NULL`（audit_logs 除外）
**And** `audit_logs` 无 deleted_at 列，不可软删除
**And** 执行 `goose down` 可完整回滚，不报错

---

### Story 1.3: AES-256-GCM 加密基础设施

作为系统，
我想要从环境变量 `SMARTCDN_ENCRYPT_KEY` 读取加密密钥并提供加解密 API，
以便所有模块安全存储 Provider 凭证和 SSL 私钥（AD-8）。

**Acceptance Criteria:**

**Given** 环境变量 `SMARTCDN_ENCRYPT_KEY` 未设置
**When** 应用启动
**Then** 启动失败，输出明确错误：`SMARTCDN_ENCRYPT_KEY is required (32-byte hex)`，退出码非零
**And** 密钥不出现在任何日志输出中

**Given** `SMARTCDN_ENCRYPT_KEY` 已设置（64 字符 hex）
**When** 调用 `crypto.Encrypt(plaintext)` 和 `crypto.Decrypt(ciphertext)`
**Then** 加密产物格式 = `nonce(12 bytes) || ciphertext`，Base64 编码
**And** 每次加密产生不同密文（随机 nonce）
**And** 解密结果与明文完全一致

**Given** 密文格式错误或密钥不匹配
**When** 调用 `crypto.Decrypt`
**Then** 返回 `ErrDecryptionFailed`，不 panic

---

### Story 1.4: 结构化日志（zap）

作为运营者，
我想要所有服务端日志以 JSON 格式输出，
以便日志收集系统能解析告警并排查问题（FR-4）。

**Acceptance Criteria:**

**Given** 应用以生产模式启动
**When** 任意请求处理中产生日志
**Then** 日志为单行 JSON，包含：`timestamp`（RFC3339）、`level`、`module`、`request_id`、`message`
**And** 日志中不包含任何匹配 `-----BEGIN` 或长度 > 32 字节 hex/base64 的字符串
**And** 日志级别可通过环境变量 `LOG_LEVEL` 调整（debug/info/warn/error）

---

### Story 1.5: 安装向导后端（环境检测 + 初始化 API）

作为首次运营者，
我想要通过 API 完成环境检测和系统初始化，
以便安装向导前端可以引导我完成配置（FR-1，FR-2）。

**Acceptance Criteria:**

**Given** `install.lock` 不存在
**When** `GET /api/v1/setup/check`
**Then** 返回 JSON 检测结果：`{mysql: {status:"ok"}, redis:{status:"ok"}, ...}`，每项含 status 和 message
**And** MySQL 连接失败时返回具体错误原因，不暴露密码

**Given** 安装步骤全部完成
**When** `POST /api/v1/setup/complete`（含管理员账号、加密密钥确认）
**Then** 执行 Database Migration → 创建管理员账号（密码 bcrypt hash）→ 写入 `install.lock`
**And** 响应 200，含 `{installed: true}`

**Given** `install.lock` 已存在
**When** 任意 `POST /api/v1/setup/*`
**Then** 返回 403 `{code:"SETUP_LOCKED"}`

---

### Story 1.6: 安装向导前端（9 步向导 UI）

作为首次运营者，
我想要通过可视化向导完成安装配置，
以便不需要手动编辑配置文件（UX-DR1）。

**Acceptance Criteria:**

**Given** 访问 `/setup`，`install.lock` 不存在
**When** 页面加载
**Then** 展示 9 步步骤条（欢迎+环境检测 → DB配置 → Redis配置 → Migration → 管理员账号 → 加密密钥 → 访问地址 → 确认摘要 → 完成）
**And** 步骤条顶部固定，当前步骤高亮

**Given** 环境检测步骤
**When** 自动执行
**Then** 逐项展示 PASS（绿色）/ FAIL（红色），任意 FAIL 阻止进入下一步，展示具体错误

**Given** 管理员密码不符合复杂度要求（< 8 位，无大写/数字）
**When** 提交
**Then** 密码输入框下方显示规则提示，按钮保持禁用

**Given** 安装成功
**When** 进入步骤 9
**Then** 展示成功状态，「前往登录」按钮跳转 `/login`

---

## Epic 2: 用户认证与 RBAC

**Epic 目标：** 管理员可以安全登录、RBAC 权限控制至 API 端点粒度、所有变更写入不可变审计日志。

### Story 2.1: 用户登录与 JWT 认证

作为管理员，
我想要使用用户名+密码登录并获取 JWT，
以便访问受保护的 API（FR-5）。

**Acceptance Criteria:**

**Given** 正确的用户名和密码
**When** `POST /api/v1/auth/login`
**Then** 返回 `{token: "...", expires_at: "..."}，JWT 有效期 24 小时（可配置）

**Given** 错误密码连续 5 次
**When** 第 5 次提交
**Then** 账号锁定 15 分钟，响应 429，含剩余解锁时间，不区分「用户名错误」和「密码错误」

**Given** 有效 JWT
**When** 请求受保护 API，Authorization: Bearer {token}
**Then** 请求通过，`request_id` 和 `user_id` 注入到 Context

**Given** 过期或无效 JWT
**When** 请求受保护 API
**Then** 返回 401

---

### Story 2.2: RBAC 角色与权限控制

作为超级管理员，
我想要为用户分配角色并控制其可访问的 API 端点，
以便不同职责的运营人员只能操作其授权范围（FR-6）。

**Acceptance Criteria:**

**Given** 系统预置角色：超级管理员（全权）、DNS/CDN 运营管理员（操作）、只读审计员（只读）
**When** 只读审计员调用 `POST /api/v1/dns-records`（写操作）
**Then** 返回 403 `{code:"FORBIDDEN"}`，不返回 404

**Given** 超级管理员修改某用户角色
**When** 变更后，该用户在 ≤5 秒内发起请求
**Then** 新权限立即生效（不需要重新登录），Redis 缓存已清除

**Given** 管理员创建自定义角色并分配 `dns:records:write` 权限
**When** 该角色用户调用 `POST /api/v1/dns-records`
**Then** 请求通过

---

### Story 2.3: 审计日志中间件

作为超级管理员，
我想要所有资源变更操作自动记录到不可变审计日志，
以便事后追溯任何配置变更（FR-7，AD-15）。

**Acceptance Criteria:**

**Given** 任意管理员执行 DNS 记录新增操作
**When** API 返回成功
**Then** `audit_logs` 表写入一条记录，含：operator_id、operator_ip、action=`dns.record.create`、resource_type=`dns_record`、resource_id、before_state=`null`、after_state（≤4KB JSON 摘要）、created_at
**And** 写入为异步（不阻塞 API 响应）

**Given** 审计日志表
**When** 任意应用层代码尝试调用删除审计日志的 Service 方法
**Then** 编译报错（方法不存在）

**Given** 筛选条件：operator_id=1，time_range 最近 7 天
**When** `GET /api/v1/admin/audit-logs?operator_id=1&start=...&end=...`
**Then** 返回符合条件的审计记录，cursor-based 分页

---

### Story 2.4: CSRF / XSS / 安全响应头

作为系统，
我想要所有 API 响应包含安全响应头并验证 CSRF Token，
以便防止 CSRF 和 XSS 攻击（FR-8）。

**Acceptance Criteria:**

**Given** 任意 API 响应
**When** 检查 HTTP 响应头
**Then** 包含：`X-Content-Type-Options: nosniff`、`X-Frame-Options: DENY`、`Content-Security-Policy: default-src 'self'`

**Given** 修改类请求（POST/PUT/DELETE）
**When** 未携带有效 CSRF Token
**Then** 返回 403

---

## Epic 3: DNS Provider 插件框架

**Epic 目标：** 建立 DNS Provider Port 接口和 Registry，确保新增服务商不修改调度层代码。

### Story 3.1: DNS Provider Port 接口定义

作为架构，
我想要在 domain/port/ 中定义完整的 DNSProvider Go 接口，
以便所有适配器实现统一契约，调度层永不 import 具体适配器（AD-1，AD-2）。

**Acceptance Criteria:**

**Given** `internal/domain/port/dns_provider.go` 文件
**When** 编译
**Then** 文件中定义的 `DNSProvider` 接口包含全部 13 个方法（见 AD-2）
**And** 文件中无任何外部包 import（仅 `context`, `time`, 自定义 value objects）
**And** `UnsupportedCapabilityError` 类型定义存在

---

### Story 3.2: Provider Registry 与依赖注入

作为应用层，
我想要通过 ProviderRegistry 获取 Provider 实例，
以便在不 import 具体适配器包的情况下调用 Provider（AD-1）。

**Acceptance Criteria:**

**Given** AliyunDNS 适配器已注册到 Registry
**When** `registry.GetDNS("aliyun")`
**Then** 返回实现了 `DNSProvider` 接口的实例

**Given** 未注册的 Provider ID
**When** `registry.GetDNS("nonexistent")`
**Then** 返回 `ErrProviderNotFound`

**Given** 编译检查
**When** Application 层代码
**Then** 无任何对 `internal/infra/provider/*` 包的直接 import

---

### Story 3.3: Provider 能力检测与缓存

作为运营者，
我想要系统检测并缓存每个 DNS 账号的套餐能力，
以便在操作时提前拦截不支持的功能（FR-10）。

**Acceptance Criteria:**

**Given** 阿里云 DNS 免费版账号
**When** 调用 `DetectCapabilities`
**Then** 返回 `Capabilities{Lines: ["default"], MaxRecordsPerSet: 10, MinTTL: 600, ISPRouting: false}`
**And** 能力档案写入 `dns_accounts.capabilities_json` 和 Redis（TTL 1 小时）

**Given** 账号能力缓存存在
**When** 运营者在 UI 点击「手动刷新能力」
**Then** `PUT /api/v1/dns-accounts/:id/refresh-capabilities` 清除缓存并重新检测

---

### Story 3.4: Provider API 限速控制（令牌桶）

作为适配器，
我想要内置令牌桶限速保护，
以便 API 超限时本地排队重试，不抛出未处理异常（FR-11）。

**Acceptance Criteria:**

**Given** Provider 适配器配置 `rate_limit=300/min`
**When** 1 秒内发出 400 次 API 调用
**Then** 超出限额的调用在本地等待令牌，不返回 429 给上层
**And** 限速事件写入 DEBUG 日志：`{module:"aliyun-dns", event:"rate_limited", waited_ms:xxx}`

---

### Story 3.5: 凭证加密存储与 API 管理

作为运营者，
我想要 Provider API 凭证在数据库中加密存储，
以便数据库泄露时凭证不被读取（FR-12，AD-8）。

**Acceptance Criteria:**

**Given** 新增阿里云 DNS 账号（AccessKeyId + Secret）
**When** `POST /api/v1/dns-accounts`
**Then** `dns_accounts.credentials_enc` 存储 AES-256-GCM 密文，不存明文

**Given** 查询账号详情 API
**When** `GET /api/v1/dns-accounts/:id`
**Then** 响应中 credentials 显示为 `{access_key_id: "****ABCD", secret: "****"}` 脱敏格式

**Given** 凭证轮换
**When** `PUT /api/v1/dns-accounts/:id/credentials`（先验证新凭证）
**Then** 验证通过后原子替换旧凭证，验证失败则旧凭证不变

---

## Epic 4: DNS Provider 适配器（阿里云 / DNSPod / Cloudflare）

**Epic 目标：** 三个 DNS Provider 适配器全部实现，准确报告套餐能力，能力限制在 UI 层前置拦截。

### Story 4.1: 阿里云 DNS 适配器

作为运营者，
我想要添加阿里云 DNS 账号并管理解析记录，
以便通过 SmartCDN 统一管理阿里云域名（FR-13）。

**Acceptance Criteria:**

**Given** 有效的阿里云 AccessKeyId + Secret
**When** `ValidateCredentials` 调用
**Then** 返回 nil，账号状态变为「已连接」

**Given** 阿里云免费套餐账号
**When** `DetectCapabilities`
**Then** 返回 `ISPRouting: false`，UI 中三网调度显示「不支持，当前套餐限制」

**Given** 阿里云专业版账号
**When** 创建分运营商 A 记录（电信线路，IP=1.2.3.4）
**Then** 记录在阿里云 DNS 控制台可见，SmartCDN 本地 dns_records 写入一条记录

**Given** API 调用频率超过 300 次/分钟
**When** 继续发起请求
**Then** 本地令牌桶等待，不抛出异常，限速事件记录 DEBUG 日志

---

### Story 4.2: 腾讯云 DNSPod 适配器

作为运营者，
我想要添加 DNSPod 账号并管理解析记录，
以便通过 SmartCDN 统一管理 DNSPod 域名（FR-14）。

**Acceptance Criteria:**

**Given** DNSPod 免费版账号
**When** 在同一记录集下创建第 3 条 A 记录
**Then** API 调用前系统拦截，返回 `{code:"PROVIDER_LIMIT_EXCEEDED", message:"DNSPod 免费版同记录集上限 2 条"}`

**Given** DNSPod 免费版
**When** `DetectCapabilities`
**Then** 返回 `ISPRouting: false, MaxRecordsPerSet: 2`

**Given** DNSPod 专业版
**When** 创建电信线路 A 记录
**Then** 记录包含 `line=telecom` 参数，API 调用成功

---

### Story 4.3: Cloudflare DNS 适配器

作为运营者，
我想要添加 Cloudflare API Token 并管理 Zone 内的 DNS 记录，
以便通过 SmartCDN 统一管理 Cloudflare 域名（FR-15）。

**Acceptance Criteria:**

**Given** 有效的 Cloudflare API Token
**When** `ValidateCredentials`
**Then** 返回 nil，账号状态「已连接」

**Given** Proxied 状态的 A 记录（proxied=true）
**When** SmartCDN 更新该记录的值
**Then** 记录的 proxy 状态保持不变（proxied=true），仅修改 value

**Given** 超出 API 限额（1200 次/5 分钟）
**When** 继续发起调用
**Then** 本地令牌桶等待，不抛出异常

---

## Epic 5: 多域名与 DNS 记录管理

**Epic 目标：** 运营者可以管理多个域名的 DNS 配置，完整的记录 CRUD，以及历史快照回滚。

### Story 5.1: 多账号管理

作为超级管理员，
我想要为同一 Provider 添加多个账号并独立管理，
以便不同域名使用不同凭证（FR-16）。

**Acceptance Criteria:**

**Given** 添加两个阿里云 DNS 账号
**When** 查看账号列表
**Then** 两个账号独立显示，分别显示连接状态、今日 API 调用次数

**Given** 禁用账号 A
**When** 下次调度任务执行时依赖账号 A 的域名
**Then** 该域名的调度任务状态标记为 BLOCKED，并触发告警

---

### Story 5.2: 域名 DNS 配置管理

作为运营者，
我想要为每个域名独立配置 DNS Provider、TTL 和发布策略，
以便精细控制每个域名的解析行为（FR-17）。

**Acceptance Criteria:**

**Given** 创建域名 `jdmv.net`，绑定阿里云 DNS 账号
**When** `POST /api/v1/domains`
**Then** 域名记录创建，`icp_filed` 默认 false，从 Provider API 拉取现有记录同步

**Given** 配置最大发布 IP 数 = 5
**When** 当前账号套餐 `MaxRecordsPerSet = 2`
**Then** API 返回 422：`{code:"EXCEEDS_PROVIDER_LIMIT", message:"账号套餐限制为 2 条"}`

**Given** `icp_filed=false` 的域名
**When** 通过 `GET /api/v1/domains/:id/available-cdn`
**Then** 响应列表中不包含任何大陆境内 CDN 节点

---

### Story 5.3: DNS 记录 CRUD

作为运营者，
我想要创建、修改、删除多种类型的 DNS 记录，
以便精细管理域名解析（FR-18）。

**Acceptance Criteria:**

**Given** 创建 A 记录（host=www, value=1.2.3.4, ttl=300, line=telecom）
**When** `POST /api/v1/domains/:id/records`
**Then** 调用 Provider `CreateRecord`，记录写入 `dns_records`，审计日志记录
**And** SSRF 防护：若 value 为 RFC1918 地址，API 返回 400 `{code:"SSRF_BLOCKED"}`（AD-19）

**Given** 未由 SmartCDN 托管的记录（is_external=true）
**When** 批量删除操作包含外部记录
**Then** 操作拒绝，提示「外部记录不允许批量操作，请逐条确认」

**Given** 批量操作（删除 10 条记录）
**When** 触发高危确认对话框，用户未输入域名确认文本
**Then** 确认按钮保持 disabled

---

### Story 5.4: DNS 历史快照与回滚

作为运营者，
我想要查看任意时间点的 DNS 记录快照并回滚，
以便在误操作后快速恢复（FR-19）。

**Acceptance Criteria:**

**Given** 执行 DNS 记录变更
**When** 变更发起前
**Then** `dns_snapshots` 表已写入一条快照（snapshot_json 含完整记录集）

**Given** 选择 7 天前的快照
**When** 点击「预览回滚」
**Then** 展示 Dry Run 结果（将要删除的记录 + 将要创建的记录），不实际执行

**Given** 确认回滚操作（通过高危确认对话框，输入域名）
**When** `POST /api/v1/domains/:id/snapshots/:snapshotId/rollback`
**Then** 通过 AD-5 13 步流程执行，回滚操作写入审计日志

---

## Epic 6: DNS 安全发布引擎（13 步）

**Epic 目标：** 所有 DNS 写操作强制经过 13 步安全发布流程，含分布式锁、幂等检查、权威验证和失败自动回滚。

### Story 6.1: DNSPublishJob 核心引擎

作为系统，
我想要所有 DNS 写操作经过 DNSPublishJob 编排的 13 步流程，
以便确保每次变更安全、幂等、可回滚（FR-21，AD-5）。

**Acceptance Criteria:**

**Given** 发起一次 DNS A 记录更新
**When** `publishing_jobs` 中存在 RUNNING 状态记录
**Then** 该记录完成前，同一域名同一线路的新发布任务被 PENDING 排队

**Given** 步骤 6（CreateBatch）执行后 Provider API 超时
**When** 超时发生
**Then** 系统执行回滚（从 dns_snapshots 恢复），publishing_jobs.status=FAILED，审计日志记录失败

**Given** 目标状态与当前状态相同（幂等检查）
**When** 步骤 3 执行
**Then** 任务短路返回 NO_OP，不调用 Provider API，不写 dns_snapshots

---

### Story 6.2: 分布式锁保护

作为系统，
我想要同一域名同一线路同时只允许一个发布任务执行，
以便防止并发 DNS 写入导致状态不一致（FR-22，AD-6）。

**Acceptance Criteria:**

**Given** 域名 A 线路 telecom 正在执行发布任务（持有 Redis 锁）
**When** 第二个发布任务尝试获取锁
**Then** 返回 `LockContentionError`，第二个任务放弃本次轮次

**Given** 锁 TTL = 10min，发布任务在 9min 内完成
**When** 任务完成
**Then** Lua CAS 脚本安全释放锁，锁 key 被删除

**Given** 应用进程崩溃，锁未释放
**When** 10min 后 TTL 到期
**Then** 锁自动释放，后续发布任务可正常执行

---

### Story 6.3: 权威 DNS 验证与外部 HTTPS 检查

作为系统，
我想要 DNS 记录更新后通过权威 NS 查询和 HTTPS 外部检查确认变更生效，
以便在记录异常时自动触发回滚（AD-5 步骤 7-8）。

**Acceptance Criteria:**

**Given** DNS 记录成功更新
**When** 步骤 7 权威验证（非递归查询权威 NS）
**Then** 轮询超时 60 秒，每 10 秒查询一次；超时后标记任务 FAILED 并回滚

**Given** 新 IP（1.2.3.4）通过权威验证
**When** 步骤 8 外部 HTTPS 检查
**Then** 检查 HTTPS 返回 200，检查超时 15 秒；失败则回滚，不让空响应 IP 进入 Published Set

---

## Epic 7: DNS 迁移与 Multi-DNS

**Epic 目标：** 运营者可以安全地将域名从一个 Provider 迁移到另一个，NS 切换需人工授权，可配置主备双权威 DNS。

### Story 7.1: DNS 迁移工作流（模式 B）

作为运营者，
我想要通过向导完成 DNS Provider 迁移，
以便在切换 NS 前确保所有记录在目标 Provider 已同步（FR-20，FR-25，UX-DR6）。

**Acceptance Criteria:**

**Given** 迁移工作流步骤 4（完整性校验）
**When** 目标 Provider 中存在 2 条记录未同步
**Then** 展示差异列表，阻止进入步骤 5，提供「自动补全缺失记录」按钮

**Given** 步骤 5（NS 切换授权）
**When** 展示给运营者
**Then** 显示当前 NS 和目标 NS 的完整对比，仅提供「我已修改 NS」按钮，系统不自动执行 NS 修改

**Given** 运营者确认 NS 切换后
**When** 步骤 6（监测委派变化）
**Then** 系统轮询 `GetNameservers`，每 5 分钟检查一次，最多检查 72 小时；超时后告警

---

### Story 7.2: 迁移回滚与异常恢复

作为运营者，
我想要在迁移出现问题时回滚到迁移前状态，
以便最小化迁移失败的影响（FR-26）。

**Acceptance Criteria:**

**Given** 迁移步骤 3（预同步）失败（目标 Provider API 错误）
**When** 失败后
**Then** 迁移任务状态=FAILED，原 Provider 记录未改动，可重新发起迁移

---

### Story 7.3: Multi-DNS 双权威配置

作为运营者，
我想要为域名配置主+备双权威 DNS，
以便主 Provider 故障时备用 Provider 可接管（FR-23，FR-24）。

**Acceptance Criteria:**

**Given** 域名配置了主权威（阿里云）+ 备权威（Cloudflare）
**When** 执行 DNS 记录更新
**Then** 主权威和备权威均执行更新操作（通过各自的 AD-5 流程）

**Given** 主权威更新成功，备权威更新失败
**When** 失败后
**Then** 告警推送「备权威同步失败」，主权威记录不回滚，记录差异状态标记

---

## Epic 8: CDN Provider 框架与 Cloudflare SaaS / 阿里云 CDN

**Epic 目标：** CDN Provider 接口建立，Cloudflare Zone 管理、Custom Hostname 配置（含 DCV 自动化）、阿里云 CDN 接入。

### Story 8.1: CDN Provider Port 接口与 Registry

作为架构，
我想要在 domain/port/ 中定义 CDNProvider 和 CloudflareSaaSProvider 接口，
以便 CDN 功能不直接依赖 SDK（AD-3）。

**Acceptance Criteria:**

**Given** `internal/domain/port/cdn_provider.go`
**When** 编译
**Then** 文件定义 `CDNProvider` 接口（含 `CloudflareSaaS() (CloudflareSaaSProvider, error)` 方法）和 `CloudflareSaaSProvider` 子接口（5 个方法）
**And** 非 Cloudflare CDN 实现调用 `CloudflareSaaS()` 返回 `UnsupportedCapabilityError`

---

### Story 8.2: Cloudflare Zone 管理

作为运营者，
我想要查看和管理 Cloudflare 账号下的 Zone 列表，
以便监控 Zone 状态和触发缓存清理（FR-28）。

**Acceptance Criteria:**

**Given** 有效 Cloudflare API Token（含 Zone.Read 权限）
**When** `GET /api/v1/cdn/cloudflare/zones`
**Then** 返回 Zone 列表（name, status, plan, record_count）

**Given** 触发 Zone 缓存全清
**When** `POST /api/v1/cdn/cloudflare/zones/:zoneId/purge`
**Then** 调用 Cloudflare API `purge_everything`，返回成功，写入审计日志

---

### Story 8.3: Cloudflare SaaS Custom Hostname（含 DCV）

作为运营者，
我想要通过 4 步向导配置 Cloudflare SaaS Custom Hostname，
以便将客户域名接入 Cloudflare CDN 加速（FR-29，FR-30，UX-DR7）。

**Acceptance Criteria:**

**Given** 步骤 1（账号选择）
**When** 展示给运营者
**Then** 页面顶部 Banner 明确显示：「标准模式仅支持 CNAME 接入，A 记录需确认套餐支持」

**Given** Fallback Origin（origin.wueym.com）填写后
**When** 点击「验证源站」
**Then** 系统发起 HTTPS HEAD 请求（通过 SSRFGuard 验证），展示 HTTP 状态码、TLS 证书信息、响应时间；若源站返回非 200，阻止进入步骤 3

**Given** 选择 DCV 类型=DNS TXT，选择自动添加
**When** Custom Hostname 创建
**Then** 系统调用绑定的 DNS Provider 自动添加 `_acme-challenge.jdmv.net` TXT 记录

**Given** DCV 验证轮询
**When** 验证通过（Cloudflare API status=active）
**Then** 展示 CNAME 配置指引（`www.jdmv.net CNAME edge.wueym.com`），Certificate Status=active

---

### Story 8.4: 阿里云 CDN 管理接入

作为运营者，
我想要将阿里云 CDN 账号接入 SmartCDN 并查看域名加速状态，
以便在统一界面管理阿里云 CDN（FR-31）。

**Acceptance Criteria:**

**Given** 有效的阿里云 AccessKeyId + Secret（含 CDN 权限）
**When** 添加阿里云 CDN 账号
**Then** `DetectCapabilities` 返回接入域名列表

**Given** `icp_filed=false` 的域名
**When** 尝试绑定到阿里云 CDN（大陆节点）
**Then** API 返回 `{code:"ICP_VIOLATION"}`（AD-20）

---

### Story 8.5: CDN 缓存规则与清理

作为运营者，
我想要为 CDN Zone 配置缓存规则并手动触发缓存清理，
以便精细控制内容缓存策略（FR-32）。

**Acceptance Criteria:**

**Given** 配置缓存规则：`*.jpg` 缓存 7 天
**When** `POST /api/v1/cdn/cache-rules`
**Then** 规则写入数据库，可选择是否同步到 Provider

**Given** 手动清理指定 URL 缓存
**When** `POST /api/v1/cdn/cache-purge` `{urls: ["https://jdmv.net/logo.jpg"]}`
**Then** 调用对应 Provider 的缓存清理 API，返回任务 ID，写入审计日志

---

## Epic 9: 自有节点管理、Node Agent 与缓存

**Epic 目标：** 运营者可以添加自有家宽/服务器节点，部署 Node Agent，配置节点缓存规则，查看实时状态。

### Story 9.1: 自有节点 CRUD 与 HTTPS 检查

作为运营者，
我想要管理自有反向代理节点的基础信息，
以便将节点加入 IP 候选池（FR-33）。

**Acceptance Criteria:**

**Given** 添加节点（IP=1.2.3.4，ISP=电信，地区=上海，回源地址=origin.wueym.com）
**When** `POST /api/v1/cdn/nodes`（含「我已获授权使用此节点」确认）
**Then** 节点创建，立即触发 HTTPS 检查（通过 SSRFGuard，检查 1.2.3.4:443）
**And** HTTPS 检查结果写入节点详情（HTTP 状态、TLS 证书域名、响应时间）

---

### Story 9.2: Node Agent 安装与 mTLS 认证

作为运营者，
我想要为节点安装 Node Agent 并建立 mTLS 通信，
以便管理中心可以获取节点状态和下发任务（FR-34，AD-11）。

**Acceptance Criteria:**

**Given** 在节点详情页
**When** 点击「查看安装命令」
**Then** 显示一键安装脚本（含管理中心地址和节点认证 Token）

**Given** Node Agent 以正确 mTLS 证书连接
**When** `POST /api/agent/v1/node-status`
**Then** 接受并更新节点 CPU/内存/HTTPS 状态

**Given** Node Agent 使用非法/过期证书
**When** 建立 TLS 连接
**Then** 握手失败，服务器拒绝连接

**Given** 下发任务（reload_nginx）
**When** Node Agent 拉取任务 `GET /api/agent/v1/tasks`
**Then** 执行 nginx reload，回写 `node_task_results.status=SUCCEEDED`

**Given** 下发任务类型为 `arbitrary_shell`（非白名单）
**When** Agent 收到任务
**Then** Agent 拒绝执行，记录 `FORBIDDEN_TASK_TYPE` 日志

---

### Story 9.3: 节点资源池管理

作为运营者，
我想要将多个节点组织为资源池，
以便域名的 IP 候选集可以灵活选择（FR-35）。

**Acceptance Criteria:**

**Given** 创建资源池「香港节点池」并添加 3 个香港节点
**When** `POST /api/v1/cdn/pools`
**Then** 资源池创建，关联 3 个节点

**Given** 域名绑定资源池「香港节点池」
**When** 调度器计算 IP 候选集
**Then** 只从该资源池内的节点 IP 中选择

---

## Epic 10: 三网智能调度与 Probe Agent

**Epic 目标：** 分布式 Probe Agent 在真实 ISP 网络收集性能数据，调度器定期计算最优 IP 并发布，含熔断保护。

### Story 10.1: Probe Agent Binary 与 mTLS 通信

作为系统管理员，
我想要在真实 ISP 网络部署 Probe Agent，
以便收集真实的三网性能数据（FR-37，AD-10）。

**Acceptance Criteria:**

**Given** Probe Agent binary 以 ISP=电信、Region=上海配置部署
**When** Agent 以 mTLS 客户端证书连接 `POST /api/agent/v1/probe-results`
**Then** 管理中心验证客户端证书（CA 签发），接受数据写入 `probe_results`

**Given** Probe Agent 使用未经 CA 签发的自签名证书
**When** 尝试建立连接
**Then** mTLS 握手失败，管理中心拒绝连接

**Given** Agent 拉取配置 `GET /api/agent/v1/config`
**When** 管理中心更新了目标 IP 列表
**Then** Agent 下次拉取获得新配置，开始对新 IP 探测

---

### Story 10.2: IP 候选池管理与状态机

作为系统，
我想要维护每个域名每条线路的 IP 候选池，
以便调度器从池中选择最优 IP（FR-39）。

**Acceptance Criteria:**

**Given** 新增候选 IP（1.2.3.4）到电信线路
**When** 通过 HTTPS 安全检查后
**Then** 状态从 Candidate → Eligible

**Given** IP 通过手动锁定
**When** 调度器计算时
**Then** 锁定 IP 不被替换，显示「🔒 已锁定」标识

**Given** IP HTTPS 检查连续 3 次失败
**When** 状态更新
**Then** 状态从 Eligible → Offline

---

### Story 10.3: 三网调度评分与自动发布

作为系统，
我想要定期计算三网各线路的最优 IP 并自动发布到 DNS，
以便用户始终访问最快节点（FR-40，FR-41，AD-12）。

**Acceptance Criteria:**

**Given** 调度间隔 10 分钟，电信线路有 5 个 Eligible IP
**When** 调度器运行一轮
**Then** 聚合各 Probe Agent 最近 5 次探测数据，按评分公式（延迟40%+TTFB40%+成功率20%）排序
**And** 若排名第一的 IP 不在 Published Set 中，触发 AD-5 发布流程

**Given** 30 分钟内电信线路已切换 3 次
**When** 第 4 次需要切换
**Then** 触发熔断：暂停该线路自动调度，仪表盘显示熔断状态 Banner，推送告警
**And** 运营者手动点击「恢复自动调度」后熔断解除

---

## Epic 11: Node Agent 状态上报与管理

**Epic 目标：** Node Agent 定期上报节点 CPU/内存/HTTPS 状态，管理中心可下发授权任务。

### Story 11.1: Node Agent 状态上报与监控集成

作为运营者，
我想要在节点详情页查看实时的 CPU/内存/HTTPS 状态，
以便监控自有节点的健康状况（FR-42）。

**Acceptance Criteria:**

**Given** Node Agent 每 60 秒上报 `{cpu_percent, memory_percent, https_status, last_checked_at}`
**When** `POST /api/agent/v1/node-status`
**Then** `node_agents.last_seen_at` 更新，状态写入节点详情

**Given** Node Agent 超过 5 分钟未上报
**When** 状态检查
**Then** 节点状态变为「离线」，触发告警「节点 X 已离线」

---

## Epic 12: SSL 证书管理与 ACME 自动化

**Epic 目标：** 运营者可以管理 ACME 账号、申请/导入证书、配置自动续期，系统在证书到期前自动完成 DNS-01 验证。

### Story 12.1: ACME 账号管理

作为运营者，
我想要管理 Let's Encrypt 和 ZeroSSL 的 ACME 账号，
以便用于自动证书申请（FR-43）。

**Acceptance Criteria:**

**Given** 添加 ACME 账号（email=admin@jdmv.net，CA=Let's Encrypt）
**When** `POST /api/v1/ssl/acme-accounts`
**Then** 生成 ECDSA P-256 密钥对，向 Let's Encrypt 注册账号，私钥以 AES-256-GCM 加密存储
**And** 账号注册成功后状态显示「已激活」

---

### Story 12.2: SSL 证书申请（DNS-01 Challenge）

作为运营者，
我想要通过 DNS-01 Challenge 自动申请 Let's Encrypt 证书，
以便无需配置 HTTP 服务器（FR-44）。

**Acceptance Criteria:**

**Given** 申请 `www.jdmv.net` 证书，选择 DNS Provider 账号和 ACME 账号
**When** `POST /api/v1/ssl/certificates/apply`
**Then** 创建 ACME Order → 获取 DNS-01 Challenge → 自动在 jdmv.net 的 DNS Provider 创建 `_acme-challenge` TXT 记录 → 轮询验证（最多 10 次，每次 60s）→ 验证通过下载证书

**Given** DNS-01 Challenge TXT 传播超时（10 次轮询均未检测到）
**When** 超时
**Then** 任务状态=FAILED，审计日志记录，自动清理 `_acme-challenge` TXT 记录，不残留

---

### Story 12.3: 证书导入与管理

作为运营者，
我想要导入外部颁发的证书并管理证书元数据，
以便统一管理所有证书（FR-44 部分，FR-51）。

**Acceptance Criteria:**

**Given** 上传 PEM 格式证书文件（cert chain + private key）
**When** `POST /api/v1/ssl/certificates/import`
**Then** 解析证书 SAN、有效期、颁发者；私钥 AES-256-GCM 加密存储；证书状态=VALID

**Given** 私钥与证书不匹配
**When** 导入
**Then** 返回 422 `{code:"KEY_CERT_MISMATCH"}`

---

### Story 12.4: 证书到期监控与自动续期触发

作为系统，
我想要监控证书到期时间并在到期前触发续期工作流，
以便证书不会在无人干预的情况下过期（FR-45，FR-50）。

**Acceptance Criteria:**

**Given** 证书剩余有效期 ≤ 30 天
**When** 每日扫描任务运行
**Then** 触发证书关联的续期工作流，工作流状态写入 `workflow_executions`

**Given** 证书剩余 ≤ 14 天，工作流执行失败
**When** 失败后
**Then** 推送 CRITICAL 告警「证书 jdmv.net 续期失败，剩余 14 天，需人工介入」

---

## Epic 13: 证书工作流 DAG 引擎与部署

**Epic 目标：** 可视化 DAG 工作流编排证书的申请、验证、部署全流程，支持并行分支部署到宝塔 Nginx 和阿里云 CDN。

### Story 13.1: DAG 工作流引擎核心

作为系统，
我想要执行以 DAG 结构定义的工作流，
以便证书申请、验证、部署可以按依赖关系并行执行（FR-46，AD-7）。

**Acceptance Criteria:**

**Given** 一个包含 2 个并行部署分支的 DAG（BaotaNginx + AliyunCDN）
**When** ACME 验证节点成功后
**Then** 两个部署分支同时启动（Go goroutine 并行执行）

**Given** 一个部署分支（AliyunCDN）失败
**When** 失败后
**Then** 另一个分支（BaotaNginx）不受影响，继续执行
**And** 失败分支按 retryPolicy 重试（默认 3 次，间隔 60 秒）

**Given** 节点执行过程
**When** 状态变化
**Then** 实时写入 `workflow_node_runs`，前端通过 SSE 接收更新（见 Epic 15）

---

### Story 13.2: 证书部署到宝塔 Nginx（Deploy Target）

作为系统，
我想要将证书安全部署到宝塔面板的 Nginx，
以便网站 HTTPS 自动更新（FR-48，AD-4）。

**Acceptance Criteria:**

**Given** 证书部署到宝塔 Nginx（通过 Node Agent）
**When** `DeployCertificate` 执行
**Then** 步骤：备份旧证书 → 写入新证书到临时目录 → `nginx -t` 语法检查 → 原子替换 → nginx reload → 外部 HTTPS 检查（通过 SSRFGuard）
**And** 若 `nginx -t` 失败，回滚到旧证书，标记部署 FAILED

---

### Story 13.3: 证书部署到阿里云 CDN

作为系统，
我想要将证书通过 API 更新到阿里云 CDN 加速域名，
以便 CDN 节点使用最新证书（FR-49）。

**Acceptance Criteria:**

**Given** 证书部署到阿里云 CDN（accelerated domain: jdmv.net）
**When** `DeployCertificate` 执行
**Then** 调用阿里云 CDN 证书更新 API，轮询任务状态至 SUCCESS（最多等待 120 秒）

---

### Story 13.4: DAG 工作流 UI 编辑器

作为运营者，
我想要通过可视化编辑器设计和修改证书工作流，
以便配置证书申请和部署的完整流程（FR-47，UX-DR9）。

**Acceptance Criteria:**

**Given** 访问 `/ssl/workflows/:id/edit`
**When** 页面加载
**Then** 展示画布（使用 React Flow 或 @xyflow/react），已有节点从 `cert_workflows.dag_json` 渲染
**And** 7 种节点类型视觉区分（触发/操作/条件/并行/通知/成功/失败）

**Given** 拖拽节点到画布并连线
**When** 点击「保存」
**Then** DAG JSON 更新到 `cert_workflows.dag_json`，工作流版本号递增

---

### Story 13.5: 工作流执行日志（实时 SSE）

作为运营者，
我想要实时查看证书工作流的执行进度和每个节点的日志，
以便监控证书续期过程（FR-52）。

**Acceptance Criteria:**

**Given** 工作流执行中
**When** 访问 `/ssl/execution-logs/:runId`
**Then** 左侧 DAG 拓扑已执行节点绿色高亮，当前节点蓝色脉冲，失败节点红色
**And** 右侧展示选中节点的实时日志（通过 SSE 流 `GET /api/v1/tasks/:id/log-stream`）

---

## Epic 14: 监控、告警与通知

**Epic 目标：** 统一仪表盘展示三网性能趋势、DNS/CDN 健康，可配置告警规则，通过邮件/Webhook 通知。

### Story 14.1: 监控仪表盘与性能趋势

作为运营者，
我想要在仪表盘实时查看系统运行状态，
以便快速发现和定位问题（FR-53，UX-DR4）。

**Acceptance Criteria:**

**Given** 仪表盘页面（/dashboard）
**When** 页面加载
**Then** 展示 4 个统计卡片（域名数/在线节点数/有效证书数/未解决告警数）
**And** 三网性能趋势折线图（ECharts）展示电信/联通/移动三条线，时间选择器（1h/6h/24h/7d）
**And** SSL 到期列表展示最近 5 条即将到期证书（剩余天数）

---

### Story 14.2: 告警规则与记录管理

作为运营者，
我想要配置告警规则并查看告警历史，
以便在问题发生时第一时间知晓（FR-54，FR-55）。

**Acceptance Criteria:**

**Given** 配置告警规则「节点离线告警」（触发条件：node.status=offline，等级=CRITICAL）
**When** 节点离线
**Then** `alert_records` 写入记录，前端实时收到告警推送（Epic 15）

**Given** 运营者点击「确认」告警
**When** `PUT /api/v1/monitoring/alerts/:id/acknowledge`
**Then** `alert_records.status=ACKNOWLEDGED`，记录确认者和时间

---

### Story 14.3: 通知渠道（邮件 / Webhook）

作为运营者，
我想要将告警通知发送到邮件或 Webhook，
以便在远程也能接收重要告警（FR-56）。

**Acceptance Criteria:**

**Given** 配置 Webhook 通知渠道（URL=https://hooks.example.com/smartcdn）
**When** 触发 CRITICAL 告警
**Then** 系统通过 SSRFGuard 验证 URL（允许外部 HTTPS，拒绝内网），发起 POST 请求，body 含告警详情

**Given** 邮件通知渠道已配置（SMTP）
**When** 证书剩余 ≤ 14 天
**Then** 发送邮件至配置的收件人，主题：`[SmartCDN] 证书 jdmv.net 即将到期（剩余 14 天）`

---

## Epic 15: 实时推送（SSE / WebSocket）

**Epic 目标：** 管理控制台通过 SSE 实时接收任务状态、告警和指标更新，工作流执行进度通过 WebSocket 推送。

### Story 15.1: SSE 全局事件流

作为前端，
我想要通过 SSE 订阅全局事件流，
以便仪表盘数据和告警实时更新（FR-57，AD-18）。

**Acceptance Criteria:**

**Given** 客户端连接 `GET /api/v1/events` （Authorization: Bearer token）
**When** 告警触发、任务状态变更
**Then** SSE 推送对应事件（`event: alert.created` / `event: job.status.changed`）

**Given** SSE 连接空闲 30 秒
**When** 无事件
**Then** 服务端发送 `: keepalive` comment，客户端不断连

**Given** SSE 连接断开
**When** 客户端自动重连
**Then** 通过 `Last-Event-ID` header 重新连接，服务端补发错过的事件（最多 1 分钟内）

---

### Story 15.2: 任务日志流（SSE）与工作流 WebSocket

作为前端，
我想要通过 SSE 实时查看任务日志，通过 WebSocket 接收工作流 DAG 执行进度，
以便监控长时间运行的操作（FR-57，AD-18）。

**Acceptance Criteria:**

**Given** 访问 `GET /api/v1/tasks/:id/log-stream`
**When** 任务执行中
**Then** SSE 流实时推送每条日志行（`event: log`, `data: {line: "...", ts: "..."}`)

**Given** 工作流执行页面打开 WebSocket
**When** 节点状态变更（RUNNING → SUCCEEDED）
**Then** WebSocket 推送 `{nodeId: "n1", status: "SUCCEEDED", finished_at: "..."}`

---

## Epic 16: 部署打包

**Epic 目标：** SmartCDN 可通过宝塔 Go 项目（P0）、systemd（P1）、Docker/Compose（P1）三种方式部署。

### Story 16.1: 单 Binary 打包（前端 embed）

作为运营者，
我想要通过单个 Go binary 运行整个 SmartCDN，
以便在宝塔 Go 项目中直接部署（FR-58，AD-16）。

**Acceptance Criteria:**

**Given** 执行 `make build`（先 `npm run build`，再 `go build -ldflags "-s -w"`)
**When** 构建完成
**Then** `dist/smartcdn` 单文件，大小 < 50MB
**And** 访问 `/` 返回 React SPA HTML
**And** 访问 `/api/v1/setup/check` 返回 JSON
**And** 无需 Node.js 或其他 runtime

---

### Story 16.2: systemd 单元文件

作为运维，
我想要通过 systemd 管理 SmartCDN 服务，
以便在原生 Linux 服务器上部署（FR-58）。

**Acceptance Criteria:**

**Given** 复制 `deploy/systemd/smartcdn.service` 到 `/etc/systemd/system/`
**When** `systemctl start smartcdn`
**Then** 服务启动，监听 8080 端口，`journalctl -u smartcdn` 输出 JSON 结构化日志

**Given** 应用崩溃
**When** systemd 检测
**Then** 自动重启（Restart=on-failure，RestartSec=5s）

---

### Story 16.3: Docker / Docker Compose 部署

作为运维，
我想要通过 Docker Compose 一键启动 SmartCDN，
以便容器化部署（FR-58）。

**Acceptance Criteria:**

**Given** 配置 `.env` 文件（DB/Redis/加密密钥）
**When** `docker-compose up -d`
**Then** 启动 smartcdn（8080）+ mysql:8.0 + redis:7 三个容器，服务健康检查通过

**Given** SmartCDN Dockerfile
**When** 构建
**Then** 使用多阶段构建（node:20 build frontend → golang:1.23 build binary → alpine:3 final），最终镜像 < 100MB

---

## Epic 17: 前端基础与 UX 完整性

**Epic 目标：** 前端 React SPA 实现完整设计规范：双主题切换、全局搜索（Cmd+K）、响应式布局、骨架屏、高危确认组件。

### Story 17.1: 前端项目脚手架与路由

作为前端开发者，
我想要建立 React 18 + TypeScript + Vite + Arco Design 项目脚手架，
以便后续所有前端 Story 在统一框架下开发（UX-DR3）。

**Acceptance Criteria:**

**Given** 执行 `npm run dev`
**When** 访问 http://localhost:5173
**Then** 加载 React SPA，路由 `/login` 显示登录页
**And** Arco Design 组件库正常渲染
**And** TypeScript 严格模式，无编译错误

---

### Story 17.2: 全局布局（左侧导航 + TopBar）

作为运营者，
我想要通过统一的导航结构访问所有模块，
以便快速在各功能间切换（UX-DR3）。

**Acceptance Criteria:**

**Given** 登录后访问 `/dashboard`
**When** 页面渲染
**Then** 左侧导航展开（200px），7 个一级菜单各含子菜单，当前路由高亮
**And** TopBar（48px）展示 Logo、面包屑、主题切换按钮、通知图标、用户菜单

**Given** 点击导航折叠按钮
**When** 切换
**Then** 导航宽度变为 48px（icon-only），折叠状态 localStorage 持久化

---

### Story 17.3: 双主题切换（浅色 / 深色）

作为运营者，
我想要切换深色和浅色主题，
以便在不同光线环境下舒适使用（UX-DR12）。

**Acceptance Criteria:**

**Given** 点击 TopBar 主题切换按钮
**When** 切换到深色
**Then** Arco Design `ConfigProvider theme="dark"` 生效，页面背景变为 `#17171A`，导航背景保持 `#001529`
**And** 选择持久化到 localStorage，刷新后保持

---

### Story 17.4: 全局搜索（Cmd+K）

作为运营者，
我想要通过 Cmd+K 快速搜索域名、IP、证书，
以便不需要逐层导航（UX-DR3）。

**Acceptance Criteria:**

**Given** 按下 Cmd+K（Mac）/ Ctrl+K（Windows）
**When** 搜索框激活
**Then** TopBar 搜索输入框聚焦，下方展示结果面板（max-height 400px）

**Given** 输入「jdmv」（防抖 300ms）
**When** 搜索
**Then** 结果分组展示：域名/IP/证书/节点，每组最多 3 条，点击跳转对应详情页

---

### Story 17.5: 高危操作二次确认组件

作为运营者，
我想要在执行高危操作时通过文本输入验证确认，
以便防止误操作（UX-DR8）。

**Acceptance Criteria:**

**Given** 触发 NS 切换操作
**When** 展示确认对话框
**Then** 对话框含：「⚠ 高危操作」标题、危险描述列表、文本输入框（提示「请输入域名以确认」）、确认按钮默认 disabled

**Given** 用户输入与目标域名完全匹配
**When** 确认按钮激活
**Then** 点击后执行操作，不绑定 Enter 键

---

### Story 17.6: 骨架屏与响应式布局

作为运营者，
我想要在数据加载时看到骨架屏，在平板端看到自适应布局，
以便在不同场景下有良好体验（UX-DR11，UX-DR15）。

**Acceptance Criteria:**

**Given** 表格数据加载中（< 500ms 未返回）
**When** 展示
**Then** 显示骨架表格行（Skeleton），数量与上次加载行数相同或默认 5 行

**Given** 浏览器宽度 768–1279px
**When** 访问任意列表页
**Then** 左侧导航自动折叠（48px），表格隐藏次要列（如「更新时间」），表单切换单列布局

**Given** 浏览器宽度 < 768px
**When** 访问任意列表页
**Then** 表格切换为卡片列表视图，高危操作（NS 切换/批量删除）按钮不显示

---

## 生产保护注意事项（PROD_CAUTION）

以下 Story 在测试和开发阶段**严禁使用 `www.jdmv.net` 进行测试**，须使用测试专用域名：

- Story 5.3（DNS 记录 CRUD）：测试使用 `test-dev.example.com`
- Story 6.1（DNS 安全发布引擎）：测试使用 `test-dev.example.com`
- Story 7.1（DNS 迁移工作流）：测试使用 `test-migration.example.com`
- Story 8.3（Cloudflare SaaS Custom Hostname）：测试使用非生产域名
- Story 10.3（三网调度评分与自动发布）：测试使用 `sched-test.example.com`
- Story 12.2（SSL 证书申请 DNS-01）：测试使用 Let's Encrypt Staging 环境 + 测试域名

任何对生产域名 `www.jdmv.net` 的 DNS、CDN、SSL 变更必须经过明确授权后方可执行。
