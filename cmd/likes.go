package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/paolo/x-cli/internal/api"
	"github.com/paolo/x-cli/internal/auth"
	"github.com/paolo/x-cli/internal/models"
	"github.com/paolo/x-cli/internal/output"
	"github.com/spf13/cobra"
)

var likesCount int
var likesCursor string
var likesAll bool
var likesMaxPages int

var likesCmd = &cobra.Command{
	Use:   "likes [@handle]",
	Short: "List a user's liked tweets (defaults to your own)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		userID, err := likesUserID(client, args)
		if err != nil {
			return err
		}

		if likesAll {
			return paginateLikes(client, func(cursor string) (models.TimelinePage, []byte, error) {
				return client.GetLikes(context.Background(), userID, likesCount, cursor)
			})
		}

		page, rawJSON, err := client.GetLikes(context.Background(), userID, likesCount, likesCursor)
		if err != nil {
			return fmt.Errorf("fetch likes: %w", err)
		}

		output.PrintTweets(page.Tweets, jsonOutput, rawJSON)
		if !jsonOutput {
			output.PrintCursor("x-cli likes", page.NextCursor)
		}
		return nil
	},
}

func likesUserID(client *api.Client, args []string) (string, error) {
	if len(args) == 0 {
		creds, err := auth.Load()
		if err != nil {
			return "", err
		}
		return creds.UserID, nil
	}

	handle := strings.TrimPrefix(args[0], "@")
	userID, err := client.ResolveUserID(context.Background(), handle)
	if err != nil {
		return "", fmt.Errorf("resolve @%s: %w", handle, err)
	}
	return userID, nil
}

func paginateLikes(client *api.Client, fetch timelineFetcher) error {
	cursor := likesCursor
	total := 0

	for page := 0; page < likesMaxPages; page++ {
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

		// Like other timelines, X can end the pagination with a cursor-only
		// page whose Bottom cursor repeats the previous one. Terminate on the
		// cursor no longer advancing rather than on an empty tweet list.
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
		fmt.Printf("\nFetched %d likes.\n", total)
	}
	return nil
}

func init() {
	likesCmd.Flags().IntVar(&likesCount, "count", 20, "Number of liked tweets per page")
	likesCmd.Flags().StringVar(&likesCursor, "cursor", "", "Pagination cursor")
	likesCmd.Flags().BoolVar(&likesAll, "all", false, "Auto-paginate through all results")
	likesCmd.Flags().IntVar(&likesMaxPages, "max-pages", 10, "Max pages when using --all")
	rootCmd.AddCommand(likesCmd)
}
