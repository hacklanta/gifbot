package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/nlopes/slack"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	requestGifRegex = regexp.MustCompile("^\\.gif ([^ ]+)$")

	storeGifRegex = regexp.MustCompile("^\\.storegif ([^ ]+) <([^ ]+)>$")

	botId = ""
)

func handleMessage(db *sql.DB, rtm *slack.RTM, msg slack.Msg) {
	if requestGifRegex.MatchString(msg.Text) {
		keyword := requestGifRegex.FindStringSubmatch(msg.Text)[1]
		gifRows, err := db.Query("SELECT url FROM gifbot_gifs WHERE keyword = ? AND _ROWID_ >= (abs(random()) % (SELECT max(_ROWID_) FROM gifbot_gifs)) LIMIT 1;", keyword)
		if err != nil {
			log.Fatalf("Could not retrieve gif: %s", err)
		}

		if gifRows.Next() == true {
			gifUrl := ""
			gifRows.Scan(&gifUrl)

			rtm.SendMessage(rtm.NewOutgoingMessage(gifUrl, msg.Channel))
		} else {
			rtm.SendMessage(rtm.NewOutgoingMessage("You haven't given me anything for that, you silly goose.", msg.Channel))
		}
		return
	}

	if storeGifRegex.MatchString(msg.Text) {
			matches := storeGifRegex.FindStringSubmatch(msg.Text)
			keyword := matches[1]
			url := matches[2]

			existingGifRows, err := db.Query("SELECT url FROM gifbot_gifs WHERE keyword = ? AND url = ?", keyword, url)
			if err != nil {
				log.Fatalf("DB communication error: %v", err)
			}

			if existingGifRows.Next() == false {
				_, err := db.Exec("INSERT INTO gifbot_gifs VALUES (?, ?, ?)", keyword, url, msg.User)
				if err != nil {
					log.Fatalf("DB communication error: %v", err)
				}
			}

			rtm.SendMessage(rtm.NewOutgoingMessage("Got it.", msg.Channel))
			return
	}

	helpRegex := regexp.MustCompile(fmt.Sprintf("^<@%s> help$", botId))
	if helpRegex.MatchString(msg.Text) {
		helpText := "Hi I'm gifbot. Supported commands:\n\n```\n.gif <keyword> Get a stored gif for a keyword\n.storegif <keyword> <url> Store a URL under a keyword\n```"
		rtm.SendMessage(rtm.NewOutgoingMessage(helpText, msg.Channel))
	}
}

func migrate(db *sql.DB) {
	existingTableRows, err := db.Query("SELECT name FROM sqlite_temp_master WHERE type='table';")
	if err != nil {
		log.Fatalf("%v", err)
	}

	// No tables exist
	if existingTableRows.Next() == false {
		db.Exec("CREATE TABLE gifbot_metadata (key text, value text);")
		db.Exec("CREATE TABLE gifbot_gifs (keyword text, url text, creator text);")
		db.Exec("CREATE INDEX idx_gifbot_gifs_keyword_url ON gifbot_gifs (keyword, url);")
		db.Exec("INSERT INTO gifbot_metadata (\"schema_version\", \"1\");")
		return
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	db, err := sql.Open("sqlite3", os.Getenv("DATABASE_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err)
	}

	// Run migrations
	migrate(db)

	// Set up slack connection
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("User ID: ", ev.Info.User.ID)
			botId = ev.Info.User.ID
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			handleMessage(db, rtm, ev.Msg)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}
