# Roadmap

<details open>
<summary><strong>English</strong> / <a href="#中文-1">中文</a></summary>

## v0.1 ✅
- [x] Plugin architecture
- [x] SSL certificate check
- [x] Server list & SSH
- [x] Domain add/delete/list
- [x] Todo management
- [x] Email notification (SMTP) on domain expiry

## v0.2 ✅ (Current)
- [x] Environment variable management — .env hub + encryption
- [x] CLI dynamic metadata — config set/get/delete via JSON path
- [x] Interactive TUI — `sw tui` for optional menu navigation
- [x] Default overview — `sw` with no args shows resource summary (same as `sw all`)
- [x] JSON config import/export — `config import/export`
- [x] Secret management — AES-256-GCM encrypted storage for API keys / tokens / passwords
- [x] Config priority adjustment — `~/.solo/config.yaml` takes precedence over `./.solo.yaml`
- [x] Project CRUD (`project add/update/delete`)
- [x] Project & todo auto-increment IDs — CRUD by ID for easier management
- [x] Project path — `project path <id>` prints absolute path for `cd`
- [x] Todo update — `todo update <id>` with `--name` / `--desc`
- [x] Todo notes — `todo note <id> <text>` for per-item notes
- [x] Todo stats — `todo stats` summary (total, pending, done, with note, archived)
- [x] Time log — `log` plugin for quick timestamped daily entries
- [x] Default list behavior — all resource commands default to `list` without subcommand
- [x] Overview improvements — projects first, then todos; secrets shown in `sw all`
- [x] Table & TUI formatting — aligned column separators, stable terminal restore
- [x] PowerShell / Git Bash completion installation

## v0.3 (Planned)
- [ ] Project relationships
- [ ] Cost tracking
- [ ] SQLite backend
- [ ] Indie Business Tracking

## v0.4 (Planned)
- [ ] Docker integration
- [ ] GitHub integration

## v1.0 (Future)
- [ ] Web Dashboard
- [ ] Plugin Marketplace

### Optional / Backlog
- [ ] Abstract storage interface (SQLite)
- [ ] Team collaboration
- [ ] Community task board
- [ ] DNS propagation checker — verify global resolution after record changes
- [ ] Let's Encrypt auto-issuance — local application + deploy to server
- [ ] Auto certificate renewal — renew before expiry + reload nginx
- [ ] Certificate deployment — push local cert to multiple servers
- [ ] Uptime monitoring — HTTP health checks with downtime alerts
- [ ] Project scaffolding — `sw new <template>` to generate project skeletons
- [ ] Server cost tracking — monthly/per-project cloud cost statistics
- [ ] Project P&L — revenue minus cost, per project
- [ ] Vercel / Netlify integration — deployment status, environment variables
- [ ] Sentry integration — error summaries, trends
- [ ] Custom plugin install — `sw plugin install <source>`
- [ ] Task module for projects — task CRUD / status management
- [ ] Cloudflare DNS management — CRUD for DNS records / Cloudflare API integration (DNS + Workers + Pages)
- [ ] Bubble Tea terminal dashboard
- [ ] Link domains/servers to projects
- [ ] Project management enhancements — tags, search, filter, sort
- [ ] Docker container management
- [ ] GitHub repo integration
- [ ] Cost tracking per project/server

</details>

<details>
<summary id="中文-1"><strong>中文</strong> / <a href="#">English</a></summary>

## v0.1 ✅
- [x] 插件架构
- [x] SSL 证书检查
- [x] 服务器列表与 SSH
- [x] 域名增删查
- [x] 待办事项管理
- [x] 邮件通知（域名到期 SMTP 提醒）

## v0.2 ✅（当前版本）
- [x] 环境变量管理 — .env 集中管理 + 加密
- [x] CLI 动态元数据 — config set/get/delete 按 JSON Path 操作
- [x] 交互式 TUI — `sw tui` 可选菜单导航
- [x] 默认概览 — `sw` 无参数展示资源总览（等同 `sw all`）
- [x] JSON 配置导入导出 — `config import/export`
- [x] 机密管理 — AES-256-GCM 加密存储 API key / token / password
- [x] 配置优先级调整 — `~/.solo/config.yaml` 优先于 `./.solo.yaml`
- [x] 项目增删改查（`project add/update/delete`）
- [x] 项目 & 待办自增 ID — 按 ID 操作，管理更方便
- [x] 项目路径跳转 — `project path <id>` 输出绝对路径，配合 `cd` 使用
- [x] 待办编辑 — `todo update <id>` 支持 `--name` / `--desc`
- [x] 待办备注 — `todo note <id> <text>` 给待办添加备注
- [x] 待办统计 — `todo stats` 展示总数/待完成/已完成/有备注/已归档
- [x] 时间日志 — `log` 插件，带时间戳的快捷日志
- [x] 默认列表 — 所有资源命令无子命令时默认展示列表
- [x] 概览优化 — 项目优先、待办其次；`sw all` 展示机密名称
- [x] 表格 & TUI 格式 — 列分隔符对齐、终端状态稳定恢复
- [x] PowerShell / Git Bash 补全安装

## v0.3（计划中）
- [ ] 项目关联关系
- [ ] 成本追踪
- [ ] SQLite 存储后端

## v0.4（计划中）
- [ ] Docker 集成
- [ ] GitHub 集成

## v1.0（远期）
- [ ] Web 仪表盘
- [ ] 插件市场

### 可选 / 待办池
- [ ] 抽象存储接口（自由选择实现）
- [ ] 团队协作
- [ ] 社区任务看板
- [ ] DNS 传播检查 — 修改记录后验证全球生效
- [ ] Let's Encrypt 自动签发 — 本地申请 + 部署到服务器
- [ ] 证书自动续期 — 到期前自动 renew + reload nginx
- [ ] 证书部署 — 本地证书推送到多台服务器
- [ ] 运行监控 — HTTP 探活，宕机通知
- [ ] 项目脚手架 — `sw new <template>` 生成项目骨架
- [ ] 服务器成本追踪 — 按月/按项目统计云服务器费用
- [ ] 项目盈亏 — 收入减成本，按项目
- [ ] Vercel / Netlify 集成 — 部署状态、环境变量
- [ ] Sentry 集成 — 错误摘要、趋势
- [ ] 自定义插件安装 — `sw plugin install <source>`
- [ ] 项目任务模块 — 任务增删改查 / 状态管理
- [ ] Cloudflare DNS 管理 — DNS 记录增删改查 / Cloudflare API 集成（DNS + Workers + Pages）
- [ ] Bubble Tea 终端仪表盘
- [ ] 域名/服务器关联到项目
- [ ] 项目管理增强 — 标签、搜索、过滤、排序
- [ ] Docker 容器管理
- [ ] GitHub 仓库集成
- [ ] 按项目/服务器的成本追踪

</details>
