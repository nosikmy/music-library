package handler

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/nosikmy/music-library/internal/app/errors"
	"github.com/nosikmy/music-library/internal/app/models"
	"github.com/sirupsen/logrus"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

// GetSongText Handler to get the verses for a certain song
//
//	@Summary		Get the verses for a certain song
//	@Description	Supports pagination(limit, page params)
//	@Tags			song
//	@Produce		json
//	@Param			id		path		int	true	"id of the chosen song"
//	@Param			limit	query		int	false	"limit of received data"				default(2)	example(2)
//	@Param			page	query		int	false	"page of data that you want to receive"	default(0)	example(1)
//	@Success		200		{object}	models.SongTextResponse
//	@Failure		400,500	{object}	errors.MusicLibraryError
//	@Router			/song/{id}/text [get]
func (h *Handler) GetSongText(ctx *gin.Context) {
	const op = "handler.song.GetSongText"
	idStr := ctx.Param("id")
	limitStr := ctx.Query("limit")
	if limitStr == "" {
		limitStr = "1"
	}
	pageStr := ctx.Query("page")
	if pageStr == "" {
		pageStr = "0"
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
		ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(mlErr, "id is not a number"))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
		ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(mlErr, "limit is not a number"))
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
		ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(mlErr, "page is not a number"))
		return
	}

	h.logger.Info("Getting song text", slog.Int("id", id))

	count, song, err := h.songService.GetSongText(id, limit, page)
	if err != nil {
		h.logger.Error("Error while getting song text " + op + ": " + err.Error())
		ctx.JSON(http.StatusInternalServerError, errors.GetHTTPError(
			errors.NewMusicLibraryError(errors.InternalError, err)),
		)
	}

	h.logger.Info("Got song text", slog.Int("id", id))

	ctx.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Payload: gin.H{
			"count": count,
			"text":  song,
		},
	})
}

// DeleteSong Handler to delete a certain song
//
//	@Summary	Delete a certain song
//	@Tags		song
//	@Produce	json
//	@Param		id		path		int	true	"id of the chosen song"
//	@Success	200		{object}	models.Response
//	@Failure	400,500	{object}	errors.MusicLibraryError
//	@Router		/song [delete]
func (h *Handler) DeleteSong(ctx *gin.Context) {
	const op = "handler.song.DeleteSong"
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
		ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(mlErr, "id is not a number"))
		return
	}

	h.logger.Info("Deleting song", slog.Int("id", id))

	err = h.songService.DeleteSong(id)
	if err != nil {
		h.logger.Error("Error while deleting text " + op + ": " + err.Error())
		ctx.JSON(http.StatusInternalServerError, errors.GetHTTPError(
			errors.NewMusicLibraryError(errors.InternalError, err)),
		)
		return
	}

	h.logger.Info("Song deleted", slog.Int("id", id))

	ctx.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Payload: nil,
	})
}

