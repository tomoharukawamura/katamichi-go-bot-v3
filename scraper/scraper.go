package scraper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const targetURL = "https://cp.toyota.jp/rentacar/?padid=ag270_fr_sptop_onewayma"

type CarItem struct {
	StartShop  string
	StartArea  string
	ReturnShop string
	ReturnArea string
	CarType    string
	Condition  string
	Period     string
	Tel        string
	Available  bool
}

func (c CarItem) StartCity() string     { return startCityFromShop(c.StartShop) }
func (c CarItem) ReturnCity() string    { return returnCityFromShop(c.ReturnShop) }
func (c CarItem) StartCityIcon() string { return iconFromShop(c.StartShop) }
func (c CarItem) ReturnCityIcon() string { return iconFromShop(c.ReturnShop) }
func (c CarItem) StartCityGroup() string  { g, _ := GroupForCity(c.StartCity()); return g }
func (c CarItem) ReturnCityGroup() string { g, _ := GroupForCity(c.ReturnCity()); return g }
func (c CarItem) ReturnShopURL() string { return returnShopURLFromShop(c.ReturnShop) }

func (c CarItem) Key() string {
	return c.CarType
}

func (c CarItem) String() string {
	status := "受付中"
	if !c.Available {
		status = "受付終了"
	}
	return fmt.Sprintf("・%s → %s\n  車種: %s\n  期間: %s\n  条件: %s\n  電話: %s\n  状態: %s",
		c.StartShop, c.ReturnShop, c.CarType, c.Period, c.Condition, c.Tel, status)
}

func Fetch() ([]CarItem, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}
	return parse(doc), nil
}

func parse(doc *goquery.Document) []CarItem {
	shopText := func(s *goquery.Selection, cls string) string {
		p := s.Find(cls + " p:not(.label-sp)").First().Clone()
		p.Find("small").Remove()
		return clean(p.Text())
	}
	text := func(s *goquery.Selection, cls string) string {
		return clean(s.Find(cls + " p:not(.label-sp)").First().Text())
	}
	attr := func(s *goquery.Selection, key string) string { v, _ := s.Attr(key); return v }

	validation := func(i CarItem) bool {
		// Add your validation logic here
		parts := strings.Fields(i.StartShop)
		if len(parts) == 0 {
			return false
		}
		if _, ok := areaMap[parts[0]]; !ok {
			fmt.Printf("Unknown company in StartShop: %s\n", i.StartShop)
			return false
		}
		return true
	}

	var items []CarItem
	doc.Find("#service-items-shop-type-start .service-item").Each(func(_ int, s *goquery.Selection) {
		item := CarItem{
			StartShop:  shopText(s, ".service-item__shop-start"),
			StartArea:  attr(s, "data-start-area"),
			ReturnShop: shopText(s, ".service-item__shop-return"),
			ReturnArea: attr(s, "data-return-area"),
			CarType:    text(s, ".service-item__info__car-type"),
			Condition:  text(s, ".service-item__info__condition"),
			Period:     text(s, ".service-item__date"),
			Tel:        formatTel(clean(s.Find(".service-item__reserve-tel").First().Text())),
			Available:  !s.Find(".service-item__body").HasClass("show-entry-end"),
		}
		if validation(item) {
			items = append(items, item)
		}
	})

	return items
}

func clean(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// formatTel は電話番号を XXX-XXX-XXXX (3-3-4) 形式に正規化する。
// ハイフンの位置や全角数字は問わず、数字のみ抽出して再フォーマットする。
// 10桁以外はそのまま返す。
func formatTel(s string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r >= '０' && r <= '９' {
			return r - '０' + '0'
		}
		return -1
	}, s)
	if len(digits) != 10 {
		return s
	}
	return digits[:3] + "-" + digits[3:6] + "-" + digits[6:]
}
