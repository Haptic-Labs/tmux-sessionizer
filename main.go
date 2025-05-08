package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	git "github.com/Haptic-Labs/tmux-sessionizer/git"
	tmux "github.com/Haptic-Labs/tmux-sessionizer/tmux"
	ui "github.com/Haptic-Labs/tmux-sessionizer/ui"
	utils "github.com/Haptic-Labs/tmux-sessionizer/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Define flags
	var forceAttach bool
	var forceRecreate bool

	flag.BoolVar(&forceAttach, "a", false, "Automatically attach to existing session if it exists")
	flag.BoolVar(&forceAttach, "attach", false, "Automatically attach to existing session if it exists")
	flag.BoolVar(&forceRecreate, "k", false, "Automatically kill and recreate existing session if it exists")
	flag.BoolVar(&forceRecreate, "kill", false, "Automatically kill and recreate existing session if it exists")

	// Parse flags
	flag.Parse()

	var searchDir string

	// Check if a directory argument was provided
	args := flag.Args()
	if len(args) > 0 {
		// Use the provided directory
		providedDir := args[0]
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
	repos, err := git.FindGitRepos(searchDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding git repositories: %v\n", err)
		os.Exit(1)
	}

	if len(repos) == 0 {
		fmt.Println("No git repositories found.")
		os.Exit(0)
	}

	// Get directory names and create mapping to full paths
	dirMap := utils.GetDirectoryNames(repos)

	// Create a slice of directory names for the selection
	var options []string
	for name := range dirMap {
		options = append(options, name)
	}

	// Sort the options alphabetically (case-insensitive)
	sort.Slice(options, func(i, j int) bool {
		return strings.ToLower(options[i]) < strings.ToLower(options[j])
	})

	// Create bubbletea model for repository selection
	model := ui.InitializeModel(options, dirMap)
	p := tea.NewProgram(&model)
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running bubbletea program: %v\n", err)
		os.Exit(1)
	}

	// Get the selected repository from the model
	m, ok := result.(*ui.BubbleteaModel)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: could not get model from program\n")
		os.Exit(1)
	}

	if m.Selected == -1 {
		fmt.Println("No repository selected.")
		os.Exit(0)
	}

	selected := options[m.Selected]
	selectedPath := dirMap[selected]

	// Create tmux session
	err = tmux.CreateTmuxSession(selected, selectedPath, forceAttach, forceRecreate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tmux session: %v\n", err)
		os.Exit(1)
	}
}
