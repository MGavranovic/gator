package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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
	cmds.register("users", handleGetUsers)
	cmds.register("agg", handeFetchRSSFeed)

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

func handleGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting all users from users table: %s", err)
	}
	for _, u := range users {
		if u.Name == s.cfg.Current_user_name {
			fmt.Printf("* %s (current)", u.Name)
		}
		fmt.Printf("* %s", u.Name)
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// make req
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Error: %w", err)
	}
	req.Header.Set("User-Agent", "gator")

	// send req to client
	client := http.Client{}
	res, errRes := client.Do(req)
	if errRes != nil {
		return nil, fmt.Errorf("Error %w:", errRes)
	}

	// read resp
	data, errData := io.ReadAll(res.Body)
	if errData != nil {
		return nil, fmt.Errorf("Error %w:", errData)
	}
	defer res.Body.Close()

	// unmarshal resp
	var feed RSSFeed
	errUnMarsh := xml.Unmarshal(data, &feed)
	if errUnMarsh != nil {
		return nil, errUnMarsh
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func handeFetchRSSFeed(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}
	fmt.Println(feed)
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
