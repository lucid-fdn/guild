package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type agentDeskDoctorReport struct {
	SchemaVersion string                 `json:"schema_version"`
	Ready         bool                   `json:"ready"`
	Checks        []agentDeskDoctorCheck `json:"checks"`
}

type agentDeskDoctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func runAgentDeskDoctor(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "optional mandate/taskpack UUID for proof readiness checks")
	repo := fs.String("repo", "", "optional GitHub owner/repo override for token and label checks")
	if err := fs.Parse(args); err != nil {
		return err
	}
	report := buildAgentDeskDoctorReport(*id, *repo)
	if err := writeJSON(stdout, report); err != nil {
		return err
	}
	if !report.Ready {
		return errors.New("agentdesk doctor found failing checks")
	}
	return nil
}

func buildAgentDeskDoctorReport(mandateID, repoOverride string) agentDeskDoctorReport {
	report := agentDeskDoctorReport{
		SchemaVersion: "v1alpha1",
		Ready:         true,
	}
	add := func(name, status, message string) {
		report.Checks = append(report.Checks, agentDeskDoctorCheck{Name: name, Status: status, Message: message})
		if status == "fail" {
			report.Ready = false
		}
	}

	root, rootErr := findAgentDeskRoot()
	if rootErr != nil {
		add("agentdesk_config", "fail", rootErr.Error())
		return report
	}
	config, configErr := loadAgentDeskConfig()
	if configErr != nil {
		add("agentdesk_config", "fail", configErr.Error())
		return report
	}
	add("agentdesk_config", "pass", "agentdesk.yaml is present and valid")

	requiredDirs := map[string]bool{
		".agentdesk/mandates": true,
		".agentdesk/proof":    true,
		".agentdesk/replay":   true,
		".agentdesk/handoffs": true,
	}
	for _, dir := range []string{".agentdesk/mandates", ".agentdesk/proof", ".agentdesk/replay", ".agentdesk/handoffs", ".agentdesk/approvals", ".agentdesk/claims", ".agentdesk/closed"} {
		if info, err := os.Stat(filepath.Join(root, dir)); err != nil {
			if requiredDirs[dir] {
				add("agentdesk_directory:"+dir, "fail", dir+" is missing; run `guild agentdesk init`")
			} else {
				add("agentdesk_directory:"+dir, "warn", dir+" is missing; it will be created when needed")
			}
		} else if !info.IsDir() {
			add("agentdesk_directory:"+dir, "fail", dir+" exists but is not a directory")
		} else {
			add("agentdesk_directory:"+dir, "pass", dir+" exists")
		}
	}

	repo := strings.TrimSpace(repoOverride)
	if repo == "" {
		for _, source := range config.TaskSources {
			if source.Type == "github_issues" && source.Repo != "" {
				repo = source.Repo
				break
			}
		}
	}
	if repo == "" {
		repo = os.Getenv("GITHUB_REPOSITORY")
	}
	token := os.Getenv("GITHUB_TOKEN")
	if repo == "" {
		add("github_token", "warn", "no GitHub repo configured; pass --repo or add a github_issues task_source")
		add("github_labels", "warn", "skipped because no GitHub repo is configured")
	} else if token == "" {
		add("github_token", "warn", "GITHUB_TOKEN is not set; GitHub issue intake may be rate-limited or unavailable for "+repo)
		add("github_labels", "warn", "skipped because GITHUB_TOKEN is missing")
	} else {
		add("github_token", "pass", "GITHUB_TOKEN is set for "+repo)
		labels, err := fetchGitHubLabelNames(repo)
		if err != nil {
			add("github_labels", "fail", err.Error())
		} else if !stringSliceContains(labels, "agent:ready") {
			add("github_labels", "fail", "label agent:ready is missing in "+repo)
		} else {
			add("github_labels", "pass", "label agent:ready exists in "+repo)
		}
	}

	if executable, err := os.Executable(); err != nil {
		add("mcp_command", "warn", "could not resolve current executable: "+err.Error())
	} else {
		add("mcp_command", "pass", fmt.Sprintf("single-binary MCP server is available: %s mcp serve", filepath.Base(executable)))
	}

	if strings.TrimSpace(mandateID) == "" {
		add("proof_readiness", "warn", "skipped; pass --id <mandate-id> to verify proof readiness")
		return report
	}
	mandate, err := loadMandate(mandateID)
	if err != nil {
		add("proof_readiness", "fail", err.Error())
		return report
	}
	verify, err := verifyMandate(mandate)
	if err != nil {
		add("proof_readiness", "fail", err.Error())
		return report
	}
	if !verify.Ready {
		add("proof_readiness", "fail", strings.Join(verify.OpenIssues, "; "))
		return report
	}
	add("proof_readiness", "pass", fmt.Sprintf("mandate %s is ready with %d proof artifacts", mandate.TaskpackID, verify.ProofCount))
	return report
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
