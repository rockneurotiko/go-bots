package rssbot

import (
	"bytes"
	"fmt"
	"sort"
)

/*
Changelogs:

0.1.0:
- Like you are seeing, now I can send messages with the new things!
- The sub, rm and delete command now handle multiple spaces correctly and in rm/delete you can write with a last ) ex: "/rm 1)"
- Now I'm able to handle preferences, right now I only have one: "image" preference, that set if you want to receive the images from the news or don't (don't by default, you will have to enable it)
Use:

/preference image -> It will ask you to enable or disable, you will have a keyboard to do that :)

/enable or /disable -> Done! :)

You can cancel this process with /cancel

*/

const (
	version       = "0.2.0"
	lastchangelog = `- Now I support sending more type of media, jpg, png, wepb and gifs.
- `
)

var messageUpdate = `Hi!
This is an automatic message to let you know that I'm waking up, probably I'm a new version and you have new features, or less bugs ^^
I won't be able to send you the updates that had been made while I was down, I'm so sorry if it was something important...
But blame my master @rock_neurotiko not me :'(
Thanks for your understanding!`

// start - Start the bot
// help - Show this help
// sub  - Subscribe to that RSS
// list - Return your RSS subscriptions
// delete - Remove your subscription of the RSS <id> (an integer)
// rm - Remove your subscription of the RSS <id> (an integer)
// preference - Change a preference

var availableCommands = map[string]string{
	"/start":              "Start the bot",
	"/help":               "Show this help",
	"/sub <url>":          "Subscribe to that RSS, the URL need to have 'http://' or 'https://'",
	"/list":               "Return your RSS subscriptions",
	"/delete <id>":        "Remove your subscription of the RSS <id> (an integer)",
	"/rm <id>":            "Remove your subscription of the RSS <id> (an integer)",
	"/preference (image)": "Change the preference",
}

var helptoptext = `Hi! I'm the RSS Bot, with me you can subscribe to the RSS of your favourites webpages, and I'll send you the updates, when they update their RSS.
Please, before downvote me, I only work with RSS, not with webpages, you should search for an icon similar to my avatar. And, if you have any trouble, before go to downvote, talk to me in @rock_neurotiko, maybe it's a bug and I can fix it :-)

This is the available commands:`
var helpbottomtext = fmt.Sprintf(`Please, if you like and use this bot, consider vote in https://telegram.me/storebot?start=RSSNewsBot

Also you have any suggestion or issue you can contact with the main developer of this bot: @rock_neurotiko

All the code of this bot is Open Source, you can see it or contribute in https://github.com/rockneurotiko/go-bots/tree/master/rssbot

@RSSNewsBot version: %s`, version)

func buildHelp() string {
	var buffer bytes.Buffer

	buffer.WriteString(helptoptext + "\n\n")

	orderk := []string{}
	for cmd := range availableCommands {
		orderk = append(orderk, cmd)
	}
	sort.Strings(orderk)
	for _, cmd := range orderk {
		htext := availableCommands[cmd]
		buffer.WriteString(fmt.Sprintf("%s - %s\n", cmd, htext))
	}

	buffer.WriteString("\n" + helpbottomtext)
	return buffer.String()
}
