package main

import (
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	registry map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registry[name] = f

}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.registry[cmd.name]
	if !ok {
		return fmt.Errorf("command not in registry")
	}

	err := f(s, cmd)
	if err != nil {
		return err
	}

	return nil
}
