package scraper

import (
	"testing"

	"github.com/tomok/katamichi-go-bot-v3/storage"
)

func makeItem(start, carType, period, condition, tel string) CarItem {
	return CarItem{
		StartShop: start,
		CarType:   carType,
		Period:    period,
		Condition: condition,
		Tel:       tel,
		Available: true,
	}
}

func stateWith(active map[string]storage.StoredItem, messageTS map[string]string) *storage.State {
	if messageTS == nil {
		messageTS = map[string]string{}
	}
	return &storage.State{Active: active, MessageTS: messageTS}
}

var (
	itemA = makeItem("盛岡店", "ヤリス", "5月1日", "禁煙 AT", "019-606-3076")
	itemB = makeItem("仙台店", "プリウス", "5月2日", "AT", "022-111-2222")
)

func TestDetect_added(t *testing.T) {
	state := stateWith(map[string]storage.StoredItem{}, nil)
	d := Detect([]CarItem{itemA}, state)

	if len(d.Added) != 1 || d.Added[0].Key() != itemA.Key() {
		t.Errorf("want 1 added item, got %+v", d)
	}
	if len(d.Reopened)+len(d.Updated)+len(d.Removed) != 0 {
		t.Errorf("unexpected changes: %+v", d)
	}
}

func TestDetect_noChange(t *testing.T) {
	key := itemA.Key()
	state := stateWith(
		map[string]storage.StoredItem{key: toStoredItem(itemA)},
		map[string]string{key: "ts-abc"},
	)
	d := Detect([]CarItem{itemA}, state)

	if d.HasChange() {
		t.Errorf("want no change, got %+v", d)
	}
}

func TestDetect_removed(t *testing.T) {
	key := itemA.Key()
	state := stateWith(
		map[string]storage.StoredItem{key: toStoredItem(itemA)},
		map[string]string{key: "ts-abc"},
	)
	d := Detect([]CarItem{}, state)

	if len(d.Removed) != 1 || d.Removed[0] != key {
		t.Errorf("want 1 removed, got %+v", d)
	}
}

func TestDetect_reopened(t *testing.T) {
	key := itemA.Key()
	soldOut := toStoredItem(itemA)
	soldOut.Available = false
	state := stateWith(
		map[string]storage.StoredItem{key: soldOut},
		map[string]string{key: "ts-abc"},
	)
	available := itemA
	available.Available = true
	d := Detect([]CarItem{available}, state)

	if len(d.Reopened) != 1 || d.Reopened[0].Key() != key {
		t.Errorf("want 1 reopened (Available false→true), got %+v", d)
	}
	if len(d.Added) != 0 {
		t.Errorf("reopened item should not appear in Added")
	}
}

func TestDetect_updated(t *testing.T) {
	key := itemA.Key()
	state := stateWith(
		map[string]storage.StoredItem{key: toStoredItem(itemA)},
		map[string]string{key: "ts-abc"},
	)
	modified := itemA
	modified.Condition = "禁煙 AT ETC"
	d := Detect([]CarItem{modified}, state)

	if len(d.Updated) != 1 {
		t.Errorf("want 1 updated, got %+v", d)
	}
	if d.Updated[0].Condition != "禁煙 AT ETC" {
		t.Errorf("Condition = %q, want %q", d.Updated[0].Condition, "禁煙 AT ETC")
	}
}

func TestDetect_mixed(t *testing.T) {
	keyA := itemA.Key()
	keyB := itemB.Key()

	state := stateWith(
		map[string]storage.StoredItem{keyA: toStoredItem(itemA)},
		map[string]string{keyA: "ts-a", keyB: "ts-b"},
	)
	availableB := itemB
	availableB.Available = true
	itemC := makeItem("札幌店", "アクア", "6月1日", "AT", "011-222-3333")
	d := Detect([]CarItem{availableB, itemC}, state)

	if len(d.Removed) != 1 || d.Removed[0] != keyA {
		t.Errorf("Removed: want [%s], got %v", keyA, d.Removed)
	}
	if len(d.Reopened) != 0 {
		t.Errorf("Reopened: want [], got %v", d.Reopened)
	}
	if len(d.Added) != 2 {
		t.Errorf("Added: want [%s, %s], got %v", keyB, itemC.Key(), d.Added)
	}
}

func TestDetect_periodChanged(t *testing.T) {
	key := itemA.Key()
	state := stateWith(
		map[string]storage.StoredItem{key: toStoredItem(itemA)},
		map[string]string{key: "ts-abc"},
	)
	modified := itemA
	modified.Period = "6月1日"
	d := Detect([]CarItem{modified}, state)

	if len(d.Updated) != 1 {
		t.Errorf("want 1 updated for period change, got %+v", d)
	}
	if d.Updated[0].Period != "6月1日" {
		t.Errorf("Period = %q, want %q", d.Updated[0].Period, "6月1日")
	}
	if len(d.Added)+len(d.Removed)+len(d.Reopened) != 0 {
		t.Errorf("unexpected changes: %+v", d)
	}
}

func TestDetect_soldOut(t *testing.T) {
	available := itemA
	available.Available = true
	key := available.Key()
	state := stateWith(
		map[string]storage.StoredItem{key: toStoredItem(available)},
		map[string]string{key: "ts-abc"},
	)
	unavailable := available
	unavailable.Available = false
	d := Detect([]CarItem{unavailable}, state)

	if len(d.SoldOut) != 1 || d.SoldOut[0].Key() != key {
		t.Errorf("want 1 soldout, got %+v", d)
	}
	if len(d.Updated) != 0 {
		t.Errorf("soldout should not appear in Updated, got %+v", d.Updated)
	}
}

func TestDetect_soldOutNotTriggeredWhenAlreadyUnavailable(t *testing.T) {
	unavailable := itemA
	unavailable.Available = false
	key := unavailable.Key()
	state := stateWith(
		map[string]storage.StoredItem{key: toStoredItem(unavailable)},
		map[string]string{key: "ts-abc"},
	)
	d := Detect([]CarItem{unavailable}, state)

	if d.HasChange() {
		t.Errorf("want no change when already unavailable, got %+v", d)
	}
}

func TestApplyDiff(t *testing.T) {
	state := stateWith(map[string]storage.StoredItem{}, nil)
	d := Detect([]CarItem{itemA}, state)

	current := map[string]CarItem{itemA.Key(): itemA}
	ApplyDiff(state, d, current)

	if _, ok := state.Active[itemA.Key()]; !ok {
		t.Error("item should be in Active after apply")
	}
}
