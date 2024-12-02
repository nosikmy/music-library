package errors

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type MusicLibraryError struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"bad request error"`
}

func (m MusicLibraryError) Error() string {
	return m.Message
}

var (
	InternalError = MusicLibraryError{
		Status:  http.StatusInternalServerError,
		Message: "internal error",
	}
	BadRequestError = MusicLibraryError{
		Status:  http.StatusBadRequest,
		Message: "bad request error",
	}
)

func NewMusicLibraryError(merr MusicLibraryError, err error) error {
	return fmt.Errorf("%w: %s", merr, err.Error())
}

func GetHTTPErrorWithMessage(err error, message string) MusicLibraryError {
	var merr MusicLibraryError
	if errors.As(err, &merr) {
		merr.Message = merr.Message + ": " + message
		return merr
	}
	return MusicLibraryError{
		Message: err.Error() + ": " + message,
		Status:  520,
	}
}

func GetHTTPError(err error) MusicLibraryError {
	var merr MusicLibraryError
	if errors.As(err, &merr) {
		return merr
	}
	return MusicLibraryError{
		Message: err.Error(),
		Status:  520,
	}
}
