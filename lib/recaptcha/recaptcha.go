package recaptcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

type RecaptchaResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

func VerifyToken(ctx context.Context, secretKey, token string) (*RecaptchaResponse, error) {
	if strings.TrimSpace(secretKey) == "" {
		return nil, errors.New("reCAPTCHA secret key is not configured")
	}

	form := url.Values{"secret": {secretKey}, "response": {token}}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://www.google.com/recaptcha/api/siteverify",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var recaptchaResp RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&recaptchaResp); err != nil {
		return nil, err
	}

	return &recaptchaResp, nil
}
