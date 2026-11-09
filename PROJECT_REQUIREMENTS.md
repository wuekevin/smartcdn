# SmartCDN — 商业级多 DNS、智能 CDN 与 SSL 自动化管理平台

**项目类型：** 商业级 Web 管理平台

**项目名称：** SmartCDN

**开发方法：** BMad Method

**开发工具：** Claude Code / Codex

**主要开发语言：** Go + TypeScript

**主要部署环境：** Linux + 宝塔面板 Go 项目

**目标：** 完整源码交付、生产可用、可持续扩展、多服务商兼容

---

# 第一章：项目总体目标

## 1.1 项目背景

开发一套完全自主控制、可私有化部署的商业级 DNS、CDN、SSL 证书和智能线路调度综合管理平台。

SmartCDN 的主要用途包括：

1. 统一管理多个 DNS 服务商。
2. 管理多个云账号和多个网站域名。
3. 为中国大陆电信、联通、移动分别进行网络质量检测。
4. 对可用的 CDN 和自有反向代理节点进行智能选线。
5. 根据不同 DNS 服务商的套餐能力发布解析记录。
6. 管理 Cloudflare CDN 和 Cloudflare for SaaS。
7. 管理自有家宽 HTTPS 反向代理节点。
8. 自动申请、续期、部署 SSL 证书。
9. 统一管理跨平台 SSL 证书工作流。
10. 监测网站可用性、DNS 状态、HTTPS 性能和节点健康。
11. 自动识别故障、执行安全回滚及发送告警。
12. 为未来增加 DNS/CDN 服务商提供插件扩展能力。

系统必须具有专业、完整、统一的管理后台。

不得将产品降级为演示程序、MVP 或单纯的 DNS CRUD 系统。

## 1.2 开发原则

必须遵守：

- 使用 BMad Method 管理需求、架构、开发和验收。
- 实现真实可运行的完整业务。
- 所有第三方能力必须经过官方文档及真实 API 验证。
- 不得编造 DNS、CDN 或证书 API 能力。
- 不得使用模拟成功结果代替真实调用。
- 不得使用虚假的监控数据冒充实际测速。
- 不得留下无功能按钮、空实现或不可用的占位模块。
- 不得跳过测试、日志、安全和异常处理。
- 不得为了完成任务而破坏已有生产网站配置。
- 默认先进行只读发现、预检查及 Dry Run，再申请生产变更授权。
- 敏感或不可逆操作必须经过人工确认。
- 所有重要架构决策必须生成 ADR。
- 开发 Agent 应主动发现问题、实施修复并重新验证。

商业级要求不等于一次提交全部代码。必须通过可验证的开发阶段逐步交付，最终完成全部约定范围。

---

# 第二章：强制技术栈

## 2.1 后端

- Go 稳定版本。
- Gin Web Framework。
- MVC / Service / Repository 分层。
- MySQL。
- Redis。
- RESTful API。
- OpenAPI 3.1。
- RBAC 权限控制。
- WebSocket 或 SSE 实时事件推送。
- 数据库 Migration。
- 结构化日志。
- 后台任务调度系统。
- 分布式任务锁。
- 审计日志。

禁止引入 NestJS、Java、Python 或其他运行时来替代 Go 核心后端。

允许使用成熟的 Go 开源依赖，但必须进行许可证、安全和维护状态审查。

## 2.2 前端

- React。
- Vite。
- TypeScript。
- Arco Design React。
- React Router。
- TanStack Query。
- ECharts。
- 专业响应式管理后台。
- 中文默认，支持英文国际化。
- 深色和浅色主题。

前后端必须分离。

## 2.3 部署模式

### P0：主要部署方式

**Linux + 宝塔面板 Go 项目管理器**

必须支持：

- 宝塔创建 Go 项目。
- Go 二进制或受支持的项目运行方式。
- 宝塔 Nginx 网站反向代理。
- 宝塔 MySQL。
- 宝塔 Redis。
- 宝塔环境变量配置。
- Go 项目启动、停止、重启和更新。
- 宝塔网站 HTTPS。
- 首次安装向导。
- 日志和备份。

不能用 systemd 方案冒充宝塔 Go 项目管理方案。

### P1：次要部署方式

- 原生 Linux + Nginx + systemd。
- Docker。
- Docker Compose。

必须交付四种部署方式对应的文件和文档。

Docker 不允许成为宝塔 Go 项目部署的必要依赖。

---

# 第三章：首个业务网站与现网保护

## 3.1 初始业务域名

| 域名             | 用途                            | 当前环境                                   |
| ---------------- | ------------------------------- | ------------------------------------------ |
| www.jdmv.net     | 正式网站                        | 未备案，目前使用 Cloudflare                |
| origin.wueym.com | 计划中的 Cloudflare SaaS 回退源 | 需检测实际源站和 HTTPS                     |
| static.01233.xyz | 静态资源                        | 已备案，目前走阿里云全球加速，包含中国大陆 |
| cdnapi.01233.xyz | 独立 API                        | 已备案，目前走阿里云全球加速，包含中国大陆 |

上述域名只用于初始化。

系统必须支持后台增加其他业务域名，不能写死任何具体域名。

## 3.2 现网保护要求

开发前必须建立现网资产和配置基线，包括：

