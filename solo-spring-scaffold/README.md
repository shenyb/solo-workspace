# solo-spring-scaffold

<details>
<summary>🇨🇳 中文</summary>

> Spring Boot 3.x 多模块项目脚手架生成器 — 用 CLI 一键生成企业级项目骨架。

`solo-spring-scaffold` 是一个 Python CLI 工具，根据模板生成标准化的 Spring Boot 3.x 多模块项目。开箱即用（H2 内存库），也可以一键切换到 MySQL + MyBatis-Plus 代码生成。

**这不是一个框架，而是一个项目生成器。** 它把我的架构决策打包成可复用的模板，让你不用每次从零搭项目。

---

## 快速开始

```bash
# 1. 生成项目
python solo-spring-scaffold init my-app --package com.example.myapp --jdk 21

# 2. 进入项目、启动（H2 内存库，零配置）
cd my-app
mvn spring-boot:run -pl my-app-web -am

# 3. 打开浏览器
# http://localhost:8080/swagger-ui.html
```

Windows 用户也可以用 `solo-spring-scaffold.bat`（需将所在目录加入 PATH）。

### 带 MySQL 代码生成

```bash
python solo-spring-scaffold init my-app \
    --package com.example.myapp \
    --jdk 21 \
    --db-url jdbc:mysql://192.168.1.100:3306/mydb \
    --db-user root \
    --db-pass xxx
```

生成的 `bin/gen-tables.sh` 可从数据库表自动生成 Entity/Mapper/Service/Controller。

---

## 前置依赖

| 工具 | 版本要求 |
|------|---------|
| JDK | 17 / 21 / 25 |
| Maven | 3.8+ |
| Python | 3.10+ |
| IDEA (推荐) | 2023+ |

---

## 生成的项目包含什么

```
my-app/
├── my-app-common/       # 公共模块
│   └── api/             # Result, PageResult 统一返回体
│   └── exception/       # BizException, ErrorCode, 全局异常处理器
│   └── web/             # LoggingAspect（请求日志AOP）, MDC Filter
├── my-app-dao/          # 数据层
│   └── config/          # MyBatis-Plus 配置 + MetaObjectHandler
│   └── user/            # User Entity + Mapper（演示）
│   └── generator/       # MyBatis-Plus 代码生成器（可选）
├── my-app-service/      # 业务层
│   └── user/            # UserService + DTO + 实现
├── my-app-web/          # Web 层
│   └── user/            # UserController（CRUD 演示）
│   └── resources/
│       ├── application.yml          # 主配置（H2 内存库）
│       ├── application-dev.yml      # 开发环境
│       ├── application-prod.yml     # 生产环境（MySQL）
│       ├── logback-spring.xml       # 日志分级切割
│       ├── db/migration/            # Flyway 迁移脚本
│       └── messages/validation.properties  # 校验国际化
├── bin/                              # 部署脚本
│   ├── start.sh / start.bat
│   ├── stop.sh / stop.bat
│   ├── deploy.sh
│   └── gen-tables.sh / .bat（可选）
└── pom.xml                           # 父 POM
```

### 内置的技术选型

| 维度 | 选择 | 理由 |
|------|------|------|
| ORM | MyBatis-Plus 3.5.9 | 灵活写 SQL，分页/自动填充开箱即用 |
| 数据库迁移 | Flyway | 版本化、可回滚、团队协作友好 |
| API 文档 | SpringDoc OpenAPI 2.6 | 原生支持 OpenAPI 3，UI 现代 |
| 参数校验 | Jakarta Validation + 国际化 | 统一错误消息 |
| 日志 | Logback + MDC traceId | 分级切割、分布式追踪准备 |
| 开发期数据库 | H2（内存） | 零配置启动，PR 评审无需搭数据库 |

---

> **⚠️ 安全提醒**
>
> 生成的项目 **不包含 Spring Security / JWT / OAuth2**。如果需要认证和权限控制，请自行添加。
> 请勿在生产环境直接使用默认配置部署到公网。

## Roadmap（计划中）

- [ ] Spring Security + JWT 认证模板
- [ ] 单元/集成测试模板（TestContainers）
- [ ] Redis + Caffeine 缓存配置
- [ ] Dockerfile + docker-compose.yml
- [ ] Spring Boot Actuator + Prometheus
- [ ] 多 Domain 示例（订单、商品）
- [ ] 文件上传/下载示例
- [ ] GitHub Actions CI 模板
- [ ] 交互式 CLI 向导

