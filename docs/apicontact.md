# Aalto Talent Network API 文档

## 概述

本文档描述了 Aalto Talent Network 后端 API 的完整接口规范。所有 API 请求都通过 Gateway (端口 8080) 进行路由，Gateway 负责认证、授权和请求转发。

**Base URL**: `http://localhost:8080/api/v1`

**API 版本**: v1

**认证方式**: JWT Bearer Token (通过 Gateway 认证中间件)

---

## 目录

1. [认证与授权 (Authentication)](#认证与授权-authentication)
2. [用户管理 (User Management)](#用户管理-user-management)
3. [作品集 (Portfolio)](#作品集-portfolio)
4. [社区 (Community)](#社区-community)
5. [活动 (Events)](#活动-events)
6. [机会 (Opportunities)](#机会-opportunities)
7. [通知 (Notifications)](#通知-notifications)
8. [文件管理 (Files)](#文件管理-files)
9. [内部 API (Internal APIs)](#内部-api-internal-apis)

---

## 认证与授权 (Authentication)

所有认证相关的 API 都在 `/api/v1/auth` 路径下。

### 注册 (Register)

**POST** `/api/v1/auth/register`

创建新用户账户。

**请求体**:
```json
{
  "username": "string (必填, 3-30字符, 字母数字下划线)",
  "email": "string (必填, 有效邮箱地址)",
  "password": "string (必填, 最少8字符)",
  "first_name": "string (可选)",
  "last_name": "string (可选)"
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "email_verified": false
    },
    "access_token": "string (JWT access token)",
    "refresh_token": "string (JWT refresh token)"
  }
}
```

**错误响应**:
- `400 Bad Request`: 输入验证失败
- `409 Conflict`: 用户名或邮箱已存在

---

### 登录 (Login)

**POST** `/api/v1/auth/login`

用户登录，获取访问令牌。

**请求体**:
```json
{
  "email": "string (必填)",
  "password": "string (必填)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe"
    },
    "access_token": "string (JWT access token)",
    "refresh_token": "string (JWT refresh token)"
  }
}
```

**错误响应**:
- `401 Unauthorized`: 邮箱或密码错误

---

### 刷新令牌 (Refresh Token)

**POST** `/api/v1/auth/refresh`

使用 refresh token 获取新的 access token。

**请求体**:
```json
{
  "refresh_token": "string (必填)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "access_token": "string (新的 JWT access token)",
    "refresh_token": "string (新的 JWT refresh token)"
  }
}
```

---

### 登出 (Logout)

**POST** `/api/v1/auth/logout`

用户登出，使 refresh token 失效。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

### 邮箱验证 - 发送验证邮件

**POST** `/api/v1/auth/verify-email`

发送邮箱验证邮件。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Verification email sent"
}
```

---

### 邮箱验证 - 验证链接

**GET** `/api/v1/auth/verify?token=<verification_token>`

通过验证链接验证邮箱。

**查询参数**:
- `token` (必填): 邮箱验证令牌

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Email verified successfully"
}
```

---

### 忘记密码

**POST** `/api/v1/auth/forgot-password`

发送密码重置邮件。

**请求体**:
```json
{
  "email": "string (必填)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Password reset email sent"
}
```

---

### 重置密码

**POST** `/api/v1/auth/reset-password`

使用重置令牌设置新密码。

**请求体**:
```json
{
  "token": "string (必填, 密码重置令牌)",
  "new_password": "string (必填, 最少8字符)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Password reset successfully"
}
```

---

## 用户管理 (User Management)

### 检查用户名可用性

**GET** `/api/v1/users/check-username?username=<username>`

检查用户名是否可用（注册时使用）。

**查询参数**:
- `username` (必填): 要检查的用户名

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "available": true
  }
}
```

---

### 检查邮箱可用性

**GET** `/api/v1/users/check-email?email=<email>`

检查邮箱是否可用（注册时使用）。

**查询参数**:
- `email` (必填): 要检查的邮箱

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "available": true
  }
}
```

---

### 获取用户信息（公开）

**GET** `/api/v1/users/:id`

获取指定用户的公开信息。

**路径参数**:
- `id` (必填): 用户 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "john_doe",
    "first_name": "John",
    "last_name": "Doe",
    "avatar_url": "string (可选)",
    "bio": "string (可选)",
    "profile_visibility": "public|aalto_only|private",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
- `403 Forbidden`: 用户资料不可见（根据 profile_visibility 设置）
- `404 Not Found`: 用户不存在

---

### 获取用户摘要（公开）

**GET** `/api/v1/users/:id/summary`

获取用户的轻量级摘要信息（用于卡片/列表显示）。

**路径参数**:
- `id` (必填): 用户 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "john_doe",
    "first_name": "John",
    "last_name": "Doe",
    "avatar_url": "string (可选)"
  }
}
```

---

### 获取当前用户信息

**GET** `/api/v1/users/me`

获取当前登录用户的完整信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "avatar_url": "string (可选)",
    "bio": "string (可选)",
    "profile_visibility": "public|aalto_only|private",
    "email_verified": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新当前用户信息

**PATCH** `/api/v1/users/me`

更新当前登录用户的信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "first_name": "string (可选)",
  "last_name": "string (可选)",
  "bio": "string (可选)",
  "profile_visibility": "public|aalto_only|private (可选)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "bio": "Updated bio",
    "profile_visibility": "public",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 上传头像

**POST** `/api/v1/users/me/avatar`

上传用户头像。

**请求头**:
```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**请求体** (multipart/form-data):
- `avatar` (必填): 图片文件 (支持 jpg, png, gif, 最大 5MB)

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "avatar_url": "https://example.com/files/avatar_123.jpg"
  }
}
```

---

### 修改密码

**PATCH** `/api/v1/users/me/password`

修改当前用户的密码。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "current_password": "string (必填)",
  "new_password": "string (必填, 最少8字符)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Password updated successfully"
}
```

**错误响应**:
- `400 Bad Request`: 当前密码错误

---

### 获取可用性设置

**GET** `/api/v1/users/me/availability`

获取当前用户的可用性设置。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "available": true,
    "available_from": "2024-06-01",
    "available_until": "2024-08-31",
    "location": "Helsinki, Finland",
    "remote": true
  }
}
```

---

### 更新可用性设置

**PATCH** `/api/v1/users/me/availability`

更新当前用户的可用性设置。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "available": true,
  "available_from": "2024-06-01 (可选, ISO 8601 日期)",
  "available_until": "2024-08-31 (可选, ISO 8601 日期)",
  "location": "string (可选)",
  "remote": true
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "available": true,
    "available_from": "2024-06-01",
    "available_until": "2024-08-31",
    "location": "Helsinki, Finland",
    "remote": true
  }
}
```

