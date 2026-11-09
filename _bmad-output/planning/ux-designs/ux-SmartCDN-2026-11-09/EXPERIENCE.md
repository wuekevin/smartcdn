---
title: SmartCDN Experience Design
status: final
created: 2026-11-09
updated: 2026-11-09
visual-identity: DESIGN.md
prd: "../prds/prd-SmartCDN-2026-11-09/prd.md"
---

# SmartCDN Experience Design

## Foundation

**Form Factor：** Web 应用，桌面为主（1280px+），响应式支持平板（768–1279px）和移动端（<768px，只读）。

**UI 体系：** Arco Design React。DESIGN.md 定义了所有视觉令牌。本文档只描述行为增量——交互逻辑、状态模式、IA、流程，不重述 Arco 组件的默认外观。

**视觉身份参考：** DESIGN.md。每个颜色、字体、间距引用使用 `{path.to.token}` 语法，与 DESIGN.md frontmatter 对应。

---

## Information Architecture

### 全局导航结构

```
TopBar (48px fixed)
  Logo | Breadcrumb | GlobalSearch | ThemeToggle | Notifications | UserMenu

SideNav (200px expanded / 48px collapsed)
  仪表盘
  DNS 管理
    ├── DNS 服务商
    ├── DNS 账号
    ├── 域名管理
    ├── DNS 记录
    ├── 智能解析线路
    ├── DNS 迁移
    ├── Multi-DNS
    └── DNS 历史快照
  CDN 管理
    ├── CDN 服务商
    ├── Cloudflare Zones
    ├── Cloudflare SaaS
    ├── 阿里云 CDN/ESA
    ├── 自有节点
    ├── 节点资源池
    ├── 缓存规则
    └── 缓存清理
  智能调度
    ├── 调度仪表盘
    ├── 三网测速
    ├── 探针管理
    ├── 优选 IP 池
    ├── 调度策略
    ├── 自动发布任务
    ├── 发布历史
    └── 故障回滚
  SSL 证书
    ├── 证书总览
    ├── ACME 账号
    ├── 证书申请
    ├── 已签发证书
    ├── 证书导入
    ├── 自动续期
    ├── 证书部署目标
    ├── 证书工作流
    ├── 执行日志
    └── 证书监控
  监控中心
    ├── 实时监控
    ├── 性能趋势
    ├── DNS 健康
    ├── CDN 健康
    ├── HTTPS 状态
    └── 告警记录
  系统管理
    ├── 用户管理
    ├── 角色权限
    ├── API 凭证
    ├── 通知渠道
    ├── 审计日志
    ├── 系统配置
    ├── 备份恢复
    └── 系统诊断
```

**路由规则：**

| 路径 | 页面 |
|------|------|
| `/setup` | 安装向导（无 nav，仅在 install.lock 不存在时可访问） |
| `/login` | 登录页（无 nav） |
| `/dashboard` | 仪表盘 |
| `/dns/providers` | DNS 服务商 |
| `/dns/accounts` | DNS 账号 |
| `/dns/domains` | 域名管理 |
| `/dns/domains/:id` | 域名详情 |
| `/dns/records` | DNS 记录（跨域名视图） |
| `/dns/lines` | 智能解析线路 |
| `/dns/migrations` | DNS 迁移 |
| `/dns/multi` | Multi-DNS |
| `/dns/snapshots` | DNS 历史快照 |
| `/cdn/providers` | CDN 服务商 |
| `/cdn/cloudflare/zones` | Cloudflare Zones |
| `/cdn/cloudflare/zones/:id` | Zone 详情 |
| `/cdn/cloudflare/saas` | Cloudflare SaaS |
| `/cdn/cloudflare/saas/:id` | Custom Hostname 详情 |
| `/cdn/aliyun` | 阿里云 CDN/ESA |
| `/cdn/nodes` | 自有节点 |
| `/cdn/nodes/:id` | 节点详情 |
| `/cdn/pools` | 节点资源池 |
| `/cdn/cache-rules` | 缓存规则 |
| `/cdn/cache-purge` | 缓存清理 |
| `/scheduling/dashboard` | 调度仪表盘 |
| `/scheduling/speed-test` | 三网测速 |
| `/scheduling/probes` | 探针管理 |
| `/scheduling/ip-pool` | 优选 IP 池 |
| `/scheduling/policies` | 调度策略 |
| `/scheduling/jobs` | 自动发布任务 |
| `/scheduling/history` | 发布历史 |
| `/scheduling/rollback` | 故障回滚 |
| `/ssl/overview` | 证书总览 |
| `/ssl/acme-accounts` | ACME 账号 |
| `/ssl/apply` | 证书申请 |
| `/ssl/certificates` | 已签发证书 |
| `/ssl/certificates/:id` | 证书详情 |
| `/ssl/import` | 证书导入 |
| `/ssl/renewal` | 自动续期 |
| `/ssl/deploy-targets` | 证书部署目标 |
| `/ssl/workflows` | 证书工作流 |
| `/ssl/workflows/:id/edit` | 工作流 DAG 编辑器 |
| `/ssl/execution-logs` | 执行日志 |
| `/ssl/monitoring` | 证书监控 |
| `/monitoring/realtime` | 实时监控 |
| `/monitoring/trends` | 性能趋势 |
| `/monitoring/dns-health` | DNS 健康 |
| `/monitoring/cdn-health` | CDN 健康 |
| `/monitoring/https` | HTTPS 状态 |
| `/monitoring/alerts` | 告警记录 |
| `/admin/users` | 用户管理 |
| `/admin/roles` | 角色权限 |
| `/admin/api-credentials` | API 凭证 |
| `/admin/notifications` | 通知渠道 |
| `/admin/audit-logs` | 审计日志 |
| `/admin/settings` | 系统配置 |
| `/admin/backup` | 备份恢复 |
| `/admin/diagnostics` | 系统诊断 |

