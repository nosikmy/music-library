CREATE TABLE IF NOT EXISTS songs
(
    id           SERIAL PRIMARY KEY,
    name         VARCHAR,
    link         VARCHAR,
    release_date TIMESTAMP,
    first_verse_id  INTEGER,
    last_verse_id  INTEGER,
    FOREIGN KEY (first_verse_id) REFERENCES verses (id),
    FOREIGN KEY (last_verse_id) REFERENCES verses (id) ON DELETE CASCADE
)