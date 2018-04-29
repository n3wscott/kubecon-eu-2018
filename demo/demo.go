package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"encoding/json"
	"strconv"
	"net/url"
)

const (
	port               = "8080"
	gcpProjectEnvName  = "GOOGLE_CLOUD_PROJECT"
	pubsubTopicEnvName = "PUBSUB_TOPIC"
	tokenEnvName       = "TOKEN"
)

var topic *pubsub.Topic
var token string

func main() {
	ctx := context.Background()

	token = os.Getenv(tokenEnvName)

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
	defer topic.Stop()

	http.HandleFunc("/", getHandler)
	http.HandleFunc("/publish", postHandler)

	log.Println("Listening on port:", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

type LightRequest struct {
	Token     string  `json:"token"`
	Intensity float32 `json:"intensity"`
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	formIntensity := r.FormValue("intensity")

	i, err := strconv.Atoi(formIntensity)

	intensity := float32(i) / 100.0

	json, err := json.Marshal(LightRequest{
		Token:     token,
		Intensity: intensity,
	})

	result := topic.Publish(ctx, &pubsub.Message{Data: json})
	serverID, err := result.Get(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Published message ID=%s", serverID)
	url := fmt.Sprintf("/?i=%s", formIntensity)
	http.Redirect(w, r, url, http.StatusFound)
}

var body = `<!doctype html>
	<form method="POST" action="/publish">
		<input type="range" min="1" max="100" value="%s" class="slider" name="intensity">
		<input type="submit" value="Set">
	</form>
</html>
`

func getHandler(w http.ResponseWriter, r *http.Request) {
	value := "0"
	m, _ := url.ParseQuery(r.URL.RawQuery)
	if m != nil && len(m["i"]) > 0 && m["i"][0] != "" {
		value = m["i"][0]
	}
	fmt.Fprintf(w, body, value)
}
