package tmux

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CreateTmuxSession creates a new tmux session with the specified name and directory
// If forceAttach is true and a session exists, it will attach to the existing session without prompting
// If forceRecreate is true and a session exists, it will kill and recreate the session without prompting
func CreateTmuxSession(name string, directory string, forceAttach bool, forceRecreate bool) error {
	// Check if session already exists
	checkCmd := exec.Command("tmux", "has-session", "-t", name)
	err := checkCmd.Run()

	if err == nil {
		// Session exists, determine what to do
		if forceAttach {
			// Automatically attach to existing session
			attachCmd := exec.Command("tmux", "attach", "-t", name)
			attachCmd.Stdin = os.Stdin
			attachCmd.Stdout = os.Stdout
			attachCmd.Stderr = os.Stderr
			return attachCmd.Run()
		} else if forceRecreate {
			// Automatically kill existing session
			killCmd := exec.Command("tmux", "kill-session", "-t", name)
			if err := killCmd.Run(); err != nil {
				return fmt.Errorf("failed to kill existing session: %w", err)
			}
			// Continue to create new session below
		} else {
			// Prompt user for action
			fmt.Printf("Session '%s' already exists. Choose an option:\n", name)
			fmt.Println("[a/y] Attach")
			fmt.Println("[k/n] Kill and recreate")
			fmt.Println("[q/c] Cancel")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			switch input {
			case "a", "y":
				// Attach to existing session
				attachCmd := exec.Command("tmux", "attach", "-t", name)
				attachCmd.Stdin = os.Stdin
				attachCmd.Stdout = os.Stdout
				attachCmd.Stderr = os.Stderr
				return attachCmd.Run()
			case "k", "n":
				// Kill existing session
				killCmd := exec.Command("tmux", "kill-session", "-t", name)
				if err := killCmd.Run(); err != nil {
					return fmt.Errorf("failed to kill existing session: %w", err)
				}
				// Continue to create new session below
			case "q", "c", "":
				// Cancel operation
				return nil
			default:
				fmt.Println("Invalid option, canceling operation")
				return nil
			}
		}
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
