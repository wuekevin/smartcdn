---
name: SmartCDN Design System
description: Professional cloud infrastructure management console — Arco Design React extended with SmartCDN brand tokens.
status: final
created: 2026-11-09
updated: 2026-11-09

colors:
  # --- Primary palette ---
  primary: '#165DFF'
  primary-hover: '#4080FF'
  primary-pressed: '#0E42D2'
  primary-disabled: '#94BFFF'

  # --- Light theme surfaces ---
  bg-page: '#F2F3F5'
  bg-container: '#FFFFFF'
  bg-container-secondary: '#F7F8FA'
  bg-nav: '#001529'
  bg-nav-hover: '#002040'
  bg-nav-active: '#165DFF'
  bg-mask: 'rgba(0,0,0,0.6)'

  # --- Dark theme surfaces ---
  bg-page-dark: '#17171A'
  bg-container-dark: '#232324'
  bg-container-secondary-dark: '#2A2A2B'
  bg-nav-dark: '#141414'
  bg-nav-hover-dark: '#1E1E1E'

  # --- Neutral text ---
  text-primary: '#1D2129'
  text-secondary: '#4E5969'
  text-tertiary: '#86909C'
  text-disabled: '#C9CDD4'
  text-on-nav: 'rgba(255,255,255,0.65)'
  text-on-nav-active: '#FFFFFF'

  # --- Dark theme text ---
  text-primary-dark: '#F5F6F7'
  text-secondary-dark: '#9DA3AC'
  text-tertiary-dark: '#6B6F77'

  # --- Semantic ---
  success: '#00B42A'
  success-bg: '#E8FFEA'
  success-bg-dark: '#0D2B14'
  warning: '#FF7D00'
  warning-bg: '#FFF7E8'
  warning-bg-dark: '#2B1D00'
  danger: '#F53F3F'
  danger-bg: '#FFECE8'
  danger-bg-dark: '#2B0E0E'
  info: '#165DFF'
  info-bg: '#E8F4FF'
  info-bg-dark: '#0A1929'

  # --- Borders ---
  border: '#E5E6EB'
  border-secondary: '#F2F3F5'
  border-dark: '#3A3A3B'
  border-secondary-dark: '#2D2D2E'

  # --- Special ---
  link: '#165DFF'
  link-hover: '#4080FF'
  scrollbar: '#C9CDD4'
  highlight: '#FFF3C8'

typography:
  display:
    fontFamily: "'PingFang SC', 'Helvetica Neue', Arial, 'Microsoft YaHei', sans-serif"
    fontSize: '28px'
    fontWeight: '600'
    lineHeight: '1.4'
  h1:
    fontFamily: "'PingFang SC', 'Helvetica Neue', Arial, 'Microsoft YaHei', sans-serif"
    fontSize: '20px'
    fontWeight: '600'
    lineHeight: '1.5'
  h2:
    fontSize: '16px'
    fontWeight: '600'
    lineHeight: '1.5'
  h3:
    fontSize: '14px'
    fontWeight: '600'
    lineHeight: '1.57'
  body:
    fontSize: '14px'
    fontWeight: '400'
    lineHeight: '1.57'
  body-sm:
    fontSize: '12px'
    fontWeight: '400'
    lineHeight: '1.67'
  code:
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Courier New', monospace"
    fontSize: '13px'
    fontWeight: '400'
    lineHeight: '1.6'
  label:
    fontSize: '12px'
    fontWeight: '500'
    lineHeight: '1.67'
    letterSpacing: '0.02em'
  caption:
    fontSize: '12px'
    fontWeight: '400'
    lineHeight: '1.67'
    color: '{colors.text-secondary}'
  stat:
    fontSize: '24px'
    fontWeight: '600'
    lineHeight: '1.33'
  stat-sm:
    fontSize: '20px'
    fontWeight: '600'
    lineHeight: '1.4'

rounded:
  sm: '2px'
  md: '4px'
  lg: '6px'
  xl: '8px'
  card: '4px'
  modal: '4px'
  badge: '2px'
  tag: '2px'
  button: '2px'
  full: '9999px'
  DEFAULT: '4px'

