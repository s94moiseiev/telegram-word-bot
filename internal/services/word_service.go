package services

import (
	"english-words-bot/internal/db"
	"english-words-bot/internal/models"
)

type WordService struct{}

func (s *WordService) AddWord(userID uint, englishWord, translation string) error {
	word := models.Word{
		UserID:      userID,
		EnglishWord: englishWord,
		Translation: translation,
	}
	return db.DB.Create(&word).Error
}

func (s *WordService) GetUserWords(userID uint) ([]models.Word, error) {
	var words []models.Word
	err := db.DB.Where("user_id = ?", userID).Find(&words).Error
	return words, err
}

func (s *WordService) UpdateWord(wordID uint, englishWord, translation string) error {
	return db.DB.Model(&models.Word{}).Where("id = ?", wordID).
		Updates(map[string]interface{}{
			"english_word": englishWord,
			"translation":  translation,
		}).Error
}

func (s *WordService) DeleteWord(wordID uint) error {
	return db.DB.Delete(&models.Word{}, wordID).Error
}

func (s *WordService) GetRandomWord(userID uint) (*models.Word, error) {
	var word models.Word
	err := db.DB.Where("user_id = ?", userID).Order("RANDOM()").First(&word).Error
	return &word, err
}

func (s *WordService) GetWordByID(wordID uint) (*models.Word, error) {
	var word models.Word
	err := db.DB.Where("id = ?", wordID).First(&word).Error
	return &word, err
}
