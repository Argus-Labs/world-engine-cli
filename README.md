# world-engine-cli

temporary testing notes:

install this to add new commands to the cli:
`go install github.com/spf13/cobra-cli@latest`

Then run:
`cobra-cli add create -p 'configCmd` to add a command line parameter "create" with an option "-p" that takes an argument 'configCmd'

see newProject.go to see how this CLI utilizes these libraries to construct a TUI after a command has been run:
- https://github.com/charmbracelet/lipgloss for terminal rendering
- https://github.com/charmbracelet/bubbletea for the tui framework
- https://github.com/charmbracelet/bubbles for tui widgets that work with bubbletea

All this thing currently does is clone the starter-game-template.

Test with these commands; in the project dir run:
1. `go build`
2. `./world create myproject`

## Installation

## Go Install

- Install latest available release:
```
go install pkg.world.dev/world-cli@latest
```

- Install specific release tag:
```
go install pkg.world.dev/world-cli@<release tag>
```

## Curl

- Install latest available release:
```
curl https://install.world.dev/cli! | bash
```

- Install specific release tag:
```
curl https://install.world.dev/cli@<release tag>! | bash
```
