package preview

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/prettyletto/omarchy-themegen/internal/export"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

func magickPath() (string, error) {
	path, err := exec.LookPath("magick")
	if err != nil {
		return "", fmt.Errorf("ImageMagick 'magick' is not installed")
	}
	return path, nil
}

func RenderDirectionPreview(outputPath, sourcePath string, dir theme.Direction, width, height int) error {
	return export.GenerateDesktopPreview(outputPath, sourcePath, width, height, dir.Colors, dir.Label)
}

func RenderComposedPreview(outputPath, sourcePath string, tm *theme.ThemeModel, width, height int) error {
	modeLabel := tm.Mode
	if tm.DirectionLabel != "" {
		modeLabel = tm.DirectionLabel
	}
	return export.GenerateDesktopPreview(outputPath, sourcePath, width, height, tm.Colors, modeLabel)
}

func GenerateDirectionPreviews(outputDir, sourcePath string, directions []theme.Direction) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create preview dir: %w", err)
	}

	var paths []string
	for _, d := range directions {
		p := filepath.Join(outputDir, fmt.Sprintf("direction-%d.png", d.ID))
		if err := RenderDirectionPreview(p, sourcePath, d, 1600, 900); err != nil {
			return paths, fmt.Errorf("direction %d preview: %w", d.ID, err)
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func GenerateComposedPreview(outputPath, sourcePath string, tm *theme.ThemeModel) error {
	return RenderComposedPreview(outputPath, sourcePath, tm, 1600, 900)
}
