package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	feed, err := FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Println(feed)

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
