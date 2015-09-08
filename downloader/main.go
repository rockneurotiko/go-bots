package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rockneurotiko/go-bots/downloader/downloader"
)

func main() {
	godotenv.Load("secrets.env")
	token := os.Getenv("TELEGRAM_KEY")
	bot := downloader.BuildBot(token, 5)
	fmt.Println("Let's start!")
	bot.SimpleStart()
}
