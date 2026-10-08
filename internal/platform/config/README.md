# config

**Job:** read the server's settings from environment variables (and `.env`
in development).

**You call:** `config.Load()` → **you get:** `*Config` (port, database URL,
JWT secret, rate limits, question-generator settings…), or an error if a
required setting such as `DATABASE_URL` or `JWT_SECRET` is missing.

**Uses:** nothing else in this project. See `.env.example` for every setting.
