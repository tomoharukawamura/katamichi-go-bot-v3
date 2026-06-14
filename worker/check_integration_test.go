//go:build integration

package worker

import (
	"os"
	"testing"

	"github.com/tomok/katamichi-go-bot-v3/notifier"
	"github.com/tomok/katamichi-go-bot-v3/scraper"
	"github.com/tomok/katamichi-go-bot-v3/storage"
)

func TestCheck_integration(t *testing.T) {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		t.Skip("SLACK_BOT_TOKEN not set")
	}

	// 1. スクレイピングで現在の全データを取得
	items, _, err := scraper.Fetch()
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(items) == 0 {
		t.Skip("no items fetched from site")
	}
	t.Logf("fetched %d items", len(items))

	// 2. 受付中アイテムを最大5件選ぶ → Added 対象
	var available []scraper.CarItem
	for _, item := range items {
		if item.Available {
			available = append(available, item)
		}
	}
	if len(available) == 0 {
		t.Skip("no available items on site")
	}
	addCount := min(5, len(available))
	toAdd := available[:addCount]

	toAddKeys := make(map[string]bool, addCount)
	for _, item := range toAdd {
		toAddKeys[item.Key()] = true
	}

	// 3. 全アイテムをstateに登録（差分なしの初期状態）
	state := &storage.State{
		Active:         make(map[string]storage.StoredItem, len(items)),
		MessageTS:      make(map[string]string, len(items)),
		MessageChannel: make(map[string]string, len(items)),
	}
	for _, item := range items {
		state.Active[item.Key()] = scraper.StoredItemFrom(item)
		state.MessageTS[item.Key()] = "dummy_ts"
		state.MessageChannel[item.Key()] = "DUMMY_CHANNEL"
	}

	// 4. toAdd の5件をstateから削除 → Added diff になる
	for _, item := range toAdd {
		delete(state.Active, item.Key())
		delete(state.MessageTS, item.Key())
		delete(state.MessageChannel, item.Key())
	}

	// 5. toAdd以外の受付中アイテムから最大5件のConditionを書き換える → Updated diff になる
	var toUpdate []scraper.CarItem
	for _, item := range items {
		if !toAddKeys[item.Key()] && item.Available && len(toUpdate) < 5 {
			toUpdate = append(toUpdate, item)
		}
	}
	for _, item := range toUpdate {
		stored := state.Active[item.Key()]
		stored.Condition = stored.Condition + " テスト変更"
		state.Active[item.Key()] = stored
	}
	t.Logf("Added: %d items, Updated: %d items", len(toAdd), len(toUpdate))

	// 6. 加工済みstateをtempファイルに保存
	tmpDir := t.TempDir()
	tmpPath := tmpDir + "/state.json"
	if err := storage.Save(tmpPath, state); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// 7. customScan: 取得済みアイテムを固定して使い、stateはtmpPathから読む
	fixedItems := items
	customScan := func(path string) (scraper.Diff, *storage.State, []scraper.CarItem, error) {
		s, err := storage.Load(path)
		if err != nil {
			return scraper.Diff{}, nil, nil, err
		}
		diff := scraper.Detect(fixedItems, s)
		return diff, s, fixedItems, nil
	}

	ch, err := notifier.LoadChannelConfig()
	if err != nil {
		t.Fatalf("LoadChannelConfig: %v", err)
	}
	slack := notifier.NewSlack(token)

	// 8. runCheck 実行
	if err := runCheck(slack, ch, tmpPath, customScan); err != nil {
		t.Fatalf("runCheck: %v", err)
	}

	// 9. 検証: Added 5件のMessageTSが記録されているか
	finalState, err := storage.Load(tmpPath)
	if err != nil {
		t.Fatalf("Load final state: %v", err)
	}
	for _, item := range toAdd {
		ts := finalState.MessageTS[item.Key()]
		ch := finalState.MessageChannel[item.Key()]
		if ts == "" {
			t.Errorf("MessageTS not stored for Added item: %s", item.Key())
		} else {
			t.Logf("Added: %s → ts=%s channel=%s", item.Key(), ts, ch)
		}
	}
	for _, item := range toUpdate {
		// Updated はスレッド返信なのでTSは変わらない（dummy_tsのまま）が、エラーなく処理されたことを確認
		t.Logf("Updated: %s (thread reply sent)", item.Key())
	}
}
