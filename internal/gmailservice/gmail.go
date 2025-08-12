package gmailservice

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/google/uuid"
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

func (s *Service) ScanAndStageInvoices(ctx context.Context, db *database.Queries, userID string) error {
	user := "me"
	query := `subject:(invoice OR receipt OR "bill from") OR "invoice" OR "receipt"`
	pageToken := ""

	log.Printf("Performing Gmail scan for user ID %s with query: %s", userID, query)
	for {
		req := s.Users.Messages.List(user).Q(query)
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("failed to retrieve a page of messages: %w", err)
		}

		if len(resp.Messages) == 0 {
			log.Printf("No new messages found on this page.")
			break
		}

		for _, msg := range resp.Messages {
			// is the message already staged
			stagedInvoices, err := db.GetStagedInvoicesByMessageId(ctx, database.GetStagedInvoicesByMessageIdParams{
				UserID:         userID,
				GmailMessageID: msg.Id,
			})

			if err != nil {
				log.Printf("Error checking for existing staged invoices %s: %v", msg.Id, err)
				continue
			}

			if len(stagedInvoices) > 0 {
				log.Printf("Message %s already staged. Skipping", msg.Id)
				continue
			}

			fullMsg, err := s.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
			if err != nil {
				log.Printf("Failed to get message metadata for %s: %v", msg.Id, err)
				continue
			}

			var sender, subject string
			for _, h := range fullMsg.Payload.Headers {
				if h.Name == "From" {
					sender = h.Value
				}
				if h.Name == "Subject" {
					subject = h.Value
				}
			}

			receivedAt := fullMsg.InternalDate / 1000
			now := time.Now().Unix()

			_, err = db.CreateStagedInvoice(ctx, database.CreateStagedInvoiceParams{
				ID:             uuid.New().String(),
				UserID:         userID,
				GmailMessageID: fullMsg.Id,
				GmailThreadID:  fullMsg.ThreadId,
				Sender:         sender,
				Subject:        subject,
				Snippet: sql.NullString{
					String: fullMsg.Snippet,
					Valid:  true,
				},
				HasAttachment: len(fullMsg.Payload.Parts) > 1,
				ReceivedAt:    receivedAt,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
			if err != nil {
				log.Printf("Failed to create staged invoice for message %s: %v", msg.Id, err)
				continue
			}
			log.Printf("Successfully staged message for %s with subject: %s", sender, subject)
		}

		if resp.NextPageToken != "" {
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}
	log.Println("Finished scanning all pages")
	return nil
}