- 当前 NS。
- DNS 记录。
- CNAME。
- Cloudflare Zone。
- 阿里云加速产品类型。
- HTTPS 证书。
- 源站地址。
- 回源 Host。
- 缓存策略。
- API 路径。
- WebSocket 接入方式。
- 现有健康检查。
- 当前网站可用性。

没有相应账号或网络权限时，必须标记为待验证。

正式上线前不得未经授权：

- 切换 NS。
- 删除 DNS 记录。
- 修改生产 CDN。
- 修改源站。
- 替换 SSL 证书。
- 停止 Nginx。
- 关闭 Cloudflare 代理。
- 更改网站 API 路由。

## 3.3 备案及加速策略

系统必须明确区分：

- 已备案与未备案域名。
- 大陆 CDN 节点和境外 CDN 节点。
- 境外服务商网络及境内加速。
- CDN 边缘缓存和普通 HTTPS 反向代理。
- CDN 公网 IP 测速与受支持的接入方式。

未备案的 www.jdmv.net 不得被配置成依赖中国大陆境内备案准入的 CDN 接入方案。

对 static.01233.xyz 和 cdnapi.01233.xyz，必须保留现有可正常工作的阿里云全球加速能力。

---

# 第四章：多 DNS 服务商中心

## 4.1 DNS Provider 插件架构

实现统一 Provider 接口。

第一批必须支持：

- 阿里云 DNS。
- 腾讯云 DNSPod。
- Cloudflare DNS。

第二批支持：

- 华为云 DNS。
- DNS.com（以实际 API 为准）。
- 自定义第三方 DNS Provider。

后续扩展新服务商，不允许修改核心调度算法。

统一接口至少包含：

- ValidateCredentials
- DetectCapabilities
- ListZones
- ListRecords
- CreateRecord
- UpdateRecord
- DeleteRecord
- EnableRecord
- DisableRecord
- GetNameservers
- GetLimits
- VerifyPropagation

不支持的能力必须返回明确的 Unsupported Capability，而非静默忽略。

## 4.2 多账号管理

允许同一个服务商添加多个账号。

功能包括：

- API 凭证管理。
- 账号连接测试。
- 服务商权限检测。
- 凭证加密。
- 账号禁用。
- API 限速。
- 多域名绑定。
- 凭证轮换。
- 余额或配额状态（API 可提供时）。
- 调用历史。
- 失败诊断。

禁止将 SecretKey 或 Token 明文输出到日志、浏览器或错误信息。

## 4.3 域名管理

每个域名可独立配置：

- 当前权威 DNS Provider。
- 备用 DNS Provider。
- DNS 账号。
- 运营商线路。
- 地域线路。
- TTL。
- DNS 发布策略。
- 最大发布 IP 数量。
- 节点资源池。
- 自动优化开关。
- 故障回滚开关。

## 4.4 DNS 套餐能力识别

针对每个域名，系统应检测、查询或配置：

- 当前套餐。
- 可用线路。
- 同记录集条数上限。
- 最小 TTL。
- API 限流。
- 是否支持权重。
- 是否支持健康检查。
- 是否支持 A、AAAA、CNAME 等记录。
- 是否支持中国大陆运营商线路。
- 是否支持境内与境外解析。

无法通过 API 自动获取的套餐限制，必须基于官方文档和管理员确认建立能力档案。

不得仅根据服务商名称推断套餐。

## 4.5 多 DNS 服务商运行模式

必须支持：

### 模式 A：单权威 DNS

每个域名使用一个权威服务商，其他服务商保存备用配置。

### 模式 B：DNS 迁移

执行：

发现现有记录
→ 生成迁移计划
→ 预同步新服务商
→ 检查 SOA、NS、DNSSEC、MX、TXT 等记录
→ 验证记录完整性
→ 人工批准 NS 切换
→ 监测委派变化
→ 迁移完成

### 模式 C：Multi-DNS

作为独立高级功能。

只有当多个权威服务商能够正确协同，并且记录同步、委派、DNSSEC、故障恢复等条件满足时才允许启用。

不得简单混用 NS 而不做一致性检查。

## 4.6 DNS 记录管理

支持：

- A。
- AAAA。
- CNAME。
- TXT。
- MX。
- NS。
- CAA。
- SRV。
- HTTPS / SVCB（服务商支持时）。

支持批量操作、筛选、比较、导入导出、历史快照和回滚。

严格保护未由 SmartCDN 托管的解析记录。

---

# 第五章：中国大陆三网智能优选

## 5.1 核心目标

分别优化：

- 中国电信。
- 中国联通。
- 中国移动。

同时支持：

- 中国境内默认线路。
- 中国境外。
- 全局默认。
- 服务商提供的其他地域和 ISP 线路。

## 5.2 调度周期

默认每 10 分钟执行一次。

调度周期可由管理员修改。

## 5.3 候选池和发布数量

默认每条运营商线路最多维护 5 个候选 IP。

必须区分：

- Candidate Pool：候选 IP。
- Eligible Pool：通过安全及 HTTPS 验证的 IP。
- Published Set：当前 DNS 已发布的 IP。
- Standby Pool：备用节点。

实际 DNS 发布条数不得超过服务商套餐限制。

例如候选池为 5 个，而 DNS 当前套餐只能发布 2 个，则最多发布 2 个，其他保留备用。

## 5.4 分布式测速探针

必须开发独立 Go Probe Agent。

