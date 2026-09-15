package twitch_client

// GQL endpoint и web client-id, которые использует сам сайт twitch.tv
const (
	gqlURL = "https://gql.twitch.tv/gql"

	channelURL = "https://www.twitch.tv/%s"

	getOnlineQuery = `query GetOnline($login: String!) {
		user(login: $login) {
			stream {
				viewersCount
			}
		}
	}`
)

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type streamResponse struct {
	Data struct {
		User struct {
			Stream *stream `json:"stream"`
		} `json:"user"`
	} `json:"data"`
}

type stream struct {
	ViewersCount int64 `json:"viewersCount"`
}