---

### 获取收藏项

**GET** `/api/v1/users/me/saved`

获取当前用户的所有收藏项。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `type` (可选): 过滤类型 (`opportunity`, `event`, `post`)
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "type": "opportunity",
        "item_id": 123,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

### 添加收藏项

**POST** `/api/v1/users/me/saved`

添加一个收藏项。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "type": "opportunity|event|post (必填)",
  "item_id": 123
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "type": "opportunity",
    "item_id": 123,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 删除收藏项

**DELETE** `/api/v1/users/me/saved`

删除一个收藏项。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "type": "opportunity|event|post (必填)",
  "item_id": 123
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Item unsaved successfully"
}
```

---

## 作品集 (Portfolio)

### 获取项目详情（公开）

**GET** `/api/v1/portfolio/:id`

获取指定项目的详细信息。

**路径参数**:
- `id` (必填): 项目 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Project Title",
    "description": "Project description",
    "image_url": "string (可选)",
    "url": "string (可选)",
    "technologies": ["React", "Node.js"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 获取用户作品集（公开）

**GET** `/api/v1/users/:id/portfolio`

获取指定用户的所有项目。

**路径参数**:
- `id` (必填): 用户 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "projects": [
      {
        "id": 1,
        "title": "Project Title",
        "description": "Project description",
        "image_url": "string (可选)",
        "url": "string (可选)",
        "technologies": ["React", "Node.js"],
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### 获取我的作品集

**GET** `/api/v1/users/me/portfolio`

获取当前用户的所有项目。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "projects": [
      {
        "id": 1,
        "title": "Project Title",
        "description": "Project description",
        "image_url": "string (可选)",
        "url": "string (可选)",
        "technologies": ["React", "Node.js"],
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### 创建项目

**POST** `/api/v1/users/me/portfolio`

创建新项目。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "title": "string (必填)",
  "description": "string (可选)",
  "image_url": "string (可选)",
  "url": "string (可选)",
  "technologies": ["React", "Node.js"] (可选)
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Project Title",
    "description": "Project description",
    "image_url": "string (可选)",
    "url": "string (可选)",
    "technologies": ["React", "Node.js"],
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新项目

**PUT** `/api/v1/users/me/portfolio/:id`

更新指定项目。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 项目 ID

**请求体**:
```json
{
  "title": "string (可选)",
  "description": "string (可选)",
  "image_url": "string (可选)",
  "url": "string (可选)",
  "technologies": ["React", "Node.js"] (可选)
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Updated Title",
    "description": "Updated description",
    "image_url": "string (可选)",
    "url": "string (可选)",
    "technologies": ["React", "Node.js"],
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 删除项目

**DELETE** `/api/v1/users/me/portfolio/:id`

删除指定项目。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 项目 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Project deleted successfully"
}
```

---

## 社区 (Community)

### 获取帖子列表

**GET** `/api/v1/community/posts`

获取帖子列表。

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20
- `sort` (可选): 排序方式 (`newest`, `oldest`, `popular`)

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "id": 1,
        "user_id": 1,
        "username": "john_doe",
        "avatar_url": "string (可选)",
        "content": "Post content",
        "images": ["url1", "url2"],
        "likes_count": 10,
        "comments_count": 5,
        "is_liked": false,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

---

### 获取热门帖子

**GET** `/api/v1/community/posts/trending`

获取热门帖子列表（基于点赞、评论和时间的综合评分）。

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "id": 1,
        "user_id": 1,
        "username": "john_doe",
        "avatar_url": "string (可选)",
        "content": "Post content",
        "images": ["url1", "url2"],
        "likes_count": 100,
        "comments_count": 50,
        "trending_score": 95.5,
        "is_liked": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

### 获取帖子详情

**GET** `/api/v1/community/posts/:id`

获取指定帖子的详细信息。

**路径参数**:
- `id` (必填): 帖子 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "username": "john_doe",
    "avatar_url": "string (可选)",
    "content": "Post content",
    "images": ["url1", "url2"],
    "likes_count": 10,
    "comments_count": 5,
    "is_liked": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 获取帖子评论

**GET** `/api/v1/community/posts/:id/comments`

获取指定帖子的评论列表。

**路径参数**:
- `id` (必填): 帖子 ID

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "comments": [
      {
        "id": 1,
        "post_id": 1,
        "user_id": 2,
        "username": "jane_doe",
        "avatar_url": "string (可选)",
        "content": "Comment content",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

### 获取用户帖子

**GET** `/api/v1/community/users/:id/posts`

获取指定用户的所有帖子。

**路径参数**:
- `id` (必填): 用户 ID

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "id": 1,
        "user_id": 1,
        "content": "Post content",
        "images": ["url1", "url2"],
        "likes_count": 10,
        "comments_count": 5,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 30,
      "total_pages": 2
    }
  }
}
```

