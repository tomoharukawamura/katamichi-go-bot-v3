package scraper

import (
	"fmt"

	"github.com/tomok/katamichi-go-bot-v3/storage"
)

// Scan fetches the current listings, loads state, and detects changes in one step.
func Scan(statePath string) (Diff, *storage.State, []CarItem, error) {
	items, err := Fetch()
	if err != nil {
		return Diff{}, nil, nil, fmt.Errorf("scraper: %w", err)
	}
	state, err := storage.Load(statePath)
	if err != nil {
		return Diff{}, nil, nil, fmt.Errorf("storage load: %w", err)
	}
	return Detect(items, state), state, items, nil
}

type Diff struct {
	Added    []CarItem
	Reopened []CarItem
	Updated  []CarItem
	SoldOut  []CarItem // Available: true → false
	Removed  []string  // CarItem.Key()
}

func (d Diff) HasChange() bool {
	return len(d.Added)+len(d.Reopened)+len(d.Updated)+len(d.SoldOut)+len(d.Removed) > 0
}

func Detect(items []CarItem, state *storage.State) Diff {
	var d Diff

	current := make(map[string]CarItem, len(items))
	for _, item := range items {
		current[item.Key()] = item
	}

	for key, item := range current {
		stored, wasActive := state.Active[key]
		_, wasSeen := state.MessageTS[key]

		switch {
		case !wasSeen:
			d.Added = append(d.Added, item)
		case !wasActive:
			d.Reopened = append(d.Reopened, item)
		case stored.Available && !item.Available:
			d.SoldOut = append(d.SoldOut, item)
		case toStoredItem(item) != stored:
			d.Updated = append(d.Updated, item)
		}
	}

	for key := range state.Active {
		if _, exists := current[key]; !exists {
			d.Removed = append(d.Removed, key)
		}
	}

	return d
}

func ApplyDiff(state *storage.State, d Diff, current map[string]CarItem) {
	for _, item := range d.Added {
		state.Active[item.Key()] = toStoredItem(item)
	}
	for _, item := range d.Reopened {
		state.Active[item.Key()] = toStoredItem(item)
	}
	for _, item := range d.Updated {
		state.Active[item.Key()] = toStoredItem(item)
	}
	for _, item := range d.SoldOut {
		state.Active[item.Key()] = toStoredItem(item)
	}
	for _, key := range d.Removed {
		delete(state.Active, key)
	}
}

func StoredItemFrom(c CarItem) storage.StoredItem {
	return toStoredItem(c)
}

func toStoredItem(c CarItem) storage.StoredItem {
	return storage.StoredItem{
		StartShop:  c.StartShop,
		ReturnShop: c.ReturnShop,
		StartArea:  c.StartArea,
		ReturnArea: c.ReturnArea,
		Period:     c.Period,
		Condition:  c.Condition,
		Tel:        c.Tel,
		Available:  c.Available,
	}
}
