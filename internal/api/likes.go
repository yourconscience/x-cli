package api

import (
	"context"

	"github.com/paolo/x-cli/internal/models"
	"github.com/tidwall/gjson"
)

// GetLikes fetches tweets liked by a specific user.
func (c *Client) GetLikes(ctx context.Context, userID string, count int, cursor string) (models.TimelinePage, []byte, error) {
	vars := map[string]interface{}{
		"userId":                 userID,
		"count":                  count,
		"includePromotedContent": false,
		"withClientEventToken":   false,
		"withBirdwatchNotes":     false,
		"withVoice":              true,
		"withV2Timeline":         true,
	}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	data, err := c.GraphQL(ctx, Endpoints["Likes"], vars)
	if err != nil {
		return models.TimelinePage{}, nil, err
	}

	instructions := gjson.GetBytes(data, "data.user.result.timeline.timeline.instructions")
	page := models.ParseTimeline(instructions)
	return page, data, nil
}
