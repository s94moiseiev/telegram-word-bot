package bot

import (
	"english-words-bot/internal/services"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"english-words-bot/internal/version"

	tele "gopkg.in/telebot.v3"
)

type trainingStats struct {
	correct      int
	incorrect    int
	words        []uint
	currentIndex int
}

type Bot struct {
	bot           *tele.Bot
	userService   *services.UserService
	wordService   *services.WordService
	userStates    map[int64]string
	trainingWords map[int64]uint
	trainingStats map[int64]trainingStats
}

func NewBot(token string) (*Bot, error) {
	pref := tele.Settings{
		Token:     token,
		Poller:    &tele.LongPoller{Timeout: 10},
		ParseMode: tele.ModeHTML,
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	// Ініціалізуємо генератор випадкових чисел
	rand.Seed(time.Now().UnixNano())

	bot := &Bot{
		bot:           b,
		userService:   &services.UserService{},
		wordService:   &services.WordService{},
		userStates:    make(map[int64]string),
		trainingWords: make(map[int64]uint),
		trainingStats: make(map[int64]trainingStats),
	}

	bot.setupHandlers()
	return bot, nil
}

func (b *Bot) setupHandlers() {
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/menu", b.handleMenu)
	b.bot.Handle("/stop", b.handleStop)
	b.bot.Handle(tele.OnText, b.handleText)

	// Додаємо обробники для кнопок тренування
	b.bot.Handle(&tele.Btn{Text: "🎯 Training"}, func(c tele.Context) error {
		return c.Send("Choose training mode:", b.getTrainingMenu())
	})

	b.bot.Handle(&tele.Btn{Text: "🎯 10 Words Training"}, func(c tele.Context) error {
		err := b.startTraining(c.Sender().ID, "10_words")
		if err != nil {
			return c.Send(err.Error())
		}
		word, _ := b.wordService.GetWordByID(b.trainingWords[c.Sender().ID])
		return c.Send(fmt.Sprintf("Translate this word: %s", word.EnglishWord))
	})

	b.bot.Handle(&tele.Btn{Text: "🎯 Continuous Training"}, func(c tele.Context) error {
		err := b.startTraining(c.Sender().ID, "continuous")
		if err != nil {
			return c.Send(err.Error())
		}
		word, _ := b.wordService.GetWordByID(b.trainingWords[c.Sender().ID])
		return c.Send(fmt.Sprintf("Translate this word: %s\nType /stop to end training", word.EnglishWord))
	})

	b.bot.Handle(&tele.Btn{Text: "🔙 Back to Main Menu"}, func(c tele.Context) error {
		delete(b.userStates, c.Sender().ID)
		delete(b.trainingWords, c.Sender().ID)
		delete(b.trainingStats, c.Sender().ID)
		return c.Send("Choose an option:", b.getMainMenu())
	})
}

func (b *Bot) Start() {
	b.bot.Start()
}

func (b *Bot) getMainMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{
		ResizeKeyboard: true,
	}

	menu.Reply(
		menu.Row(
			tele.Btn{Text: "➕ Add Word"},
			tele.Btn{Text: "📚 My Words"},
			tele.Btn{Text: "✏️ Edit Word"},
		),
		menu.Row(
			tele.Btn{Text: "🗑 Delete Word"},
			tele.Btn{Text: "🎯 Training"},
		),
	)

	return menu
}

func (b *Bot) getTrainingMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{
		ResizeKeyboard: true,
	}

	menu.Reply(
		menu.Row(
			tele.Btn{Text: "🎯 10 Words Training"},
			tele.Btn{Text: "🎯 Continuous Training"},
			tele.Btn{Text: "🔙 Back to Main Menu"},
		),
	)

	return menu
}

func (b *Bot) handleStart(c tele.Context) error {
	user, err := b.userService.GetOrCreateUser(c.Sender().ID, c.Sender().Username)
	if err != nil {
		return c.Send("Error creating user profile")
	}

	response := fmt.Sprintf("Welcome, %s! I'm your English words learning bot.\n\n"+
		"Version: %s\n"+
		"Use the menu below to start learning!",
		user.Username,
		version.Version)

	return c.Send(response, b.getMainMenu(), &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	})
}

