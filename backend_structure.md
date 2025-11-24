# Backend 结构总览

## 架构概览

本后端采用微服务架构，通过 API Gateway 统一入口，各服务独立部署。核心设计原则：
- **领域驱动设计（DDD）**：每个服务代表一个业务域
- **服务解耦**：Notification 作为通用域，与 User 无强耦合
- **统一平台层**：共享基础设施（配置、日志、数据库、缓存等）
- **RESTful API**：通过 Gateway 统一路由和认证

## 运行时组件（`cmd/`）

### 已实现服务

#### 1. `cmd/user-service/main.go` - 用户服务
**状态**：✅ 完全实现

**功能**：
- 用户注册、登录、登出、刷新 Token
- 邮箱验证（通过 RabbitMQ 异步发送）
- 用户 Profile 管理（读取、更新、头像上传）
- 可用性管理（每周小时数、情绪状态、周可用性）
- 收藏项管理（保存/取消保存项目、机会、用户）
- 通过 HTTP 客户端调用 notification-service 发送通知
- 通过 HTTP 客户端调用 file-service 上传文件（头像等）

**启动流程**：
1. 加载 YAML/环境配置
2. 初始化 zap 日志
3. 连接 PostgreSQL 并运行 migrations
4. 连接 Redis（会话/验证码/限流）
5. 初始化 JWT（access/refresh token）
6. 初始化 File Service Client（通过 Gateway 调用 file-service）
7. 可选连接 RabbitMQ（邮件队列）
8. 构造仓储、服务与 Gin 处理器
9. 启动 HTTP 服务器

**API 端点**：
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新 Token
- `POST /api/v1/auth/logout` - 登出
- `POST /api/v1/auth/verify-email` - 请求邮箱验证
- `GET /api/v1/auth/verify` - 验证邮箱（通过 token）
- `GET /api/v1/users/:id` - 获取公开用户信息（无需认证）
- `GET /api/v1/users/me` - 获取当前用户信息（需认证）
- `PATCH /api/v1/users/me` - 更新当前用户信息（需认证）
- `POST /api/v1/users/me/avatar` - 上传头像（需认证）
- `GET /api/v1/users/me/availability` - 获取可用性（需认证）
- `PATCH /api/v1/users/me/availability` - 更新可用性（需认证）
- `GET /api/v1/users/me/saved` - 获取收藏项（需认证）
- `POST /api/v1/users/me/saved` - 保存项目/机会/用户（需认证）
- `DELETE /api/v1/users/me/saved` - 取消收藏（需认证）

**端口**：8081

**依赖**：
- PostgreSQL（用户数据）
- Redis（会话、验证码、限流）
- RabbitMQ（可选，邮件队列）
- Gateway（所有服务间调用通过 Gateway 的内部路由）
  - 通过 Gateway 调用 notification-service
  - 通过 Gateway 调用 file-service

---

#### 2. `cmd/portfolio-service/main.go` - 作品集服务
**状态**：✅ 完全实现

**功能**：
- 项目 CRUD 操作
- 用户作品集展示（支持可见性控制）
- 通过 HTTP 客户端调用 user-service 检查用户可见性

**启动流程**：
1. 加载配置
2. 初始化日志
3. 连接 PostgreSQL
4. 初始化仓储和服务
5. 创建 HTTP User Service Client（传入 baseURL，用于检查可见性）
6. 启动 HTTP 服务器

**User Service Client 配置**：
- 从环境变量 `USER_SERVICE_URL` 读取，默认 `http://user-service:8081`
- 必须显式传入 baseURL 参数创建客户端

**API 端点**：
- `GET /api/v1/portfolio/:id` - 获取项目详情（公开）
- `GET /api/v1/users/:id/portfolio` - 获取用户作品集（公开，受可见性控制）
- `GET /api/v1/users/me/portfolio` - 获取当前用户作品集（需认证）
- `POST /api/v1/users/me/portfolio` - 创建项目（需认证）
- `PUT /api/v1/users/me/portfolio/:id` - 更新项目（需认证）
- `DELETE /api/v1/users/me/portfolio/:id` - 删除项目（需认证）

**端口**：8082

**依赖**：
- PostgreSQL（项目数据）
- Gateway（通过内部路由调用 User Service 检查用户可见性）

---

#### 3. `cmd/notification-service/main.go` - 通知服务
**状态**：✅ 完全实现

**功能**：
- 创建通知（供其他服务调用）
- 获取用户通知列表
- 获取未读通知数量
- 标记通知为已读
- 标记所有通知为已读
- 删除通知

**设计理念**：
- **通用域服务**：Notification 作为独立服务，与 User 无强耦合
- **支持多种通知类型**：
  - `profile_saved` - 用户资料被保存
  - `follow` - 关注某人
  - `project_saved` - 项目被保存
  - `comment` - 评论
  - `opportunity_match` - 机会匹配
  - `project_invite` - 项目邀请
  - `ai_summary_finished` - AI 摘要完成
  - `system` - 系统通知
  - `weekly_digest` - 每周摘要
  - `message` - 消息

**启动流程**：
1. 加载配置
2. 初始化日志
3. 连接 PostgreSQL
4. 运行数据库迁移
5. 初始化仓储和服务
6. 设置路由（内部 API 和用户 API 分离）
7. 启动 HTTP 服务器

**安全特性**：
- 内部 API 使用 `RequireInternalCall` 中间件验证
- 用户 API 使用 `TrustGatewayMiddleware` + `RequireGatewayAuth` 验证
- 支持可选的内部 Token 验证（通过 `INTERNAL_API_TOKEN` 环境变量）
- 支持可选的 IP 白名单（通过 `INTERNAL_ALLOWED_IPS` 环境变量）

**API 端点**：

**内部 API**（供其他服务调用，需内部认证，通过 Gateway 访问）：
- `POST /api/v1/internal/notification/notifications` - 创建通知（通过 Gateway：`/api/v1/internal/notification/notifications`）

**用户 API**（需认证）：
- `GET /api/v1/me/notifications` - 获取通知列表
- `GET /api/v1/me/notifications/unread-count` - 获取未读数量
- `PUT /api/v1/me/notifications/:id/read` - 标记为已读
- `PUT /api/v1/me/notifications/read-all` - 标记全部已读
- `DELETE /api/v1/me/notifications/:id` - 删除通知

**端口**：8085

**依赖**：
- PostgreSQL（通知数据）

**安全配置**：
- `INTERNAL_API_TOKEN`（可选）- 内部 API Token，如果设置则必须验证
- `INTERNAL_ALLOWED_IPS`（可选）- 允许的内部 IP 列表（逗号分隔）

---

#### 4. `cmd/file-service/main.go` - 文件服务
**状态**：✅ 完全实现

**功能**：
- 统一文件上传管理（头像、项目封面、帖子图片、简历、AI 输出等）
- 文件元数据存储和管理
- 支持直接上传和预签名 URL 上传（前端直传 S3）
- 文件删除（同时删除 S3 对象和数据库记录）
- 文件大小和 MIME 类型验证
- 基于文件类型的大小限制

