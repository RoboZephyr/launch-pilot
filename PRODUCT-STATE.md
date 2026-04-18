# Launch Pilot — Product State

> **产品档案**：跨 iteration 累积的产品事实记录。由 forge release 成功后自动追加，feature mode 启动时读此档案建立上下文。

**Last updated**: 2026-04-18 (after v0.2.0 release)
**Slug**: `macos-launchd`
**Current version**: v0.2.0
**Repo**: https://github.com/A404coder/macos-launchd

---

## Product Overview

**Name**: Launch Pilot（原名 Launchboard）

**Objective**: 让 macOS 开发者通过 `brew install` 一条命令安装、在浏览器中一眼看到所有 launchd job 的运行状态，并能一键管理和诊断故障 —— 替代手动 `launchctl` CLI + 手写 plist 的痛苦工作流。

**Target User**: macOS 开发者和 power users（25-40 岁），日常使用 Homebrew，本地运行 3-10 个后台服务。主要人群：Solo Developer、Side Project Builder、DevOps/SRE、Indie macOS Developer、Data/ML Engineer。

**Business model**: 开源免费（MIT）+ Open Core/Freemium（Pro/Team tier 远期）

---

## Delivered Features

### v0.1.0 — Initial MVP (2026-04-16)

初始 5 个 P0 user stories 交付：

- **US-001 实时状态仪表盘**：列出所有 `~/Library/LaunchAgents` 和 `/Library/LaunchAgents` 下的 job；每个 job 显示 Label、PID、Last Exit Status、运行状态；颜色标记；Label 搜索；SSE 5s 自动刷新
- **US-002 一键操作**：每个 job 支持 Reload / Start / Stop，带确认弹窗；操作后自动刷新 + toast 通知；仅当前用户域（gui/UID）
- **US-003 日志查看**：每个 job 的 stdout/stderr 最近行（最多 10,000 行）；自动检测 `StandardOutPath/StandardErrorPath`；关键词搜索
- **US-004 诊断引擎**：6 项健康检查（exit code 映射、Program 路径存在、执行权限、plist owner/permission、日志路径）；红/黄/绿灯显示 + 一句话修复建议；只展示问题不自动修改 job
- **US-005 零摩擦安装**：Homebrew tap 分发单个 Go binary；启动即在 127.0.0.1 随机端口拉起 HTTP server；自动打开浏览器（`--no-open` 跳过）；`Ctrl+C` 优雅退出；`--port` 自定义端口

**附加能力**：Job classification（Mine / System / 3rd-party）、Multi-dimensional filtering（Category chips + 7-tab status + Only Mine toggle）、Label 正则校验防注入、plist mtime cache、Preact+Signals+HTM 嵌入式前端

### v0.2.0 — Scheduled / Completed / Offline Status Distinction (2026-04-18)

解决"周五 11:00 CalendarInterval 任务被 launchctl 正确触发、PID=0+exit=0、UI 仍显 offline"的状态误判 bug。

- **US-F1 scheduled**：PID=0 + exit=0 + 有周期配置（`StartInterval>0` 或 `StartCalendarInterval≠∅` 或 `RunAtLoad+已加载`）→ 新状态 `scheduled`；Job JSON 新增 `nextRunAt` 字段（CalendarInterval 纯函数计算 + 2 年 horizon guard；StartInterval 用 `lastRunAt+interval`）；UI `StatusTabs` 增加 `Scheduled` tab + 独立蓝色徽章 + hover 显示 Next run
- **US-F2 completed**：`lastRunAt ≤ recentWindow` 时 Status=`completed`（优先级高于 scheduled）；`lastRunAt` 从 `StandardOutPath/StandardErrorPath` mtime 较新者；`--recent-window` CLI 参数（`1m` ≤ window ≤ `24h`，默认 `10m`）；UI 绿色徽章 + hover 显示 Last run；无日志路径时 `lastRunAt=null` + 明示"unknown (no log path configured)"
- **US-F3 offline**：`launchctl list` 不返回但 `~/Library/LaunchAgents` 中有 plist → 合并为 `offline` 状态（旧行为是不显示）；`JobStatus` 枚举扩展为 6 值（running/error/completed/scheduled/stopped/offline）

