CREATE TABLE IF NOT EXISTS contact (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(30),
    mail VARCHAR(50),
    telegram VARCHAR (50)
);

CREATE TABLE IF NOT EXISTS account (
    id SERIAL PRIMARY KEY,
    mail VARCHAR(50) NOT NULL,
    login VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    subscribers INT NOT NULL,
    avatar VARCHAR(255),
    official BOOLEAN NOT NULL,
    description VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS subscribes (
    id SERIAL PRIMARY KEY,
    account_id INT,
    subscriber_id INT,
    FOREIGN KEY (account_id) REFERENCES account (id),
    FOREIGN KEY (subscriber_id) REFERENCES account (id)
);

CREATE TABLE IF NOT EXISTS address (
    id SERIAL PRIMARY KEY,
    country VARCHAR(50),
    city VARCHAR(50),
    full_address VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS event (
    id SERIAL PRIMARY KEY,
    organizer INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    price FLOAT NOT NULL,
    address INT,
    type VARCHAR (50) NOT NULL ,
    date TIME,
    description VARCHAR(255),
    media VARCHAR(255)[],
    contact INT,
    visible BOOLEAN NOT NULL,
    FOREIGN KEY (organizer) REFERENCES account(id),
    FOREIGN KEY (address) REFERENCES address(id),
    FOREIGN KEY (contact) REFERENCES contact(id)
);

CREATE TABLE IF NOT EXISTS event_archive (
    id SERIAL PRIMARY KEY,
    organizer INT,
    event INT,
    FOREIGN KEY (organizer) REFERENCES account(int),
    FOREIGN KEY (event) REFERENCES event(int)
);