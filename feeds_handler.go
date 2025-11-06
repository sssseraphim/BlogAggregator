package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sssseraphim/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	res := func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), sql.NullString{String: s.cfg.Current_user_name, Valid: true})
		if err != nil {
			return fmt.Errorf("failed to get current user: %v", err)
		}
		err = handler(s, cmd, user)
		if err != nil {
			return err
		}
		return nil
	}
	return res
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("no time given")
	}
	time_between_reqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("wrong time given: %v", err)
	}
	fmt.Printf("Collecting feeds every %v/n", time_between_reqs)
	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s, context.Background())
		if err != nil {
			fmt.Printf("Failed to fetch: %v", err)
		}
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("add command takes 2 arguments: name and url")
	}
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create a feed: %v", err)
	}
	fmt.Printf("Success! %v %v %v %v %v %v\n", feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.ID)
	err = handlerFollow(s, command{args: cmd.args[1:]}, user)
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to list feeds: %v", err)
	}
	for _, feed := range feeds {
		fmt.Printf("Feed: %v Url: %v User: %v\n", feed.Name, feed.Url, feed.Name_2.String)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("follow command takes 1 argument: url")
	}
	url := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Failed getting feed: %v", err)
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to follow: %v", err)
	}
	fmt.Printf("Success! %v follows %v", feedFollow.UserName.String, feedFollow.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("Coudn't get follows: %v", err)
	}
	for _, feed := range feeds {
		fmt.Printf("- %v\n", feed.Feed)
	}
	return nil

}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("unfollow takes 1 argument: url")
	}
	url := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to get a feed: %v", err)
	}
	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow: %v", err)
	}
	fmt.Printf("Unfollowed %v", feed.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) >= 1 {
		l, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			limit = 2
		} else {
			limit = l
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts: %v", err)
	}
	fmt.Println(len(posts))
	for _, post := range posts {
		fmt.Printf("%v\n%v\n%v\n%v\n\n", post.Title, post.Description.String, post.PublishedAt.Time, post.Url)
	}
	return nil

}
