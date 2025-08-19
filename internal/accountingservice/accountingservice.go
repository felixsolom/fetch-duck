package accountingservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/felixsolom/fetch-duck/internal/config"
)

type Service struct {
	cfg         config.AccountingConfig
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

func New(cfg config.AccountingConfig) (*Service, error) {
	service := &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	err := service.refreshToken(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get initial accounting token: %w", err)
	}
	return service, nil
}

func (s *Service) refreshToken(ctx context.Context) error {

	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()

	reqBody, err := json.Marshal(map[string]string{
		"id":     s.cfg.APIKey,
		"secret": s.cfg.APISecret,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal token request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx,
		"POST",
		s.cfg.BaseURL+"/account/token",
		bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("token request failed with status: %s", resp.Status)
	}

	var tokenResponse struct {
		Token   string `json:"token"`
		Expires int64  `json:"expires"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	s.token = tokenResponse.Token
	s.tokenExpiry = time.Unix(tokenResponse.Expires, 0)
	log.Printf("Successfully refreshed API token. New expiry: %s",
		s.tokenExpiry.Format(time.RFC1123))
	return nil
}

func (s *Service) getToken(ctx context.Context) (string, error) {
	s.tokenMutex.RLock()
	isExpired := time.Now().After(s.tokenExpiry.Add(-60 * time.Second))
	s.tokenMutex.RUnlock()

	if isExpired {
		log.Println("Accounting token is exipered, or about to expire. Refreshing...")
	}
	if err := s.refreshToken(ctx); err != nil {
		return "", err
	}

	s.tokenMutex.RLock()
	defer s.tokenMutex.RUnlock()
	return s.token, nil
}
