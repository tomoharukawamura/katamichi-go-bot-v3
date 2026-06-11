package worker

import (
	"fmt"
	"log"
	"time"

	"github.com/tomok/katamichi-go-bot-v3/scraper"
	"github.com/tomok/katamichi-go-bot-v3/storage"
)

func RunNightlySync(statePath string, jst *time.Location) {
	for {
		now := time.Now().In(jst)
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, jst)
		time.Sleep(time.Until(next))
		if err := SyncStorage(statePath); err != nil {
			log.Printf("nightly sync error: %v", err)
		}
	}
}

func SyncStorage(statePath string) error {
	items, err := scraper.Fetch()
	if err != nil {
		return fmt.Errorf("scraper: %w", err)
	}

	state, err := storage.Load(statePath)
	if err != nil {
		return fmt.Errorf("storage load: %w", err)
	}

	// Active をサイトの現状で完全に再構築
	newActive := make(map[string]storage.StoredItem, len(items))
	for _, item := range items {
		key := item.Key()
		if existing, ok := state.Active[key]; ok {
			newActive[key] = existing
		} else {
			newActive[key] = scraper.StoredItemFrom(item)
		}
	}
	state.Active = newActive

	// 現在サイトにないキーを MessageTS から除去
	currentKeys := make(map[string]bool, len(items))
	for _, item := range items {
		currentKeys[item.Key()] = true
	}
	storage.PruneMessageTS(state, currentKeys)

	return storage.Save(statePath, state)
}
