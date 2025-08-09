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
	pageToken := ""

	log.Printf("Performing Gmail scan for user %s with query: %s", user, query)
	for {
		req := s.Users.Messages.List(user).Q(query)
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("failed to retrieve messages: %w", err)
		}
		if len(resp.Messages) == 0 {
			fmt.Errorf("no messages found matching the query.")
			break
		}
		log.Printf("Processing %d message(s) on this page...", len(resp.Messages))

		for _, msg := range resp.Messages {
			fullMsg, err := s.Users.Messages.Get("me", msg.Id).Format("full").Do()
			if err != nil {
				log.Printf("Failed to get full message details for %s: %v", msg.Id, err)
				continue
			}

			for _, part := range fullMsg.Payload.Parts {
				if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
					log.Printf("Found Attachment! MessageID: %s, Filemane: %s, AtachmentID: %s",
						fullMsg.Id, part.Filename, part.Body.AttachmentId)
				}
			}
		}
		if resp.NextPageToken != "" {
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}
	log.Printf("Finished scanning all pages")
	return nil
}
