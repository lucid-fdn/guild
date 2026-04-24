package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
)

const defaultGitHubIssueQuery = "label:agent:ready state:open"

type githubIssueSearchResponse struct {
	Items []githubIssue `json:"items"`
}

type githubIssue struct {
	Number    int           `json:"number"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	HTMLURL   string        `json:"html_url"`
	State     string        `json:"state"`
	CreatedAt string        `json:"created_at"`
	Labels    []githubLabel `json:"labels"`
	User      *githubUser   `json:"user"`
}

type githubLabel struct {
	Name string `json:"name"`
}

type githubUser struct {
	Login string `json:"login"`
}

func syncGitHubIssueMandates(repoOverride, queryOverride string) error {
	config, err := loadAgentDeskConfig()
	if err != nil {
		return err
	}
	source, err := resolveGitHubTaskSource(config, repoOverride, queryOverride)
	if err != nil {
		return err
	}
	issues, err := fetchGitHubIssues(source.Repo, source.Query)
	if err != nil {
		return err
	}
	for _, issue := range issues {
		mandate := taskpackFromGitHubIssue(config, source.Repo, issue)
		if err := specvalidate.Taskpack(mandate); err != nil {
			return err
		}
		if err := writeAgentDeskJSON(mandatePath(mandate.TaskpackID), mandate); err != nil {
			return err
		}
	}
	return nil
}

func resolveGitHubTaskSource(config spec.WorkspaceConstitution, repoOverride, queryOverride string) (spec.TaskSource, error) {
	source := spec.TaskSource{
		Type:  "github_issues",
		Repo:  repoOverride,
		Query: queryOverride,
	}
	if source.Repo == "" || source.Query == "" {
		for _, candidate := range config.TaskSources {
			if candidate.Type != "github_issues" {
				continue
			}
			if source.Repo == "" {
				source.Repo = candidate.Repo
			}
			if source.Query == "" {
				source.Query = candidate.Query
			}
			break
		}
	}
	if source.Repo == "" {
		source.Repo = os.Getenv("GITHUB_REPOSITORY")
	}
	if source.Query == "" {
		source.Query = defaultGitHubIssueQuery
	}
	if strings.TrimSpace(source.Repo) == "" {
		return spec.TaskSource{}, errors.New("GitHub repo is required; pass --repo, set GITHUB_REPOSITORY, or add a github_issues task_source")
	}
	if !strings.Contains(source.Repo, "/") {
		return spec.TaskSource{}, fmt.Errorf("GitHub repo must be owner/repo, got %q", source.Repo)
	}
	return source, nil
}

func fetchGitHubIssues(repo, query string) ([]githubIssue, error) {
	apiURL := strings.TrimRight(envOr("GITHUB_API_URL", "https://api.github.com"), "/")
	searchQuery := query
	if !strings.Contains(searchQuery, "repo:") {
		searchQuery = "repo:" + repo + " " + searchQuery
	}
	if !strings.Contains(searchQuery, "type:") && !strings.Contains(searchQuery, "is:") {
		searchQuery += " type:issue"
	}
	endpoint := apiURL + "/search/issues?q=" + url.QueryEscape(searchQuery) + "&per_page=20"
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	setGitHubHeaders(request)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("GitHub issue search failed with %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload githubIssueSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Items, nil
}

func taskpackFromGitHubIssue(config spec.WorkspaceConstitution, repo string, issue githubIssue) spec.Taskpack {
	labels := githubLabelNames(issue.Labels)
	createdAt := issue.CreatedAt
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339)
	}
	requester := "github"
	if issue.User != nil && issue.User.Login != "" {
		requester = issue.User.Login
	}
	return spec.Taskpack{
		SchemaVersion: "v1alpha1",
		TaskpackID:    deterministicUUID("github-issue:" + issue.HTMLURL),
		Title:         issue.Title,
		Objective:     githubIssueObjective(issue),
		TaskType:      githubTaskType(labels),
		Priority:      githubPriority(labels),
		RequestedBy: spec.ActorRef{
			ActorID:      deterministicUUID("github-user:" + requester),
			ActorType:    "human",
			DisplayName:  requester,
			Orchestrator: "github",
			Endpoint:     "https://github.com/" + repo,
		},
		RoleHint: githubRoleHint(labels),
		References: []string{
			issue.HTMLURL,
		},
		Labels: uniqueStrings(append([]string{
			"agentdesk",
			"github",
			"issue-" + strconv.Itoa(issue.Number),
		}, sanitizeLabels(labels)...)),
		ContextBudget: spec.ContextBudget{
			MaxInputTokens:  max(config.Defaults.ContextBudgetTokens, 256),
			MaxOutputTokens: 1024,
			ContextStrategy: "artifact_refs_first",
		},
		Permissions: spec.Permissions{
			AllowNetwork:       false,
			AllowShell:         true,
			AllowExternalWrite: false,
			ApprovalMode:       "ask",
			Scopes:             githubScopes(config, labels),
		},
		Acceptance: defaultAcceptance(config.SuccessCriteria),
		CreatedAt:  createdAt,
	}
}

func githubIssueObjective(issue githubIssue) string {
	body := strings.TrimSpace(issue.Body)
	if body == "" {
		return issue.Title
	}
	if len(body) > 2000 {
		body = body[:2000] + "..."
	}
	return issue.Title + "\n\nSource issue:\n" + body
}

func githubLabelNames(labels []githubLabel) []string {
	names := make([]string, 0, len(labels))
	for _, label := range labels {
		if strings.TrimSpace(label.Name) != "" {
			names = append(names, label.Name)
		}
	}
	return names
}

func githubPriority(labels []string) string {
	for _, label := range labels {
		normalized := strings.ToLower(label)
		switch normalized {
		case "priority:p0", "priority:critical", "p0", "critical":
			return "critical"
		case "priority:p1", "priority:high", "p1", "high":
			return "high"
		case "priority:p3", "priority:low", "p3", "low":
			return "low"
		}
	}
	return "medium"
}

func githubTaskType(labels []string) string {
	for _, label := range labels {
		switch strings.ToLower(label) {
		case "type:review", "task:review":
			return "review"
		case "type:research", "task:research":
			return "research"
		case "type:triage", "task:triage":
			return "triage"
		case "type:ops", "type:operations", "task:operations":
			return "operations"
		}
	}
	return "implementation"
}

func githubRoleHint(labels []string) string {
	for _, label := range labels {
		switch strings.ToLower(label) {
		case "role:reviewer", "agent:reviewer":
			return "reviewer"
		case "role:explorer", "agent:explorer":
			return "explorer"
		case "role:skeptic", "agent:skeptic":
			return "skeptic"
		case "role:specialist", "agent:specialist":
			return "specialist"
		}
	}
	return "builder"
}

func githubScopes(config spec.WorkspaceConstitution, labels []string) []string {
	scopes := append([]string{}, config.Scope.Writable...)
	for _, label := range labels {
		lower := strings.ToLower(label)
		if strings.HasPrefix(lower, "scope:") {
			scope := strings.TrimSpace(strings.TrimPrefix(label, "scope:"))
			if scope != "" {
				scopes = append(scopes, scope+"/**")
			}
		}
	}
	return uniqueStrings(scopes)
}

var labelSanitizer = regexp.MustCompile(`[^a-z0-9._/-]+`)

func sanitizeLabels(labels []string) []string {
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		normalized := strings.ToLower(strings.TrimSpace(label))
		normalized = strings.ReplaceAll(normalized, ":", "-")
		normalized = labelSanitizer.ReplaceAllString(normalized, "-")
		normalized = strings.Trim(normalized, "-._/")
		if normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}

func deterministicUUID(seed string) string {
	sum := sha1.Sum([]byte(seed))
	bytes := sum[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	hexed := hex.EncodeToString(bytes)
	return hexed[0:8] + "-" + hexed[8:12] + "-" + hexed[12:16] + "-" + hexed[16:20] + "-" + hexed[20:32]
}

func setGitHubHeaders(request *http.Request) {
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "guild-agentdesk")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
}

func publishGitHubAgentWorkReport(mandate spec.Taskpack, report agentDeskVerifyReport, replayRef string) error {
	body := renderAgentWorkReport(mandate, report, replayRef)
	if summaryPath := os.Getenv("GITHUB_STEP_SUMMARY"); summaryPath != "" {
		if err := appendFile(summaryPath, body+"\n"); err != nil {
			return err
		}
	}
	if os.Getenv("GITHUB_TOKEN") == "" {
		return nil
	}
	prNumber, err := githubPullRequestNumber()
	if err != nil || prNumber == "" {
		return nil
	}
	repository := os.Getenv("GITHUB_REPOSITORY")
	if repository == "" {
		return nil
	}
	return postGitHubIssueComment(repository, prNumber, body)
}

func renderAgentWorkReport(mandate spec.Taskpack, report agentDeskVerifyReport, replayRef string) string {
	status := "failed"
	if report.Ready {
		status = "passed"
	}
	replayStatus := "not attached"
	if strings.TrimSpace(replayRef) != "" {
		replayStatus = "attached"
	}
	approvals := "resolved"
	if report.PendingApprovalCount > 0 {
		approvals = fmt.Sprintf("%d pending", report.PendingApprovalCount)
	}
	lines := []string{
		"### Agent Work Contract: " + status,
		"",
		"- Mandate: " + mandate.Title + " (`" + mandate.TaskpackID + "`)",
		"- Proof: " + strings.Join(report.PresentProofKinds, ", "),
		"- Approvals: " + approvals,
		"- Replay: " + replayStatus,
	}
	if replayRef != "" {
		lines = append(lines, "- Replay ref: `"+replayRef+"`")
	}
	if len(report.OpenIssues) > 0 {
		lines = append(lines, "", "Open issues:")
		for _, issue := range report.OpenIssues {
			lines = append(lines, "- "+issue)
		}
	}
	return strings.Join(lines, "\n")
}

func githubPullRequestNumber() (string, error) {
	if explicit := os.Getenv("GITHUB_PR_NUMBER"); explicit != "" {
		return explicit, nil
	}
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return "", nil
	}
	data, err := os.ReadFile(eventPath)
	if err != nil {
		return "", err
	}
	var payload struct {
		Number      int `json:"number"`
		PullRequest *struct {
			Number int `json:"number"`
		} `json:"pull_request"`
		Issue *struct {
			Number int `json:"number"`
		} `json:"issue"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	switch {
	case payload.PullRequest != nil && payload.PullRequest.Number > 0:
		return strconv.Itoa(payload.PullRequest.Number), nil
	case payload.Issue != nil && payload.Issue.Number > 0:
		return strconv.Itoa(payload.Issue.Number), nil
	case payload.Number > 0:
		return strconv.Itoa(payload.Number), nil
	default:
		return "", nil
	}
}

func postGitHubIssueComment(repo, number, body string) error {
	apiURL := strings.TrimRight(envOr("GITHUB_API_URL", "https://api.github.com"), "/")
	endpoint := apiURL + "/repos/" + repo + "/issues/" + number + "/comments"
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	setGitHubHeaders(request)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("GitHub comment failed with %d: %s", response.StatusCode, strings.TrimSpace(string(data)))
	}
	return nil
}

func appendFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