---

### 创建帖子

**POST** `/api/v1/community/posts`

创建新帖子。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "content": "string (必填)",
  "images": ["url1", "url2"] (可选)
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "content": "Post content",
    "images": ["url1", "url2"],
    "likes_count": 0,
    "comments_count": 0,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新帖子

**PUT** `/api/v1/community/posts/:id`

更新指定帖子（仅作者可更新）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 帖子 ID

**请求体**:
```json
{
  "content": "string (可选)",
  "images": ["url1", "url2"] (可选)
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "content": "Updated content",
    "images": ["url1", "url2"],
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
- `403 Forbidden`: 不是帖子作者

---

### 删除帖子

**DELETE** `/api/v1/community/posts/:id`

删除指定帖子（仅作者可删除）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 帖子 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Post deleted successfully"
}
```

**错误响应**:
- `403 Forbidden`: 不是帖子作者

---

### 点赞帖子

**POST** `/api/v1/community/posts/:id/like`

点赞指定帖子。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 帖子 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Post liked successfully"
}
```

---

### 取消点赞

**DELETE** `/api/v1/community/posts/:id/like`

取消对指定帖子的点赞。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 帖子 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Post unliked successfully"
}
```

---

### 创建评论

**POST** `/api/v1/community/posts/:id/comments`

在指定帖子下创建评论。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 帖子 ID

