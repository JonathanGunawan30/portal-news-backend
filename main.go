package main

import (
	"github.com/JonathanGunawan30/portal-news-backend.git/cmd"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Info().Msg("No .env file found, using environment variables from the system")
	}
	cmd.Execute()
}
