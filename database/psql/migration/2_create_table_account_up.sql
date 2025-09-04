CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    mail VARCHAR(50) NOT NULL,
    login VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    subscribers INT NOT NULL,
    avatar VARCHAR(255),
    official BOOLEAN NOT NULL,
    description VARCHAR(255)
);