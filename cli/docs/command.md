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

## Resource Management

```bash
# Project CRUD (update/delete by ID)
sw project list
sw project add <name> --path <path> --desc "..."
sw project update <id> --path <path> --desc "..."
sw project delete <id>

# Server management
sw server list
sw server ssh <name>

# Domain management
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
sw todo list                      # List all todos (with ID)
sw todo add <name> --desc "..."   # Add a todo item
sw todo update <id> --name <name>   # Rename a todo (optional)
sw todo update <id> --desc "..."  # Update description (at least one flag required)
sw todo delete <id>               # Delete a todo by ID
sw todo done <id>                 # Mark a todo as done
sw todo reopen <id>               # Reopen a completed todo
```

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

## Secrets

```bash
sw secret set api_key "sk-xxx"    # Store encrypted secret
sw secret get api_key             # Retrieve and decrypt
sw secret list                    # List all secret keys
sw secret delete api_key          # Delete a secret
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

## 资源管理

```bash
# 项目管理（update/delete 按 ID 操作）
sw project list
sw project add <name> --path <path> --desc "..."
sw project update <id> --path <path> --desc "..."
sw project delete <id>

# 服务器管理
sw server list
sw server ssh <name>

# 域名管理
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
sw todo list                      # 查看所有待办（含 ID）
sw todo add <name> --desc "..."   # 添加待办
sw todo update <id> --name <name>   # 重命名（可选）
sw todo update <id> --desc "..."  # 修改描述（至少指定一个参数）
sw todo delete <id>               # 按 ID 删除
sw todo done <id>                 # 标记完成
sw todo reopen <id>               # 重新打开
```

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

## 机密管理

```bash
sw secret set api_key "sk-xxx"    # 存储加密机密
sw secret get api_key             # 读取并解密
sw secret list                    # 查看所有机密键名
sw secret delete api_key          # 删除机密
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
