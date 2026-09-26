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
	onlineGetter   OnlineGetter
	texter         Texter
	voter          Voter
}

func New(
	port int,
	getStyleUC GetStyleUC,
	listMessagesUC ListMessagesUC,
	onlineGetter OnlineGetter,
	texter Texter,
	voter Voter,
) *Api {
	gin.SetMode(gin.ReleaseMode)
	service := gin.New()
	service.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/api/v1/messages", "/api/v1/logging_status", "/api/v1/widget_content"),
		gin.Recovery(),
	)
	api := &Api{
		service:        service,
		getStyleUC:     getStyleUC,
		listMessagesUC: listMessagesUC,
		onlineGetter:   onlineGetter,
		texter:         texter,
		voter:          voter,
		port:           port,
	}

	service.GET("/css/custom_style.css", api.cssCustomStyleCss)
	service.GET("/api/v1/messages", api.apiV1Messages)
	service.GET("/api/v1/logging_status", api.apiV1LoggingStatus)
	service.GET("/api/v1/widget_content", api.apiV1WidgetContent)
	service.GET("css/messages.css", api.cssMainStyleCss)

	service.StaticFile("js/messages.js", "./static/js/messages.js")
	service.StaticFile("messages.html", "./static/messages.html")
	service.StaticFile("js/widget.js", "./static/js/widget.js")
	service.StaticFile("css/widget.css", "./static/css/widget.css")
	service.StaticFile("widget.html", "./static/widget.html")
	service.Static("/img", "./static/img")

	return api
}

func (a *Api) Run() {
	go func() {
		logger.Info("starting self api")
		// Если что - этот вызов блокирующий
		err := a.service.Run(fmt.Sprintf(":%d", a.port))
		if err != nil {
			logger.Error(err)
		}
	}()
}
