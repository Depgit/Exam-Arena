# respond

**Job:** write every HTTP reply in the same JSON shape.

| You call | The client receives |
|---|---|
| `respond.JSON(w, status, data)` | `{"success": true, "data": …}` |
| `respond.Error(w, status, "message")` | `{"success": false, "error": "message"}` |
| `respond.JSONWithMeta(w, status, data, meta)` | adds `"meta"` (e.g. paging) |