spacing:
  '1': '4px'
  '2': '8px'
  '3': '12px'
  '4': '16px'
  '5': '20px'
  '6': '24px'
  '8': '32px'
  '10': '40px'
  '12': '48px'
  page-padding: '24px'
  card-padding: '20px'
  section-gap: '16px'
  form-gap: '20px'
  table-cell-h: '12px'
  table-cell-v: '8px'
  nav-width-collapsed: '48px'
  nav-width-expanded: '200px'
  topbar-height: '48px'
  breadcrumb-height: '40px'

components:
  button-primary:
    background: '{colors.primary}'
    background-hover: '{colors.primary-hover}'
    background-pressed: '{colors.primary-pressed}'
    color: '#FFFFFF'
    borderRadius: '{rounded.button}'
    height-default: '32px'
    height-sm: '24px'
    height-lg: '40px'
    padding: '0 16px'
  button-danger:
    background: '{colors.danger}'
    background-hover: '#D93025'
    color: '#FFFFFF'
    borderRadius: '{rounded.button}'
  button-ghost:
    background: 'transparent'
    border: '1px solid {colors.border}'
    color: '{colors.text-primary}'
    borderRadius: '{rounded.button}'
  status-badge:
    borderRadius: '{rounded.full}'
    padding: '1px 8px'
    fontSize: '{typography.body-sm.fontSize}'
  table:
    rowHeight: '54px'
    rowHeightSm: '40px'
    headerBg: '{colors.bg-container-secondary}'
    stripeBg: '{colors.bg-page}'
    hoverBg: '{colors.info-bg}'
    borderColor: '{colors.border}'
  card:
    background: '{colors.bg-container}'
    borderRadius: '{rounded.card}'
    border: '1px solid {colors.border}'
    padding: '{spacing.card-padding}'
    shadow: '0 1px 2px rgba(0,0,0,0.04)'
  nav-item:
    height: '40px'
    color-inactive: '{colors.text-on-nav}'
    color-active: '{colors.text-on-nav-active}'
    bg-active: '{colors.bg-nav-active}'
    bg-hover: '{colors.bg-nav-hover}'
    borderRadius: '0'
    paddingLeft: '24px'
  form-item:
    labelWidth: '100px'
    gapV: '{spacing.form-gap}'
    helpTextColor: '{colors.text-secondary}'
    errorColor: '{colors.danger}'
  modal-confirm-danger:
    headerColor: '{colors.danger}'
    confirmBtnVariant: 'danger'
    inputPlaceholder: '请输入名称以确认'
  stat-card:
    background: '{colors.bg-container}'
    valueColor: '{colors.text-primary}'
    labelColor: '{colors.text-secondary}'
    iconSize: '24px'
  tag-green:
    background: '{colors.success-bg}'
    color: '{colors.success}'
    borderRadius: '{rounded.tag}'
  tag-orange:
    background: '{colors.warning-bg}'
    color: '{colors.warning}'
    borderRadius: '{rounded.tag}'
  tag-red:
    background: '{colors.danger-bg}'
    color: '{colors.danger}'
    borderRadius: '{rounded.tag}'
  tag-blue:
    background: '{colors.info-bg}'
    color: '{colors.primary}'
    borderRadius: '{rounded.tag}'
  tag-gray:
    background: '{colors.bg-page}'
    color: '{colors.text-secondary}'
    borderRadius: '{rounded.tag}'
---

## Brand & Style

SmartCDN 是一套面向运维工程师和技术团队的私有化部署管理控制台。视觉语言以阿里云控制台和 Cloudflare Dashboard 为参照系——信息密度高、交互精确、无多余装饰。用户的工作对象是数据和操作，UI 应退到背景。

设计基调：**专业 · 权威 · 克制**。颜色用于传达状态和层级，不用于装饰。每个像素都要为信息密度和操作清晰度服务。系统同时支持深色和浅色主题，深色主题是数据密集视图下的首选。

UI 体系：**Arco Design React**（字节跳动出品的企业级设计体系）。SmartCDN 在 Arco 基础上定义品牌色令牌覆盖，不重新实现 Arco 组件，只通过 `ConfigProvider` 修改主题令牌。

