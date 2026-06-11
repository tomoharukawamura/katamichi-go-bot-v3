package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const slackAPIBase = "https://slack.com/api"

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short,omitempty"`
}

type Attachment struct {
	Color     string   `json:"color,omitempty"`
	Title     string   `json:"title,omitempty"`
	Fields    []Field  `json:"fields,omitempty"`
	Fallback  string   `json:"fallback,omitempty"`
	MrkdwnIn  []string `json:"mrkdwn_in,omitempty"`
}

type Slack struct {
	token string
}

func NewSlack(token string) *Slack {
	return &Slack{token: token}
}

func (s *Slack) Send(channelID, text string, attachments []Attachment) (string, error) {
	payload, err := json.Marshal(map[string]any{
		"channel":      channelID,
		"text":         text,
		"attachments":  attachments,
		"unfurl_links": false,
		"unfurl_media": false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, slackAPIBase+"/chat.postMessage", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("slack post failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		TS    string `json:"ts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("slack response decode failed: %w", err)
	}
	if !result.OK {
		return "", fmt.Errorf("slack error: %s", result.Error)
	}
	return result.TS, nil
}

func (s *Slack) Reply(channelID, threadTS, text string, attachments []Attachment) error {
	payload, err := json.Marshal(map[string]any{
		"channel":      channelID,
		"text":         text,
		"attachments":  attachments,
		"thread_ts":    threadTS,
		"unfurl_links": false,
		"unfurl_media": false,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, slackAPIBase+"/chat.postMessage", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack reply failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("slack response decode failed: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack error: %s", result.Error)
	}
	return nil
}

func (s *Slack) replyOrSend(channelID, threadTS, text string, attachments []Attachment) (newTS string, err error) {
	err = s.Reply(channelID, threadTS, text, attachments)
	if err != nil && isMessageNotFound(err) {
		newTS, err = s.Send(channelID, text, attachments)
	}
	return
}

func isMessageNotFound(err error) bool {
	return strings.Contains(err.Error(), "message_not_found")
}

func (s *Slack) Delete(channelID, ts string) error {
	payload, err := json.Marshal(map[string]string{
		"channel": channelID,
		"ts":      ts,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, slackAPIBase+"/chat.delete", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack delete failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("slack response decode failed: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack error: %s", result.Error)
	}
	return nil
}

func (s *Slack) RemoveReaction(channelID, ts, name string) error {
	return s.reactionAPI("reactions.remove", channelID, ts, name)
}

func (s *Slack) AddReaction(channelID, ts, name string) error {
	return s.reactionAPI("reactions.add", channelID, ts, name)
}

func (s *Slack) reactionAPI(method, channelID, ts, name string) error {
	payload, err := json.Marshal(map[string]string{
		"channel":   channelID,
		"timestamp": ts,
		"name":      name,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, slackAPIBase+"/"+method, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack %s failed: %w", method, err)
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("slack response decode failed: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack error: %s", result.Error)
	}
	return nil
}
