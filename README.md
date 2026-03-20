# kenban

Plain-text personal kanban for solo developers. Trello for one, in a text file.

## Install

```sh
git clone <repo-url> && cd kenban
chmod +x bin/kenban

# Add to PATH (pick one):
export PATH="$PWD/bin:$PATH"          # session only
ln -s "$PWD/bin/kenban" /usr/local/bin/kenban
ln -s "$PWD/bin/kb" /usr/local/bin/kb  # shortcut
```

Or install as a gem:

```sh
gem build kenban.gemspec
gem install kenban-0.1.0.gem
```

## Usage

`kb` is an alias for `kenban` — they are interchangeable.

### Add a task

```sh
kenban add "[goaliebook] Stripe onboarding"
kenban add "[kioskbook] Fix auto-reload bug @high +payments"
```

Tasks are created in `[todo]` state.

### List tasks

```sh
kenban list               # all tasks
kenban list goaliebook    # filter by project
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

Example file:

```
[todo] [goaliebook] Stripe onboarding @high +payments
[doing] [kioskbook] Fix auto-reload bug
[blocked] [teambook] Budget export waiting on schema
[done] [goaliebook] Initial billing model
```

## File location

kenban looks for tasks in this order:

1. `./tasks.txt` (current directory)
2. `~/.kenban/tasks.txt` (home directory)

The file and directories are created automatically if they don't exist.

## Running tests

```sh
ruby -e "Dir.glob('test/test_*.rb').each { |f| require_relative f }"
# or
rake test
```

## Future ideas

- TUI with curses/ratatui-style interface
- `kenban archive` to move done tasks to an archive file
- Due dates and priorities
- Custom states
- `kenban grep` for full-text search
- Tab completion for shells
- Color output
