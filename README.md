# rttui

A terminal user interface for [Remember The Milk](https://www.rememberthemilk.com/).

Browse, search, add, and complete tasks without leaving the terminal.

## Features

- Task list grouped by date bucket: Overdue, Today, Tomorrow, day-of-week, No Due Date
- Within each bucket, tasks sorted by priority then due time
- Color-coded priority bar on each task (red = high, yellow = medium, blue = low)
- RTM filter search with vim-style navigation
- Smart task input with syntax highlighting for due dates, priorities, lists, tags, and recurrence
- Completes tasks in-place and refreshes the list

## Requirements

- Go 1.22+
- A Remember The Milk account
- An RTM API key and shared secret — request one at https://www.rememberthemilk.com/services/api/

## Setup

1. Clone and build:

```
git clone <repo>
cd rttui
go build -o rttui .
```

2. Set your credentials as environment variables:

```
export RTM_API_KEY=<your api key>
export RTM_SHARED_SECRET=<your shared secret>
```

   See `.envrc.example` for a 1Password reference pattern if you use `direnv` + 1Password CLI.

3. Run:

```
./rttui
```

On first run, rttui will print an authorization URL. Open it in your browser, grant access, then press Enter. The token is saved to `$XDG_CONFIG_HOME/rttui/token` (typically `~/.config/rttui/token`) and reused on subsequent runs.

## Usage

```
./rttui [filter]
```

Pass an optional RTM filter string as an argument to pre-filter the task list on startup. Uses the same syntax as RTM's advanced search (e.g. `"list:Work AND priority:1"`).

### Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` or arrow keys | Move up/down |
| `/` | Open search bar |
| `n` | Add a new task |
| `x` | Complete selected task |
| `q` / `ctrl+c` | Quit |

### Search bar

The search bar accepts any RTM filter expression. It supports vim-style editing:

| Key | Action |
|-----|--------|
| `Enter` | Submit search |
| `Esc` (insert mode) | Switch to normal mode |
| `Esc` (normal mode) | Close search bar |
| `i` / `a` / `A` / `I` | Enter insert mode |
| `h` / `l` | Move cursor left/right |
| `w` / `b` | Jump word forward/backward |
| `0` / `$` | Jump to start/end of line |
| `x` | Delete character under cursor |

### Adding tasks (Smart Add)

Press `n` to open the add bar. The input supports a Smart Add syntax with live syntax highlighting:

| Prefix | Meaning | Example |
|--------|---------|---------|
| `!1` `!2` `!3` | Priority (high/medium/low) | `!1` |
| `^` | Due date | `^tomorrow`, `^friday` |
| `#` | List | `#Work` |
| `%` | Tag | `%errand` |
| `*` | Recurrence | `*weekly` |

Press `Enter` to submit, `Esc` to cancel.

> The `%tag` syntax is a local convenience — rttui converts it to `#tag` when sending to RTM, since RTM uses `#` for both lists and tags.
