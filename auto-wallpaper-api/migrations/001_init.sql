-- Core schema for auto-wallpaper-api.

CREATE TABLE IF NOT EXISTS schema_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('admin')),
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

-- Default admin: armin / dopadopa123 (bcrypt). Insert only when missing.
INSERT INTO users (username, password_hash, role)
SELECT 'armin', '$2a$10$A5YA5wI3D3DQWG0ZPLa7eOO42sQnlzgBYygP4SlbEuiyvVG1sShca', 'admin'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'armin' COLLATE NOCASE);

CREATE TABLE IF NOT EXISTS wallpapers (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    screen      TEXT NOT NULL CHECK (screen IN ('lock', 'home')),
    filename    TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    enabled     INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_wallpapers_screen ON wallpapers (screen, sort_order, id);

CREATE TABLE IF NOT EXISTS schedules (
    screen         TEXT PRIMARY KEY CHECK (screen IN ('lock', 'home')),
    mode           TEXT NOT NULL DEFAULT 'daily' CHECK (mode IN ('daily', 'hourly')),
    interval_hours INTEGER NOT NULL DEFAULT 1 CHECK (interval_hours IN (1, 2, 4, 8)),
    start_hour     INTEGER NOT NULL DEFAULT 0 CHECK (start_hour >= 0 AND start_hour <= 23),
    enabled        INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1))
);

INSERT INTO schedules (screen, mode, interval_hours, start_hour, enabled)
SELECT 'lock', 'daily', 1, 0, 1
WHERE NOT EXISTS (SELECT 1 FROM schedules WHERE screen = 'lock');

INSERT INTO schedules (screen, mode, interval_hours, start_hour, enabled)
SELECT 'home', 'daily', 1, 0, 1
WHERE NOT EXISTS (SELECT 1 FROM schedules WHERE screen = 'home');

INSERT INTO schema_meta (key, value)
VALUES ('schema', '001')
ON CONFLICT (key) DO UPDATE SET value = excluded.value;
