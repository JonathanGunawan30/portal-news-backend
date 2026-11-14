package seeds

import (
	"os"
	"strings"

	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/model"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedRoles(db *gorm.DB) {
	adminEmail := strings.TrimSpace(viper.GetString("ADMIN_EMAIL"))
	if adminEmail == "" {
		adminEmail = strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	}

	adminPassword := viper.GetString("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = os.Getenv("ADMIN_PASSWORD")
	}

	if adminEmail == "" || adminPassword == "" {
		log.Info().Msg("Skipping admin user seed: ADMIN_EMAIL or ADMIN_PASSWORD is not configured")
		return
	}

	if len(adminPassword) < 8 || strings.EqualFold(adminPassword, "rahasia") {
		log.Warn().Msg("Skipping admin user seed: ADMIN_PASSWORD must be at least 8 characters and not a weak default")
		return
	}

	bytes, err := conv.HashPassword(adminPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating password hash")
	}

	admin := model.User{
		Name:     "Admin",
		Email:    adminEmail,
		Password: bytes,
	}

	if err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoNothing: true,
	}).Create(&admin).Error; err != nil {
		log.Fatal().Err(err).Msg("Failed to seed admin user")
	}

	log.Info().Msg("Admin user seeded (created or already exists)")
}