支持将 Probe Agent 部署在：

- 中国电信网络。
- 中国联通网络。
- 中国移动网络。
- 香港。
- 日本。
- 新加坡。
- 其他地区。

探针必须真实运行在对应网络环境中。

不能通过更改 User-Agent、HTTP Header 或代理标签来冒充运营商测速。

必须采集：

- DNS 查询时间。
- TCP 连接时间。
- TLS 握手时间。
- HTTPS TTFB。
- HTTP 状态。
- 实际业务请求成功率。
- 下载吞吐量。
- 丢包或连接失败率。
- TLS 证书有效性。
- 测试时间。
- 探针网络及地区。
- IP 地址。
- 测试目标域名。

对共享 CDN IP 必须使用正确的 Host 和 SNI 进行检测。

## 5.5 智能评分算法

建立可配置综合评分：

Score = Success + Latency + TTFB + Throughput + Stability

权重由后台配置。

默认优先级：

1. HTTPS 正确性。
2. 业务请求成功率。
3. 长期稳定性。
4. TTFB。
5. 下载速度。
6. 网络延迟。

要求支持：

- 滑动时间窗口。
- 异常值排除。
- 节点冷却时间。
- 切换滞后阈值。
- 连续失败熔断。
- 自动恢复。
- 人工锁定。
- 灰度发布。
- 防止频繁 DNS 切换。

## 5.6 DNS 发布安全

调度执行流程：

1. 测速。
2. 计算评分。
3. 选择候选。
4. 校验 HTTPS。
5. 校验 Provider 能力。
6. 生成 DNS 差异。
7. 保存当前记录快照。
8. 获取分布式锁。
9. 执行幂等更新。
10. 查询权威 DNS。
11. 验证 HTTPS。
12. 记录结果。
13. 失败执行安全回滚。

必须提供 Dry Run。

DNS TTL 和递归缓存会影响实际生效时间，不得宣称 DNS 修改后立即全网生效。

---

# 第六章：Cloudflare CDN 与 SaaS

## 6.1 Cloudflare 管理

支持：

- 多 Cloudflare 账号。
- API Token。
- Zone。
- DNS 记录。
- Proxy 状态。
- Cache Rules。
- Cache Purge。
- SSL/TLS。
- 安全配置状态读取。
- 域名性能监控。

所有写入操作必须检查账号和 Token 权限。

## 6.2 Cloudflare for SaaS

支持：

- Custom Hostnames。
- Fallback Origin。
- CNAME Target。
- Custom Origin（套餐支持时）。
- 主机名验证。
- 证书验证。
- 主机名接入状态。
- SaaS 证书生命周期。
- 接入失败诊断。

针对首个网站，计划使用：

- SaaS 回退源：origin.wueym.com。
- 正式网站：www.jdmv.net。

如需单独创建 CNAME Target，可采用 edge.wueym.com，但必须由实际 Cloudflare 配置验证，不能预设已经存在。

## 6.3 Cloudflare 优选 IP 限制

系统可以维护 Cloudflare IP 测速数据。

但必须遵守：

- Cloudflare for SaaS 标准模式以官方 CNAME 接入为基础。
- 任意 Cloudflare 公网 IP 不等于可授权使用的 SaaS A 记录入口。
- 不得绕过账号功能限制。
- 不得仅凭一次 HTTP 200 即判定 IP 可发布。
- 若缺乏受支持的 A 记录接入能力，只允许测速展示，不允许自动修改正式 DNS 指向这些 IP。
- Cloudflare 特殊 A 记录能力必须经过套餐授权及配置验证。

## 6.4 Cloudflare 与自有节点的双线路策略

系统允许同一业务域名同时存在：

- Cloudflare 官方接入线路。
- 自有家宽或服务器 HTTPS 入口线路。

但 DNS 记录编排需要根据服务商实际支持的记录类型及线路匹配规则执行。

若无法安全同时编排 A 与 CNAME，必须选择受支持的替代方式，例如使用独立子域名、独立区域或经过验证的流量管理服务。

不得强行创建不被 DNS Provider 支持的配置。

---

# 第七章：自有家宽和反向代理节点

## 7.1 节点范围

允许管理员接入自己拥有或得到授权的节点。

节点类型：

- 家庭宽带。
- 香港服务器。
- 日本服务器。
- 新加坡服务器。
- 其他公网节点。

禁止把不受控的公共 VPN Gate 志愿者节点当作可信生产 CDN 入口。

## 7.2 节点管理

支持：

- 节点注册。
- 节点分组。
- 公网 IP。
- IPv6。
- 节点地区。
- ISP 标识。
- Nginx 反向代理。
- HTTPS。
- 回源地址。
- 回源 Host。
- TLS SNI。
- WebSocket。
- 缓存规则。
- 最大带宽。
- 健康检查。
- 节点启停。
- 权重。
- 自动故障摘除。

## 7.3 节点 Agent

提供独立 Go Node Agent。

功能：

- 上报节点状态。
- 上报 CPU、内存、网络指标。
- 检测 Nginx。
- 检测证书有效期。
- 检测实际 HTTPS 服务。
- 接收经过授权的配置更新任务。
- 执行配置验证和安全 reload。
- 返回执行结果。

Agent 与管理中心通信必须经过认证和加密。

不得允许未经授权的任意命令执行。

