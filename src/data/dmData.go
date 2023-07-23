package data

type DmData struct {
	Topic string `json:"topic"`
	DM    DmItem `json:"dm"`
}

type DmItem struct {
	Nlg         string `json:"nlg"`
	AudioBase64 string `json:"audioBase64"`
	AudioUrl    string `json:"audioUrl"`
}