**请求体**:
```json
{
  "content": "string (必填)"
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "post_id": 1,
    "user_id": 2,
    "username": "jane_doe",
    "avatar_url": "string (可选)",
    "content": "Comment content",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新评论

**PUT** `/api/v1/community/comments/:id`

更新指定评论（仅作者可更新）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 评论 ID

**请求体**:
```json
{
  "content": "string (必填)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "post_id": 1,
    "user_id": 2,
    "content": "Updated comment",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
- `403 Forbidden`: 不是评论作者

---

### 删除评论

**DELETE** `/api/v1/community/comments/:id`

删除指定评论（仅作者可删除）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 评论 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Comment deleted successfully"
}
```

**错误响应**:
- `403 Forbidden`: 不是评论作者

---

### 获取我的帖子

**GET** `/api/v1/community/users/me/posts`

获取当前用户的所有帖子。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "id": 1,
        "user_id": 1,
        "content": "Post content",
        "images": ["url1", "url2"],
        "likes_count": 10,
        "comments_count": 5,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 30,
      "total_pages": 2
    }
  }
}
```

---

## 活动 (Events)

### 获取活动列表

**GET** `/api/v1/events`

获取活动列表。

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20
- `sort` (可选): 排序方式 (`newest`, `oldest`, `upcoming`)
- `category` (可选): 活动类别

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": 1,
        "user_id": 1,
        "username": "john_doe",
        "title": "Event Title",
        "description": "Event description",
        "location": "Helsinki, Finland",
        "start_time": "2024-06-01T10:00:00Z",
        "end_time": "2024-06-01T18:00:00Z",
        "image_url": "string (可选)",
        "interested_count": 50,
        "going_count": 30,
        "comments_count": 10,
        "is_interested": false,
        "is_going": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

---

### 获取活动详情

**GET** `/api/v1/events/:id`

获取指定活动的详细信息。

**路径参数**:
- `id` (必填): 活动 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "username": "john_doe",
    "avatar_url": "string (可选)",
    "title": "Event Title",
    "description": "Event description",
    "location": "Helsinki, Finland",
    "start_time": "2024-06-01T10:00:00Z",
    "end_time": "2024-06-01T18:00:00Z",
    "image_url": "string (可选)",
    "interested_count": 50,
    "going_count": 30,
    "comments_count": 10,
    "is_interested": false,
    "is_going": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 获取活动评论

**GET** `/api/v1/events/:id/comments`

获取指定活动的评论列表。

**路径参数**:
- `id` (必填): 活动 ID

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "comments": [
      {
        "id": 1,
        "event_id": 1,
        "user_id": 2,
        "username": "jane_doe",
        "avatar_url": "string (可选)",
        "content": "Comment content",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

### 创建活动

**POST** `/api/v1/events`

创建新活动。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "title": "string (必填)",
  "description": "string (可选)",
  "location": "string (可选)",
  "start_time": "2024-06-01T10:00:00Z (必填, ISO 8601)",
  "end_time": "2024-06-01T18:00:00Z (可选, ISO 8601)",
  "image_url": "string (可选)"
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Event Title",
    "description": "Event description",
    "location": "Helsinki, Finland",
    "start_time": "2024-06-01T10:00:00Z",
    "end_time": "2024-06-01T18:00:00Z",
    "image_url": "string (可选)",
    "interested_count": 0,
    "going_count": 0,
    "comments_count": 0,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新活动

**PATCH** `/api/v1/events/:id`

更新指定活动（仅创建者可更新）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 活动 ID

**请求体**:
```json
{
  "title": "string (可选)",
  "description": "string (可选)",
  "location": "string (可选)",
  "start_time": "2024-06-01T10:00:00Z (可选, ISO 8601)",
  "end_time": "2024-06-01T18:00:00Z (可选, ISO 8601)",
  "image_url": "string (可选)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Updated Title",
    "description": "Updated description",
    "location": "Helsinki, Finland",
    "start_time": "2024-06-01T10:00:00Z",
    "end_time": "2024-06-01T18:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
- `403 Forbidden`: 不是活动创建者

---

### 删除活动

**DELETE** `/api/v1/events/:id`

删除指定活动（仅创建者可删除）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 活动 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Event deleted successfully"
}
```

**错误响应**:
- `403 Forbidden`: 不是活动创建者

---

### 表示感兴趣

**POST** `/api/v1/events/:id/interested`

表示对活动感兴趣。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 活动 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Interest added successfully"
}
```

---

### 取消感兴趣

**DELETE** `/api/v1/events/:id/interested`

取消对活动的感兴趣。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 活动 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Interest removed successfully"
}
```

---

### 表示参加

**POST** `/api/v1/events/:id/going`

表示将参加活动。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 活动 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Going status added successfully"
}
```

---

### 取消参加

**DELETE** `/api/v1/events/:id/going`

取消参加活动。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 活动 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Going status removed successfully"
}
```