## 7.4 节点缓存

支持：

- 静态资源缓存。
- Cache-Control。
- 缓存容量限制。
- 缓存过期。
- 缓存清理。
- Cache HIT / MISS 指标。
- 按路径配置绕过缓存。
- 可配置请求头和查询参数缓存键。

默认禁止缓存敏感 API、鉴权响应和 WebSocket。

---

# 第八章：SSL 证书自动化中心

## 8.1 参考项目

Certimate 官方文档：

https://docs.certimate.me/zh-CN/docs/introduction/

GitHub：

https://github.com/certimate-go/certimate

目标是在 SmartCDN 中原生实现类似 Certimate 的完整 SSL 自动化能力。

借鉴其：

- 证书申请。
- 自动续期。
- 多 DNS Provider。
- 多部署目标。
- 可视化工作流。
- 执行记录。
- 告警。

不得直接依赖独立的 Certimate 实例作为核心功能。

如复用 Certimate 源代码，必须先检查许可证兼容性。

## 8.2 ACME 证书申请

支持：

- Let's Encrypt。
- ZeroSSL。
- 其他符合条件的 ACME CA。

支持：

- 单域名。
- 多域名 SAN。
- 泛域名。
- RSA。
- ECDSA。
- CSR。
- PEM。
- 证书链。
- 证书导入。
- 证书导出。
- 有效期监控。
- 自动续期。

私钥必须加密保存。

不得明文记录私钥。

## 8.3 DNS-01 验证

与现有 DNS Provider 系统共用账号。

自动执行：

创建 ACME Order
→ 获取 Challenge
→ 创建 TXT 记录
→ 检查权威 DNS 传播
→ 完成 ACME 验证
→ 下载证书
→ 保存证书
→ 清理临时 TXT
→ 触发部署工作流

必须支持并发 TXT Challenge，不得误删其他正在使用的验证记录。

支持 `_acme-challenge` 委派。

## 8.4 HTTP-01 验证

支持可选 HTTP-01。

必须处理：

- Nginx 验证路径。
- 80 端口。
- HTTP 重定向。
- CDN 缓存。
- Cloudflare 代理。
- 源站路由。
- 验证失败诊断。

对于 CDN 场景默认优先 DNS-01。

## 8.5 自动续期

要求：

- 自动计算续期窗口。
- 支持 ACME ARI（CA 支持时）。
- 续期并发控制。
- CA 限流控制。
- 失败重试。
- 指数退避。
- 续期历史。
- 过期预警。
- 续期成功自动部署。
- 多目标部署失败处理。

## 8.6 证书部署目标

第一阶段支持：

- 宝塔 Nginx 网站。
- 原生 Linux Nginx。
- SSH 远程服务器。
- Cloudflare 支持的证书场景。
- 阿里云 CDN。
- 阿里云 ESA。

第二阶段支持：

- 腾讯云 CDN。
- 腾讯云 EdgeOne。
- 华为云 CDN。
- 自定义 Webhook。

必须根据云产品的实际 API 进行部署。

不得假定 CDN、ESA、EdgeOne 使用相同接口。

## 8.7 Cloudflare SSL 区分

必须区分：

- Cloudflare 托管边缘证书。
- SaaS Custom Hostname 证书。
- 用户上传的边缘证书。
- Cloudflare Origin CA。
- 自有节点公开可信证书。

Origin CA 证书不能直接用于面向普通浏览器的 HTTPS 公网入口。

## 8.8 SSL 自动部署工作流

开发可视化 DAG 工作流引擎。

支持节点：

- 开始。
- 申请证书。
- DNS 验证。
- HTTP 验证。
- 导入证书。
- 校验证书。
- 部署证书。
- 条件判断。
- 并行分支。
- 等待。
- Webhook。
- 通知。
- 回滚。
- 结束。

支持：

- 工作流编辑。
- 运行日志。
- 定时触发。
- 手动触发。
- 自动续期触发。
- 失败重试。
- 幂等执行。
- 暂停和取消。
- 历史版本。
- 执行回放。
- 运行审计。

## 8.9 安全部署与回滚

Nginx 证书部署必须采用：

1. 检查证书链。
2. 检查域名 SAN。
3. 检查证书与私钥匹配。
4. 保存旧证书。
5. 写入受保护的临时目录。
6. 校验文件权限。
7. 执行 Nginx 配置测试。
8. 安全替换证书文件。
9. Reload Nginx。
10. 从外部检查 HTTPS。
11. 成功确认或失败回滚。

对于 CDN 云端证书部署，需要另外适配云厂商的任务状态和回滚机制。

---

# 第九章：监控、告警及故障恢复

## 9.1 监控范围

监控：

- DNS 解析结果。
- 当前 NS。
- DNS Provider API。
- 三网 HTTPS 成功率。
- 节点可用性。
- CDN 命中率（API 支持时）。
- 源站 TTFB。
- TLS 握手。
- SSL 到期。
- 证书部署。
- 自动调度任务。
- MySQL。
- Redis。
- CPU。
- 内存。
- 网络吞吐量。

## 9.2 告警渠道

支持：

- Telegram Bot。
- SMTP 邮件。
- Webhook。

后续扩展：

- 企业微信。
- 飞书。
- 钉钉。

## 9.3 告警机制

支持：

