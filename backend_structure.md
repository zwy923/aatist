# Backend 结构总览

## 运行时组件（`cmd/`）
- `cmd/user-service/main.go`：完整实现的 REST 服务。
  - 启动顺序：加载 YAML/环境配置 → 初始化 zap 日志 → 建立 PostgreSQL（`internal/platform/db`）连接并运行 `migrations/` → 连接 Redis（会话/验证码/限流）→ 初始化 JWT（`internal/platform/auth`）→ 构造 `internal/user` 仓储、服务与 Gin 处理器。
  - Web 层：`/api/v1/auth` 提供注册/登录/刷新/登出/邮箱验证；`/api/v1/users/me` 支持个人档案读取与 PATCH 更新，`/api/v1/users/me/avatar` 处理多部分表单上传（统一走 S3/MinIO）。
  - 可选 RabbitMQ（`internal/platform/mq`）用于异步发送验证邮件；若连接失败只影响邮件流程。
  - 通过 `http.Server`+关机完成服务托管。
- `cmd/email-worker/main.go`：SendGrid 驱动的后台消费者。
  - 读取同一份配置，初始化日志与 RabbitMQ 连接，消费 `EmailVerificationRequest` 消息（定义见 `internal/user/model/email_message.go`），调用 `EmailService.SendVerificationEmail` 生成 `FRONTEND_URL/verify?token=...` 链接。
  - 依赖 `SENDGRID_API_KEY`/`FROM_EMAIL` 环境变量；失败时会写入结构化日志。
- `cmd/gateway/main.go`：轻量 API Gateway。
  - 负责公共 `/api/v1/auth/*` 以及 `/auth/*` 的反向代理，使用 `proxyToService("user-service", 8081, …)` 将请求透传到后端，同时转发请求头与 `X-Request-ID`。
  - 附加 CORS、请求 ID、可选 Redis（用于限流）以及 `GatewayAuthMiddleware`（JWT 校验+透传用户上下文）。
  - Protected 路由预留 `users/portfolio/opportunities/community` 等服务端口映射。
- `cmd/{community-service, opp-service, portfolio-service, ai-worker}/main.go`：目前仅含 `TODO` 占位，表明架构计划扩充成多服务体系。

## 平台基础层（`internal/platform/`）
- `auth/jwt.go`：封装 access/refresh token 生成与解析，供 `authService`、Gateway 使用。
- `cache/redis.go`：集中创建 go-redis v9 客户端，附带健康检查。
- `config/`：`load.go` 从 YAML+环境变量加载配置并填充默认值（HTTP 端口、JWT TTL、Redis addr等）；`config.go` 提供强类型结构体（App/Postgres/Redis/MQ/S3/JWT/Email）。
- `db/`：`postgres.go` 创建 sqlx 连接；`migrate.go` 用 `golang-migrate` 运行 `migrations/`。
- `http/server.go`（若有）提供通用 HTTP 启动工具；当前主要通过各服务 main 手动构建。
- `log/logger.go`：对 zap 的轻量封装，按 env 切换开发/生产配置。
- `metrics/metrics.go`：定义 Prometheus 计数器（注册/登录成功、锁定等），在 `auth_service.go` 中调用。
- `middleware/`：
  - `recovery.go` 捕获 panic。
  - `request_id.go` 注入/透传 `X-Request-ID`。
  - `cors.go` 依据环境变量控制跨域。
  - `trust_gateway.go` 信任上游 Gateway 设定的客户端 IP。
  - `rate_limit.go` 使用 Redis `INCR` + TTL 做滑窗限流。
  - `gateway_auth.go` 执行 JWT 认证并把用户信息写入请求头。
  - `auth.go`、`rbac.go` 等为将来服务提供的鉴权辅助。
- `mq/`：RabbitMQ 抽象，提供连接、发布 `PublishEmailVerification`、消费邮件任务等能力，亦带重连逻辑。
- `storage/s3.go`：封装 MinIO/S3 客户端、Bucket 初始化、对象上传与 `BuildPublicURL` 逻辑，现由 User Service 的头像与后续资产上传复用。

## 业务域模块（`internal/`）

