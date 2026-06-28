package scraper

import (
	"testing"
)

func TestIdentify(t *testing.T) {
	m, err := NewCarMatcher()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input     string
		wantName  string
		wantMeta  CarMeta
		wantFound bool
	}{
		// prefix なし
		{"アクア", "アクア", CarMeta{Type: "aqua"}, true},
		// prefix あり: prefix+canonical（スペースなし）
		{"GR86", "86", CarMeta{Type: "sports"}, true},
		{"ダイハツタント", "タント", CarMeta{Type: "light"}, true},
		{"スズキスペーシア", "スペーシア", CarMeta{Type: "light"}, true},
		// prefix あり: prefix+" "+canonical（スペースあり）
		{"GR 86", "86", CarMeta{Type: "sports"}, true},
		{"ダイハツ タント", "タント", CarMeta{Type: "light"}, true},
		{"カローラ ツーリング", "ツーリング", CarMeta{Type: "wagon_normal"}, true},
		{"カローラ・ツーリング", "ツーリング", CarMeta{Type: "wagon_normal"}, true},
		// prefix+canonical+プレート
		{"ダイハツタント 品川300あ1234", "タント", CarMeta{Type: "light"}, true},
		// 境界の記号トリム
		{"アクア・", "アクア", CarMeta{Type: "aqua"}, true},
		{"・アクア", "アクア", CarMeta{Type: "aqua"}, true},
		{"　アクア　", "アクア", CarMeta{Type: "aqua"}, true},
		{"アクア - 品川300あ1234", "アクア", CarMeta{Type: "aqua"}, true},
		{"プリウス", "プリウス", CarMeta{Type: "prius_normal"}, true},
		{"プリウスPHEV", "プリウスPHV", CarMeta{Type: "prius_plugin"}, true},
		{"プリウスアルファ", "プリウスα", CarMeta{Type: "prius_van"}, true},
		{"アクア 品川 500 あ 1234", "アクア", CarMeta{Type: "aqua"}, true},
		{"アクア さいたま 500 あ 1234", "アクア", CarMeta{Type: "aqua"}, true},
		{"プリウスＰＨＶ", "プリウスPHV", CarMeta{Type: "prius_plugin"}, true},
		{"ｱｸｱ", "アクア", CarMeta{Type: "aqua"}, true},
		// HV / HEV
		{"アルファードHV 品川 300 あ 1234", "アルファード", CarMeta{Type: "minivan_luxury", HasHybridLabel: true}, true},
		{"ヤリスHV", "ヤリス", CarMeta{Type: "compact_normal", HasHybridLabel: true}, true},
		{"ノアHEV", "ノア", CarMeta{Type: "minivan_normal", HasHybridLabel: true}, true},
		{"ヤリスhv", "ヤリス", CarMeta{Type: "compact_normal", HasHybridLabel: true}, true},
		{"アクア", "アクア", CarMeta{Type: "aqua"}, true},
		{"存在しない車種", "", CarMeta{}, false},
		{"", "", CarMeta{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, meta, ok := m.Identify(tt.input)
			if ok != tt.wantFound {
				t.Errorf("Identify(%q) found=%v, want %v", tt.input, ok, tt.wantFound)
			}
			if got != tt.wantName {
				t.Errorf("Identify(%q) = %q, want %q", tt.input, got, tt.wantName)
			}
			if meta != tt.wantMeta {
				t.Errorf("Identify(%q) meta=%+v, want %+v", tt.input, meta, tt.wantMeta)
			}
		})
	}
}

func TestIdentifyPrefixAlias(t *testing.T) {
	const testJSON = `{"cars":[{
		"canonical": "PHV",
		"aliases": ["PHEV", "プラグインハイブリッド"],
		"prefix": ["テスト"],
		"meta": {"type": "prius_plugin"}
	}]}`
	m, err := newCarMatcherFromBytes([]byte(testJSON))
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct{ input, want string }{
		{"PHV", "PHV"},
		{"PHEV", "PHV"},
		{"テストPHV", "PHV"},
		{"テストPHEV", "PHV"},
		{"テスト PHV", "PHV"},
		{"テスト PHEV", "PHV"},
		{"テストプラグインハイブリッド", "PHV"},
		{"テスト プラグインハイブリッド", "PHV"},
	}
	for _, c := range cases {
		got, _, ok := m.Identify(c.input)
		if !ok || got != c.want {
			t.Errorf("Identify(%q) = %q, ok=%v; want %q, true", c.input, got, ok, c.want)
		}
	}
}

func TestNormalizeKana(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ｱｸｱ", "アクア"},
		{"ｶﾞ", "ガ"},
		{"ﾊﾟ", "パ"},
		{"ＡＢＣ", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeKana(tt.input)
			if got != tt.want {
				t.Errorf("normalizeKana(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