// ChangeSong Handler to change all fields of a song
//
//	@Summary		change all fields of a song
//	@Description	Fields will be changed if the required parameters for this are specified
//	@Tags			song
//	@Produce		json
//	@Param			id				path		int		true	"id of the chosen song"
//	@Param			name			query		string	false	"new name for song"
//	@Param			newGroup		query		string	false	"new group name to add to the song"
//	@Param			groupToDelete	query		string	false	"id of the group to be deleted from the song"
//	@Param			newVersePrevId	query		string	false	"verse id, after which a new verse should be inserted. id = 0 - for insertion at the beginning"
//	@Param			newVerseText	query		string	false	"text for a new verse"
//	@Param			verseId			query		string	false	"id of the verse whose text must be changed"
//	@Param			verseText		query		string	false	"new text for a verse"
//	@Param			deleteVerseId	query		string	false	"id of the verse to be deleted"
//	@Success		200				{object}	models.Response
//	@Failure		400,500			{object}	errors.MusicLibraryError
//	@Router			/song [put]
func (h *Handler) ChangeSong(ctx *gin.Context) {
	const op = "handler.song.ChangeSong"
	idStr := ctx.Param("id")
	newName := ctx.Query("name")
	newGroup := ctx.Query("newGroup")
	deleteGroupIdStr := ctx.Query("groupToDelete")
	newVerseIdPrevStr := ctx.Query("newVersePrevId")
	newVerseText := ctx.Query("newVerseText")
	changeVerseIdStr := ctx.Query("verseId")
	changeVerseText := ctx.Query("verseText")
	deleteVerseIdStr := ctx.Query("deleteVerseId")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return
	}

	var deleteGroupId int
	if deleteGroupIdStr != "" {
		deleteGroupId, err = strconv.Atoi(deleteGroupIdStr)
		if err != nil {
			mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
			ctx.JSON(
				http.StatusBadRequest, errors.GetHTTPErrorWithMessage(
					mlErr, "id for deleting group is not a number"))
			return
		}
	}
	var deleteVerseId int
	if deleteVerseIdStr != "" {
		deleteVerseId, err = strconv.Atoi(deleteVerseIdStr)
		if err != nil {
			mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
			ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(
				mlErr, "id for deleting verse is not a number"))
			return
		}
	}

	var newVerse *models.Verse
	if newVerseIdPrevStr != "" {
		newVerseIdPrev, err := strconv.Atoi(newVerseIdPrevStr)
		if err != nil {
			mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
			ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(
				mlErr, "id of previous verse is not a number"))
			return
		}
		newVerse = &models.Verse{
			Id:   newVerseIdPrev,
			Text: newVerseText,
		}
	}

	var changeVerse *models.Verse
	if changeVerseIdStr != "" {
		changeVerseId, err := strconv.Atoi(changeVerseIdStr)
		if err != nil {
			mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
			ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(
				mlErr, "id of verse for changing is not a number"))
			return
		}
		changeVerse = &models.Verse{
			Id:   changeVerseId,
			Text: changeVerseText,
		}
	}

	h.logger.Info("Changing song", slog.Int("id", id))

	err = h.songService.ChangeSong(id, newName, newGroup, deleteGroupId, newVerse, changeVerse, deleteVerseId)
	if err != nil {
		h.logger.Error("Error while changing song " + op + ": " + err.Error())
		ctx.JSON(http.StatusInternalServerError, errors.GetHTTPError(
			errors.NewMusicLibraryError(errors.InternalError, err)),
		)
		return
	}

	h.logger.Info("Song changed", slog.Int("id", id))

	ctx.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Payload: nil,
	})
}

// AddSong Handler to add a song to the library
//
//	@Summary	Add a song to the library
//	@Tags		song
//	@Produce	json
//	@Param		input	body		models.ApiMusicRequest	true	"Data for adding a song"
//	@Success	200		{object}	models.AddSongResponse
//	@Failure	400,500	{object}	errors.MusicLibraryError
//	@Router		/song [post]
func (h *Handler) AddSong(ctx *gin.Context) {
	const op = "handler.song.AddSong"
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("Error while loading .env file: %s", err.Error())
	}
	//group := ctx.Query("group")
	//name := ctx.Query("song")
	var input models.ApiMusicRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
		ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(mlErr, "bad format of input data"))
		return
	}

	client := resty.New()

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := client.R().SetContext(ctxTimeout).SetQueryParams(map[string]string{
		"group": input.Group,
		"song":  input.Song,
	})
	resp, err := request.Get(os.Getenv("API_MUSIC_ADDRESS") + "/info")
	if err != nil {
		ctx.JSON(520, errors.GetHTTPErrorWithMessage(err, ""))
		return
	}
	h.logger.Info("Adding new song", slog.String("group", input.Group), slog.String("song", input.Song))
	musicData := models.ApiMusicResponse{}
	err = json.Unmarshal(resp.Body(), &musicData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errors.GetHTTPError(
			errors.NewMusicLibraryError(errors.InternalError, err)),
		)
		return
	}

	h.logger.Debug("Data from music api", slog.Any("data", musicData))

	id, err := h.songService.AddSong(input.Group, input.Song, musicData)
	if err != nil {
		h.logger.Error("Error while adding new song " + op + ": " + err.Error())
		ctx.JSON(http.StatusInternalServerError, errors.GetHTTPError(err))
		return
	}

	h.logger.Info("New song added", slog.String("group", input.Group), slog.String("song", input.Song))

	ctx.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Payload: gin.H{
			"id": id,
		},
	})
}
