package worker

import (
	"fmt"

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
		return fmt.Errorf("slack: %w", err)
	}

	current := make(map[string]scraper.CarItem, len(items))
	for _, item := range items {
		current[item.Key()] = item
	}
	scraper.ApplyDiff(state, d, current)
	return storage.Save(statePath, state)
}
