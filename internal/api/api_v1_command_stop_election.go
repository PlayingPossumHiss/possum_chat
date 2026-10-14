package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1CommandStopElection(ctx *gin.Context) {
	a.voter.StopElection()

	ctx.JSON(http.StatusOK, nil)
}
