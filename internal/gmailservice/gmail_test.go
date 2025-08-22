package gmailservice

import (
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestFindAttachmentPart(t *testing.T) {
	testCases := []struct {
		name             string
		payload          *gmail.MessagePart
		expectedFound    bool
		expectedID       string
		expectedFilename string
	}{
		{
			name: "Simple Case - Top Level Attachment",
			payload: &gmail.MessagePart{
				Parts: []*gmail.MessagePart{
					{Filename: "invoice.pdf", Body: &gmail.MessagePartBody{AttachmentId: "ATTACH_ID_1"}},
					{Filename: "", Body: &gmail.MessagePartBody{}},
				},
			},
			expectedFound:    true,
			expectedID:       "ATTACH_ID_1",
			expectedFilename: "invoice.pdf",
		},
		{
			name: "Nested Case - Attachment inside multipart/mixed",
			payload: &gmail.MessagePart{
				Parts: []*gmail.MessagePart{
					{
						Parts: []*gmail.MessagePart{
							{Filename: "", Body: &gmail.MessagePartBody{}},
							{Filename: "", Body: &gmail.MessagePartBody{}},
						},
					},
					{
						Filename: "receipt.pdf",
						Body:     &gmail.MessagePartBody{AttachmentId: "ATTACH_ID_2"},
					},
				},
			},
			expectedFound:    true,
			expectedID:       "ATTACH_ID_2",
			expectedFilename: "receipt.pdf",
		},
		{
			name: "Deeply Nested Case",
			payload: &gmail.MessagePart{
				Parts: []*gmail.MessagePart{
					{
						Parts: []*gmail.MessagePart{
							{
								Parts: []*gmail.MessagePart{
									{Filename: "deep-invoice.pdf", Body: &gmail.MessagePartBody{AttachmentId: "ATTACH_ID_3"}},
								},
							},
						},
					},
				},
			},
			expectedFound:    true,
			expectedID:       "ATTACH_ID_3",
			expectedFilename: "deep-invoice.pdf",
		},
		{
			name: "No Attachment Case",
			payload: &gmail.MessagePart{
				Parts: []*gmail.MessagePart{
					{Filename: "", Body: &gmail.MessagePartBody{}},
					{Filename: "inline-image.jpg", Body: &gmail.MessagePartBody{AttachmentId: ""}}, // image, not a real attachment
				},
			},
			expectedFound: false,
		},
		{
			name:          "Empty Payload",
			payload:       &gmail.MessagePart{},
			expectedFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			foundPart, foundID := findAttachmemtPart(tc.payload)
			if (foundPart != nil) != tc.expectedFound {
				t.Errorf("Expected to find attachment: %v, but got: %v", tc.expectedFound, (foundPart != nil))
			}

			if tc.expectedFound {
				if foundID != tc.expectedID {
					t.Errorf("Expected attachment ID %s, but got %s", tc.expectedID, foundID)
				}
				if foundPart.Filename != tc.expectedFilename {
					t.Errorf("Expected filename %s, but got %s", tc.expectedFilename, foundPart.Filename)
				}
			}
		})
	}
}
