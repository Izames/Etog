CREATE TABLE subscribes (
    id SERIAL PRIMARY KEY,
    account_id INT,
    subscriber_id INT,
    FOREIGN KEY (account_id) REFERENCES account (id),
    FOREIGN KEY (subscriber_id) REFERENCES account (id)
)