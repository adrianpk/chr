package workflow

import (
	"fmt"
	"net/url"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func prepareTicketReviewReference(content string, document TicketDocument, request TicketEditRequest) (string, error) {
	if request.ReviewRef == nil {
		return content, nil
	}

	if reason := document.ReadOnly["review_refs"]; reason != "" {
		return "", fmt.Errorf("review references are read-only: %s", reason)
	}

	if err := ticketcore.ValidateReviewRef(*request.ReviewRef); err != nil {
		return "", err
	}

	parsed, err := url.Parse(request.ReviewRef.URL)
	if err != nil || !validTicketPullRequestURL(parsed) {
		return "", fmt.Errorf("review reference requires an HTTP URL")
	}

	return updateTicketReviewMetadataContent(content, UpdateTicketReviewMetadataOptions{ReviewRef: request.ReviewRef, UpdatedAt: time.Now().UTC()})
}
