// Pi coding-agent wrapper — translates StrawPot protocol to pi CLI.
//
// This wrapper is a pure translation layer: it maps StrawPot protocol args
// to "pi" CLI flags.  It does NOT manage processes, sessions, or any
// infrastructure — that is handled by WrapperRuntime in StrawPot core.
//
// Subcommands: setup, build
package main

import (
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
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// setup
// ---------------------------------------------------------------------------

func cmdSetup() {
	piPath, err := exec.LookPath("pi")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: pi CLI not found on PATH.")
		fmt.Fprintln(os.Stderr, "Install it with: npm install -g @mariozechner/pi-coding-agent")
		os.Exit(1)
	}

	cmd := exec.Command(piPath, "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
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

	// Create workspace directory
	if err := os.MkdirAll(ba.AgentWorkspaceDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create workspace dir: %v\n", err)
		os.Exit(1)
	}

	// Write prompt file (AGENTS.md) into workspace.
	// Pi coding-agent automatically discovers AGENTS.md as context/system instructions.
	promptFile := filepath.Join(ba.AgentWorkspaceDir, "AGENTS.md")
	var parts []string
	if ba.RolePrompt != "" {
		parts = append(parts, ba.RolePrompt)
	}
	if ba.MemoryPrompt != "" {
		parts = append(parts, ba.MemoryPrompt)
	}
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
			// Skip if already exists (first-wins deduplication)
			if _, err := os.Lstat(link); err == nil {
				continue
			}
			if err := symlink(src, link); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to link role %s: %v\n", entry.Name(), err)
				os.Exit(1)
			}
		}
	}

	// Symlink project files directories into workspace if provided
	for i, filesDir := range ba.FilesDirs {
		if filesDir == "" {
			continue
		}
		linkName := "files"
		if i > 0 {
			linkName = fmt.Sprintf("files_%d", i)
		}
		filesLink := filepath.Join(ba.AgentWorkspaceDir, linkName)
		if _, err := os.Lstat(filesLink); os.IsNotExist(err) {
			if err := symlink(filesDir, filesLink); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to link files dir: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Build pi command
	cmd := []string{"pi"}

	if ba.Task != "" {
		cmd = append(cmd, "-p", ba.Task)
	}

	if model, ok := config["model"].(string); ok && model != "" {
		cmd = append(cmd, "--model", model)
	}

	// Output JSON
	result := map[string]interface{}{
		"cmd": cmd,
		"cwd": ba.WorkingDir,
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}
