package scraper

import (
	_ "embed"
	"encoding/json"
)

//go:embed data/area_sub_group.json
var areaSubGroupJSON []byte

var cityToGroup map[string]string

func init() {
	var data struct {
		Groups []struct {
			Name   string   `json:"name"`
			Cities []string `json:"cities"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(areaSubGroupJSON, &data); err != nil {
		panic(err)
	}
	cityToGroup = make(map[string]string)
	for _, g := range data.Groups {
		for _, city := range g.Cities {
			cityToGroup[city] = g.Name
		}
	}
}

// GroupForCity はcity_nameからエリアサブグループ名を返す。
// 該当するグループがない場合は ("", false) を返す。
func GroupForCity(cityName string) (group string, ok bool) {
	group, ok = cityToGroup[cityName]
	return
}
