package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rockneurotiko/go-bots/kimsufi/kimsufi"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	godotenv.Load("secrets.env")
	token := os.Getenv("TELEGRAM_KEY")
	dbpath := os.Getenv("DB_PATH")

	db, err := leveldb.OpenFile(dbpath, nil)
	if err != nil {
		fmt.Printf("Error opening database %s\n", dbpath)
		return
	}
	defer db.Close()

	bot := kimsufi.BuildBot(token, db)
	bot.SimpleStart()
}
