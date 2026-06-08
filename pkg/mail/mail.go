package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	klog "github.com/hanasakis/kotoha/pkg/log"
)

type Client struct {
	apiKey   string
	from     string
	fromName string
	client   *http.Client
}

type emailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func New(host, port, username, password, from, fromName string) *Client {
	return &Client{
		apiKey:   password,
		from:     from,
		fromName: fromName,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) Send(to, subject, htmlBody string) error {
	if c.apiKey == "" {
		klog.Warnf("[mail] API key not configured, skipping send to %s (subject: %s)", to, subject)
		return fmt.Errorf("mail: api key not configured")
	}

	fromAddr := fmt.Sprintf("%s <%s>", c.fromName, c.from)
	payload := emailPayload{
		From:    fromAddr,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		klog.Warnf("[mail] failed to send email to %s: %v", to, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		klog.Warnf("[mail] API error for %s (HTTP %d): %s", to, resp.StatusCode, string(respBody))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	klog.Infof("[mail] email sent to %s (subject: %s)", to, subject)
	return nil
}

func (c *Client) SendPasswordReset(to, token string) error {
	resetURL := "http://localhost:3000/reset-password?token=" + token
	klog.Infof("[mail] password reset link for %s: %s", to, resetURL)

	subject := "Kotoha Password Reset"
	body := "<html><body>\n<p>You requested a password reset for your Kotoha account.</p>\n<p>Click the link below to reset your password (valid for 1 hour):</p>\n<p><a href=\"" + resetURL + "\">" + resetURL + "</a></p>\n<p>If you did not request this, please ignore this email.</p>\n</body></html>"
	return c.Send(to, subject, body)
}
