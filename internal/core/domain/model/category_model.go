package model

import "time"

type Category struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CreatedByID int64     `gorm:"column:created_by_id;not null"`
	Title       string    `gorm:"column:title;type:varchar(200);not null"`
	Slug        string    `gorm:"column:slug;type:varchar(200);unique;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`

	User User `gorm:"foreignKey:CreatedByID;references:ID"`
}
