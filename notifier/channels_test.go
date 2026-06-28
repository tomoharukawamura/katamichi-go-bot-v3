package notifier

import (
	"testing"
)

func newTestChannelConfig() *ChannelConfig {
	return &ChannelConfig{
		index: map[[2]string]string{
			{"2", "3"}:         "ch_2_3",
			{"2", "2"}:         "ch_2_2",
			{"北東北", "3"}:     "ch_kitaohoku_3",
			{"北東北", "南東北"}: "ch_kitaohoku_minamitohoku",
		},
	}
}

func TestChannelFor(t *testing.T) {
	cfg := newTestChannelConfig()

	cases := []struct {
		desc        string
		startGroup  string
		returnGroup string
		startArea   string
		returnArea  string
		wantID      string
		wantOK      bool
	}{
		// --- エリアコードのみで一致（groupが空） ---
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
		{
			desc:       "同一エリアコードで一致",
			startArea: "2", returnArea: "2",
			wantID: "ch_2_2", wantOK: true,
		},

		// --- グループ名が優先される ---
		{
			desc:        "startGroupがエリアコードより優先してマッチ",
			startGroup: "北東北", returnGroup: "",
			startArea: "2", returnArea: "3", // index["2","3"]も存在するが index["北東北","3"]が優先
			wantID: "ch_kitaohoku_3", wantOK: true,
		},
		{
			desc:        "両方グループ名でマッチ",
			startGroup: "北東北", returnGroup: "南東北",
			startArea: "2", returnArea: "3",
			wantID: "ch_kitaohoku_minamitohoku", wantOK: true,
		},
		{
			desc:        "グループ名逆順でもマッチ",
			startGroup: "南東北", returnGroup: "北東北",
			startArea: "3", returnArea: "2",
			wantID: "ch_kitaohoku_minamitohoku", wantOK: true,
		},

		// --- グループ未登録でエリアコードにフォールバック ---
		{
			desc:        "グループ名がsectionに未登録でエリアコードにフォールバック",
			startGroup: "甲信越", returnGroup: "",
			startArea: "2", returnArea: "3",
			wantID: "ch_2_3", wantOK: true,
		},

		// --- 一致なし ---
		{
			desc:       "どのエントリにも一致しない",
			startArea: "99", returnArea: "99",
			wantID: "", wantOK: false,
		},
		{
			desc:        "グループ・エリアともに未登録",
			startGroup: "未知", returnGroup: "未知",
			startArea: "99", returnArea: "99",
			wantID: "", wantOK: false,
		},

		// --- 空文字は無視 ---
		{
			desc:        "空のgroupは無視されてエリアコードで一致",
			startGroup: "", returnGroup: "",
			startArea: "2", returnArea: "2",
			wantID: "ch_2_2", wantOK: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			gotID, gotOK := cfg.ChannelFor(tc.startGroup, tc.returnGroup, tc.startArea, tc.returnArea)
			if gotOK != tc.wantOK || gotID != tc.wantID {
				t.Errorf("ChannelFor(%q,%q,%q,%q) = (%q,%v), want (%q,%v)",
					tc.startGroup, tc.returnGroup, tc.startArea, tc.returnArea,
					gotID, gotOK, tc.wantID, tc.wantOK)
			}
		})
	}
}
