package scraper

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed data/area_shop_master.json
var areaShopMasterJSON []byte

var companyPrefixes = []string{
	"トヨタS＆Dレンタシェア西東京",
	"トヨタS＆Dレンタシェア",
	"トヨタモビリティサービス",
	"トヨタレンタリース",
}

type areaEntry struct {
	cityName          string
	icon              string
	returnShopURL     string
	shopCityException map[string]string
	shopIconException map[string]string
}

// areaMap は会社名 → areaEntry のルックアップテーブル
var areaMap map[string]areaEntry

func init() {
	var master struct {
		Areas []struct {
			Name              string            `json:"name"`
			CityName          string            `json:"city_name"`
			Icon              string            `json:"icon"`
			ReturnShopURL     string            `json:"return_shop_url"`
			ShopCityException map[string]string `json:"shop_city_exception"`
			ShopIconException map[string]string `json:"icon_exception"`
		} `json:"areas"`
	}
	if err := json.Unmarshal(areaShopMasterJSON, &master); err != nil {
		panic("area_shop_master.json parse error: " + err.Error())
	}
	areaMap = make(map[string]areaEntry, len(master.Areas))
	for _, a := range master.Areas {
		areaMap[a.Name] = areaEntry{
			cityName:          a.CityName,
			icon:              a.Icon,
			returnShopURL:     a.ReturnShopURL,
			shopCityException: a.ShopCityException,
			shopIconException: a.ShopIconException,
		}
	}
}

// startCityFromShop は出発店舗名から都市名を返す。
// 会社名を area_shop_master.json で引き、shop_city_exception があれば優先する。
func startCityFromShop(shop string) string {
	parts := strings.Fields(shop)
	if len(parts) == 0 {
		return ""
	}
	entry, ok := areaMap[parts[0]]
	if !ok {
		return cityFromShop(shop)
	}
	if len(parts) >= 2 {
		storeName := parts[len(parts)-1]
		if city, ok := entry.shopCityException[storeName]; ok {
			return city
		}
	}
	return entry.cityName
}

// returnShopURLFromShop は返却店舗名から店舗一覧URLを返す。
func returnShopURLFromShop(shop string) string {
	parts := strings.Fields(shop)
	if len(parts) == 0 {
		return ""
	}
	if entry, ok := areaMap[parts[0]]; ok {
		return entry.returnShopURL
	}
	return ""
}

// returnCityFromShop は返却店舗名から都市名を返す。
// 「トヨタレンタリース〇〇 返却可能店舗」のようなサフィックスがある場合も parts[0] で対応する。
func returnCityFromShop(shop string) string {
	parts := strings.Fields(shop)
	if len(parts) == 0 {
		return ""
	}
	if entry, ok := areaMap[parts[0]]; ok {
		return entry.cityName
	}
	return cityFromShop(shop)
}

// iconFromShop は店舗名から都市アイコンを返す。未登録の場合は空文字。
func iconFromShop(shop string) string {
	parts := strings.Fields(shop)
	if len(parts) == 0 {
		return ""
	}
	entry, ok := areaMap[parts[0]]
	if !ok {
		return ""
	}
	if len(parts) >= 2 {
		storeName := parts[len(parts)-1]
		if icon, ok := entry.shopIconException[storeName]; ok {
			return icon
		}
	}
	return entry.icon
}

// cityFromShop はフォールバック用。会社名プレフィックスを除いた店舗名を返す。
func cityFromShop(shop string) string {
	parts := strings.Fields(shop)
	if len(parts) == 0 {
		return ""
	}
	if len(parts) >= 2 {
		return strings.TrimSuffix(parts[len(parts)-1], "店")
	}
	s := parts[0]
	for _, prefix := range companyPrefixes {
		if strings.HasPrefix(s, prefix) {
			return strings.TrimSuffix(strings.TrimPrefix(s, prefix), "店")
		}
	}
	return s
}
