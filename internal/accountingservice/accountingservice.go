package accountingservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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

//special struct to hold the response of GET file upload url data. to further POST in aws of green invoice

type UploadURLFields struct {
	Key                 string `json:"key"`
	Bucket              string `json:"bucket"`
	XAmzAlgorithm       string `json:"X-Amz-Algorithm"`
	XAmzCredential      string `json:"X-Amz-Credential"`
	XAmzDate            string `json:"X-Amz-Date"`
	XAmzSecurityToken   string `json:"X-Amz-Security-Token"`
	Policy              string `json:"Policy"`
	XAmzSignature       string `json:"X-Amz-Signature"`
	XAmzMetaAccountID   string `json:"x-amz-meta-account-id"`
	XAmzMetaUserID      string `json:"x-amz-meta-user-id"`
	XAmzMetaBusinessID  string `json:"x-amz-meta-business-id"`
	XAmzMetaFileContext string `json:"x-amz-meta-file-context"`
	XAmzMetaFileData    string `json:"x-amz-meta-file-data"`
}

type UploadURLResponse struct {
	URL    string          `json:"url"`
	Fields UploadURLFields `json:"fields"`
}

func New(cfg config.AccountingConfig) (*Service, error) {
	service := &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
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

func (s *Service) getUploadURL(ctx context.Context) (*UploadURLResponse, error) {
	token, err := s.getToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", s.cfg.BaseURL+"/expenses/file", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload URL request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer: "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute upload URL request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("get upload URL failed with status: %s", resp.Status)
	}

	var uploadURLresp UploadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadURLresp); err != nil {
		return nil, fmt.Errorf("failed to decode upload URL response: %w", err)
	}
	return &uploadURLresp, nil
}

func (s *Service) StagedInvoiceFile(ctx context.Context, filename string, fileData []byte) error {
	log.Println("getting pre-signed URL for invoice upload...")
	uploadConfig, err := s.getUploadURL(ctx)
	if err != nil {
		return fmt.Errorf("failed to get upload config: %w", err)
	}
	log.Printf("Uploading file %s to %s", filename, uploadConfig.URL)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fields := map[string]string{
		"key":                     uploadConfig.Fields.Key,
		"bucket":                  uploadConfig.Fields.Bucket,
		"X-Amz-Algorithm":         uploadConfig.Fields.XAmzAlgorithm,
		"X-Amz-Credential":        uploadConfig.Fields.XAmzCredential,
		"X-Amz-Date":              uploadConfig.Fields.XAmzDate,
		"X-Amz-Security-Token":    uploadConfig.Fields.XAmzSecurityToken,
		"Policy":                  uploadConfig.Fields.Policy,
		"X-Amz-Signature":         uploadConfig.Fields.XAmzSignature,
		"x-amz-meta-account-id":   uploadConfig.Fields.XAmzMetaAccountID,
		"x-amz-meta-user-id":      uploadConfig.Fields.XAmzMetaUserID,
		"x-amz-meta-business-id":  uploadConfig.Fields.XAmzMetaBusinessID,
		"x-amz-meta-file-context": uploadConfig.Fields.XAmzMetaFileContext,
		"x-amz-meta-file-data":    uploadConfig.Fields.XAmzMetaFileData,
	}

	for key, val := range fields {
		if err = w.WriteField(key, val); err != nil {
			return fmt.Errorf("Failed to write field %s, %w", key, err)
		}
	}

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := fw.Write(fileData); err != nil {
		return fmt.Errorf("failed to write file data to form: %w", err)
	}

	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", uploadConfig.URL, &b)
	if err != nil {
		return fmt.Errorf("failed to create final upload request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute final upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("final upload request failed with status: %s: %s", resp.Status, string(body))
	}
	log.Printf("Successfully staged invoice file: %s", filename)
	return nil
}
