package main

import (
	"english-words-bot/internal/bot"
	"english-words-bot/internal/db"
	"english-words-bot/internal/version"
	"flag"
	"log"
	"os"
)

var (
	showVersion bool
)

func init() {
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()
}

func main() {
	if showVersion {
		log.Printf("Version: %s\nBuild Time: %s\nGit Commit: %s\n",
			version.Version,
			version.BuildTime,
			version.GitCommit)
		return
	}

	// Initialize database
	db.InitDB()

	// Get bot token from environment variable
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}

	// Create and start bot
	b, err := bot.NewBot(token)
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}

	log.Printf("Bot started... (version %s)", version.Version)
	b.Start()
}