- 告警规则。
- 告警等级。
- 连续失败阈值。
- 告警去重。
- 静默窗口。
- 恢复通知。
- 告警历史。
- 告警确认。
- 故障事件关联。

## 9.4 自动故障处理

支持：

- 节点摘除。
- 备用节点启用。
- DNS 安全回滚。
- 发布任务暂停。
- API 凭证失效告警。
- 证书部署回滚。
- 人工干预。
- 故障恢复验证。

必须设置熔断和切换速率限制，避免多个控制器同时操作同一域名导致抖动。

---

# 第十章：管理后台完整设计

## 10.1 UI 设计原则

参考专业云计算控制台。

采用 Arco Design React 组件体系。

要求：

- 专业商业化视觉设计。
- 无明显 AI 生成风格。
- 左侧可展开多级菜单。
- 顶部导航和全局搜索。
- 清晰的表格及筛选。
- 统一交互规范。
- 深色/浅色主题。
- 桌面及移动端适配。
- 完整加载、错误和空状态。
- 图表数据可追溯。
- 危险操作二次确认。

不得把大量功能挤在单页。

## 10.2 后台菜单信息架构

### 仪表盘

- 系统概览
- DNS 服务状态
- CDN 节点状态
- SSL 证书状态
- 三网性能趋势
- 最近告警

### DNS 管理

- DNS 服务商
- DNS 账号
- 域名管理
- DNS 记录
- 智能解析线路
- DNS 迁移
- Multi-DNS
- DNS 历史快照

### CDN 管理

- CDN 服务商
- Cloudflare Zones
- Cloudflare SaaS
- 阿里云 CDN / ESA
- 自有家宽节点
- 节点资源池
- 缓存规则
- 缓存清理

### 智能调度

- 调度仪表盘
- 三网测速
- 探针管理
- 优选 IP 池
- 调度策略
- 自动发布任务
- 发布历史
- 故障回滚

### SSL 证书

- 证书总览
- ACME 账号
- 证书申请
- 已签发证书
- 证书导入
- 自动续期
- 证书部署目标
- 证书工作流
- 执行日志
- 证书监控

### 监控中心

- 实时监控
- 性能趋势
- DNS 健康
- CDN 健康
- HTTPS 状态
- 告警记录

### 系统管理

- 用户管理
- 角色权限
- API 凭证
- 通知渠道
- 审计日志
- 系统配置
- 备份恢复
- 系统诊断

## 10.3 设计交付要求

在开发前必须完成：

- 页面信息架构。
- 设计规范。
- 色彩与字体规范。
- 组件规范。
- 高保真页面设计。
- 页面状态设计。
- 交互说明。
- 响应式方案。

每个页面需独立评审，不允许仅提供一张包含所有页面的总览图。

---

# 第十一章：数据库及系统架构

## 11.1 数据模型

MySQL 至少覆盖：

### 用户与权限

- users
- roles
- permissions
- user_roles
- audit_logs

### DNS

- dns_providers
- dns_accounts
- domains
- domain_provider_bindings
- dns_records
- dns_record_snapshots
- dns_migrations
- dns_capabilities

### CDN 和节点

- cdn_providers
- cdn_accounts
- cdn_nodes
- node_groups
- probe_agents
- probe_results
- node_health_events

### 调度

- network_lines
- scheduling_policies
- scheduling_runs
- publishing_jobs
- publishing_changes
- rollback_jobs

### Cloudflare

- cloudflare_zones
- cloudflare_custom_hostnames
- cloudflare_origins

### SSL

- acme_accounts
- certificate_authorities
- certificates
- certificate_domains
- certificate_versions
- certificate_orders
- certificate_challenges
- certificate_deploy_targets
- certificate_deployments
- certificate_renewal_jobs
- certificate_workflows
- certificate_workflow_runs

### 监控

- monitor_targets
- monitor_results
- alert_rules
- alert_events
- notification_channels
- system_settings

允许根据数据库规范化及实际功能需要调整表名和结构。

## 11.2 Redis 8

用于：

- 分布式锁。
- 任务队列。
- API 限流。
- 节点实时状态。
- 测速结果短期缓存。
- 会话状态。
- 工作流执行协调。
- 调度任务去重。

关键配置、审计记录、发布状态必须持久化到 MySQL。

## 11.3 项目目录

建议采用：

smartcdn/

- backend/
  - cmd/
  - internal/
    - api/
    - controller/
    - service/
    - repository/
    - dns/
    - cdn/
    - cloudflare/
    - scheduler/
    - probe/
    - certificate/
    - workflow/
    - security/
    - monitor/
  - migrations/
  - configs/
- frontend/
- agent/
  - probe/
  - node/
- deploy/
  - baota/
  - linux/
  - docker/
- scripts/
- tests/
- docs/
- _bmad/
- _bmad-output/

实际路径需要与安装的 BMAD 版本保持兼容。

---

# 第十二章：首次安装与部署

## 12.1 安装向导

初次启动时检测：

- Go 服务环境。
- MySQL。
- Redis。
- 数据库权限。
- 文件目录权限。
- 端口。
- Nginx。
- 服务器时间。
- 加密密钥配置。

首次安装流程：

1. 环境检测。
2. 数据库连接配置。
3. Redis 配置。
4. 初始化 Migration。
5. 创建管理员。
6. 设置应用密钥。
7. 配置后台访问地址。
8. 写入 install.lock。
9. 完成安装。

