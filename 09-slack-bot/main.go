package main

import (
	"log"
	"os"

	"github.com/Krognol/go-wolfram"
	"github.com/christianrondeau/go-wit"
	"github.com/nlopes/slack"
)

const confidenceThreshold = 0.5

var (
	witClient     *wit.Client
	slackClient   *slack.Client
	wolframClient *wolfram.Client
)

func main() {
	witClient = wit.NewClient(os.Getenv("WIT_AI_ACCESS_TOKEN"))
	slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
	wolframClient = &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	rtm := slackClient.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if len(ev.BotID) == 0 {
				go handleMessage(ev)
			}
		}
	}
}

func handleMessage(ev *slack.MessageEvent) {
	result, err := witClient.Message(ev.Msg.Text)
	if err != nil {
		log.Printf("unable to get wit.ai result: %v", err)
		return
	}

	var (
		topEntity    *wit.MessageEntity
		topEntityKey string
	)

	for key, entityList := range result.Entities {
		for _, entity := range entityList {
			moreConfident := topEntity == nil || entity.Confidence > topEntity.Confidence
			if entity.Confidence > confidenceThreshold && moreConfident {
				topEntity = &entity
				topEntityKey = key
			}
		}
	}

	switch topEntityKey {
	case "greetings":
		slackClient.PostMessage(ev.User, "Hello user! How can I help you?", slack.PostMessageParameters{
			AsUser: true,
		})
		return
	case "wolfram_search_query":
		res, err := wolframClient.GetSpokentAnswerQuery(topEntity.Value.(string), wolfram.Metric, 1000)
		if err == nil {
			slackClient.PostMessage(ev.User, res, slack.PostMessageParameters{
				AsUser: true,
			})
			return
		}

		log.Printf("unable to get data from wolfram: %v", err)
	}

	slackClient.PostMessage(ev.User, "¯\\_(o_o)_/¯", slack.PostMessageParameters{
		AsUser: true,
	})
}
