package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// pubDateLayouts are the formats feeds use for <pubDate>, tried in order.
var pubDateLayouts = []string{
	time.RFC1123Z, // Mon, 02 Jan 2006 15:04:05 -0700
	time.RFC1123,  // Mon, 02 Jan 2006 15:04:05 MST
	time.RFC822Z,  // 02 Jan 06 15:04 -0700
	time.RFC822,   // 02 Jan 06 15:04 MST
	time.RFC3339,  // 2006-01-02T15:04:05Z07:00
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// parsePubDate converts a feed's pubDate string into a nullable timestamp.
// An empty or unrecognized value yields NULL rather than an error: a post is
// still worth storing even when we can't tell exactly when it was published.
func parsePubDate(pubDate string) sql.NullTime {
	pubDate = strings.TrimSpace(pubDate)
	if pubDate == "" {
		return sql.NullTime{}
	}

	for _, layout := range pubDateLayouts {
		t, err := time.Parse(layout, pubDate)
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}

	return sql.NullTime{}
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")

	httpClient := http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch %s: %w", feedURL, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read response body: %w", err)
	}

	var rssFeed RSSFeed
	err = xml.Unmarshal(data, &rssFeed)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse feed: %w", err)
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}

	return &rssFeed, nil
}
