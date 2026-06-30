package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/anomalyco/omarchy-themegen/internal/export"
	"github.com/anomalyco/omarchy-themegen/internal/gen"
	"github.com/anomalyco/omarchy-themegen/internal/image"
	"github.com/anomalyco/omarchy-themegen/internal/theme"
	"github.com/anomalyco/omarchy-themegen/internal/validate"
)

var version = "0.2.0-dev"

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
	archivePath := flagSet.String("archive-path", "", "Archive output path (default: current directory)")
	showVersion := flagSet.Bool("version", false, "Print version and exit")
	directionID := flagSet.Int("direction", 0, "Direction to export (1, 2, or 3)")
	lightMode := flagSet.Bool("light", false, "Generate light theme directions")
	seed := flagSet.Int("seed", 0, "Generation seed for deterministic output")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "omarchy-themegen %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen <image>              Start interactive/TUI theme generation\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen --image <path> --name <name> --direction <1-3> [flags]\n")
		fmt.Fprintf(os.Stderr, "                                        Export a generated theme non-interactively\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flagSet.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen wallpaper.png\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen --image wallpaper.png --name mytheme --direction 1\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen --image wallpaper.png --name mytheme --direction 2 --light --yes\n")
		fmt.Fprintf(os.Stderr, "  omarchy-themegen --image wallpaper.png --name mytheme --direction 3 --json\n")
	}

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if *showVersion {
		fmt.Printf("omarchy-themegen %s\n", version)
		return 0
	}

	args := flagSet.Args()

	// Positional <image> argument (interactive/TUI intent)
	if *imagePath == "" && len(args) > 0 {
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Error: only one source image is supported")
			return 1
		}
		srcImage := args[0]

		if *jsonMode {
			outputJSON(map[string]string{
				"status":  "unimplemented",
				"message": "TUI is not implemented yet. Use --image, --name, and --direction flags for non-interactive export.",
			})
			return 0
		}

		fmt.Println("Interactive TUI theme generation is not implemented yet.")
		fmt.Printf("\nTo export a generated theme non-interactively:\n")
		fmt.Printf("  omarchy-themegen --image %s --name <theme-name> --direction <1|2|3>\n", srcImage)
		fmt.Printf("\nFlags: --json, --yes, --archive, --archive-path, --output, --light, --seed\n")
		return 0
	}

	// Non-interactive export mode
	if *imagePath == "" {
		flagSet.Usage()
		return 1
	}

	if *themeName == "" {
		fmt.Fprintln(os.Stderr, "Error: --name is required for non-interactive export")
		return 1
	}

	if *directionID < 1 || *directionID > 3 {
		fmt.Fprintln(os.Stderr, "Error: --direction is required (must be 1, 2, or 3)")
		return 1
	}

	// Validate source image
	imgResult := image.Validate(*imagePath)
	if !imgResult.Valid {
		for _, errMsg := range imgResult.Errors {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		}
		return 1
	}
	for _, warn := range imgResult.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warn)
	}

	// Create generation options
	opts, err := gen.NewGenerationOptions(*imagePath, *seed, *lightMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	// Extract dominant colors
	dominantColors, err := gen.ExtractDominantColors(*imagePath, 12)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: color extraction failed: %s\n", err)
		return 1
	}

	// Generate palette candidates
	candidates, err := gen.GeneratePalettes(dominantColors, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: palette generation failed: %s\n", err)
		return 1
	}

	// Select the requested direction
	dirIdx := *directionID - 1
	if dirIdx < 0 || dirIdx >= len(candidates) {
		fmt.Fprintf(os.Stderr, "Error: invalid direction %d\n", *directionID)
		return 1
	}
	selected := candidates[dirIdx]

	// Report warnings from palette validation
	for _, w := range selected.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: direction %d: %s\n", selected.ID, w)
	}

	// Build direction
	dir := theme.Direction{
		ID:          selected.ID,
		Label:       selected.Label,
		Fingerprint: opts.Fingerprint,
		Colors:      selected.Colors,
		Warnings:    selected.Warnings,
		LightMode:   opts.LightMode,
	}

	// Build Theme Model from direction
	tm, err := theme.NewThemeModelFromDirection(*themeName, *imagePath, imgResult, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}
	tm.Version = version

	// Pre-export validation
	if err := validate.PreExport(tm); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %s\n", err)
		return 1
	}

	// Determine output directory
	exportDir := *outputDir
	if exportDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
			return 1
		}
		exportDir = filepath.Join(home, ".config", "omarchy", "themes", tm.NormalizedName)
	}

	// Export theme directory
	exportResult, err := export.ThemeDirectory(tm, exportDir, *yesMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	// Post-export validation
	postResult := validate.PostExport(exportDir, tm.NormalizedName)

	// Create archive if requested
	var archiveResults *export.ArchiveResult
	if *archiveMode {
		arcPath := *archivePath
		if arcPath == "" {
			arcPath = tm.NormalizedName + ".tar.gz"
		}
		archiveResults, err = export.CreateArchive(exportDir, tm.NormalizedName, arcPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: archive creation failed: %v\n", err)
		}
	}

	if *jsonMode {
		outputJSON(buildJSONResult(tm, exportResult, archiveResults, postResult))
	} else {
		printResults(tm, exportResult, archiveResults, postResult)
	}

	if !postResult.Passed {
		return 1
	}
	return 0
}

