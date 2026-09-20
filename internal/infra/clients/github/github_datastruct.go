package github

type release struct {
	Version string   `json:"tag_name"`
	Asserts []assert `json:"assets"`
}

type assert struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}
