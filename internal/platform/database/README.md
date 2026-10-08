# database

**Job:** everything about *reaching* Postgres — not what's stored in it.

| You call | You get back |
|---|---|
| `NewPostgres(ctx, url, max, min)` | a connection pool |
| `RunMigrations(ctx, pool)` | applies new files from `/migrations`, in order, once each |
| `IsUniqueViolation(err)` / `IsForeignKeyViolation(err)` / `IsInvalidInput(err)` | whether a DB error is of that kind, so stores can turn it into a friendly message |

**Uses:** nothing else in this project.
