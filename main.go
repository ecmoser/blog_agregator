package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	cfg "github.com/ecmoser/blog_aggregator/internal/config"
)

type state struct {
	config *cfg.Config
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

	err = s.config.SetUser(cmd.args[0], wd)
	if err != nil {
		return err
	}

	fmt.Println("The user " + cmd.args[0] + " has been set")

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

	file.WriteString(`{"db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"}`)

	return nil
}

func main() {
	workingDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

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

	cfgState := state{config: &config}

	cmds := commands{functions: make(map[string]func(s *state, cmd command) error)}

	cmds.register("login", handlerLogin)

	args := os.Args

	if len(args) < 2 {
		log.Fatal("not enough arguments")
	}

	cmd := command{name: args[1], args: args[2:]}

	err = cmds.run(&cfgState, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
