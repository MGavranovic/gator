package main

import (
	"fmt"
	"os"

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
	conf := config.Read()
	s := state{cfg: &conf}

	cmds := commands{
		handler: map[string]func(*state, command) error{},
	}
	cmds.register("login", handleLogin)

	if len(os.Args) < 2 {
		fmt.Println("Not enough args")
		os.Exit(1)
	}
	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	err := cmds.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_ = config.Read()
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Username is required")
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Println("The user has been set")
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	val, ok := c.handler[cmd.name]
	if ok {
		return val(s, cmd)
	}
	return fmt.Errorf("Command doesn't exist")
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handler[name] = f
}