不得使用公开固定的默认管理员密码。

首次安装完成后，安装接口必须进入锁定状态。

## 12.2 宝塔 Go 项目

必须提供：

- Go 项目配置指导。
- Nginx 反向代理模板。
- 前端构建部署。
- 环境变量模板。
- 数据库初始化。
- Redis 配置。
- 权限设置。
- 日志目录。
- 更新脚本。
- 回滚脚本。
- 备份脚本。
- SSL 自动部署配置。

宝塔 Go 项目需实际验证启动、停止、重启、日志和环境变量管理。

## 12.3 原生 Linux

必须提供：

- systemd 服务文件。
- Nginx 配置。
- 安装脚本。
- 环境变量模板。
- 更新及回滚脚本。

## 12.4 Docker

提供：

- Dockerfile。
- Docker Compose。
- 健康检查。
- 持久化卷。
- 环境变量。
- 数据库初始化。
- 网络隔离。
- 备份方案。

## 12.5 生产环境安全

必须实施：

- 最小权限。
- API 凭证加密。
- TLS。
- CSRF 防护。
- XSS 防护。
- SSRF 防护。
- SQL 注入防护。
- RBAC。
- 登录限流。
- 审计。
- 数据库备份。
- 备份恢复验证。
- 敏感字段脱敏。
- 探针目标地址限制。
- 节点 Agent 认证。
- 证书私钥保护。

不得默认开放 MySQL、Redis 或内部 Agent 管理接口到公网。

---

# 第十三章：API 和插件扩展

## 13.1 API 规范

必须生成完整 OpenAPI 文档。

包括：

- 认证 API。
- DNS API。
- 域名 API。
- CDN API。
- Cloudflare API。
- 调度 API。
- Probe API。
- SSL API。
- 工作流 API。
- 监控 API。
- 管理员 API。

支持版本控制，例如 /api/v1。

## 13.2 插件扩展设计

DNS Provider、CDN Provider、证书部署目标必须采用独立适配接口。

要求：

- 能力声明。
- 配置校验。
- 凭证测试。
- 统一错误模型。
- 重试与限流。
- 日志和审计。
- 版本兼容。
- 单元和契约测试。

第三方自定义插件默认不允许获得任意操作系统权限。

---

# 第十四章：BMAD 开发流程

必须按照当前安装的 BMad Method 版本执行。

先识别项目是 Brownfield 还是 Greenfield。

## Phase 1 — Analysis

输出：

- Product Brief。
- 市场与竞品研究。
- 服务商 API 能力矩阵。
- DNS/CDN 可行性报告。
- 现网风险清单。
- 技术限制报告。

重点研究：

- 阿里云 DNS。
- DNSPod。
- Cloudflare DNS。
- Cloudflare for SaaS。
- 阿里云 CDN / ESA。
- Certimate。
- ACME。
- 宝塔 Go 项目。

## Phase 2 — Planning

输出：

- PRD。
- 用户画像。
- 用户角色。
- 用户故事。
- 完整功能清单。
- 验收标准。
- 非功能需求。
- 功能依赖关系。

## Phase 3 — Solutioning

输出：

- Architecture。
- ADR。
- 数据库 ERD。
- API Specification。
- Provider 接口设计。
- 调度状态机。
- DNS 发布状态机。
- SSL 工作流状态机。
- 故障回滚方案。
- 安全方案。
- UX Design。
- UI 组件规范。
- Epics and Stories。
- Implementation Readiness。

## Phase 4 — Implementation

建议划分以下 Epics：

1. 项目基础架构与安装系统。
2. 用户、认证和 RBAC。
3. DNS Provider 插件框架。
4. 阿里云、DNSPod、Cloudflare DNS 接入。
5. 多域名及权威 DNS 管理。
6. DNS 迁移和记录发布引擎。
7. 分布式测速探针。
8. 三网智能调度系统。
9. Cloudflare CDN 与 SaaS。
10. 自有家宽节点和 Agent。
11. SSL ACME 证书管理。
12. SSL 多平台部署。
13. SSL 可视化工作流。
14. 监控与告警。
15. 专业管理后台。
16. 宝塔及多环境部署。
17. 全链路测试、性能优化与生产验收。

每个 Epic 拆分成可独立验收的 Stories。

每个 Story 都必须包含：

- 需求描述。
- 前置依赖。
- API。
- 数据结构。
- 前端界面。
- 安全约束。
- 测试用例。
- 验收标准。
- 文档更新。

## Phase 5 — Production Readiness

必须完成：

- Go 单元测试。
- API 集成测试。
- Provider 契约测试。
- 前端功能测试。
- E2E 测试。
- 三网探针真实测试。
- DNS 安全发布测试。
- DNS 回滚测试。
- SSL 申请测试。
- SSL 续期测试。
- Nginx 证书部署测试。
- Cloudflare SaaS 接入测试。
- 数据库恢复测试。
- Redis 故障恢复测试。
- 宝塔 Go 项目部署测试。
- Docker Compose 部署测试。
- 安全测试。
- 压力测试。
- 生产可用性验证。

缺少真实凭证或环境时，必须标注待验证。

严禁虚构测试结果。

---

# 第十五章：最终交付

必须交付：

