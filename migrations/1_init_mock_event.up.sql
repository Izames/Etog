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