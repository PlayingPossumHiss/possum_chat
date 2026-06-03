package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) cssCustomStyleCss(ctx *gin.Context) {
	ctx.Writer.Header().Set("Content-Type", "text/css")
	ctx.Writer.Header().Set("Cache-Control", "max-age=5")
	ctx.String(http.StatusOK, a.getStyleUC.GetCustomStyle())
}
