package storage

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const s3Key = "state.json"

type S3Backup struct {
	client *s3.Client
	bucket string
}

func NewS3Backup(ctx context.Context, bucket, region string) (*S3Backup, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	return &S3Backup{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

func (b *S3Backup) Upload(ctx context.Context, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	_, err = b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(s3Key),
		Body:   f,
	})
	return err
}

// Restore downloads state.json from S3 to localPath.
// Returns (false, nil) if the object does not exist yet.
func (b *S3Backup) Restore(ctx context.Context, localPath string) (bool, error) {
	resp, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, fmt.Errorf("s3 get: %w", err)
	}
	defer resp.Body.Close()

	f, err := os.Create(localPath)
	if err != nil {
		return false, fmt.Errorf("create: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := f.Write(buf[:n]); writeErr != nil {
				return false, writeErr
			}
		}
		if readErr != nil {
			break
		}
	}
	return true, nil
}
