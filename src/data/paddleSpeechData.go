package data

type PaddlespeechData struct {
	Result PaddleSpeechResult `json:"result"`
}

type PaddleSpeechResult struct {
	Audio string `json:"audio"`
}
