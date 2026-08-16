package api

import (
	"context"

	"github.com/paolo/x-cli/internal/models"
	"github.com/tidwall/gjson"
)

// GetBookmarks fetches the authenticated user's bookmarked tweets.
func (c *Client) GetBookmarks(ctx context.Context, count int, cursor string) (models.TimelinePage, []byte, error) {
	vars := map[string]interface{}{
		"count":                  count,
		"includePromotedContent": false,
	}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	data, err := c.GraphQL(ctx, Endpoints["Bookmarks"], vars)
	if err != nil {
		return models.TimelinePage{}, nil, err
	}

	instructions := gjson.GetBytes(data, "data.bookmark_timeline_v2.timeline.instructions")
	page := models.ParseTimeline(instructions)
	return page, data, nil
}
