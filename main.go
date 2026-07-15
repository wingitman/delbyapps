package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

var defaultWorkDir string

type app struct {
	Name        string
	Description string
	RepoDir     string
	ConfigFile  string
}

var apps = []app{
	{Name: "atob", Description: "universal text converter", RepoDir: "atob", ConfigFile: "atob.toml"},
	{Name: "crosspile", Description: "AI agent history browser", RepoDir: "crosspile", ConfigFile: "crossfile.toml"},
	{Name: "devtools", Description: "remote environment command TUI", RepoDir: "DevToolsCLI", ConfigFile: "devtools.toml"},
	{Name: "gcss", Description: "terminal HTML and CSS builder", RepoDir: "gcss", ConfigFile: "gcss.toml"},
	{Name: "lambit", Description: "local AWS Lambda tester", RepoDir: "lambit", ConfigFile: "lambit.toml"},
	{Name: "listicles", Description: "terminal file explorer", RepoDir: "listicles", ConfigFile: "listicles.toml"},
	{Name: "scry", Description: "AI task harness runner", RepoDir: "scry", ConfigFile: "scry.toml"},
	{Name: "sqwee", Description: "terminal database client", RepoDir: "sqwee", ConfigFile: "sqwee.toml"},
	{Name: "teapi", Description: "HTTP client TUI", RepoDir: "teapi", ConfigFile: "teapi.toml"},
	{Name: "ticky", Description: "focus timer and task scheduler", RepoDir: "ticky", ConfigFile: "ticky.toml"},
	{Name: "lup", Description: "code documentation lookup", RepoDir: "lup", ConfigFile: "lup.toml"},
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "delbyapps:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "list", "ls":
		return listApps()
	case "config":
		if len(args) != 2 {
			return errors.New("usage: delbyapps config <app>")
		}
		return openConfig(args[1])
	case "install":
		if len(args) != 2 {
			return errors.New("usage: delbyapps install <app>")
		}
		return installApp(args[1])
	case "help":
		if len(args) == 1 {
			printUsage()
			return nil
		}
		if len(args) != 2 {
			return errors.New("usage: delbyapps help <app>")
		}
		return describeApp(args[1])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println("delbyapps manages delbysoft CLI applications")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  delbyapps list")
	fmt.Println("  delbyapps config <app>")
	fmt.Println("  delbyapps install <app>")
	fmt.Println("  delbyapps help <app>")
}

func listApps() error {
	configDir, err := delbysoftConfigDir()
	if err != nil {
		return err
	}
	workDir, _ := findWorkDir()

	installed := 0
	for _, a := range apps {
		if !fileExists(filepath.Join(configDir, a.ConfigFile)) {
			continue
		}

		installed++
		status := updateStatus(a, workDir, configDir)
		fmt.Printf("%s - %s | %s\n", a.Name, a.Description, status)
	}

	if installed == 0 {
		fmt.Printf("No delbysoft app configs found in %s\n", configDir)
	}

	return nil
}

func openConfig(name string) error {
	a, err := lookupApp(name)
	if err != nil {
		return err
	}
	configDir, err := delbysoftConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(configDir, a.ConfigFile)
	if !fileExists(path) {
		return fmt.Errorf("config not found for %s: %s", a.Name, path)
	}

	return openInEditor(path)
}

func installApp(name string) error {
	a, err := lookupApp(name)
	if err != nil {
		return err
	}
	workDir, err := findWorkDir()
	if err != nil {
		return err
	}
	repoDir := filepath.Join(workDir, a.RepoDir)
	if !dirExists(repoDir) {
		return fmt.Errorf("source repo not found for %s: %s", a.Name, repoDir)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		installer := filepath.Join(repoDir, "install.ps1")
		if !fileExists(installer) {
			return fmt.Errorf("installer not found: %s", installer)
		}
		cmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", installer)
	} else if fileExists(filepath.Join(repoDir, "Makefile")) {
		cmd = exec.Command("make", "install")
	} else if fileExists(filepath.Join(repoDir, "install.sh")) {
		cmd = exec.Command("sh", "install.sh")
	} else {
		return fmt.Errorf("no supported installer found in %s", repoDir)
	}

	cmd.Dir = repoDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func describeApp(name string) error {
	a, err := lookupApp(name)
	if err != nil {
		return err
	}
	configDir, err := delbysoftConfigDir()
	if err != nil {
		return err
	}
	workDir, workErr := findWorkDir()
	configPath := filepath.Join(configDir, a.ConfigFile)

	fmt.Printf("%s - %s\n", a.Name, a.Description)
	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Installed: %s\n", yesNo(fileExists(configPath)))
	if workErr == nil {
		repoPath := filepath.Join(workDir, a.RepoDir)
		fmt.Printf("Source: %s\n", repoPath)
		fmt.Printf("Update: %s\n", updateStatus(a, workDir, configDir))
		if runtime.GOOS == "windows" {
			fmt.Println("Install: run install.ps1 from the source repo")
		} else if fileExists(filepath.Join(repoPath, "Makefile")) {
			fmt.Println("Install: run make install from the source repo")
		} else {
			fmt.Println("Install: run install.sh from the source repo")
		}
	} else {
		fmt.Printf("Source: %v\n", workErr)
		fmt.Println("Update: update unknown")
	}

	return nil
}

func lookupApp(name string) (app, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, a := range apps {
		if a.Name == name {
			return a, nil
		}
	}

	names := make([]string, 0, len(apps))
	for _, a := range apps {
		names = append(names, a.Name)
	}
	sort.Strings(names)
	return app{}, fmt.Errorf("unknown app %q; known apps: %s", name, strings.Join(names, ", "))
}

func delbysoftConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not locate user config directory: %w", err)
	}
	return filepath.Join(dir, "delbysoft"), nil
}

