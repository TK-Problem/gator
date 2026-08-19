package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TK-Problem/gator/internal/database"
	"github.com/google/uuid"
)

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't find current user %s: %w", s.cfg.CurrentUserName, err)
	}

	_, err = s.db.GetFeedByURL(context.Background(), url)
	if err == nil {
		return fmt.Errorf("feed with url %s already exists", url)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't look up feed %s: %w", url, err)
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't follow new feed: %w", err)
	}

	fmt.Println("Feed created successfully:")
	fmt.Printf("ID:         %s\n", feed.ID)
	fmt.Printf("Created At: %s\n", feed.CreatedAt)
	fmt.Printf("Updated At: %s\n", feed.UpdatedAt)
	fmt.Printf("Name:       %s\n", feed.Name)
	fmt.Printf("URL:        %s\n", feed.Url)
	fmt.Printf("User ID:    %s\n", feed.UserID)
	fmt.Printf("%s is now followed by %s\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}
