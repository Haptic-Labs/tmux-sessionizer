# tmux-sessionizer

A simple CLI tool to create tmux sessions based on git repositories.

## Features

- Recursively searches for git repositories in the current directory
- Provides an interactive selection list of repository directories
- Creates a new tmux session with the selected directory name with 3 windows:
  - "nvim" - Opens Neovim
  - "server" - Empty window for running servers
  - "term" - Terminal window for miscellaneous commands

## Installation

```bash
# Clone the repository
git clone https://github.com/braydenbabbitt/tmux-sessionizer.git
cd tmux-sessionizer

# Install dependencies and build
go mod tidy
go build -o tmux-sessionizer

# Move to a directory in your PATH
sudo mv tmux-sessionizer /usr/local/bin/
# tmux-sessionizer

```