### User 模块
- `model/user.go`：定义用户实体（与 `users` 表对应）及角色辅助方法；`email_message.go` 定义发送邮件所需的消息体。
- `repository/`：`repository.go` 约定接口；`postgres_repository.go` 提供 SQL 实现，覆盖查询、创建、登录信息更新、锁定、验证状态更新等操作。
- `service/`：
  - `auth_service.go`：核心业务逻辑。
    - 注册：密码校验、角色/学校信息归一化，自动验证 `@aalto.fi` 邮件，写库后生成 access/refresh token，并通过 `storeRefreshToken` 把 refresh token 写入 Redis （键 `refresh_token:<token>` → userID，TTL 30 天）。
    - 登录：含限流、锁定、密码匹配、验证状态检查、JWT 签发与 refresh token 存储。
    - Refresh/Logout：校验刷新 token 是否仍存在 Redis，轮换/删除对应键。
    - 邮箱验证：调用 `EmailVerificationService` 校验 token 并更新 `is_verified_email`。
  - `profile_service.go`：集中处理 Profile 字段归一化、Redis 限流（PATCH、Avatar）、头像上传（仅允许 JPEG/PNG/WebP、S3 域名校验、Prometheus 计数）与技能前缀搜索。
  - `email_verification.go`：管理邮箱验证 token。
    - Redis 键 `email_verify_token:<uuid>`，值为 `{"user_id":..,"email":..}` JSON，TTL 30 分钟。
    - `VerifyToken` 会从 Redis 取值→反序列化→确认用户/邮箱一致→删除键（一次性使用）。
    - `MarkEmailAsVerified` 更新数据库标记。
  - `service.go`：对外暴露接口别名，便于 handler 注入。
- `handler/handler.go & dto.go`：
  - 使用 Gin 绑定/校验请求体，调用 `authService` 方法并通过 `pkg/response` 统一返回格式。
  - `/users/me` DTO 带长度约束，PATCH 前附加 Redis 限流；头像上传接口校验 MIME、限制频率并返回 `file_size/last_updated_at`。
  - 注册成功后若 MQ 可用就生成验证 token 并发布 `EmailVerificationRequest`；错误由 `handleServiceError` 封装成 `AppError`。

### Portfolio / Opportunity / Community / AI Worker
- 当前目录下已预置 `handler/model/repository/service` 子目录，但多为空/待实现，代表未来会以与 user 模块类似的分层模式扩展。
- 这几个服务在 `cmd/` 中已有入口，Gateway 也预留了代理路径，方便后续接入。

### AI Worker
- `internal/aiworker/consumer`、`service` 为空目录，表示计划中的异步/AI 处理流程将在此实现。

## 共享工具（`pkg/`）
- `pkg/errs/errors.go`：定义 `AppError`、错误代码常量、HTTP 状态映射等，贯穿 handler 和响应层。
- `pkg/response/response.go`：统一 API 返回结构（`success/data/error`）。
- `pkg/types`、`pkg/utils`：通用类型和工具函数，当前使用较少但可扩展。

## 配置与部署
- `configs/config.example.yaml`：示例配置，包含 app/redis/postgres/mq/s3/jwt/email 等段落，并记录 access/refresh TTL 与前端 URL。
- `docker-compose.yml`：本地一键启动栈。
  - 服务：Postgres、Redis、RabbitMQ（含管理 UI）、MinIO、`user-service`、`email-worker`、`gateway`。
  - `user-service`、`gateway` 通过 `REDIS_ADDR=redis:6379`、`POSTGRES_DSN=postgres://...@postgres` 等方式连接容器网络。
  - `email-worker` 使用 `.env` 中的 SendGrid/Fronend URL 参数。
- `deploy/docker/`：各服务的 Dockerfile，负责多阶段构建与运行镜像。
- `migrations/001_create_users_table.sql`：初始 `users` 表结构。
- `migrations/002_add_profile_fields.sql`：补充昵称/头像/专业/技能/项目 JSONB（默认 `[]`），为 Profile/Avatar 功能提供列支持。

## 数据流与依赖关系
1. **注册流程**：`gateway -> user-service (/auth/register)`→写 `users` 表→生成 JWT/refresh token→写 Redis → 发布邮件消息（RabbitMQ）→`email-worker` 发 SendGrid 邮件（链接指向 `frontend/src/pages/Verify`）。
2. **邮箱验证**：用户访问 `GET /api/v1/auth/verify?token=` → Gateway 代理至 user-service → `emailVerificationSvc.VerifyToken` （Redis）→ `MarkEmailAsVerified` （Postgres）→ 返回成功并删除 Redis token。
3. **登录/刷新/登出**：完全由 user-service 处理，Redis 仅保存刷新 token；访问受限 API 时 gateway 负责 JWT 校验与转发。
4. **监控与限流**：Prometheus Counter 位于 `internal/platform/metrics`，新增 `avatar_upload_success_total`/`failure_total`；Redis 限流键包括 `login:<email>:<ip>`、`register:<fingerprint>:<ip>`、`rate:profile_update:<userID>`、`rate:avatar:<userID>`，覆盖登录/注册/档案 PATCH/头像上传。

借助上述分层（cmd → handler → service → repository → platform/pkg），后端可以较容易地扩展更多微服务，同时复用配置、日志、鉴权、消息与存储等基础设施。

