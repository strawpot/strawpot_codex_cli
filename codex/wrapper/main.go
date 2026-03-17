// Codex CLI wrapper — translates StrawPot protocol to Codex CLI.
//
// This wrapper is a pure translation layer: it maps StrawPot protocol args
// to "codex" CLI flags.  It does NOT manage processes, sessions, or any
// infrastructure — that is handled by WrapperRuntime in StrawPot core.
//
// Subcommands: setup, build, filter
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: wrapper <setup|build> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		cmdSetup()
	case "build":
		cmdBuild(os.Args[2:])
	case "filter":
		cmdFilter()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// setup
// ---------------------------------------------------------------------------

func cmdSetup() {
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: codex CLI not found on PATH.")
		fmt.Fprintln(os.Stderr, "Install it with: npm install -g @openai/codex")
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Starting Codex CLI login...")
	fmt.Fprintln(os.Stderr, "If a browser window does not open, copy the URL from the output below.")
	fmt.Fprintln(os.Stderr)

	cmd := exec.Command(codexPath, "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Inherit full environment so DISPLAY/WAYLAND_DISPLAY are available
	// for browser opening on Linux.
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// filter — reads Codex JSONL from stdin, emits only agent_message text
// ---------------------------------------------------------------------------

func cmdFilter() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB lines
	for scanner.Scan() {
		var event struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if event.Type == "item.completed" && event.Item.Type == "agent_message" {
			fmt.Println(event.Item.Text)
		}
	}
}

// ---------------------------------------------------------------------------
// build
// ---------------------------------------------------------------------------

type buildArgs struct {
	AgentID           string
	WorkingDir        string
	AgentWorkspaceDir string
	RolePrompt        string
	MemoryPrompt      string
	Task              string
	Config            string
	SkillsDirs        []string
	RolesDirs         []string
	FilesDirs         []string
}

func parseBuildArgs(args []string) buildArgs {
	var ba buildArgs
	ba.Config = "{}"

	for i := 0; i < len(args); i++ {
		if i+1 >= len(args) {
			break
		}
		switch args[i] {
		case "--agent-id":
			i++
			ba.AgentID = args[i]
		case "--working-dir":
			i++
			ba.WorkingDir = args[i]
		case "--agent-workspace-dir":
			i++
			ba.AgentWorkspaceDir = args[i]
		case "--role-prompt":
			i++
			ba.RolePrompt = args[i]
		case "--memory-prompt":
			i++
			ba.MemoryPrompt = args[i]
		case "--task":
			i++
			ba.Task = args[i]
		case "--config":
			i++
			ba.Config = args[i]
		case "--skills-dir":
			i++
			ba.SkillsDirs = append(ba.SkillsDirs, args[i])
		case "--roles-dir":
			i++
			ba.RolesDirs = append(ba.RolesDirs, args[i])
		case "--files-dir":
			i++
			ba.FilesDirs = append(ba.FilesDirs, args[i])
		}
	}
	return ba
}

// symlink creates a symlink from dst pointing to src.
func symlink(src, dst string) error {
	return os.Symlink(src, dst)
}

func cmdBuild(args []string) {
	ba := parseBuildArgs(args)

	// Parse config JSON
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ba.Config), &config); err != nil {
		config = map[string]interface{}{}
	}

	// Validate required args
	if ba.AgentWorkspaceDir == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent-workspace-dir is required")
		os.Exit(1)
	}

	// Use agent workspace dir directly as the --add-dir for Codex.
	if err := os.MkdirAll(ba.AgentWorkspaceDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create workspace dir: %v\n", err)
		os.Exit(1)
	}

	// Write prompt file into workspace
	promptFile := filepath.Join(ba.AgentWorkspaceDir, "prompt.md")
	var parts []string
	if ba.RolePrompt != "" {
		parts = append(parts, ba.RolePrompt)
	}
	if ba.MemoryPrompt != "" {
		parts = append(parts, ba.MemoryPrompt)
	}
	hasPrompt := len(parts) > 0
	if err := os.WriteFile(promptFile, []byte(strings.Join(parts, "\n\n")), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write prompt file: %v\n", err)
		os.Exit(1)
	}

	// Symlink each subdirectory from each skills-dir into skills/<name>/
	for _, skillsDir := range ba.SkillsDirs {
		if skillsDir == "" {
			continue
		}
		entries, err := os.ReadDir(skillsDir)
		if err == nil && len(entries) > 0 {
			skillsTarget := filepath.Join(ba.AgentWorkspaceDir, "skills")
			if err := os.MkdirAll(skillsTarget, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create skills dir: %v\n", err)
				os.Exit(1)
			}
			for _, entry := range entries {
				if !entry.IsDir() && entry.Type()&fs.ModeSymlink == 0 {
					continue
				}
				src := filepath.Join(skillsDir, entry.Name())
				link := filepath.Join(skillsTarget, entry.Name())
				if _, err := os.Lstat(link); err == nil {
					continue
				}
				if err := symlink(src, link); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to link skill %s: %v\n", entry.Name(), err)
					os.Exit(1)
				}
			}
		}
	}

	// Symlink each subdirectory from each roles-dir into roles/<name>/
	for _, rolesDir := range ba.RolesDirs {
		if rolesDir == "" {
			continue
		}
		entries, err := os.ReadDir(rolesDir)
		if err != nil || len(entries) == 0 {
			continue
		}
		rolesTarget := filepath.Join(ba.AgentWorkspaceDir, "roles")
		if err := os.MkdirAll(rolesTarget, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create roles dir: %v\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if !entry.IsDir() && entry.Type()&fs.ModeSymlink == 0 {
				continue
			}
			src := filepath.Join(rolesDir, entry.Name())
			link := filepath.Join(rolesTarget, entry.Name())
			// Skip if already exists (e.g. pre-placed by another roles-dir)
			if _, err := os.Lstat(link); err == nil {
				continue
			}
			if err := symlink(src, link); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to link role %s: %v\n", entry.Name(), err)
				os.Exit(1)
			}
		}
	}

	// Build codex command
	cmd := []string{"codex", "exec"}

	// JSON output for machine parsing (suppresses prompt echo)
	cmd = append(cmd, "--json")

	if ba.Task != "" {
		cmd = append(cmd, ba.Task)
	}

	// System prompt via model_instructions_file config override
	if hasPrompt {
		cmd = append(cmd, "-c", fmt.Sprintf("model_instructions_file=%q", promptFile))
	}

	if model, ok := config["model"].(string); ok && model != "" {
		cmd = append(cmd, "-m", model)
	}

	if sm := os.Getenv("SANDBOX_MODE"); sm != "" {
		cmd = append(cmd, "--sandbox", sm)
	}

	// Default: enable --dangerously-bypass-approvals-and-sandbox unless explicitly disabled.
	if skip, ok := config["dangerously_skip_permissions"].(bool); !ok || skip {
		cmd = append(cmd, "--dangerously-bypass-approvals-and-sandbox")
	}

	// Working directory via -C flag
	if ba.WorkingDir != "" {
		cmd = append(cmd, "-C", ba.WorkingDir)
	}

	// Single --add-dir pointing to the agent workspace.
	cmd = append(cmd, "--add-dir", ba.AgentWorkspaceDir)

	// Add project files directories if provided
	for _, filesDir := range ba.FilesDirs {
		if filesDir != "" {
			cmd = append(cmd, "--add-dir", filesDir)
		}
	}

	// Resolve wrapper binary path for the filter pipe.
	wrapperBin, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve wrapper path: %v\n", err)
		os.Exit(1)
	}

	// Build shell command that pipes codex JSONL through the filter.
	shellCmd := shellJoin(cmd) + " | " + shellEscape(wrapperBin) + " filter"

	// Output JSON
	result := map[string]interface{}{
		"cmd": []string{"sh", "-c", shellCmd},
		"cwd": ba.WorkingDir,
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

// shellEscape wraps a string in single quotes for sh.
func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// shellJoin quotes and joins args for sh -c.
func shellJoin(args []string) string {
	escaped := make([]string, len(args))
	for i, a := range args {
		escaped[i] = shellEscape(a)
	}
	return strings.Join(escaped, " ")
}
