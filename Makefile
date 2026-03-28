install-deps:
	GOBIN=${CURDIR}/bin go install github.com/gojuno/minimock/v3/cmd/minimock@latest

mock-generate:
	${CURDIR}/bin/minimock -g -i ./internal/service/message_queue.ConfigStorage -o ./internal/service/message_queue/mocks -s _mock.go
	${CURDIR}/bin/minimock -g -i ./internal/utils/time.Clock -o ./internal/utils/time/mocks -s _mock.go
	${CURDIR}/bin/minimock -g -i ./internal/service/scrapers/twitch.TwitchIrcClient -o ./internal/service/scrapers/twitch/mocks -s _mock.go
	${CURDIR}/bin/minimock -g -i ./internal/service/scrapers/vk_play_live.VkPlayLiveApi -o ./internal/service/scrapers/vk_play_live/mocks -s _mock.go
	${CURDIR}/bin/minimock -g -i ./internal/service/scrapers/vk_play_live.VkPlayLiveWs -o ./internal/service/scrapers/vk_play_live/mocks -s _mock.go
	${CURDIR}/bin/minimock -g -i ./internal/service/scrapers/youtube.YoutubeClient -o ./internal/service/scrapers/youtube/mocks -s _mock.go

cover:
	go test -timeout 240s -short -count=1 -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

test:
	go test ./...

lint:
	golangci-lint run

