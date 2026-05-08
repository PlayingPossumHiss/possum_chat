package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) cssCustomStyleCss(ctx *gin.Context) {
	ctx.Writer.Header().Set("Content-Type", "text/css")
	ctx.String(http.StatusOK, a.getStyleUC.GetStyle())
}
