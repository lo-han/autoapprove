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
	pullRequests, err := getPullRequestFromCLI()

	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
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
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("please provide a resource kind and name")
	}

	for _, pullRequest := range os.Args[1:] {
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