func (b *Bot) handleMenu(c tele.Context) error {
	return c.Send("Choose an option:", b.getMainMenu())
}

func (b *Bot) handleStop(c tele.Context) error {
	userID := c.Sender().ID
	if stats, ok := b.trainingStats[userID]; ok {
		total := stats.correct + stats.incorrect
		accuracy := 0.0
		if total > 0 {
			accuracy = float64(stats.correct) / float64(total) * 100
		}

		response := fmt.Sprintf("Training stopped!\nResults:\nCorrect: %d\nIncorrect: %d\nAccuracy: %.1f%%",
			stats.correct, stats.incorrect, accuracy)

		delete(b.userStates, userID)
		delete(b.trainingWords, userID)
		delete(b.trainingStats, userID)

		return c.Send(response, b.getMainMenu(), &tele.SendOptions{
			ParseMode: tele.ModeHTML,
		})
	}
	return c.Send("No active training session", b.getMainMenu(), &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	})
}

func (b *Bot) startTraining(userID int64, mode string) error {
	user, _ := b.userService.GetOrCreateUser(userID, "")
	words, err := b.wordService.GetUserWords(user.ID)
	if err != nil {
		return err
	}

	if len(words) == 0 {
		return fmt.Errorf("no words available for training")
	}

	// Очищаємо попередні результати
	delete(b.trainingWords, userID)
	delete(b.trainingStats, userID)

	// Створюємо нову структуру для статистики
	stats := trainingStats{
		correct:      0,
		incorrect:    0,
		words:        make([]uint, 0),
		currentIndex: 0,
	}

	// Вибираємо випадкові слова
	if mode == "10_words" {
		// Для режиму 10 слів
		selectedWords := make([]uint, 0)
		// Перемішуємо слова
		for i := len(words) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			words[i], words[j] = words[j], words[i]
		}
		// Беремо перші 10 слів (або менше, якщо слів менше 10)
		count := 10
		if len(words) < count {
			count = len(words)
		}
		for i := 0; i < count; i++ {
			selectedWords = append(selectedWords, words[i].ID)
		}
		stats.words = selectedWords
		fmt.Printf("Selected words for training: %v\n", stats.words)
	} else {
		// Для режиму безлімітного тренування
		stats.words = []uint{words[0].ID}
	}

	// Зберігаємо статистику
	b.trainingStats[userID] = stats

	// Встановлюємо перше слово
	word, err := b.wordService.GetWordByID(stats.words[0])
	if err != nil {
		fmt.Printf("Error getting first word: %v\n", err)
		return err
	}
	fmt.Printf("First word set: ID=%d, Word=%s\n", word.ID, word.EnglishWord)
	b.trainingWords[userID] = word.ID
	b.userStates[userID] = "training_" + mode

	return nil
}

