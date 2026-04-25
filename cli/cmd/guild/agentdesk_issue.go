package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type agentDeskIssueCreateReport struct {
	SchemaVersion string   `json:"schema_version"`
	Repo          string   `json:"repo"`
	IssueNumber   int      `json:"issue_number"`
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	Labels        []string `json:"labels"`
}

type githubIssueCreateRequest struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels,omitempty"`
}

func runAgentDeskIssue(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk issue command is required")
	}
	switch args[0] {
	case "create":
		return runAgentDeskIssueCreate(args[1:], stdout)
	default:
		return fmt.Errorf("unknown agentdesk issue command %q", args[0])
	}
}

func runAgentDeskIssueCreate(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk issue create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repo", "", "GitHub owner/repo override")
	objective := fs.String("objective", "", "objective; defaults to title")
	scope := fs.String("scope", "", "allowed scope for the agent, for example docs/**")
	acceptance := fs.String("acceptance", "", "acceptance criteria; can be repeated as newline-separated text")
	priority := fs.String("priority", "priority:p2", "priority label, for example priority:p1")
	labels := fs.String("labels", "", "comma-separated extra labels")
	notes := fs.String("notes", "", "additional notes for the agent")
	if err := fs.Parse(reorderInterspersedFlags(args, map[string]bool{
		"repo":       true,
		"objective":  true,
		"scope":      true,
		"acceptance": true,
		"priority":   true,
		"labels":     true,
		"notes":      true,
	})); err != nil {
		return err
	}
	title := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if title == "" {
		return errors.New("issue title is required")
	}
	resolvedRepo, err := resolveGitHubIssueRepo(*repo)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*objective) == "" {
		*objective = title
	}
	if strings.TrimSpace(*scope) == "" {
		*scope = "docs/**"
	}
	if strings.TrimSpace(*acceptance) == "" {
		*acceptance = strings.Join([]string{
			"Tests pass or failure is explained.",
			"Every modified file is listed.",
			"A proof artifact is attached.",
			"A reviewer handoff is created.",
		}, "\n")
	}

	issueLabels := uniqueStrings(append([]string{"agent:ready", strings.TrimSpace(*priority)}, splitCSV(*labels)...))
	issue, err := createGitHubIssue(resolvedRepo, githubIssueCreateRequest{
		Title:  title,
		Body:   renderAgentReadyIssueBody(*objective, *scope, *acceptance, *notes),
		Labels: issueLabels,
	})
	if err != nil {
		return err
	}
	return writeJSON(stdout, agentDeskIssueCreateReport{
		SchemaVersion: "v1alpha1",
		Repo:          resolvedRepo,
		IssueNumber:   issue.Number,
		Title:         issue.Title,
		URL:           issue.HTMLURL,
		Labels:        issueLabels,
	})
}

func resolveGitHubIssueRepo(repoOverride string) (string, error) {
	repo := strings.TrimSpace(repoOverride)
	if repo == "" {
		repo = os.Getenv("GITHUB_REPOSITORY")
	}
	if repo == "" {
		if config, err := loadAgentDeskConfig(); err == nil {
			for _, source := range config.TaskSources {
				if source.Type == "github_issues" && source.Repo != "" {
					repo = source.Repo
					break
				}
			}
		}
	}
	if repo == "" {
		return "", errors.New("GitHub repo is required; pass --repo, set GITHUB_REPOSITORY, or add a github_issues task_source")
	}
	if !strings.Contains(repo, "/") {
		return "", fmt.Errorf("GitHub repo must be owner/repo, got %q", repo)
	}
	return repo, nil
}

func createGitHubIssue(repo string, payload githubIssueCreateRequest) (githubIssue, error) {
	apiURL := strings.TrimRight(envOr("GITHUB_API_URL", "https://api.github.com"), "/")
	body, err := json.Marshal(payload)
	if err != nil {
		return githubIssue{}, err
	}
	request, err := http.NewRequest(http.MethodPost, apiURL+"/repos/"+repo+"/issues", bytes.NewReader(body))
	if err != nil {
		return githubIssue{}, err
	}
	setGitHubHeaders(request)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return githubIssue{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return githubIssue{}, fmt.Errorf("GitHub issue create failed with %d: %s", response.StatusCode, strings.TrimSpace(string(data)))
	}
	var issue githubIssue
	if err := json.NewDecoder(response.Body).Decode(&issue); err != nil {
		return githubIssue{}, err
	}
	return issue, nil
}

func renderAgentReadyIssueBody(objective, scope, acceptance, notes string) string {
	sections := []string{
		"## Objective\n" + strings.TrimSpace(objective),
		"## Allowed scope\n" + strings.TrimSpace(scope),
		"## Acceptance criteria\n" + normalizeIssueList(acceptance),
	}
	if strings.TrimSpace(notes) != "" {
		sections = append(sections, "## Notes for the agent\n"+strings.TrimSpace(notes))
	}
	return strings.Join(sections, "\n\n") + "\n"
}

func normalizeIssueList(value string) string {
	lines := strings.Split(strings.TrimSpace(value), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			out = append(out, line)
		} else {
			out = append(out, "- "+line)
		}
	}
	if len(out) == 0 {
		return "- Complete the mandate."
	}
	return strings.Join(out, "\n")
}
