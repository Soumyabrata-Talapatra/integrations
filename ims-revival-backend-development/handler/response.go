package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleSuccess sends a success response with the specified status code and optional data
func handleSuccess(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, data)
}
func handleCreateSuccess(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusCreated, data)
}
