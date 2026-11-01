package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"os"
	"time"

	"github.com/MGavranovic/gator/internal/config"
	"github.com/MGavranovic/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
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

	db, errDB := sql.Open("postgres", conf.DB_URL)
	if errDB != nil {
		fmt.Println(errDB)
	}
	dbQueries := database.New(db)
	s.db = dbQueries

	cmds := commands{
		handler: map[string]func(*state, command) error{},
	}
	cmds.register("login", handleLogin)
	cmds.register("register", handleRegister)
	cmds.register("reset", handleReset)

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

	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println(err)
	}
	if cmd.args[0] != user.Name {
		os.Exit(1)
	}

	fmt.Println("The user has been set")
	return nil
}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Name is required")
	}
	// check if user exists
	user2, err2 := s.db.GetUser(context.Background(), cmd.args[0])
	if err2 != nil {
		fmt.Println(err2)
	}
	if cmd.args[0] == user2.Name {
		os.Exit(1)
	}

	// create user
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]})
	if err != nil {
		return fmt.Errorf("Error creating a user: %s", err)
	}

	s.cfg.SetUser(user.Name)

	fmt.Println("User was created", user)
	return nil
}

func handleReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error clearing users table: %s", err)
	}
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
