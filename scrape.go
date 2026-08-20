package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/TK-Problem/gator/internal/database"
	"github.com/google/uuid"
)

// scrapeFeeds fetches the least recently fetched feed and saves its posts.
// It logs its own failures instead of returning them so that the agg loop
// keeps running when a single feed or request goes bad.
func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println("couldn't get next feed to fetch:", err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("couldn't fetch feed %s: %v", feed.Name, err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: parsePubDate(item.PubDate),
			FeedID:      feed.ID,
		})
		if err != nil {
			// Duplicate URLs are swallowed by ON CONFLICT DO NOTHING, so
			// anything landing here is a real problem worth seeing.
			log.Printf("couldn't save post %s: %v", item.Link, err)
			continue
		}
	}

	fmt.Printf("Feed %s collected, %d posts found\n", feed.Name, len(rssFeed.Channel.Item))
}
