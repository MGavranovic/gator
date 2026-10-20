package main

import (
	"fmt"

	"github.com/MGavranovic/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handler map[string]func(*state, command) error
}

func main() {
	cfg := config.Read()
	cfg.SetUser("milos")
	_ = config.Read()
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No arguments specified")
	}
	if err := s.cfg.SetUser(cmd.name); err != nil {
		return err
	}
	fmt.Println("The user has been set")
	return nil
}

func (c *commands) run(s *state, cmd command) error {}

func (c *commands) register(name string, f func(*state, command) error) {}