**设计理念**：
- **独立服务**：将文件上传从 user-service 解耦，统一管理所有文件
- **支持多种文件类型**：
  - `avatar` - 用户头像
  - `project_cover` - 项目封面图
  - `post_image` - 社区帖子图片
  - `resume` - 简历 PDF
  - `ai_output` - AI 输出文件
  - `other` - 其他文件

**启动流程**：
1. 加载配置
2. 初始化日志
3. 连接 PostgreSQL
4. 运行数据库迁移
5. 初始化 S3/MinIO 存储
6. 初始化仓储和服务
7. 设置路由（内部 API 和用户 API 分离）
8. 启动 HTTP 服务器

**安全特性**：
- 内部 API 使用 `RequireInternalCall` 中间件验证
- 用户 API 使用 `TrustGatewayMiddleware` + `RequireGatewayAuth` 验证
- 支持可选的内部 Token 验证（通过 `INTERNAL_API_TOKEN` 环境变量）
- 支持可选的 IP 白名单（通过 `INTERNAL_ALLOWED_IPS` 环境变量）
- 文件大小验证（基于文件类型）
- Content-Type 白名单验证
- Magic number 验证（防止 content-type 伪造）

**文件大小限制**：
- Avatar: 5MB
- Project Cover/Post Image: 10MB
- Resume: 5MB
- AI Output/Other: 100MB

**API 端点**：

**内部 API**（供其他服务调用，需内部认证，通过 Gateway 访问）：
- `POST /api/v1/internal/file/upload` - 上传文件（通过 Gateway：`/api/v1/internal/file/upload`）
- `GET /api/v1/internal/file/:id` - 获取文件信息
- `DELETE /api/v1/internal/file/:id` - 删除文件

**用户 API**（需认证）：
- `POST /api/v1/files/upload` - 上传文件（multipart/form-data）
- `POST /api/v1/files/presigned-upload` - 生成预签名上传 URL（用于前端直传 S3）
- `POST /api/v1/files/confirm-upload` - 确认预签名上传完成
- `GET /api/v1/files` - 获取用户文件列表（支持 `?type=avatar` 过滤）
- `GET /api/v1/files/:id` - 获取文件信息
- `DELETE /api/v1/files/:id` - 删除文件

**预签名 URL 流程**：
1. 前端请求上传凭证 → `POST /api/v1/files/presigned-upload`
2. file-service 生成 PUT pre-signed URL（1 小时有效期）
3. 前端用浏览器直接 PUT 上传到 S3
4. 前端调用 `POST /api/v1/files/confirm-upload` 确认上传
5. file-service 更新文件元数据并返回最终 URL

**优势**：
- 大文件（可达 5GB）不占用 Gateway/Go 服务的带宽
- 上传性能提升
- 成本更低

**端口**：8086

**依赖**：
- PostgreSQL（文件元数据）
- S3/MinIO（文件存储）

**安全配置**：
- `INTERNAL_API_TOKEN`（可选）- 内部 API Token，如果设置则必须验证
- `INTERNAL_ALLOWED_IPS`（可选）- 允许的内部 IP 列表（逗号分隔）

---

#### 5. `cmd/email-worker/main.go` - 邮件工作器
**状态**：✅ 完全实现

**功能**：
- 从 RabbitMQ 消费邮件验证请求
- 通过 SendGrid 发送验证邮件
- 生成验证链接（指向前端验证页面）

**启动流程**：
1. 加载配置
2. 初始化日志
3. 连接 RabbitMQ
4. 初始化邮件服务（SendGrid）
5. 启动消费者循环

**环境变量**：
- `SENDGRID_API_KEY` - SendGrid API 密钥
- `SENDGRID_FROM_EMAIL` - 发件人邮箱
- `FRONTEND_URL` - 前端 URL（用于生成验证链接）

**依赖**：
- RabbitMQ（邮件队列）
- SendGrid（邮件发送服务）

---

#### 6. `cmd/gateway/main.go` - API 网关
**状态**：✅ 完全实现

**功能**：
- 统一 API 入口
- JWT 认证和授权
- 请求路由到后端服务
- CORS 处理
- 请求 ID 追踪
- 可选 Redis 限流
- 内部 API 路由（服务间调用）
- 请求体正确转发（解决 Body 复用问题）
- Hop-by-hop headers 过滤（符合 RFC 2616）

**路由配置**：

**公开路由**（无需认证）：
- `/api/v1/auth/*` → user-service:8081
- `/api/v1/portfolio/*` → portfolio-service:8082
- `/api/v1/users/:id/portfolio` → portfolio-service:8082（**注意**：必须在 `/users/:id` 之前，避免路由冲突）
- `/api/v1/users/:id` → user-service:8081

**受保护路由**（需认证）：
- `/api/v1/users/*` → user-service:8081
- `/api/v1/portfolio/*` → portfolio-service:8082
- `/api/v1/me/notifications/*` → notification-service:8085
- `/api/v1/files/*` → file-service:8086
- `/api/v1/opportunities/*` → opp-service:8083（未实现）
- `/api/v1/community/*` → community-service:8084（未实现）

**内部 API 路由**（服务间调用，需内部认证，所有服务间流量统一经过 Gateway）：
- `/api/v1/internal/user/*` → user-service:8081（用于其他服务调用 user-service）
- `/api/v1/internal/notification/*` → notification-service:8085（用于其他服务调用 notification-service）
- `/api/v1/internal/portfolio/*` → portfolio-service:8082（用于未来其他服务调用 portfolio-service）
- `/api/v1/internal/file/*` → file-service:8086（用于其他服务调用 file-service）

**架构优势**：
- 所有内部流量统一经过 Gateway，便于统一限流、监控、追踪
- 服务之间零耦合，只依赖 Gateway
- 端口变化不影响调用链（只需更新 Gateway 路由）
- 统一的内部 Token/headers 约束

**启动流程**：
1. 加载配置
2. 初始化日志
3. 可选连接 Redis（限流）
4. 初始化 JWT
5. 设置路由和中间件
6. 启动 HTTP 服务器

**技术实现细节**：

**路由顺序优化**：
- 更具体的路由（如 `/users/:id/portfolio`）必须在通用路由（如 `/users/:id`）之前
- Gin 按注册顺序匹配，避免路由冲突和误匹配

**请求体转发修复**：
- 使用 `io.ReadAll` 读取请求体后再转发
- 解决 Gin JSON 绑定后 Body 被读尽（EOF）的问题
- 确保下游服务能正确接收请求体

**Hop-by-hop Headers 过滤**：
- 自动过滤 RFC 2616 规定的 Hop-by-hop headers：
  - `Connection`、`Transfer-Encoding`、`Keep-Alive`
  - `Proxy-Authenticate`、`Proxy-Authorization`
  - `Trailer`、`TE`、`Upgrade`
