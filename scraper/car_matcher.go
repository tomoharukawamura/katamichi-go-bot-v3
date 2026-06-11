package scraper

import (
	_ "embed"
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

//go:embed data/car_master.json
var defaultMasterJSON []byte

// フルプレートパターン: 地名(漢字) + 分類番号 + ひらがな + 一連番号
var fullPlateRe = regexp.MustCompile(`\s*[一-龥]{1,4}\s*\d{1,3}\s*[あ-ん]\s*\d{1,4}$`)

// 記号・空白など非文字・非数字を全て除去
var symbolRe = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// 半角カタカナ → 全角カタカナ
var halfKanaToFull = map[rune]rune{
	'ｦ': 'ヲ', 'ｧ': 'ァ', 'ｨ': 'ィ', 'ｩ': 'ゥ', 'ｪ': 'ェ', 'ｫ': 'ォ',
	'ｬ': 'ャ', 'ｭ': 'ュ', 'ｮ': 'ョ', 'ｯ': 'ッ', 'ｰ': 'ー',
	'ｱ': 'ア', 'ｲ': 'イ', 'ｳ': 'ウ', 'ｴ': 'エ', 'ｵ': 'オ',
	'ｶ': 'カ', 'ｷ': 'キ', 'ｸ': 'ク', 'ｹ': 'ケ', 'ｺ': 'コ',
	'ｻ': 'サ', 'ｼ': 'シ', 'ｽ': 'ス', 'ｾ': 'セ', 'ｿ': 'ソ',
	'ﾀ': 'タ', 'ﾁ': 'チ', 'ﾂ': 'ツ', 'ﾃ': 'テ', 'ﾄ': 'ト',
	'ﾅ': 'ナ', 'ﾆ': 'ニ', 'ﾇ': 'ヌ', 'ﾈ': 'ネ', 'ﾉ': 'ノ',
	'ﾊ': 'ハ', 'ﾋ': 'ヒ', 'ﾌ': 'フ', 'ﾍ': 'ヘ', 'ﾎ': 'ホ',
	'ﾏ': 'マ', 'ﾐ': 'ミ', 'ﾑ': 'ム', 'ﾒ': 'メ', 'ﾓ': 'モ',
	'ﾔ': 'ヤ', 'ﾕ': 'ユ', 'ﾖ': 'ヨ',
	'ﾗ': 'ラ', 'ﾘ': 'リ', 'ﾙ': 'ル', 'ﾚ': 'レ', 'ﾛ': 'ロ',
	'ﾜ': 'ワ', 'ﾝ': 'ン',
}

// 半角濁点(ﾞ)結合マップ
var voicedCombine = map[rune]rune{
	'カ': 'ガ', 'キ': 'ギ', 'ク': 'グ', 'ケ': 'ゲ', 'コ': 'ゴ',
	'サ': 'ザ', 'シ': 'ジ', 'ス': 'ズ', 'セ': 'ゼ', 'ソ': 'ゾ',
	'タ': 'ダ', 'チ': 'ヂ', 'ツ': 'ヅ', 'テ': 'デ', 'ト': 'ド',
	'ハ': 'バ', 'ヒ': 'ビ', 'フ': 'ブ', 'ヘ': 'ベ', 'ホ': 'ボ',
	'ウ': 'ヴ',
}

// 半角半濁点(ﾟ)結合マップ
var semiVoicedCombine = map[rune]rune{
	'ハ': 'パ', 'ヒ': 'ピ', 'フ': 'プ', 'ヘ': 'ペ', 'ホ': 'ポ',
}

func normalizeKana(s string) string {
	result := make([]rune, 0, len([]rune(s)))
	for _, r := range s {
		switch {
		case r >= 0xFF01 && r <= 0xFF5E:
			result = append(result, r-0xFEE0)
		case r == 0xFF9E && len(result) > 0:
			if v, ok := voicedCombine[result[len(result)-1]]; ok {
				result[len(result)-1] = v
			} else {
				result = append(result, r)
			}
		case r == 0xFF9F && len(result) > 0:
			if v, ok := semiVoicedCombine[result[len(result)-1]]; ok {
				result[len(result)-1] = v
			} else {
				result = append(result, r)
			}
		default:
			if fw, ok := halfKanaToFull[r]; ok {
				result = append(result, fw)
			} else {
				result = append(result, r)
			}
		}
	}
	return string(result)
}

// CarMeta は master.json の meta フィールドに対応する
type CarMeta struct {
	Type string `json:"type"`
}

type matchResult struct {
	canonical string
	meta      CarMeta
}

type carEntry struct {
	Canonical string   `json:"canonical"`
	Aliases   []string `json:"aliases"`
	Prefix    []string `json:"prefix"`
	Meta      CarMeta  `json:"meta"`
}

type masterData struct {
	Cars []carEntry `json:"cars"`
}

// CarMatcher は車種名の正規化・特定を行う
type CarMatcher struct {
	aliasMap map[string]matchResult
}

// NewCarMatcher はバイナリに埋め込まれた master.json から CarMatcher を生成する
func NewCarMatcher() (*CarMatcher, error) {
	return newCarMatcherFromBytes(defaultMasterJSON)
}

// LoadCarMatcher は指定パスの master.json から CarMatcher を生成する（主にテスト用）
func LoadCarMatcher(path string) (*CarMatcher, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return newCarMatcherFromBytes(data)
}

func normalizeForLookup(s string) string {
	return symbolRe.ReplaceAllString(normalizeKana(s), "")
}

func newCarMatcherFromBytes(data []byte) (*CarMatcher, error) {
	var md masterData
	if err := json.Unmarshal(data, &md); err != nil {
		return nil, err
	}
	aliasMap := make(map[string]matchResult)
	for _, car := range md.Cars {
		r := matchResult{canonical: car.Canonical, meta: car.Meta}
		names := append([]string{car.Canonical}, car.Aliases...)
		for _, name := range names {
			aliasMap[normalizeForLookup(name)] = r
		}
		for _, pfx := range car.Prefix {
			if pfx == "" {
				continue
			}
			for _, name := range names {
				aliasMap[normalizeForLookup(pfx+name)] = r
			}
		}
	}
	return &CarMatcher{aliasMap: aliasMap}, nil
}

// Identify は車種文字列（プレート番号含む可能性あり）から正規車種名とメタデータを返す
func (m *CarMatcher) Identify(carType string) (canonical string, meta CarMeta, ok bool) {
	var r matchResult

	stripped := normalizeForLookup(strings.TrimSpace(fullPlateRe.ReplaceAllString(carType, "")))
	if r, ok = m.aliasMap[stripped]; ok {
		return r.canonical, r.meta, true
	}

	runes := []rune(normalizeKana(strings.TrimSpace(carType)))
	for i := len(runes) - 1; i > 0; i-- {
		if r, ok = m.aliasMap[normalizeForLookup(string(runes[:i]))]; ok {
			return r.canonical, r.meta, true
		}
	}

	return "", CarMeta{}, false
}
