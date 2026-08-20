package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/TK-Problem/gator/internal/database"
)

const defaultBrowseLimit = 2

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := defaultBrowseLimit
	switch len(cmd.Args) {
	case 0:
	case 1:
		parsed, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit %q: must be a number", cmd.Args[0])
		}
		if parsed < 1 {
			return fmt.Errorf("limit must be at least 1, got %d", parsed)
		}
		limit = parsed
	default:
		return fmt.Errorf("usage: %s [limit]", cmd.Name)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found. Follow a feed and run agg to collect some.")
		return nil
	}

	for _, post := range posts {
		published := "unknown date"
		if post.PublishedAt.Valid {
			published = post.PublishedAt.Time.Format("Mon Jan 2 15:04")
		}
		fmt.Printf("* %s\n", post.Title)
		fmt.Printf("  %s | %s\n", post.FeedName, published)
		fmt.Printf("  %s\n", post.Url)
	}

	return nil
}