---

## Colors

**主色 (Primary Blue)：** `{colors.primary}` — Arco 标准蓝，用于主操作按钮、选中态、链接、高亮。避免用于背景大面积填充。在深色主题中保持相同值（深蓝背景下可读性已验证）。

**导航背景 (Nav Dark)：** `{colors.bg-nav}` — 深色海军蓝，始终保持深色（无论浅/深主题），形成内容区与导航区的强烈层级感。这是全局导航的主色调，参照阿里云控制台侧边栏风格。

**内容区背景：**
- 浅色主题：`{colors.bg-page}`（淡灰）+ `{colors.bg-container}`（白色容器）
- 深色主题：`{colors.bg-page-dark}`（深灰黑）+ `{colors.bg-container-dark}`（稍亮容器）

**语义色：**
- 成功绿 `{colors.success}`：DNS 解析正常、证书有效、节点在线。在浅色主题以 `{colors.success-bg}` 作底色衬托。
- 警告橙 `{colors.warning}`：即将到期（证书剩余 ≤ 30 天）、探针延迟偏高、DNS TTL 未传播。
- 危险红 `{colors.danger}`：证书已过期、节点离线、DNS 发布失败、API 凭证失效。
- 信息蓝 `{colors.info}`：与主色共用，用于提示性状态（非操作性）。

**禁止用法：** 不将危险红用于主操作按钮（除删除确认外）；不将多种语义色叠用于同一状态标签；不在导航区使用语义色背景。

---

## Typography

字体栈首选 PingFang SC（macOS/iOS 系统中文字体），回退到 Helvetica Neue / Arial / 微软雅黑。正文字号 14px，行高 1.57，满足中文阅读节奏。

代码块（DNS 记录值、API 凭证预览、命令行指令、JSON 配置）统一使用等宽字体 `{typography.code.fontFamily}`，背景使用 `{colors.bg-container-secondary}`，圆角 `{rounded.sm}`。

**数字密度：** 统计卡片和仪表盘数字使用 `{typography.stat}` 或 `{typography.stat-sm}`，配合 tabular-nums 特性保证对齐。

页面标题（h1）不超过 20px，保持密度控制台风格，区别于营销页面。

---

## Layout & Spacing

**全局布局：**
```
┌────────────────────────────────────────┐
│ TopBar  48px                           │
├─────────┬──────────────────────────────┤
│ SideNav │  Content Area               │
│ 200px   │  min-width: 980px           │
│(expand) │  padding: 16px 24px         │
│ 48px    │                             │
│(collapse│                             │
└─────────┴──────────────────────────────┘
```

- 顶部栏高度：`{spacing.topbar-height}` = 48px，固定定位
- 导航宽度展开：`{spacing.nav-width-expanded}` = 200px；折叠：`{spacing.nav-width-collapsed}` = 48px（仅显示 Icon）
- 内容区 padding：`{spacing.page-padding}` = 24px（上下左右）
- 卡片内边距：`{spacing.card-padding}` = 20px
- 卡片间距：`{spacing.section-gap}` = 16px
- 表单项间距：`{spacing.form-gap}` = 20px

**响应式断点：**
- 桌面（≥1280px）：标准双栏或三栏，导航展开状态为默认
- 平板（768–1279px）：导航默认折叠（icon-only），内容区适配减少列数
- 移动（<768px）：导航完全隐藏（汉堡按钮弹出），表格换为卡片列表视图，只读功能可访问，危险操作不显示

**内容宽度：** 表单和详情页最大宽度 760px，居中展示（不无限拉伸）；列表和仪表盘页铺满内容区。

---

## Elevation & Depth

层级规则（从低到高）：
1. **页面背景** — `{colors.bg-page}`，无阴影
2. **容器/卡片** — `{colors.bg-container}`，`shadow: 0 1px 2px rgba(0,0,0,0.04)` + border
3. **悬浮面板/Dropdown** — `shadow: 0 4px 10px rgba(0,0,0,0.12)`，border
4. **抽屉/侧边 Drawer** — `shadow: -4px 0 12px rgba(0,0,0,0.15)`
5. **模态框** — `{colors.bg-mask}` 遮罩 + `shadow: 0 8px 24px rgba(0,0,0,0.20)` + border
6. **Toast 通知** — 最高层，不受内容区遮挡，出现在右上角

