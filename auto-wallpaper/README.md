# auto-wallpaper (autowall)

Headless Windows CLI/service that applies lock and home wallpapers from `auto-wallpaper-api`.

## Commands

```
autowall doctor
autowall service start
autowall service stop
autowall service restart
autowall uninstall
autowall run
```

## Build (on Windows or cross-compile)

```bash
GOOS=windows GOARCH=amd64 go build -o autowall.exe ./cmd/autowall
```

## Config

`%ProgramData%\AutoWallpaper\config.env` or env:

| Variable | Default |
|----------|---------|
| `AUTOWALL_API_URL` | `http://127.0.0.1:8101` |
| `AUTOWALL_USERNAME` | `armin` |
| `AUTOWALL_PASSWORD` | `dopadopa123` |
| `AUTOWALL_POLL_SECONDS` | `60` |

Service start installs Windows service `AutoWallpaper` pointing at this binary with arg `service`.
