package main

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Read()
	appState := state{config: &cfg}
	cliCommands := commands{registry: make(map[string]func(*state, command) error)}
	cliCommands.register("login", handlerLogin)
	cliCommands.register("register", handlerRegister)
	cliCommands.register("reset", handlerReset)
	cliCommands.register("users", handlerListUsers)
	cliCommands.register("agg", handlerAgg)
	cliCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cliCommands.register("feeds", handlerListFeeds)
	cliCommands.register("follow", middlewareLoggedIn(handlerFollow))
	cliCommands.register("following", middlewareLoggedIn(handlerFollowing))
	cliCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cliCommands.register("browse", middlewareLoggedIn(handlerBrowse))

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Println("failed to open database")
		os.Exit(1)
	}
	dbQueries := database.New(db)
	appState.db = dbQueries

	args := os.Args
	if len(args) < 2 {
		fmt.Println("expected cli arg")
		os.Exit(1)
	}
	commandName := args[1]
	args = args[1:]
	cmd := command{
		name: commandName,
		args: args,
	}

	if err := cliCommands.run(&appState, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{UpdatedAt: time.Now(), ID: feed.ID})
	if err != nil {
		return err
	}

	fetched_feed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, feed_item := range fetched_feed.Channel.Item {
		pubDate, err := parseDate(feed_item.PubDate)
		if err != nil {
			return err
		}

		params := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       feed_item.Title,
			Url:         feed_item.Link,
			Description: sql.NullString{String: feed_item.Description},
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		}
		_, err = s.db.CreatePost(context.Background(), params)
		if err != nil {
			return err
		}

	}

	return nil
}
