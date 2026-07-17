package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const telegramAPI = "https://api.telegram.org/bot"

// Sender posts alert messages to Telegram.
type Sender interface {
	Send(text string) error
}

// TelegramClient sends messages via the Telegram Bot API.
type TelegramClient struct {
	Token  string
	ChatID string
	Client *http.Client
}

func NewTelegramClient(token, chatID string) *TelegramClient {
	return &TelegramClient{
		Token:  token,
		ChatID: chatID,
		Client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *TelegramClient) Send(text string) error {
	body, err := json.Marshal(map[string]string{
		"chat_id":    c.ChatID,
		"text":       text,
		"parse_mode": "HTML",
	})
	if err != nil {
		return err
	}
	url := telegramAPI + c.Token + "/sendMessage"
	resp, err := c.Client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("telegram send: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	return nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
