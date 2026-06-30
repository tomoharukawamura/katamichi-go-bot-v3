package notifier

import (
	"testing"

	"github.com/tomok/katamichi-go-bot-v3/scraper"
)

func TestItemAttachment_Color(t *testing.T) {
	cases := []struct {
		carType       string
		expectedColor string
		desc          string
	}{
		// default（decorationマップ未登録の型 or 未知の車種）
		{"ヤリスHV テスト用", "#36a64f", "未知の車種はdefault"},

		// crown: #ffd700
		{"クラウン 品川300あ1234", "#ffd700", "クラウン（crown）"},

		// randcruiser: #808000
		{"ランドクルーザープラド 品川300あ1234", "#808000", "ランクルプラド（randcruiser）"},
		{"ランクルプラド 品川300あ1234", "#808000", "ランクル（randcruiserエイリアス）"},

		// minivan_luxury: #b8860b
		{"アルファード 品川300あ1234", "#b8860b", "アルファード（minivan_luxury）"},
		{"ヴェルファイア 品川300あ1234", "#b8860b", "ヴェルファイア（minivan_luxury）"},

		// minivan_normal: #36a64f
		{"ノア 品川300あ1234", "#36a64f", "ノア（minivan_normal）"},

		// minivan_rare: #f58220
		{"ヴォクシー 品川300あ1234", "#f58220", "ヴォクシー（minivan_rare）"},

		// sports: #ff0000
		{"GR86 品川300あ1234", "#ff0000", "GR86（sports）"},
		{"アクアGR 品川300あ1234", "#ff0000", "アクアGR（sports）"},

		// suv_rare: #00bfff
		{"ハリアー 品川300あ1234", "#00bfff", "ハリアー（suv_rare）"},
		{"RAV4 品川300あ1234", "#00bfff", "RAV4（suv_rare）"},

		// sedan_rare: #a6a6a6
		{"カムリ 品川300あ1234", "#a6a6a6", "カムリ（sedan_rare）"},
		{"マークX 品川300あ1234", "#a6a6a6", "マークX（sedan_rare）"},

		// prius_plugin: #a6a6a6
		{"プリウスPHV 品川300あ1234", "#a6a6a6", "プリウスPHV（prius_plugin）"},
		{"プリウスPHEV 品川300あ1234", "#a6a6a6", "プリウスPHEV（prius_pluginエイリアス）"},

		// prius_normal: #36a64f
		{"プリウス 品川300あ1234", "#36a64f", "プリウス（prius_normal）"},

		// prius_van: #f58220
		{"プリウスα 品川300あ1234", "#f58220", "プリウスα（prius_van）"},
		{"プリウスアルファ 品川300あ1234", "#f58220", "プリウスアルファ（prius_vanエイリアス）"},

		// van_normal: #36a64f
		{"ハイエース 品川300あ1234", "#36a64f", "ハイエース（van_normal）"},
		{"ハイエースグランドキャビン 品川300あ1234", "#f58220", "ハイエースGC（minivan_rare）"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			item := scraper.CarItemForMessage{CarItem: scraper.CarItem{CarType: tc.carType}}
			att := itemAttachment(item, "新着")
			if att.Color != tc.expectedColor {
				t.Errorf("CarType=%q: want color %q, got %q", tc.carType, tc.expectedColor, att.Color)
			}
		})
	}
}
