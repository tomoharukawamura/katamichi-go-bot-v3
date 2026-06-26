package worker

import (
	"errors"
	"testing"

	"github.com/tomok/katamichi-go-bot-v3/scraper"
	"github.com/tomok/katamichi-go-bot-v3/storage"
)

func makeCarItem(carType string) scraper.CarItem {
	return scraper.CarItem{
		StartShop: "盛岡店",
		CarType:   carType,
		Period:    "5月1日",
		Condition: "AT",
		Tel:       "019-606-3076",
		Available: true,
	}
}

func saveState(t *testing.T, path string, state *storage.State) {
	t.Helper()
	if err := storage.Save(path, state); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

func loadState(t *testing.T, path string) *storage.State {
	t.Helper()
	s, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return s
}

func tmpPath(t *testing.T) string {
	t.Helper()
	return t.TempDir() + "/state.json"
}

func TestSyncStorage_removesStaleKeys(t *testing.T) {
	itemA := makeCarItem("ヤリス")
	itemB := makeCarItem("プリウス")

	path := tmpPath(t)
	saveState(t, path, &storage.State{
		Active: map[string]storage.StoredItem{
			itemA.Key(): scraper.StoredItemFrom(itemA),
			itemB.Key(): scraper.StoredItemFrom(itemB),
		},
		MessageTS:      map[string]string{itemA.Key(): "ts-a", itemB.Key(): "ts-b"},
		MessageChannel: map[string]string{itemA.Key(): "C1", itemB.Key(): "C2"},
	})

	// itemB だけ返す → itemA が削除対象
	if err := syncStorage(path, func() ([]scraper.CarItem, error) {
		return []scraper.CarItem{itemB}, nil
	}); err != nil {
		t.Fatalf("syncStorage: %v", err)
	}

	s := loadState(t, path)
	if _, ok := s.Active[itemA.Key()]; ok {
		t.Errorf("itemA should be removed from Active")
	}
	if _, ok := s.Active[itemB.Key()]; !ok {
		t.Errorf("itemB should remain in Active")
	}
	if _, ok := s.MessageTS[itemA.Key()]; ok {
		t.Errorf("itemA MessageTS should be pruned")
	}
	if s.MessageTS[itemB.Key()] != "ts-b" {
		t.Errorf("itemB MessageTS should be preserved")
	}
}

func TestSyncStorage_preservesExistingStoredItem(t *testing.T) {
	item := makeCarItem("ヤリス")
	stored := scraper.StoredItemFrom(item)
	stored.Condition = "AT ETC" // stateに保存済みの値

	path := tmpPath(t)
	saveState(t, path, &storage.State{
		Active:         map[string]storage.StoredItem{item.Key(): stored},
		MessageTS:      map[string]string{item.Key(): "ts-x"},
		MessageChannel: map[string]string{},
	})

	// サイト側ではConditionが変わっている
	fetched := item
	fetched.Condition = "AT"
	if err := syncStorage(path, func() ([]scraper.CarItem, error) {
		return []scraper.CarItem{fetched}, nil
	}); err != nil {
		t.Fatalf("syncStorage: %v", err)
	}

	s := loadState(t, path)
	if s.Active[item.Key()].Condition != "AT ETC" {
		t.Errorf("existing stored value should be preserved, got %q", s.Active[item.Key()].Condition)
	}
}

func TestSyncStorage_addsNewItems(t *testing.T) {
	item := makeCarItem("アクア")

	path := tmpPath(t)
	saveState(t, path, &storage.State{
		Active:         map[string]storage.StoredItem{},
		MessageTS:      map[string]string{},
		MessageChannel: map[string]string{},
	})

	if err := syncStorage(path, func() ([]scraper.CarItem, error) {
		return []scraper.CarItem{item}, nil
	}); err != nil {
		t.Fatalf("syncStorage: %v", err)
	}

	s := loadState(t, path)
	if _, ok := s.Active[item.Key()]; !ok {
		t.Errorf("new item should be added to Active")
	}
}

func TestSyncStorage_emptyFetch(t *testing.T) {
	item := makeCarItem("ヤリス")

	path := tmpPath(t)
	saveState(t, path, &storage.State{
		Active:         map[string]storage.StoredItem{item.Key(): scraper.StoredItemFrom(item)},
		MessageTS:      map[string]string{item.Key(): "ts-a"},
		MessageChannel: map[string]string{item.Key(): "C1"},
	})

	// サイトが空になった → 全件削除
	if err := syncStorage(path, func() ([]scraper.CarItem, error) {
		return []scraper.CarItem{}, nil
	}); err != nil {
		t.Fatalf("syncStorage: %v", err)
	}

	s := loadState(t, path)
	if len(s.Active) != 0 {
		t.Errorf("Active should be empty, got %d items", len(s.Active))
	}
	if len(s.MessageTS) != 0 {
		t.Errorf("MessageTS should be empty, got %d entries", len(s.MessageTS))
	}
}

func TestSyncStorage_fetchError(t *testing.T) {
	path := tmpPath(t)
	saveState(t, path, &storage.State{
		Active:         map[string]storage.StoredItem{},
		MessageTS:      map[string]string{},
		MessageChannel: map[string]string{},
	})

	fetchErr := errors.New("network error")
	err := syncStorage(path, func() ([]scraper.CarItem, error) {
		return nil, fetchErr
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, fetchErr) {
		t.Errorf("expected fetchErr, got %v", err)
	}
}