- 防止代理行为异常

**HTTP Client 优化**：
- 使用全局 `httpClient` 实例（30 秒超时）
- 支持连接池复用，提升性能
- 避免每次请求创建新客户端

**内部 API 路由**：
- 自动为内部 API 请求设置 `X-Internal-Call: true` 头
- 如果配置了 `INTERNAL_API_TOKEN`，自动设置 `X-Internal-Token` 头
- 自动转发用户身份 headers（`X-User-ID`, `X-User-Role`, `X-User-Email`）到下游服务
- 保护内部 API 不被外部直接访问

**端口**：8080

**依赖**：
- Redis（可选，限流）
- JWT Secret（用于验证 Token）

---

### 未实现服务

#### 7. `cmd/opp-service/main.go` - 机会服务
**状态**：❌ 未实现（仅占位符）

**计划功能**：
- 机会（Opportunity）的 CRUD 操作
- 机会匹配算法
- 机会申请管理
- 通过 notification-service 发送机会匹配通知

**目录结构**（已创建但为空）：
- `internal/opportunity/handler/` - HTTP 处理器
- `internal/opportunity/model/` - 数据模型
- `internal/opportunity/repository/` - 数据访问层
- `internal/opportunity/service/` - 业务逻辑层

**计划端口**：8083

**计划依赖**：
- PostgreSQL（机会数据）
- Notification Service（发送通知）

---

#### 8. `cmd/community-service/main.go` - 社区服务
**状态**：❌ 未实现（仅占位符）

**计划功能**：
- 社区帖子/讨论管理
- 评论系统
- 点赞/收藏功能
- 通过 notification-service 发送评论通知

**目录结构**（已创建但为空）：
- `internal/community/handler/` - HTTP 处理器
- `internal/community/model/` - 数据模型
- `internal/community/repository/` - 数据访问层
- `internal/community/service/` - 业务逻辑层

**计划端口**：8084

**计划依赖**：
- PostgreSQL（社区数据）
- Notification Service（发送通知）

---

#### 9. `cmd/ai-worker/main.go` - AI 工作器
**状态**：❌ 未实现（仅占位符）

**计划功能**：
- 从消息队列消费 AI 处理任务
- 生成用户摘要/推荐
- 处理 AI 相关异步任务
- 通过 notification-service 发送 AI 摘要完成通知

**目录结构**（已创建但为空）：
- `internal/aiworker/consumer/` - 消息消费者
- `internal/aiworker/service/` - AI 处理服务

**计划依赖**：
- RabbitMQ（任务队列）
- AI 服务（待定）
- Notification Service（发送通知）

---

## 平台基础层（`internal/platform/`）

### 认证与授权

#### `auth/jwt.go`
- **功能**：JWT Token 生成与解析
- **方法**：
  - `GenerateAccessToken(userID, email, role)` - 生成访问 Token
  - `GenerateRefreshToken(userID, email, role)` - 生成刷新 Token
  - `ParseToken(token)` - 解析 Token
  - `ValidateToken(token)` - 验证 Token
- **配置**：Secret、Access TTL、Refresh TTL

---

### 缓存

#### `cache/redis.go`
- **功能**：Redis 客户端封装
- **用途**：
  - 会话存储（Refresh Token）
  - 邮箱验证 Token
  - 限流计数器
  - 缓存用户数据
- **方法**：
  - `NewRedis(addr, db)` - 创建客户端
  - `Get/Set/Delete` - 基本操作
  - `Incr/Decr` - 计数器操作
  - `Expire` - 设置过期时间

---

### 配置管理

#### `config/config.go`
- **功能**：配置结构体定义
- **配置项**：
  - `App` - 应用配置（环境、端口、名称）
  - `Postgres` - 数据库配置（DSN）
  - `Redis` - 缓存配置（地址、数据库）
  - `MQ` - 消息队列配置（Broker URL）
  - `S3` - 对象存储配置（端点、密钥、Bucket）
  - `JWT` - JWT 配置（Secret、TTL）
  - `Email` - 邮件配置（SendGrid API Key）

#### `config/load.go`
- **功能**：从 YAML 文件和环境变量加载配置
- **优先级**：环境变量 > YAML 文件 > 默认值
- **方法**：
  - `Load(path)` - 加载配置

---

### 数据库

#### `db/postgres.go`
- **功能**：PostgreSQL 连接管理
- **技术栈**：sqlx
- **方法**：
  - `NewPostgres(dsn)` - 创建连接
  - `GetDB()` - 获取 sqlx.DB
  - `GetSQLDB()` - 获取标准 database/sql.DB
  - `Close()` - 关闭连接

#### `db/migrate.go`
- **功能**：数据库迁移
- **技术栈**：golang-migrate
- **方法**：
  - `RunMigrations(db, dir)` - 运行迁移
- **迁移文件位置**：`migrations/`

---

### HTTP 服务器

#### `http/server.go`
- **功能**：通用 HTTP 服务器工具（当前未使用，各服务直接使用 `http.Server`）

---

### 日志

#### `log/logger.go`
- **功能**：结构化日志封装
- **技术栈**：zap
- **方法**：
  - `NewLogger(env)` - 创建日志器
  - `Info/Debug/Warn/Error/Fatal` - 日志级别
- **配置**：根据环境（dev/prod）切换配置

---

### 指标监控

#### `metrics/metrics.go`
- **功能**：Prometheus 指标定义
- **指标**：
  - `registration_success_total` - 注册成功计数
  - `registration_failure_total` - 注册失败计数
  - `login_success_total` - 登录成功计数
  - `login_failure_total` - 登录失败计数
  - `account_locked_total` - 账户锁定计数
  - `avatar_upload_success_total` - 头像上传成功计数
  - `avatar_upload_failure_total` - 头像上传失败计数

---

### 中间件

#### `middleware/recovery.go`
- **功能**：Panic 恢复中间件
- **行为**：捕获 panic，记录日志，返回 500 错误

#### `middleware/request_id.go`
- **功能**：请求 ID 追踪
- **行为**：生成/读取 `X-Request-ID` 头，用于分布式追踪

#### `middleware/cors.go`
- **功能**：CORS 处理
- **配置**：通过环境变量 `CORS_ORIGINS` 配置允许的源

#### `middleware/trust_gateway.go`
- **功能**：信任 Gateway 设置的用户身份信息
- **行为**：
  - `TrustGatewayMiddleware()` - 从请求头读取用户信息（`X-User-ID`, `X-User-Email`, `X-User-Role`）
  - 如果存在，注入到 Gin 上下文
  - 用于微服务中信任 Gateway 传递的用户身份
  - **注意**：不应全局使用，只在需要用户身份的路由组使用
- `RequireGatewayAuth()` - 要求必须存在用户身份头（用于受保护的路由）
  - 检查 `X-User-ID` 是否存在
  - 不存在则返回 401 Unauthorized

