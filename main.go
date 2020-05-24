package main

import (
	"log"
	"time"

	"github.com/alecthomas/kong"
	"github.com/smulube/scavenge/bot"
	"go.uber.org/zap"
)

var cli struct {
	Verbose       bool          `kong:"help='Enable verbose mode',default=false"`
	Start         time.Time     `kong:"help='Start time of the game as an ISO8601 string',required"`
	Duration      time.Duration `kong:"help='Duration of the game',default='1h'"`
	TelegramToken string        `kong:"help='Telegram token for the bot',required,env='TELEGRAM_TOKEN'"`
	Database      string        `kong:"help='Connection string for Postgres',required,env='POSTGRES_CONN'"`
	Admins        []string      `kong:"help='Comma separated list of admin users'"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Unable to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	ctx := kong.Parse(
		&cli,
		kong.Name("scavenge"),
		kong.Description("Telegram bot providing the back end for our lockdown scavenger hunt"),
	)

	err = bot.Run(logger, cli.TelegramToken, cli.Start, cli.Duration, cli.Verbose, cli.Database, cli.Admins)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

}
