package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

type ExportResult struct {
	Path       string
	BackupPath string
	Warnings   []string
}

func ThemeDirectory(tm *theme.ThemeModel, exportDir string, forceOverwrite bool) (*ExportResult, error) {
	return themeDirectory(tm, exportDir, forceOverwrite, false)
}

func ThemeDirectoryWithLivePreview(tm *theme.ThemeModel, exportDir string, forceOverwrite bool) (*ExportResult, error) {
	return themeDirectory(tm, exportDir, forceOverwrite, true)
}

func themeDirectory(tm *theme.ThemeModel, exportDir string, forceOverwrite, livePreview bool) (*ExportResult, error) {
	result := &ExportResult{Path: exportDir}
	srcData, err := os.ReadFile(tm.SourceImage)
	if err != nil {
		return nil, fmt.Errorf("cannot read source image: %w", err)
	}
	previewSource, cleanup, err := snapshotSourceImage(tm.SourceImage, srcData)
	if err != nil {
		return nil, err
	}
	defer cleanup()

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
	if _, err := GenerateOmarchyThemedConfigs(exportDir, tm.Colors); err != nil {
		return nil, fmt.Errorf("failed to generate Omarchy themed configs: %w", err)
	}

	// Create backgrounds directory and copy source image
	bgDir := filepath.Join(exportDir, "backgrounds")
	if err := os.MkdirAll(bgDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create backgrounds directory: %w", err)
	}

	bgFilename := filepath.Base(tm.SourceImage)
	if err := os.WriteFile(filepath.Join(bgDir, bgFilename), srcData, 0644); err != nil {
		return nil, fmt.Errorf("failed to copy source image to backgrounds: %w", err)
	}

	// Generate preview.png (1800x1012) as an Omarchy-style desktop mock.
	if err := GenerateDesktopPreview(
		filepath.Join(exportDir, "preview.png"),
		previewSource,
		1800, 1012,
		tm.Colors,
		tm.DirectionLabel,
	); err != nil {
		return nil, fmt.Errorf("failed to generate preview.png: %w", err)
	}

	// Generate preview-unlock.png (1920x1080)
	if err := GenerateDesktopPreview(
		filepath.Join(exportDir, "preview-unlock.png"),
		previewSource,
		1920, 1080,
		tm.Colors,
		"Lock Preview",
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

	// Write neovim.lua as a LazyVim plugin spec consumed through Omarchy's current/theme symlink.
	if err := GenerateNeovimLua(exportDir, tm.Colors, tm.LightMode); err != nil {
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

	if livePreview {
		if err := GenerateLiveDesktopPreview(filepath.Join(exportDir, "preview.png"), exportDir, 1800, 1012); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("live preview failed; kept generated preview.png: %v", err))
		}
	}

	return result, nil
}

func snapshotSourceImage(sourcePath string, data []byte) (string, func(), error) {
	tmp, err := os.CreateTemp("", "omarchy-themegen-source-*"+filepath.Ext(sourcePath))
	if err != nil {
		return "", func() {}, fmt.Errorf("snapshot source image: %w", err)
	}
	path := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(path)
		return "", func() {}, fmt.Errorf("snapshot source image: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(path)
		return "", func() {}, fmt.Errorf("snapshot source image: %w", err)
	}
	return path, func() { os.Remove(path) }, nil
}