#### `middleware/rate_limit.go`
- **功能**：基于 Redis 的限流
- **算法**：滑动窗口（INCR + TTL）
- **限流键**：
  - `login:<email>:<ip>` - 登录限流
  - `register:<fingerprint>:<ip>` - 注册限流
  - `rate:profile_update:<userID>` - Profile 更新限流
  - `rate:avatar:<userID>` - 头像上传限流

#### `middleware/gateway_auth.go`
- **功能**：Gateway JWT 认证中间件
- **行为**：
  1. 验证 JWT Token
  2. 提取用户信息（ID、Email、Role）
  3. 将用户信息写入请求头（`X-User-ID`, `X-User-Email`, `X-User-Role`）
  4. 传递给下游服务

#### `middleware/auth.go`
- **功能**：服务层 JWT 认证中间件（用于 user-service）
- **行为**：验证 Token 并设置用户上下文

#### `middleware/internal_auth.go`
- **功能**：内部 API 认证中间件
- **行为**：
  1. 验证 `X-Internal-Call: true` 头
  2. 可选验证 `X-Internal-Token`（通过环境变量 `INTERNAL_API_TOKEN` 配置）
  3. 可选 IP 白名单验证（通过环境变量 `INTERNAL_ALLOWED_IPS` 配置）
- **用途**：保护内部 API，防止外部直接访问

#### `middleware/rbac.go`
- **功能**：基于角色的访问控制（待实现）

#### `middleware/middleware.go`
- **功能**：中间件工具函数

---

### 消息队列

#### `mq/mq.go`
- **功能**：消息队列接口定义
- **方法**：
  - `PublishEmailVerification(message)` - 发布邮件验证消息

#### `mq/rabbitmq.go`
- **功能**：RabbitMQ 实现
- **技术栈**：amqp091-go
- **方法**：
  - `NewRabbitMQ(broker, logger)` - 创建连接
  - `PublishEmailVerification(message)` - 发布消息
  - `ConsumeEmailVerification(handler)` - 消费消息
  - `Close()` - 关闭连接
- **队列**：
  - `email_verification` - 邮箱验证队列

---

### 对象存储

#### `storage/s3.go`
- **功能**：S3/MinIO 客户端封装
- **技术栈**：minio-go
- **接口**：`ObjectStorage`
  - `Upload(ctx, objectName, reader, size, contentType, metadata)` - 上传对象
  - `Delete(ctx, objectName)` - 删除对象
  - `PresignedPutURL(ctx, objectName, expires)` - 生成预签名 PUT URL（用于前端直传）
  - `BuildPublicURL(objectName)` - 构建公开 URL
- **方法**：
  - `NewS3(config)` - 创建客户端
  - `BaseURL()` - 获取基础 URL
- **用途**：
  - 文件服务（file-service）统一管理所有文件存储
  - 用户头像、项目封面、帖子图片、简历、AI 输出等

---

## 业务域模块（`internal/`）

### User 模块

#### 模型（`model/`）

**`user.go`**：
- `User` - 用户实体
  - 基础信息：ID、Email、PasswordHash、Name、Nickname、AvatarURL
  - 角色：Role（student/company/admin）
  - 学校信息：StudentID、School、Faculty、Major
  - Profile：Skills（JSONB）、Bio、WeeklyHours、EmotionalStatus、WeeklyAvailability（JSONB）
  - 可见性：ProfileVisibility（public/aalto_only/private）
  - 安全：IsVerifiedEmail、OAuthProvider、OAuthSubject、LastLoginAt、FailedAttempts、LockedUntil
  - 时间戳：CreatedAt、UpdatedAt

**`profile.go`**：
- `Project` - Profile 中的项目信息（JSONB）
- `Projects` - Project 数组，带数据库序列化
- `Skill` - 技能（名称 + 等级）
- `Skills` - Skill 数组，带数据库序列化
- `WeeklyAvailability` - 周可用性
- `WeeklyAvailabilityArray` - 周可用性数组，带数据库序列化
- `StringArray` - 字符串数组，带数据库序列化

**`saved_item.go`**：
- `SavedItemType` - 收藏项类型（project/opportunity/user）
- `SavedItem` - 收藏项实体

**`email_message.go`**：
- `EmailVerificationRequest` - 邮件验证请求消息体

#### 仓储（`repository/`）

**`repository.go`**：
- `UserRepository` - 用户数据访问接口
  - `FindByEmail/FindByID` - 查询
  - `CreateUser` - 创建
  - `UpdateProfile` - 更新 Profile
  - `UpdateAvatarURL` - 更新头像
  - `UpdateLoginInfo` - 更新登录信息
  - `SetEmailVerified` - 设置邮箱已验证
  - `SetFailedAttempts/LockAccount` - 安全相关
- `SavedItemRepository` - 收藏项数据访问接口
  - `FindByUserID/FindByUserIDAndType` - 查询
  - `Create/Delete` - 创建/删除
  - `Exists` - 检查是否存在

**`postgres_repository.go`**：
- `postgresRepository` - UserRepository 实现
- `postgresSavedItemRepository` - SavedItemRepository 实现

#### 服务（`service/`）

**`auth_service.go`**：
- `AuthService` - 认证服务
  - `Register` - 注册（密码校验、角色归一化、自动验证 @aalto.fi 邮箱、生成 Token、存储 Refresh Token 到 Redis）
  - `Login` - 登录（限流、锁定检查、密码验证、JWT 签发）
  - `RefreshToken` - 刷新 Token（验证 Refresh Token、生成新 Token）
  - `Logout` - 登出（删除 Refresh Token）
  - `VerifyEmail` - 邮箱验证（调用 EmailVerificationService）

**`profile_service.go`**：
- `ProfileService` - Profile 服务
  - `GetProfile` - 获取 Profile（带缓存）
  - `UpdateProfile` - 更新 Profile（字段归一化、限流、Prometheus 计数）
  - `UploadAvatar` - 上传头像（MIME 校验、大小限制、S3 上传、限流）
  - `SearchBySkills` - 按技能搜索（前缀匹配）

**`email_verification.go`**：
- `EmailVerificationService` - 邮箱验证服务
  - `GenerateToken` - 生成验证 Token（存储到 Redis，TTL 30 分钟）
  - `VerifyToken` - 验证 Token（从 Redis 读取、一次性使用）
  - `MarkEmailAsVerified` - 标记邮箱已验证

**`saved_item_service.go`**：
- `SavedItemService` - 收藏项服务
  - `GetSavedItems/GetSavedItemsByType` - 获取收藏项
  - `SaveItem/UnsaveItem` - 保存/取消保存
  - `IsSaved` - 检查是否已保存

**`notification_client.go`**：
- `NotificationClient` - 通知客户端接口
- `httpNotificationClient` - HTTP 实现
  - `CreateNotification` - 通过 Gateway 的内部路由调用 notification-service 创建通知
  - 调用路径：`/api/v1/internal/notification/notifications`（通过 Gateway）
  - 使用 `GATEWAY_URL` 环境变量（默认 `http://gateway:8080`）
  - Gateway 自动设置内部调用头
