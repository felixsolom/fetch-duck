package gmailservice

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Service struct {
	*gmail.Service
}

func New(client *http.Client) (*Service, error) {
	gmailService, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service %w", err)
	}
	return &Service{gmailService}, nil
}

func (s *Service) GetMessage(messageID string) (*gmail.Message, error) {
	return s.Users.Messages.Get("me", messageID).Format("metadata").Do()
}

func (s *Service) ScanForInvoices(ctx context.Context) error {
	user := "me"
	query := "has:attachment invoice"

	log.Printf("Performing Gmail scan for user %s with query: %s", user, query)
	req := s.Users.Messages.List(user).Q(query)
	resp, err := req.Do()
	if err != nil {
		return fmt.Errorf("failed to retrieve messages: %w", err)
	}
	if len(resp.Messages) == 0 {
		return fmt.Errorf("No messages found matching the query.")
	}
	log.Printf("Found %d message(s):", len(resp.Messages))
	for _, msg := range resp.Messages {
		fullMsg, err := s.GetMessage(msg.Id)
		if err != nil {
			log.Printf("Failed to get message %s: %v", msg.Id, err)
		}
		log.Printf("- Message ID: %s, Message Snippet: %s", fullMsg.Id, fullMsg.Snippet)
	}
	return nil
}
