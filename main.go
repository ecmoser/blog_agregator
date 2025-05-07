package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	cfg "github.com/ecmoser/blog_aggregator/internal/config"
	"github.com/ecmoser/blog_aggregator/internal/database"
)

const dbUrl = "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"

type state struct {
	config    *cfg.Config
	dbQueries *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	functions map[string]func(s *state, cmd command) error
}

func (c *commands) run(s *state, cmd command) error {
	for name, f := range c.functions {
		if name == cmd.name {
			err := f(s, cmd)
			return err
		}
	}

	return fmt.Errorf("command %s not found", cmd.name)
}

func (c *commands) register(name string, f func(s *state, cmd command) error) {
	c.functions[name] = f
}

func handlerLogin(s *state, cmd command) error {
	wd, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if len(cmd.args) == 0 {
		return fmt.Errorf("username is required for login")
	}

	_, err = s.dbQueries.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("user %s not found", cmd.args[0])
	}

	err = s.config.SetUser(cmd.args[0], wd)
	if err != nil {
		return err
	}

	fmt.Println("The user " + cmd.args[0] + " has been set")

	return nil
}

func handlerRegister(s *state, cmd command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if len(cmd.args) == 0 {
		return fmt.Errorf("username is required for registration")
	}

	_, err = s.dbQueries.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		return fmt.Errorf("user %s already exists", cmd.args[0])
	}

	s.dbQueries.CreateUser(context.Background(), database.CreateUserParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})

	s.config.SetUser(cmd.args[0], homeDir)

	fmt.Println("The user " + cmd.args[0] + " has been registered")

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.dbQueries.EmptyUsersTable(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("The users table has been reset")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.dbQueries.GetUsers(context.Background())
	if err != nil {
		return err
	}

	output := "\n"

	for _, user := range users {
		output += "* " + user.Name
		if user.Name == s.config.CurrentUserName {
			output += " (current)"
		}
		output += "\n"
	}

	fmt.Println(output)

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Time between fetches is required")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid time duration: %s", cmd.args[0])
	}

	fmt.Println("Collecting feeds evert " + timeBetweenRequests.String())

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("adding feed requires a name and a url")
	}

	user, err := s.dbQueries.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	feed, err := s.dbQueries.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})

	if err != nil {
		return err
	}

	_, err = s.dbQueries.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	fmt.Println("The feed " + cmd.args[0] + " has been added with url " + cmd.args[1] + " for user with ID " + string(user.ID))

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	output := "\nName, URL, User\n"

	feeds, err := s.dbQueries.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		userName, err := s.dbQueries.GetUserName(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		output += feed.Name + ", " + feed.Url + ", " + userName + "\n"
	}

	fmt.Println(output)

	return nil
}

func handlerFollow(s *state, cmd command) error {
	user, err := s.dbQueries.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	if len(cmd.args) == 0 {
		return fmt.Errorf("url is required for following")
	}

	feed, err := s.dbQueries.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	_, err = s.dbQueries.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	fmt.Println("The feed " + feed.Name + " has been followed by user " + user.Name)

	return nil
}

func handlerFollowing(s *state, cmd command) error {
	feedFollows, err := s.dbQueries.GetFeedFollowsForUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	output := "\n"

	for _, ff := range feedFollows {
		output += "* " + ff.FeedName + " (" + ff.Url + ")\n"
	}

	fmt.Println(output)

	return nil
}

func handlerUnfollow(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("url is required for unfollowing")
	}

	user, err := s.dbQueries.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	feed, err := s.dbQueries.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	_, err = s.dbQueries.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	})

	if err != nil {
		return err
	}

	fmt.Println("The feed " + feed.Name + " has been unfollowed by user " + user.Name)

	return nil
}

func handlerBrowse(s *state, cmd command) error {
	var limit int32
	if len(cmd.args) == 0 {
		limit = 2
	} else {
		limit64, err := strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil {
			return err
		}
		limit = int32(limit64)
	}

	user, err := s.dbQueries.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	posts, err := s.dbQueries.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})

	for _, post := range posts {
		fmt.Println(post.Title + ": " + post.Description.String)
	}

	return nil
}

func scrapeFeeds(s *state) error {
	feed, err := s.dbQueries.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	_, err = s.dbQueries.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}

	rssFeed, err := FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, item := range rssFeed.Channel.Items {
		_, err = s.dbQueries.CreatePost(context.Background(), database.CreatePostParams{
			ID:          int32(uuid.New().ID()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: sql.NullString{String: item.PubDate, Valid: true},
			FeedID:      feed.ID,
		})
		if err != nil && !strings.Contains(err.Error(), "duplicate key") {
			return err
		}
	}

	return nil
}

func createConfigFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(homeDir, ".gatorconfig.json"))
	if err != nil {
		return err
	}

	defer file.Close()

	file.WriteString(`{"db_url": "` + dbUrl + `"}`)

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if _, err := s.dbQueries.GetUser(context.Background(), s.config.CurrentUserName); err != nil {
			// User is not logged in
			return fmt.Errorf("user not logged in")
		}
		return handler(s, cmd)
	}
}

func main() {
	workingDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	if _, err := os.Stat(filepath.Join(workingDir, ".gatorconfig.json")); os.IsNotExist(err) {
		err = createConfigFile()
		if err != nil {
			log.Fatal(err)
		}
	}

	config, err := cfg.Read(workingDir)
	if err != nil {
		log.Fatal(err)
	}

	dbState := state{config: &config, dbQueries: dbQueries}

	cmds := commands{functions: make(map[string]func(s *state, cmd command) error)}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args

	if len(args) < 2 {
		log.Fatal("not enough arguments")
	}

	cmd := command{name: args[1], args: args[2:]}

	err = cmds.run(&dbState, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
