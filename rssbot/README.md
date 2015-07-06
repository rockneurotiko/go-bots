# RSS Bot

This is the source code of the RSS bot developed with my [Telegram Bot's API library, tgbot](https://www.github.com/rockneurotiko/go-tgbot).

I have this bot running in [@RSSNewsBot](https://telegram.me/RSSNewsBot), but this project allow you to run your own RSS bot. The code that runs [@RSSNewsBot](https://telegram.me/RSSNewsBot) is exactly the same that you find here.

If you see that [@RSSNewsBot](https://telegram.me/RSSNewsBot) don't answer to /help or /start in one hour, please [let me know](https://telegram.me/rock_neurotiko), I have it running in a really limited server, and I think that they are going to stop it when I reach some limit xD (The one hour limit is because maybe I'm deploying a new version, or doing some troubleshooting)

If you have any suggestion, issue, review or whatever, you can talk to me in [telegram](https://telegram.me/rock_neurotiko)

If you like this project, you can vote my bot [here](https://telegram.me/storebot?start=rssnewsbot).


# Features!

- Binary for all major platforms and architectures, just search yours in the bin directory :)
- Configured with flags.
- Database in disk, the state will be saved between runs.
- Caches

# Configure

If you don't like (or can't) to use the scrit I provide to execute (`run.sh` or `run.cmd`) you can execute in the terminal like any program. You can execute the binary, or the go code.

Parameters to configure the program (All the paths are OS-like you are using, in unix with / in windows with \, but will expand ~ in both):
- `-db=~/path/to/db`: The final path where will be saved the db. Default: `book.db`
- `-env=~/path/to/secrets.env`: The path where the secrets.env is stored. Default: `secrets.env`
- `-deploy=https://example.org`: If you provides this parameter, the program will start as a server, using `$HOST` and `$PORT`, and setting a webhook with base the url provided, if you don't provide this parameter, it will use the `getUpdates` way (easier if you don't have a server or SSL). Default: No parameter

# Execute

- Talk with [@botfather](https://telegram.me/botfather) and create a new bot following the steps.
- Create a file called `secrets.env` and add the token with the format:
  `TELEGRAM_TOKEN=1111111:AAAAAAAABBBBBBBBBBB`
- If you are in unix system open `run.sh` if you are in windows open `run.cmd`, if you are not in any of these, see some of these to know how to run it xD
- Change the values like you want, the values are the three explained in the [configure](#configure) section one the binary:
  - BINARY: The path for the binary, make sure that is the one you need!
  - DBPATH: Directory of the database
  - ENVDIR: The path where the secrets.env file are.
  - URL: The URL to deploy, this can be empty (actually, let this empty if you don't have a server)


That's all! Just run the file you edited! The program will show you the paths configured (if you see that it's not what you want, just stop it and modify the file).

I hope you that you enjoy it, and [vote it here](https://telegram.me/storebot?start=rssnewsbot)


# Other things

- Commands of the bot:
  - `/start`: Shows the help
  - `/help`: Shows the help
  - `sub <rss_url>`: Subscribe to the RSS url
  - `/list`: Return your RSS subscriptions.
  - `/delete <id>`: Remove your subscription of the RSS <id> (an integer)
  - `/rm <id>`: Just a short way of remove :)


This is what you can send to @botfather when you use the `/setcommands`:
```
start - Start the bot
help - Show this help
sub  - Subscribe to that RSS
list - Return your RSS subscriptions
delete - Remove your subscription of the RSS <id> (an integer)
rm - Remove your subscription of the RSS <id> (an integer)
```
