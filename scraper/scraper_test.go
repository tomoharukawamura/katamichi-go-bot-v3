package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func mustParse(html string) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		panic(err)
	}
	return doc
}

const sampleHTML = `
<div id="service-items-shop-type-start">
  <div class="service-item" data-start-area="2" data-return-area="3">
    <div class="service-item__body">
      <div class="service-item__shop-start">
        <p>トヨタレンタリース岩手 盛岡駅南口店<small>（補足）</small></p>
      </div>
      <div class="service-item__shop-return">
        <p>トヨタレンタリース青森 青森空港店</p>
      </div>
      <div class="service-item__info__car-type">
        <p class="label-sp">車種</p>
        <p>ヤリスHV 青森501わ3002</p>
      </div>
      <div class="service-item__info__condition">
        <p class="label-sp">条件</p>
        <p>禁煙 AT</p>
      </div>
      <div class="service-item__date">
        <p class="label-sp">期間</p>
        <p>2026年5月22日 ～ 5月29日</p>
      </div>
      <div class="service-item__reserve-tel">0196063076</div>
    </div>
  </div>
</div>`

func TestParse_basic(t *testing.T) {
	items := parse(mustParse(sampleHTML))

	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	it := items[0]

	checks := []struct{ got, want string }{
		{it.StartShop, "トヨタレンタリース岩手 盛岡駅南口店"},
		{it.StartArea, "2"},
		{it.ReturnShop, "トヨタレンタリース青森 青森空港店"},
		{it.ReturnArea, "3"},
		{it.CarType, "ヤリスHV 青森501わ3002"},
		{it.Condition, "禁煙 AT"},
		{it.Period, "2026年5月22日 ～ 5月29日"},
		{it.Tel, "019-606-3076"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
	if !it.Available {
		t.Errorf("Available = false, want true")
	}
}

func TestParse_unavailable(t *testing.T) {
	html := `
	<div id="service-items-shop-type-start">
	  <div class="service-item" data-start-area="1" data-return-area="2">
	    <div class="service-item__body show-entry-end">
	      <div class="service-item__shop-start"><p>トヨタレンタリース岩手 盛岡店</p></div>
	      <div class="service-item__shop-return"><p>トヨタレンタリース青森 青森店</p></div>
	      <div class="service-item__info__car-type"><p>ヤリス</p></div>
	      <div class="service-item__info__condition"><p></p></div>
	      <div class="service-item__date"><p>5月1日</p></div>
	      <div class="service-item__reserve-tel"></div>
	    </div>
	  </div>
	</div>`
	items := parse(mustParse(html))
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if items[0].Available {
		t.Errorf("Available = true, want false for show-entry-end")
	}
}

func TestParse_skipsEmptyShop(t *testing.T) {
	html := `
	<div id="service-items-shop-type-start">
	  <div class="service-item" data-start-area="1" data-return-area="1">
	    <div class="service-item__shop-start"><p></p></div>
	  </div>
	</div>`
	items := parse(mustParse(html))
	if len(items) != 0 {
		t.Errorf("want 0 items for empty shop, got %d", len(items))
	}
}

func TestParse_smallTagStripped(t *testing.T) {
	html := `
	<div id="service-items-shop-type-start">
	  <div class="service-item" data-start-area="1" data-return-area="2">
	    <div class="service-item__body">
	      <div class="service-item__shop-start"><p>トヨタレンタリース岩手 本店<small>（直営）</small></p></div>
	      <div class="service-item__shop-return"><p>トヨタレンタリース青森 空港店</p></div>
	      <div class="service-item__info__car-type"><p>アクア</p></div>
	      <div class="service-item__info__condition"><p></p></div>
	      <div class="service-item__date"><p>期間未定</p></div>
	      <div class="service-item__reserve-tel"></div>
	    </div>
	  </div>
	</div>`
	items := parse(mustParse(html))
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if items[0].StartShop != "トヨタレンタリース岩手 本店" {
		t.Errorf("small tag not stripped: got %q", items[0].StartShop)
	}
}

func TestParse_labelSpIgnored(t *testing.T) {
	html := `
	<div id="service-items-shop-type-start">
	  <div class="service-item" data-start-area="1" data-return-area="2">
	    <div class="service-item__body">
	      <div class="service-item__shop-start"><p>トヨタレンタリース岩手 駅前店</p></div>
	      <div class="service-item__shop-return"><p>トヨタレンタリース青森 港店</p></div>
	      <div class="service-item__info__car-type">
	        <p class="label-sp">車種ラベル</p>
	        <p>プリウス</p>
	      </div>
	      <div class="service-item__info__condition"><p></p></div>
	      <div class="service-item__date"><p>5月1日</p></div>
	      <div class="service-item__reserve-tel"></div>
	    </div>
	  </div>
	</div>`
	items := parse(mustParse(html))
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if items[0].CarType != "プリウス" {
		t.Errorf("label-sp not ignored: got %q", items[0].CarType)
	}
}

func TestFormatTel(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0196063076", "019-606-3076"},
		{"019-606-3076", "019-606-3076"},
		{"０１９６０６３０７６", "019-606-3076"},
		{"03-1234-5678", "031-234-5678"},
		{"短い", "短い"},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatTel(tt.input)
		if got != tt.want {
			t.Errorf("formatTel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestClean(t *testing.T) {
	tests := []struct{ input, want string }{
		{"  hello   world  ", "hello world"},
		{"\t改行\nテスト\t", "改行 テスト"},
		{"", ""},
	}
	for _, tt := range tests {
		got := clean(tt.input)
		if got != tt.want {
			t.Errorf("clean(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCarItemKey(t *testing.T) {
	it := CarItem{StartShop: "盛岡店", CarType: "ヤリス", Period: "5月22日"}
	want := "ヤリス"
	if got := it.Key(); got != want {
		t.Errorf("Key() = %q, want %q", got, want)
	}
}
