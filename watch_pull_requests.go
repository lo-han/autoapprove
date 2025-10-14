package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	autoApproveLabel = "autoapprove"
)

type PullRequestWatch struct{}

func NewPullRequestWatch() *PullRequestWatch {
	return &PullRequestWatch{}
}

func (p *PullRequestWatch) Watch(client *http.Client) (pullRequests []*GithubPullRequest, err error) {
	responses, err := p.fetchPullRequests(client)
	if err != nil {
		return nil, fmt.Errorf("error fetching pull requests: %w", err)
	}

	for _, resp := range responses {
		hasTag, err := p.checkIfHasAutoapproveTag(client, resp.Subject.Url)
		if err != nil {
			return nil, fmt.Errorf("error checking for autoapprove tag: %w", err)
		}

		approvedByMe, err := p.HasBeenApprovedByMe(client, resp.Subject.Url)
		if err != nil {
			return nil, fmt.Errorf("error checking if has been approved by me: %w", err)
		}

		if hasTag || !approvedByMe {
			owner, repo, prNumber, err := p.splitLink(resp.Subject.Url)
			if err != nil {
				return nil, fmt.Errorf("error splitting link: %w", err)
			}
			pullRequests = append(pullRequests, NewGithubPullRequest(owner, repo, prNumber))
		}
	}

	return pullRequests, nil
}

func (p *PullRequestWatch) HasBeenApprovedByMe(client *http.Client, url string) (approved bool, err error) {
	url += "/reviews"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response []hasBeenApprovedResponse

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading hasBeenApprovedResponse body: %w", err)
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return false, fmt.Errorf("error unmarshalling hasBeenApprovedResponse: %w", err)
	}

	for _, review := range response {
		if review.User.Login == os.Getenv("USER") && review.State == "APPROVED" {
			return true, nil
		}
	}

	return false, nil
}

func (p *PullRequestWatch) checkIfHasAutoapproveTag(client *http.Client, url string) (hasTag bool, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response readPullRequestResponse

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading readPullRequestResponse body: %w", err)
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return false, fmt.Errorf("error unmarshalling readPullRequestResponse: %w", err)
	}

	for _, label := range response.Labels {
		if label.Name == autoApproveLabel {
			return true, nil
		}
	}
	return false, nil
}

func (p *PullRequestWatch) fetchPullRequests(client *http.Client) (responses []watchResponse, err error) {
	url := "https://api.github.com/notifications"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var watchResponse []watchResponse

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading watchResponse body: %w", err)
	}

	if err := json.Unmarshal(body, &watchResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling watchResponse: %w", err)
	}
	return watchResponse, nil
}

func (p *PullRequestWatch) splitLink(link string) (owner, repository string, prNumber int, err error) {
	splitted := strings.Split(link, "/")
	if len(splitted) < 7 {
		err = fmt.Errorf("invalid pull request format")
		return
	}

	owner = splitted[4]
	repository = splitted[5]

	prNumber, err = strconv.Atoi(splitted[7])
	if err != nil {
		err = fmt.Errorf("invalid pull request number: %w", err)
		return
	}

	return
}

type subject struct {
	Url string `json:"url"`
}

type label struct {
	Name string `json:"name"`
}

type watchResponse struct {
	Id      string  `json:"id"`
	Subject subject `json:"subject"`
}

type readPullRequestResponse struct {
	Id     int64   `json:"id"`
	Labels []label `json:"labels"`
}

type user struct {
	Login string `json:"login"`
}

type hasBeenApprovedResponse struct {
	User  user   `json:"user"`
	State string `json:"state"`
}
