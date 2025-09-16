package model

import "time"

type User struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;type:varchar(100);not null"`
	Email     string    `gorm:"column:email;type:varchar(100);unique;not null"`
	Password  string    `gorm:"column:password;type:varchar(100);not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
}
