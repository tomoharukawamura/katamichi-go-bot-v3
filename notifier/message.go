package notifier

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/tomok/katamichi-go-bot-v3/scraper"
	"github.com/tomok/katamichi-go-bot-v3/storage"
)

func (s *Slack) Notify(d scraper.Diff, state *storage.State, ch *ChannelConfig) error {
	type tsResult struct {
		key       string
		ts        string
		channelID string
		err       error
	}

	var wg sync.WaitGroup

	// storedChannel は保存済みチャンネルIDを返す。未保存の場合は現在のエリア・都市名から解決する。
	storedChannel := func(item scraper.CarItemForMessage) (string, bool) {
		if cid := state.MessageChannel[item.Key()]; cid != "" {
			return cid, true
		}
		// チャンネルが見つからない場合は通知しない
		return ch.ChannelFor(item.StartCityGroup(), item.ReturnCityGroup(), item.StartArea, item.ReturnArea)
	}

	addedResults := make([]tsResult, len(d.Added))
	for i, item := range d.Added {
		wg.Add(1)
		go func(i int, item scraper.CarItemForMessage) {
			defer wg.Done()
			cid, ok := ch.ChannelFor(item.StartCityGroup(), item.ReturnCityGroup(), item.StartArea, item.ReturnArea)
			if !ok {
				return
			}
			ts, err := s.Send(cid, buildHeaderText(item), []Attachment{itemAttachment(item)})
			addedResults[i] = tsResult{item.Key(), ts, cid, err}
		}(i, item)
	}

	// channelChangedUpdated: エリア修正によりチャンネルが変わったUpdatedアイテム
	// → 旧チャンネルをSoldOut扱いにして新チャンネルに新着投稿する
	channelChangedUpdatedResults := make([]tsResult, len(d.Updated))
	updatedResults := make([]tsResult, len(d.Updated))
	for i, item := range d.Updated {
		storedCID := state.MessageChannel[item.Key()]
		currentCID, ok := ch.ChannelFor(item.StartCityGroup(), item.ReturnCityGroup(), item.StartArea, item.ReturnArea)
		if !ok {
			continue
		}
		storedTS := state.MessageTS[item.Key()]
		if storedTS == "" {
			continue
		}

		if storedCID != "" && storedCID != currentCID {
			// チャンネル変化: 旧チャンネルにSoldOut（ベストエフォート）+ 新チャンネルに新着
			wg.Add(2)
			go func(i int, storedCID, storedTS string) {
				defer wg.Done()
				_ = s.Delete(storedCID, storedTS)
			}(i, storedCID, storedTS)
			go func(i int, item scraper.CarItemForMessage, currentCID string) {
				defer wg.Done()
				ts, err := s.Send(currentCID, buildHeaderText(item, "新着"), []Attachment{itemAttachment(item, "新着")})
				channelChangedUpdatedResults[i] = tsResult{item.Key(), ts, currentCID, err}
			}(i, item, currentCID)
		} else {
			// 同一チャンネル: スレッド返信
			cid := currentCID
			if storedCID != "" {
				cid = storedCID
			}
			wg.Add(1)
			go func(i int, item scraper.CarItemForMessage, cid, storedTS string) {
				defer wg.Done()
				newTS, err := s.replyOrSend(cid, storedTS, buildHeaderText(item), []Attachment{itemAttachment(item)})
				updatedResults[i] = tsResult{item.Key(), newTS, cid, err}
			}(i, item, cid, storedTS)
		}
	}

	reopenedReplyResults := make([]tsResult, len(d.Reopened))
	reopenedReactionErrs := make([]error, len(d.Reopened))
	channelChangedReopenedResults := make([]tsResult, len(d.Reopened))
	reopenedNewResults := make([]tsResult, len(d.Reopened))
	for i, item := range d.Reopened {
		storedCID := state.MessageChannel[item.Key()]
		currentCID, ok := ch.ChannelFor(item.StartCityGroup(), item.ReturnCityGroup(), item.StartArea, item.ReturnArea	)
		if !ok {
			continue
		}
		storedTS := state.MessageTS[item.Key()]
		if storedTS == "" {
			// storedTSが未登録の場合は新着として投稿
			wg.Add(1)
			go func(i int, item scraper.CarItemForMessage, currentCID string) {
				defer wg.Done()
				ts, err := s.Send(currentCID, buildHeaderText(item, "新着"), []Attachment{itemAttachment(item, "新着")})
				reopenedNewResults[i] = tsResult{item.Key(), ts, currentCID, err}
			}(i, item, currentCID)
			continue
		}

		if storedCID != "" && storedCID != currentCID {
			// チャンネル変化: 旧チャンネルにSoldOut（ベストエフォート）+ 新チャンネルに新着
			wg.Add(2)
			go func(i int, storedCID, storedTS string) {
				defer wg.Done()
				_ = s.Delete(storedCID, storedTS)
			}(i, storedCID, storedTS)
			go func(i int, item scraper.CarItemForMessage, currentCID string) {
				defer wg.Done()
				ts, err := s.Send(currentCID, buildHeaderText(item, "新着"), []Attachment{itemAttachment(item, "新着")})
				channelChangedReopenedResults[i] = tsResult{item.Key(), ts, currentCID, err}
			}(i, item, currentCID)
		} else {
			// 同一チャンネル: SoldOutリアクション除去 + スレッド返信
			cid := currentCID
			if storedCID != "" {
				cid = storedCID
			}
			wg.Add(2)
			go func(i int, cid, storedTS string) {
				defer wg.Done()
				reopenedReactionErrs[i] = s.RemoveReaction(cid, storedTS, "sold_out")
			}(i, cid, storedTS)
			go func(i int, item scraper.CarItemForMessage, cid, storedTS string) {
				defer wg.Done()
				newTS, err := s.replyOrSend(cid, storedTS, buildHeaderText(item), []Attachment{itemAttachment(item)})
				reopenedReplyResults[i] = tsResult{item.Key(), newTS, cid, err}
			}(i, item, cid, storedTS)
		}
	}

	soldOutErrs := make([]error, len(d.SoldOut))
	for i, item := range d.SoldOut {
		wg.Add(1)
		go func(i int, item scraper.CarItemForMessage) {
			defer wg.Done()
			cid, ok := storedChannel(item)
			if !ok {
				return
			}
			if ts := state.MessageTS[item.Key()]; ts != "" {
				soldOutErrs[i] = s.AddReaction(cid, ts, "sold_out")
			}
		}(i, item)
	}

	wg.Wait()

	for _, r := range addedResults {
		if r.err == nil && r.ts != "" {
			state.MessageTS[r.key] = r.ts
			state.MessageChannel[r.key] = r.channelID
		}
	}
	for _, r := range channelChangedUpdatedResults {
		if r.err == nil && r.ts != "" {
			state.MessageTS[r.key] = r.ts
			state.MessageChannel[r.key] = r.channelID
		}
	}
	for _, r := range updatedResults {
		if r.err == nil && r.ts != "" {
			state.MessageTS[r.key] = r.ts
		}
	}
	for _, r := range channelChangedReopenedResults {
		if r.err == nil && r.ts != "" {
			state.MessageTS[r.key] = r.ts
			state.MessageChannel[r.key] = r.channelID
		}
	}
	for _, r := range reopenedReplyResults {
		if r.err == nil && r.ts != "" {
			state.MessageTS[r.key] = r.ts
		}
	}
	for _, r := range reopenedNewResults {
		if r.err == nil && r.ts != "" {
			state.MessageTS[r.key] = r.ts
			state.MessageChannel[r.key] = r.channelID
		}
	}

	// channelChangedUpdatedReactionErrs / channelChangedReopenedReactionErrs は
	// 旧チャンネルへの後始末リアクションなのでベストエフォート（エラー無視）

	var allErrs []error
	for _, r := range addedResults {
		allErrs = append(allErrs, r.err)
	}
	for _, r := range channelChangedUpdatedResults {
		allErrs = append(allErrs, r.err)
	}
	for _, r := range updatedResults {
		allErrs = append(allErrs, r.err)
	}
	for _, r := range channelChangedReopenedResults {
		allErrs = append(allErrs, r.err)
	}
	allErrs = append(allErrs, reopenedReactionErrs...)
	for _, r := range reopenedReplyResults {
		allErrs = append(allErrs, r.err)
	}
	for _, r := range reopenedNewResults {
		allErrs = append(allErrs, r.err)
	}
	allErrs = append(allErrs, soldOutErrs...)
	return errors.Join(allErrs...)
}

