//go:build integration

package notifier

import (
	"os"
	"testing"
	"time"

	"github.com/tomok/katamichi-go-bot-v3/scraper"
	"github.com/tomok/katamichi-go-bot-v3/storage"
)

func sp(s string) *string { return &s }

func toMsg(status *string, items ...scraper.CarItem) []scraper.CarItemForMessage {
	out := make([]scraper.CarItemForMessage, len(items))
	for i, it := range items {
		out[i] = scraper.NewCarItemForMessage(it, status)
	}
	return out
}

func TestNotify_integration(t *testing.T) {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		t.Skip("SLACK_BOT_TOKEN not set")
	}

	ch, err := LoadChannelConfig()
	if err != nil {
		t.Fatalf("LoadChannelConfig: %v", err)
	}

	slack := NewSlack(token)

	// area "2"→"3" は channels.dev.json に定義済み
	added := scraper.CarItem{
		StartShop:  "トヨタレンタリース岩手 盛岡駅南口店",
		StartArea:  "2",
		ReturnShop: "トヨタレンタリース青森 返却可能店舗",
		ReturnArea: "2",
		CarType:    "ヤリスHV テスト用",
		Period:     "2026年6月1日 〜 6月7日",
		Condition:  "禁煙 AT ETC",
		Tel:        "019-606-3076",
		Available:  true,
	}

	state := &storage.State{
		Active:         map[string]storage.StoredItem{},
		MessageTS:      map[string]string{},
		MessageChannel: map[string]string{},
	}

	pause := func() { time.Sleep(5 * time.Second) }

	// --- Added ---
	t.Run("Added", func(t *testing.T) {
		d := scraper.Diff{Added: toMsg(sp("新着"), added)}
		if err := slack.Notify(d, state, ch); err != nil {
			t.Errorf("Notify(Added) error: %v", err)
		}
		if ts := state.MessageTS[added.Key()]; ts == "" {
			t.Error("MessageTS not stored after Added")
		}
		t.Logf("ts = %s", state.MessageTS[added.Key()])
	})

	pause()

	// --- Updated (スレッド返信) ---
	t.Run("Updated", func(t *testing.T) {
		updated := added
		updated.Condition = "禁煙 AT ETC バックモニター"
		d := scraper.Diff{Updated: toMsg(sp("更新"), updated)}
		if err := slack.Notify(d, state, ch); err != nil {
			t.Errorf("Notify(Updated) error: %v", err)
		}
	})

	pause()

	// --- SoldOut (リアクション付与) ---
	t.Run("SoldOut", func(t *testing.T) {
		soldOut := added
		soldOut.Available = false
		d := scraper.Diff{SoldOut: toMsg(nil, soldOut)}
		if err := slack.Notify(d, state, ch); err != nil {
			t.Errorf("Notify(SoldOut) error: %v", err)
		}
	})

	pause()

	// --- Reopened (リアクション除去 + スレッド返信) ---
	t.Run("Reopened", func(t *testing.T) {
		reopened := added
		reopened.Available = true
		d := scraper.Diff{Reopened: toMsg(sp("受付再開"), reopened)}
		if err := slack.Notify(d, state, ch); err != nil {
			t.Errorf("Notify(Reopened) error: %v", err)
		}
	})

	pause()

	// --- アルファード（minivan_luxury: color=#b8860b）---
	alphardItem := scraper.CarItem{
		StartShop:  "トヨタレンタリース岩手 盛岡駅南口店",
		StartArea:  "2",
		ReturnShop: "トヨタレンタリース青森 返却可能店舗",
		ReturnArea: "2",
		CarType:    "アルファード 品川300あ1234 テスト用",
		Period:     "2026年6月1日 〜 6月7日",
		Condition:  "禁煙 AT ETC",
		Tel:        "019-606-3076",
		Available:  true,
	}

	t.Run("Alphard_Added", func(t *testing.T) {
		d := scraper.Diff{Added: toMsg(sp("新着"), alphardItem)}
		if err := slack.Notify(d, state, ch); err != nil {
			t.Errorf("Notify(Alphard Added) error: %v", err)
		}
		if ts := state.MessageTS[alphardItem.Key()]; ts == "" {
			t.Error("MessageTS not stored after Alphard Added")
		}
		t.Logf("alphard ts = %s", state.MessageTS[alphardItem.Key()])
	})

	type itemOpts struct {
		StartShop  string
		StartArea  string
		ReturnShop string
		ReturnArea string
	}
	baseItem := func(carType string, opts ...itemOpts) scraper.CarItem {
		o := itemOpts{
			StartShop:  "トヨタレンタリース岩手 盛岡駅南口店",
			StartArea:  "2",
			ReturnShop: "トヨタレンタリース青森 返却可能店舗",
			ReturnArea: "2",
		}
		if len(opts) > 0 {
			if opts[0].StartShop != "" {
				o.StartShop = opts[0].StartShop
			}
			if opts[0].StartArea != "" {
				o.StartArea = opts[0].StartArea
			}
			if opts[0].ReturnShop != "" {
				o.ReturnShop = opts[0].ReturnShop
			}
			if opts[0].ReturnArea != "" {
				o.ReturnArea = opts[0].ReturnArea
			}
		}
		return scraper.CarItem{
			StartShop:  o.StartShop,
			StartArea:  o.StartArea,
			ReturnShop: o.ReturnShop,
			ReturnArea: o.ReturnArea,
			CarType:    carType,
			Period:     "2026年6月1日 〜 6月7日",
			Condition:  "禁煙 AT ETC",
			Tel:        "019-606-3076",
			Available:  true,
		}
	}

	specialItems := []struct {
		name string
		item scraper.CarItem
	}{
		// ["2","2"] 東北→東北: 岩手🍇→青森🍎 (prefix: GR 86)
		{"GR86_sports", baseItem("GR 86 品川300あ1234 テスト用")},
		// ["宮城","2"] 東北→東北(宮城起点): 宮城🫛→山形🍒
		{"Harrier_suv_rare", baseItem("ハリアー 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタレンタリース宮城 仙台空港店",
			ReturnShop: "トヨタレンタリース山形 返却可能店舗",
		})},
		// ["3","3"] 関東→関東: 神奈川🏖️→多摩🐎
		{"Camry_sedan_rare", baseItem("カムリ 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタレンタリース神奈川 横浜みなとみらい店",
			StartArea:  "3",
			ReturnShop: "トヨタS＆Dレンタシェア西東京 返却可能店舗",
			ReturnArea: "3",
		})},
		// ["2","3"] 東北→関東: 岩手🍇→東京🗼
		{"PriusPHEV_prius_plugin", baseItem("プリウスPHEV 品川300あ1234 テスト用", itemOpts{
			ReturnShop: "トヨタモビリティサービス 返却可能店舗",
			ReturnArea: "3",
		})},
		// ["3","静岡"] 関東→東海: 神奈川🏖️→静岡🍵
		{"Prius_prius_normal", baseItem("プリウス 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタレンタリース神奈川 横浜みなとみらい店",
			StartArea:  "3",
			ReturnShop: "トヨタレンタリース静岡 返却可能店舗",
		})},
		// ["3","愛知"] 関東→東海: 東京🗼→名古屋🏭
		{"PriusAlpha_prius_van", baseItem("プリウスα 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタモビリティサービス 品川高輪口店",
			StartArea:  "3",
			ReturnShop: "トヨタレンタリース名古屋 返却可能店舗",
		})},
		// ["3","5"] 関東→関西: 成田✈️→大阪🏯
		{"Crown_crown", baseItem("クラウン 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタモビリティサービス 成田空港店",
			StartArea:  "3",
			ReturnShop: "トヨタレンタリース大阪 返却可能店舗",
			ReturnArea: "5",
		})},
		// ["4","5"] 東海→関西: 静岡🍵→京都⛩️
		{"LandCruiserPrado_randcruiser", baseItem("ランドクルーザープラド 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタレンタリース静岡 静岡駅前店",
			StartArea:  "4",
			ReturnShop: "トヨタレンタリース京都 返却可能店舗",
			ReturnArea: "5",
		})},
		// ["8","8"] 九州→九州: 福岡🍜→鹿児島🌋
		{"Hiace_van_normal", baseItem("ハイエース 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタレンタリース福岡 博多駅前店",
			StartArea:  "8",
			ReturnShop: "トヨタレンタリース鹿児島 返却可能店舗",
			ReturnArea: "8",
		})},
		// ["2","2"] 東北→東北: 山形🍒→福島🍑
		{"HiaceGC_minivan_rare", baseItem("ハイエースグランドキャビン 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタレンタリース山形 山形駅前店",
			ReturnShop: "トヨタレンタリース福島 返却可能店舗",
		})},
		// ["8","8"] 九州→九州: 長崎⛪️→大分🐻 (prefix: ダイハツタント)
		{"Tanto_light", baseItem("ダイハツタント 品川300あ1234 テスト用", itemOpts{
			StartShop:  "トヨタレンタリース長崎 長崎駅前店",
			StartArea:  "8",
			ReturnShop: "トヨタレンタリース大分 返却可能店舗",
			ReturnArea: "8",
		})},
	}

	for _, tc := range specialItems {
		tc := tc
		t.Run(tc.name+"_Added", func(t *testing.T) {
			t.Parallel()
			localState := &storage.State{
				Active:         map[string]storage.StoredItem{},
				MessageTS:      map[string]string{},
				MessageChannel: map[string]string{},
			}
			d := scraper.Diff{Added: toMsg(sp("新着"), tc.item)}
			if err := slack.Notify(d, localState, ch); err != nil {
				t.Errorf("Notify(%s Added) error: %v", tc.name, err)
			}
			if ts := localState.MessageTS[tc.item.Key()]; ts == "" {
				t.Errorf("MessageTS not stored after %s Added", tc.name)
			}
			t.Logf("%s ts = %s", tc.name, localState.MessageTS[tc.item.Key()])
		})
	}

	pause()

	// --- チャンネル変化シナリオ ---
	// 新着時に誤ったエリアで投稿 → 正しいエリアに修正されUpdatedとして検知
	// 期待: 旧チャンネルにSoldOutリアクション（ベストエフォート）+ 新チャンネルに新着投稿
	t.Run("ChannelChanged_Updated", func(t *testing.T) {
		// 1. 誤ったエリア(area "2"→"2")でAddedとして通知
		wrongAreaItem := baseItem("プリウス テスト用（チャンネル変化）")
		addedDiff := scraper.Diff{Added: toMsg(sp("新着"), wrongAreaItem)}
		if err := slack.Notify(addedDiff, state, ch); err != nil {
			t.Fatalf("Notify(Added wrong area) error: %v", err)
		}
		oldTS := state.MessageTS[wrongAreaItem.Key()]
		oldChannel := state.MessageChannel[wrongAreaItem.Key()]
		if oldTS == "" {
			t.Fatal("MessageTS not stored after Added")
		}
		t.Logf("old ts = %s, old channel = %s", oldTS, oldChannel)

		pause()

		// 2. 正しいエリア(area "2"→"3")に修正されたアイテムがUpdatedとして検知
		correctAreaItem := baseItem("プリウス テスト用（チャンネル変化）", itemOpts{
			ReturnArea: "3",
			ReturnShop: "トヨタモビリティサービス 返却可能店舗",
		})
		updatedDiff := scraper.Diff{Updated: toMsg(sp("更新"), correctAreaItem)}
		if err := slack.Notify(updatedDiff, state, ch); err != nil {
			t.Errorf("Notify(Updated channel changed) error: %v", err)
		}
		newTS := state.MessageTS[wrongAreaItem.Key()]
		newChannel := state.MessageChannel[wrongAreaItem.Key()]
		if newTS == "" {
			t.Error("MessageTS not stored after channel-changed Updated")
		}
		if newTS == oldTS {
			t.Error("MessageTS should be updated to new message after channel change")
		}
		if newChannel == oldChannel {
			t.Error("MessageChannel should be updated to correct channel after channel change")
		}
		t.Logf("new ts = %s, new channel = %s", newTS, newChannel)
	})
}
