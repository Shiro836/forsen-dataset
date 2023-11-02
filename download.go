package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
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

var regex = regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

func LinkExists(text string) bool {
	return regex.MatchString(text)
}

type cache struct {
	OldestID string `yaml:"oldest_id"`
}

func getOldestId() string {
	data, err := os.ReadFile("cache.yaml")
	if err != nil {
		return ""
	}

	var cache *cache
	if err := yaml.Unmarshal(data, &cache); err != nil {
		return ""
	}

	return cache.OldestID
}

func updateOldestId(oldestID string) error {
	data, err := os.ReadFile("cache.yaml")
	if err != nil {
		return fmt.Errorf("failed to open cache.yaml")
	}

	var cache *cache
	if err := yaml.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("failed to unmarshal cache.yaml")
	}

	cache.OldestID = oldestID
	data, err = yaml.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache")
	}

	return os.WriteFile("cache.yaml", data, 0o644)
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

	oldestID := getOldestId()
	oldestTime := time.Now()

	for i := 0; i < 20; i++ {
		// posts, _, err := client.Subreddit.SearchPosts(ctx, "", "forsen", &reddit.ListPostSearchOptions{
		// 	ListPostOptions: reddit.ListPostOptions{
		// 		ListOptions: reddit.ListOptions{
		// 			Limit:  100,
		// 			Before: oldestID,
		// 		},
		// 		Time: "all",
		// 	},
		// 	Sort: "new",
		// })
		posts, _, err := client.Subreddit.NewPosts(ctx, "forsen", &reddit.ListOptions{
			Limit:  100,
			Before: oldestID,
		})
		if err != nil {
			return fmt.Errorf("failed to get top forsen posts: %w", err)
		}

		for _, post := range posts {
			if post.Created.Time.Before(oldestTime) {
				oldestID = post.FullID
				oldestTime = post.Created.Time
				if err := updateOldestId(post.FullID); err != nil {
					return fmt.Errorf("failed to update oldest id: %w", err)
				}
			}
			if !IsNonImageLink(post.URL) {
				continue
			}
			if LinkExists(post.Body) {
				continue
			}
			fmt.Println(post.ID, post.FullID, post.Title, post.Author, post.Body, post.URL)
		}
	}

	return nil
}
