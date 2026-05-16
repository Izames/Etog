CREATE TABLE IF NOT EXISTS official_request(
    user_id INT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    comment TEXT
);