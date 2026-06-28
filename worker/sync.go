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
	return syncStorage(statePath, scraper.Fetch)
}

func syncStorage(statePath string, fetch func() ([]scraper.CarItem, error)) error {
	items, err := fetch()
	if err != nil {
		return fmt.Errorf("scraper: %w", err)
	}

	state, err := storage.Load(statePath)
	if err != nil {
		return fmt.Errorf("storage load: %w", err)
	}

	prevActiveCount := len(state.Active)
	prevTSCount := len(state.MessageTS)

	currentKeys := make(map[string]bool, len(items))
	for _, item := range items {
		currentKeys[item.Key()] = true
	}

	// stateにあってサイトから消えたキーをまとめて削除
	var removedCount int
	for key := range state.Active {
		if !currentKeys[key] {
			log.Printf("sync: removing key=%s", key)
			removedCount++
		}
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
	storage.PruneMessageTS(state, currentKeys)

	if err := storage.Save(statePath, state); err != nil {
		return err
	}

	log.Printf("sync: fetched=%d removed=%d active: %d→%d ts: %d→%d",
		len(items), removedCount, prevActiveCount, len(state.Active), prevTSCount, len(state.MessageTS))
	return nil
}
