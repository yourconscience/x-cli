package cmd

import (
	"context"
	"fmt"

	"github.com/paolo/x-cli/internal/api"
	"github.com/paolo/x-cli/internal/models"
	"github.com/paolo/x-cli/internal/output"
	"github.com/spf13/cobra"
)

var bookmarksCount int
var bookmarksCursor string
var bookmarksAll bool
var bookmarksMaxPages int

var bookmarksCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "List your bookmarked tweets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		if bookmarksAll {
			return paginateBookmarks(client, func(cursor string) (models.TimelinePage, []byte, error) {
				return client.GetBookmarks(context.Background(), bookmarksCount, cursor)
			})
		}

		page, rawJSON, err := client.GetBookmarks(context.Background(), bookmarksCount, bookmarksCursor)
		if err != nil {
			return fmt.Errorf("fetch bookmarks: %w", err)
		}

		output.PrintTweets(page.Tweets, jsonOutput, rawJSON)
		if !jsonOutput {
			output.PrintCursor("x-cli bookmarks", page.NextCursor)
		}
		return nil
	},
}

func paginateBookmarks(client *api.Client, fetch timelineFetcher) error {
	cursor := bookmarksCursor
	total := 0

	for page := 0; page < bookmarksMaxPages; page++ {
		result, rawJSON, err := fetch(cursor)
		if err != nil {
			return err
		}

		if len(result.Tweets) > 0 {
			if jsonOutput {
				output.PrintTweets(nil, true, rawJSON)
			} else {
				output.PrintTweets(result.Tweets, false, nil)
			}
			total += len(result.Tweets)
		}

		// X ends the bookmark timeline with a cursor-only page whose Bottom
		// cursor repeats the previous one. Terminate on the cursor no longer
		// advancing rather than on an empty tweet list, so an occasional page
		// that parses zero tweets (e.g. all tombstones) does not cut
		// pagination short and miss later bookmarks.
		if result.NextCursor == "" || result.NextCursor == cursor {
			break
		}
		cursor = result.NextCursor

		// Respect rate limits between pages.
		if client.LastRateLimit != nil {
			client.LastRateLimit.WaitIfNeeded()
		}
	}

	if !jsonOutput {
		fmt.Printf("\nFetched %d bookmarks.\n", total)
	}
	return nil
}

func init() {
	bookmarksCmd.Flags().IntVar(&bookmarksCount, "count", 20, "Number of bookmarks per page")
	bookmarksCmd.Flags().StringVar(&bookmarksCursor, "cursor", "", "Pagination cursor")
	bookmarksCmd.Flags().BoolVar(&bookmarksAll, "all", false, "Auto-paginate through all results")
	bookmarksCmd.Flags().IntVar(&bookmarksMaxPages, "max-pages", 10, "Max pages when using --all")
	rootCmd.AddCommand(bookmarksCmd)
}
