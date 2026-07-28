package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"p-mon-backend/internal/config"
)

func SendEmail(to, subject, body string) error {
	cfg := config.Cfg.SMTP
	if cfg.Host == "" {
		return fmt.Errorf("SMTP not configured")
	}
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return smtp.SendMail(addr, auth, cfg.From, []string{to}, msg)
}

// papiBaseURL is the PAPI endpoint host. Overridable via config.WhatsApp.APIURL.
const papiBaseURL = "https://api.papi.api.br"

// SendWhatsAppPAPI sends a text message through the PAPI provider.
//
//	POST {base}/api/instances/{instance}/send-text
//	Headers: x-api-key, Content-Type: application/json
//	Body: { "jid": "...", "text": "..." }
func SendWhatsAppPAPI(instance, apiKey, jid, text string) error {
	if instance == "" || apiKey == "" || jid == "" {
		return fmt.Errorf("PAPI channel incomplete (instance, api_key and jid are required)")
	}

	base := papiBaseURL
	if config.Cfg.WhatsApp.APIURL != "" {
		base = strings.TrimRight(config.Cfg.WhatsApp.APIURL, "/")
	}

	endpoint := fmt.Sprintf("%s/api/instances/%s/send-text", base, url.PathEscape(instance))
	payload := map[string]string{
		"jid":  jid,
		"text": text,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PAPI returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func SendWebhook(url, payloadJSON string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(payloadJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
