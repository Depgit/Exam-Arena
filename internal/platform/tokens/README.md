# tokens

**Job:** create and check login tokens (JWT).

- `tokens.Generate(userID, username, role, secret, hours)` → token string
- `tokens.Validate(token, secret)` → `*Claims` (user id, username, role) or an error