---

### 创建活动评论

**POST** `/api/v1/events/:id/comments`

在指定活动下创建评论。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 活动 ID

**请求体**:
```json
{
  "content": "string (必填)"
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "event_id": 1,
    "user_id": 2,
    "username": "jane_doe",
    "avatar_url": "string (可选)",
    "content": "Comment content",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新活动评论

**PATCH** `/api/v1/events/comments/:id`

更新指定活动评论（仅作者可更新）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 评论 ID

**请求体**:
```json
{
  "content": "string (必填)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "event_id": 1,
    "user_id": 2,
    "content": "Updated comment",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
- `403 Forbidden`: 不是评论作者

---

### 删除活动评论

**DELETE** `/api/v1/events/comments/:id`

删除指定活动评论（仅作者可删除）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 评论 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Comment deleted successfully"
}
```

**错误响应**:
- `403 Forbidden`: 不是评论作者

---

## 机会 (Opportunities)

### 获取机会列表

**GET** `/api/v1/opportunities`

获取机会列表。

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20
- `sort` (可选): 排序方式 (`newest`, `oldest`, `deadline`)
- `type` (可选): 机会类型 (`job`, `internship`, `project`)
- `location` (可选): 地点过滤
- `remote` (可选): 是否远程 (`true`, `false`)

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "opportunities": [
      {
        "id": 1,
        "user_id": 1,
        "username": "john_doe",
        "title": "Software Engineer",
        "description": "Job description",
        "type": "job",
        "company": "Company Name",
        "location": "Helsinki, Finland",
        "remote": true,
        "deadline": "2024-06-30",
        "is_favorite": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

---

### 获取机会详情

**GET** `/api/v1/opportunities/:id`

获取指定机会的详细信息。

**路径参数**:
- `id` (必填): 机会 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "username": "john_doe",
    "avatar_url": "string (可选)",
    "title": "Software Engineer",
    "description": "Job description",
    "type": "job",
    "company": "Company Name",
    "location": "Helsinki, Finland",
    "remote": true,
    "deadline": "2024-06-30",
    "requirements": ["Skill 1", "Skill 2"],
    "is_favorite": false,
    "has_applied": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 创建机会

**POST** `/api/v1/opportunities`

创建新机会。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "title": "string (必填)",
  "description": "string (必填)",
  "type": "job|internship|project (必填)",
  "company": "string (可选)",
  "location": "string (可选)",
  "remote": true,
  "deadline": "2024-06-30 (可选, ISO 8601 日期)",
  "requirements": ["Skill 1", "Skill 2"] (可选)
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Software Engineer",
    "description": "Job description",
    "type": "job",
    "company": "Company Name",
    "location": "Helsinki, Finland",
    "remote": true,
    "deadline": "2024-06-30",
    "requirements": ["Skill 1", "Skill 2"],
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新机会

**PATCH** `/api/v1/opportunities/:id`

更新指定机会（仅创建者可更新）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 机会 ID

**请求体**:
```json
{
  "title": "string (可选)",
  "description": "string (可选)",
  "type": "job|internship|project (可选)",
  "company": "string (可选)",
  "location": "string (可选)",
  "remote": true,
  "deadline": "2024-06-30 (可选, ISO 8601 日期)",
  "requirements": ["Skill 1", "Skill 2"] (可选)
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Updated Title",
    "description": "Updated description",
    "type": "job",
    "company": "Company Name",
    "location": "Helsinki, Finland",
    "remote": true,
    "deadline": "2024-06-30",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
- `403 Forbidden`: 不是机会创建者

---

### 删除机会

**DELETE** `/api/v1/opportunities/:id`

删除指定机会（仅创建者可删除）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 机会 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Opportunity deleted successfully"
}
```

**错误响应**:
- `403 Forbidden`: 不是机会创建者

---

### 收藏机会

**POST** `/api/v1/opportunities/:id/favorite`

收藏指定机会。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 机会 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Opportunity favorited successfully"
}
```

