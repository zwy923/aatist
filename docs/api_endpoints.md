# Backend API Endpoints

Base URL: `/api/v1`

## Authentication & Security
- **Access Token TTL**: 15 minutes.
- **Refresh Token TTL**: 30 days (Rotation enabled).
- **Logout**: Invalidates the Refresh Token. Access Token remains valid until expiry (short TTL).
- **SSO**: Supports Google and LinkedIn via OAuth (browser redirect flow).

## Public Routes
These routes are accessible without authentication.

> **Note**: Some public endpoints support **optional authentication**. If a valid `Authorization` header is provided, interaction state fields (e.g., `likedByMe`, `savedByMe`) will reflect the current user's status. Without a token, these fields default to `false` or `null`.

### Authentication (User Service)
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/auth/register` | Register a new user |
| `POST` | `/auth/login` | Login with email/password |
| `POST` | `/auth/refresh` | Refresh access token |
| `POST` | `/auth/logout` | Logout (Invalidates Refresh Token) |
| `GET` | `/auth/oauth/:provider` | Initiate OAuth flow (google, linkedin) |
| `GET` | `/auth/oauth/:provider/callback` | OAuth callback (browser redirect) |

### User Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/users/check-username` | Check if username is available |
| `GET` | `/users/check-email` | Check if email is registered |
| `GET` | `/users/:userId` | Get public user profile |
| `GET` | `/users/:userId/summary` | Get user profile summary (Includes `onlineStatus`, `lastSeenAt`) |
| `GET` | `/talents` | Search/Filter talents (See [Filters](#talent-search-filters)) |
| `GET` | `/stats/overview` | Get dashboard statistics |
| `GET` | `/skills/popular` | Get popular skills |

### Portfolio Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/portfolio/:portfolioId` | Get specific portfolio item |
| `GET` | `/users/:userId/portfolio` | Get user's portfolio |

### Community Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/community/posts` | List posts (See [Filters](#community-filters)) |
| `GET` | `/community/posts/:postId` | Get specific post (**Optional Auth**: `likedByMe`, `likeCount`) |
| `GET` | `/community/posts/:postId/comments` | Get comments for a post (Supports pagination) |
| `GET` | `/community/users/:userId/posts` | Get posts by specific user |

### Events Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/events` | List events (See [Filters](#event-filters)) |
| `GET` | `/events/:eventId` | Get specific event (**Optional Auth**: `interestedByMe`, `goingByMe`) |
| `GET` | `/events/:eventId/comments` | Get comments for an event (Supports pagination) |

### Opportunity Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/opportunities` | List opportunities (See [Filters](#opportunity-filters)) |
| `GET` | `/opportunities/:opportunityId` | Get opportunity details (**Optional Auth**: `savedByMe`) |

---

## Protected Routes
These routes require a valid JWT token in the `Authorization` header (`Bearer <token>`).

### Real-time
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/ws` | WebSocket connection for online status and notifications |

### User Service (Current User)
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/users/me` | Get current user's profile |
| `PATCH` | `/users/me` | Update current user's profile |
| `POST` | `/users/me/avatar` | Upload/Update avatar |
| `PATCH` | `/users/me/password` | Change password |
| `GET` | `/users/me/availability` | Get availability settings |
| `PATCH` | `/users/me/availability` | Update availability settings |
| `GET` | `/users/me/saved` | Get saved items (Supports `type` filter, pagination) |
| `POST` | `/users/me/saved` | Save an item (See [Saved Items Body](#saved-items-body)) |
| `DELETE` | `/users/me/saved/:savedItemId` | Remove saved item (by saved ID) |
| `DELETE` | `/users/me/saved` | Remove saved item (by `type` & `targetId` query params) |
| `GET` | `/users/me/applications` | Get my job applications |

### Portfolio Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/users/me/portfolio` | Get my portfolio |
| `POST` | `/users/me/portfolio` | Create portfolio item |
| `PATCH` | `/users/me/portfolio/:portfolioId` | Update portfolio item |
| `DELETE` | `/users/me/portfolio/:portfolioId` | Delete portfolio item |

### Notification Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/notifications` | List notifications (Query: `unread=true`) |
| `GET` | `/notifications/:notificationId` | Get notification details |
| `PATCH` | `/notifications/read-all` | Mark ALL notifications as read |
| `PATCH` | `/notifications/:notificationId/read` | Mark notification as read |
| `DELETE` | `/notifications/:notificationId` | Delete notification |

### File Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/files` | List files |
| `GET` | `/files/:fileId` | Get file metadata |
| `GET` | `/files/:fileId/download` | Download file content |
| `POST` | `/files` | Upload file |
| `DELETE` | `/files/:fileId` | Delete file |

### Opportunity Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/opportunities` | Create opportunity |
| `PATCH` | `/opportunities/:opportunityId` | Update opportunity |
| `DELETE` | `/opportunities/:opportunityId` | Delete opportunity |
| `POST` | `/opportunities/:opportunityId/applications` | Apply for opportunity |
| `GET` | `/opportunities/:opportunityId/applications` | View applications (**Recruiter Only**) |
| `GET` | `/applications/:applicationId` | Get application details (**Owner/Recruiter Only**) |
| `PATCH` | `/applications/:applicationId` | Update status (See [Application Status](#application-status)) |
| `DELETE` | `/applications/:applicationId` | Withdraw application (**Candidate Only**) |

### Events Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/events` | Create event |
| `PATCH` | `/events/:eventId` | Update event |
| `DELETE` | `/events/:eventId` | Delete event |
| `POST` | `/events/:eventId/interested` | Mark as interested |
| `DELETE` | `/events/:eventId/interested` | Unmark interested |
| `POST` | `/events/:eventId/going` | Mark as going |
| `DELETE` | `/events/:eventId/going` | Unmark going |
| `POST` | `/events/:eventId/comments` | Add comment to event |
| `PATCH` | `/events/comments/:commentId` | Update comment |
| `DELETE` | `/events/comments/:commentId` | Delete comment |

### Community Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/community/posts` | Create post |
| `PATCH` | `/community/posts/:postId` | Update post |
| `DELETE` | `/community/posts/:postId` | Delete post |
| `POST` | `/community/posts/:postId/like` | Like post |
| `DELETE` | `/community/posts/:postId/like` | Unlike post |
| `POST` | `/community/posts/:postId/comments` | Add comment to post |
| `PATCH` | `/community/comments/:commentId` | Update comment |
| `DELETE` | `/community/comments/:commentId` | Delete comment |
| `GET` | `/community/users/me/posts` | Get my posts |

---

## Internal Routes
These routes are for service-to-service communication only.
Base Path: `/api/v1/internal`

**Security**: Requires mTLS + allowlist service account. Not accessible from public network.

| Service | Path | Description |
| :--- | :--- | :--- |
| User Service | `/user/:userId/*path` | Internal user operations |
| Notification Service | `/notification/:notificationId/*path` | Internal notification operations |
| Portfolio Service | `/portfolio/:portfolioId/*path` | Internal portfolio operations |
| File Service | `/file/:fileId/*path` | Internal file operations |

---

## Additional Documentation

### Standard Response Format
All list endpoints support pagination:
```json
{
  "items": [],
  "page": 1,
  "pageSize": 20,
  "total": 123,
  "hasNext": true
}
```

### Stats Overview Response
Endpoint: `GET /stats/overview`
```json
{
  "activeStudents": 1250,
  "openOpportunities": 45,
  "ongoingProjects": 12,
  "upcomingEvents": 8
}
```

### Online Status
User objects include:
- `onlineStatus`: `online` | `offline` | `away`
- `lastSeenAt`: Timestamp (ISO 8601)

*Source: Real-time WebSocket presence + Redis TTL.*

### Interaction Fields
Resource objects include interaction state for the current user (requires optional auth on public endpoints):
- **Posts**: `likedByMe` (bool | null), `likeCount` (int), `commentCount` (int)
- **Events**: `interestedByMe` (bool | null), `goingByMe` (bool | null)
- **Opportunities**: `savedByMe` (bool | null)

### Application Status
Endpoint: `PATCH /applications/:applicationId`

**Status Values**: `submitted` | `reviewed` | `accepted` | `rejected` | `withdrawn`

**Permissions**:
- **Recruiter**: Can change to `reviewed`, `accepted`, `rejected`
- **Candidate**: Can only change to `withdrawn`

Body:
```json
{
  "status": "reviewed"
}
```

### Talent Search Filters
Endpoint: `GET /talents`
Query Parameters (camelCase):
- `q`: Fuzzy search by name, skill, or field.
- `skills`: List of skills (e.g., `skills=React&skills=Python`).
- `school`: Filter by school name.
- `availabilityMin` / `availabilityMax`: Weekly hours range.
- `online`: `true` to filter by online status.
- `page`, `pageSize`: Pagination.
- `sort`: `createdAt`, `name`.

### Opportunity Filters
Endpoint: `GET /opportunities`
Query Parameters:
- `q`: Search query.
- `category`: Job category.
- `location`: Job location.
- `workLanguage`: Language requirement.
- `budgetMin` / `budgetMax`: Budget range.
- `payType`: `hourly` or `fixed`.
- `sort`: `startDate`, `publishedAt`, `budget`.
- `order`: `asc` or `desc`.
- `page`, `pageSize`.

### Event Filters
Endpoint: `GET /events`
Query Parameters:
- `q`: Search query.
- `time`: `today`, `week`, `month`, `all`.
- `type`: `workshop`, `talk`, `hackathon`, etc.
- `school`: Filter by school.
- `free`: `true` or `false`.
- `sort`: `startsAt`.
- `page`, `pageSize`.

### Community Filters
Endpoint: `GET /community/posts`
Query Parameters:
- `type`: `discussion` or `sticky_note`.
- `tag`: Filter by tag.
- `category`: Filter by category.
- `q`: Search query.
- `sort`: `createdAt` or `trending`.
- `page`, `pageSize`.

### Saved Items Body
Endpoint: `POST /users/me/saved`
```json
{
  "type": "opportunity|user|event|post",
  "targetId": "uuid-of-the-item"
}
```
