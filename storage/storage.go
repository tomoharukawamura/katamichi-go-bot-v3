package storage

import (
	"encoding/json"
	"os"
)

// StoredItem holds the fields used for change detection (non-key fields of CarItem).
type StoredItem struct {
	StartShop  string `json:"start_shop"`
	ReturnShop string `json:"return_shop"`
	StartArea  string `json:"start_area"`
	ReturnArea string `json:"return_area"`
	Period     string `json:"period"`
	Condition  string `json:"condition"`
	Tel        string `json:"tel"`
	Available  bool   `json:"available"`
}

type State struct {
	Active         map[string]StoredItem `json:"active"`
	MessageTS      map[string]string     `json:"message_ts"`
	MessageChannel map[string]string     `json:"message_channel"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return empty(), nil
	}
	if err != nil {
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return empty(), nil
	}
	if s.Active == nil {
		s.Active = map[string]StoredItem{}
	}
	if s.MessageTS == nil {
		s.MessageTS = map[string]string{}
	}
	if s.MessageChannel == nil {
		s.MessageChannel = map[string]string{}
	}
	return &s, nil
}

func Save(path string, s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// PruneMessageTS removes keys from MessageTS that are no longer on the site,
// so they'll be treated as new items if they reappear.
func PruneMessageTS(s *State, currentKeys map[string]bool) {
	for key := range s.MessageTS {
		if !currentKeys[key] {
			delete(s.MessageTS, key)
		}
	}
}

func empty() *State {
	return &State{
		Active:         map[string]StoredItem{},
		MessageTS:      map[string]string{},
		MessageChannel: map[string]string{},
	}
}
