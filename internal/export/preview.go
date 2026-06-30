package export

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/anomalyco/omarchy-themegen/internal/theme"
)

func magickPath() (string, error) {
	path, err := exec.LookPath("magick")
	if err != nil {
		return "", fmt.Errorf("ImageMagick 'magick' is not installed; required for image generation")
	}
	return path, nil
}

func generatePlaceholderPNG(outputPath string, width, height int, bgColor, fgColor, label string) error {
	magick, err := magickPath()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, magick,
		"-size", fmt.Sprintf("%dx%d", width, height),
		fmt.Sprintf("canvas:%s", bgColor),
		"-fill", fgColor,
		"-gravity", "center",
		"-pointsize", "48",
		"-annotate", "0", fmt.Sprintf("%s (%dx%d)", label, width, height),
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("magick failed: %v: %s", err, string(output))
	}
	return nil
}

func GeneratePreviewWithSource(outputPath, sourcePath string, width, height int, bgColor, fgColor, accentColor string) error {
	magick, err := magickPath()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a simple preview:
	// Resize source to left half, put color palette swatches on right
	cmd := exec.CommandContext(ctx, magick,
		sourcePath,
		"-resize", fmt.Sprintf("%dx%d^", width, height),
		"-gravity", "center",
		"-extent", fmt.Sprintf("%dx%d", width, height),
		"-fill", fgColor,
		"-gravity", "south",
		"-pointsize", "24",
		"-annotate", "0", fmt.Sprintf(" background: %s  foreground: %s  accent: %s", bgColor, fgColor, accentColor),
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("magick preview failed: %v: %s", err, string(output))
	}
	return nil
}

func writeNeovimLua(exportDir string, colors *theme.Colors) error {
	content := generateNeovimLuaContent(colors)
	return os.WriteFile(filepath.Join(exportDir, "neovim.lua"), []byte(content), 0644)
}

func generateNeovimLuaContent(colors *theme.Colors) string {
	var b strings.Builder

	b.WriteString("return {\n")
	b.WriteString(fmt.Sprintf("  accent = \"%s\",\n", colors.Accent))
	b.WriteString(fmt.Sprintf("  cursor = \"%s\",\n", colors.Cursor))
	b.WriteString(fmt.Sprintf("  foreground = \"%s\",\n", colors.Foreground))
	b.WriteString(fmt.Sprintf("  background = \"%s\",\n", colors.Background))
	b.WriteString(fmt.Sprintf("  color0 = \"%s\",\n", colors.Color0))
	b.WriteString(fmt.Sprintf("  color1 = \"%s\",\n", colors.Color1))
	b.WriteString(fmt.Sprintf("  color2 = \"%s\",\n", colors.Color2))
	b.WriteString(fmt.Sprintf("  color3 = \"%s\",\n", colors.Color3))
	b.WriteString(fmt.Sprintf("  color4 = \"%s\",\n", colors.Color4))
	b.WriteString(fmt.Sprintf("  color5 = \"%s\",\n", colors.Color5))
	b.WriteString(fmt.Sprintf("  color6 = \"%s\",\n", colors.Color6))
	b.WriteString(fmt.Sprintf("  color7 = \"%s\",\n", colors.Color7))
	b.WriteString(fmt.Sprintf("  color8 = \"%s\",\n", colors.Color8))
	b.WriteString(fmt.Sprintf("  color9 = \"%s\",\n", colors.Color9))
	b.WriteString(fmt.Sprintf("  color10 = \"%s\",\n", colors.Color10))
	b.WriteString(fmt.Sprintf("  color11 = \"%s\",\n", colors.Color11))
	b.WriteString(fmt.Sprintf("  color12 = \"%s\",\n", colors.Color12))
	b.WriteString(fmt.Sprintf("  color13 = \"%s\",\n", colors.Color13))
	b.WriteString(fmt.Sprintf("  color14 = \"%s\",\n", colors.Color14))
	b.WriteString(fmt.Sprintf("  color15 = \"%s\",\n", colors.Color15))
	b.WriteString("}\n")

	return b.String()
}
