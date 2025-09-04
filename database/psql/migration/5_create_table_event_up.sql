CREATE TABLE event (
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