---

## License

[MIT](../LICENSE)

</details>

<details open>
<summary>🇬🇧 English</summary>

> Spring Boot 3.x multi-module project scaffold generator — generate a production-grade project skeleton from CLI in seconds.

`solo-spring-scaffold` is a Python CLI tool that generates standardized Spring Boot 3.x multi-module projects from templates. Ships with H2 in-memory database (zero-config startup) and optional MySQL + MyBatis-Plus code generation.

**This is a project generator, not a framework.** It packages architectural decisions into reusable templates so you don't start from scratch every time.

---

## Quick Start

```bash
# 1. Generate a project
python solo-spring-scaffold init my-app --package com.example.myapp --jdk 21

# 2. Enter project directory and start (H2 in-memory, zero config)
cd my-app
mvn spring-boot:run -pl my-app-web -am

# 3. Open browser
# http://localhost:8080/swagger-ui.html
```

Windows users can also use `solo-spring-scaffold.bat` (add its directory to PATH first).

### With MySQL Code Generation

```bash
python solo-spring-scaffold init my-app \
    --package com.example.myapp \
    --jdk 21 \
    --db-url jdbc:mysql://192.168.1.100:3306/mydb \
    --db-user root \
    --db-pass xxx
```

The generated `bin/gen-tables.sh` auto-generates Entity/Mapper/Service/Controller from database tables.

---

## Prerequisites

| Tool | Version |
|------|---------|
| JDK | 17 / 21 / 25 |
| Maven | 3.8+ |
| Python | 3.10+ |
| IDEA (recommended) | 2023+ |

---

## What You Get

```
my-app/
├── my-app-common/       # Shared module
│   └── api/             # Result, PageResult unified response
│   └── exception/       # BizException, ErrorCode, global handler
│   └── web/             # LoggingAspect, MDC Filter
├── my-app-dao/          # Data layer
│   └── config/          # MyBatis-Plus config + MetaObjectHandler
│   └── user/            # User Entity + Mapper (demo)
│   └── generator/       # Code generator (optional)
├── my-app-service/      # Business layer
│   └── user/            # UserService + DTO + impl
├── my-app-web/          # Web layer
│   └── user/            # UserController (CRUD demo)
│   └── resources/
│       ├── application.yml          # Main config (H2)
│       ├── application-dev.yml      # Dev profile
│       ├── application-prod.yml     # Prod profile (MySQL)
│       ├── logback-spring.xml       # Logging with rotation
│       ├── db/migration/            # Flyway migrations
│       └── messages/validation.properties  # i18n validation
├── bin/                              # Scripts
│   ├── start.sh / start.bat
│   ├── stop.sh / stop.bat
│   ├── deploy.sh
│   └── gen-tables.sh / .bat (optional)
└── pom.xml                           # Parent POM
```

### Built-in Tech Stack

| Area | Choice | Why |
|------|--------|-----|
| ORM | MyBatis-Plus 3.5.9 | Flexible SQL, paging + auto-fill out of the box |
| Migration | Flyway | Versioned, rollback-friendly, team-collab ready |
| API Docs | SpringDoc OpenAPI 2.6 | Native OpenAPI 3, modern Swagger UI |
| Validation | Jakarta Validation + i18n | Unified error messages |
| Logging | Logback + MDC traceId | Level-based rotation, distributed tracing ready |
| Dev DB | H2 (in-memory) | Zero-config startup, no DB needed for PRs |

---

> **⚠️ Security Notice**
>
> The generated project does **NOT include Spring Security / JWT / OAuth2**. Add authentication and authorization before deploying to production.
> Do NOT deploy the default configuration to a public network.

## Roadmap

- [ ] Spring Security + JWT template
- [ ] Unit/integration test templates (TestContainers)
- [ ] Redis + Caffeine caching
- [ ] Dockerfile + docker-compose.yml
- [ ] Spring Boot Actuator + Prometheus
- [ ] Multi-domain examples (order, product)
- [ ] File upload/download example
- [ ] GitHub Actions CI template
- [ ] Interactive CLI wizard

---

## License

[MIT](../LICENSE)

</details>
