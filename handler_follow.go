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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	url := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't find feed with url %s: %w", url, err)
	}

	_, err = s.db.GetFeedFollowForUser(context.Background(), database.GetFeedFollowForUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err == nil {
		return fmt.Errorf("you already follow %s", feed.Name)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't look up feed follow: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w", err)
	}

	fmt.Printf("%s is now followed by %s\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("couldn't get feed follows: %w", err)
	}

	for _, feedFollow := range feedFollows {
		fmt.Printf("* %s\n", feedFollow.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	url := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't find feed with url %s: %w", url, err)
	}

	_, err = s.db.GetFeedFollowForUser(context.Background(), database.GetFeedFollowForUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("you don't follow %s", feed.Name)
	}
	if err != nil {
		return fmt.Errorf("couldn't look up feed follow: %w", err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
	}

	fmt.Printf("%s unfollowed by %s\n", feed.Name, user.Name)
	return nil
}