---

## Voice and Tone

**风格：** 专业、直接、无废话。用动词主导操作文案（「添加」「验证」「发布」），不用「请」「您好」等礼貌性废话。

**错误文案：** 给出具体原因和行动建议，不显示堆栈 trace。
- 错误：「出错了」
- 正确：「阿里云 DNS API 凭证无效，请检查 AccessKeyId 和 Secret 是否正确，或前往阿里云控制台重新生成凭证」

**危险操作文案：** 明确描述不可逆后果。
- 错误：「确认删除此记录？」
- 正确：「删除此 A 记录将立即从 DNS 服务商移除，当前通过此 IP 访问的用户将受到影响。如需恢复，请通过历史快照手动回滚。」

**状态文案：** 现在进行时（「发布中」「验证中」「续期中」），完成后立即切换为过去式（「已发布」「已验证」「已续期」）。

**数字单位：** 延迟始终显示 ms（如「234 ms」），TTL 显示 s 或人类可读（「300s / 5分钟」），证书剩余天数显示「X 天」。

---

## Component Patterns（行为增量）

### 全局搜索 (Global Search)

- **触发：** 点击顶栏搜索框；`Cmd+K` / `Ctrl+K`
- **行为：** 搜索框内联（不弹出遮罩），下方出现结果面板（max-height: 400px，可滚动）
- **搜索范围：** 域名、IP 地址、证书通用名、节点名称、API 账号名；实时搜索（防抖 300ms）
- **结果分组显示：** 分「域名」「IP」「证书」「节点」等小节，每节最多 3 条，全结果链接
- **空结果：** 「未找到匹配结果」+ 二级提示（「试试搜索域名、IP 或证书名称」）
- **键盘导航：** ↑↓ 选择，Enter 跳转，Esc 关闭

### 实时更新 (SSE / WebSocket)

- 仪表盘、调度仪表盘、执行日志页面通过 SSE 实时刷新
- 连接状态指示器：顶栏右侧显示连接状态小圆点（`{colors.success}`=已连接，`{colors.warning}`=重连中，`{colors.text-tertiary}`=离线）
- 断连后以指数退避自动重连（1s→2s→4s→8s→最大 30s）
- 重连成功后静默补齐错过的状态变更

### 分页与筛选 (Table Toolbar)

列表页标准工具栏（从左到右）：
1. 筛选下拉（主要筛选维度，如"状态""服务商"）
2. 时间范围选择器（若适用）
3. 搜索输入框（当前表格内搜索）
4. 分隔线
5. 刷新按钮（手动刷新，含最后更新时间 tooltip）
6. 列配置按钮（显示/隐藏列）

表格底部：分页控件（总条数、每页条数选择器、页码）。

### Provider 账号下拉选择

在所有需要选择 DNS/CDN Provider 账号的场景，下拉选项展示：账号名 + Provider Logo + 状态标签（已连接/凭证失效）。失效账号置灰但仍可选（允许查看，不允许触发操作）。

### 操作结果 Toast

- 成功：绿色图标 + 单行结果描述，2 秒后自动消失
- 失败：红色图标 + 单行错误摘要，不自动消失，提供「查看详情」链接
- 长时间任务（DNS 发布、证书申请）：初始 Toast 变为进度指示，完成后转为结果 Toast

### 权限缺失提示

用户无权限访问当前页面：显示 403 状态页（非跳转到 /login），提示「当前角色 (X) 无权访问此功能，联系管理员」。操作级别无权限（如只读用户点击编辑）：操作按钮置灰，hover 时显示 tooltip「当前角色无权限执行此操作」。

---

## State Patterns

