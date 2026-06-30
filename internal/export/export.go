package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anomalyco/omarchy-themegen/internal/theme"
)

type ExportResult struct {
	Path       string
	BackupPath string
}

func ThemeDirectory(tm *theme.ThemeModel, exportDir string, forceOverwrite bool) (*ExportResult, error) {
	result := &ExportResult{Path: exportDir}

	// Check for existing directory
	if _, err := os.Stat(exportDir); err == nil {
		if !forceOverwrite {
			return nil, fmt.Errorf(
				"theme directory already exists at %s; use --yes to replace with backup",
				exportDir,
			)
		}

		// Create timestamped backup
		backupPath, err := CreateBackup(exportDir)
		if err != nil {
			return nil, fmt.Errorf("backup failed: %w", err)
		}
		result.BackupPath = backupPath
	}

	// Create the export directory
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create export directory: %w", err)
	}

	// Write colors.toml
	if err := os.WriteFile(
		filepath.Join(exportDir, "colors.toml"),
		[]byte(tm.Colors.ToTOML()),
		0644,
	); err != nil {
		return nil, fmt.Errorf("failed to write colors.toml: %w", err)
	}

	// Create backgrounds directory and copy source image
	bgDir := filepath.Join(exportDir, "backgrounds")
	if err := os.MkdirAll(bgDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create backgrounds directory: %w", err)
	}

	srcData, err := os.ReadFile(tm.SourceImage)
	if err != nil {
		return nil, fmt.Errorf("cannot read source image: %w", err)
	}

	bgFilename := filepath.Base(tm.SourceImage)
	if err := os.WriteFile(filepath.Join(bgDir, bgFilename), srcData, 0644); err != nil {
		return nil, fmt.Errorf("failed to copy source image to backgrounds: %w", err)
	}

	// Generate preview.png (1800x1012) using source image + direction colors
	if err := GeneratePreviewWithSource(
		filepath.Join(exportDir, "preview.png"),
		tm.SourceImage,
		1800, 1012,
		tm.Colors.Background,
		tm.Colors.Foreground,
		tm.Colors.Accent,
	); err != nil {
		return nil, fmt.Errorf("failed to generate preview.png: %w", err)
	}

	// Generate preview-unlock.png (1920x1080)
	if err := GeneratePreviewWithSource(
		filepath.Join(exportDir, "preview-unlock.png"),
		tm.SourceImage,
		1920, 1080,
		tm.Colors.Background,
		tm.Colors.Foreground,
		tm.Colors.Accent,
	); err != nil {
		return nil, fmt.Errorf("failed to generate preview-unlock.png: %w", err)
	}

	// Generate unlock.png
	if err := generatePlaceholderPNG(
		filepath.Join(exportDir, "unlock.png"),
		1920, 1080,
		tm.Colors.Background,
		tm.Colors.Accent,
		"Unlock",
	); err != nil {
		return nil, fmt.Errorf("failed to generate unlock.png: %w", err)
	}

	// Write neovim.lua (minimal Aether config)
	if err := writeNeovimLua(exportDir, tm.Colors); err != nil {
		return nil, fmt.Errorf("failed to write neovim.lua: %w", err)
	}

	// Write light.mode for explicit light themes
	if tm.LightMode {
		if err := os.WriteFile(filepath.Join(exportDir, "light.mode"), []byte("true\n"), 0644); err != nil {
			return nil, fmt.Errorf("failed to write light.mode: %w", err)
		}
	}

	// Write README.md
	readmeContent := GenerateREADME(tm)
	if err := os.WriteFile(filepath.Join(exportDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write README.md: %w", err)
	}

	return result, nil
}
