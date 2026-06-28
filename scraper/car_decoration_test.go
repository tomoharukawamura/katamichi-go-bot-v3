package scraper

import "testing"

func TestDecoration_icon(t *testing.T) {
	tests := []struct {
		carType  string
		wantIcon string
	}{
		// アイコンがJSONで指定したものと一致するか
		{"アクア", "🌏"},          // type=aqua
		{"プリウス", "🚀"},         // type=prius_normal
		{"プリウスPHV", "🔌"},      // type=prius_plugin
		{"プリウスα", "🚀"},        // type=prius_van
		{"ヤリスクロス", "🚙"},       // type=suv_normal
		{"ハリアー", "🚙"},          // type=suv_rare
		{"カローラ", "🚔"},          // type=sedan_normal
		{"クラウン", "👑"},          // type=crown
		{"ノア", "🚌"},             // type=minivan_normal
		{"アルファード", "🧑‍💼"},    // type=minivan_luxury
		{"ランドクルーザープラド", "⛰️"}, // type=randcruiser
		{"ハイエース", "🚐"},         // type=van_normal
		{"ダイハツタント", "🚗"},      // type=light
		{"ヤリス", "🚗"},            // type=compact_normal → decoration未定義なのでdefault
		{"存在しない車種", "🚗"},      // 未マッチ → default
	}

	for _, tt := range tests {
		t.Run(tt.carType, func(t *testing.T) {
			item := CarItem{CarType: tt.carType}
			got := item.Decoration().Icon
			if got != tt.wantIcon {
				t.Errorf("Decoration(%q).Icon = %q, want %q", tt.carType, got, tt.wantIcon)
			}
		})
	}
}

func TestDecoration_hvLabelSwitchesToEcoIcon(t *testing.T) {
	tests := []struct {
		carType  string
		wantIcon string
		desc     string
	}{
		// ecoが設定されている車種: HV表記でecoアイコンに切り替わる
		{"ヤリスHV", "🌏", "compact_normal(default)のecoアイコン"},
		{"ヤリスhv", "🌏", "小文字hvでも切り替わる"},
		{"カローラHEV", "🌏", "sedan_normalのecoアイコン"},
		{"ヤリスクロスHV", "🌏", "suv_normalのecoアイコン"},
		// ecoが設定されていない車種: HV表記でもアイコンは変わらない
		{"ノアHEV", "🚌", "minivan_normalにecoなし → そのまま"},
		{"アルファードHV", "🧑‍💼", "minivan_luxuryにecoなし → そのまま"},
		{"ハリアーHV", "🚙", "suv_rareにecoなし → そのまま"},
	}

	for _, tt := range tests {
		t.Run(tt.carType, func(t *testing.T) {
			item := CarItem{CarType: tt.carType}
			got := item.Decoration().Icon
			if got != tt.wantIcon {
				t.Errorf("[%s] Decoration(%q).Icon = %q, want %q", tt.desc, tt.carType, got, tt.wantIcon)
			}
		})
	}
}
