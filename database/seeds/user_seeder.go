package seeds

import (
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/domain/model"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/conv"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	if err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoNothing: true,
	}).Create(&admin).Error; err != nil {
		log.Fatal().Err(err).Msg("Failed to seed admin user")
	} else {
		log.Info().Msg("Admin user seeded (created or already exists)")
	}

}
