package main

import (
	"context"
	"log"

	"github.com/PlayingPossumHiss/possum_chat/internal/container"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	app := container.New(ctx)
	err := app.Run()
	if err != nil {
		cancel()
		log.Fatalln(err)
	}
}
