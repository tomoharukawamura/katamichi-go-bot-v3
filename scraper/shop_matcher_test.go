package scraper

import "testing"

func TestStartCityFromShop(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// city_name ルックアップ（通常）
		{"トヨタレンタリース岩手 盛岡駅南口店", "岩手"},
		{"トヨタレンタリース岩手 二戸駅新幹線口店", "岩手"},
		{"トヨタレンタリース宮城 仙台空港店", "宮城"},
		{"トヨタレンタリース仙台 仙台新幹線口店", "宮城"},
		{"トヨタレンタリース山形 山形空港店", "山形"},
		{"トヨタレンタリース新福島 会津若松駅店", "福島"},
		{"トヨタレンタリース神奈川 横浜みなとみらい店", "神奈川"},
		{"トヨタS＆Dレンタシェア西東京 武蔵境店", "東京(多摩)"},
		{"トヨタS＆Dレンタシェア西東京 八王子駅前店", "東京(多摩)"},
		{"トヨタレンタリース静岡 三島新幹線口店", "静岡"},
		{"トヨタレンタリース名古屋 丸の内駅前店", "愛知"},
		{"トヨタレンタリース京都 三条京阪北店", "京都"},

		// shop_city_exception ルックアップ
		{"トヨタモビリティサービス 成田空港店", "成田"},

		// exception なし → city_name
		{"トヨタモビリティサービス 羽田空港(国内線)店", "東京"},
		{"トヨタモビリティサービス 品川高輪口店", "東京"},

		// エッジケース
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := startCityFromShop(tt.input)
			if got != tt.want {
				t.Errorf("startCityFromShop(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReturnCityFromShop(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// city_name が会社名から自明でないケース
		{"トヨタレンタリース仙台 返却可能店舗", "宮城"},
		{"トヨタレンタリース名古屋 返却可能店舗", "愛知"},
		{"トヨタレンタリース新大阪 返却可能店舗", "大阪"},
		{"トヨタレンタリース新福島 返却可能店舗", "福島"},
		{"トヨタモビリティサービス 返却可能店舗", "東京"},
		{"トヨタS＆Dレンタシェア西東京 返却可能店舗", "東京(多摩)"},

		// 通常ケース
		{"トヨタレンタリース青森 返却可能店舗", "青森"},
		{"トヨタレンタリース岩手 返却可能店舗", "岩手"},
		{"トヨタレンタリース宮城 返却可能店舗", "宮城"},
		{"トヨタレンタリース山形 返却可能店舗", "山形"},
		{"トヨタレンタリース福島 返却可能店舗", "福島"},
		{"トヨタレンタリース神奈川 返却可能店舗", "神奈川"},
		{"トヨタレンタリース静岡 返却可能店舗", "静岡"},
		{"トヨタレンタリース愛知 返却可能店舗", "愛知"},
		{"トヨタレンタリース京都 返却可能店舗", "京都"},
		{"トヨタレンタリース大阪 返却可能店舗", "大阪"},
		{"トヨタレンタリース福岡 返却可能店舗", "福岡"},
		{"トヨタレンタリース熊本 返却可能店舗", "熊本"},
		{"トヨタレンタリース鹿児島 返却可能店舗", "鹿児島"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := returnCityFromShop(tt.input)
			if got != tt.want {
				t.Errorf("returnCityFromShop(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIconFromShop(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// area icon ルックアップ（通常）
		{"トヨタレンタリース岩手 盛岡駅南口店", "🍇"},
		{"トヨタレンタリース山形 山形空港店", "🍒"},
		{"トヨタレンタリース宮城 仙台空港店", "🫛"},
		{"トヨタレンタリース仙台 仙台新幹線口店", "🫛"},

		// icon_exception ルックアップ
		{"トヨタモビリティサービス 成田空港店", "✈️"},
		{"トヨタモビリティサービス 東久留米駅前店", "🐎"},
		{"トヨタモビリティサービス 調布店", "🐎"},

		// exception なし → area icon
		{"トヨタモビリティサービス 品川高輪口店", "🗼"},
		{"トヨタモビリティサービス 羽田空港(国内線)店", "🗼"},

		// 未登録 → 空文字
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := iconFromShop(tt.input)
			if got != tt.want {
				t.Errorf("iconFromShop(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCarItemCityMethods(t *testing.T) {
	tests := []struct {
		startShop   string
		returnShop  string
		wantStart   string
		wantReturn  string
		wantStartIcon  string
		wantReturnIcon string
	}{
		{"トヨタレンタリース神奈川 横浜みなとみらい店", "トヨタレンタリース静岡 返却可能店舗", "神奈川", "静岡", "🏖️", "🍵"},
		{"トヨタモビリティサービス 成田空港店", "トヨタレンタリース新大阪 返却可能店舗", "成田", "大阪", "✈️", "🏯"},
		{"トヨタモビリティサービス 品川高輪口店", "トヨタモビリティサービス 返却可能店舗", "東京", "東京", "🗼", "🗼"},
		{"トヨタS＆Dレンタシェア西東京 武蔵境店", "トヨタレンタリース静岡 返却可能店舗", "東京(多摩)", "静岡", "🐎", "🍵"},
		{"トヨタレンタリース仙台 仙台新幹線口店", "トヨタレンタリース新福島 返却可能店舗", "宮城", "福島", "🫛", "🍑"},
	}

	for _, tt := range tests {
		item := CarItem{StartShop: tt.startShop, ReturnShop: tt.returnShop}
		if got := item.StartCity(); got != tt.wantStart {
			t.Errorf("StartCity(%q) = %q, want %q", tt.startShop, got, tt.wantStart)
		}
		if got := item.ReturnCity(); got != tt.wantReturn {
			t.Errorf("ReturnCity(%q) = %q, want %q", tt.returnShop, got, tt.wantReturn)
		}
		if got := item.StartCityIcon(); got != tt.wantStartIcon {
			t.Errorf("StartCityIcon(%q) = %q, want %q", tt.startShop, got, tt.wantStartIcon)
		}
		if got := item.ReturnCityIcon(); got != tt.wantReturnIcon {
			t.Errorf("ReturnCityIcon(%q) = %q, want %q", tt.returnShop, got, tt.wantReturnIcon)
		}
	}
}
