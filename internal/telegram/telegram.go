package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Bot struct {
	token  string
	client *http.Client
}

func New(token string) *Bot {
	return &Bot{
		token:  token,
		client: &http.Client{Timeout: 40 * time.Second},
	}
}

func (b *Bot) apiURL(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method)
}

func (b *Bot) SendMessage(chatID, text string) error {
	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", text)
	form.Set("parse_mode", "Markdown")

	resp, err := b.client.PostForm(b.apiURL("sendMessage"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendMessage failed: %s: %s", resp.Status, string(body))
	}
	return nil
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type updatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

func (b *Bot) GetUpdates(offset int) ([]Update, error) {
	form := url.Values{}
	form.Set("offset", fmt.Sprintf("%d", offset))
	form.Set("timeout", "30")

	resp, err := b.client.PostForm(b.apiURL("getUpdates"), form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed updatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if !parsed.OK {
		return nil, fmt.Errorf("getUpdates: not ok")
	}
	return parsed.Result, nil
}
