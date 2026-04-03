package api

import (
	"fmt"

	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/gin-gonic/gin"
)

type Api struct {
	service        *gin.Engine
	port           int
	getStyleUC     GetStyleUC
	listMessagesUC ListMessagesUC
}

func New(
	port int,
	getStyleUC GetStyleUC,
	listMessagesUC ListMessagesUC,
) *Api {
	gin.SetMode(gin.ReleaseMode)
	service := gin.New()
	service.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/api/v1/messages"),
		gin.Recovery(),
	)
	api := &Api{
		service:        service,
		getStyleUC:     getStyleUC,
		listMessagesUC: listMessagesUC,
		port:           port,
	}

	service.GET("/css/custom_style.css", api.cssCustomStyleCss)
	service.GET("/api/v1/messages", api.apiV1Messages)
	service.StaticFile("js/messages.js", "./static/js/messages.js")
	service.StaticFile("css/messages.css", "./static/css/messages.css")
	service.StaticFile("messages.html", "./static/messages.html")
	service.Static("/img", "./static/img")

	return api
}

func (a *Api) Run() error {
	logger.Warn("starting self api")
	return a.service.Run(fmt.Sprintf(":%d", a.port))
}
