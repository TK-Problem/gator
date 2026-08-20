package main

import (
	"context"
	"fmt"
	"log"
)

// scrapeFeeds fetches the least recently fetched feed and prints its posts.
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

	fmt.Printf("Found %d posts in %s\n", len(rssFeed.Channel.Item), feed.Name)
	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("* %s\n", item.Title)
	}
}
