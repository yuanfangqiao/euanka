package config

type Configuration struct {
	App App `mapstructure:"app" json:"app" yaml:"app"`
	Service Service `mapstructure:"service" json:"service" yaml:"service"`
	Zap Zap `mapstructure:"zap" json:"zap" yaml:"zap"`
}