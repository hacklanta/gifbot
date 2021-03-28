package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/nlopes/slack"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-redis/redis/v8"
)

var (
	requestGifRegex = regexp.MustCompile("^\\.gif ([^ ]+)$")

	storeGifRegex = regexp.MustCompile("^\\.gifstore ([^ ]+) <([^ ]+)>$")

	deleteGifRegex = regexp.MustCompile("^\\.gifdelete ([^ ]+) <([^ ]+)>$")

	attributeGifRegex = regexp.MustCompile("^\\.gifattribute ([^ ]+) <([^ ]+)>$")

	botId = ""

	helpRegex = regexp.MustCompile(".*")
)

func handleMessage(rdb *redis.Client, rtm *slack.RTM, msg slack.Msg) {
	if requestGifRegex.MatchString(msg.Text) {
		keyword := requestGifRegex.FindStringSubmatch(msg.Text)[1]
		key := "gif_" + keyword

		gifUrl, err := rdb.SRandMember(context.TODO(), key).Result()
		switch {
		case err == redis.Nil:
			rtm.SendMessage(rtm.NewOutgoingMessage("No matches for that keyword", msg.Channel))
			return

		case err != nil:
			rtm.SendMessage(rtm.NewOutgoingMessage("Internal error, attempting restart", msg.Channel))
			log.Fatalf("Error communicating with redis", err)
		}

		rtm.SendMessage(rtm.NewOutgoingMessage(gifUrl, msg.Channel))
		return
	}

	if storeGifRegex.MatchString(msg.Text) {
		matches := storeGifRegex.FindStringSubmatch(msg.Text)
		keyword := matches[1]
		url := matches[2]

		key := "gif_" + keyword

		_, err := rdb.SAdd(context.TODO(), key, url).Result()
		if err != nil {
			rtm.SendMessage(rtm.NewOutgoingMessage("Internal error, attempting restart", msg.Channel))
			log.Fatalf("Error communicating with redis", err)
		}

		rtm.SendMessage(rtm.NewOutgoingMessage("Got it.", msg.Channel))
		return
	}

	if deleteGifRegex.MatchString(msg.Text) {
		matches := deleteGifRegex.FindStringSubmatch(msg.Text)
		keyword := matches[1]
		url := matches[2]

		key := "gif_" + keyword

		_, err := rdb.SRem(context.TODO(), key, url).Result()
		if err != nil {
			rtm.SendMessage(rtm.NewOutgoingMessage("Internal error, attempting restart", msg.Channel))
			log.Fatalf("Error communicating with redis", err)
		}

		rtm.SendMessage(rtm.NewOutgoingMessage("GIF Removed.", msg.Channel))
		return
	}

	if attributeGifRegex.MatchString(msg.Text) {
		return
	}

	if helpRegex.MatchString(msg.Text) {
		helpText := "Hi I'm gifbot. Supported commands:\n" +
			"```\n" +
			".gif <keyword> Get a stored gif for a keyword\n" +
			".gifstore <keyword> <url> Store a URL under a keyword\n" +
			".gifdelete <keyword> <url> Delete a URL from a keyword\n" +
			".gifattribute <keyword> <url> Figure out who is responsible for a URL.\n" +
			"```"
		rtm.SendMessage(rtm.NewOutgoingMessage(helpText, msg.Channel))
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	rdbopt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(rdbopt)
	defer rdb.Close()

	db, err := sql.Open("sqlite3", os.Getenv("DATABASE_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err)
	}

	// Set up slack connection
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	api := slack.New(
		os.Getenv("SLACK_TOKEN"),
		slack.OptionLog(logger),
		slack.OptionDebug(true),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("User ID: ", ev.Info.User.ID)
			botId = ev.Info.User.ID
			helpRegex = regexp.MustCompile(fmt.Sprintf("^<@%s> help$", botId))
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			handleMessage(rdb, rtm, ev.Msg)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:
			// Ignore other events..
		}
	}
}
