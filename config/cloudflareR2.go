package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/rs/zerolog/log"
)

func (cfg *Config) LoadAwsConfig() aws.Config {
	conf, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.R2.ApiKey, cfg.R2.ApiSecret, "",
			)), config.WithRegion("auto"))

	if err != nil {
		log.Fatal().Err(err).Msg("unable to load AWS Config")
	}

	log.Info().Msg("Success to load AWS Config")
	return conf
}