---

### 取消收藏

**DELETE** `/api/v1/opportunities/:id/favorite`

取消对指定机会的收藏。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 机会 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Opportunity unfavorited successfully"
}
```

---

### 申请机会

**POST** `/api/v1/opportunities/:id/apply`

申请指定机会。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 机会 ID

**请求体**:
```json
{
  "cover_letter": "string (可选)",
  "resume_url": "string (可选)"
}
```

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "opportunity_id": 1,
    "user_id": 2,
    "cover_letter": "Cover letter text",
    "resume_url": "string (可选)",
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 获取我的申请

**GET** `/api/v1/opportunities/applications`

获取当前用户的所有申请。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20
- `status` (可选): 申请状态过滤 (`pending`, `accepted`, `rejected`)

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "applications": [
      {
        "id": 1,
        "opportunity_id": 1,
        "opportunity_title": "Software Engineer",
        "cover_letter": "Cover letter text",
        "resume_url": "string (可选)",
        "status": "pending",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 10,
      "total_pages": 1
    }
  }
}
```

---

### 获取机会的申请列表

**GET** `/api/v1/opportunities/:id/applications`

获取指定机会的所有申请（仅创建者可查看）。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 机会 ID

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20
- `status` (可选): 申请状态过滤 (`pending`, `accepted`, `rejected`)

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "applications": [
      {
        "id": 1,
        "opportunity_id": 1,
        "user_id": 2,
        "username": "jane_doe",
        "avatar_url": "string (可选)",
        "cover_letter": "Cover letter text",
        "resume_url": "string (可选)",
        "status": "pending",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

**错误响应**:
- `403 Forbidden`: 不是机会创建者

---

## 通知 (Notifications)

### 获取通知列表

**GET** `/api/v1/notifications`

获取当前用户的所有通知。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20
- `unread_only` (可选): 仅未读通知 (`true`, `false`)

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": 1,
        "user_id": 1,
        "type": "post_liked",
        "title": "Your post was liked",
        "message": "jane_doe liked your post",
        "related_id": 123,
        "related_type": "post",
        "read": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

### 获取未读通知数量

**GET** `/api/v1/notifications/unread-count`

获取当前用户的未读通知数量。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "count": 5
  }
}
```

---

### 标记通知为已读

**PUT** `/api/v1/notifications/:id/read`

标记指定通知为已读。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 通知 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Notification marked as read"
}
```

---

### 标记所有通知为已读

**PUT** `/api/v1/notifications/read-all`

标记当前用户的所有通知为已读。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "All notifications marked as read"
}
```

---

### 删除通知

**DELETE** `/api/v1/notifications/:id`

删除指定通知。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 通知 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Notification deleted successfully"
}
```

---

### 批量删除通知

**DELETE** `/api/v1/notifications`

批量删除通知。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "ids": [1, 2, 3] (必填, 通知 ID 数组)
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "Notifications deleted successfully",
  "data": {
    "deleted_count": 3
  }
}
```

---

## 文件管理 (Files)

### 上传文件

**POST** `/api/v1/files/upload`

上传文件。

**请求头**:
```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**请求体** (multipart/form-data):
- `file` (必填): 文件
- `description` (可选): 文件描述

**响应** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "filename": "document.pdf",
    "file_url": "https://example.com/files/document_123.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "description": "File description",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 生成预签名上传 URL

**POST** `/api/v1/files/presigned-upload`

