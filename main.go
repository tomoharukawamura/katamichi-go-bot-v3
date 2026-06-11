package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tomok/katamichi-go-bot-v3/notifier"
	"github.com/tomok/katamichi-go-bot-v3/storage"
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

	ctx := context.Background()
	slack := notifier.NewSlack(token)

	var s3b *storage.S3Backup
	if os.Getenv("APP_ENV") == "pro" {
		if b, err := worker.NewS3Backup(ctx); err != nil {
			log.Printf("S3 backup disabled: %v", err)
		} else {
			s3b = b
		}
	}

	if s3b != nil {
		if restored, err := s3b.Restore(ctx, statePath); err != nil {
			log.Printf("S3 restore failed: %v", err)
		} else if restored {
			} else {
		}
	}

	if err := worker.SyncStorage(statePath); err != nil {
		log.Printf("initial sync error: %v", err)
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("failed to load JST, using UTC: %v", err)
		jst = time.UTC
	}
	go worker.RunNightlySync(statePath, jst)

	if s3b != nil {
		go worker.RunS3Backup(ctx, s3b, statePath, backupInterval)
	}

	for {
		if err := worker.Check(slack, ch, statePath); err != nil {
			log.Printf("check error: %v", err)
		}
		time.Sleep(checkInterval)
	}
}