### 加载状态

| 场景 | 组件 | 行为 |
|------|------|------|
| 表格数据加载 | Skeleton Table | 展示 N 行骨架（N=上次加载行数或默认 5）|
| 卡片数据加载 | Skeleton Card | 轮廓骨架，shimmer 动效 |
| 操作提交中 | 操作按钮 disabled + Spin | 按钮文案变为「提交中…」|
| 页面首次加载 | 全页 Spin | 居中加载动画，< 500ms 不显示（防闪烁）|

### 空状态

每个列表页和功能区域必须有对应空状态。模式：

```
   [图标]
   主标题：描述为什么为空
   副标题：引导下一步操作
   [主操作按钮]（若有）
```

图标使用淡灰色线条图标（`{colors.text-tertiary}`）。常见文案：

| 场景 | 主标题 | 副标题 |
|------|--------|--------|
| 未添加 DNS Provider | 还没有 DNS 服务商 | 添加第一个 DNS 服务商开始管理 DNS 记录 |
| 未添加域名 | 还没有托管域名 | 添加域名并绑定 DNS 服务商账号 |
| 筛选无结果 | 没有匹配的结果 | 尝试修改筛选条件 |
| 未触发告警 | 暂无告警记录 | 系统运行正常 |

### 错误状态

| 类型 | 展示方式 |
|------|---------|
| API 加载失败 | 表格/卡片区域显示错误 Banner（红色背景，错误摘要 + 重试按钮）|
| Provider API 凭证失效 | 账号状态标签变为红色「凭证失效」，悬浮提示操作建议 |
| 任务执行失败 | 任务行状态变为红色 FAILED，可展开查看日志 |
| 网络断连 | 顶栏连接状态指示红点 + Toast 通知 |

### 确认对话框——交互规范

**标准确认（见 DESIGN.md）：**
- 标题 + 描述 + 取消/确认
- Enter 键触发确认（不适用于高危操作）

**高危确认（见 DESIGN.md components.modal-confirm-danger）：**
- 标题前置「⚠ 高危操作」
- 危险描述列表（影响范围）
- 文本输入验证（用户需输入目标名称）
- 确认按钮默认 disabled，输入匹配后激活
- 不绑定 Enter 键（防止误触）

**应用场景：**
- NS 切换 → 高危确认（输入域名）
- 批量 DNS 记录删除 → 高危确认（输入「确认删除」或域名）
- 删除域名绑定 → 高危确认（输入域名）
- 证书强制替换 → 高危确认（输入证书域名）
- 删除账号 → 标准确认
- 触发 Dry Run → 无需确认

---

## Interaction Primitives

### 导航展开/折叠

- 折叠态：宽度 `{spacing.nav-width-collapsed}` = 48px，显示 icon，hover 显示 tooltip 显示菜单文字
- 展开态：宽度 `{spacing.nav-width-expanded}` = 200px，显示 icon + 文字
- 切换：点击导航底部折叠按钮，状态持久化到 localStorage
- 二级菜单：在展开态显示缩进列表，在折叠态显示悬浮面板（fly-out）

### 面包屑导航

TopBar 中展示当前路径，格式：`模块 / 子模块 / 资源名`。子模块层级单击可跳转，当前页不可点击（非链接样式）。

### 表格行内操作

- 编辑/查看按钮：hover 时出现，不占列宽
- 超过 3 个操作：折叠到「更多」Dropdown
- 危险操作（删除）：在 Dropdown 末尾，使用 `{colors.danger}` 颜色

### 抽屉 (Drawer) vs 对话框 (Modal)

| 场景 | 组件 | 原因 |
|------|------|------|
| 表单（添加/编辑账号、节点、域名） | 右侧 Drawer（480px）| 表单内容较多，Drawer 有更多空间 |
| 快速预览（DNS 记录详情、告警详情） | 右侧 Drawer（640px）| 展示详细信息，不打断上下文 |
| 危险操作确认 | Modal（480px）| 居中焦点，强调决策性质 |
| 进度展示（工作流执行） | Modal（640px）| 居中展示进度，可关闭 |
| 工作流 DAG 编辑器 | 全页（独立路由）| 画布操作需要全屏空间 |

### 任务进度追踪

长时间任务（DNS 发布、SSL 续期、工作流执行）启动后：
1. 右下角 Toast：「任务已提交，正在执行…」+ 「查看详情」链接
2. 点击「查看详情」打开进度 Modal 或跳转到日志页
3. 进度条或步骤指示器展示当前执行步骤
4. 完成后自动推送结果通知（成功/失败）

---

## Responsive & Platform

### 桌面（≥1280px）

- 标准双栏布局（Nav + Content）
- 导航默认展开
- 表格显示所有列（含操作列）
- 表单使用双列布局（label 左对齐，input 右侧）
- 仪表盘统计卡片：4 列 → 适配宽度

