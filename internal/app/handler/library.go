package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nosikmy/music-library/internal/app/errors"
	"github.com/nosikmy/music-library/internal/app/models"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// GetLibrary Handler to get a list of songs
//
//	@Summary		Get a list of songs
//	@Description	Supports pagination(limit, page params)
//	@Description	Supports filtration(search, dateFrom, dateTo params)
//	@Tags			library
//	@Produce		json
//	@Param			limit		query		int		false	"limit of received data"				default(10)	example(10)
//	@Param			page		query		int		false	"page of data that you want to receive"	default(0)	example(2)
//	@Param			search		query		string	false	"search query for filtering by song and group names"
//	@Param			dateFrom	query		string	false	"the date from which the release dates of the songs begin"
//	@Param			dateTo		query		string	false	"the date from which the release dates of the songs end"
//	@Success		200			{object}	models.LibraryResponse
//	@Failure		400,500		{object}	errors.MusicLibraryError
//	@Router			/library [get]
func (h *Handler) GetLibrary(ctx *gin.Context) {
	const op = "handler.library.GetLibrary"
	limitStr := ctx.Query("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	pageStr := ctx.Query("page")
	if pageStr == "" {
		pageStr = "0"
	}
	search := ctx.Query("search")
	dateFrom := ctx.Query("dateFrom")
	dateTo := ctx.Query("dateTo")

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

	h.logger.Info("Getting library")

	var dateFromTime time.Time
	if dateFrom == "" {
		dateFromTime = time.Time{}
	} else {
		dateFromTime, err = time.Parse("01.02.2006", dateFrom)
		if err != nil {
			mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
			ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(mlErr, "bad date format"))
			return
		}
	}
	var dateToTime time.Time
	if dateTo == "" {
		dateToTime = time.Time{}
	} else {
		dateToTime, err = time.Parse("01.02.2006", dateTo)
		if err != nil {
			mlErr := errors.NewMusicLibraryError(errors.BadRequestError, err)
			ctx.JSON(http.StatusBadRequest, errors.GetHTTPErrorWithMessage(mlErr, "bad date format"))
			return
		}
	}
	count, library, err := h.libraryService.GetLibrary(limit, page, search, dateFromTime, dateToTime)
	if err != nil {
		h.logger.Error("Error while getting library " + op + ": " + err.Error())
		ctx.JSON(http.StatusInternalServerError, errors.GetHTTPError(
			errors.NewMusicLibraryError(errors.InternalError, err)),
		)
		return
	}

	h.logger.Info("Got library", slog.Int("rowsCount", count))

	ctx.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "ok",
		Payload: gin.H{
			"count":   count,
			"library": library,
		},
	})
}
