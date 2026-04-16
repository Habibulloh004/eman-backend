package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TelegramService struct {
	botToken string
	chatID   string
	client   *http.Client
}

type telegramSendMessageResponse struct {
	OK          bool `json:"ok"`
	Description string `json:"description"`
}

func sanitizeEnvValue(v string) string {
	trimmed := strings.TrimSpace(v)
	trimmed = strings.Trim(trimmed, "\"")
	trimmed = strings.Trim(trimmed, "'")
	return strings.TrimSpace(trimmed)
}

func NewTelegramService(botToken, chatID string) *TelegramService {
	return &TelegramService{
		botToken: sanitizeEnvValue(botToken),
		chatID:   sanitizeEnvValue(chatID),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *TelegramService) Enabled() bool {
	return s != nil && s.botToken != "" && s.chatID != ""
}

func (s *TelegramService) SendMessage(text string) error {
	if !s.Enabled() {
		return nil
	}

	form := url.Values{}
	form.Set("chat_id", s.chatID)
	form.Set("text", text)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build telegram request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read telegram response failed: %w", err)
	}

	var parsed telegramSendMessageResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("parse telegram response failed: %w (body: %s)", err, string(respBody))
	}

	if !parsed.OK {
		if parsed.Description == "" {
			parsed.Description = "unknown error"
		}
		return fmt.Errorf("telegram API error: %s", parsed.Description)
	}

	return nil
}
