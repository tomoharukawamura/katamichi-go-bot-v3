package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tomok/katamichi-go-bot-v3/notifier"
	"github.com/tomok/katamichi-go-bot-v3/worker"
)

const (
	statePath      = "/data/state.json"
	checkInterval  = 10 * time.Second
	backupInterval = 1 * time.Hour
)

func main() {
	_ = godotenv.Load()

	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		log.Fatal("SLACK_BOT_TOKEN is required")
	}

	ch, err := notifier.LoadChannelConfig()
	if err != nil {
		log.Fatalf("channel config load failed: %v", err)
	}

	slack := notifier.NewSlack(token)

	if err := worker.SyncStorage(statePath); err != nil {
		log.Printf("initial sync error: %v", err)
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("failed to load JST, using UTC: %v", err)
		jst = time.UTC
	}
	go worker.RunNightlySync(statePath, jst)

	for {
		if err := worker.Check(slack, ch, statePath); err != nil {
			log.Printf("check error: %v", err)
		}
		time.Sleep(checkInterval)
	}
}



