# windows-auto-wallpaper

Local stack to manage and apply Windows lock-screen and home-screen wallpapers on a schedule.

## Components

| Folder | Role |
|--------|------|
| [`auto-wallpaper-api`](auto-wallpaper-api/) | Gin + SQLite API (`:8101`) |
| [`auto-wallpaper-webui`](auto-wallpaper-webui/) | Vue 3 management UI (`:5181`) |
| [`auto-wallpaper`](auto-wallpaper/) | Headless Windows CLI/service `autowall` |

Default login: **`armin` / `dopadopa123`**

## Quick start (dev)

```bash
# API
cd auto-wallpaper-api && cp .env.example .env && go run ./cmd/server

# WebUI (another terminal)
cd auto-wallpaper-webui && cp .env.example .env && npm install && npm run dev

# Windows agent (on Windows, after building)
cd auto-wallpaper
GOOS=windows GOARCH=amd64 go build -o autowall.exe ./cmd/autowall
autowall doctor
autowall service start
```

## WebUI pages

- `/lock-screen` — browse / upload / delete / enable-disable wallpapers + schedule
- `/home-screen` — identical UI for desktop wallpaper

## Schedule

Per screen (`lock` or `home`):

- **Daily** — rotate once per day starting at hour `HH`
- **Hourly** — every **1 / 2 / 4 / 8** hours, slots aligned from start hour `HH`

Rotation is sequential among enabled wallpapers.

## `autowall` CLI (no GUI)

```
autowall doctor
autowall service start
autowall service stop
autowall service restart
autowall uninstall
```

Service polls `GET /api/agent/state`, downloads the current images, applies home via `SystemParametersInfo`, lock via `PersonalizationCSP` registry.
