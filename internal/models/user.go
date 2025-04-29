package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID         uint  `gorm:"primarykey"`
	TelegramID int64 `gorm:"uniqueIndex"`
	Username   string
	Words      []Word `gorm:"foreignKey:UserID"`
}
