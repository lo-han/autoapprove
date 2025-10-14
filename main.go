package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("please provide a resource kind and name")
		return
	}
	mode := os.Args[1]

	client := &http.Client{Timeout: 30 * time.Second}

	switch mode {
	case "watch":
		approveFromNotificationsWatch(client)
		return

	case "cli":
		approveFromCLI(client)
		return

	default:
		log.Fatalf("unknown mode: %s", mode)
		return
	}
}

func approveFromNotificationsWatch(client *http.Client) {
	watcher := NewPullRequestWatch()
	fmt.Println("Watching pull requests for approval...")

	for {
		pullRequests, err := watcher.Watch(client)

		if err != nil {
			log.Fatalf("Error watching pull requests: %v", err)
			return
		}

		if len(pullRequests) > 0 {
			approve(client, pullRequests)
		}
		time.Sleep(1 * time.Minute)
	}

}

func approveFromCLI(client *http.Client) {
	pullRequests, err := getPullRequestFromCLI()
	if err != nil {
		log.Fatalf("Error getting pull requests from CLI: %v", err)
		return
	}

	approve(client, pullRequests)
}

func approve(client *http.Client, pullRequests []*GithubPullRequest) {
	hasError := false

	for _, pr := range pullRequests {
		if err := pr.Approve(client); err != nil {
			hasError = true
			fmt.Printf("Error approving PR %s: %v\n", pr.ID(), err)
		}
	}

	if !hasError {
		fmt.Println("PRs approved successfully")
	}
}

func getPullRequestFromCLI() (pullRequests []*GithubPullRequest, err error) {
	for _, pullRequest := range os.Args[2:] {
		splitted := strings.Split(pullRequest, "/")
		if len(splitted) < 7 {
			return nil, fmt.Errorf("invalid pull request format")
		}

		repository := splitted[4]
		owner := splitted[3]

		prNumber, err := strconv.Atoi(splitted[6])
		if err != nil {
			return nil, fmt.Errorf("invalid pull request number: %w", err)
		}

		pullRequests = append(pullRequests, NewGithubPullRequest(owner, repository, prNumber))
	}

	return pullRequests, nil
}
