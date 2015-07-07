package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rockneurotiko/go-bots/rssbot/rssbot"
	"github.com/rockneurotiko/go-homedir"
)

func isValidDbDir(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return true
	}
	return fi.IsDir()
}

func isValidFile(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func main() {
	var dbdir string
	var deploy string
	var envdir string

	flag.StringVar(&dbdir, "db", "./rssdb.db", "database file")
	flag.StringVar(&deploy, "deploy", "", "Run in deploy")
	flag.StringVar(&envdir, "env", "secrets.env", "Environment file (secret.env)")
	flag.Parse()
	edbdir, err := homedir.Expand(dbdir)
	if err != nil || edbdir == "" || !isValidDbDir(edbdir) {
		fmt.Println("Database path not valid")
		return
	}
	dbdir = edbdir

	eenvdir, err := homedir.Expand(envdir)
	if err != nil || eenvdir == "" || !isValidFile(eenvdir) {
		fmt.Println("Environment path not valid")
		return
	}
	envdir = eenvdir

	fmt.Println("DataBase dir: " + dbdir)
	fmt.Println("Environment dir: " + envdir)
	fmt.Println("Deploy url: " + deploy)

	godotenv.Load(envdir)
	// TELEGRAM_TOKEN=yourtoken
	token := os.Getenv("TELEGRAM_TOKEN")
	notify := os.Getenv("NOTIFY") == "true"
	bot := rssbot.BuildBot(dbdir, token, notify)

	if deploy != "" {
		bot.ServerStart(deploy, "/webhook")
	} else {
		bot.SimpleStart()
	}
}
