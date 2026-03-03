# 独立游戏工作室全矩阵中台平台

基于 Go + PostgreSQL + Redis 的云原生游戏工作室中台系统。

## 技术栈

- **后端**: Go 1.23, Gin, pgx, go-redis
- **数据库**: PostgreSQL 16, Redis 7
- **认证**: JWT (Access Token + Refresh Token)
- **日志**: Zap (结构化日志)
- **配置**: Viper (支持环境变量覆盖)

## 快速开始

### 1. 安装依赖

```bash
make setup
```

### 2. 启动基础设施

```bash
make docker-up
```

### 3. 运行数据库迁移

```bash
# 需要先安装 golang-migrate
# https://github.com/golang-migrate/migrate

make migrate-up
```

### 4. 配置环境变量

复制 `.env.example` 到 `.env` 并修改配置:

```bash
cp .env.example .env
```

### 5. 运行服务

```bash
make run
```

服务将在 `http://localhost:8080` 启动。

## API 端点

### 健康检查

- `GET /health` - 健康检查
- `GET /ready` - 就绪检查

### 认证 (v1)

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新令牌
- `POST /api/v1/auth/logout` - 用户登出 (需要认证)

## 项目结构

```
go-web/
├── cmd/server/          # 应用入口
├── configs/             # 配置文件
├── internal/            # 私有代码
│   ├── domain/          # 领域实体和接口
│   ├── usecase/         # 业务逻辑
│   ├── infra/           # 基础设施实现
│   ├── transport/       # HTTP 传输层
│   └── pkg/             # 内部工具包
├── migrations/          # 数据库迁移
├── docker-compose.yml   # Docker 编排
└── Makefile            # 构建脚本
```

## 开发规范

详见 `skill.md` 文件，包含完整的编码规范、架构设计和最佳实践。

## 许可证

MIT
