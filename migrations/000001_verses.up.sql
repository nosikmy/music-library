CREATE TABLE IF NOT EXISTS verses(
    id SERIAL PRIMARY KEY ,
    text VARCHAR,
    next INTEGER,
    FOREIGN KEY (next) REFERENCES verses (id) ON DELETE CASCADE
)