### 平板（768–1279px）

- 导航默认折叠（icon-only，48px）
- 表格减少次要列（隐藏"更新时间""操作人"等低频列）
- 表单切换为单列布局
- 仪表盘统计卡片：2 列

### 移动端（<768px）

- 导航完全隐藏，顶栏显示汉堡按钮（触发侧滑抽屉式导航）
- 表格换为卡片列表视图（每行数据折叠为卡片）
- 高危操作（NS 切换、批量删除）不展示（移动端不允许执行破坏性操作）
- 仪表盘统计卡片：1 列
- 只读功能（查看域名列表、查看监控数据、查看告警）完整可用

---

## Accessibility Floor

- **颜色对比度：** 正文文字与背景对比度 ≥ 4.5:1（WCAG AA）；图标/控件 ≥ 3:1（见 DESIGN.md 状态标签设计）
- **键盘可访问：** 所有交互元素（按钮、链接、表单项、模态框）可通过 Tab/Shift+Tab 导航；焦点环可见（Arco 默认提供）
- **模态框焦点陷阱：** 模态框打开时焦点锁定在模态框内，关闭后焦点回到触发元素
- **表格操作：** 操作按钮的 aria-label 包含行标识（如 `aria-label="编辑域名 jdmv.net"`）
- **错误提示：** 表单验证错误与输入框通过 aria-describedby 关联
- **颜色非唯一信息载体：** 状态标签同时含颜色 + 文字（不仅靠颜色区分状态）

---

## Key Flows（用户旅程映射）

### KF-1：首次添加 DNS Provider 并配置域名（映射 UJ-1）

**路径：** `/dns/providers` → 点击「添加服务商」→ 选择阿里云 DNS → Drawer 表单（填写名称/AccessKeyId/AccessKeySecret）→ 点击「验证连接」→ 加载态（2–5 秒）→ 验证成功（Drawer 下方绿色 Banner：账号已连接，套餐：专业版，支持分运营商线路）→ 点击「保存」→ Drawer 关闭，列表出现新行 →

继续：`/dns/domains` → 点击「添加域名」→ Drawer 表单（填写域名/选择 Provider 账号/选择迁移方式）→ 选择「从 Provider 自动同步」→ 保存后进入 `/dns/domains/:id` → 域名详情页展示同步进度 → 完成后展示 DNS 记录列表 →

**关键交互点：**
- 验证连接按钮独立（不与保存合并），方便用户测试凭证
- 套餐能力报告在验证成功后立即展示（不等到保存）
- 若套餐不支持分运营商线路，能力报告显示警告：「当前套餐（免费版）不支持三网调度，升级到企业版或使用支持的 Provider」

**关键状态：**
- 验证失败：Drawer 内红色 Alert，含具体 API 错误码和建议操作
- 保存成功：Toast「DNS 账号已添加」

---

### KF-2：查看三网测速结果和调度状态（映射 UJ-2）

**路径：** `/scheduling/dashboard` → 仪表盘展示三网调度状态概览（电信/联通/移动各线路当前 Published Set、上次切换时间、当前评分 Top IP）→

点击某线路「查看详情」→ `/scheduling/ip-pool?line=telecom` → IP 候选池列表，展示每个 IP 的：当前状态标签（Published/Eligible/Standby/Offline）、各探针的延迟/TTFB/成功率热图 →

选择某 IP → 右侧抽屉展示该 IP 的历史测速趋势（ECharts 折线图，可选时间范围 1h/6h/24h）、当前三网评分详情 →

返回仪表盘 → 点击「调度历史」→ `/scheduling/history` → 时间线展示每次切换事件（切换原因/切换前后 IP/执行耗时）

**关键状态：**
- 自动调度已暂停（熔断触发）：页面顶部显示警告 Banner「电信线路自动调度已暂停（触发熔断：30 分钟内切换 3 次）」+ 「手动恢复」按钮
- 探针离线：IP 池中该探针列标记为 N/A，不影响其他探针的评分

---

### KF-3：SSL 证书续期工作流执行（映射 UJ-3）

**路径：** `/ssl/monitoring` → 证书到期提醒列表，jdmv.net 证书剩余 28 天，状态「即将到期」→

点击「查看工作流」→ `/ssl/workflows` → 找到对应工作流 → 点击「查看上次执行」→ `/ssl/execution-logs/:runId` →

工作流执行日志页：左侧为 DAG 拓扑（已执行节点绿色高亮，当前节点蓝色脉冲，失败节点红色）；右侧为当前选中节点的日志输出（实时滚动）→

执行完成：所有节点绿色 → 顶部 Banner「jdmv.net 证书已成功续期，新证书有效期至 YYYY-MM-DD」→

