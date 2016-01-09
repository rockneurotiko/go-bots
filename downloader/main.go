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
	youtubeurl := os.Getenv("YOUTUBE_URL")
	instaid := os.Getenv("INSTAGRAM_ID")

	slideapi := os.Getenv("SLIDESHARE_KEY")
	slidesecret := os.Getenv("SLIDESHARE_SECRET")

	bot := downloader.BuildBot(token, 1, youtubeurl, instaid, slideapi, slidesecret)
	fmt.Println("Let's start!")
	bot.SimpleStart()
}
