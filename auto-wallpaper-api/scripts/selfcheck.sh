#!/usr/bin/env bash
# Minimal API self-check: login, schedule, upload, agent state, delete.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
API_BIN="${API_BIN:-}"
PORT="${PORT:-18102}"
BASE="http://127.0.0.1:${PORT}"
DATA="$(mktemp -d)"
DB="$DATA/test.db"

cleanup() {
  if [[ -n "${PID:-}" ]]; then kill "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; fi
  rm -rf "$DATA"
}
trap cleanup EXIT

if [[ -z "$API_BIN" ]]; then
  API_BIN="$DATA/api"
  (cd "$ROOT" && go build -o "$API_BIN" ./cmd/server)
fi

ADDR=":${PORT}" DATABASE_URL="$DB" DATA_DIR="$DATA" MIGRATIONS_DIR="$ROOT/migrations" \
  "$API_BIN" >"$DATA/api.log" 2>&1 &
PID=$!
for i in $(seq 1 50); do
  curl -sf "$BASE/health" >/dev/null && break
  sleep 0.1
done
curl -sf "$BASE/health" >/dev/null

TOKEN=$(curl -sf -X POST "$BASE/api/auth/login" -H 'Content-Type: application/json' \
  -d '{"username":"armin","password":"dopadopa123"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')
AUTH="Authorization: Bearer $TOKEN"

curl -sf -X PUT "$BASE/api/schedules/lock" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"mode":"hourly","intervalHours":4,"startHour":8,"enabled":true}' >/dev/null

PNG="$DATA/t.png"
python3 - "$PNG" <<'PY'
import struct, zlib, sys
path = sys.argv[1]
def chunk(t, d):
    return struct.pack(">I", len(d)) + t + d + struct.pack(">I", zlib.crc32(t + d) & 0xFFFFFFFF)
raw = b"\x00" + b"\x00\x00\x00" + b"\x00"
open(path, "wb").write(
    b"\x89PNG\r\n\x1a\n"
    + chunk(b"IHDR", struct.pack(">IIBBBBB", 1, 1, 8, 2, 0, 0, 0))
    + chunk(b"IDAT", zlib.compress(raw))
    + chunk(b"IEND", b"")
)
PY

UP=$(curl -sf -X POST "$BASE/api/screens/lock/wallpapers" -H "$AUTH" -F "file=@${PNG};filename=t.png")
ID=$(python3 -c 'import sys,json; print(json.load(sys.stdin)["id"])' <<<"$UP")

curl -sf "$BASE/api/agent/state" -H "$AUTH" >"$DATA/state.json"
python3 - "$DATA/state.json" <<'PY'
import json, sys
s = json.load(open(sys.argv[1]))
assert any(x["screen"] == "lock" and x.get("currentWallpaper") for x in s["screens"]), s
print("agent state ok")
PY

curl -sf -X DELETE "$BASE/api/wallpapers/$ID" -H "$AUTH" >/dev/null
echo "self-check passed"
