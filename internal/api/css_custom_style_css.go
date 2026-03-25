package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) cssCustomStyleCss(ctx *gin.Context) {
	ctx.String(http.StatusOK, a.getStyleUC.GetStyle())
}
