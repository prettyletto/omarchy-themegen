package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/prettyletto/omarchy-themegen/internal/export"
	"github.com/prettyletto/omarchy-themegen/internal/gen"
	"github.com/prettyletto/omarchy-themegen/internal/omarchy"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
	"github.com/prettyletto/omarchy-themegen/internal/tui"
	"github.com/prettyletto/omarchy-themegen/internal/workflow"
)

var version = "1.0.0"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				version = fmt.Sprintf("%s (%s)", version, s.Value[:7])
			}
		}
	}
}

func main() {
	os.Exit(run())
}

func run() int {
	flagSet := flag.NewFlagSet("omarchy-themegen", flag.ExitOnError)

	imagePath := flagSet.String("image", "", "Path to theme source image")
	themeName := flagSet.String("name", "", "Theme name for export")
	outputDir := flagSet.String("output", "", "Output directory (default: ~/.config/omarchy/themes/<name>)")
	jsonMode := flagSet.Bool("json", false, "Emit JSON output")
	yesMode := flagSet.Bool("yes", false, "Confirm overwrite with backup")
	archiveMode := flagSet.Bool("archive", false, "Also create a finished-theme archive")
	archivePath := flagSet.String("archive_path", "", "Archive output path (default: current directory)")
	archiveOnly := flagSet.Bool("archive_only", false, "Create archive only, don't write to local theme directory")
	showVersion := flagSet.Bool("version", false, "Print version and exit")
	directionID := flagSet.Int("direction", 0, "Direction to export (1, 2, or 3) for whole-theme mode")
	lightMode := flagSet.Bool("light", false, "Generate light theme directions")
	seed := flagSet.Int("seed", 0, "Generation seed for deterministic output")
	mode := flagSet.String("mode", "whole-theme", "Selection mode: whole-theme or component-mix")
	recipeFlag := flagSet.String("recipe", "", "Export a recipe file to this path")
	reproducible := flagSet.Bool("reproducible", false, "Create reproducible archive (requires --yes)")
	livePreview := flagSet.Bool("live_preview", false, "Generate preview.png by applying the theme, opening real Hyprland windows, and taking a screenshot")
	replayFile := flagSet.String("replay", "", "Replay a recipe file (requires --image)")
	forceFingerprint := flagSet.Bool("force_fingerprint", false, "Skip fingerprint validation on recipe replay")

	groupDesktopShell := flagSet.Int("group_desktop_shell", 0, "")
	groupTerminals := flagSet.Int("group_terminals_and_tui", 0, "")
	groupEditor := flagSet.Int("group_editor", 0, "")
	groupAssets := flagSet.Int("group_assets_and_system", 0, "")

	overrideWaybar := flagSet.Int("override_waybar", 0, "")
	overrideHyprland := flagSet.Int("override_hyprland", 0, "")
	overrideHyprlock := flagSet.Int("override_hyprlock", 0, "")
	overrideMako := flagSet.Int("override_mako", 0, "")
	overrideWalker := flagSet.Int("override_walker", 0, "")
	overrideGhostty := flagSet.Int("override_ghostty", 0, "")
	overrideAlacritty := flagSet.Int("override_alacritty", 0, "")
	overrideKitty := flagSet.Int("override_kitty", 0, "")
	overrideBtop := flagSet.Int("override_btop", 0, "")
	overrideNeovim := flagSet.Int("override_neovim", 0, "")
	overrideChromium := flagSet.Int("override_chromium", 0, "")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "omarchy-themegen %s — generate Omarchy themes from images\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen <image>\n")
		fmt.Fprintf(os.Stderr, "      Open keyboard-only TUI.\n\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen --image <path> --name <name> --direction <1-3> [flags]\n")
		fmt.Fprintf(os.Stderr, "      Non-interactive whole-theme export.\n\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen --image <path> --name <name> --mode component-mix\n")
		fmt.Fprintf(os.Stderr, "      --group_desktop_shell <1-3> [other group flags] [flags]\n")
		fmt.Fprintf(os.Stderr, "      Non-interactive component-mix export.\n\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen --image <path> --name <name> --replay <recipe.json> [flags]\n")
		fmt.Fprintf(os.Stderr, "      Replay a previously exported recipe.\n\n")
		fmt.Fprintf(os.Stderr, "Modes: whole-theme (default), component-mix\n")
		fmt.Fprintf(os.Stderr, "Artifacts: --archive, --recipe <path>, --reproducible (requires --yes)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flagSet.PrintDefaults()
	}

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if *showVersion {
		printVersionInfo(*jsonMode)
		return 0
	}

	args := flagSet.Args()
	if *imagePath == "" && len(args) > 0 {
		srcImage := strings.Join(args, " ")
		if *jsonMode {
			outputJSON(map[string]string{"status": "error", "message": "JSON mode is only available with --image, --name options"})
			return 1
		}
		if err := tui.Run(srcImage, *archiveMode); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	if *imagePath == "" {
		flagSet.Usage()
		return 1
	}
	if *themeName == "" {
		fmt.Fprintln(os.Stderr, "Error: --name is required for non-interactive export")
		return 1
	}
	if *mode != "whole-theme" && *mode != "component-mix" {
		fmt.Fprintf(os.Stderr, "Error: invalid mode %q; must be 'whole-theme' or 'component-mix'\n", *mode)
		return 1
	}

	// Recipe replay
	if *replayFile != "" {
		return runRecipeReplay(*replayFile, *imagePath, *themeName, *outputDir, *jsonMode, *yesMode, *archiveMode, *archivePath, *recipeFlag, *reproducible, *forceFingerprint, *livePreview)
	}

	// Build options
	opts := workflow.Options{
		ImagePath:   *imagePath,
		ThemeName:   *themeName,
		OutputDir:   *outputDir,
		Seed:        *seed,
		LightMode:   *lightMode,
		DirectionID: *directionID,
		Mode:        *mode,
		Overwrite:   *yesMode,
		Archive:     *archiveMode || *archiveOnly,
		ArchivePath: *archivePath,
		ArchiveOnly: *archiveOnly,
		RecipePath:  *recipeFlag,
		LivePreview: *livePreview,
	}

	if *mode == "component-mix" {
		opts.GroupSources = buildGroupSources(*groupDesktopShell, *groupTerminals, *groupEditor, *groupAssets)
		opts.Overrides = buildOverrides(*overrideWaybar, *overrideHyprland, *overrideHyprlock, *overrideMako, *overrideWalker,
			*overrideGhostty, *overrideAlacritty, *overrideKitty, *overrideBtop, *overrideNeovim, *overrideChromium)

		if len(opts.GroupSources) == 0 {
			fmt.Fprintln(os.Stderr, "Error: component-mix mode requires at least one --group_* flag")
			return 1
		}
	} else {
		if opts.DirectionID < 1 || opts.DirectionID > 3 {
			fmt.Fprintln(os.Stderr, "Error: --direction is required (must be 1, 2, or 3)")
			return 1
		}
	}

	if *reproducible && !*yesMode {
		fmt.Fprintln(os.Stderr, "Error: --reproducible requires --yes to confirm inclusion of source image bytes")
		return 1
	}
	opts.Reproducible = *reproducible

	result := workflow.Run(opts)
	printWorkflowResult(result, *jsonMode)
	if !result.Success {
		return 1
	}
	return 0
}

