package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"git.codegoalie.com/rttui.git/internal/rtm"
	"git.codegoalie.com/rttui.git/internal/ui"
)

func main() {
	filter := ""
	if len(os.Args) > 1 {
		filter = os.Args[1]
	}

	apiKey := os.Getenv("RTM_API_KEY")
	secret := os.Getenv("RTM_SHARED_SECRET")
	if apiKey == "" || secret == "" {
		fmt.Fprintln(os.Stderr, "RTM_API_KEY and RTM_SHARED_SECRET must be set")
		os.Exit(1)
	}

	client := rtm.NewClient(apiKey, secret)

	token, err := rtm.EnsureAuthenticated(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth error: %v\n", err)
		os.Exit(1)
	}

	tasks, err := client.GetTasks(token, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch tasks error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(client, token, filter, tasks))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
