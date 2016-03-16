package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rockneurotiko/go-bots/pocketbot/pocket"
	"github.com/rockneurotiko/go-homedir"
	"github.com/syndtr/goleveldb/leveldb"
)

func isValidFile(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func main() {
	var configfile string
	flag.StringVar(&configfile, "config", "config.json", "Config file")
	flag.Parse()

	ccfile, err := homedir.Expand(configfile)
	if err != nil || ccfile == "" || !isValidFile(ccfile) {
		fmt.Println("Config path not valid")
		return
	}
	configfile = ccfile

	fmt.Println("Config file: " + configfile)

	file, e := ioutil.ReadFile(configfile)

	if e != nil {
		fmt.Printf("Error reading config file %s: %s\n", configfile, e.Error())
		os.Exit(1)
	}

	var configj pocket.ConfigJ
	json.Unmarshal(file, &configj)
	token := configj.Token

	db, err := leveldb.OpenFile(configj.DbPath, nil)
	if err != nil {
		fmt.Printf("Error opening database %s\n", configj.DbPath)
		return
	}
	defer db.Close()

	bot := pocket.BuildBot(token, configj, db)
	bot.SimpleStart()
}
