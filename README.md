# omarchy-themegen

`omarchy-themegen` turns one opaque still image into a complete Omarchy theme directory.

It generates five theme directions, lets you choose a whole-theme or component-mix composition, previews the resolved Theme Model, then exports files Omarchy can apply.

## Demo

[Watch the demo](assets/demo.mp4)

## Install

Requirements:

- Go 1.26.3 or newer;
- ImageMagick with `magick` on `PATH`;
- Omarchy is optional for export, but required to apply themes.

From this repository:

```bash
go install ./cmd/omarchy-themegen
```

From a published module version:

```bash
go install github.com/prettyletto/omarchy-themegen/cmd/omarchy-themegen@latest
```

## Usage

Open the keyboard-only TUI:

```bash
omarchy-themegen /path/to/wallpaper.png
```

Export a whole-theme direction non-interactively:

```bash
omarchy-themegen --image /path/to/wallpaper.png --name my-theme --direction 1
```

Export a component mix non-interactively:

```bash
omarchy-themegen \
  --image /path/to/wallpaper.png \
  --name my-mix \
  --mode component-mix \
  --group_desktop_shell 2 \
  --group_terminals_and_tui 3 \
  --group_editor 1 \
  --group_assets_and_system 5
```

Show version information:

```bash
omarchy-themegen --version
```

## Outputs

By default, exports go to:

```text
~/.config/omarchy/themes/<theme-name>
```

Generated themes include:

- `colors.toml`;
- `backgrounds/` with the source image;
- `preview.png`, `preview-unlock.png`, and `unlock.png`;
- `neovim.lua`;
- `README.md`;
- optional archive and recipe artifacts when requested.

Export and apply are separate. The TUI asks before applying with Omarchy.

## Component Mix

Component-mix mode assigns Surface Groups to generated directions, then allows per-surface overrides. The selection resolves into one final Theme Model before preview or export.

Surface Groups:

- Desktop Shell;
- Terminals And TUI;
- Editor;
- Assets And System.

## Development

Run the automated checks:

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./cmd/omarchy-themegen
```

## License

MIT. See `LICENSE`.
