---
title: "SmartCDN — 商业级多 DNS、智能 CDN 与 SSL 自动化管理平台"
status: final
created: 2026-11-09
updated: 2026-11-09
project: SmartCDN
version: 1.0
---

# Product Brief: SmartCDN

## Executive Summary

SmartCDN 是一套可私有化部署的商业级网络基础设施管理平台，专为需要在多个 DNS 服务商、CDN 节点和 SSL 证书体系之间统一运营的技术团队和个人站长而设计。平台以 Go + React 构建，核心价值在于：通过可扩展插件架构统一管理阿里云 DNS、腾讯云 DNSPod、Cloudflare 等多家 DNS/CDN 服务商，同时对中国大陆电信、联通、移动三网分别执行真实探针测速，并依据结果安全自动地完成智能 DNS 解析调度。

平台整合了 SSL 证书的全生命周期管理（ACME 自动申请、续期、多目标部署）和可视化工作流引擎，并对生产网络实施严格保护：所有变更均需 Dry Run 预检和显式授权，DNS 发布内置快照与安全回滚。SmartCDN 不是演示程序，也不是单纯的 DNS CRUD 工具——它是面向复杂多云多节点场景的完整运营中台。

在中国互联网的特殊环境下，三网优化、ICP 备案限制、境内外 CDN 接入边界、Cloudflare SaaS 接入方式等问题往往需要在多个服务商控制台之间反复切换操作。SmartCDN 将这些分散的操作纳入统一的平台，同时保持对各服务商 API 真实能力的精确建模，不编造不存在的接口。

## The Problem

**问题一：多服务商割裂，运营成本高。**
一个典型的运营场景：网站主域名 DNS 托管于阿里云，静态资源走阿里云 CDN，同时在 Cloudflare 保留一个 Zone 作为国际线路，自有家宽节点作为备用出口，SSL 证书分布在三个不同平台。任何一次 IP 更换、证书续期或线路切换，都需要在四五个控制台之间手动完成并保持一致性。漏操作或操作顺序错误将导致网站中断。

**问题二：中国三网优化缺乏自动化工具。**
电信、联通、移动三网对同一域名的最优 IP 可能完全不同，且随网络状况实时变化。现有开源工具（如 DNSPod 控制台）提供静态分线路配置，但无法感知节点实时健康状态并自动调度。企业级解决方案价格昂贵且不支持私有部署。

**问题三：SSL 证书管理分散，续期风险高。**
Let's Encrypt 证书 90 天过期，涉及多个域名、多个部署目标（Nginx、CDN、Cloudflare SaaS）时，手动续期极易遗漏。DNS-01 验证需要调用 DNS Provider API，与 CDN 证书部署流程完全解耦，使得端到端自动化难以实现。

**问题四：缺乏统一的操作审计与安全回滚。**
在多服务商环境中，无法简单地回答"三小时前 DNS 状态是什么"或"是谁触发了这次解析切换"。故障发生时，回滚操作需要人工翻查多个控制台历史。

## The Solution

SmartCDN 以统一的 Provider 插件接口为核心，将各服务商差异封装在适配层内。主控程序通过标准接口操作 DNS 记录、CDN 配置和证书部署，对上层业务逻辑屏蔽服务商差异。

**DNS 统一管理层**：支持阿里云 DNS、腾讯云 DNSPod、Cloudflare DNS 作为第一批 Provider，后续扩展华为云、DNS.com 等无需修改核心调度代码。对每个 Provider 进行套餐能力检测（线路支持、同记录集条数上限、最小 TTL、是否支持权重等），并基于实际能力约束调度决策。

**三网智能调度系统**：独立的 Go Probe Agent 真实部署于电信、联通、移动网络环境，采集 DNS 查询时间、TCP 连接时间、TLS 握手时间、HTTPS TTFB、业务成功率等多维指标。评分算法基于滑动时间窗口和可配置权重，包含熔断、冷却、灰度发布、防频繁切换机制。DNS 发布流程内置 13 步安全流程（快照→锁→幂等更新→验证→失败回滚）。

**SSL 自动化中心**：与 DNS Provider 系统共用凭证，自动完成 ACME DNS-01 验证全流程；支持向宝塔 Nginx、原生 Nginx、SSH 远程服务器、阿里云 CDN/ESA、Cloudflare 等多目标部署；可视化 DAG 工作流引擎支持并行分支、条件判断、失败重试和执行回放。

**专业管理后台**：基于 Arco Design React 构建，涵盖仪表盘、DNS 管理、CDN 管理、智能调度、SSL 证书、监控中心、系统管理七大模块，深色/浅色主题，桌面和移动端响应式。

## What Makes This Different

**真实能力建模，不编造 API。** 系统对每个 Provider 的套餐限制（如 DNSPod 免费版不支持分运营商线路，Cloudflare SaaS A 记录接入需要套餐授权）进行精确建模，不允许向用户展示实际不支持的操作选项。这避免了用户在生产环境遭遇"界面可操作但 API 失败"的陷阱。

**生产保护优先。** 默认 Dry Run，默认只读发现，敏感操作（NS 切换、SSL 替换、DNS 批量修改）需显式授权。运行中的生产网站（如 www.jdmv.net）受到明确的变更保护机制保障。

**插件化架构防止技术债。** DNS Provider、CDN Provider、证书部署目标均遵循统一接口，不支持的能力返回 `UnsupportedCapability` 而非静默失败。新增服务商只需实现适配器，不涉及核心调度代码。

**私有部署，数据自主。** 无 SaaS 依赖，所有配置、凭证、历史数据保存在用户自己的 MySQL 实例中。API 凭证和证书私钥在数据库中加密存储，不明文出现在日志、HTTP 响应或错误消息中。

