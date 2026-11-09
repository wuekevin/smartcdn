---
title: "SmartCDN Product Brief — 技术附录与深度分析"
status: final
created: 2026-11-09
updated: 2026-11-09
brief: brief.md
---

# SmartCDN — 技术附录

> 本文件存放简报正文中不适合放置的技术细节、能力矩阵、约束分析和待验证事项。
> 供 PRD、架构设计和 Epic 拆分阶段参考。

---

## A. DNS Provider API 能力矩阵（规划期基线）

> 以下数据基于官方文档截至 2026-11，实施阶段必须通过真实 API 调用验证。

| 能力 | 阿里云 DNS | 腾讯云 DNSPod | Cloudflare DNS |
|------|-----------|--------------|----------------|
| 分运营商线路（电信/联通/移动） | 付费套餐 | 付费套餐 | 不支持（Geo 路由需 Load Balancing 附加产品） |
| 分境内/境外线路 | 是 | 是 | 是（Geo 分区路由） |
| 同记录集最大条数 | 套餐依赖（企业版更高） | 套餐依赖 | 无固定限制（实际受 Zone API 速率限制） |
| 最小 TTL | 免费版 600s；企业版 1s | 免费版 600s；付费版 1s | 300s（代理关闭）；自动（代理开启） |
| 权重轮询 | 付费版支持 | 付费版支持 | 需 Load Balancing 产品 |
| API 限流 | 300次/分钟（通用） | 1000次/小时 | 1200次/5分钟（Zone API）|
| 健康检查 | GTM 产品独立 | 付费版 D 监控 | 需 Load Balancing |
| CAA 记录 | 是 | 是 | 是 |
| HTTPS/SVCB 记录 | 部分支持 | 否 | 是 |

**关键约束：**
- DNSPod 免费版：单域名 DNS 记录 ≤ 2条 A 记录（同记录集），不支持分运营商线路 → 三网调度必须检测套餐后决策
- 阿里云 DNS 免费版：不支持分运营商线路（需企业版或云解析 DNS 专业版）
- Cloudflare DNS：分运营商线路不原生支持，三网优化依赖 Cloudflare for SaaS + 自有节点双线策略

---

## B. Cloudflare SaaS 技术约束（高优先级验证）

### B.1 标准 Custom Hostnames 接入方式

Cloudflare for SaaS 标准模式：
- 客户在自己的 DNS 服务商创建 CNAME → `{saas-zone}.cdn.cloudflare.net`（CNAME Target）
- Cloudflare 验证主机名所有权（DCV），颁发托管 SSL 证书
- **A 记录接入**：标准 SaaS 不提供 A 记录接入；若用户 DNS 服务商不支持 CNAME Flattening（如根域 `@`），需使用 CNAME Flattening 或迁移到 Cloudflare DNS
- **付费要求**：Cloudflare for SaaS 需要 Business 或 Enterprise 套餐，或按 Custom Hostname 计费（$0.10/月/主机名，含 100 个免费）

### B.2 计划中的 SmartCDN 场景

| 域名 | 角色 | 计划接入方式 | 状态 |
|------|------|------------|------|
| www.jdmv.net | SaaS 客户域名 | CNAME → edge.wueym.com（待确认） | 待 Cloudflare 配置验证 |
| origin.wueym.com | Fallback Origin | 实际源站，需 HTTPS | 待检测实际服务 |
| edge.wueym.com | CNAME Target | Cloudflare SaaS Zone 目标 | 待确认是否需要独立创建 |

### B.3 优选 IP 场景限制

- Cloudflare 公网 IP 属于 Cloudflare 共享基础设施，不保证路由行为
- 通过测速找到延迟低的 Cloudflare IP 并手动发布 A 记录：
  - 技术可行但不受官方支持，且 Cloudflare IP 随时可能变动
  - **SmartCDN 策略**：可维护测速数据用于展示，但禁止在无套餐授权的情况下自动将优选 IP 发布到生产 DNS

---

## C. 阿里云 CDN / ESA 能力区分

| 产品 | 类型 | 备案要求 | 适用场景 |
|------|------|---------|---------|
| 阿里云 CDN | 传统 CDN，回源型 | 大陆节点加速需 ICP | static.01233.xyz, cdnapi.01233.xyz |
| 阿里云 ESA（边缘安全加速） | Cloudflare-like 全站代理 | 大陆节点需 ICP | 可替代部分 Cloudflare 功能，API 不同 |
| 阿里云全球加速 GA | 专线加速 | 否（但域名需 ICP 才可用大陆弹性接入） | 当前 static.01233.xyz, cdnapi.01233.xyz 在用 |

**重要区分：**
- 阿里云 CDN API 和 ESA API 完全不同，SmartCDN 必须分别实现适配器
- 当前 static.01233.xyz 和 cdnapi.01233.xyz 走阿里云全球加速（非 CDN），迁移前需确认是 GA 还是 CDN 产品

---

## D. 宝塔 Go 项目管理器行为（待实际验证）

### D.1 已知机制（文档层面）

- 宝塔面板"Go 项目"模块支持：创建项目、指定二进制路径、启动参数、工作目录
- 进程管理：start/stop/restart 通过宝塔面板 API 或控制台执行
- 日志：stdout/stderr 重定向到宝塔日志目录
- 环境变量：通过宝塔项目配置设置，注入到进程环境

### D.2 待验证项目（架构阶段必须确认）