func buildGroupSources(gDesktop, gTerminals, gEditor, gAssets int) map[string]int {
	m := make(map[string]int)
	if gDesktop >= 1 && gDesktop <= 3 {
		m[theme.GroupDesktopShell.ID] = gDesktop
	}
	if gTerminals >= 1 && gTerminals <= 3 {
		m[theme.GroupTerminalsAndTUI.ID] = gTerminals
	}
	if gEditor >= 1 && gEditor <= 3 {
		m[theme.GroupEditor.ID] = gEditor
	}
	if gAssets >= 1 && gAssets <= 3 {
		m[theme.GroupAssetsAndSystem.ID] = gAssets
	}
	return m
}

func buildOverrides(ovWaybar, ovHyprland, ovHyprlock, ovMako, ovWalker, ovGhostty, ovAlacritty, ovKitty, ovBtop, ovNeovim, ovChromium int) map[string]int {
	entries := []struct {
		dirID   int
		surface string
	}{
		{ovWaybar, "waybar"}, {ovHyprland, "hyprland"}, {ovHyprlock, "hyprlock"},
		{ovMako, "mako"}, {ovWalker, "walker"}, {ovGhostty, "ghostty"},
		{ovAlacritty, "alacritty"}, {ovKitty, "kitty"}, {ovBtop, "btop"},
		{ovNeovim, "neovim"}, {ovChromium, "chromium"},
	}
	m := make(map[string]int)
	for _, e := range entries {
		if e.dirID >= 1 && e.dirID <= 3 {
			m[e.surface] = e.dirID
		}
	}
	return m
}