func (b *Bot) handleText(c tele.Context) error {
	text := c.Text()
	userID := c.Sender().ID

	switch text {
	case "/stop":
		if stats, ok := b.trainingStats[userID]; ok {
			total := stats.correct + stats.incorrect
			accuracy := 0.0
			if total > 0 {
				accuracy = float64(stats.correct) / float64(total) * 100
			}

			response := fmt.Sprintf("Training stopped!\nResults:\nCorrect: %d\nIncorrect: %d\nAccuracy: %.1f%%",
				stats.correct, stats.incorrect, accuracy)

			delete(b.userStates, userID)
			delete(b.trainingWords, userID)
			delete(b.trainingStats, userID)

			return c.Send(response, b.getMainMenu(), &tele.SendOptions{
				ParseMode: tele.ModeHTML,
			})
		}
		return c.Send("No active training session", b.getMainMenu(), &tele.SendOptions{
			ParseMode: tele.ModeHTML,
		})

	case "➕ Add Word":
		b.userStates[userID] = "waiting_for_word"
		return c.Send(`Please send words in one of these formats:
1. Single word: english_word - translation
2. Multiple words (new line): 
   english_word1 - translation1
   english_word2 - translation2
3. Multiple words (comma): english_word1 - translation1, english_word2 - translation2`, &tele.SendOptions{
			ParseMode: tele.ModeHTML,
		})

	case "📚 My Words":
		user, _ := b.userService.GetOrCreateUser(userID, c.Sender().Username)
		words, err := b.wordService.GetUserWords(user.ID)
		if err != nil {
			return c.Send("Error getting words")
		}

		if len(words) == 0 {
			return c.Send("You don't have any words yet. Add some!")
		}

		var response strings.Builder
		response.WriteString("Your words:\n\n")
		for i, word := range words {
			response.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, word.EnglishWord, word.Translation))
		}
		return c.Send(response.String(), &tele.SendOptions{
			ParseMode: tele.ModeHTML,
		})

	case "✏️ Edit Word":
		user, _ := b.userService.GetOrCreateUser(userID, c.Sender().Username)
		words, err := b.wordService.GetUserWords(user.ID)
		if err != nil {
			return c.Send("Error getting words")
		}

		if len(words) == 0 {
			return c.Send("You don't have any words to edit. Add some first!")
		}

		var response strings.Builder
		response.WriteString("Select word number to edit:\n\n")
		for i, word := range words {
			response.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, word.EnglishWord, word.Translation))
		}

		b.userStates[userID] = "waiting_for_word_number_to_edit"
		return c.Send(response.String(), &tele.SendOptions{
			ParseMode: tele.ModeHTML,
		})

	case "🗑 Delete Word":
		user, _ := b.userService.GetOrCreateUser(userID, c.Sender().Username)
		words, err := b.wordService.GetUserWords(user.ID)
		if err != nil {
			return c.Send("Error getting words")
		}

		if len(words) == 0 {
			return c.Send("You don't have any words to delete. Add some first!")
		}

		var response strings.Builder
		response.WriteString("Select word number to delete:\n\n")
		for i, word := range words {
			response.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, word.EnglishWord, word.Translation))
		}

		b.userStates[userID] = "waiting_for_word_number_to_delete"
		return c.Send(response.String(), &tele.SendOptions{
			ParseMode: tele.ModeHTML,
		})

	default:
		if state, ok := b.userStates[userID]; ok {
			if strings.HasPrefix(state, "training_") {
				mode := strings.TrimPrefix(state, "training_")
				wordID, exists := b.trainingWords[userID]
				if !exists {
					fmt.Printf("No word ID found for user %d\n", userID)
					return c.Send("Something went wrong. Please start training again.")
				}

				word, err := b.wordService.GetWordByID(wordID)
				if err != nil {
					fmt.Printf("Error getting word for training: %v\n", err)
					return c.Send("Error getting word for training")
				}

				stats, ok := b.trainingStats[userID]
				if !ok {
					fmt.Printf("No training stats found for user %d\n", userID)
					return c.Send("Something went wrong. Please start training again.")
				}

				fmt.Printf("Current training state - Mode: %s, Index: %d, Words: %v, Current Word ID: %d\n",
					mode, stats.currentIndex, stats.words, wordID)

				if strings.ToLower(text) == strings.ToLower(word.Translation) {
					stats.correct++
					fmt.Printf("Correct answer for word %s\n", word.EnglishWord)

					if mode == "10_words" {
						stats.currentIndex++
						fmt.Printf("Moving to next word. New index: %d\n", stats.currentIndex)
						if stats.currentIndex >= len(stats.words) {
							// Тренування завершено
							total := stats.correct + stats.incorrect
							accuracy := float64(stats.correct) / float64(total) * 100
							response := fmt.Sprintf("Correct! 🎉\n\nTraining completed!\nResults:\nCorrect: %d\nIncorrect: %d\nAccuracy: %.1f%%",
								stats.correct, stats.incorrect, accuracy)
							delete(b.userStates, userID)
							delete(b.trainingWords, userID)
							delete(b.trainingStats, userID)
							return c.Send(response, b.getMainMenu(), &tele.SendOptions{
								ParseMode: tele.ModeHTML,
							})
						}
						// Переходимо до наступного слова
						nextWord, err := b.wordService.GetWordByID(stats.words[stats.currentIndex])
						if err != nil {
							fmt.Printf("Error getting next word: %v\n", err)
							return c.Send("Error getting next word")
						}
						fmt.Printf("Next word set: ID=%d, Word=%s\n", nextWord.ID, nextWord.EnglishWord)
						b.trainingWords[userID] = nextWord.ID
						b.trainingStats[userID] = stats // Оновлюємо статистику
						return c.Send(fmt.Sprintf("Correct! 🎉\n\nNext word: %s", nextWord.EnglishWord), &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					} else {
						// Для безлімітного режиму беремо нове випадкове слово
						user, _ := b.userService.GetOrCreateUser(userID, "")
						words, _ := b.wordService.GetUserWords(user.ID)
						if len(words) > 0 {
							// Перемішуємо слова
							for i := len(words) - 1; i > 0; i-- {
								j := rand.Intn(i + 1)
								words[i], words[j] = words[j], words[i]
							}
							nextWord := words[0]
							b.trainingWords[userID] = nextWord.ID
							b.trainingStats[userID] = stats // Оновлюємо статистику
							return c.Send(fmt.Sprintf("Correct! 🎉\n\nNext word: %s\nType /stop to end training", nextWord.EnglishWord), &tele.SendOptions{
								ParseMode: tele.ModeHTML,
							})
						}
					}
				} else {
					stats.incorrect++
					fmt.Printf("Incorrect answer for word %s. Expected: %s, Got: %s\n",
						word.EnglishWord, word.Translation, text)

					if mode == "10_words" {
						stats.currentIndex++
						fmt.Printf("Moving to next word. New index: %d\n", stats.currentIndex)
						if stats.currentIndex >= len(stats.words) {
							// Тренування завершено
							total := stats.correct + stats.incorrect
							accuracy := float64(stats.correct) / float64(total) * 100
							response := fmt.Sprintf("Incorrect. The correct translation is: %s\n\nTraining completed!\nResults:\nCorrect: %d\nIncorrect: %d\nAccuracy: %.1f%%",
								word.Translation, stats.correct, stats.incorrect, accuracy)
							delete(b.userStates, userID)
							delete(b.trainingWords, userID)
							delete(b.trainingStats, userID)
							return c.Send(response, b.getMainMenu(), &tele.SendOptions{
								ParseMode: tele.ModeHTML,
							})
						}
						// Переходимо до наступного слова
						nextWord, err := b.wordService.GetWordByID(stats.words[stats.currentIndex])
						if err != nil {
							fmt.Printf("Error getting next word: %v\n", err)
							return c.Send("Error getting next word")
						}
						fmt.Printf("Next word set: ID=%d, Word=%s\n", nextWord.ID, nextWord.EnglishWord)
						b.trainingWords[userID] = nextWord.ID
						b.trainingStats[userID] = stats // Оновлюємо статистику
						return c.Send(fmt.Sprintf("Incorrect. The correct translation is: %s\n\nNext word: %s\nType /stop to end training",
							word.Translation, nextWord.EnglishWord), &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					} else {
						// Для безлімітного режиму беремо нове випадкове слово
						user, _ := b.userService.GetOrCreateUser(userID, "")
						words, _ := b.wordService.GetUserWords(user.ID)
						if len(words) > 0 {
							// Перемішуємо слова
							for i := len(words) - 1; i > 0; i-- {
								j := rand.Intn(i + 1)
								words[i], words[j] = words[j], words[i]
							}
							nextWord := words[0]
							b.trainingWords[userID] = nextWord.ID
							b.trainingStats[userID] = stats // Оновлюємо статистику
							return c.Send(fmt.Sprintf("Incorrect. The correct translation is: %s\n\nNext word: %s\nType /stop to end training",
								word.Translation, nextWord.EnglishWord), &tele.SendOptions{
								ParseMode: tele.ModeHTML,
							})
						}
					}
				}
			} else {
				switch state {
				case "waiting_for_word":
					user, _ := b.userService.GetOrCreateUser(userID, c.Sender().Username)
					addedCount := 0
					errorCount := 0

					// Split by newlines first
					lines := strings.Split(text, "\n")
					for _, line := range lines {
						// Skip empty lines
						if strings.TrimSpace(line) == "" {
							continue
						}

						// Split by comma if present
						words := strings.Split(line, ",")
						for _, word := range words {
							// Skip empty words
							if strings.TrimSpace(word) == "" {
								continue
							}

							// Split by dash
							parts := strings.Split(strings.TrimSpace(word), "-")
							if len(parts) != 2 {
								errorCount++
								continue
							}

							englishWord := strings.TrimSpace(parts[0])
							translation := strings.TrimSpace(parts[1])

							// Skip if either part is empty
							if englishWord == "" || translation == "" {
								errorCount++
								continue
							}

							err := b.wordService.AddWord(user.ID, englishWord, translation)
							if err != nil {
								errorCount++
							} else {
								addedCount++
							}
						}
					}

					delete(b.userStates, userID)

					if addedCount > 0 {
						response := fmt.Sprintf("Successfully added %d word(s)", addedCount)
						if errorCount > 0 {
							response += fmt.Sprintf(", but %d word(s) had errors", errorCount)
						}
						return c.Send(response, &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					} else {
						return c.Send("No words were added. Please use the correct format: english_word - translation", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

				case "waiting_for_word_number_to_edit":
					wordNum, err := strconv.Atoi(text)
					if err != nil {
						return c.Send("Please enter a valid number", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

					user, _ := b.userService.GetOrCreateUser(userID, c.Sender().Username)
					words, err := b.wordService.GetUserWords(user.ID)
					if err != nil {
						return c.Send("Error getting words", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

					if wordNum < 1 || wordNum > len(words) {
						return c.Send("Invalid word number", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

					b.userStates[userID] = fmt.Sprintf("waiting_for_word_edit_%d", words[wordNum-1].ID)
					return c.Send("Please send the new word in format: english_word - translation", &tele.SendOptions{
						ParseMode: tele.ModeHTML,
					})

				case "waiting_for_word_number_to_delete":
					wordNum, err := strconv.Atoi(text)
					if err != nil {
						return c.Send("Please enter a valid number", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

					user, _ := b.userService.GetOrCreateUser(userID, c.Sender().Username)
					words, err := b.wordService.GetUserWords(user.ID)
					if err != nil {
						return c.Send("Error getting words", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

					if wordNum < 1 || wordNum > len(words) {
						return c.Send("Invalid word number", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

					err = b.wordService.DeleteWord(words[wordNum-1].ID)
					if err != nil {
						return c.Send("Error deleting word", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}

					delete(b.userStates, userID)
					return c.Send("Word deleted successfully!", &tele.SendOptions{
						ParseMode: tele.ModeHTML,
					})

				default:
					if strings.HasPrefix(state, "waiting_for_word_edit_") {
						parts := strings.Split(text, " - ")
						if len(parts) != 2 {
							return c.Send("Please use format: english_word - translation", &tele.SendOptions{
								ParseMode: tele.ModeHTML,
							})
						}

						wordID, _ := strconv.ParseUint(strings.TrimPrefix(state, "waiting_for_word_edit_"), 10, 32)
						err := b.wordService.UpdateWord(uint(wordID), parts[0], parts[1])
						if err != nil {
							return c.Send("Error updating word", &tele.SendOptions{
								ParseMode: tele.ModeHTML,
							})
						}

						delete(b.userStates, userID)
						return c.Send("Word updated successfully!", &tele.SendOptions{
							ParseMode: tele.ModeHTML,
						})
					}
				}
			}
		}
	}

	return c.Send("Please use the menu buttons to interact with the bot", &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	})
}
