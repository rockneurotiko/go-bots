package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/rockneurotiko/go-bots/color-image/imagepicker"
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

	var configj imagepicker.ConfigJ
	json.Unmarshal(file, &configj)
	token := configj.Token

	bot := imagepicker.BuildBot(token, configj)
	bot.SimpleStart()
}
