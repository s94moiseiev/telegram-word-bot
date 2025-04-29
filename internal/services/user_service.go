package services

import (
	"english-words-bot/internal/db"
	"english-words-bot/internal/models"
)

type UserService struct{}

func (s *UserService) GetOrCreateUser(telegramID int64, username string) (*models.User, error) {
	var user models.User
	result := db.DB.Where("telegram_id = ?", telegramID).First(&user)
	
	if result.Error != nil {
		user = models.User{
			TelegramID: telegramID,
			Username:   username,
		}
		if err := db.DB.Create(&user).Error; err != nil {
			return nil, err
		}
	}
	
	return &user, nil
} 