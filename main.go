package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/smulube/scavengerbot/bot"
	"go.uber.org/zap"
)

var cli struct {
	Verbose       bool     `kong:"help='Enable verbose mode',default=false"`
	GameFile      string   `kong:"help='Path to the game config file',required,default=./game.yaml"`
	ImageFolder   string   `kong:"help='Path to folder within which images are saved',required,default=./gallery"`
	TelegramToken string   `kong:"help='Telegram token for the bot',required,env='TELEGRAM_TOKEN'"`
	Database      string   `kong:"help='Connection string for Postgres',required,env='POSTGRES_CONN'"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Unable to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	ctx := kong.Parse(
		&cli,
		kong.Name("scavengerbot"),
		kong.Description("Telegram bot providing the back end for our lockdown scavenger hunt"),
	)

	err = bot.Run(logger, cli.TelegramToken, cli.GameFile, cli.ImageFolder, cli.Verbose, cli.Database)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

}
