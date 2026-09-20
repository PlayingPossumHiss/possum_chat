package github

type release struct {
	Version string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}
