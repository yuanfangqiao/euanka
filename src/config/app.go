package config

type App struct {
	Env     string `mapstructure:"env" json:"env" yaml:"env"`
	Port    int    `mapstructure:"port" json:"port" yaml:"port"`
	AppName string `mapstructure:"app_name" json:"app_name" yaml:"app_name"`
	AppUrl  string `mapstructure:"app_url" json:"app_url" yaml:"app_url"`
	DbType  string `mapstructure:"db_type" json:"db_type" yaml:"db_type"`
	ExternalTtsHostPort string `mapstructure:"external_tts_host_port" json:"external_tts_host_port" yaml:"external_tts_host_port"`
}