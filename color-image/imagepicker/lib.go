package imagepicker

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strconv"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/rockneurotiko/go-tgbot"
)

type ConfigJ struct {
	Token string `json:"token"`
}

var (
	white color.Color = color.RGBA{255, 255, 255, 255}
	black color.Color = color.RGBA{0, 0, 0, 255}
	blue  color.Color = color.RGBA{0, 0, 255, 255}
)

func createImage(s1 int, s2 int, s3 int, s4 int, color color.Color) image.Image {
	m := image.NewRGBA(image.Rect(s1, s2, s3, s4))
	draw.Draw(m, m.Bounds(), &image.Uniform{color}, image.ZP, draw.Src)
	return m.SubImage(m.Bounds())
}

func createDefaultImage(c color.Color) image.Image {
	return createImage(0, 0, 640, 480, c)
}

func rgba(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	fmt.Println(args)

	r := "0"
	g := "0"
	b := "0"
	a := "255"

	if len(args) >= 2 {
		r = args[1]
	}
	if len(args) >= 3 {
		g = args[2]
	}
	if len(args) >= 4 {
		b = args[3]
	}
	if len(args) >= 5 {
		a = args[4]
	}

	ri, er := strconv.ParseUint(r, 10, 8)
	gi, eg := strconv.ParseUint(g, 10, 8)
	bi, eb := strconv.ParseUint(b, 10, 8)
	ai, ea := strconv.ParseUint(a, 10, 8)

	if er != nil || eg != nil || eb != nil || ea != nil || ri > 255 || gi > 255 || bi > 255 || ai > 255 {
		return nil
	}

	bot.Answer(msg).
		Photo(createDefaultImage(color.RGBA{uint8(ri), uint8(gi), uint8(bi), uint8(ai)})).
		End()

	return nil
}

func hex_hand(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	c, err := colorful.Hex(fmt.Sprintf("#%s", args[1]))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	image := createDefaultImage(c)
	bot.Answer(msg).Photo(image).End()
	return nil
}

func help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).
		Text(fmt.Sprintf(`Welcome to Color Sample Bot! How are you %s?

You can ask me for some color :)

This are the commands right now:

/rgb - Return the RGB color. Example: /rgb 123 123 123
You can provide 1 argument (Red), 2 arguments (Red and Green) or the 3 arguments (Red Green Blue)

/hex - Return the HEX color. Example: /hex #7B7B7B
You can remove the # (/hex 7B7B7B)

Icon made by Freepik (License CC BY 3.0)

This bot is open source and has been created by @rockneurotiko, I hope that you like it ;-)

The source code can be founded in: https://github.com/rockneurotiko/go-bots/tree/master/color-image
`, msg.From.FirstName)).
		DisablePreview(true).
		End()
	return nil
}

func BuildBot(token string, localconfig ConfigJ) *tgbot.TgBot {
	bot := tgbot.New(token).
		SimpleCommandFn(`help`, help).
		SimpleCommandFn(`start`, help).
		CommandFn(`hex (?:#)?([0-9A-F]{3}|[0-9A-F]{5,6})`, hex_hand).
		MultiCommandFn([]string{
		`rgba?`,
		`rgba? (\d{1,3})`,
		`rgba? (\d{1,3}) (\d{1,3})`,
		`rgba? (\d{1,3}) (\d{1,3}) (\d{1,3})`,
		// `rgba (\d{1,3}) (\d{1,3}) (\d{1,3}) (\d{1,3})`,
	}, rgba)
	return bot
}
