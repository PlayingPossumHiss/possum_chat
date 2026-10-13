package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1CommandMessage(ctx *gin.Context) {
	request := apiV1CommandMessageRequest{}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, responseError{Message: err.Error()})

		return
	}

	a.screenTitle.SetMessage(request.Text)

	ctx.JSON(http.StatusOK, nil)
}