func printWorkflowResult(r *workflow.Result, jsonMode bool) {
	if jsonMode {
		data := map[string]interface{}{
			"status":           "ok",
			"theme_name":       r.NormalizedName,
			"mode":             r.Mode,
			"direction":        r.DirectionID,
			"direction_label":  r.DirectionLabel,
			"light_mode":       r.LightMode,
			"export_directory": r.ExportPath,
		}
		if r.BackupPath != "" {
			data["backup_path"] = r.BackupPath
		}
		if r.ArchivePath != "" {
			data["archive_path"] = r.ArchivePath
		}
		if r.RecipePath != "" {
			data["recipe_path"] = r.RecipePath
		}
		if r.ReproduciblePath != "" {
			data["reproducible_archive_path"] = r.ReproduciblePath
		}
		data["omarchy_detected"] = r.OmarchyDetected
		if len(r.GroupSelections) > 0 {
			data["group_selections"] = r.GroupSelections
		}
		if len(r.Overrides) > 0 {
			data["overrides"] = r.Overrides
		}
		if len(r.Warnings) > 0 {
			data["warnings"] = r.Warnings
		}
		if len(r.Errors) > 0 {
			data["errors"] = r.Errors
			data["status"] = "error"
		}
		outputJSON(data)
		return
	}

	if !r.Success {
		for _, e := range r.Errors {
			fmt.Fprintf(os.Stderr, "Error: %s\n", e)
		}
		return
	}

	fmt.Printf("\nExport complete.\n")
	fmt.Printf("  Theme name:  %s\n", r.NormalizedName)
	fmt.Printf("  Mode:        %s\n", r.Mode)
	fmt.Printf("  Direction:   %d (%s)\n", r.DirectionID, r.DirectionLabel)
	if r.LightMode {
		fmt.Printf("  Light:       yes\n")
	}
	fmt.Printf("  Directory:   %s\n", r.ExportPath)
	if r.BackupPath != "" {
		fmt.Printf("  Backup:      %s\n", r.BackupPath)
	}
	if r.ArchivePath != "" {
		fmt.Printf("  Archive:     %s\n", r.ArchivePath)
	}
	if r.RecipePath != "" {
		fmt.Printf("  Recipe:      %s\n", r.RecipePath)
	}
	if r.ReproduciblePath != "" {
		fmt.Printf("  Reproducible: %s\n", r.ReproduciblePath)
	}
	if r.OmarchyDetected {
		fmt.Printf("  Omarchy:     detected\n")
	} else {
		fmt.Printf("  Omarchy:     not detected (reduced confidence)\n")
	}
	for _, w := range r.Warnings {
		fmt.Printf("  Warning:     %s\n", w)
	}
	if len(r.GroupSelections) > 0 && r.Mode == "component-mix" {
		fmt.Printf("  Groups:\n")
		for gid, did := range r.GroupSelections {
			fmt.Printf("    %s → direction %d\n", gid, did)
		}
	}
	if len(r.Overrides) > 0 {
		fmt.Printf("  Overrides:\n")
		for surf, did := range r.Overrides {
			fmt.Printf("    %s → direction %d\n", surf, did)
		}
	}
	fmt.Printf("\nApply manually with: omarchy theme set %s\n", r.NormalizedName)
}

