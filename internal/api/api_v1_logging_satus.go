package api

import (
	"net/http"

	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1LoggingStatus(ctx *gin.Context) {
	status := logger.GetStatus()
	resp := apiV1LoggingStatusResponse{
		ErrorCount: status.ErrorCount,
		WarnCount:  status.WarnCount,
	}

	ctx.JSON(http.StatusOK, resp)
}