func updateStatus(a app, workDir, configDir string) string {
	repoDir := repoDirFromMeta(configDir, a, workDir)
	if repoDir == "" || !dirExists(repoDir) {
		return "update unknown"
	}

	local, err := gitOutput(repoDir, "rev-parse", "HEAD")
	if err != nil {
		return "update unknown"
	}
	branch, err := gitOutput(repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "update unknown"
	}

	remoteRef := "HEAD"
	if branch != "HEAD" && branch != "" {
		remoteRef = "refs/heads/" + branch
	}

	remote, err := gitOutput(repoDir, "ls-remote", "origin", remoteRef)
	if err != nil || strings.TrimSpace(remote) == "" {
		remote, err = gitOutput(repoDir, "ls-remote", "origin", "HEAD")
	}
	if err != nil {
		return "update unknown"
	}

	fields := strings.Fields(remote)
	if len(fields) == 0 {
		return "update unknown"
	}
	if fields[0] != local {
		return "update available"
	}
	return "up to date"
}

func repoDirFromMeta(configDir string, a app, workDir string) string {
	metaPath := filepath.Join(configDir, a.Name+"-install-meta.toml")
	if fileExists(metaPath) {
		if repoDir := readMetaValue(metaPath, "repo_dir"); repoDir != "" {
			return repoDir
		}
	}
	if workDir == "" {
		return ""
	}
	return filepath.Join(workDir, a.RepoDir)
}

func readMetaValue(path, key string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	prefix := key + " = "
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		return strings.Trim(value, `"`)
	}
	return ""
}

func gitOutput(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", ctx.Err()
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func findWorkDir() (string, error) {
	for _, dir := range []string{os.Getenv("DELBYAPPS_WORKDIR"), os.Getenv("DELBYSOFT_WORKDIR"), defaultWorkDir} {
		if isWorkDir(dir) {
			return filepath.Clean(dir), nil
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if dir := findWorkDirFrom(wd); dir != "" {
			return dir, nil
		}
	}
	if exe, err := os.Executable(); err == nil {
		if dir := findWorkDirFrom(filepath.Dir(exe)); dir != "" {
			return dir, nil
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		for _, dir := range []string{filepath.Join(home, "Work"), filepath.Join(home, "work"), filepath.Join(home, "Projects")} {
			if isWorkDir(dir) {
				return filepath.Clean(dir), nil
			}
		}
	}

	return "", errors.New("could not find delbysoft work directory; set DELBYAPPS_WORKDIR")
}

func findWorkDirFrom(start string) string {
	dir := filepath.Clean(start)
	for {
		if isWorkDir(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isWorkDir(dir string) bool {
	if dir == "" {
		return false
	}
	matches := 0
	for _, a := range apps {
		if dirExists(filepath.Join(dir, a.RepoDir)) {
			matches++
		}
	}
	return matches >= 3
}

func openInEditor(path string) error {
	if editor := strings.TrimSpace(os.Getenv("VISUAL")); editor != "" {
		return runEditor(editor, path)
	}
	if editor := strings.TrimSpace(os.Getenv("EDITOR")); editor != "" {
		return runEditor(editor, path)
	}

	switch runtime.GOOS {
	case "windows":
		return runCommand("notepad", path)
	case "darwin":
		return runCommand("open", "-t", path)
	default:
		for _, candidate := range [][]string{{"xdg-open", path}, {"nano", path}, {"vim", path}, {"vi", path}} {
			if _, err := exec.LookPath(candidate[0]); err == nil {
				return runCommand(candidate[0], candidate[1:]...)
			}
		}
	}

	return errors.New("no editor found; set EDITOR or VISUAL")
}

func runEditor(editor, path string) error {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return errors.New("editor command is empty")
	}
	return runCommand(parts[0], append(parts[1:], path)...)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func yesNo(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}
