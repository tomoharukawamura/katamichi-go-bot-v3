package notifier

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed data/channels.dev.json
var channelsDev []byte

//go:embed data/channels.sta.json
var channelsSta []byte

//go:embed data/channels.pro.json
var channelsPro []byte

type channelEntry struct {
	ID      string    `json:"id"`
	Section [2]string `json:"section"`
}

type ChannelConfig struct {
	index map[[2]string]string
}

func LoadChannelConfig() (*ChannelConfig, error) {
	env := os.Getenv("APP_ENV")
	var data []byte
	switch env {
	case "pro":
		data = channelsPro
	case "sta":
		data = channelsSta
	default:
		data = channelsDev
	}

	var raw struct {
		Channels []channelEntry `json:"channels"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("channels config parse failed: %w", err)
	}

	index := make(map[[2]string]string, len(raw.Channels))
	for _, c := range raw.Channels {
		index[c.Section] = c.ID
	}
	return &ChannelConfig{index: index}, nil
}

func (c *ChannelConfig) ChannelFor(startArea, returnArea, startCity, returnCity string) (string, bool) {
	// (startCity|startArea) × (returnCity|returnArea) の全組み合わせを試す。
	// starts/returns の先頭が都市名なので、都市名を含む section エントリが先にマッチする。
	starts := unique(startCity, startArea)
	returns := unique(returnCity, returnArea)

	for _, s := range starts {
		for _, r := range returns {
			if id, ok := c.index[[2]string{s, r}]; ok {
				return id, true
			}
			if id, ok := c.index[[2]string{r, s}]; ok {
				return id, true
			}
		}
	}
	return "", false
}

// unique は重複を除いた非空の文字列スライスを返す（先頭優先）。
func unique(vals ...string) []string {
	seen := make(map[string]bool, len(vals))
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
