package scraper

import (
	_ "embed"
	"encoding/json"
)

//go:embed data/car_decoration_master.json
var carDecorationMasterJSON []byte

type Decoration struct {
	Icon  string
	Color string
}

var (
	decorationMap     map[string]Decoration
	defaultDecoration Decoration
	carMatcher        *CarMatcher
)

func init() {
	var master struct {
		Decorations []struct {
			Type  string `json:"type"`
			Icon  string `json:"icon"`
			Color string `json:"color"`
		} `json:"decorations"`
	}
	if err := json.Unmarshal(carDecorationMasterJSON, &master); err != nil {
		panic(err)
	}
	decorationMap = make(map[string]Decoration, len(master.Decorations))
	for _, d := range master.Decorations {
		dec := Decoration{Icon: d.Icon, Color: d.Color}
		decorationMap[d.Type] = dec
		if d.Type == "default" {
			defaultDecoration = dec
		}
	}

	var err error
	carMatcher, err = NewCarMatcher()
	if err != nil {
		panic(err)
	}
}

func decorationFromCarType(carType string) Decoration {
	_, meta, ok := carMatcher.Identify(carType)
	if !ok {
		return defaultDecoration
	}
	if dec, ok := decorationMap[meta.Type]; ok {
		return dec
	}
	return defaultDecoration
}

func (c CarItem) Decoration() Decoration {
	return decorationFromCarType(c.CarType)
}
