CREATE TABLE event_archive (
    id SERIAL PRIMARY KEY,
    organizer INT,
    event INT,
    FOREIGN KEY (organizer) REFERENCES account(int),
    FOREIGN KEY (event) REFERENCES event(int)
);