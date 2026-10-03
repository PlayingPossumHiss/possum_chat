package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1Command(ctx *gin.Context) {
	request := apiV1CommandRequest{}
	ctx.ShouldBindBodyWithJSON(&request)

	err := a.doCommandUC.DoCommand(ctx, request.Command)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, apiV1CommandResponse{})
}
