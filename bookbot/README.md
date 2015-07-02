# Book Bot (File Share Bot)!

This is the source code of the first bot developed with my [Telegram Bot's API library, tgbot](https://www.github.com/rockneurotiko/go-tgbot).

I called Book Bot because is the purpose that I use, but actually is more like a file server, because you can send any document :smile:

You don't have to know any coding, just look the [execute](#execute) section.

If you have any suggestion, issue, review or whatever, you can talk to me in [telegram](https://telegram.me/rock_neurotiko)

If you like this project, you can vote my bot (it's password protected, but you it's a way of showing me how many people like this ^^): [vote here](https://telegram.me/storebot?start=b00kbot). Also, if you really really want to see my books, you can ask me in private.

There is an example imagen of what can do ;-)

![alt screenshot](http://web.neurotiko.com/bookbot_screenshot.png)

(And yes, I have that book purchased :P)

# Features!

- Binary for all major platforms and architectures, just search yours in the bin directory :)
- Configured with flags.
- Protected by password if you want.
  If you have sensible files, or you don't want that anyone can use your bot, you can set a password that the users/groups will have to set in the `/start` command. (`/start <pwd>`)
- Real send only one time. If you send a file once, the next times you send the file to any user, you don't have to upload it again, so you only upload the files once.
- Database, save all the users and book ids, so you can close the program and start again without loosing anything.

# Configure

If you don't like (or can't) to use the scrit I provide to execute (`run.sh` or `run.cmd`) you can execute in the terminal like any program. You can execute the binary, or the go code.

Parameters to configure the program (All the paths are OS-like you are using, in linux with / in windows with \, but will expand ~ in both):
- `--dir=~/path/to/your/files`: The path of the base directory of the files you want to share.
- `--db=~/path/to/db`: The final path where will be saved the db.
- `--pwd=password`: The password to protect who can execute commands and who don't. In the chats, if one enable it, it will be enabled for all the chat.

# Execute

- Talk with [@botfather](https://telegram.me/botfather) and create a new bot following the steps.
- Change the "YOURKEY" in `secrets.env` file to the one that @botfather gived to you.
- If you are in unix system open `run.sh` if you are in windows open `run.cmd`, if you are not in any of these, see some of these to know how to run it xD
- Change the values like you want, the values are the three explained in the [configure](#configure) section one the binary:
  - BINARY: The path for the binary, make sure that the one you need!
  - BOOKSPATH: Base directory of the files
  - DBPATH: Directory of the database
  - PWD: The password to protect the bot. It can be empty if you don't want to protect.


That's all! Just run the file you edited! The program will show you the paths configured (if you see that it's not what you want, just stop it and modify the file) and what had been loaded from the database.

I hope you that you enjoy it!

# Other things

- Commands of the bot:
  - `/start <pwd>`: Start the bot, use it when the bot have password
  - `/help`: Get the help.
  - `/cd`: Change your relative path to $HOME (the base)
  - `/cd <name|id>`: Change your relative path to the specified (I recommend use the number listed)
  - `/ls`: Show the files and directories in the current path
  - `/ls <name|id>`: Show the files and directories in the path specified.
  - `/download <name|id>` | `/dl <name|id>` | `/dw <name|id>`: Download the file specified.


This is what you can send to @botfather when you use the `/setcommands`:
```
start - Start the bot
help - Show this help
cd - Change user directory
ls - Show current user directory content
download - Download specified file
dw - Shortcut to /download :)
dl - Shourcut to /download :)
```
