package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	errors2 "github.com/nosikmy/music-library/internal/app/errors"
)

type SongChangerRepository struct {
	db *sqlx.DB
}

func NewSongChangerRepository(db *sqlx.DB) *SongChangerRepository {
	return &SongChangerRepository{
		db: db,
	}
}

func (s *SongChangerRepository) ChangeSongName(id int, newName string) error {
	const op = "repository.song_changer.ChangeSongName"
	query := fmt.Sprintf(`UPDATE %s SET name = $1 WHERE id = $2`, songsTable)
	_, err := s.db.Exec(query, newName, id)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s: %w", op, mlErr)
	}
	return nil
}

func (s *SongChangerRepository) AddGroupToSong(id int, group string) error {
	const op = "repository.song_changer.AddGroupToSong"
	tx, err := s.db.Beginx()
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed begin transaction): %w", op, mlErr)
	}
	defer tx.Rollback()

	groupId, err := addGroup(tx, group)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed add Group): %w", op, mlErr)
	}

	queryAddRelation := fmt.Sprintf(`INSERT INTO %s (song_id, group_id)  VALUES ($1, $2)`, songsGroupsTable)

	_, err = tx.Exec(queryAddRelation, id, groupId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed add relaton beetween song and group): %w", op, mlErr)
	}

	if err = tx.Commit(); err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s (failed to commit): %w", op, mlErr)
	}
	return nil
}

func (s *SongChangerRepository) DeleteGroupFromSong(id int, groupId int) error {
	const op = "repository.song_changer.DeleteGroupFromSong"
	query := fmt.Sprintf(`DELETE FROM %s WHERE song_id = $1 AND group_id = $2`, songsGroupsTable)
	_, err := s.db.Exec(query, id, groupId)
	if err != nil {
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return fmt.Errorf("%s: %w", op, mlErr)
	}
	return nil
}
