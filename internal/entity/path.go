package entity

import (
	"time"
)

type Path struct {
	UserID    uint      `gorm:"primary_key;not null"`
	Path      string    `gorm:"primary_key;size:255;not null"`
	Deploy    bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	User      User      `gorm:"foreignkey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UpdatedAt time.Time `gorm:"not null"`
}
