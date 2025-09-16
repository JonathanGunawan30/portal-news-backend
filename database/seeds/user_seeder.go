package seeds

import (
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/model"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	bytes, err := conv.HashPassword("rahasia")
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating password hash")
	}

	admin := model.User{
		Name:     "Admin",
		Email:    "admin@mail.com",
		Password: bytes,
	}

	result := db.FirstOrCreate(&admin, model.User{Email: admin.Email})

	if result.Error != nil {
		log.Fatal().Err(result.Error).Msg("Failed to seed admin user")
		return
	}

	if result.RowsAffected > 0 {
		log.Info().Msg("Admin user seeded successfully")
	} else {
		log.Info().Msg("Admin user already exists, skipping seed.")
	}

}
