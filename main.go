package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/nlopes/slack"
	"github.com/xyproto/simplebolt"
)

var (
	requestGifRegex = regexp.MustCompile("^\\.gif ([^ ]+)$")

	storeGifRegex = regexp.MustCompile("^\\.storegif ([^ ]+) <([^ ]+)>$")

	botId = ""
)

type storedgif struct {
	Url string `json:"url"`
	Creator string `json:"creator"`
}

func handleMessage(db *simplebolt.Database, rtm *slack.RTM, msg slack.Msg) {
	if requestGifRegex.MatchString(msg.Text) {
		keyword := requestGifRegex.FindStringSubmatch(msg.Text)[1]

		setstore, err := simplebolt.NewSet(db, keyword)
		if err != nil {
			log.Fatalf("Could not retrieve set for keyword %s: %s", keyword, err)
		}

		gifs, err := setstore.GetAll()
		if err != nil {
			log.Fatalf("Could not retrieve values for keyword: %s", err)
		}

		if len(gifs) > 0 {
			rtm.SendMessage(rtm.NewOutgoingMessage(gifs[rand.Intn(len(gifs))], msg.Channel))
		} else {
			rtm.SendMessage(rtm.NewOutgoingMessage("You haven't given me anything for that, you silly goose.", msg.Channel))
		}
		return
	}

	if storeGifRegex.MatchString(msg.Text) {
		keyword := storeGifRegex.FindStringSubmatch(msg.Text)[1]
		url := storeGifRegex.FindStringSubmatch(msg.Text)[2]

		setstore, err := simplebolt.NewSet(db, keyword)
		if err != nil {
			log.Fatalf("Could not retrieve set for keyword %s: %s", keyword, err)
		}

		setstore.Add(url)
		rtm.SendMessage(rtm.NewOutgoingMessage("Got it.", msg.Channel))
		return
	}

	helpRegex := regexp.MustCompile(fmt.Sprintf("^<@%s> help$", botId))
	if helpRegex.MatchString(msg.Text) {
		helpText := "Hi I'm gifbot. Supported commands:\n\n```\n.gif <keyword> Get a stored gif for a keyword\n.storegif <keyword> <url> Store a URL under a keyword\n```"
		rtm.SendMessage(rtm.NewOutgoingMessage(helpText, msg.Channel))
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	// New bolt database
	db, err := simplebolt.New(os.Getenv("DATABASE_PATH"))
	if err != nil {
		log.Fatalf("Could not create database! %s", err)
	}
	defer db.Close()

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