func buildHeaderText(item scraper.CarItemForMessage, override ...string) string {
	icon := item.Decoration().Icon
	status := ""
	if s := item.Status(); s != nil {
		status = *s
	}
	if len(override) > 0 {
		status = override[0]
	}
	return fmt.Sprintf("%s%s%s", icon, status, icon)
}

func itemAttachment(item scraper.CarItemForMessage, override ...string) Attachment {
	label := ""
	if s := item.Status(); s != nil {
		label = *s
	}
	if len(override) > 0 {
		label = override[0]
	}
	dec := item.Decoration()
	return Attachment{
		Color:    dec.Color,
		Fields:   buildFields(item),
		Fallback: fmt.Sprintf("%s %s", label, item.CarType),
		MrkdwnIn: []string{"fields"},
	}
}

func buildFields(item scraper.CarItemForMessage) []Field {
	return []Field{
		{Title: "区間", Value: sectionField(item)},
		{Title: "車種", Value: item.CarType},
		{Title: "出発店舗", Value: item.StartShop},
		{Title: "返却エリア", Value: returnShopField(item)},
		{Title: "期間", Value: item.Period},
		{Title: "条件", Value: item.Condition},
		{Title: "電話", Value: item.Tel},
	}
}

func sectionField(item scraper.CarItemForMessage) string  {
	return fmt.Sprintf("%s %s → %s %s", item.StartCity(), item.StartCityIcon(), item.ReturnCity(), item.ReturnCityIcon())
}

func returnShopField(item scraper.CarItemForMessage) string {
	url := item.ReturnShopURL()
	if url == "" {
		return item.ReturnShop
	}
	parts := strings.Fields(item.ReturnShop)
	company := parts[0]
	return fmt.Sprintf("%s（返却可能店舗は<%s|こちら>）", company, url)
}
