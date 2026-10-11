package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1CommandVote(ctx *gin.Context) {
	request := apiV1CommandVoteRequest{}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, responseError{Message: err.Error()})

		return
	}

	if len(request.Candidates) == 0 {
		a.voter.StopElection()
	} else {
		a.voter.StartElection(request.Candidates)
	}

	ctx.JSON(http.StatusOK, nil)
}