**宝塔 Go 项目原生适配。** 主要部署路径为宝塔 Go 项目管理器，不用 systemd 冒充。同时提供原生 Linux、Docker、Docker Compose 四种部署方式的完整文档和脚本。

## Who This Serves

**主要用户：技术型网站运营者**
运营 1-20 个网站，使用多家 DNS/CDN 服务商，有中国大陆三网优化需求，需要自主控制所有基础设施配置。典型场景：有 Cloudflare 国际线路，同时有阿里云/腾讯云大陆 CDN，自有家宽节点作为备用，需要根据网络状况自动切换。对控制台操作熟悉，但希望减少手动重复工作和人为失误风险。

**次要用户：小型技术团队**
3-10 人的技术团队共同管理多个业务域名，需要操作审计、RBAC 权限控制，以及统一的故障响应流程。SSL 证书管理交给平台自动完成，团队成员不需要手动处理续期。

**系统用户（非人类）：Probe Agent 和 Node Agent**
运行于各网络节点的轻量 Go 程序，负责实时测速上报和受控配置变更执行。

## Success Criteria

- DNS 记录在三网调度下的可用性均在 99.9% 以上，自动切换响应时间不超过两个调度周期（默认 20 分钟）。
- SSL 证书到期前自动续期成功率 100%（排除 CA 侧故障）。
- 所有 DNS 变更均有完整快照，可在 5 分钟内完成安全回滚。
- 支持在宝塔 Go 项目管理器中完成完整部署、启停、日志查看和环境变量管理。
- Provider 插件新增服务商无需修改核心调度代码，仅新增适配器文件。
- 管理后台支持全部 7 个菜单模块，页面完整覆盖功能范围，无占位按钮或空实现。
- 通过安全检测：API 凭证加密存储、私钥保护、RBAC、CSRF/XSS/SSRF/SQL 注入防护全部通过。

## Scope

**纳入（V1 完整交付）**

- 多 DNS Provider 插件框架 + 阿里云 DNS、DNSPod、Cloudflare DNS 适配
- 多域名管理、DNS 记录 CRUD、历史快照、安全回滚
- DNS 迁移工作流（模式 A/B），Multi-DNS 作为独立高级功能
- 分布式 Probe Agent（Go）真实三网+香港/日本/新加坡探测
- 三网智能调度系统（评分算法、发布安全流程、熔断回滚）
- Cloudflare 管理（Zone、DNS、Proxy、Cache、SSL）+ Cloudflare for SaaS
- 自有节点注册管理 + Node Agent（Go）
- SSL ACME（Let's Encrypt/ZeroSSL）+ DNS-01/HTTP-01 验证
- SSL 多部署目标：宝塔 Nginx、原生 Nginx、SSH、阿里云 CDN/ESA、Cloudflare
- SSL 可视化 DAG 工作流引擎
- 监控告警（Telegram/SMTP/Webhook）
- 专业管理后台（React + Arco Design，7 大模块，深/浅色主题）
- RBAC + 审计日志
- 安装向导（首次安装流程，install.lock）
- 四种部署方式完整文档和脚本

**二期支持（架构预留，V1 不交付）**

- 华为云 DNS、DNS.com、自定义 DNS Provider
- 腾讯云 CDN/EdgeOne、华为云 CDN 证书部署
- 企业微信/飞书/钉钉告警渠道
- 阿里云 CDN 管理（读取状态，非 SmartCDN 直管）

**明确排除**

- 任何未经授权的生产 DNS/CDN/SSL 变更
- 使用未备案 CDN 节点为 www.jdmv.net 进行大陆直连
- Cloudflare 优选 IP 自动发布到正式 DNS（无套餐授权情况下）
- 对不受控的公共 VPN Gate 节点的接入支持
- 将 Certimate 实例作为外部依赖（原生实现，检查许可证后可参考源码）

## Vision

SmartCDN 未来两到三年的目标是成为中文技术圈私有化部署基础设施管理平台的标准选择。当平台覆盖了主流的中国 DNS/CDN 服务商生态，并被证明在生产环境中稳定运行后，可以向以下方向演进：

- 多租户 SaaS 化，服务需要托管管理能力但不想自部署的中小企业
- 更完整的 Cloudflare API 覆盖（Workers Routes、R2、Pages 等）
- 海外 CDN 服务商接入（BunnyCDN、Fastly、Akamai 等）
- 通过 Webhook 和 API 对外暴露自动化能力，接入 CI/CD 和 GitOps 工作流
- 移动端（iOS/Android）告警和审批应用

---

## 技术风险与已知约束摘要

> 以下条目为规划阶段已识别风险，详细分析见 addendum.md 和架构 ADR。

| 风险 | 等级 | 说明 |
|------|------|------|
| Cloudflare SaaS A 记录接入 | 高 | 需要付费套餐授权，标准模式仅支持 CNAME，不得假设 A 记录可用 |
| DNSPod 免费版线路限制 | 高 | 免费版不支持分运营商线路，需按实际套餐能力检测后决策 |
| 阿里云 DNS 分线路记录数量上限 | 中 | 不同套餐上限不同，调度系统必须尊重套餐限制 |
| 未备案域名 CDN 接入 | 高 | www.jdmv.net 未备案，不能使用境内 CDN 节点，只能走 Cloudflare 或境外节点 |
| 分布式探针真实性 | 高 | 探针必须真实部署于对应网络，不能以 HTTP Header 伪造 |
| 宝塔 Go 项目实际行为 | 中 | 宝塔 Go 项目管理器的具体启动/停止/环境变量机制需实际验证 |
| DNS TTL 传播时间 | 中 | DNS 变更不立即全网生效，系统不得宣称即时生效 |
| 现网 www.jdmv.net 保护 | 高 | 任何变更须明确授权，默认只读发现 |
