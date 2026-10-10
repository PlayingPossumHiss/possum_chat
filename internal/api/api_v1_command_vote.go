package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1CommandVote(ctx *gin.Context) {
	request := apiV1CommandVoteRequest{}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, responseError{Message: err.Error()})

		return
	}

	var command string
	if len(request.Candidates) == 0 {
		command = "--vote"
	} else {
		command = fmt.Sprintf("--vote %s", strings.Join(request.Candidates, ";"))
	}

	err := a.doCommandUC.DoCommand(ctx, command)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, responseError{Message: err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, nil)
}
