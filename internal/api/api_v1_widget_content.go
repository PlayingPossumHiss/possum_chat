package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1WidgetContent(ctx *gin.Context) {
	electionResult := a.voter.ElectionResult()
	resp := apiV1WidgetContentResponse{
		Vote: make([]vote, 0, len(electionResult)),
	}
	for _, candidate := range electionResult {
		resp.Vote = append(resp.Vote, vote{
			Text:  candidate.Text,
			Count: candidate.Counter,
		})
	}

	ctx.JSON(http.StatusOK, resp)
}