点击「查看证书」→ `/ssl/certificates/:id` → 证书详情页展示新证书信息和部署记录

**关键状态：**
- DNS-01 验证等待传播：等待节点变为蓝色「等待 DNS 传播」，倒计时显示已等待时间
- 某部署目标失败（如阿里云 CDN API 超时）：该分支节点红色，右侧展示 API 错误详情，「从失败节点重试」按钮

---

### KF-4：DNS 迁移操作（映射 UJ-4）

**路径：** `/dns/migrations` → 点击「新建迁移」→ Drawer（选择域名/源 Provider/目标 Provider）→ 创建后进入迁移详情页 →

迁移详情页分步展示：Step 1（发现现有记录）→ Step 2（生成迁移计划）→ Step 3（预同步目标 Provider）→ Step 4（完整性检查）→ **Step 5：人工审批（NS 切换）** →

Step 5 展示：当前 NS 记录、目标 NS 记录、切换前完整性报告（所有记录是否同步完成）。操作按钮：「我已在域名注册商修改了 NS，标记为已切换」→

系统进入 Step 6（监测 NS 委派变化，轮询更新）→ 委派成功 → Step 7（迁移完成标记）

**关键状态：**
- 完整性检查发现缺失记录：Step 4 显示差异列表，阻止进入 Step 5，提供「自动补全缺失记录」按钮
- NS 切换后超过 72 小时未检测到委派变化：显示告警，提供诊断建议

---

### KF-5：配置 Cloudflare SaaS Custom Hostname（映射 UJ-5）

**路径：** `/cdn/cloudflare/saas` → 点击「添加 Custom Hostname」→ Drawer（4 步向导）：

**Step 1 — 基础配置：** 选择 Cloudflare 账号 / Zone → 验证账号权限（Auto：显示检测结果，Custom Hostname 配额剩余 N 个）

**Step 2 — 源站配置：** 填写 Fallback Origin（`origin.wueym.com`）→ 点击「验证源站」→ 显示 HTTPS 检查结果（HTTP 状态、TLS 证书、响应时间）→ 通过则绿色 ✓

**Step 3 — 主机名配置：** 填写 Custom Hostname（`www.jdmv.net`）→ 选择 DCV 类型（DNS TXT 验证）→ 选择 TXT 记录添加方式（自动/手动）→ 自动：选择绑定的 DNS Provider 账号自动添加 → 手动：展示 TXT 记录内容 + 复制按钮

**Step 4 — 验证等待：** 展示主机名验证状态（轮询更新）→ 通过后展示成功状态、CNAME 配置指引（`www.jdmv.net CNAME edge.wueym.com`）→ 完成

**关键状态：**
- Fallback Origin 不可访问：Step 2 显示红色错误，阻止进入 Step 3
- DCV 验证超时：Step 4 展示诊断信息（DNS 传播进度、已等待时间），提供「重试 DCV」按钮
- **UI 明确提示（Step 1）：** 「标准 Cloudflare SaaS 仅支持 CNAME 接入。若需 A 记录接入，请联系 Cloudflare 确认您的套餐是否支持 Custom Origin 功能。」

---

## 页面规范（逐页）

以下为所有 55+ 页面的设计规范。格式：页面名 → 用途 → 主要组件 → 关键状态。

---

### 安装向导（/setup）

**用途：** 首次启动时的步骤式配置向导。无左侧导航，无顶栏，居中布局。

**布局：** 步骤条（顶部，9 步）+ 内容区（居中卡片）+ 底部操作按钮（上一步/下一步/完成）

**步骤：**
1. 欢迎 + 环境检测（自动执行，逐项展示 PASS/FAIL）
2. 数据库配置（表单：Host/Port/Database/User/Password + 测试连接按钮）
3. Redis 配置（表单 + 测试连接）
4. 执行数据库 Migration（进度展示，自动执行）
5. 创建管理员账号（用户名/密码/确认密码，密码强度指示器）
6. 应用加密密钥（生成随机密钥建议 + 自定义输入）
7. 配置访问地址（后台 URL 配置）
8. 确认配置摘要（汇总展示，可返回修改）
9. 安装完成（成功状态页，跳转登录按钮）

**关键状态：**
- 环境检测 FAIL（如 MySQL 连接失败）：当前步骤阻止前进，展示详细错误和修复建议
- Migration 执行失败：步骤 4 展示错误日志，提供重试按钮，不写入 install.lock
- 安装完成后：所有 `/api/v1/setup/*` 接口返回 403

---

### 登录页（/login）

**用途：** 用户名 + 密码登录。

**布局：** 左侧品牌区域（DESIGN.md 主色背景，产品 Logo 和 Tagline）+ 右侧登录表单（白色卡片）

