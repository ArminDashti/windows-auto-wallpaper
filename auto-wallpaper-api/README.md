# auto-wallpaper-api

Gin + SQLite API for Windows auto wallpaper.

## Quick start

```bash
cp .env.example .env
go run ./cmd/server
```

Default login: `armin` / `dopadopa123`  
Listens on `:8101`.

## Env

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDR` | `:8101` | Listen address |
| `DATABASE_URL` | `./data/auto-wallpaper.db` | SQLite path |
| `MIGRATIONS_DIR` | `migrations` | SQL migrations |
| `DATA_DIR` | `./data` | Wallpaper files root |
