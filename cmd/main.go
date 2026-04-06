package main

import (
	"context"
	"log"
	"os"

	"github.com/PlayingPossumHiss/possum_chat/internal/container"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

func main() {
	ctx := context.Background()
	app, err := container.New(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	err = app.Run()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
