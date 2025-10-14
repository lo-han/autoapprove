package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GithubPullRequest struct {
	headers    map[string]string
	owner      string
	repo       string
	pullNumber int
	event      string
}

func NewGithubPullRequest(owner, repo string, pullNumber int) *GithubPullRequest {
	return &GithubPullRequest{
		headers: map[string]string{
			"Accept": "application/vnd.github+json",
		},
		event:      "APPROVE",
		owner:      owner,
		repo:       repo,
		pullNumber: pullNumber,
	}
}

func (g *GithubPullRequest) ID() string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", g.owner, g.repo, g.pullNumber)
}

func (g *GithubPullRequest) Approve(client *http.Client) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews", g.owner, g.repo, g.pullNumber)

	requestBody := map[string]string{
		"event": g.event,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	for key, value := range g.headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
