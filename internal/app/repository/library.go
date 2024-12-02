package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	errors2 "github.com/nosikmy/music-library/internal/app/errors"
	"github.com/nosikmy/music-library/internal/app/models"
	"time"
)

type LibraryRepository struct {
	db *sqlx.DB
}

func NewLibraryRepository(db *sqlx.DB) *LibraryRepository {
	return &LibraryRepository{
		db: db,
	}
}

func (l *LibraryRepository) GetLibrary(limit, offset int, searchText string, dateFrom, dateTo time.Time) ([]models.SongDBFormat, error) {
	const op = "repository.library.GetLibrary"
	query := fmt.Sprintf(`
		SELECT
			s.id, s.name, s.link, s.release_date,
			g.id AS group_id, g.name AS group_name
		FROM %s s
		JOIN %s sg ON s.id = sg.song_id
		JOIN %s g ON sg.group_id = g.id `, songsTable, songsGroupsTable, groupsTable)
	if searchText != "" || !dateFrom.IsZero() || !dateTo.IsZero() {
		query += " WHERE "
	}
	args := 0
	if searchText != "" {
		query += `(s.name ILIKE '%' || :search_text || '%' OR g.name ILIKE '%' || :search_text || '%') `
		args++
	}
	if !dateFrom.IsZero() {
		if args > 0 {
			query += "AND "
		}
		query += `s.release_date >= :start_date `
		args++
	}
	if !dateTo.IsZero() {
		if args > 0 {
			query += "AND "
		}
		query += `s.release_date <= :end_date `
	}
	query += `LIMIT :limit OFFSET :offset`

	filters := map[string]interface{}{
		"search_text": searchText,
		"start_date":  dateFrom,
		"end_date":    dateTo,
		"limit":       limit,
		"offset":      offset,
	}

	var songsData []models.SongDBFormat
	rows, err := l.db.NamedQuery(query, filters)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return songsData, nil
		}
		mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
		return nil, fmt.Errorf("%s: %w", op, mlErr)
	}
	defer rows.Close()

	for rows.Next() {
		var songDB models.SongDBFormat
		if err = rows.StructScan(&songDB); err != nil {
			mlErr := errors2.NewMusicLibraryError(errors2.InternalError, err)
			return nil, fmt.Errorf("%s: %w", op, mlErr)
		}
		songsData = append(songsData, songDB)
	}
	return songsData, nil
}
