# Command Reference

<details open>
<summary><strong>English</strong> / <a href="#中文">中文</a></summary>

## Overview

```bash
sw                           # Show resource overview (same as sw all)
sw all                       # Show all configured resources
```

Overview display order: **Projects → Todos → Servers → Domains → Notifications → Env Vars → Secrets**

Projects and todos include an auto-increment **ID** column. Use IDs for update/delete/done/reopen operations.

Commands default to `list` when no subcommand is given — `sw project`, `sw todo`, `sw server`, `sw domain`, `sw secret`, and `sw env` all show their list view.

## Resource Management

```bash
# Project CRUD (update/delete by ID)
sw project                          # List all projects (same as sw project list)
sw project list
sw project add <name> --path <path> --desc "..."
sw project update <id> --path <path> --desc "..."
sw project delete <id>
sw project path <id>                # Print project absolute path (use: cd "$(sw project path 1)")

# Server management
sw server                           # List all servers (same as sw server list)
sw server list
sw server add <name> --host <ip> --user <user>  # Add a server
sw server update <name> --host <ip>              # Update server fields
sw server delete <name>                          # Delete a server
sw server ssh <name>

# Domain management
sw domain                           # List all domains (same as sw domain list)
sw domain list
sw domain add <domain>
sw domain delete <domain>
```

## SSL Certificate

```bash
sw ssl check                 # Check all domain SSL certificates
```

## Todo

```bash
sw todo                           # List all todos (same as sw todo list)
sw todo list                      # List all todos (with ID, Created, Updated, Note)
sw todo add <name> --desc "..."   # Add a todo item
sw todo update <id> --name <name>   # Rename a todo (optional)
sw todo update <id> --desc "..."  # Update description (at least one flag required)
sw todo note <id> <text>          # Add a note to a todo item
sw todo delete <id>               # Delete a todo by ID
sw todo done <id>                 # Mark a todo as done
sw todo reopen <id>               # Reopen a completed todo
sw todo stats                     # Show summary: total, pending, completed, with note, archived
sw todo archive run               # Archive todos inactive for 2+ weeks
sw todo archive list              # List archived todos
```

Stale todos are moved to `todos-archive.yaml` in the same directory as the active config file.

## Notification

```bash
sw notify test                # Send a test email notification
sw notify send "<subject>" "<body>"  # Send custom notification
```

## Configuration

```bash
sw config show                # Show config as YAML
sw config show --format json  # Show config as JSON
sw config export > backup.yaml   # Export config to file
sw config export --format json > backup.json
sw config import backup.yaml  # Import and merge config
sw config validate config.yaml   # Validate a config file
sw config set <path> <value>  # Set value by path (e.g. servers.my-vps.host 1.2.3.4)
sw config get <path>          # Get value by path
sw config delete <path>       # Delete key by path
```

## Environment Variables

```bash
sw env set DB_HOST localhost               # Store plaintext var
sw env set secret_db_password "p@ss"       # Store encrypted var
sw env get DB_HOST                         # Retrieve variable
sw env list                                # List all variables
sw env export > .env                       # Export as .env format
sw env export --prefix secret_             # Export only secret-prefixed vars
sw env delete DB_HOST                      # Remove variable
```

Data files (`env.yaml`, `secrets.enc`) live alongside the active config file — use `-c <path>` to control their location.

## Secrets

```bash
sw secret set api_key "sk-xxx"    # Store encrypted secret
sw secret get api_key             # Retrieve and decrypt
sw secret list                    # List all secret keys
sw secret delete api_key          # Delete a secret
```

## Time Log

```bash
sw log "fixed login bug"          # Record a timestamped log entry
sw log today                      # Show today's entries
sw log list                       # Show last 20 entries
sw log since 3d                   # Show entries from last 3 days
sw log since 24h                  # Show entries from last 24 hours
```

## Shell Completion

```bash
sw completion bash                     # Generate Bash completion script
sw completion install bash             # Install Bash completion
sw completion install zsh              # Install Zsh completion
sw completion install powershell       # Install PowerShell completion
sw completion uninstall bash           # Remove Bash completion
sw completion uninstall zsh            # Remove Zsh completion
sw completion uninstall powershell     # Remove PowerShell completion
```

After installing, reload your shell or open a new terminal session.

> **Windows Git Bash:** run `sw completion install bash` then `source ~/.bashrc`

## Interactive Mode

```bash
sw tui                       # Launch interactive TUI menu
```

CLI commands are the recommended workflow. The TUI is optional for quick navigation.

</details>

<details>
<summary id="中文"><strong>中文</strong> / <a href="#">English</a></summary>

## 概览

```bash
sw                           # 展示资源概览（等同 sw all）
sw all                       # 查看所有配置的资源
```