- `NotifyProfileSaved` - 辅助函数，发送"资料被保存"通知

**`service.go`**：
- 服务接口别名定义

#### 处理器（`handler/`）

**`handler.go`**：
- `AuthHandler` - HTTP 请求处理器
  - 注册/登录/刷新/登出处理器
  - Profile 管理处理器
  - 头像上传处理器
  - 可用性管理处理器
  - 收藏项管理处理器
  - 错误处理和响应格式化

**`dto.go`**：
- `RegisterRequest` - 注册请求 DTO
- `LoginRequest` - 登录请求 DTO
- `UpdateProfileRequest` - 更新 Profile 请求 DTO
- `ProjectInput` - Profile 项目输入（已废弃，项目移至 portfolio-service）

---

### Portfolio 模块

#### 模型（`model/`）

**`project.go`**：
- `Project` - 项目实体
  - ID、UserID、Title、Description、Year
  - Tags（StringArray，JSONB）
  - CoverImageURL、ProjectLink
  - CreatedAt、UpdatedAt
- `StringArray` - 字符串数组，带数据库序列化

#### 仓储（`repository/`）

**`repository.go`**：
- `ProjectRepository` - 项目数据访问接口
  - `FindByUserID/FindByID` - 查询
  - `Create/Update/Delete` - CRUD

**`postgres_repository.go`**：
- `postgresProjectRepository` - ProjectRepository 实现

#### 服务（`service/`）

**`service.go`**：
- `ProjectService` - 项目服务
  - `GetUserProjects` - 获取用户项目列表
  - `GetProject` - 获取项目详情
  - `CreateProject` - 创建项目（验证标题）
  - `UpdateProject` - 更新项目（验证所有权）
  - `DeleteProject` - 删除项目（验证所有权）

**`user_client.go`**：
- `UserServiceClient` - 用户服务客户端接口
  - `CheckProfileVisibility` - 检查用户 Profile 可见性
- `HTTPUserServiceClient` - HTTP 实现
  - 通过 Gateway 的内部路由调用 user-service 检查可见性
  - 调用路径：`/api/v1/internal/user/users/:id`（通过 Gateway）
  - `NewHTTPUserServiceClient(baseURL)` - 创建客户端（需传入 baseURL，默认从 `GATEWAY_URL` 环境变量读取，默认值 `http://gateway:8080`）

#### 处理器（`handler/`）

**`handler.go`**：
- `PortfolioHandler` - HTTP 请求处理器
  - `GetProjectDetailHandler` - 获取项目详情（公开）
  - `GetUserPortfolioHandler` - 获取用户作品集（公开，受可见性控制）
  - `GetMyPortfolioHandler` - 获取当前用户作品集（需认证）
  - `CreateProjectHandler` - 创建项目（需认证）
  - `UpdateProjectHandler` - 更新项目（需认证）
  - `DeleteProjectHandler` - 删除项目（需认证）

---

### Notification 模块

#### 模型（`model/`）

**`notification.go`**：
- `NotificationType` - 通知类型枚举
  - `profile_saved` - 用户资料被保存
  - `follow` - 关注某人
  - `project_saved` - 项目被保存
  - `comment` - 评论
  - `opportunity_match` - 机会匹配
  - `project_invite` - 项目邀请
  - `ai_summary_finished` - AI 摘要完成
  - `system` - 系统通知
  - `weekly_digest` - 每周摘要
  - `message` - 消息
- `NotificationData` - 通知数据（JSONB，map[string]interface{}）
- `Notification` - 通知实体
  - ID、UserID、Type、Title、Message、Data、IsRead、CreatedAt

#### 仓储（`repository/`）

**`repository.go`**：
- `NotificationRepository` - 通知数据访问接口
  - `Create` - 创建通知
  - `FindByUserID/FindUnreadByUserID` - 查询
  - `MarkAsRead/MarkAllAsRead` - 标记已读
  - `CountUnread` - 统计未读
  - `Delete` - 删除

**`postgres_repository.go`**：
- `postgresNotificationRepository` - NotificationRepository 实现

#### 服务（`service/`）

**`service.go`**：
- `NotificationService` - 通知服务
  - `CreateNotification` - 创建通知
  - `GetNotifications/GetUnreadNotifications` - 获取通知列表（带分页和限制）
  - `MarkAsRead/MarkAllAsRead` - 标记已读
  - `GetUnreadCount` - 获取未读数量
  - `DeleteNotification` - 删除通知

#### 处理器（`handler/`）

**`handler.go`**：
- `NotificationHandler` - HTTP 请求处理器
  - `GetNotificationsHandler` - 获取通知列表（支持 unread 过滤，路径：`/api/v1/me/notifications`）
  - `GetUnreadCountHandler` - 获取未读数量（路径：`/api/v1/me/notifications/unread-count`）
  - `MarkNotificationAsReadHandler` - 标记为已读（路径：`/api/v1/me/notifications/:id/read`）
  - `MarkAllNotificationsAsReadHandler` - 标记全部已读（路径：`/api/v1/me/notifications/read-all`）
  - `DeleteNotificationHandler` - 删除通知（路径：`/api/v1/me/notifications/:id`）
  - `CreateNotificationHandler` - 创建通知（内部 API，路径：`/api/v1/internal/notifications`）

**`dto.go`**：
- `CreateNotificationRequest` - 创建通知请求 DTO

---

### File 模块

#### 模型（`model/`）

**`file.go`**：
- `FileType` - 文件类型枚举
  - `FileTypeAvatar` - 用户头像
  - `FileTypeProjectCover` - 项目封面图
  - `FileTypePostImage` - 社区帖子图片
  - `FileTypeResume` - 简历 PDF
  - `FileTypeAIOutput` - AI 输出文件
  - `FileTypeOther` - 其他文件
- `File` - 文件实体
  - ID、UserID、Type、ObjectKey、URL、Filename
  - ContentType、Size、Metadata（TEXT，JSON 字符串）
  - CreatedAt、UpdatedAt

#### 仓储（`repository/`）

**`repository.go`**：
- `FileRepository` - 文件数据访问接口
  - `Create` - 创建文件记录
  - `FindByID/FindByUserID/FindByUserIDAndType` - 查询
  - `FindByObjectKey` - 通过 object key 查找（用于预签名上传确认）
  - `Update` - 更新文件记录
  - `Delete/DeleteByUserID` - 删除

**`postgres_repository.go`**：
- `postgresFileRepository` - FileRepository 实现

#### 服务（`service/`）

