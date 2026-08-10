package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type WebhookOTPSender struct {
	url    string
	apiKey string
	client *http.Client
}

func NewWebhookOTPSender(url, apiKey string) (*WebhookOTPSender, error) {
	url = strings.TrimSpace(url)
	apiKey = strings.TrimSpace(apiKey)
	if url == "" || apiKey == "" {
		return nil, ErrUnsafeConfig
	}
	return &WebhookOTPSender{
		url:    url,
		apiKey: apiKey,
		client: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func newWebhookOTPSenderForTest(url, apiKey string, client *http.Client) *WebhookOTPSender {
	return &WebhookOTPSender{url: url, apiKey: apiKey, client: client}
}

func (s *WebhookOTPSender) Send(ctx context.Context, phone, code string) error {
	payload, err := json.Marshal(map[string]string{
		"phone":   phone,
		"code":    code,
		"message": "Ma OTP FlashX cua ban la " + code + ". Khong chia se ma nay cho bat ky ai.",
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send otp: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("send otp: provider status %d", resp.StatusCode)
	}
	return nil
}
