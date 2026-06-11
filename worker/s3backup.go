package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tomok/katamichi-go-bot-v3/storage"
)

func NewS3Backup(ctx context.Context) (*storage.S3Backup, error) {
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")
	if bucket == "" || region == "" {
		return nil, fmt.Errorf("S3_BUCKET or AWS_REGION not set")
	}
	return storage.NewS3Backup(ctx, bucket, region)
}

func RunS3Backup(ctx context.Context, s3b *storage.S3Backup, statePath string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s3b.Upload(ctx, statePath); err != nil {
				log.Printf("S3 backup failed: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