**关键状态：**
- 登录失败：表单上方红色 Alert（错误摘要，不区分"用户名错误"还是"密码错误"）
- 账号锁定：显示剩余解锁时间，不显示锁定原因
- 密码输入框：右侧图标切换可见/不可见

---

### 仪表盘（/dashboard）

**用途：** 系统全局状态一览，快速发现问题。

**布局：**
- 第一行：4 个统计卡片（域名总数、在线节点数、有效证书数、未解决告警数）
- 第二行：DNS 服务状态列表（Provider 账号连接状态）+ CDN 节点状态概览（在线/离线节点环形图）
- 第三行：三网性能趋势折线图（ECharts，电信/联通/移动三条线，时间选择器：1h/6h/24h/7d）
- 第四行：SSL 证书状态列表（即将到期前 5 条）+ 最近告警（最新 10 条，按等级分色）

**组件：**
- Arco Grid（统计卡片 4 列，折叠为 2 列在平板）
- Arco Tag（告警等级：CRITICAL=红，WARNING=橙，INFO=蓝）
- ECharts Line Chart（三网趋势）
- ECharts Donut Chart（节点状态）
- Arco Table（SSL 到期列表）
- SSE 实时更新（告警和状态数字自动刷新）

**关键状态：**
- 加载中：所有卡片和图表显示骨架屏
- 无数据（新安装）：统计卡片显示 0，图表显示空状态（「暂无监控数据，添加域名和节点后自动采集」）
- 告警触发：未解决告警数 Badge 变红，最近告警列表展示新告警

---

### DNS 服务商（/dns/providers）

**用途：** 管理支持的 DNS Provider 类型配置。

**布局：** 页头（标题 + 描述文字）+ Provider 卡片网格（每个 Provider 一张卡片：Logo + 名称 + 状态 + 账号数量 + 操作）

**关键状态：**
- Provider 卡片状态：「已支持」（蓝色 Badge）、「二期计划」（灰色 Badge）
- 点击已支持 Provider 卡片 → 跳转到 `/dns/accounts?provider=xxx` 筛选该 Provider 账号

---

### DNS 账号（/dns/accounts）

**用途：** 管理多个 DNS Provider 的 API 账号凭证。

**布局：** 工具栏（Provider 筛选 + 状态筛选 + 添加按钮）+ 表格

**表格列：** 账号名、Provider（Logo + 名称）、状态标签（已连接/凭证失效/检测中）、绑定域名数、上次检测时间、API 调用次数（今日）、操作（编辑/检测连接/禁用/删除）

**Drawer（添加/编辑）：**
- Provider 选择（单选）
- 账号名（自定义标签）
- API 凭证字段（根据 Provider 动态变化，如阿里云需要 AccessKeyId + Secret，Cloudflare 需要 API Token）
- 凭证输入框使用 password 类型（`type="password"`）
- 保存前必须点击「验证连接」按钮，验证通过后才能保存
- 验证结果展示：账号连接状态 + 套餐类型 + 能力列表（线路/最小 TTL/记录条数上限）

**关键状态：**
- 验证失败：内联 Alert（具体 API 错误码 + 文档链接），保存按钮保持禁用
- 凭证失效（已保存账号）：状态标签红色 + 表格行 hover 显示「凭证已失效，点击编辑重新配置」

---

### 域名管理（/dns/domains）

**用途：** 管理所有托管域名的 DNS 配置绑定。

**表格列：** 域名、权威 Provider（账号名）、备用 Provider、三网调度状态（已启用/未满足条件/未启用）、记录数量、上次同步时间、ICP 备案状态标签（已备案/未备案）、操作

**重要交互：**
- ICP 备案状态用于控制可用 CDN 选项（未备案域名禁止配置大陆 CDN，行内 tooltip 说明）
- 未满足调度条件的域名，状态 tooltip 说明具体缺失条件（如 Provider 套餐不支持）

---

### 域名详情（/dns/domains/:id）

**用途：** 单个域名的完整 DNS 配置视图。

**布局：** 顶部域名信息栏（域名名称 + Provider 信息 + 状态）+ Tab 标签页切换

**Tab 标签：**
1. **DNS 记录** — 表格列（类型/主机名/值/TTL/线路/状态/操作），按类型分组，支持筛选/搜索
2. **套餐能力** — Provider 能力报告卡片（支持的线路/最小TTL/记录上限/权重支持/健康检查支持）
3. **智能调度** — 此域名的调度策略摘要 + 候选池快览
4. **证书关联** — 此域名关联的证书列表
5. **历史快照** — 此域名的快照时间线