生成用于直接上传到 S3/MinIO 的预签名 URL。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "filename": "string (必填)",
  "content_type": "string (必填, MIME type)",
  "file_size": 1024000
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "upload_url": "https://s3.example.com/upload?signature=...",
    "file_id": "unique-file-id",
    "expires_in": 3600
  }
}
```

---

### 确认上传

**POST** `/api/v1/files/confirm-upload`

确认文件已通过预签名 URL 上传完成。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "file_id": "string (必填)",
  "filename": "string (必填)",
  "description": "string (可选)"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "filename": "document.pdf",
    "file_url": "https://example.com/files/document_123.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "description": "File description",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 获取我的文件列表

**GET** `/api/v1/files`

获取当前用户的所有文件。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "files": [
      {
        "id": 1,
        "user_id": 1,
        "filename": "document.pdf",
        "file_url": "https://example.com/files/document_123.pdf",
        "file_size": 1024000,
        "mime_type": "application/pdf",
        "description": "File description",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

### 获取文件信息

**GET** `/api/v1/files/:id`

获取指定文件的信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 文件 ID

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "filename": "document.pdf",
    "file_url": "https://example.com/files/document_123.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "description": "File description",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
- `403 Forbidden`: 不是文件所有者
- `404 Not Found`: 文件不存在

---

### 删除文件

**DELETE** `/api/v1/files/:id`

删除指定文件。

**请求头**:
```
Authorization: Bearer <access_token>
```

**路径参数**:
- `id` (必填): 文件 ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "File deleted successfully"
}
```

**错误响应**:
- `403 Forbidden`: 不是文件所有者
- `404 Not Found`: 文件不存在

---

## 内部 API (Internal APIs)

内部 API 用于服务间通信，需要内部服务认证。这些 API 不对外公开，仅用于微服务之间的调用。

### 内部用户 API

**路径**: `/api/v1/internal/user/*`

**说明**: 用于其他服务查询用户信息（如检查用户资料可见性）。

**认证**: 需要内部服务认证（`X-Internal-Request: true` header）

**示例**:
- `GET /api/v1/internal/user/users/:id` - 获取用户信息（内部调用）
- `GET /api/v1/internal/user/users/:id/summary` - 获取用户摘要（内部调用）

---

### 内部通知 API

**路径**: `/api/v1/internal/notification/*`

**说明**: 用于其他服务创建通知。

**认证**: 需要内部服务认证

**示例**:
- `POST /api/v1/internal/notification/notifications` - 创建通知（内部调用）

---

### 内部文件 API

**路径**: `/api/v1/internal/file/*`

**说明**: 用于其他服务上传和管理文件。

**认证**: 需要内部服务认证

**示例**:
- `POST /api/v1/internal/file/upload` - 上传文件（内部调用）
- `GET /api/v1/internal/file/:id` - 获取文件信息（内部调用）
- `DELETE /api/v1/internal/file/:id` - 删除文件（内部调用）

---

### 内部作品集 API

**路径**: `/api/v1/internal/portfolio/*`

**说明**: 用于服务间调用作品集相关功能。

**认证**: 需要内部服务认证

---

## 错误响应格式

所有错误响应都遵循以下格式：

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {} (可选, 包含额外错误详情)
  }
}
```

### 常见 HTTP 状态码

- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未认证或认证失败
- `403 Forbidden`: 无权限访问
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突（如用户名已存在）
- `500 Internal Server Error`: 服务器内部错误

---

## 认证说明

### JWT Token 使用

所有需要认证的 API 都需要在请求头中包含 JWT token：

```
Authorization: Bearer <access_token>
```

### Token 刷新

Access token 有过期时间，当 token 过期时，需要使用 refresh token 获取新的 access token：

1. 调用 `POST /api/v1/auth/refresh` 接口
2. 使用返回的新 access token 继续请求

### Token 过期处理

当 access token 过期时，API 会返回 `401 Unauthorized` 错误。客户端应该：
1. 使用 refresh token 获取新的 access token
2. 重试原始请求

---

## 分页说明

所有列表 API 都支持分页，使用以下查询参数：

- `page`: 页码（从 1 开始）
- `limit`: 每页数量（默认 20，最大 100）

响应中包含分页信息：

```json
{
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

---

## 日期时间格式

所有日期时间字段都使用 ISO 8601 格式：

```
2024-01-01T00:00:00Z
```

日期字段使用 ISO 8601 日期格式：

```
2024-01-01
```

---

## 文件上传限制

- **最大文件大小**: 5MB（头像），10MB（其他文件）
- **支持的图片格式**: JPG, PNG, GIF
- **支持的文件格式**: PDF, DOC, DOCX, TXT 等

---

## 速率限制

API 请求有速率限制（通过 Gateway 实现）：

- **认证接口**: 每分钟 10 次请求
- **其他接口**: 每分钟 100 次请求

超过限制会返回 `429 Too Many Requests` 错误。

---

## 健康检查

**GET** `/gateway/health`

检查 Gateway 服务状态。

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "service": "gateway"
  }
}
```