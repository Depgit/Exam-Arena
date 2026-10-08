# models

**Job:** the shared data shapes passed between folders — `User`, `Question`,
`Match`, `Friendship`, `DailyAttempt`… — and their JSON field names.

No logic and no database code lives here. Shapes used by only one feature
live in that feature's folder instead.
