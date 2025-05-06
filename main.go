package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// isGitRepo checks if the given directory is a git repository
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// findGitRepos searches for git repos recursively from the given root
func findGitRepos(root string) ([]string, error) {
	var repos []string

	// Check if the root directory itself is a git repository
	if isGitRepo(root) {
		repos = append(repos, root)
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "Permission denied: %v\n", path)
				return filepath.SkipDir
			}
			return err
		}

		// Skip the root directory since we've already checked it
		if path == root {
			return nil
		}

		// Skip hidden directories (those starting with .)
		if info.IsDir() && strings.HasPrefix(filepath.Base(path), ".") && path != root {
			return filepath.SkipDir
		}

		// If this directory is a git repository, add it to our list
		if info.IsDir() && isGitRepo(path) {
			repos = append(repos, path)
			return filepath.SkipDir // Skip traversing into git repositories
		}

		return nil
	})

	return repos, err
}

// getDirectoryNames extracts just the directory names (not full paths) from a list of paths
func getDirectoryNames(paths []string) map[string]string {
	// Using a map to store name->path mapping
	dirMap := make(map[string]string)

	for _, path := range paths {
		name := filepath.Base(path)
		dirMap[name] = path
	}

	return dirMap
}

// createTmuxSession creates a new tmux session with the specified name and directory
func createTmuxSession(name, directory string) error {
	// Check if session already exists
	checkCmd := exec.Command("tmux", "has-session", "-t", name)
	err := checkCmd.Run()

	if err == nil {
		// Session exists, attach to it
		attachCmd := exec.Command("tmux", "attach", "-t", name)
		attachCmd.Stdin = os.Stdin
		attachCmd.Stdout = os.Stdout
		attachCmd.Stderr = os.Stderr
		return attachCmd.Run()
	}

	// Create new session with first window named "nvim"
	createCmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", directory, "-n", "nvim")
	if err := createCmd.Run(); err != nil {
		return err
	}

	// Run nvim in the first window
	nvimCmd := exec.Command("tmux", "send-keys", "-t", name+":0", "nvim", "Enter")
	if err := nvimCmd.Run(); err != nil {
		return err
	}

	// Create second window named "server"
	serverCmd := exec.Command("tmux", "new-window", "-t", name+":1", "-n", "server", "-c", directory)
	if err := serverCmd.Run(); err != nil {
		return err
	}

	// Create third window named "term"
	termCmd := exec.Command("tmux", "new-window", "-t", name+":2", "-n", "term", "-c", directory)
	if err := termCmd.Run(); err != nil {
		return err
	}

	// Select the first window
	selectCmd := exec.Command("tmux", "select-window", "-t", name+":0")
	if err := selectCmd.Run(); err != nil {
		return err
	}

	// Attach to the session
	attachCmd := exec.Command("tmux", "attach", "-t", name)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}

// model represents the bubbletea UI state
type model struct {
	options  []string
	cursor   int
	selected int
	dirMap   map[string]string
}

// initialModel initializes the bubbletea model
func initialModel(options []string, dirMap map[string]string) model {
	return model{
		options:  options,
		cursor:   0,
		selected: -1,
		dirMap:   dirMap,
	}
}

// Init is the bubbletea initialization function
func (m model) Init() tea.Cmd {
	return nil
}

// Update is the bubbletea update function that handles messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			return m, tea.Quit
		}
	}
	return m, nil
}

// View is the bubbletea view function that renders the UI
func (m model) View() string {
	s := "Select a repository:\n\n"

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}

	s += "\nPress q to quit.\n"
	return s
}

func main() {
	var searchDir string

	// Check if a directory argument was provided
	if len(os.Args) > 1 {
		// Use the provided directory
		providedDir := os.Args[1]
		absDir, err := filepath.Abs(providedDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}
		searchDir = absDir
	} else {
		// Use current directory as default
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		searchDir = currentDir
	}

	fmt.Printf("Searching for git repositories in: %s\n", searchDir)

	// Find git repositories
	repos, err := findGitRepos(searchDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding git repositories: %v\n", err)
		os.Exit(1)
	}

	if len(repos) == 0 {
		fmt.Println("No git repositories found.")
		os.Exit(0)
	}

	// Get directory names and create mapping to full paths
	dirMap := getDirectoryNames(repos)

	// Create a slice of directory names for the selection
	var options []string
	for name := range dirMap {
		options = append(options, name)
	}

	// Create bubbletea model for repository selection
	p := tea.NewProgram(initialModel(options, dirMap))
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running bubbletea program: %v\n", err)
		os.Exit(1)
	}

	// Get the selected repository from the model
	m, ok := result.(model)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: could not get model from program\n")
		os.Exit(1)
	}

	if m.selected == -1 {
		fmt.Println("No repository selected.")
		os.Exit(0)
	}

	selected := options[m.selected]
	selectedPath := dirMap[selected]

	// Create tmux session
	err = createTmuxSession(selected, selectedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tmux session: %v\n", err)
		os.Exit(1)
	}
}
