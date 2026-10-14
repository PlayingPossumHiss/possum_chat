package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1CommandStartElection(ctx *gin.Context) {
	request := apiV1CommandStartElectionRequest{}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, responseError{Message: err.Error()})

		return
	}

	if len(request.Candidates) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, responseError{Message: "must be more than 0 variants"})

		return
	}

	a.voter.StartElection(request.Candidates)

	ctx.JSON(http.StatusOK, nil)
}
