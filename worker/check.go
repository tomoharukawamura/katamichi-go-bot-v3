package worker

import (
	"log"

	"github.com/tomok/katamichi-go-bot-v3/notifier"
	"github.com/tomok/katamichi-go-bot-v3/scraper"
	"github.com/tomok/katamichi-go-bot-v3/storage"
)

type slackNotifier interface {
	Notify(scraper.Diff, *storage.State, *notifier.ChannelConfig) error
}

type scanFn func(string) (scraper.Diff, *storage.State, []scraper.CarItem, error)

func Check(slack *notifier.Slack, ch *notifier.ChannelConfig, statePath string) error {
	return runCheck(slack, ch, statePath, scraper.Scan)
}

func runCheck(slack slackNotifier, ch *notifier.ChannelConfig, statePath string, scan scanFn) error {
	d, state, items, err := scan(statePath)
	if err != nil {
		return err
	}

	if !d.HasChange() {
		return nil
	}

	if err := slack.Notify(d, state, ch); err != nil {
		log.Printf("slack notify error (continuing): %v", err)
	}

	current := make(map[string]scraper.CarItem, len(items))
	for _, item := range items {
		current[item.Key()] = item
	}
	scraper.ApplyDiff(state, d, current)
	if err := storage.Save(statePath, state); err != nil {
		return err
	}

	log.Printf("diff: added=%d reopened=%d updated=%d soldout=%d (fetched=%d active=%d ts=%d)",
		len(d.Added), len(d.Reopened), len(d.Updated), len(d.SoldOut),
		len(items), len(state.Active), len(state.MessageTS))
	for _, item := range d.Added {
		log.Printf("  [added]    %s", item.Key())
	}
	for _, item := range d.Reopened {
		log.Printf("  [reopened] %s (storedTS=%q)", item.Key(), state.MessageTS[item.Key()])
	}
	return nil
}