**DNS 记录操作：**
- 添加记录 → Drawer 表单（类型/主机名/值/TTL/线路，线路下拉根据 Provider 能力动态填充）
- 批量操作：勾选记录 → 工具栏出现「批量删除」「批量修改 TTL」按钮（批量删除触发高危确认）
- SmartCDN 未托管的记录：行标记「外部记录」灰色 Badge，操作列无删除按钮（只允许查看）
- Dry Run 开关：全局开关在域名详情页顶部，开启后所有操作进入预检模式，不实际提交

---

### DNS 迁移（/dns/migrations）

**用途：** 执行域名从一个 DNS Provider 迁移到另一个 Provider 的完整工作流。

**列表页：** 迁移任务列表（域名/源Provider/目标Provider/状态/当前步骤/创建时间/操作）

**详情页（步骤式）：**
步骤条展示 7 步（见 KF-4 中描述），当前步骤高亮。每步卡片包含：当前步骤说明、执行结果、下一步操作按钮。

**Step 5 — 人工授权 NS 切换 UI：**
```
当前 NS:  ns1.aliyun.com  ns2.aliyun.com
目标 NS:  mia.ns.cloudflare.com  nash.ns.cloudflare.com

✅ 记录同步完整 (127/127 条记录已同步)
✅ 完整性检查通过

⚠ 下一步需要您在域名注册商处修改 NS 记录。
   此操作将影响域名的全部 DNS 解析，建议在低峰期执行。
   
[我已修改 NS，开始监测委派变化]  [稍后操作]
```

---

### Cloudflare Zones（/cdn/cloudflare/zones）

**用途：** 管理 Cloudflare 账号下的 Zone 列表，查看和操作 Zone 配置。

**表格列：** Zone 名称、账号、状态（Active/Paused）、套餐、Proxy 记录数、上次缓存清理时间、操作（进入详情/清理缓存）

**Zone 详情页（Tab 切换）：**
1. **DNS 记录** — 通过 Cloudflare 适配器展示和管理记录，区分 Proxied/DNS-only
2. **缓存规则** — Cache Rules 列表（读取/配置）
3. **SSL/TLS** — 当前 SSL 模式（Full/Full(strict)/Flexible）
4. **安全配置** — 安全设置状态读取（DDoS/防火墙规则摘要）
5. **性能** — 请求量/缓存命中率趋势（Cloudflare API 提供时）

---

### Cloudflare SaaS（/cdn/cloudflare/saas）

**用途：** 管理 Cloudflare for SaaS 的 Custom Hostname 列表。

**表格列：** 主机名、所属 Zone、验证状态（待验证/验证中/已验证/失败）、SSL 状态（颁发中/有效/即将到期/已过期）、CNAME Target、创建时间、操作

**重要提示 Banner（页面顶部）：**
> 「Cloudflare for SaaS 标准模式通过 CNAME 接入。A 记录直接接入需要 Cloudflare Custom Origin 功能，请确认您的套餐支持后操作。」

**Custom Hostname 详情 Drawer：**
- 验证状态详情（DCV TXT 记录值、传播状态）
- SSL 证书状态（颁发中/有效期/证书类型）
- 失败诊断（若状态为 Failed，展示具体原因）
- 重新触发验证按钮

---

### 自有节点（/cdn/nodes）

**用途：** 管理自有家宽和服务器节点。

**表格列：** 节点名称、IP（v4/v6）、地区、ISP 标识（电信/联通/移动/香港/日本等）、Node Agent 状态（在线/离线/未安装）、CPU/内存（来自 Agent 上报）、HTTPS 检查状态、所属资源池、权重、操作

**节点详情页（Tab 切换）：**
1. **基础信息** — 节点配置（回源地址/Host/SNI/WebSocket 支持/缓存配置）
2. **性能监控** — CPU/内存/网络吞吐量趋势图（ECharts）
3. **HTTPS 状态** — 当前证书信息、外部 HTTPS 检查结果
4. **缓存规则** — 该节点的缓存策略配置
5. **Agent 日志** — Node Agent 最近日志（实时 SSE）

**重要交互：**
- 添加节点时必须勾选「我确认拥有或已获授权使用此节点」
- Node Agent 安装指引：节点详情页展示一键安装命令（含认证 Token）

---

### 探针管理（/scheduling/probes）

**用途：** 管理分布式 Probe Agent 节点。

**表格列：** 探针名称、部署位置（地区/ISP）、Agent IP（实际注册 IP）、在线状态、上次心跳时间、探测任务数、成功率（近 1 小时）、操作

**探针详情 Drawer：**
- 注册信息（ISP 标识/地区/注册时间）
- 探测配置（目标 IP 列表/频率/超时）
- 近期探测结果统计（成功率/平均延迟）
- 下发配置更新按钮

**重要说明 Banner：**
> 「Probe Agent 必须真实部署于对应运营商网络中才能反映真实网络性能。请勿在同一网络部署多个标记为不同运营商的探针。」

