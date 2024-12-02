CREATE TABLE IF NOT EXISTS songs_groups
(
    song_id  INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    FOREIGN KEY (song_id) REFERENCES songs (id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE
)