package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"strconv"
	"strings"
)

/*

#ledhouse 4b green 1

#ledhouse <[1-4]><a|b|c> <red|green|blue> <[0-10]>

*/

const (
	gcpProjectEnvName  = "GOOGLE_CLOUD_PROJECT"
	pubsubTopicEnvName = "PUBSUB_TOPIC"
)

var topic *pubsub.Topic
var tokens map[string]string

type LedHouse struct {
	Room  string
	Color string
	Value string
}

type LightRequest struct {
	Token     string  `json:"token"`
	Intensity float32 `json:"intensity"`
}

func processCommand(cmd *LedHouse) {
	token := getToken(cmd.Room, cmd.Color)
	if token != "" {
		i, err := strconv.Atoi(cmd.Value)
		if err != nil {
			return
		}

		if i < 0 {
			return
		}

		if i > 10 {
			i = 10
		}

		intensity := float32(i) / 10.0

		json, err := json.Marshal(LightRequest{
			Token:     token,
			Intensity: intensity,
		})

		if topic != nil {
			ctx := context.Background()
			result := topic.Publish(ctx, &pubsub.Message{Data: json})
			serverID, err := result.Get(ctx)
			if err != nil {
				log.Printf("Failed to publish: %v", err)
				return
			}
			log.Printf("Published message ID=%s", serverID)
		}
	}
}

func parseTweet(tweet *twitter.Tweet) {
	fmt.Println(tweet.Text)

	var cmd *LedHouse
	t := strings.ToLower(tweet.Text)
	words := strings.Fields(t)
	for i, w := range words {
		if w == "#ledhouse" {
			if len(words) > i+3 {
				cmd = &LedHouse{
					Room:  words[i+1],
					Color: words[i+2],
					Value: words[i+3],
				}
			}
			break
		}
	}

	if cmd != nil {
		processCommand(cmd)
	}
}

func configure() {
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	oauthToken := os.Getenv("TWITTER_TOKEN")
	oauthSecret := os.Getenv("TWITTER_TOKEN_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(oauthToken, oauthSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	demux := twitter.NewSwitchDemux()

	demux.Tweet = parseTweet

	fmt.Println("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"#ledhouse"},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	fmt.Println("Stopping Stream...")
	stream.Stop()

}

func prepPubSub() {
	ctx := context.Background()

	projectID := os.Getenv(gcpProjectEnvName)
	if projectID == "" {
		log.Fatalf("Couldn't find %s in env", gcpProjectEnvName)
	}

	topicName := os.Getenv(pubsubTopicEnvName)
	if topicName == "" {
		log.Fatalf("Couldn't find %s in env", pubsubTopicEnvName)
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Println("Created client")

	topic = client.Topic(topicName)

	// The topic existence test requires the binding to have the 'viewer' role.
	ok, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Error finding topic: %v", err)
	}
	if !ok {
		log.Fatalf("Couldn't find topic %v", topic)
	}
}

func key(room, color string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(room), strings.ToLower(color))
}

func setToken(room, color, token string) {
	tokens[key(room, color)] = token
}

func getToken(room, color string) string {
	return tokens[key(room, color)]
}

func prepTokens() {
	tokens = make(map[string]string, 30)

	rooms := []string{
		"1a", "1b", "1c",
		"2a", "2b", "2c",
		"3a", "3b", "3c",
		"4a",
	}
	colors := []string{"red", "green", "blue"}

	for _, r := range rooms {
		for _, c := range colors {
			envName := fmt.Sprintf("TOKEN_%s_%s", strings.ToUpper(r), strings.ToUpper(c))
			token := os.Getenv(envName)
			if token != "" {
				setToken(r, c, token)
			}
		}
	}
	log.Printf("%+v", tokens)
}

func main() {
	fmt.Println("LedHouse Twitter Bot")

	prepTokens()
	prepPubSub()

	configure()
}
