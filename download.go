package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vartanbeno/go-reddit/v2/reddit"
	"gopkg.in/yaml.v3"
)

type App struct {
	Creds Creds `yaml:"app"`
}

type Creds struct {
	ID       string `yaml:"id"`
	Secret   string `yaml:"secret"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func IsNonImageLink(url string) bool {
	prefix := "https://www.reddit.com/r/forsen/comments"

	if len(url) < len(prefix) {
		return false
	}

	return url[:len(prefix)] == prefix
}

func downloadRawRedditData(redditCredsFilePath, outputFolder string) error {
	redditCredsFile, err := os.ReadFile(redditCredsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read reddit creds file: %w", err)
	}

	var app *App
	if err := yaml.Unmarshal(redditCredsFile, &app); err != nil {
		return fmt.Errorf("failed to unmarshal reddit creds file: %w", err)
	}
	creds := app.Creds

	credentials := reddit.Credentials{ID: creds.ID, Secret: creds.Secret, Username: creds.Username, Password: creds.Password}
	client, err := reddit.NewClient(credentials, reddit.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))
	if err != nil {
		return fmt.Errorf("failed to create reddit client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	posts, _, err := client.Subreddit.NewPosts(ctx, "forsen", &reddit.ListOptions{
		Limit: 300,
	})
	if err != nil {
		return fmt.Errorf("failed to get top forsen posts: %w", err)
	}

	for _, post := range posts {
		if !IsNonImageLink(post.URL) {
			continue
		}
		fmt.Println(post.ID, post.FullID, post.Title, post.Author, post.Body, post.URL)
	}

	return nil
}
