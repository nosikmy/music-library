package services

import (
	"fmt"
	"github.com/nosikmy/music-library/internal/app/models"
	"log/slog"
	"strings"
	"time"
)

type SongRepository interface {
	GetSongText(id int, limit int, offset int) (int, []models.Verse, error)
	DeleteSong(id int) error
	AddSong(group string, song string, releaseDate time.Time, verses []string, link string) (int, error)
}

type SongChangerRepository interface {
	ChangeSongName(id int, newName string) error
	AddGroupToSong(id int, group string) error
	DeleteGroupFromSong(id int, groupId int) error
}

type VersesRepository interface {
	AddVerse(id int, newVerse *models.Verse) error
	ChangeVerse(changeVerse *models.Verse) error
	DeleteVerse(id int, verseId int) error
}

type SongService struct {
	logger                *slog.Logger
	songRepository        SongRepository
	songChangerRepository SongChangerRepository
	versesRepository      VersesRepository
}

func NewSongService(logger *slog.Logger, s SongRepository, sc SongChangerRepository, v VersesRepository) *SongService {
	return &SongService{
		logger:                logger,
		songRepository:        s,
		songChangerRepository: sc,
		versesRepository:      v,
	}
}

func (s *SongService) GetSongText(id int, limit int, page int) (int, []models.Verse, error) {
	const op = "service.song.GetSongText"
	offset := limit * page
	count, song, err := s.songRepository.GetSongText(id, limit, offset)
	if err != nil {
		return 0, nil, fmt.Errorf("%s: %w", op, err)
	}

	return count, song, nil
}

func (s *SongService) DeleteSong(id int) error {
	const op = "service.song.DeleteSong"
	err := s.songRepository.DeleteSong(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *SongService) ChangeSong(id int, changeName string, newGroup string, deleteGroupId int,
	newVerse *models.Verse, changeVerse *models.Verse, deleteVerseId int) error {
	const op = "service.song.ChangeSong"
	if changeName != "" {
		err := s.songChangerRepository.ChangeSongName(id, changeName)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.logger.Info("Song name updated", slog.Int("songId", id), slog.String("newName", changeName))
	}
	if newGroup != "" {
		err := s.songChangerRepository.AddGroupToSong(id, newGroup)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.logger.Info("Added new group to song", slog.Int("songId", id), slog.String("newGroup", newGroup))
	}
	if deleteGroupId != 0 {
		err := s.songChangerRepository.DeleteGroupFromSong(id, deleteGroupId)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.logger.Info("Song changed", slog.Int("songId", id), slog.Int("deletedGroupId", deleteGroupId))
	}
	if newVerse != nil {
		err := s.versesRepository.AddVerse(id, newVerse)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.logger.Info("Added new verse to song", slog.Int("songId", id))
	}
	if changeVerse != nil {
		err := s.versesRepository.ChangeVerse(changeVerse)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.logger.Info("Changed the verse of the song", slog.Int("songId", id))
	}
	if deleteVerseId != 0 {
		err := s.versesRepository.DeleteVerse(id, deleteVerseId)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.logger.Info("Deleted verse from the song", slog.Int("songId", id))
	}
	return nil
}

func (s *SongService) AddSong(group string, song string, songData models.ApiMusicResponse) (int, error) {
	const op = "service.song.AddSong"
	verses := strings.Split(songData.Text, "\n\n")
	releaseDate, err := time.Parse("02.01.2006", songData.ReleaseDate)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.songRepository.AddSong(group, song, releaseDate, verses, songData.Link)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}
