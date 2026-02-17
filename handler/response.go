package handler

import (
	"net/http"
	"scratch/core/port"

	"github.com/gin-gonic/gin"
)

type response struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"  example:"success"`
	Data    any    `json:"data,omitempty"`
}

func newResponse(success bool, message string, data any) response {
	return response{
		Success: success,
		Message: message,
		Data:    data,
	}
}

var errorStatusMap = map[error]int{
	port.ErrDataNotFound:               http.StatusNotFound,
	port.ErrConflictingData:            http.StatusConflict,
	port.ErrInvalidCredentials:         http.StatusUnauthorized,
	port.ErrUnauthorized:               http.StatusUnauthorized,
	port.ErrEmptyAuthorizationHeader:   http.StatusUnauthorized,
	port.ErrInvalidAuthorizationHeader: http.StatusUnauthorized,
	port.ErrInvalidAuthorizationType:   http.StatusUnauthorized,
	port.ErrInvalidToken:               http.StatusUnauthorized,
	port.ErrExpiredToken:               http.StatusUnauthorized,
	port.ErrForbidden:                  http.StatusForbidden,
	port.ErrNoUpdatedData:              http.StatusBadRequest,
	port.ErrInsufficientStock:          http.StatusBadRequest,
	port.ErrInsufficientPayment:        http.StatusBadRequest,
}

func handleSuccess(ctx *gin.Context, data any) {
	rsp := newResponse(true, "success", data)
	ctx.JSON(http.StatusOK, rsp)
}

func handleFailure(c *gin.Context, data any) {
	rsp := newResponse(false, "Failure", data)
	c.JSON(http.StatusNotFound, rsp)
}