1. 完整 Go 后端源码。
2. 完整 React 前端源码。
3. DNS Provider 驱动。
4. Cloudflare 集成。
5. 分布式 Probe Agent。
6. 节点管理 Agent。
7. 智能调度算法。
8. DNS 发布和回滚引擎。
9. SSL ACME 模块。
10. 证书工作流引擎。
11. 多平台证书部署器。
12. 所有后台页面。
13. MySQL Migration。
14. Redis 配置。
15. 宝塔 Go 项目部署资料。
16. 原生 Linux 部署脚本。
17. Dockerfile 和 Docker Compose。
18. OpenAPI 文档。
19. 自动化测试。
20. 管理员手册。
21. 运维手册。
22. 安全文档。
23. 故障恢复手册。
24. BMAD 规划文档。
25. 最终验收报告。

不能以已经生成文档、编译成功或运行演示环境作为全部验收通过的依据。

必须证明核心功能在对应真实环境中可运行。

---

# 第十六章：完成标准

最终 SmartCDN 必须具备：

- 多 DNS 服务商。
- 多 DNS 账号。
- 多域名。
- 独立权威 DNS 管理。
- DNS 记录自动同步。
- 三网真实测速。
- 智能 IP 候选池。
- 安全 DNS 发布。
- 自动故障回滚。
- Cloudflare SaaS。
- 自有家宽节点。
- SSL 自动申请。
- SSL 自动续期。
- 多 CDN / 服务器证书部署。
- 可视化工作流。
- 监控告警。
- 专业完整管理后台。
- 宝塔 Go 项目正式部署。
- 原生 Linux 部署。
- Docker 和 Docker Compose 部署。

最终交付应为可用于正式生产环境的完整项目。

---

# 第十七章：立即执行指令

现在开始使用 BMad Method 执行 SmartCDN 项目。

首先检查：

1. 当前 Git 仓库。
2. 已有源码。
3. BMad Method 安装状态及版本。
4. 已有规划文档。
5. 已有 UI 设计。
6. 当前开发环境。

如果项目尚未初始化，按照当前 BMAD 版本创建项目工作区。

如果已有项目，必须保护现有代码，不得无故覆盖。

请先完成 Analysis、Planning、Solutioning 和 Implementation Readiness，再启动正式开发。

在架构阶段，必须先重点解决以下高风险问题：

- Cloudflare SaaS A/CNAME 接入限制。
- 三网分线路 DNS 记录数量与套餐限制。
- 多 DNS 权威解析和迁移一致性。
- 未备案域名的 CDN 接入边界。
- SSL 证书自动部署安全。
- 宝塔 Go 项目实际运行机制。
- 分布式测速真实性。
- DNS 发布与故障恢复。
- API 凭证和证书私钥保护。

对于未确定的普通实现细节，由架构师制定合理技术决策并记录 ADR。

对于生产 DNS 切换、外部资源付费、敏感数据删除、证书替换等操作，必须经过明确授权。

项目开发过程中允许自动执行安全的本地代码生成、构建、静态检查、测试和文档更新。

每个阶段需要产出真实成果，发现问题立即修复。

最终目标：完成商业级 SmartCDN 全部源码、文档、测试和部署交付，而不只是需求设计。

---

# 第十八章：BMAD 开发质量管理与强制验收机制

## 18.1 Epic 完成度报告（强制）

每个 Epic 完成后，必须自动生成一份独立的完成度报告。

文件保存位置：

`docs/progress/epic-{编号}-completion-report.md`

报告必须包含：

1. Epic 编号与名称。
2. 关联的 Stories。
3. 已实现功能。
4. 未实现功能。
5. 前端页面完成情况。
6. 后端 API 完成情况。
7. 数据库 Migration 状态。
8. 第三方服务集成状态。
9. 自动化测试结果。
10. 安全检查结果。
11. 已知问题。
12. 未解决的依赖。
13. 真实环境验证情况。
14. 后续改进任务。
15. Epic 最终验收状态。

不得仅以代码已提交或能够编译作为 Epic 完成依据。

## 18.2 四级功能验收标准

每项功能必须明确标记以下验证等级：

| 等级 | 状态            | 验收要求                                           |
| ---- | --------------- | -------------------------------------------------- |
| L1   | 代码已实现      | 源码完整、编译成功、静态检查通过                   |
| L2   | 测试环境已验证  | 单元测试、集成测试或模拟环境测试通过               |
| L3   | 真实 API 已验证 | 使用真实服务商 API、凭证或可访问的实际服务完成验证 |
| L4   | 生产环境已验收  | 已授权的生产环境测试通过，并确认功能实际可用       |

特别要求：

- L1 不等于 L2。
- L2 不等于 L3。
- L3 不等于 L4。
- 禁止根据模拟测试结果宣称真实 API 已验证。
- 缺少第三方密钥时，功能最多只能标记到有证据支持的等级。
- 未经过生产环境确认，不能标记为 L4。

例如：

- DNSPod SDK 单元测试通过：L2。
- 使用真实 DNSPod 账号成功查询 DNS 记录：L3。
- 在授权生产域名成功执行变更并完成端到端验证：L4。

## 18.3 Epic 完成度报告模板

每个 Epic 必须生成以下结构：

### 基本信息

- Epic：
- 关联 Stories：
- 当前版本：
- 开始日期：
- 完成日期：
- 负责人 Agent：
- Git Commit：
- 验收等级：

