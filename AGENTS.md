# rttui

A Terminal User Interface for [Remember The Milk](https://www.rememberthemilk.com/) built in Go using Bubble Tea (v2) and Lip Gloss.

## Structure

- `main.go` — Entrypoint; reads env vars, authenticates, fetches tasks, launches TUI.
- `internal/rtm/` — Remember The Milk API client (auth, task fetching).
- `internal/ui/` — Bubble Tea TUI model and views.

## Environment

Requires `RTM_API_KEY` and `RTM_SHARED_SECRET` env vars. Secrets are managed via 1Password CLI (`op inject`). Copy `.envrc.example` → `.envrc` with `make init-env`.

## Make Targets

- `make` / `make all` — Init env + run.
- `make run` — `go run ./...`
- `make test` — `go test ./...`
- `make init-env` — Inject secrets from 1Password into `.envrc`.
- `make clean` — Remove built binary.

## Conventions

- Go module: `git.codegoalie.com/rttui.git`
- Use Bubble Tea v2 (`charm.land/bubbletea/v2`) and Bubbles v2 (`charm.land/bubbles/v2`).