---

### 优选 IP 池（/scheduling/ip-pool）

**用途：** 展示和管理各域名各线路的 IP 候选池及当前状态。

**筛选器：** 域名选择 + 线路选择（全部/电信/联通/移动/境外）

**IP 列表：** 每个 IP 一行，列：IP 地址、池状态（Published/Eligible/Standby/Offline，分色标签）、当前评分、各探针延迟/TTFB/成功率（热图单元格，颜色深浅代表优劣）、上次更新时间、手动操作

**手动操作 Dropdown：**
- 锁定为当前线路（固定使用此 IP，绕过自动调度）
- 强制发布到 DNS
- 临时排除（冷却 N 小时）
- 移出候选池

**锁定状态展示：** 被手动锁定的 IP 行显示「🔒 已锁定」徽章，解锁按钮

---

### 证书工作流（/ssl/workflows）

**用途：** 管理 SSL 证书的申请/续期/部署工作流。

**列表页：** 工作流名称、关联证书（域名）、触发方式（定时/自动续期/手动）、状态（启用/禁用）、上次执行结果、下次执行时间、操作

---

### 工作流 DAG 编辑器（/ssl/workflows/:id/edit）

**用途：** 可视化编辑证书工作流 DAG（全页编辑器）。

**布局：**
- 顶部工具栏：工作流名称（可编辑）+ 保存/另存为/运行/撤销/重做
- 左侧节点面板（可用节点类型，拖拽到画布）
- 中央画布（DAG 拓扑，节点连线）
- 右侧属性面板（选中节点的配置表单）

**画布交互：**
- 拖拽添加节点
- 节点间连线（从输出端口拖到输入端口）
- 右键节点：编辑/删除/设置条件分支
- 画布缩放（鼠标滚轮/触控板）
- 自动对齐（启用/禁用）

**节点类型视觉区分：**
- 触发节点（蓝色圆角矩形）
- 操作节点（白色边框矩形）
- 条件节点（菱形，橙色边框）
- 并行节点（平行四边形）
- 通知节点（消息气泡形状）
- 结束节点（圆形，绿色）
- 失败节点（圆形，红色）

---

### 告警记录（/monitoring/alerts）

**用途：** 查看所有告警历史，支持确认和静默操作。

**表格列：** 告警等级（CRITICAL/WARNING/INFO 图标）、告警规则名称、触发对象（域名/节点/证书）、告警内容、触发时间、持续时长、当前状态（触发中/已恢复/已确认）、操作（确认/静默 N 分钟）

**告警详情 Drawer：**
- 告警规则详情（触发条件/当前指标值/历史趋势小图）
- 相关对象快速跳转链接
- 告警时间线（触发→更新→恢复/确认 记录）
- 操作：确认告警/设置静默/查看相关日志

---

### 审计日志（/admin/audit-logs）

**用途：** 查看所有用户和系统操作的完整审计记录，不可修改或删除。

**工具栏筛选：** 操作者（用户选择）、操作类型（DNS变更/证书操作/账号管理等）、时间范围、目标资源类型

**表格列：** 时间戳、操作者（用户名 + IP）、操作类型、目标资源（可点击跳转）、操作摘要、详情展开按钮

**详情展开（行内 Expand 或 Drawer）：** 展示操作前后状态的 JSON diff（使用代码字体 `{typography.code}`，新增行绿色高亮，删除行红色高亮）

---

### 系统诊断（/admin/diagnostics）

**用途：** 实时检测系统组件健康状态，辅助排查问题。

**布局：** 多个诊断卡片，每张卡片对应一个组件：MySQL 连接/Redis 连接/各 DNS Provider API 可访问性/各 Probe Agent 在线状态/证书 ACME CA 可达性

**每张卡片：** 组件名 + 状态指示（绿/橙/红）+ 检测时间 + 详情（延迟/错误信息）+ 重新检测按钮

---

## 设计规范补充

### 色彩主题切换

TopBar 右侧提供主题切换按钮（太阳/月亮图标），当前选择持久化到 localStorage。浅色/深色主题通过 Arco Design `ConfigProvider` 的 `theme` 属性切换，所有 `{colors.*-dark}` 令牌在深色模式下替换对应的浅色令牌。

### 国际化（i18n）

默认语言：简体中文。TopBar 右侧提供语言切换（中/英），切换后页面内容（菜单、标签、按钮、错误提示）切换语言，数据内容（域名、IP、凭证名）不翻译。语言选择持久化到 localStorage。

### 数据脱敏展示规范

API Key/Secret：始终展示 `****` + 末 4 位（如 `****ABCD`）。
证书私钥：在证书详情页不展示，提供「导出私钥」操作（需 2FA 确认或密码确认）。
密码：表单中始终使用 `type="password"`。
