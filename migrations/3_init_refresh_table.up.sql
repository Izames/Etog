CREATE TABLE IF NOT EXISTS refresh (
    token_id VARCHAR(255) UNIQUE NOT NULL,
    user_id INT NOT NULL
);

CREATE INDEX IF NOT EXISTS user_refresh ON refresh(user_id);