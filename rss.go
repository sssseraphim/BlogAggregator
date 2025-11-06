package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
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
