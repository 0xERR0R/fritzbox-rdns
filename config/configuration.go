package config

import (
	"context"
	"github.com/sethvargo/go-envconfig"
)

type FritzboxConfig struct {
	Url      string `env:"FB_URL,required"`
	User     string `env:"FB_USER,required"`
	Password string `env:"FB_PASSWORD,required"`
}

func LoadConfig() (FritzboxConfig, error) {
	ctx := context.Background()

	var c FritzboxConfig
	return c, envconfig.Process(ctx, &c)
}