**`service.go`**：
- `FileService` - 文件服务
  - `UploadFile` - 上传文件（验证大小、MIME 类型，上传到 S3，创建数据库记录）
  - `GetFile/GetUserFiles` - 获取文件信息
  - `DeleteFile` - 删除文件（先删除 S3 对象，再删除数据库记录，保持幂等性）
  - `GeneratePresignedUploadURL` - 生成预签名上传 URL（用于前端直传 S3）
  - `ConfirmUpload` - 确认预签名上传完成（更新文件元数据）
- `PresignedUploadResponse` - 预签名上传响应结构
  - UploadURL、ObjectKey、FileID、ExpiresIn、PublicURL

**文件验证**：
- `validateFileSize` - 基于文件类型的大小限制验证
- `validateContentType` - Content-Type 白名单验证
- `isValidFileType` - 文件类型验证
- `getAllowedContentTypes` - 获取允许的 Content-Type 列表

#### 处理器（`handler/`）

**`handler.go`**：
- `FileHandler` - HTTP 请求处理器
  - `UploadFileHandler` - 上传文件（支持 multipart/form-data，使用 magic number 验证）
  - `GetFileHandler` - 获取文件信息
  - `GetUserFilesHandler` - 获取用户文件列表（支持 `?type=avatar` 过滤）
  - `DeleteFileHandler` - 删除文件
  - `GeneratePresignedUploadURLHandler` - 生成预签名上传 URL
  - `ConfirmUploadHandler` - 确认预签名上传完成

**安全特性**：
- 使用 `io.LimitReader` 防止内存耗尽
- Magic number 验证（前 512 字节）防止 content-type 伪造
- 基于文件类型的大小限制
- Content-Type 白名单验证

---

### Opportunity 模块（未实现）

#### 目录结构
- `handler/` - HTTP 处理器（空）
- `model/` - 数据模型（空）
- `repository/` - 数据访问层（空）
- `service/` - 业务逻辑层（空）

#### 计划功能
- 机会（Opportunity）的 CRUD 操作
- 机会匹配算法
- 机会申请管理
- 通过 notification-service 发送机会匹配通知

---

### Community 模块（未实现）

#### 目录结构
- `handler/` - HTTP 处理器（空）
- `model/` - 数据模型（空）
- `repository/` - 数据访问层（空）
- `service/` - 业务逻辑层（空）

#### 计划功能
- 社区帖子/讨论管理
- 评论系统
- 点赞/收藏功能
- 通过 notification-service 发送评论通知

---

### AI Worker 模块（未实现）

#### 目录结构
- `consumer/` - 消息消费者（空）
- `service/` - AI 处理服务（空）

#### 计划功能
- 从消息队列消费 AI 处理任务
- 生成用户摘要/推荐
- 处理 AI 相关异步任务
- 通过 notification-service 发送 AI 摘要完成通知

---

## 共享工具（`pkg/`）

### 错误处理

#### `errs/errors.go`
- **功能**：统一错误处理
- **类型**：
  - `AppError` - 应用错误（包含错误码、HTTP 状态码、消息、详情）
  - 错误代码常量（`CodeUserNotFound`, `CodeInvalidCredentials`, 等）
  - 错误变量（`ErrUserNotFound`, `ErrInvalidCredentials`, 等）
- **方法**：
  - `NewAppError` - 创建应用错误
  - `ToHTTPStatus` - 错误转 HTTP 状态码
  - `GetErrorCode` - 提取错误代码
  - `Is` - 检查错误代码

---

### 响应格式化

#### `response/response.go`
- **功能**：统一 API 响应格式
- **结构**：
  ```json
  {
    "success": true/false,
    "data": {...},
    "error": {
      "code": "...",
      "message": "...",
      "details": {...}
    }
  }
  ```
- **方法**：
  - `Success(data)` - 成功响应
  - `Error(err)` - 错误响应
  - `ErrorWithCode(err, code, message)` - 带自定义代码的错误响应

---

### 类型和工具

#### `types/types.go`
- **功能**：通用类型定义（当前使用较少）

#### `utils/utils.go`
- **功能**：通用工具函数（当前使用较少）

---

## 配置与部署

### 配置文件

#### `configs/config.example.yaml`
- **结构**：
  - `app` - 应用配置（环境、端口、名称）
  - `postgres` - 数据库配置（DSN）
  - `redis` - 缓存配置（地址、数据库）
  - `mq` - 消息队列配置（Broker URL）
  - `s3` - 对象存储配置（端点、密钥、Bucket、公开 URL）
  - `jwt` - JWT 配置（Secret、Access TTL、Refresh TTL）
  - `email` - 邮件配置（SendGrid API Key、发件人邮箱、前端 URL）

---

### Docker 部署

#### `docker-compose.yml`
- **服务**：
  - `postgres` - PostgreSQL 15（端口 5432）
  - `redis` - Redis 7（端口 6379）
  - `backend-mq` - RabbitMQ 3（端口 5672，管理 UI 15672）
  - `minio` - MinIO（端口 9000，控制台 9001）
  - `user-service` - 用户服务（端口 8081）
  - `email-worker` - 邮件工作器
  - `portfolio-service` - 作品集服务（端口 8082）
  - `notification-service` - 通知服务（端口 8085）
  - `file-service` - 文件服务（端口 8086）
  - `gateway` - API 网关（端口 8080）

- **依赖关系**：
  - user-service 依赖：postgres、redis、backend-mq
  - email-worker 依赖：backend-mq
  - portfolio-service 依赖：postgres
  - notification-service 依赖：postgres
  - file-service 依赖：postgres、minio
  - gateway 依赖：postgres、redis、backend-mq、user-service、portfolio-service、notification-service、file-service

- **环境变量配置**：
  - `INTERNAL_API_TOKEN` - 内部 API Token（可选，用于保护内部 API）
  - `INTERNAL_ALLOWED_IPS` - 内部 IP 白名单（可选，逗号分隔）
  - `NOTIFICATION_SERVICE_URL` - Notification Service 地址（user-service 使用，默认 `http://notification-service:8085`）

#### `deploy/docker/`
- **Dockerfile**：
  - `Dockerfile.user-service` - 用户服务镜像
  - `Dockerfile.portfolio-service` - 作品集服务镜像
  - `Dockerfile.notification-service` - 通知服务镜像
  - `Dockerfile.file-service` - 文件服务镜像
  - `Dockerfile.email-worker` - 邮件工作器镜像
  - `Dockerfile.gateway` - 网关镜像

---

### 数据库迁移

#### `migrations/001_initial_schema.sql`
- **表结构**：
  - `users` - 用户表
    - 基础字段：id、email、password_hash、name、nickname、avatar_url
    - 角色：role（student/company/admin）
    - 学校信息：student_id（VARCHAR(255)）、school、faculty、major
    - Profile：skills（JSONB）、bio、weekly_hours、emotional_status、weekly_availability（JSONB）
    - 可见性：profile_visibility（public/aalto_only/private）
    - 安全：is_verified_email、oauth_provider、oauth_subject、last_login_at、failed_attempts、locked_until
    - 时间戳：created_at、updated_at
  - `projects` - 项目表
    - id、user_id、title、description、year
    - tags（JSONB）、cover_image_url、project_link
    - created_at、updated_at
  - `saved_items` - 收藏项表
    - id、user_id、item_id、item_type（project/opportunity/user）
    - created_at
  - `notifications` - 通知表
    - id、user_id、type、title、message、data（JSONB）
    - is_read、created_at
  - `files` - 文件表（file-service）
    - id、user_id、type（avatar/project_cover/post_image/resume/ai_output/other）
    - object_key、url、filename、content_type、size
    - metadata（TEXT，JSON 字符串）、created_at、updated_at

