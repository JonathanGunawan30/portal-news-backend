package model

import "time"

type Content struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CategoryID  int64     `gorm:"column:category_id;not null"`
	CreatedByID int64     `gorm:"column:created_by_id;not null"`
	Title       string    `gorm:"column:title;type:varchar(200);not null"`
	Excerpt     string    `gorm:"column:excerpt;type:varchar(250);not null"`
	Description string    `gorm:"column:description;type:text;not null"`
	Image       string    `gorm:"column:image;type:text"`
	Status      string    `gorm:"column:status;type:varchar(20);not null;default:PUBLISH"`
	Tags        string    `gorm:"column:tags;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`

	Category Category `gorm:"foreignKey:CategoryID;references:ID"`
	User     User     `gorm:"foreignKey:CreatedByID;references:ID"`
}