func printResults(tm *theme.ThemeModel, exportRes *export.ExportResult, archiveRes *export.ArchiveResult, postRes *validate.PostExportResult) {
	fmt.Printf("\nExport complete.\n")
	fmt.Printf("  Theme name:  %s\n", tm.NormalizedName)
	fmt.Printf("  Direction:   %d (%s)\n", tm.DirectionID, tm.DirectionLabel)
	if tm.LightMode {
		fmt.Printf("  Mode:        light\n")
	} else {
		fmt.Printf("  Mode:        dark\n")
	}
	fmt.Printf("  Directory:   %s\n", exportRes.Path)

	if exportRes.BackupPath != "" {
		fmt.Printf("  Backup:      %s\n", exportRes.BackupPath)
	}

	if archiveRes != nil {
		fmt.Printf("  Archive:     %s\n", archiveRes.Path)
	}

	if postRes.OmarchyInstalled {
		fmt.Printf("  Omarchy:     detected (theme list check passed)\n")
	} else {
		fmt.Printf("  Omarchy:     not detected (reduced validation confidence)\n")
	}

	for _, w := range postRes.Warnings {
		fmt.Printf("  Warning:     %s\n", w)
	}

	fmt.Printf("\nApply manually with: omarchy theme set %s\n", tm.NormalizedName)
}

func buildJSONResult(tm *theme.ThemeModel, exportRes *export.ExportResult, archiveRes *export.ArchiveResult, postRes *validate.PostExportResult) map[string]interface{} {
	result := map[string]interface{}{
		"status":           "ok",
		"theme_name":       tm.NormalizedName,
		"direction":        tm.DirectionID,
		"direction_label":  tm.DirectionLabel,
		"light_mode":       tm.LightMode,
		"export_directory": exportRes.Path,
	}

	if exportRes.BackupPath != "" {
		result["backup_path"] = exportRes.BackupPath
	}

	if archiveRes != nil {
		result["archive_path"] = archiveRes.Path
	}

	result["omarchy_detected"] = postRes.OmarchyInstalled

	validationMsgs := []string{}
	for _, w := range postRes.Warnings {
		validationMsgs = append(validationMsgs, "warning: "+w)
	}
	if !postRes.Passed {
		for _, e := range postRes.Errors {
			validationMsgs = append(validationMsgs, "error: "+e)
		}
	}
	result["post_export_validation"] = validationMsgs

	return result
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