- **索引**：
  - users：
    - `idx_users_email_ci` - email 大小写不敏感唯一索引（`LOWER(email)`）
    - role、profile_visibility、school、faculty、major
    - OAuth 唯一索引（partial index，仅当 oauth_provider 和 oauth_subject 不为 NULL 时）
  - projects：user_id、created_at
  - saved_items：user_id、item（item_id + item_type）、created_at
  - notifications：user_id、is_read、created_at、type
  - files：user_id、type、user_id+type 组合索引、created_at

- **重要设计决策**：
  - **Email 大小写不敏感**：使用 `CREATE UNIQUE INDEX idx_users_email_ci ON users (LOWER(email))` 确保 "AAA@example.com" 和 "aaa@example.com" 被视为同一个邮箱
  - **代码层面规范化**：在 Register 和 Login 时自动将 email 转换为小写存储
  - **OAuth 唯一性**：使用 partial unique index 确保同一 OAuth provider+subject 组合的唯一性，同时允许 NULL 值

---

## 数据流与依赖关系

### 注册流程
1. 客户端 → Gateway (`POST /api/v1/auth/register`)
2. Gateway → user-service（代理请求）
3. user-service：
   - 验证输入
   - 密码哈希
   - 写入 `users` 表
   - 生成 JWT/Refresh Token
   - 存储 Refresh Token 到 Redis（键：`refresh_token:<token>`，TTL 30 天）
   - 生成邮箱验证 Token（存储到 Redis，键：`email_verify_token:<uuid>`，TTL 30 分钟）
   - 发布邮件消息到 RabbitMQ（`EmailVerificationRequest`）
4. email-worker：
   - 消费 RabbitMQ 消息
   - 调用 SendGrid 发送验证邮件
   - 邮件包含链接：`FRONTEND_URL/verify?token=<uuid>`

### 邮箱验证流程
1. 用户点击邮件链接 → 前端 (`GET /api/v1/auth/verify?token=`)
2. Gateway → user-service（代理请求）
3. user-service：
   - `EmailVerificationService.VerifyToken`（从 Redis 读取 Token）
   - 验证用户/邮箱一致性
   - 删除 Redis Token（一次性使用）
   - `MarkEmailAsVerified`（更新数据库 `is_verified_email`）
4. 返回成功响应

### 登录流程
1. 客户端 → Gateway (`POST /api/v1/auth/login`)
2. Gateway → user-service（代理请求）
3. user-service：
   - Redis 限流检查（键：`login:<email>:<ip>`）
   - 检查账户锁定状态
   - 验证密码
   - 检查邮箱验证状态
   - 生成 JWT/Refresh Token
   - 存储 Refresh Token 到 Redis
   - 更新 `last_login_at`、重置 `failed_attempts`
4. 返回 Token

### 通知创建流程（示例：资料被保存）
1. user-service 处理保存操作
2. 检测到保存的是用户资料（`SavedItemTypeUser`）
3. 调用 `NotificationClient.CreateNotification`（HTTP 请求到 Gateway）
   - 请求路径：`POST /api/v1/internal/notification/notifications`
   - 目标：Gateway（`GATEWAY_URL`）
4. Gateway：
   - 接收内部 API 请求
   - 自动设置 `X-Internal-Call: true` 头
   - 如果配置了 `INTERNAL_API_TOKEN`，也会设置 `X-Internal-Token` 头
   - 代理请求到 notification-service
5. notification-service：
   - 接收请求（`POST /api/v1/internal/notifications`）
   - `RequireInternalCall` 中间件验证内部调用头
   - 创建通知记录到 `notifications` 表
6. 用户可通过 `GET /api/v1/me/notifications` 获取通知

**优势**：所有内部流量经过 Gateway，便于统一监控、限流和追踪

### 文件上传流程（示例：头像上传）
1. 客户端 → Gateway (`POST /api/v1/users/me/avatar`)
2. Gateway → user-service（代理请求，携带用户身份）
3. user-service：
   - 验证用户身份
   - 调用 `FileClient.UploadFile`（HTTP 请求到 Gateway）
   - 请求路径：`POST /api/v1/internal/file/upload`
   - 目标：Gateway（`GATEWAY_URL`）
4. Gateway：
   - 接收内部 API 请求
   - 自动设置 `X-Internal-Call: true` 头
   - 转发用户身份 headers（`X-User-ID`）
   - 代理请求到 file-service
5. file-service：
   - 接收请求（`POST /api/v1/internal/file/upload`）
   - `RequireInternalCall` 中间件验证内部调用头
   - `TrustGatewayMiddleware` 提取用户身份
   - 验证文件大小和 MIME 类型
   - 上传到 S3/MinIO
   - 创建文件记录到 `files` 表
   - 返回文件信息（包括 URL）
6. Gateway → user-service（返回文件信息）
7. user-service：
   - 更新用户的 `avatar_url` 字段
   - 返回更新后的用户信息

**优势**：
- 文件服务独立，不占用 user-service 资源
- 统一文件管理，便于扩展（项目封面、帖子图片等）
- 支持预签名 URL，前端可直接上传大文件到 S3

### 文件删除流程
1. 客户端 → Gateway (`DELETE /api/v1/files/:id`)
2. Gateway → file-service（代理请求，携带用户身份）
3. file-service：
   - 验证文件所有权
   - **先删除 S3 对象**（使用 `object_key`）
   - **再删除数据库记录**（保持幂等性）
   - 如果 S3 删除失败，记录警告但继续删除数据库记录

**优势**：确保 S3 对象和数据库记录的一致性，防止存储泄漏

### 作品集可见性检查流程
1. 客户端 → Gateway (`GET /api/v1/users/:id/portfolio`)
2. Gateway → portfolio-service（代理请求）
3. portfolio-service：
   - 调用 `UserServiceClient.CheckProfileVisibility`（HTTP 请求到 Gateway 内部路由）
   - 请求路径：`GET /api/v1/internal/user/users/:id`
4. Gateway：
   - 接收内部 API 请求
   - 自动设置内部调用头
   - 代理请求到 user-service
5. user-service：
   - 查询用户 Profile 可见性设置
   - 根据可见性规则返回结果（public/aalto_only/private）
6. Gateway → portfolio-service（返回结果）
7. portfolio-service：
   - 根据可见性结果决定是否返回作品集

