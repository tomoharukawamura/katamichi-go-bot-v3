package notifier

import (
	"testing"
)

func newTestChannelConfig() *ChannelConfig {
	return &ChannelConfig{
		index: map[[2]string]string{
			{"2", "3"}:    "ch_2_3",
			{"2", "2"}:    "ch_2_2",
			{"岩手", "3"}: "ch_iwate_3", // sectionに都市名が混在するケース
			{"盛岡", "3"}: "ch_morioka_3", // 都市名×エリアコード混在
		},
	}
}

func TestChannelFor(t *testing.T) {
	cfg := newTestChannelConfig()

	cases := []struct {
		desc       string
		startArea  string
		returnArea string
		startCity  string
		returnCity string
		wantID     string
		wantOK     bool
	}{
		// --- エリアコードのみで一致 ---
		{
			desc:       "エリアコード正順で一致",
			startArea: "2", returnArea: "3",
			wantID: "ch_2_3", wantOK: true,
		},
		{
			desc:       "エリアコード逆順で一致",
			startArea: "3", returnArea: "2",
			wantID: "ch_2_3", wantOK: true,
		},

		// --- sectionに都市名が入っているケース ---
		{
			desc:       "sectionに都市名が入っているケース 正順",
			startArea: "岩手", returnArea: "3",
			wantID: "ch_iwate_3", wantOK: true,
		},
		{
			desc:       "sectionに都市名が入っているケース 逆順",
			startArea: "3", returnArea: "岩手",
			wantID: "ch_iwate_3", wantOK: true,
		},

		// --- startCityによる都市名優先マッチ ---
		{
			desc:       "startCityが都市名でsectionに都市名エントリが存在する場合優先される",
			startArea: "2", returnArea: "3", // index["2","3"] も存在する
			startCity: "岩手", returnCity: "",  // index["岩手","3"] を優先
			wantID: "ch_iwate_3", wantOK: true,
		},
		{
			desc:       "都市名×エリアコード混在のsectionにヒット",
			startArea: "99", returnArea: "3",
			startCity: "盛岡", returnCity: "",
			wantID: "ch_morioka_3", wantOK: true,
		},
		{
			desc:       "都市名がsectionに未登録でもエリアコードにフォールバック",
			startArea: "2", returnArea: "3",
			startCity: "不明市", returnCity: "不明市",
			wantID: "ch_2_3", wantOK: true,
		},

		// --- 一致なし ---
		{
			desc:       "どのエントリにも一致しない",
			startArea: "99", returnArea: "99",
			startCity: "不明", returnCity: "不明",
			wantID: "", wantOK: false,
		},
		{
			desc:       "空文字は無視されてエリアコードで一致",
			startArea: "2", returnArea: "2",
			startCity: "", returnCity: "",
			wantID: "ch_2_2", wantOK: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			gotID, gotOK := cfg.ChannelFor(tc.startArea, tc.returnArea, tc.startCity, tc.returnCity)
			if gotOK != tc.wantOK || gotID != tc.wantID {
				t.Errorf("ChannelFor(%q,%q,%q,%q) = (%q,%v), want (%q,%v)",
					tc.startArea, tc.returnArea, tc.startCity, tc.returnCity,
					gotID, gotOK, tc.wantID, tc.wantOK)
			}
		})
	}
}
