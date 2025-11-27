-- CREATE TABLE IF NOT EXISTS account (
--     id SERIAL PRIMARY KEY,
--     mail VARCHAR(50) UNIQUE NOT NULL,
--     login VARCHAR(50) UNIQUE NOT NULL,
--     password VARCHAR(255) NOT NULL,
--     avatar VARCHAR(255),
--     official BOOLEAN NOT NULL,
--     description VARCHAR(255),
--     rating INT NOT NULL,
--     deleted BOOLEAN NOT NULL
-- );
--
-- CREATE TABLE IF NOT EXISTS subscribes (
--     account_id INT NOT NULL,
--     subscriber_id INT NOT NULL,
--     FOREIGN KEY (account_id) REFERENCES account (id),
--     FOREIGN KEY (subscriber_id) REFERENCES account (id)
-- );
--
-- -- поиск всех подписчиков пользователя
-- CREATE INDEX IF NOT EXISTS idx_subscribers_account
-- ON subscribes (account_id);
--
-- -- поиск всех на кого пользователь подписан
-- CREATE INDEX IF NOT EXISTS idx_account_subscribes
-- ON subscrubes (subscriber_id);
--
-- -- поиск конкретной подписки
-- CREATE INDEX IF NOT EXISTS idx_current_subscribe
-- ON subscribes (account_id, subscriber_id);
--
-- -- запретить повторение подписок
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_subscribe_unique
-- ON subscribes (account_id, subscriber_id);
--
-- CREATE TABLE IF NOT EXISTS event (
--     id SERIAL PRIMARY KEY,
--     organizer INT NOT NULL,
--     name VARCHAR(50) NOT NULL,
--     price FLOAT,
--     address VARCHAR(255),
--     type VARCHAR (50) NOT NULL ,
--     date TIME,
--     description VARCHAR(510),
--     passed VARCHAR(50) NOT NULL,
--     max_people INT,
--     media VARCHAR(255)[],
--     contact INT NOT NULL,
--     deleted BOOLEAN NOT NULL,
--     FOREIGN KEY (organizer) REFERENCES account(id),
--     FOREIGN KEY (contact) REFERENCES contact(id)
-- );
--
-- CREATE TABLE IF NOT EXISTS contact (
--     id SERIAL PRIMARY KEY,
--     phone VARCHAR(30),
--     mail VARCHAR(50),
--     telegram VARCHAR (50),
--     deleted BOOLEAN NOT NULL
-- );
--
-- CREATE TABLE IF NOT EXISTS joined_on_event (
--     event_id INT NOT NULL,
--     user_id INT NOT NULL,
--     rate INT NOT NULL,
--     FOREIGN KEY (event_id) REFERENCES event(id),
--     FOREIGN KEY (user_id) REFERENCES account(id)
-- );

CREATE TABLE IF NOT EXISTS mock_event (
    id SERIAL PRIMARY KEY,
    organizer VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    price FLOAT,
    address VARCHAR(255),
    type VARCHAR (50) NOT NULL ,
    date TIMESTAMP,
    description VARCHAR(510),
    passed VARCHAR(50) NOT NULL,
    max_people INT,
    media JSONB,
    phone VARCHAR(30),
    mail VARCHAR(50),
    telegram VARCHAR (50),
    deleted BOOLEAN NOT NULL
)