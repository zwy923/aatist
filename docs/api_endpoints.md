# Backend API Endpoints

Base URL: `/api/v1`

## Authentication & Security
- **Access Token TTL**: 15 minutes.
- **Refresh Token TTL**: 30 days (Rotation enabled).
- **Logout**: Invalidates the Refresh Token. Access Token remains valid until expiry (short TTL).
- **SSO**: Supports Google and LinkedIn via OAuth.

## Public Routes
These routes are accessible without authentication.

### Authentication (User Service)
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/auth/register` | Register a new user |
| `POST` | `/auth/login` | Login with email/password |
| `POST` | `/auth/refresh` | Refresh access token |
| `POST` | `/auth/logout` | Logout (Invalidates Refresh Token) |
| `GET` | `/auth/oauth/:provider` | Initiate OAuth flow (google, linkedin) |
| `GET` | `/auth/oauth/:provider/callback` | OAuth callback |

### User Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/users/check-username` | Check if username is available |
| `GET` | `/users/check-email` | Check if email is registered |
| `GET` | `/users/:id` | Get public user profile |
| `GET` | `/users/:id/summary` | Get user profile summary (Includes `onlineStatus`, `lastSeenAt`) |
| `GET` | `/talents` | Search/Filter talents (See [Filters](#talent-search-filters)) |
| `GET` | `/stats/overview` | Get dashboard statistics |
| `GET` | `/skills/popular` | Get popular skills |

### Portfolio Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/portfolio/:id` | Get specific portfolio item |
| `GET` | `/users/:id/portfolio` | Get user's portfolio |

### Community Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/community/posts` | List posts (See [Filters](#community-filters)) |
| `GET` | `/community/posts/trending` | List trending posts |
| `GET` | `/community/posts/:id` | Get specific post (Includes `likedByMe`, `likeCount`) |
| `GET` | `/community/posts/:id/comments` | Get comments for a post |
| `GET` | `/community/users/:id/posts` | Get posts by specific user |

### Events Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/events` | List events (See [Filters](#event-filters)) |
| `GET` | `/events/:id` | Get specific event (Includes `interestedByMe`, `goingByMe`) |
| `GET` | `/events/:id/comments` | Get comments for an event |

### Opportunity Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/opportunities` | List opportunities (See [Filters](#opportunity-filters)) |
| `GET` | `/opportunities/:id` | Get opportunity details (Includes `savedByMe`) |

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
| `GET` | `/users/me` | Get current user's profile (Alias: `GET /auth/me`) |
| `PATCH` | `/users/me` | Update current user's profile |
| `POST` | `/users/me/avatar` | Upload/Update avatar |
| `PATCH` | `/users/me/password` | Change password |
| `GET` | `/users/me/availability` | Get availability settings |
| `PATCH` | `/users/me/availability` | Update availability settings |
| `GET` | `/users/me/saved` | Get saved items (Supports `type` filter, pagination) |
| `POST` | `/users/me/saved` | Save an item (See [Saved Items Body](#saved-items-body)) |
| `DELETE` | `/users/me/saved/:id` | Remove saved item (by saved ID) |
| `DELETE` | `/users/me/saved` | Remove saved item (by `type` & `targetId` query params) |
| `GET` | `/users/me/applications` | Get my job applications |

### Portfolio Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/users/me/portfolio` | Get my portfolio |
| `POST` | `/users/me/portfolio` | Create portfolio item |
| `PATCH` | `/users/me/portfolio/:id` | Update portfolio item |
| `DELETE` | `/users/me/portfolio/:id` | Delete portfolio item |

### Notification Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/notifications` | List notifications (Query: `unread=true`) |
| `GET` | `/notifications/:id` | Get notification details |
| `PATCH` | `/notifications/read-all` | Mark ALL notifications as read |
| `PATCH` | `/notifications/:id/read` | Mark notification as read |
| `DELETE` | `/notifications/:id` | Delete notification |

### File Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/files` | List files |
| `GET` | `/files/:id` | Get file metadata |
| `GET` | `/files/:id/download` | Download file content |
| `POST` | `/files` | Upload file |
| `DELETE` | `/files/:id` | Delete file |

### Opportunity Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/opportunities` | Create opportunity |
| `PATCH` | `/opportunities/:id` | Update opportunity |
| `DELETE` | `/opportunities/:id` | Delete opportunity |
| `POST` | `/opportunities/:id/applications` | Apply for opportunity |
| `GET` | `/opportunities/:id/applications` | View applications (**Recruiter Only**) |
| `GET` | `/applications/:id` | Get application details (**Owner/Recruiter Only**) |
| `PATCH` | `/applications/:id` | Update status (**Recruiter Only**) |
| `DELETE` | `/applications/:id` | Withdraw application (**Candidate Only**) |

### Events Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/events` | Create event |
| `PATCH` | `/events/:id` | Update event |
| `DELETE` | `/events/:id` | Delete event |
| `POST` | `/events/:id/interested` | Mark as interested |
| `DELETE` | `/events/:id/interested` | Unmark interested |
| `POST` | `/events/:id/going` | Mark as going |
| `DELETE` | `/events/:id/going` | Unmark going |
| `POST` | `/events/:id/comments` | Add comment to event |
| `PATCH` | `/events/comments/:id` | Update comment |
| `DELETE` | `/events/comments/:id` | Delete comment |

### Community Service
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/community/posts` | Create post |
| `PATCH` | `/community/posts/:id` | Update post |
| `DELETE` | `/community/posts/:id` | Delete post |
| `POST` | `/community/posts/:id/like` | Like post |
| `DELETE` | `/community/posts/:id/like` | Unlike post |
| `POST` | `/community/posts/:id/comments` | Add comment to post |
| `PATCH` | `/community/comments/:id` | Update comment |
| `DELETE` | `/community/comments/:id` | Delete comment |
| `GET` | `/community/users/me/posts` | Get my posts |

---

## Internal Routes
These routes are for service-to-service communication and are typically not called directly by the frontend.
Base Path: `/api/v1/internal`

| Service | Path | Description |
| :--- | :--- | :--- |
| User Service | `/user/*path` | Internal user operations |
| Notification Service | `/notification/*path` | Internal notification operations |
| Portfolio Service | `/portfolio/*path` | Internal portfolio operations |
| File Service | `/file/*path` | Internal file operations |

## Additional Documentation

### Standard Response Format
All list endpoints support pagination and return the following structure:
```json
{
  "items": [],
  "page": 1,
  "pageSize": 20,
  "total": 123,
  "hasNext": true
}
```

### Online Status
User objects (in summary or details) include:
- `onlineStatus`: `online` | `offline` | `away`
- `lastSeenAt`: Timestamp (ISO 8601)
*Source: Real-time WebSocket presence + Redis TTL.*

### Interaction Fields
Resource objects include interaction state for the current user:
- **Posts**: `likedByMe` (bool), `likeCount` (int), `commentCount` (int)
- **Events**: `interestedByMe` (bool), `goingByMe` (bool)
- **Opportunities**: `savedByMe` (bool)

### Talent Search Filters
Endpoint: `GET /talents`
Query Parameters (camelCase):
- `q`: Fuzzy search by name, skill, or field.
- `skills`: List of skills (e.g., `skills=React&skills=Python` or `skills=React,Python`).
- `school`: Filter by school name.
- `availabilityMin`: Minimum weekly hours.
- `availabilityMax`: Maximum weekly hours.
- `online`: `true` to filter by online status.
- `page`: Page number (default 1).
- `pageSize`: Items per page (default 20).
- `sort`: Sort field (e.g., `createdAt`, `name`).

### Opportunity Filters
Endpoint: `GET /opportunities`
Query Parameters:
- `q`: Search query.
- `category`: Job category.
- `location`: Job location.
- `workLanguage`: Language requirement.
- `budgetMin` / `budgetMax`: Budget range.
- `payType`: `hourly` or `fixed`.
- `sort`: `startDate`, `publishedDate`, `budget`.
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
- `sort`: `latest` or `trending`.
- `page`, `pageSize`.

### Saved Items Body
Endpoint: `POST /users/me/saved`
Body:
```json
{
  "type": "opportunity|user|event|post",
  "targetId": "uuid-of-the-item"
}
```
