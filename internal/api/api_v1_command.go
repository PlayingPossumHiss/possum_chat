package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1Command(ctx *gin.Context) {
	request := apiV1CommandRequest{}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apiV1CommandError{Message: err.Error()})

		return
	}

	if request.Command == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, apiV1CommandError{Message: "empty command"})

		return
	}

	err := a.doCommandUC.DoCommand(ctx, request.Command)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, apiV1CommandError{Message: err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, apiV1CommandResponse{})
}