深色主题阴影将 rgba 值调整：0.04→0.12，0.12→0.24，0.20→0.40（深色背景下阴影对比度需更强）。

---

## Shapes

圆角使用 `{rounded.DEFAULT}` = 4px（Arco Design 标准），形成整体方正感，区别于消费级产品的大圆角风格。

例外：
- 状态徽章 (Badge)：`{rounded.full}` = 9999px，以圆形为标志性形态
- 标签 (Tag)：`{rounded.tag}` = 2px，更方正
- 告警 Banner：`{rounded.sm}` = 2px，贴近页面边缘时更协调

---

## Components

### 状态标签 (Status Tag)

用于节点/证书/DNS 记录的状态展示。颜色语义见 Colors 章节。

| 状态 | 颜色令牌 | 典型文案 |
|------|---------|---------|
| 正常/在线 | `tag-green` | 正常、在线、已连接、有效 |
| 警告/待处理 | `tag-orange` | 即将到期、降级、待验证 |
| 错误/离线 | `tag-red` | 离线、已过期、失败、凭证失效 |
| 进行中 | `tag-blue` | 发布中、验证中、续期中 |
| 禁用/未知 | `tag-gray` | 已禁用、未知、N/A |

### 确认对话框——标准型

触发操作：删除账号、禁用账号、批量操作（非生产域名）。
- 标题：「确认删除 {名称}？」
- 副文本：描述操作不可逆后果
- 操作按钮：取消（`button-ghost`）+ 确认（`button-primary` 或 `button-danger`）
- 不要求用户重复输入名称

### 确认对话框——高危型 (Destructive Confirm)

触发操作：NS 切换授权、批量删除生产 DNS 记录、替换生产证书、删除域名绑定。
- 标题：用「⚠」前缀 + 危险红 `{colors.danger}` 标题色（见 `{components.modal-confirm-danger}`）
- 危害说明：明确列出影响范围（影响哪些域名、哪些服务）
- 文本输入框：要求用户准确输入目标名称（如域名）以激活确认按钮
- 操作按钮：取消（`button-ghost`）+ 确认（`button-danger`，默认禁用，输入正确后激活）

### 表格行操作

行内操作按钮（编辑/查看/更多）默认隐藏，行 hover 时出现。「更多」展开 Dropdown 包含低频操作（禁用/删除）。危险操作（删除）使用 `{colors.danger}` 文字色。

### 空状态 (Empty State)

每个列表页和功能区域均有独立空状态：居中图标（淡灰色线条图标）+ 主文案（告知原因）+ 副文案（操作提示）+ 操作按钮（若有）。

### 加载骨架 (Skeleton Loading)

表格和卡片列表使用骨架屏（非旋转加载器）。骨架屏背景色：`{colors.bg-container-secondary}`，动画：shimmer 左→右扫描。

### 全局搜索 (Global Search)

弹出搜索面板，覆盖内容区（不覆盖导航）。搜索结果分组：域名、IP、证书、节点、API 凭证。快捷键：`Cmd/Ctrl + K`。

---

## Do's and Don'ts

**Do：**
- 对所有状态（正常/警告/错误）使用语义色令牌，保持一致
- 在表格中对 IP 地址、证书指纹等技术值使用代码字体 `{typography.code}`
- 危险操作始终使用高危确认对话框（见 Components）
- 保持导航条背景色为深色，无论当前主题

**Don't：**
- 不将大量操作堆砌在单个页面（遵循 PRD §4.15 要求）
- 不使用明亮渐变背景（消费级风格，不适合运维控制台）
- 不省略空状态——每个列表必须有空状态设计
- 不在加载完成前显示 0 或占位数字（使用骨架屏）
- 不为低频操作（如删除）放置突出按钮位置（主操作区优先展示添加/编辑）
- 不在错误提示中展示"出错了"类模糊文案——必须给出具体错误原因