- [ ] 宝塔 Go 项目是否支持设置工作目录（SmartCDN 需要读取 config/ 相对路径）
- [ ] 宝塔环境变量变更后是否需要重启生效
- [ ] 宝塔 Go 项目的进程守护机制（是否使用 supervisor、pm2 或自定义）
- [ ] 宝塔版本要求（确认支持 Go 项目的最低版本）
- [ ] 宝塔 Nginx 反向代理 WebSocket 支持（需要 Upgrade 和 Connection 头）
- [ ] 前端静态文件部署方式（Nginx 静态托管还是 Go 内嵌 embed.FS）

---

## E. ACME 协议与 Certimate 参考

### E.1 ACME 标准实现

- RFC 8555（ACME v2）标准 Go 库：`golang.org/x/crypto/acme`
- 更完整的客户端库：`github.com/go-acme/lego`（支持 100+ DNS Provider，含阿里云/DNSPod/Cloudflare）
- Certimate 使用 `github.com/go-acme/lego`，MIT 许可证，可参考但需独立实现

### E.2 DNS-01 并发控制要求

- 同一域名可能同时有多个 ACME 订单（如主域名和泛域名同时续期）
- `_acme-challenge` TXT 记录可能有多个值（RFC 8555 允许）
- 实现必须追踪每个 Challenge 对应的 TXT 记录，清理时只删除当前 Challenge 的值

### E.3 ARI（ACME Renewal Information）

- Let's Encrypt 和 ZeroSSL 均已支持 RFC 草案 ARI
- ARI 允许 CA 在证书仍有效时主动建议续期窗口，提高续期成功率
- SmartCDN SSL 续期系统应优先查询 ARI，退回默认"到期前 30 天"逻辑

---

## F. 初始业务域名现网基线（待手动采集）

> 以下数据需要在开发开始前人工采集并记录。没有对应账号或网络权限时标注"待验证"。

| 域名 | 当前 DNS 服务商 | NS 记录 | 主要 A/CNAME | CDN 类型 | HTTPS 证书 | 备案状态 |
|------|-------------|---------|------------|---------|-----------|---------|
| www.jdmv.net | Cloudflare | Cloudflare NS | Proxied | Cloudflare | Cloudflare 托管 | 未备案 |
| origin.wueym.com | 待查 | 待查 | 待查 | 无/源站 | 待查 | 待查 |
| static.01233.xyz | 待查 | 待查 | CNAME → 阿里云全球加速 | 阿里云全球加速 | 待查 | 已备案 |
| cdnapi.01233.xyz | 待查 | 待查 | CNAME → 阿里云全球加速 | 阿里云全球加速 | 待查 | 已备案 |

**现网变更保护清单（严禁在未授权情况下执行）：**
1. 切换 www.jdmv.net 的 Cloudflare 代理状态
2. 修改 www.jdmv.net 的任何 DNS 记录
3. 修改 static.01233.xyz / cdnapi.01233.xyz 的阿里云全球加速配置
4. 替换任何域名的现有 SSL 证书
5. 修改 origin.wueym.com 的源站配置

---

## G. 数据库关键设计要点

### G.1 凭证安全

- DNS/CDN Provider API Key/Secret 必须使用 AES-256-GCM 加密，密钥从环境变量读取
- SSL 证书私钥必须单独加密存储，与证书元数据分离
- 加密密钥不得硬编码，不得写入代码仓库

### G.2 审计日志要求

- 所有 DNS 变更、证书操作、调度执行、凭证变更均需写入 `audit_logs`
- 记录字段：操作者 ID、操作类型、操作对象、操作前状态、操作后状态、IP 地址、时间戳
- 审计日志不得修改或删除（软删除不适用于审计）

### G.3 分布式锁场景

- DNS 发布：同一域名的 DNS 变更需要互斥锁（Redis SETNX + TTL）
- 调度任务：同一调度任务不允许并发执行（幂等键 + 锁）
- 证书续期：同一证书的续期任务需要互斥（防止并发向 CA 提交重复 Order）

---

## H. 二期能力预留（架构需设计插槽）

| 能力 | V1 状态 | 预留要求 |
|------|---------|---------|
| 华为云 DNS | 不实现 | Provider 接口已定义，插槽预留 |
| 腾讯云 CDN 证书部署 | 不实现 | DeployTarget 接口预留 |
| 企业微信/飞书/钉钉 | 不实现 | NotificationChannel 接口预留 |
| Webhook DNS Provider | 框架层面 | 允许通过 Webhook 接入自定义 DNS |

---

## I. 开发阶段优先研究项目（Phase 1 Analysis 必须完成）

1. **Cloudflare SaaS 实际 API**：Custom Hostnames 创建、验证状态轮询、DCV 类型选项
2. **DNSPod API v3**：记录线路枚举、套餐限制查询接口
3. **阿里云 DNS API**：分运营商线路 ID、套餐能力查询、记录数量限制
4. **阿里云 CDN vs ESA vs 全球加速**：三个产品的证书部署 API 区别
5. **宝塔 Go 项目**：实际部署测试（需要测试服务器）
6. **go-acme/lego 许可证**：MIT，可用；确认具体版本 API 兼容性
7. **分布式探针网络**：确认可用的测试用运营商 IP 或 VPS

---

*附录版本 1.0 — 2026-11-09*
*下一更新：架构阶段 ADR 完成后*
