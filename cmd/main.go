package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Event struct {
	Type      string    `json:"type"`
	Repo      Repo      `json:"repo"`
	Payload   Payload   `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type Repo struct {
	Name string `json:"name"`
}

type Payload struct {
	Action      string      `json:"action"`
	PushID      int64       `json:"push_id"`
	Size        int         `json:"size"`
	Commits     []Commit    `json:"commits"`
	Issue       Issue       `json:"issue"`
	PullRequest PullRequest `json:"pull_request"`
}

type Commit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
}

type Issue struct {
	Title string `json:"title"`
}

type PullRequest struct {
	Title string `json:"title"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: activity <username>")
		os.Exit(1)
	}

	username := os.Args[1]
	events, err := fetchGitHubActivity(username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	displayActivity(events)
}

func fetchGitHubActivity(username string) ([]Event, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "GitHub-Activity-CLI")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user '%s' not found", username)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return events, nil
}

func displayActivity(events []Event) {
	if len(events) == 0 {
		fmt.Println("No recent activity found.")
		return
	}

	for _, event := range events {
		switch event.Type {
		case "PushEvent":
			fmt.Printf("- Pushed %d commits to %s\n", event.Payload.Size, event.Repo.Name)

		case "CreateEvent":
			fmt.Printf("- Created repository %s\n", event.Repo.Name)

		case "IssuesEvent":
			fmt.Printf("- %s issue in %s: %s\n",
				capitalize(event.Payload.Action),
				event.Repo.Name,
				event.Payload.Issue.Title)

		case "PullRequestEvent":
			fmt.Printf("- %s pull request in %s: %s\n",
				capitalize(event.Payload.Action),
				event.Repo.Name,
				event.Payload.PullRequest.Title)

		case "WatchEvent":
			fmt.Printf("- Starred %s\n", event.Repo.Name)

		default:
			fmt.Printf("- %s event in %s\n", event.Type, event.Repo.Name)
		}
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
