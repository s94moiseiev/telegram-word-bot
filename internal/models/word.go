package models

import (
	"gorm.io/gorm"
)

type Word struct {
	gorm.Model
	UserID       uint
	EnglishWord  string
	Translation  string
	User         User `gorm:"foreignKey:UserID"`
} 