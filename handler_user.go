package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/database"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 1 {
		return errors.New("expected arg")
	}

	username := cmd.args[1]
	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := s.config.SetUser(username); err != nil {
		return err
	}

	fmt.Println("User has been set: ", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return errors.New("expected arg")
	}
	userName := cmd.args[1]

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	}
	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.config.SetUser(userName)
	fmt.Println("User was created")
	fmt.Println(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) > 1 {
		return errors.New("reset takes no args")
	}

	if err := s.db.ClearUsers(context.Background()); err != nil {
		return err
	}
	fmt.Println("users table cleared")
	return nil
}

func handlerListUsers(s *state, cmd command) error {
	if len(cmd.args) > 1 {
		return errors.New("users takes no args")
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, name := range users {
		if name == s.config.CurrentUserName {
			fmt.Println("* ", name, "(current)")
		} else {
			fmt.Println("* ", name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return errors.New("agg takes one arg")
	}

	time_between_reqs := cmd.args[1]
	duration, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %s", time_between_reqs)

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
		fmt.Println("req made")
	}

}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 3 {
		return errors.New("addfeed takes two args: name, url")
	}
	name := cmd.args[1]
	url := cmd.args[2]

	feed_params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), feed_params)
	if err != nil {
		return err
	}

	ff_params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    feed_params.UserID,
		FeedID:    feed_params.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), ff_params)
	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	if len(cmd.args) > 1 {
		return errors.New("feeds takes no args")
	}

	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Feeds in database:")
	fmt.Println("--------------------------------")
	for _, feed := range feeds {
		name, err := s.db.GetUserIDName(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(name)
		fmt.Println("--------------------------------")
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return errors.New("follow takes one arg")
	}

	url := cmd.args[1]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	res, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("%s is now following %s\n", res.UserName, res.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 1 {
		return errors.New("following takes no args")
	}

	feed_follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("%s is following:\n", user.Name)
	for _, feed_follow := range feed_follows {
		fmt.Println(feed_follow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return errors.New("unfollow takes one arg")
	}

	url := cmd.args[1]

	if err := s.db.UnfollowFeed(context.Background(), database.UnfollowFeedParams{UserID: user.ID, Url: url}); err != nil {
		return err
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 2 {
		return errors.New("browse takes one optional arg")
	}
	limit := 2
	if len(cmd.args) == 2 {
		val, err := strconv.Atoi(cmd.args[1])
		if err != nil {
			return err
		}
		limit = val
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit)})
	if err != nil {
		return err
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
