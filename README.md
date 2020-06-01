# scavengerbot

This is a Telegram bot to support my little lockdown scavenger hunt game for
friends and family.

## Building

To build the project you will need to:

- install Go (latest build is best)
- add the Go bin folder into your \$PATH - mine is like `$HOME/go/bin`
- run `go install` which should build the binary, and copy it into `$HOME/go/bin`

## Running

Included in this repo is a simple docker-compose file to run a Postgre
instance inside Docker. I included this so I didn't have to try and install a
local Postgres instance just for this. If you already have Postgres installed
locally, you can create a new DB, and pass in the connection string as shown
in the API below.

To run using Docker, make sure you have a recent version of docker + compose
installed, then start the DB with:

```bash
$ docker-compose up
```

Once you see Postgres has started you should then be able to start and
connect to using the connection string shown in `.env.local.example`

## CLI API

```bash
Usage: scavengerbot --game-file="./game.yaml" --image-folder="./gallery" --telegram-token=STRING --database=STRING

Telegram bot providing the back end for our lockdown scavenger hunt

Flags:
  --help                        Show context-sensitive help.
  --verbose                     Enable verbose mode
  --game-file="./game.yaml"     Path to the game config file
  --image-folder="./gallery"    Path to folder within which images are saved
  --telegram-token=STRING       Telegram token for the bot ($TELEGRAM_TOKEN)
  --database=STRING             Connection string for Postgres ($POSTGRES_CONN)
```

Note that the values of TELEGRAM_TOKEN and POSTGRES_CONN can be passed via
environment variables of those names - this can be seen in the included
`.env.local.example` file.

## Chat API

The Chat API is currently very clunky - users have to type a command like
`/jointeam My Team Name` where precise case sensitivity is required.

The current list of commands understood by the bot are:

```
Hello, I know the following commands:

  - /listteams - list the current teams
  - /createteam - used to create a new team
  - /jointeam - used to join an existing team
  - /leaveteam - used to leave your current team
  - /me - show your current status
  - /rules - list the rules of the game
  - /items - list the items we are currently looking for
  - /game - show the current game status
```

## Future work

- Use Telegram buttons in order to simplify the above UI. These are buttons
  that the Telegram app shows to the user, so they might type `/jointeam`, the
  Telegram app would return a message that renders a button for each team. The
  user could then just click on the name of a team to join it.

- Add a `/start` message which will give a nicer experience when you go into
  private chat with the bot

- Fix the humanized time - gives bad output when say game starts in 90 minutes

  - in this case it returns the message: "game will start in one hour" instead
    of maybe - "game will start in over an hour", or better "game will start in
    90 minutes"

- Write some tests for the above functionality

- See whether the bot can proactively send messages out to the group chat, i.e.
  The game is about to start, or "The game has 5 minutes to go", or even "The
  game has finished!" - currently it just responds to incoming messages

- Find a service with an easily usable gallery API that will upload all of a
  teams photos into a shared gallery such that we can easily then review and
  score the items.
