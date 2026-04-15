package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"git.codegoalie.com/rttui/internal/config"
	"git.codegoalie.com/rttui/internal/rtm"
	"git.codegoalie.com/rttui/internal/ui"
)

func main() {
	apiKey := os.Getenv("RTM_API_KEY")
	secret := os.Getenv("RTM_SHARED_SECRET")
	if apiKey == "" || secret == "" {
		fmt.Fprintln(os.Stderr, "RTM_API_KEY and RTM_SHARED_SECRET must be set")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	filter := cfg.DefaultFilter
	if len(os.Args) > 1 {
		filter = os.Args[1]
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

	refreshInterval := time.Duration(cfg.RefreshIntervalSecs) * time.Second
	if cfg.RefreshIntervalSecs == 0 {
		refreshInterval = 60 * time.Second
	}

	p := tea.NewProgram(ui.NewModel(client, token, filter, cfg.AddPreset, refreshInterval, tasks))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
