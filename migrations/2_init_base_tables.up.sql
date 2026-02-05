CREATE EXTENSION vector;

CREATE TABLE IF NOT EXISTS account (
    id SERIAL PRIMARY KEY,
    mail VARCHAR(255) UNIQUE NOT NULL,
    login VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    official BOOLEAN NOT NULL,
    description TEXT,
    rating INT NOT NULL,
    deleted BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS account_login ON account(login) WHERE deleted is NOT TRUE;


CREATE TABLE IF NOT EXISTS subscribes (
    account_id INT NOT NULL,
    subscriber_id INT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES account (id),
    FOREIGN KEY (subscriber_id) REFERENCES account (id)
);
-- подписки аккаунта
CREATE INDEX IF NOT EXISTS subscribers_on_account ON subscribes(account_id);

-- кто подписан на аккаунт
CREATE INDEX IF NOT EXISTS accounts_subscribers ON subscribes(subscriber_id);

-- поиск конкретной подписки
CREATE INDEX IF NOT EXISTS idx_current_subscribe
ON subscribes (account_id, subscriber_id);

-- запретить повторение подписок
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscribe_unique
ON subscribes (account_id, subscriber_id);

CREATE TABLE IF NOT EXISTS contact (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(30),
    mail VARCHAR(50),
    telegram VARCHAR (50)
);

CREATE TABLE IF NOT EXISTS event (
    id SERIAL PRIMARY KEY,
    organizer INT NOT NULL,
    name VARCHAR(50) NOT NULL,
    price FLOAT,
    address VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    cords vector(2),
    type VARCHAR (50) NOT NULL,
    date TIME,
    description TEXT,
    passed VARCHAR(50) NOT NULL,
    max_people INT,
    media JSONB,
    contact INT NOT NULL,
    deleted BOOLEAN NOT NULL,
    FOREIGN KEY (organizer) REFERENCES account(id),
    FOREIGN KEY (contact) REFERENCES contact(id)
);

CREATE INDEX IF NOT EXISTS idx_event_organizer ON event (organizer) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_name ON event (name) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_price ON event (price) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_city ON event (city) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_cords ON event (cords) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_type ON event (type) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_date ON event (date) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_max_people ON event (max_people) WHERE deleted IS NOT TRUE;
CREATE INDEX IF NOT EXISTS idx_event_passed ON event (passed) WHERE deleted IS NOT TRUE;
--прикинуть индекс cords когда сайт станет доступен

CREATE TABLE IF NOT EXISTS joined_on_event (
    event_id INT NOT NULL,
    user_id INT NOT NULL,
    rate INT NOT NULL,
    FOREIGN KEY (event_id) REFERENCES event(id),
    FOREIGN KEY (user_id) REFERENCES account(id)
);


-- подписки мероприятия
CREATE INDEX IF NOT EXISTS idx_joined_event ON joined_on_event(event_id);

-- на какие ивенты пользователь пойдет
CREATE INDEX IF NOT EXISTS idx_on_event_join ON joined_on_event(user_id);

-- поиск конкретной подписки
CREATE INDEX IF NOT EXISTS idx_current_join
    ON joined_on_event (user_id, event_id);

-- запретить повторение подписок
CREATE UNIQUE INDEX IF NOT EXISTS idx_join_unique
    ON joined_on_event (user_id, event_id);
