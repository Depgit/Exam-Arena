# admin

**Job:** the admin console's endpoints.

| Endpoint (admin only) | Does |
|---|---|
| `GET /api/v1/admin/stats` | users, online players, matches, questions, open reports |
| `POST /api/v1/admin/questions` | create a question |
| `PUT /api/v1/admin/questions/{id}/publish` | publish it |
| `PUT /api/v1/admin/categories/{id}/sort-order` | change category order |

Topic creation, question reports and the question generator have their own
admin endpoints in `questions/` and `questionpool/` (listed in `cmd/server/routes.go`).

**Uses:** `questions`, `users`, `match`, `platform/realtime`.
