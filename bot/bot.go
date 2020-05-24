package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/smulube/scavenge/store"
	"go.uber.org/zap"
	"gopkg.in/guregu/null.v4"
)

// Run starts our bot running with the appropriate configuration
func Run(logger *zap.Logger, token string, start time.Time, duration time.Duration, verbose bool, connStr string, admins []string) error {

	logger.Info(
		"Starting Lockdown Scavenger Bot",
		zap.String("startTime", start.Format(time.RFC3339)),
		zap.String("duration", duration.String()),
		zap.Bool("verbose", verbose),
		zap.String("connStr", connStr),
		zap.String("admins", strings.Join(admins, ",")),
	)

	db, err := store.New(connStr, logger)
	if err != nil {
		return fmt.Errorf("Failed to initialize DB: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("Failed to instantiate bot API: %v", err)
	}

	bot.Debug = verbose

	logger.Info("Bot authorized successfully")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		tx, err := db.Beginx()
		if err != nil {
			logger.Error("Failed to begin transaction", zap.Error(err))
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// bundle operations inside anonymous function for cleaner transaction handling
		err = func() error {
			user, err := store.GetUser(tx, update.Message.From.ID)
			if err != nil && err != store.ErrNotFound {
				return fmt.Errorf("Unexpected error while retrieving user: %v", err)
			}

			if update.Message.Chat.IsPrivate() || update.Message.IsCommand() {
				if user == nil {
					user = &store.User{
						ID:        int64(update.Message.From.ID),
						UserName:  null.StringFrom(update.Message.From.UserName),
						FirstName: update.Message.From.FirstName,
						Admin:     contains(admins, update.Message.From.UserName),
					}

					err = store.SaveUser(tx, user)
					if err != nil {
						return fmt.Errorf("Unexpected error while saving user: %v", err)
					}
				}

				// we have a saved user now
				if update.Message.IsCommand() {
					switch update.Message.Command() {
					case "listteams":
						teams, err := store.GetTeams(tx)
						if err != nil {
							return err
						}

						if len(teams) == 0 {
							msg.Text = "There are currently no registered teams"
							return nil
						}

						var buf strings.Builder
						buf.WriteString("Available teams: \n\n")
						for _, team := range teams {
							buf.WriteString(" - ")
							buf.WriteString(team.Name)
							buf.WriteString("\n")
						}

						msg.Text = buf.String()

					case "createteam":
						if update.Message.CommandArguments() == "" {
							msg.Text = "You must tell me a non-empty team name"
							return nil
						}

						if len(update.Message.CommandArguments()) < 5 {
							msg.Text = "Your team name must have more than five characters"
							return nil
						}

						team := &store.Team{Name: update.Message.CommandArguments()}

						err = store.CreateTeam(tx, team)
						if err != nil {
							return err
						}

						msg.Text = fmt.Sprintf("Team '%s' successfully created!", update.Message.CommandArguments())

					case "jointeam":
						if update.Message.CommandArguments() == "" {
							msg.Text = "You must tell me a non-empty team name"
							return nil
						}

						team, err := store.GetTeamByName(tx, update.Message.CommandArguments())
						if err != nil {
							if err == store.ErrNotFound {
								msg.Text = fmt.Sprintf("I'm sorry, I can't find a team with the name '%s', please check and tell me again", update.Message.CommandArguments())
								return nil
							}
							return err
						}

						user.TeamID = null.IntFrom(team.ID)

						err = store.SaveUser(tx, user)
						if err != nil {
							return err
						}

						msg.Text = fmt.Sprintf("You have now joined the team: %s", team.Name)

					case "me":
						var buf strings.Builder
						buf.WriteString("Your name is: ")
						buf.WriteString(user.FirstName)
						buf.WriteString("\n")

						if user.TeamID.Valid {
							team, err := store.GetTeamById(tx, int(user.TeamID.Int64))
							if err != nil {
								return err
							}

							buf.WriteString("You are currently in team: ")
							buf.WriteString(team.Name)
						} else {
							buf.WriteString(("You are not currently in any team"))
						}

						msg.Text = buf.String()
					case "help":
						msg.Text = "Helpful message"

					default:
						msg.Text = "I'm afraid I don't know that command, please type /help to see a list of available commands"
					}
				}
			}

			return nil
		}()

		if err != nil {
			logger.Error("Unexpected error", zap.Error(err))
			msg.Text = fmt.Sprintf("I'm sorry I encountered an unexpected error. Please tell Sam about this: %v", err.Error())
			tx.Rollback()
		}

		tx.Commit()

		bot.Send(msg)

	}

	return nil
}

func contains(haystack []string, needle string) bool {
	for _, elem := range haystack {
		if elem == needle {
			return true
		}
	}
	return false
}
