[English](README.md)

# Solo Workspace

> 独立开发者的开源操作系统。

在终端里管理项目、服务器、域名、SSL 证书、环境变量、机密信息等——一站式搞定。

**开源 · 插件架构 · 开发者优先**

---

## 为什么需要 Solo Workspace？

作为独立开发者，你要同时应付一堆工具：终端连服务器、电子表格记域名、便签写待办、`.env` 文件散落各处、手动查 SSL。**Solo Workspace** 把它们统一到一个 CLI 里——你的独立开发工作台的唯一真相源。

- **一个配置文件**管理服务器、域名、项目和待办
- **加密存储** API key，不再明文裸奔
- **插件架构**——用 Go 包扩展，零框架锁定
- **为独立开发者而生**——无 SaaS、无云依赖，数据留在本地

---

## 功能一览

| 分类 | 能力 |
|------|------|
| 🖥️ 服务器 | 列表查看、增删改、SSH 连接 |
| 🌐 域名 | 域名追踪、SSL 证书检查 |
| 📁 项目 | 本地项目增删改查，自增 ID |
| ✅ 待办 | 任务管理，支持编辑、完成/重开（按 ID）；超过 2 周未更新可手动归档 |
| 🔐 机密 | AES-256-GCM 加密存储 API key 和 token |
| 🌍 环境变量 | 集中式 `.env` 管理，支持加密 |
| 📧 通知 | SMTP 邮件告警（域名到期、自定义消息） |
| ⚙️ 配置 | YAML/JSON 导入导出，按路径增删改查 |
| 📋 概览 | `sw` 无参数展示全部资源 |
| 🎮 交互菜单 | 可选 TUI 导航（`sw tui`） |
| 📦 补全 | Bash / Zsh / PowerShell Tab 补全 |

![sw all](cli/docs/img/sw-all.png)

---

## 快速体验

```bash
# 添加服务器、域名、项目
sw server add my-vps --host 1.2.3.4 --user root --port 22
sw domain add example.com
sw project add my-saas --path ~/code/my-saas --desc "我的 SaaS 产品"

# 检查所有域名 SSL 证书
sw ssl check

# 安全存储 API key
sw secret set stripe_key "sk_live_xxx"

# 一览全局（直接运行 sw 也可以）
sw

# 按 ID 管理待办
sw todo add fix-bug --desc "修复登录问题"
sw todo update 1 --desc "修复 OAuth 登录"
sw todo done 1
sw todo archive run              # 归档超过 2 周未更新的待办
sw todo archive list             # 查看已归档待办（todos-archive.yaml）

# 可选交互菜单
sw tui
```
---

## 安装

### macOS / Linux
```bash
cd cli/go && go build -o ~/bin/sw . && cd -
```

### Windows (Git Bash)
```bash
cd cli/go && go build -o ~/bin/sw.exe . && cd -
```

### Windows (PowerShell)
```powershell
cd cli\go
go build -o "$env:USERPROFILE\bin\sw.exe" .
```

> **验证安装:** `sw ssl check`

如果 `~/bin` 不在 PATH 中，请手动添加。

### Shell 补全

```bash
sw completion install bash   # 或 zsh, fish, powershell
```

![shell 补全演示](cli/docs/img/completion.png)

---

## 配置

SW 按以下优先级加载配置（先找到的生效）：

| 优先级 | 路径 | 用途 |
|--------|------|------|
| 1 | `-c <path>` / `--config <path>` | 手动指定 |
| 2 | `~/.solo/config.yaml` | 全局配置（跨项目） |
| 3 | `.solo.yaml`（当前目录） | 项目级配置 |
| 4 | _（无）_ | 空默认值 |

数据文件（`env.yaml`、`secrets.enc`）与当前使用的配置文件同级目录存放——默认 `~/.solo/config.yaml` 时在 `~/.solo/`；使用 `-c /path/to/config.yaml` 时跟随到 `/path/to/`。

**最小 `~/.solo/config.yaml` 示例：**

```yaml
servers:
  my-vps:
    host: 123.123.123.123
    user: root
    port: 22

domains:
  - example.com

notify:
  email:
    enabled: true
    host: smtp.example.com
    port: 587
    username: user@example.com
    password: app-password
    from: user@example.com
    to:
      - admin@example.com
```

> 📖 完整命令参考：[cli/docs/command.md](cli/docs/command.md)

---

## 插件架构

每个功能都是独立插件——像乐高积木，可自由替换和扩展：

```
cli/go/
├── cmd/                # CLI 入口（cobra + TUI）
├── internal/           # 配置、输出、插件接口
├── plugins/
│   ├── ssl/            # SSL 证书管理
│   ├── server/         # 服务器管理
│   ├── domain/         # 域名管理
│   ├── project/        # 项目增删改查
│   ├── todo/           # 待办管理
│   ├── notify/         # 邮件通知
│   ├── config/         # 配置导入导出/设值取值
│   ├── env/            # 环境变量管理
│   └── secret/         # AES-256-GCM 加密机密
└── main.go
```

**三步添加插件：**
1. 创建 `cli/go/plugins/<name>/plugin.go`
2. 实现 cobra 命令
3. 在 `cli/go/cmd/root.go` 注册

---

## 路线图

| 版本 | 状态 | 亮点 |
|------|------|------|
| v0.1 | ✅ 已完成 | 插件架构、SSL 检查、服务器 SSH、域名、待办、通知 |
| v0.2 | ✅ 当前 | 环境变量、机密加密、配置导入导出、ID 化项目/待办、概览与 TUI 优化 |
| v0.3 | 🔨 计划中 | 项目关联、成本追踪、SQLite 后端 |
| v0.4 | 📋 计划中 | Docker 集成、GitHub 集成 |
| v1.0 | 🚀 远期 | Web 仪表盘、插件市场 |

> 📖 完整路线图与待办池：[cli/docs/roadmap.md](cli/docs/roadmap.md)

---

## 参与贡献

欢迎贡献！插件架构让添加新功能非常简单。

1. Fork 仓库
2. 在 `cli/go/plugins/<name>/` 下创建插件
3. 在 `cli/go/cmd/root.go` 注册
4. 提交 PR

---

## 许可证

MIT © Solo Workspace
