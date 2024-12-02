package handler

import (
	"github.com/gin-gonic/gin"
	_ "github.com/nosikmy/music-library/docs"
	"github.com/nosikmy/music-library/internal/app/models"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log/slog"
	"net/http"
	"time"
)

type LibraryService interface {
	GetLibrary(limit int, page int,
		searchText string, dateFrom, dateTo time.Time) (int, []models.Song, error)
}

type SongService interface {
	GetSongText(id int, limit int, page int) (int, []models.Verse, error)
	DeleteSong(id int) error
	ChangeSong(id int, changeName string, newGroup string, deleteGroupId int,
		newVerse *models.Verse, changeVerse *models.Verse, deleteVerseId int) error
	AddSong(group string, song string, songData models.ApiMusicResponse) (int, error)
}

type Handler struct {
	logger         *slog.Logger
	libraryService LibraryService
	songService    SongService
}

func NewHandler(logger *slog.Logger, l LibraryService, s SongService) *Handler {
	return &Handler{
		logger:         logger,
		libraryService: l,
		songService:    s,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/swagger", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store")
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	router.GET("/library", h.GetLibrary)
	songRouter := router.Group("/song")
	{
		songRouter.POST("", h.AddSong)
		songRouterId := songRouter.Group("/:id")
		{
			songRouterId.GET("/text", h.GetSongText)
			songRouterId.DELETE("", h.DeleteSong)
			songRouterId.PUT("", h.ChangeSong)
		}

	}

	return router
}
