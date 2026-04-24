package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const mcpUsage = `Guild MCP

Usage:
  guild mcp serve
`

type mcpRequest struct {
	JSONRPC string         `json:"jsonrpc,omitempty"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string       `json:"jsonrpc"`
	ID      any          `json:"id"`
	Result  any          `json:"result,omitempty"`
	Error   *mcpRPCError `json:"error,omitempty"`
}

type mcpRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type mcpToolResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func runMCP(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, mcpUsage)
		return errors.New("mcp command is required")
	}
	switch args[0] {
	case "serve":
		fs := flag.NewFlagSet("mcp serve", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return serveMCPWithReader(bufio.NewReaderSize(os.Stdin, 64*1024), stdout)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, mcpUsage)
		return nil
	default:
		fmt.Fprint(stderr, mcpUsage)
		return fmt.Errorf("unknown mcp command %q", args[0])
	}
}

func serveMCPWithReader(reader *bufio.Reader, stdout io.Writer) error {
	for {
		payload, err := readMCPFrame(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		var request mcpRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			if writeErr := writeMCPFrame(stdout, mcpResponse{JSONRPC: "2.0", ID: nil, Error: &mcpRPCError{Code: -32700, Message: err.Error()}}); writeErr != nil {
				return writeErr
			}
			continue
		}
		response := handleMCPRequest(request)
		if response == nil {
			continue
		}
		if err := writeMCPFrame(stdout, response); err != nil {
			return err
		}
	}
}

func handleMCPRequest(request mcpRequest) *mcpResponse {
	if strings.HasPrefix(request.Method, "notifications/") {
		return nil
	}
	switch request.Method {
	case "initialize":
		return mcpOK(request.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "guild-agentdesk-mcp",
				"version": "0.1.0-alpha.1",
			},
		})
	case "tools/list":
		return mcpOK(request.ID, map[string]any{"tools": guildMCPTools()})
	case "tools/call":
		name, _ := request.Params["name"].(string)
		args, _ := request.Params["arguments"].(map[string]any)
		if name == "" {
			return mcpFail(request.ID, -32602, "tools/call requires params.name")
		}
		result := callGuildMCPTool(name, args)
		return mcpOK(request.ID, result)
	default:
		return mcpFail(request.ID, -32601, "unknown MCP method: "+request.Method)
	}
}

func callGuildMCPTool(name string, args map[string]any) mcpToolResult {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var err error
	switch name {
	case "guild_get_next_mandate":
		cliArgs := []string{"next"}
		if value := stringArg(args, "source"); value != "" {
			cliArgs = append(cliArgs, "--source", value)
		}
		if value := stringArg(args, "repo"); value != "" {
			cliArgs = append(cliArgs, "--repo", value)
		}
		if value := stringArg(args, "query"); value != "" {
			cliArgs = append(cliArgs, "--query", value)
		}
		err = runAgentDesk(cliArgs, &stdout, &stderr)
	case "guild_claim_mandate":
		cliArgs := []string{"claim", "--id", stringArg(args, "taskpack_id")}
		if value := stringArg(args, "agent"); value != "" {
			cliArgs = append(cliArgs, "--agent", value)
		}
		if value := intArg(args, "ttl_minutes"); value > 0 {
			cliArgs = append(cliArgs, "--ttl-minutes", strconv.Itoa(value))
		}
		if boolArg(args, "force") {
			cliArgs = append(cliArgs, "--force")
		}
		err = runAgentDesk(cliArgs, &stdout, &stderr)
	case "guild_compile_context":
		cliArgs := []string{"context", "compile", "--id", stringArg(args, "taskpack_id"), "--role", stringArgDefault(args, "role", "coder")}
		if value := intArg(args, "budget_tokens"); value > 0 {
			cliArgs = append(cliArgs, "--budget", strconv.Itoa(value))
		}
		err = runAgentDesk(cliArgs, &stdout, &stderr)
	case "guild_check_preflight":
		cliArgs := []string{"preflight", "--id", stringArg(args, "taskpack_id"), "--action", stringArg(args, "action")}
		if value := stringArg(args, "path"); value != "" {
			cliArgs = append(cliArgs, "--path", value)
		}
		if value := stringArg(args, "command"); value != "" {
			cliArgs = append(cliArgs, "--command", value)
		}
		err = runAgentDesk(cliArgs, &stdout, &stderr)
	case "guild_request_approval":
		cliArgs := []string{"approval", "request", "--id", stringArg(args, "taskpack_id"), "--reason", stringArg(args, "reason")}
		if value := intArg(args, "required_approvals"); value > 0 {
			cliArgs = append(cliArgs, "--required", strconv.Itoa(value))
		}
		err = runAgentDesk(cliArgs, &stdout, &stderr)
	case "guild_publish_artifact":
		cliArgs := []string{"proof", "add", "--id", stringArg(args, "taskpack_id"), "--kind", stringArgDefault(args, "kind", "custom"), "--path", stringArg(args, "path")}
		if value := stringArg(args, "summary"); value != "" {
			cliArgs = append(cliArgs, "--summary", value)
		}
		err = runAgentDesk(cliArgs, &stdout, &stderr)
	case "guild_create_handoff":
		err = runAgentDesk([]string{"handoff", "create", "--id", stringArg(args, "taskpack_id"), "--to", stringArg(args, "to"), "--summary", stringArg(args, "summary")}, &stdout, &stderr)
	case "guild_verify_mandate":
		err = runAgentDesk([]string{"verify", "--id", stringArg(args, "taskpack_id")}, &stdout, &stderr)
	case "guild_close_mandate":
		err = runAgentDesk([]string{"close", "--id", stringArg(args, "taskpack_id")}, &stdout, &stderr)
	case "guild_export_replay_bundle":
		err = runAgentDesk([]string{"replay", "export", "--id", stringArg(args, "taskpack_id")}, &stdout, &stderr)
	default:
		return mcpTextResult("unknown Guild MCP tool: "+name, true)
	}
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return mcpTextResult(message, true)
	}
	return mcpTextResult(strings.TrimSpace(stdout.String()), false)
}

func guildMCPTools() []mcpTool {
	return []mcpTool{
		mcpToolDef("guild_get_next_mandate", "Return the next open mandate from local files or GitHub Issues.", map[string]any{
			"source": map[string]any{"type": "string", "enum": []string{"local", "github"}},
			"repo":   map[string]any{"type": "string"},
			"query":  map[string]any{"type": "string"},
		}, []string{}),
		mcpToolDef("guild_claim_mandate", "Create a local lease so agents do not take the same mandate.", map[string]any{
			"taskpack_id": map[string]any{"type": "string", "format": "uuid"},
			"agent":       map[string]any{"type": "string"},
			"ttl_minutes": map[string]any{"type": "integer", "minimum": 1},
			"force":       map[string]any{"type": "boolean"},
		}, []string{"taskpack_id"}),
		mcpToolDef("guild_compile_context", "Compile bounded context for one mandate.", map[string]any{
			"taskpack_id":   map[string]any{"type": "string", "format": "uuid"},
			"role":          map[string]any{"type": "string"},
			"budget_tokens": map[string]any{"type": "integer", "minimum": 256},
		}, []string{"taskpack_id", "role"}),
		mcpToolDef("guild_check_preflight", "Check whether an agent action is allowed, denied, or needs approval.", map[string]any{
			"taskpack_id": map[string]any{"type": "string", "format": "uuid"},
			"action":      map[string]any{"type": "string"},
			"path":        map[string]any{"type": "string"},
			"command":     map[string]any{"type": "string"},
		}, []string{"taskpack_id", "action"}),
		mcpToolDef("guild_request_approval", "Request human approval for a mandate.", map[string]any{
			"taskpack_id":        map[string]any{"type": "string", "format": "uuid"},
			"reason":             map[string]any{"type": "string"},
			"required_approvals": map[string]any{"type": "integer", "minimum": 1},
		}, []string{"taskpack_id", "reason"}),
		mcpToolDef("guild_publish_artifact", "Publish a proof artifact reference.", map[string]any{
			"taskpack_id": map[string]any{"type": "string", "format": "uuid"},
			"kind":        map[string]any{"type": "string"},
			"path":        map[string]any{"type": "string"},
			"summary":     map[string]any{"type": "string"},
		}, []string{"taskpack_id", "path"}),
		mcpToolDef("guild_create_handoff", "Create a handoff proof artifact.", map[string]any{
			"taskpack_id": map[string]any{"type": "string", "format": "uuid"},
			"to":          map[string]any{"type": "string"},
			"summary":     map[string]any{"type": "string"},
		}, []string{"taskpack_id", "to", "summary"}),
		mcpToolDef("guild_verify_mandate", "Verify proof, approvals, and handoff readiness.", map[string]any{
			"taskpack_id": map[string]any{"type": "string", "format": "uuid"},
		}, []string{"taskpack_id"}),
		mcpToolDef("guild_close_mandate", "Close a mandate after proof is ready.", map[string]any{
			"taskpack_id": map[string]any{"type": "string", "format": "uuid"},
		}, []string{"taskpack_id"}),
		mcpToolDef("guild_export_replay_bundle", "Export a replay bundle for one mandate.", map[string]any{
			"taskpack_id": map[string]any{"type": "string", "format": "uuid"},
		}, []string{"taskpack_id"}),
	}
}

func mcpToolDef(name, description string, properties map[string]any, required []string) mcpTool {
	return mcpTool{
		Name:        name,
		Description: description,
		InputSchema: map[string]any{
			"type":                 "object",
			"properties":           properties,
			"required":             required,
			"additionalProperties": false,
		},
	}
}

func readMCPFrame(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, err
			}
			contentLength = parsed
		}
	}
	if contentLength < 0 {
		return nil, errors.New("MCP frame missing Content-Length")
	}
	body := make([]byte, contentLength)
	_, err := io.ReadFull(reader, body)
	return body, err
}

func writeMCPFrame(stdout io.Writer, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err = stdout.Write(body)
	return err
}

func mcpOK(id any, result any) *mcpResponse {
	return &mcpResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func mcpFail(id any, code int, message string) *mcpResponse {
	return &mcpResponse{JSONRPC: "2.0", ID: id, Error: &mcpRPCError{Code: code, Message: message}}
}

func mcpTextResult(text string, isError bool) mcpToolResult {
	return mcpToolResult{
		IsError: isError,
		Content: []mcpContent{
			{Type: "text", Text: text},
		},
	}
}

func stringArg(args map[string]any, name string) string {
	if args == nil {
		return ""
	}
	value, _ := args[name].(string)
	return value
}

func stringArgDefault(args map[string]any, name, fallback string) string {
	if value := stringArg(args, name); value != "" {
		return value
	}
	return fallback
}

func intArg(args map[string]any, name string) int {
	if args == nil {
		return 0
	}
	switch value := args[name].(type) {
	case int:
		return value
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := value.Int64()
		return int(parsed)
	default:
		return 0
	}
}

func boolArg(args map[string]any, name string) bool {
	if args == nil {
		return false
	}
	value, _ := args[name].(bool)
	return value
}
