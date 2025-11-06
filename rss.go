package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sssseraphim/gator/internal/database"
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res RSSFeed
	err = xml.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	res.Channel.Title = html.UnescapeString(res.Channel.Title)
	res.Channel.Description = html.UnescapeString(res.Channel.Description)
	for i := range len(res.Channel.Item) {
		res.Channel.Item[i].Title = html.UnescapeString(res.Channel.Item[i].Title)
		res.Channel.Item[i].Description = html.UnescapeString(res.Channel.Item[i].Description)
	}
	return &res, nil
}

func scrapeFeeds(s *state, ctx context.Context) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:            feed.ID,
	})
	if err != nil {
		return err
	}
	rssFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return err
	}

	for _, item := range rssFeed.Channel.Item {
		pubtime, err := parseRSSTime(item.PubDate)
		if err != nil {
			continue
		}
		_, err = s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: sql.NullTime{Time: pubtime, Valid: true},
			FeedID:      feed.ID,
		})
		var pqErr *pq.Error
		if err != nil {
			if errors.As(err, &pqErr) {
				if pqErr.Code == "23505" && pqErr.Constraint == "posts_url_key" {
					continue
				}
			}
			fmt.Printf("\n%v", err)
		}
	}
	return nil
}

func parseRSSTime(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	layouts := []string{
		time.RFC1123,     // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC1123Z,    // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC822,      // "02 Jan 06 15:04 MST"
		time.RFC822Z,     // "02 Jan 06 15:04 -0700"
		time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",   // Single digit day
		"Mon, 2 Jan 2006 15:04:05 -0700", // Single digit day with offset
		"January 2, 2006 3:04 PM",
		"Jan 2, 2006 3:04 PM",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", dateStr)
}
