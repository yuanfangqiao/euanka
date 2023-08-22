package config

type Service struct {
	Asr  string `mapstructure:"asr" json:"service" yaml:"asr"`
	Rasa string `mapstructure:"rasa" json:"rasa" yaml:"rasa"`
	Tts  string `mapstructure:"tts" json:"tts" yaml:"tts"`
	Llm  string `mapstructure:"llm" json:"llm" yaml:"llm"`
}