func runRecipeReplay(recipeFile, imagePath, themeName, outputDir string, jsonMode, yesMode, archiveMode bool, archivePath, recipeFlag string, reproducible, forceFingerprint, livePreview bool) int {
	recipe, err := export.LoadRecipe(recipeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot load recipe: %s\n", err)
		return 1
	}

	opts := workflow.Options{
		ImagePath:    imagePath,
		ThemeName:    themeName,
		OutputDir:    outputDir,
		Seed:         recipe.Seed,
		LightMode:    recipe.LightMode,
		DirectionID:  recipe.DirectionID,
		Mode:         recipe.Mode,
		GroupSources: recipe.GroupSelections,
		Overrides:    recipe.Overrides,
		Overwrite:    yesMode,
		Archive:      archiveMode,
		ArchivePath:  archivePath,
		RecipePath:   recipeFlag,
		LivePreview:  livePreview,
	}

	if !forceFingerprint {
		genOpts, err := gen.NewGenerationOptions(imagePath, recipe.Seed, recipe.LightMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		if genOpts.Fingerprint != recipe.Fingerprint {
			fmt.Fprintf(os.Stderr, "Error: fingerprint mismatch.\n  Recipe:  %s\n  Image:   %s\n  Use --force_fingerprint to override.\n",
				recipe.Fingerprint, genOpts.Fingerprint)
			return 1
		}
	}

	if reproducible && !yesMode {
		fmt.Fprintln(os.Stderr, "Error: --reproducible requires --yes")
		return 1
	}
	opts.Reproducible = reproducible

	result := workflow.Run(opts)
	printWorkflowResult(result, jsonMode)
	if !result.Success {
		return 1
	}
	return 0
}

func printVersionInfo(jsonMode bool) {
	if jsonMode {
		info := map[string]interface{}{
			"app_version":       version,
			"generator_version": gen.GeneratorVersion,
			"go_version":        runtime.Version(),
		}
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			for _, s := range buildInfo.Settings {
				switch s.Key {
				case "vcs.revision":
					info["vcs_revision"] = s.Value
				case "vcs.time":
					info["vcs_time"] = s.Value
				}
			}
		}
		disc := omarchy.Discover()
		info["omarchy_installed"] = disc.Installed
		if disc.Installed {
			info["omarchy_confidence"] = disc.Confidence()
		}
		_, magickErr := exec.LookPath("magick")
		info["magick_available"] = magickErr == nil
		outputJSON(info)
		return
	}
	fmt.Printf("omarchy-themegen %s\n", version)
	fmt.Printf("  Generator:      %s\n", gen.GeneratorVersion)
	fmt.Printf("  Go:             %s\n", runtime.Version())
	disc := omarchy.Discover()
	if disc.Installed {
		fmt.Printf("  Omarchy:        detected (confidence=%s)\n", disc.Confidence())
	} else {
		fmt.Printf("  Omarchy:        not detected\n")
	}
	if _, err := exec.LookPath("magick"); err == nil {
		fmt.Printf("  ImageMagick:    available\n")
	} else {
		fmt.Printf("  ImageMagick:    NOT FOUND\n")
	}
}

func outputJSON(data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: JSON encoding failed: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(bytes)
	os.Stdout.Write([]byte("\n"))
}
