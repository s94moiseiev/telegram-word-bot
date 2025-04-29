# English Words Bot

Telegram bot for learning English words. The bot helps users maintain their personal dictionary of English words and practice translations through interactive training sessions.

## Features

- 📚 Personal dictionary management
  - Add new words
  - View your word list
  - Edit existing words
  - Delete words
- 🎯 Training modes
  - 10 Words Training: Practice with 10 random words
  - Continuous Training: Practice until you decide to stop
- 📊 Training statistics
  - Track correct and incorrect answers
  - View accuracy percentage

## Requirements

- Go 1.16 or higher
- SQLite3
- Telegram Bot Token

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/english-words-bot.git
cd english-words-bot
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export BOT_TOKEN="your_telegram_bot_token"
```

## Building

To build the binary with version information:
```bash
go build -ldflags="-X 'english-words-bot/internal/version.BuildTime=$(date)' -X 'english-words-bot/internal/version.GitCommit=$(git rev-parse HEAD)'" -o english-words-bot ./cmd/main.go
```

## Usage

1. Start the bot:
```bash
./english-words-bot
```

2. View version information:
```bash
./english-words-bot --version
```

3. In Telegram:
   - Start the bot with `/start`
   - Use the menu buttons to:
     - Add new words
     - View your dictionary
     - Edit words
     - Delete words
     - Start training sessions

## Training Modes

### 10 Words Training
- Bot selects 10 random words from your dictionary
- Practice translations one by one
- Get immediate feedback
- View final results with accuracy

### Continuous Training
- Practice with random words from your dictionary
- Continue until you type `/stop`
- View statistics when finished

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 