概览展示顺序：**项目 → 待办 → 服务器 → 域名 → 通知 → 环境变量 → 机密**

项目和待办带有自增 **ID** 列，更新/删除/完成/重开等操作请使用 ID。

命令默认等于 list：`sw project`、`sw todo`、`sw server`、`sw domain`、`sw secret`、`sw env` 都不需要加 `list`，直接展示列表。

## 资源管理

```bash
# 项目管理（update/delete 按 ID 操作）
sw project                          # 查看所有项目（等同 sw project list）
sw project list
sw project add <name> --path <path> --desc "..."
sw project update <id> --path <path> --desc "..."
sw project delete <id>
sw project path <id>                # 输出项目绝对路径（配合 cd 使用：cd "$(sw project path 1)"）

# 服务器管理
sw server                           # 查看所有服务器（等同 sw server list）
sw server list
sw server add <name> --host <ip> --user <user>  # 添加服务器
sw server update <name> --host <ip>              # 更新服务器信息
sw server delete <name>                          # 删除服务器
sw server ssh <name>

# 域名管理
sw domain                           # 查看所有域名（等同 sw domain list）
sw domain list
sw domain add <domain>
sw domain delete <domain>
```

## SSL 证书

```bash
sw ssl check                 # 检查所有域名的 SSL 证书
```

## 待办事项

```bash
sw todo                           # 查看所有待办（等同 sw todo list）
sw todo list                      # 查看所有待办（含 ID、创建/更新时间、备注标记）
sw todo add <name> --desc "..."   # 添加待办
sw todo update <id> --name <name>   # 重命名（可选）
sw todo update <id> --desc "..."  # 修改描述（至少指定一个参数）
sw todo note <id> <text>          # 给待办添加备注
sw todo delete <id>               # 按 ID 删除
sw todo done <id>                 # 标记完成
sw todo reopen <id>               # 重新打开
sw todo stats                     # 查看统计：总数/待完成/已完成/有备注/已归档
sw todo archive run               # 归档超过 2 周未更新的待办
sw todo archive list              # 查看已归档待办
```

过期待办会移动到与当前配置文件同目录下的 `todos-archive.yaml`。

## 通知

```bash
sw notify test                # 发送测试邮件
sw notify send "<主题>" "<正文>"  # 发送自定义通知
```

## 配置管理

```bash
sw config show                # 以 YAML 格式查看配置
sw config show --format json  # 以 JSON 格式查看配置
sw config export > backup.yaml   # 导出配置到文件
sw config export --format json > backup.json
sw config import backup.yaml  # 导入并合并配置
sw config validate config.yaml   # 验证配置文件
sw config set <path> <value>  # 按路径设值（如 servers.my-vps.host 1.2.3.4）
sw config get <path>          # 按路径取值
sw config delete <path>       # 按路径删除键
```

## 环境变量

```bash
sw env set DB_HOST localhost               # 存储明文变量
sw env set secret_db_password "p@ss"       # 存储加密变量
sw env get DB_HOST                         # 读取变量
sw env list                                # 查看所有变量
sw env export > .env                       # 导出为 .env 格式
sw env export --prefix secret_             # 仅导出 secret_ 前缀的变量
sw env delete DB_HOST                      # 删除变量
```
数据文件（`env.yaml`、`secrets.enc`）与当前使用的配置文件同级目录存放——使用 `-c <path>` 控制其位置。

## 机密管理

```bash
sw secret set api_key "sk-xxx"    # 存储加密机密
sw secret get api_key             # 读取并解密
sw secret list                    # 查看所有机密键名
sw secret delete api_key          # 删除机密
```

## 时间日志

```bash
sw log "修复了登录 bug"            # 记录一条时间日志
sw log today                      # 查看今天的日志
sw log list                       # 查看最近 20 条
sw log since 3d                   # 查看最近 3 天
sw log since 24h                  # 查看最近 24 小时
```

## Shell 补全

```bash
sw completion bash                     # 生成 Bash 补全脚本
sw completion install bash             # 安装 Bash 补全
sw completion install zsh              # 安装 Zsh 补全
sw completion install powershell       # 安装 PowerShell 补全
sw completion uninstall bash           # 卸载 Bash 补全
sw completion uninstall zsh            # 卸载 Zsh 补全
sw completion uninstall powershell     # 卸载 PowerShell 补全
```

安装后重新加载 shell 或打开新终端。

> **Windows Git Bash:** 运行 `sw completion install bash` 然后 `source ~/.bashrc`

## 交互模式

```bash
sw tui                       # 启动交互式 TUI 菜单
```

推荐以 CLI 命令为主流程，TUI 仅作可选快捷导航。

</details>
