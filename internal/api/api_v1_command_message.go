package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1CommandMessage(ctx *gin.Context) {
	request := apiV1CommandMessageRequest{}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, responseError{Message: err.Error()})

		return
	}

	var command string
	if request.Text == "" {
		command = "--message"
	} else {
		command = fmt.Sprintf("--message %s", request.Text)
	}

	err := a.doCommandUC.DoCommand(ctx, command)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, responseError{Message: err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, nil)
}
