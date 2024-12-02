package services

import (
	"fmt"
	"github.com/nosikmy/music-library/internal/app/models"
	"log/slog"
	"time"
)

type LibraryRepository interface {
	GetLibrary(limit int, offset int,
		searchText string, dateFrom, dateTo time.Time) ([]models.SongDBFormat, error)
}

type LibraryService struct {
	logger            *slog.Logger
	libraryRepository LibraryRepository
}

func NewLibraryService(logger *slog.Logger, l LibraryRepository) *LibraryService {
	return &LibraryService{
		logger:            logger,
		libraryRepository: l,
	}
}

func (l *LibraryService) GetLibrary(limit int, page int, searchText string, dateFrom, dateTo time.Time) (int, []models.Song, error) {
	const op = "service.library.GetLibrary"
	offset := page * limit
	libraryDB, err := l.libraryRepository.GetLibrary(limit, offset, searchText, dateFrom, dateTo)
	if err != nil {
		return 0, nil, fmt.Errorf("%s: %w", op, err)
	}

	l.logger.Debug("library data from db", libraryDB)

	libraryMap := make(map[int]*models.Song)
	for _, row := range libraryDB {
		if _, exists := libraryMap[row.Id]; !exists {
			libraryMap[row.Id] = &models.Song{
				Id:          row.Id,
				Name:        row.Name,
				ReleaseDate: row.ReleaseDate,
				Link:        row.Link,
				Groups:      []models.Group{},
			}
		}
		libraryMap[row.Id].Groups = append(libraryMap[row.Id].Groups, models.Group{
			Id:   row.GroupId,
			Name: row.GroupName,
		})
	}

	// Преобразование карты в список
	var library []models.Song
	for _, music := range libraryMap {
		library = append(library, *music)
	}

	return len(library), library, nil
}