**实现要点**：`DeriveStatus` 签名变更；`howett.net/plist` 用 `UnmarshalPlist` 兼容 dict/array；纯函数 next-fire 计算；SSE schema 向后兼容（新字段可空）；**零新增第三方依赖**

---

## Current Not Doing（累积，按版本可能解锁）

| 项 | 原因 | 可能解锁版本 |
|---|---|---|
| 可视化 plist 编辑器 | V1 聚焦"看 + 管 + 诊断"，写交给用户现有工具 | V2 (Free) |
| 远程管理多台 Mac | 安全攻击面大；偏离单机本地使用定位 | Team tier |
| 系统级 LaunchDaemons (`/Library/LaunchDaemons`, `/System/...`) | 需 root + privileged helper；Big Sur+ 已限制修改；安全风险 | 长期不做 |
| AI/LLM 辅助诊断 | V1 用确定性规则引擎覆盖 80% 常见问题；AI 诊断作为 Pro 功能 | V2 Pro |
| 通知告警（job 失败推送到系统/Slack/webhook） | Pro tier 功能；V1 是被动查看工具 | V2 Pro |
| 实时长连接（FSEvents / launchctl log stream） | SSE 5s 轮询已够；长连接是独立演进方向 | 待评估 |
| 精确运行历史（launchd 内部 job history） | Apple 无稳定 API；用 mtime 启发式近似 | 不做（Apple 不给） |
| 完整 cron 表达式解析器 | launchd CalendarInterval 只支持单值/缺失；不引入 robfig/cron | 不做 |
| 历史趋势图 / 运行次数统计 | 当前只回答"现在这一刻 job 是什么状态"，不做时间序列 | 远期独立 feature |
| UI 主题/徽章视觉重设计 | 仅扩展 CSS 变量，不重写组件 | 按需 |

---

## Tech Stack（当前）

| 维度 | 选择 |
|---|---|
| 后端 | Go（单 binary，stdlib + `howett.net/plist`；零其他第三方依赖） |
| 前端 | Preact + Signals + HTM（ESM vendored，`go:embed` 嵌入 binary） |
| 数据传输 | SSE（5s 轮询式 full snapshot push） |
| 分发 | Homebrew tap（`A404coder/tap/launch-pilot`，amd64 + arm64） |
| 测试 | `go test` 表驱动用例 + Playwright E2E |
| 平台 | macOS 13+ (Ventura+) |

---

## Key Architectural Decisions

- **纯函数 next-fire 计算**：`NextCalendarFire(entry, now) → time.Time`，便于表驱动测试；2 年 horizon guard 防止 `Day=31, Month=2` 等无解组合无限循环
- **SSE 向后兼容**：新增字段一律可空 JSON 字段，旧客户端忽略即可
- **零 root 安全模型**：仅操作当前用户域（gui/UID），不触碰 system domain，不引入 privileged helper
- **Localhost only**：HTTP server 仅绑定 127.0.0.1，不暴露到网络
- **单 binary 分发**：所有前端资源用 `go:embed` 打包，无运行时依赖；首屏 < 1s，100 个 job 列表刷新 < 500ms
- **Label injection 防御**：所有 label 调用 shell 前按 `[a-zA-Z0-9._-]+` 正则校验
- **Plist mtime cache**：避免重复磁盘读取未变更的 plist 文件

---

## Current Constraints

- macOS 13 (Ventura) 及以上
- 仅当前用户域（gui/UID），不支持 system domain
- `launchctl` 命令必须可用且输出格式可解析（Apple 未提供稳定 API）
- 开源 MIT license

---

## Known Risks / Open Questions

- `launchctl` 输出格式跨 macOS 版本差异 → 已有 fuzzy parser 容错，但需持续适配新版本
- Apple 未来可能限制 `launchctl` 的 CLI 接口 → fallback 方案未定
- `lastRunAt` 基于 log mtime 的启发式语义，UI 悬停文案已明示来源（"from log mtime"），若 >3 条 issue 质疑则考虑 `wtmp/launchctl dumpstate` 回退
