package api

import (
	"net/http"

	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/gin-gonic/gin"
)

func (a *Api) cssMainStyleCss(ctx *gin.Context) {
	ctx.Writer.Header().Set("Content-Type", "text/css")
	ctx.Writer.Header().Set("Cache-Control", "max-age=5")
	content, err := a.getStyleUC.GetMainStyle()
	if err != nil {
		logger.Error(err)
		ctx.String(http.StatusInternalServerError, "")

		return
	}
	ctx.String(http.StatusOK, content)
}
