-- v1: Initial revision

CREATE TABLE IF NOT EXISTS user_refresh_tokens (
    user_id     TEXT PRIMARY KEY,
    token       TEXT
);
