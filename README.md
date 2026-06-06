# Solo Workspace

> An open-source operating system for indie developers.

Manage projects, servers, domains, SSL certificates, deployments, and business assets — all from your terminal.

**Open Source · Plugin Architecture · Developer First**

---

## Directory Structure

```
solo-workspace/
├── cli/                   # CLI 客户端
│   └── go/                # Go CLI（插件架构）
│       ├── cmd/
│       ├── internal/      # 配置、输出、插件接口
│       ├── plugins/
│       ├── main.go
│       └── go.mod
├── web/
├── .solo.yaml             # 配置示例
├── README.md
└── LICENSE
```

## Installation

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

> **Verify:** `sw ssl check`

## Configuration

### Loading order

`sw` looks for config in this order (first found wins):

| Priority | Path | Notes |
|----------|------|-------|
| 1 | `-c <path>` / `--config <path>` | Manual override |
| 2 | `.solo.yaml` (current directory) | Project-level config |
| 3 | `~/.solo/config.yaml` | Global user config |
| 4 | (none) | Returns empty defaults |

### Typical usage

**Per-project config** — place `.solo.yaml` in your project root:

```yaml
# .solo.yaml
servers:
  my-vps:
    host: 123.123.123.123
    user: root
    port: 22

domains:
  - example.com
  - mysite.org

projects:
  my-app:
    path: /home/me/my-app
    description: My awesome project

notify:
  webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/..."
```

**Global config** — place at `~/.solo/config.yaml` for settings shared across all projects.

**Manual override** — point to any config file:

```bash
sw -c /path/to/custom.yaml ssl check
```

## Commands

```bash
sw ssl check                  # Check all domain SSL certificates
sw server list                # List all configured servers
sw server ssh <name>          # SSH into a server
```

## Plugin Architecture

Each plugin is a Lego brick:

```
cli/go/
├── cmd/                # CLI entry point (cobra)
├── internal/           # 配置、输出、插件接口
│   ├── config.go       # YAML config loader
│   ├── output.go       # 表格/JSON/Spinner/彩色输出
│   └── plugin.go       # Plugin interface
├── plugins/            # Plugin implementations
│   ├── ssl/            # SSL certificate management
│   └── server/         # Server management
└── main.go
```

To add a new plugin:
1. Create `cli/go/plugins/<name>/plugin.go`
2. Implement the cobra command
3. Register in `cli/go/cmd/root.go`

## Roadmap

### v0.1 (Current)
- [x] Plugin architecture
- [x] SSL certificate check
- [x] Server list & SSH
- [ ] Domain management
- [ ] Expiry notifications

### v0.2
- [ ] Docker container management
- [ ] GitHub repo integration
- [ ] Cost tracking per project/server
- [ ] Zsh completion plugin

### v1.0
- [ ] Web Dashboard
- [ ] Team collaboration
- [ ] Plugin marketplace

## License

MIT