### 功能完成情况

| 功能   | 代码 | 测试 | 真实 API | 生产验收 |
| ------ | ---- | ---- | -------- | -------- |
| 功能 A | 状态 | 状态 | 状态     | 状态     |
| 功能 B | 状态 | 状态 | 状态     | 状态     |

状态必须使用：

- PASS：通过
- FAIL：失败
- BLOCKED：受阻
- NOT_RUN：未执行
- N/A：不适用

每个 PASS 必须关联实际证据，包括测试报告、日志、API 响应摘要或验收记录。

### 遗留问题

每个问题至少包括：

- 问题 ID。
- 严重等级。
- 影响范围。
- 复现步骤。
- 预期行为。
- 实际行为。
- 责任模块。
- 修复状态。
- 关联 Story。

### 最终结论

Epic 状态只能为：

- NOT_STARTED
- IN_PROGRESS
- CODE_COMPLETE
- TEST_VERIFIED
- EXTERNAL_VERIFIED
- PRODUCTION_ACCEPTED
- BLOCKED

不得使用模糊的“基本完成”代替正式状态。

## 18.4 自动质量门禁（Quality Gates）

每个 Story 完成时自动执行：

1. Go 格式化与静态检查。
2. Go 单元测试。
3. Go Race Detector（适用模块）。
4. 前端 TypeScript 类型检查。
5. ESLint。
6. 前端构建。
7. API 契约测试。
8. 数据库 Migration 检查。
9. 安全及敏感信息扫描。
10. 相关集成测试。

涉及 DNS 发布、证书部署和自动调度的 Story，还必须包含失败注入与幂等性测试。

如有 P0 或 P1 级严重问题，禁止进入生产发布阶段。

## 18.5 自动问题追踪

开发过程中发现的问题必须登记，不允许仅在对话中口头描述。

统一保存：

`docs/progress/issues.md`

问题分级：

- P0：数据丢失、严重安全漏洞、生产服务中断。
- P1：核心业务不可用或重大功能失效。
- P2：一般功能异常或性能明显下降。
- P3：低影响问题或体验优化。

P0/P1 必须优先修复和重新测试。

不得通过删除测试、降低断言标准、屏蔽错误日志或修改验收标准来制造测试通过结果。

## 18.6 Sprint 完成报告

每个 Sprint 完成后，生成：

`docs/progress/sprint-{编号}-report.md`

至少包含：

- 本 Sprint 目标。
- 已完成 Stories。
- 未完成 Stories。
- Epic 进度。
- 代码提交。
- 测试统计。
- 缺陷数量。
- 关键技术风险。
- 第三方依赖。
- 下一 Sprint 计划。

进度百分比必须依据实际验收的 Stories 计算，不得由 AI 主观估算。

## 18.7 项目完成度仪表盘

后台开发管理文档必须能够汇总以下数据：

- Epic 总数。
- 已验收 Epic 数量。
- Story 总数。
- 已完成 Story 数量。
- L1/L2/L3/L4 各等级数量。
- 未解决 P0/P1 问题数量。
- 第三方 API 验证完成率。
- 生产环境验收完成率。

可以先通过 Markdown 和机器可读 JSON 生成报告，不要求把开发管理工具作为 SmartCDN 正式业务后台的一部分。

## 18.8 最终交付门禁

项目只有满足以下条件，才允许标记为 `PRODUCTION_READY`：

1. 所有约定 P0 功能实现完成。
2. 所有 P0 功能通过适用的自动化测试。
3. 所有必要的生产集成完成真实验证。
4. 生产部署路径经过实际测试。
5. 不存在未解决的 P0/P1 缺陷。
6. 数据库备份与恢复验证通过。
7. DNS 发布与回滚验证通过。
8. SSL 自动续期及部署验证通过。
9. 权限、安全和敏感信息保护检查通过。
10. 监控与告警工作正常。
11. 所有未完成的低优先级事项均已登记。
12. 最终验收报告附有可核查证据。

如生产凭证、服务器或授权不足，必须标记为 `AWAITING_PRODUCTION_VALIDATION`，不得虚报已上线或已验收。

## 18.9 BMAD Agent 执行规则

所有 BMAD Agent 必须遵守本章节。

- Product Manager：确保需求及验收标准完整。
- Architect：负责技术约束、模块边界和风险设计。
- UX Designer：负责完整 UI/UX 及页面状态。
- Scrum Master：负责 Stories、Sprint 和进度追踪。
- Developer：负责实现、测试和代码证据。
- Reviewer：负责独立代码和安全审查。
- QA：负责验证结果与缺陷追踪。

上述角色可以由不同模型或 Agent 承担，但必须保留职责分离和独立审查流程。

每完成一个 Story，自动更新对应报告。

每完成一个 Epic，自动执行验收、生成完成度报告并检查质量门禁。

不得通过直接修改状态文件跳过测试。

## 18.10 持续执行要求

在已授权的开发环境中，Agent 应主动继续执行已经批准的开发任务，包括代码实现、测试、修复、构建和文档更新。

如果遇到缺少密钥、外部权限、真实服务器或生产变更授权，应明确记录阻塞原因，并继续推进不受该阻塞影响的其他工作。

未经批准不得执行高风险生产操作。

最终目标不仅是开发完成，还必须保证交付成果具有明确、真实、可核查的质量证据。
