package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	errors2 "github.com/nosikmy/music-library/internal/app/errors"
	"github.com/nosikmy/music-library/internal/app/models"
	"time"
)

type SongRepository struct {
	db *sqlx.DB
}

func NewSongRepository(db *sqlx.DB) *SongRepository {
	return &SongRepository{
		db: db,
	}
}

func (s *SongRepository) GetSongText(id int, limit int, offset int) (int, []models.Verse, error) {
	const op = "repository.song.GetSongText"
	query := fmt.Sprintf(`WITH RECURSIVE verse_chain AS (
								SELECT v.id, v.text, v.next
								FROM %s v
										 INNER JOIN %s s ON v.id = s.first_verse_id
								WHERE s.id = $1
							
								UNION ALL
							
								SELECT v.id, v.text, v.next
								FROM %s v
										 INNER JOIN verse_chain vc ON v.id = vc.next
							)
							SELECT id, text
							FROM verse_chain
							LIMIT $2 OFFSET $3`, versesTable, songsTable, versesTable)

	var text []models.Verse
	err := s.db.Select(&text, query, id, limit, offset)

	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return 0, nil, fmt.Errorf("%s: %w", op, mlErr)
	}
	return len(text), text, nil

}

func (s *SongRepository) DeleteSong(id int) error {
	const op = "repository.song.DeleteSong"
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryGetLastVerseId := fmt.Sprintf(`SELECT last_verse_id FROM %s WHERE id = $1`, songsTable)
	var verseId int
	err = s.db.Get(&verseId, queryGetLastVerseId, id)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed get verse id): %w", op, mlErr)
	}

	queryDeleteVerses := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, versesTable)
	queryDeleteSong := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, songsTable)
	_, err = tx.Exec(queryDeleteSong, id)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed delete song): %w", op, mlErr)
	}
	_, err = tx.Exec(queryDeleteVerses, verseId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed delete verses): %w", op, mlErr)
	}

	if err = tx.Commit(); err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed commit): %w", op, mlErr)
	}

	return nil
}

func (s *SongRepository) AddSong(group string, song string, releaseDate time.Time, verses []string, link string) (int, error) {
	const op = "repository.song.AddSong"
	queryCheckSongExist := fmt.Sprintf(`SELECT COALESCE((SELECT s.id FROM %s s
                								JOIN %s sg on s.id = sg.song_id
												JOIN %s g on g.id = sg.group_id
												WHERE s.name = $1 and g.name = $2), 0) AS id`,
		songsTable, songsGroupsTable, groupsTable)
	var songId int
	err := s.db.Get(&songId, queryCheckSongExist, song, group)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return 0, fmt.Errorf("%s (failed get song id): %w", op, mlErr)
	}
	if songId != 0 {
		return songId, nil
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	firstVerseId, lastVerseId, err := insertVerses(tx, verses)
	if err != nil {
		return 0, err
	}

	queryInsertSong := fmt.Sprintf(`INSERT INTO %s (name, link, release_date, first_verse_id, last_verse_id)
											VALUES ($1, $2, $3, $4, $5)
											RETURNING id;`, songsTable)
	err = tx.Get(&songId, queryInsertSong, song, link, releaseDate, firstVerseId, lastVerseId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return 0, fmt.Errorf("%s (failed insert song): %w", op, mlErr)
	}

	groupId, err := addGroup(tx, group)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return 0, fmt.Errorf("%s (failed add group): %w", op, mlErr)
	}

	queryAddRelation := fmt.Sprintf(`INSERT INTO %s (song_id, group_id)  VALUES ($1, $2)`, songsGroupsTable)

	_, err = tx.Exec(queryAddRelation, songId, groupId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return 0, fmt.Errorf("%s (failed add relation between song and group): %w", op, mlErr)
	}

	if err = tx.Commit(); err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return 0, fmt.Errorf("%s (failed to commit): %w", op, mlErr)
	}

	return songId, nil
}

func insertVerses(tx *sqlx.Tx, verses []string) (int, int, error) {
	queryInsert := fmt.Sprintf(`INSERT INTO %s (text, next)
										VALUES ($1, NULL)
										RETURNING id`, versesTable)
	queryAddNextId := fmt.Sprintf(`UPDATE %s SET next = $1 WHERE id = $2`, versesTable)

	stmtInsert, err := tx.Prepare(queryInsert)
	if err != nil {
		return 0, 0, err
	}
	defer stmtInsert.Close()

	stmtAddNextId, err := tx.Prepare(queryAddNextId)
	if err != nil {
		return 0, 0, err
	}
	defer stmtAddNextId.Close()

	var prevID *int
	var firstID int
	var lastID int

	for i, verse := range verses {
		var id int
		err = stmtInsert.QueryRow(verse).Scan(&id)
		if err != nil {
			return 0, 0, err
		}

		if i == 0 {
			firstID = id
		}
		lastID = id
		if prevID != nil {
			_, err = stmtAddNextId.Exec(id, *prevID)
			if err != nil {
				return 0, 0, err
			}
		}

		prevID = &id
	}

	return firstID, lastID, nil
}

func addGroup(tx *sqlx.Tx, groupName string) (int, error) {
	queryInsert := fmt.Sprintf(`INSERT INTO %s (name)
										SELECT $1
										WHERE NOT EXISTS (
											SELECT id FROM %s WHERE name = $2
										)
										RETURNING id;
										`, groupsTable, groupsTable)
	queryGetId := fmt.Sprintf(`SELECT id FROM %s WHERE name = $1`, groupsTable)

	_, err := tx.Exec(queryInsert, groupName, groupName)
	if err != nil {
		return 0, err
	}

	var id int
	err = tx.Get(&id, queryGetId, groupName)
	if err != nil {
		return 0, err
	}
	return id, nil
}
