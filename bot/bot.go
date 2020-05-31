package bot

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dustin/go-humanize"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/smulube/scavenge/store"
	"go.uber.org/zap"
	"gopkg.in/guregu/null.v4"
	"gopkg.in/yaml.v2"
)

//var (
//nonTeamKeyboard = tgbotapi.NewInlineKeyboardMarkup(
//tgbotapi.NewInlineKeyboardRow(
//tgbotapi.NewInlineKeyboardButtonData("Join Team", "/listteams"),
//),
//)
//)

func init() {
	ctx := context.Background()
	f, err := os.Open("foo.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	wc := client.Bucket("scavenger.mulube.com").Object("foo.txt").NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		panic(err)
	}
	if err := wc.Close(); err != nil {
		panic(err)
	}

	//bucketName := "scavenger.mulube.com"
	//bucket := client.Bucket(bucketName)

}

// Game type loaded from YAML
type Game struct {
	Title    string        `yaml:"title"`
	Start    time.Time     `yaml:"start"`
	Duration time.Duration `yaml:"duration"`
	Items    []string      `yaml:"items"`
	Bonuses  []string      `yaml:"bonuses"`
}

// Run starts our bot running with the appropriate configuration
func Run(logger *zap.Logger, token string, gameFile, imageFolder string, verbose bool, connStr string, admins []string) error {

	data, err := ioutil.ReadFile(gameFile)
	if err != nil {
		return fmt.Errorf("Unable to open game file: %v", err)
	}

	var game Game
	err = yaml.Unmarshal(data, &game)
	if err != nil {
		return fmt.Errorf("Error loading game file: %v", err)
	}

	logger.Info(
		"Starting Lockdown Scavenger Bot",
		zap.String("gameFile", gameFile),
		zap.String("gallery", imageFolder),
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
						ID: int64(update.Message.From.ID),
						// UserName:  null.StringFrom(update.Message.From.UserName),
						FirstName: update.Message.From.FirstName,
						Admin:     contains(admins, update.Message.From.UserName),
					}

					if update.Message.From.UserName != "" {
						user.UserName = null.StringFrom(update.Message.From.UserName)
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

					case "leaveteam":
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

						if user.TeamID.Int64 != team.ID {
							msg.Text = fmt.Sprintf("You are not currently in the team '%s' so you cannot leave them!", update.Message.CommandArguments())
							return nil
						}

						user.TeamID = null.Int{}

						err = store.SaveUser(tx, user)
						if err != nil {
							return err
						}

						msg.Text = fmt.Sprintf("You have now left the team: %s", team.Name)

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

					case "game":
						now := time.Now()

						if game.Start.After(now) {
							msg.Text = fmt.Sprintf("The game '%s' has not yet started, and is due to begin %s", game.Title, humanize.Time(game.Start))
							return nil
						}

						if now.After(game.Start) && game.Start.Add(game.Duration).After(now) {
							msg.Text = fmt.Sprintf("The game '%s' is currently underway and is due to finish %s", game.Title, humanize.Time(game.Start.Add(game.Duration)))
							return nil
						}

						msg.Text = fmt.Sprintf("The game '%s' has already finished", game.Title)

					case "items":
						now := time.Now()

						if now.After(game.Start) && game.Start.Add(game.Duration).After(now) {
							var buf strings.Builder
							buf.WriteString("The game is afoot! The items you are looking for are:\n\n")
							for _, item := range game.Items {
								buf.WriteString(" - ")
								buf.WriteString(item)
								buf.WriteString("\n")
							}

							if len(game.Bonuses) > 0 {
								buf.WriteString("\n\nThere are also the following bonus items to be found:\n\n")

								for _, bonus := range game.Bonuses {
									buf.WriteString(" - ")
									buf.WriteString(bonus)
									buf.WriteString("\n")
								}
							}

							msg.Text = buf.String()
							return nil
						}

						msg.Text = "The game has not yet started so I can't tell you the items we are looking for yet!"

					case "help":
						msg.Text = `Hello, I know the following commands:

  - /listteams - list the current teams
  - /createteam - used to create a new team
  - /jointeam - used to join an existing team
  - /leaveteam - used to leave your current team
  - /me - show your current status
  - /rules - list the rules of the game
  - /items - list the items we are currently looking for
  - /game - show the current game status
`

					default:
						msg.Text = "I'm afraid I don't know that command, please type /help to see a list of available commands"
					}
				}

				if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
					if !user.TeamID.Valid {
						msg.Text = "You are not in a team, so I can't save photos for you"
						return nil
					}

					photos := *update.Message.Photo
					go savePhoto(bot, user, photos[len(photos)-1], game.Title, imageFolder)
				}
			}

			return nil
		}()

		if err != nil {
			logger.Error("Unexpected error", zap.Error(err))
			msg.Text = fmt.Sprintf("I'm sorry I encountered an unexpected error. Please tell Sam about this: %v", err.Error())
			tx.Rollback()
			continue
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

func savePhoto(bot *tgbotapi.BotAPI, user *store.User, photo tgbotapi.PhotoSize, gameTitle, galleryFolder string) {
	photoURL, err := bot.GetFileDirectURL(photo.FileID)
	if err != nil {
		fmt.Println(err)
		return
	}

	teamPath := filepath.Join(galleryFolder, strings.ReplaceAll(gameTitle, " ", "-"), strconv.FormatInt(user.TeamID.Int64, 10))

	err = os.MkdirAll(teamPath, 0700)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.Get(photoURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	filename := filepath.Join(teamPath, fmt.Sprintf("%s.jpg", photo.FileID))
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return
}
