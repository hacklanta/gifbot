package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"math/rand"
  "time"

	"github.com/nlopes/slack"
	"github.com/xyproto/simplebolt"
)

var (
	requestGifRegex = regexp.MustCompile("^\\.gif ([^ ]+)$")

	storeGifRegex = regexp.MustCompile("^\\.storegif ([^ ]+) <([^ ]+)>$")

	botId = ""
)

func handleMessage(db *simplebolt.Database, rtm *slack.RTM, messageText string, channel string) {
	if (requestGifRegex.MatchString(messageText)) {
		keyword := requestGifRegex.FindStringSubmatch(messageText)[1]

		setstore, err := simplebolt.NewSet(db, keyword)
		if (err != nil) {
			log.Fatalf("Could not retrieve set for keyword %s: %s", keyword, err)
		}

		gifs, err := setstore.GetAll()
		if (err != nil) {
			log.Fatalf("Could not retrieve values for keyword: %s", err)
		}

		if (len(gifs) > 0) {
			rtm.SendMessage(rtm.NewOutgoingMessage(gifs[rand.Intn(len(gifs))], channel))
		} else {
			rtm.SendMessage(rtm.NewOutgoingMessage("You haven't given me anything for that, you silly goose.", channel))
		}
		return
	}

	if (storeGifRegex.MatchString(messageText)) {
		keyword := storeGifRegex.FindStringSubmatch(messageText)[1]
		url := storeGifRegex.FindStringSubmatch(messageText)[2]

		setstore, err := simplebolt.NewSet(db, keyword)
		if (err != nil) {
			log.Fatalf("Could not retrieve set for keyword %s: %s", keyword, err)
		}

		setstore.Add(url)
		rtm.SendMessage(rtm.NewOutgoingMessage("Got it.", channel))
		return
	}

	helpRegex := regexp.MustCompile(fmt.Sprintf("^<@%s> help$", botId))
	if (helpRegex.MatchString(messageText)) {
		helpText := "Hi I'm gifbot. Supported commands:\n\n```\n.gif <keyword> Get a stored gif for a keyword\n.storegif <keyword> <url> Store a URL under a keyword\n```"
		rtm.SendMessage(rtm.NewOutgoingMessage(helpText, channel))
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
			handleMessage(db, rtm, ev.Msg.Text, ev.Msg.Channel)

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