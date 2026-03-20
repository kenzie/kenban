# kenban

Plain-text personal kanban for solo developers. Trello for one, in a text file.

## Install

Requires Go 1.26+.

```sh
git clone <repo-url> && cd kenban
make install
```

This builds the binary and installs `kenban` (and `kb` shortcut) to `/usr/local/bin`.

## Usage

`kb` is an alias for `kenban` — they are interchangeable.

### Interactive board

```sh
kenban open
```

Opens a TUI kanban board with a 2x2 grid of columns (todo, doing, blocked, done). Use arrow keys to navigate, Enter to advance a task to the next state.

### Add a task

```sh
kenban add "[myproject] Stripe onboarding"
kenban add "[myproject] Fix auto-reload bug @high +payments"
```

Tasks are created in `[todo]` state.

### List tasks

```sh
kenban list               # all tasks
kenban list myproject     # filter by project
```

### Filter by state

```sh
kenban state doing
kenban state blocked
```

### Move a task

```sh
kenban move 3 doing       # move task 3 to doing
kenban done 3             # shortcut: move task 3 to done
```

### Edit a task

```sh
kenban edit 3             # opens in $EDITOR (default: vi)
```

### Copy a task to clipboard

```sh
kenban copy 3             # copies task 3 text to clipboard
```

Uses `pbcopy` on macOS, `xclip` or `xsel` on Linux.

### List projects

```sh
kenban projects
```

## Task format

One task per line in `tasks.txt`:

```
[state] [project] description @tag +label
```

- **state**: `todo`, `doing`, `blocked`, `done`
- **project**: any name in brackets
- **description**: free text, may include `@tags` and `+labels`

Example:

```
[todo] [myproject] Stripe onboarding @high +payments
[doing] [myproject] Fix auto-reload bug
[blocked] [myproject] Budget export waiting on schema
[done] [myproject] Initial billing model
```

## File location

kenban looks for tasks in this order:

1. `./tasks.txt` (current directory)
2. `~/.kenban/tasks.txt` (home directory)

The file and directories are created automatically if they don't exist.

## Development

```sh
make build    # build binary
make test     # run tests
make clean    # remove binary
```