**优势**：所有服务间调用统一经过 Gateway，便于统一监控和追踪

### 限流机制
- **登录限流**：`login:<email>:<ip>` - 防止暴力破解
- **注册限流**：`register:<fingerprint>:<ip>` - 防止批量注册
- **Profile 更新限流**：`rate:profile_update:<userID>` - 防止频繁更新
- **头像上传限流**：`rate:avatar:<userID>` - 防止频繁上传

### 监控指标
- Prometheus 计数器：
  - `registration_success_total` / `registration_failure_total`
  - `login_success_total` / `login_failure_total`
  - `account_locked_total`
  - `avatar_upload_success_total` / `avatar_upload_failure_total`

---

## 服务端口分配

| 服务 | 端口 | 状态 |
|------|------|------|
| Gateway | 8080 | ✅ 已实现 |
| user-service | 8081 | ✅ 已实现 |
| portfolio-service | 8082 | ✅ 已实现 |
| opp-service | 8083 | ❌ 未实现 |
| community-service | 8084 | ❌ 未实现 |
| notification-service | 8085 | ✅ 已实现 |
| file-service | 8086 | ✅ 已实现 |

---

## 技术栈

- **语言**：Go 1.21+
- **Web 框架**：Gin
- **数据库**：PostgreSQL 15（sqlx）
- **缓存**：Redis 7（go-redis v9）
- **消息队列**：RabbitMQ 3（amqp091-go）
- **对象存储**：MinIO/S3（minio-go）
- **日志**：zap
- **JWT**：自定义实现（基于 jwt-go）
- **邮件**：SendGrid
- **数据库迁移**：golang-migrate
- **容器化**：Docker + Docker Compose

---

## 未来扩展计划

1. **Opportunity Service**：
   - 实现机会 CRUD
   - 机会匹配算法
   - 申请管理

2. **Community Service**：
   - 帖子/讨论系统
   - 评论功能
   - 点赞/收藏

3. **AI Worker**：
   - AI 摘要生成
   - 推荐算法
   - 异步 AI 任务处理

4. **增强功能**：
   - WebSocket 支持（实时通知）
   - 搜索服务（Elasticsearch）
   - 分析服务（用户行为分析）
   - 文件病毒扫描（集成 ClamAV）
   - 文件 CDN 集成（CloudFront/Cloudflare）

---

## 开发指南

### 添加新服务

1. 在 `cmd/` 下创建服务入口
2. 在 `internal/` 下创建业务模块（handler/model/repository/service）
3. 在 Gateway 中添加路由配置（注意路由顺序，更具体的路由在前）
4. 在 `docker-compose.yml` 中添加服务定义
5. 创建 Dockerfile（如需要）
6. 设置统一的 Health Check 路径（`/<service-name>/health`）

### 添加内部 API

1. 在服务中创建 `/api/v1/internal/<resource>` 路由组
2. 使用 `RequireInternalCall()` 中间件保护
3. 在 Gateway 中添加内部路由，自动设置 `X-Internal-Call` 头
4. 配置 `INTERNAL_API_TOKEN` 环境变量（可选但推荐）

### 添加新通知类型

1. 在 `internal/notification/model/notification.go` 中添加新的 `NotificationType` 常量
2. 在相关服务中调用 `NotificationClient.CreateNotification` 时使用新类型

### 数据库迁移

1. 在 `migrations/` 下创建新的 SQL 文件（按序号命名）
2. 服务启动时会自动运行迁移

---

## 重要改进与最佳实践

### Gateway 改进（已实现）

1. **路由顺序优化**
   - 更具体的路由必须在通用路由之前
   - 例如：`/users/:id/portfolio` 必须在 `/users/:id` 之前
   - 避免路由冲突和误匹配

2. **请求体转发修复**
   - 使用 `io.ReadAll` 读取请求体后再转发
   - 解决 Gin JSON 绑定后 Body 被读尽的问题
   - 确保下游服务能正确接收请求体

3. **Hop-by-hop Headers 过滤**
   - 自动过滤 RFC 2616 规定的 Hop-by-hop headers
   - 防止代理行为异常

4. **HTTP Client 复用**
   - 使用全局 `httpClient` 实例
   - 支持连接池复用，提升性能

5. **CORS 配置修复**
   - 修复了环境变量为空时的逻辑错误
   - 正确检查 `CORS_ORIGINS` 环境变量

### Notification Service 改进（已实现）

1. **路由结构优化**
   - 内部 API：`/api/v1/internal/notifications`（服务间调用）
   - 用户 API：`/api/v1/me/notifications`（用户访问）
   - 避免路由冲突，API 更清晰

2. **安全增强**
   - 内部 API 使用 `RequireInternalCall` 中间件
   - 支持可选的 Token 验证（`INTERNAL_API_TOKEN`）
   - 支持可选的 IP 白名单（`INTERNAL_ALLOWED_IPS`）

3. **中间件使用优化**
   - 移除了全局 `TrustGatewayMiddleware`
   - 只在用户路由组使用，内部 API 使用独立验证

### 服务间调用统一化（已实现）

1. **所有服务间调用通过 Gateway**
   - user-service → notification-service：通过 Gateway `/api/v1/internal/notification/*`
   - portfolio-service → user-service：通过 Gateway `/api/v1/internal/user/*`
   - 使用 `GATEWAY_URL` 环境变量（默认 `http://gateway:8080`）

2. **架构优势**
   - 所有内部流量统一经过 Gateway，便于统一限流、监控、追踪
   - 服务之间零耦合，只依赖 Gateway
   - 端口变化不影响调用链（只需更新 Gateway 路由）
   - 统一的内部 Token/headers 约束

3. **Gateway 内部路由**
   - `/api/v1/internal/user/*` → user-service:8081
   - `/api/v1/internal/notification/*` → notification-service:8085
   - `/api/v1/internal/portfolio/*` → portfolio-service:8082

### Health Check 路径统一（已实现）

所有服务使用统一的 Health Check 路径格式：
- Gateway: `/gateway/health`
- user-service: `/user/health`
- portfolio-service: `/portfolio/health`
- notification-service: `/notification/health`

便于容器编排和监控系统配置。

### API 路径规范

- **用户 API**：`/api/v1/me/*` - 当前用户相关操作
- **内部 API**：`/api/v1/internal/*` - 服务间调用
- **公开 API**：`/api/v1/*` - 公开访问的资源

### 安全最佳实践

1. **内部 API 保护**
   - 使用 `RequireInternalCall` 中间件
   - 配置 `INTERNAL_API_TOKEN` 环境变量
   - 可选配置 IP 白名单

2. **Gateway 认证**
   - 所有受保护路由通过 Gateway 认证
   - Gateway 验证 JWT 后设置用户身份头
   - 下游服务信任 Gateway 设置的头

3. **CORS 配置**
   - 生产环境应明确指定允许的源
   - 开发环境可使用 `*`（允许所有）

---