# pik

[日本語](./README.ja.md) | English

A lightweight Git GUI for macOS with **line-level staging**. Built with [Wails](https://wails.io) (Go + Svelte 5).

## Install

```sh
brew install t4traw/pik/pik
```

Then launch from any repo:

```sh
pik          # current directory
pik ~/repo   # explicit path
```

## Features

- File / hunk / **line-level** staging and unstaging
- Side-by-side diff view with clickable line selection (shift-click for range)
- Commit directly from the UI (⌘↩ to submit)
- Auto-refresh on window focus — no stale diffs after editing files outside
- Undo / redo for every mutation (⌘Z / ⌘⇧Z)
- Configurable diff font size via the settings panel
- **AI commit messages** — one click generates a commit message from the staged diff via your local `claude` CLI. Respects the repo's `CLAUDE.md` conventions automatically.

## Keyboard shortcuts

| Shortcut | Action |
|---|---|
| ⌘↩ / Ctrl+↩ | Commit |
| ⌘Z / Ctrl+Z | Undo last action |
| ⌘⇧Z / Ctrl+⇧Z | Redo |
| Shift-click | Range-select lines in a hunk |

## Development

Requires Go 1.23+, [Bun](https://bun.sh), and the [Wails CLI](https://wails.io/docs/gettingstarted/installation).

```sh
make dev       # hot-reload dev server
make build     # produce build/bin/pik.app
make install   # install shim to $HOME/.local/bin/pik
```

## Release

Automated via [release-please](https://github.com/googleapis/release-please). Merge the release PR → tag push triggers a universal macOS build, GitHub Release, and Homebrew tap bump.

## License

MIT
