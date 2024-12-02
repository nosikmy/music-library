package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	errors2 "github.com/nosikmy/music-library/internal/app/errors"
	"github.com/nosikmy/music-library/internal/app/models"
)

type VersesRepository struct {
	db *sqlx.DB
}

func NewVersesRepository(db *sqlx.DB) *VersesRepository {
	return &VersesRepository{
		db: db,
	}
}

func (v *VersesRepository) AddVerse(id int, newVerse *models.Verse) error {
	const op = "repository.verses.AddVerse"
	tx, err := v.db.Beginx()
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed to begin transaction): %w", op, mlErr)
	}
	defer tx.Rollback()

	var nextVerseId int

	if newVerse.Id == 0 {
		queryGetNext := fmt.Sprintf(`SELECT COALESCE((SELECT first_verse_id FROM %s WHERE id = $1), 0) AS next`, songsTable)
		err = tx.Get(&nextVerseId, queryGetNext, id)

	} else {
		queryGetNext := fmt.Sprintf(`SELECT COALESCE((SELECT next FROM %s WHERE id = $1), 0) AS next`, versesTable)
		err = tx.Get(&nextVerseId, queryGetNext, newVerse.Id)
	}
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed to get next verse id): %w", op, mlErr)
	}

	var newVerseId int
	if nextVerseId == 0 {
		queryAddVerse := fmt.Sprintf(`INSERT INTO %s (text, next) VALUES ($1, null) RETURNING id`, versesTable)
		err = tx.Get(&newVerseId, queryAddVerse, newVerse.Text)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed to add new verse to end): %w", op, mlErr)
		}
		err = addLastVerseToSong(tx, id, newVerseId)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed to update song last verse id): %w", op, mlErr)
		}
	} else {
		queryAddVerse := fmt.Sprintf(`INSERT INTO %s (text, next) VALUES ($1, $2) RETURNING id`, versesTable)
		err = tx.Get(&newVerseId, queryAddVerse, newVerse.Text, nextVerseId)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed to add verse): %w", op, mlErr)
		}
	}

	if newVerse.Id == 0 {
		err = addFirstVerseToSong(tx, id, newVerseId)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed to update song first verse id): %w", op, mlErr)
		}
	} else {
		err = addNextVerse(tx, newVerseId, newVerse.Id)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed to update next field for previous verse): %w", op, mlErr)
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (v *VersesRepository) ChangeVerse(changeVerse *models.Verse) error {
	const op = "repository.verses.ChangeVerse"
	query := fmt.Sprintf(`UPDATE %s SET text = $1 WHERE id = $2`, versesTable)
	_, err := v.db.Exec(query, changeVerse.Text, changeVerse.Id)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s: %w", op, mlErr)
	}
	return nil
}

func (v *VersesRepository) DeleteVerse(id int, verseId int) error {
	const op = "repository.verses.DeleteVerse"
	tx, err := v.db.Beginx()
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed to begin transaction): %w", op, mlErr)
	}
	defer tx.Rollback()

	queryGetNext := fmt.Sprintf(`SELECT COALESCE((SELECT next FROM %s WHERE id = $1), 0) AS next`, versesTable)
	queryGetPrev := fmt.Sprintf(`SELECT COALESCE((SELECT id FROM %s WHERE next = $1), 0) AS next`, versesTable)

	var nextId, prevId int
	err = tx.Get(&nextId, queryGetNext, verseId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed get next verse): %w", op, mlErr)
	}
	err = tx.Get(&prevId, queryGetPrev, verseId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed get previous verse): %w", op, mlErr)
	}

	if nextId == 0 && prevId == 0 {
		queryChangeSong := fmt.Sprintf(`UPDATE %s SET first_verse_id = null, last_verse_id = null WHERE id = $1`, songsTable)
		_, err = tx.Exec(queryChangeSong, id)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed update song first and last verse id): %w", op, mlErr)
		}
	} else if nextId == 0 {
		queryNextNull := fmt.Sprintf(`UPDATE %s SET next = null WHERE id = $1`, versesTable)
		_, err = tx.Exec(queryNextNull, prevId)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed to add verse in the end): %w", op, mlErr)
		}
		err = addLastVerseToSong(tx, id, prevId)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed update song name): %w", op, mlErr)
		}
	} else if prevId == 0 {
		err = addFirstVerseToSong(tx, id, nextId)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed tu update song first verse id): %w", op, mlErr)
		}
	} else {
		err = addNextVerse(tx, nextId, prevId)
		if err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return fmt.Errorf("%s (failed to update next field for previous verse): %w", op, mlErr)
		}
	}

	queryDelete := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, versesTable)
	_, err = tx.Exec(queryDelete, verseId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed to delete verse): %w", op, mlErr)
	}

	if err = tx.Commit(); err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed to commit): %w", op, mlErr)
	}

	return nil
}

func addFirstVerseToSong(tx *sqlx.Tx, songId int, verseId int) error {
	query := fmt.Sprintf(`UPDATE %s SET first_verse_id = $1 WHERE id = $2`, songsTable)
	_, err := tx.Exec(query, verseId, songId)
	return err
}

func addLastVerseToSong(tx *sqlx.Tx, songId int, verseId int) error {
	query := fmt.Sprintf(`UPDATE %s SET last_verse_id = $1 WHERE id = $2`, songsTable)
	_, err := tx.Exec(query, verseId, songId)
	return err
}

func addNextVerse(tx *sqlx.Tx, nextId int, verseId int) error {
	query := fmt.Sprintf(`UPDATE %s SET next = $1 WHERE id = $2`, versesTable)
	_, err := tx.Exec(query, nextId, verseId)
	return err
}
