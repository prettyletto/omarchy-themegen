package export

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prettyletto/omarchy-themegen/internal/theme"
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
	colors := &theme.Colors{
		Background: bgColor,
		Foreground: fgColor,
		Accent:     accentColor,
		Color0:     bgColor,
		Color1:     accentColor,
		Color2:     accentColor,
		Color3:     accentColor,
		Color4:     accentColor,
		Color5:     accentColor,
		Color6:     accentColor,
		Color7:     fgColor,
		Color8:     bgColor,
		Color9:     accentColor,
		Color10:    accentColor,
		Color11:    accentColor,
		Color12:    accentColor,
		Color13:    accentColor,
		Color14:    accentColor,
		Color15:    fgColor,
	}
	return GenerateDesktopPreview(outputPath, sourcePath, width, height, colors, "Omarchy Theme")
}

func GenerateDesktopPreview(outputPath, sourcePath string, width, height int, colors *theme.Colors, label string) error {
	magick, err := magickPath()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sx := func(v int) int { return v * width / 1800 }
	sy := func(v int) int { return v * height / 1012 }

	type textOp struct {
		color string
		size  int
		x     int
		y     int
		text  string
	}

	draws := []string{
		fmt.Sprintf("stroke 'none' fill '%s' rectangle 0,0 %d,%d", rgba(colors.Background, 0.62), width, height),
	}
	texts := []textOp{}

	draw := func(format string, args ...interface{}) {
		draws = append(draws, fmt.Sprintf(format, args...))
	}
	text := func(color string, size, x, y int, value string) {
		texts = append(texts, textOp{color: color, size: size, x: x, y: y, text: value})
	}
	roundRect := func(color string, x, y, w, h, r int) {
		draw("stroke 'none' fill '%s' roundrectangle %d,%d %d,%d %d,%d", color, x, y, x+w, y+h, r, r)
	}
	rect := func(color string, x, y, w, h int) {
		draw("stroke 'none' fill '%s' rectangle %d,%d %d,%d", color, x, y, x+w, y+h)
	}
	line := func(color string, x1, y1, x2, y2, stroke int) {
		draw("fill 'none' stroke '%s' stroke-width %d line %d,%d %d,%d", color, stroke, x1, y1, x2, y2)
	}
	border := func(color string, x, y, w, h, r int) {
		draw("fill 'none' stroke '%s' stroke-width %d roundrectangle %d,%d %d,%d %d,%d", color, maxPreview(1, sx(2)), x, y, x+w, y+h, r, r)
	}
	bar := func(color string, x, y, w, h int) {
		roundRect(color, x, y, w, h, maxPreview(2, sy(4)))
	}

	bg := colors.Background
	fg := colors.Foreground
	accent := colors.Accent
	muted := colors.Color7
	if muted == "" {
		muted = fg
	}
	panel := rgba(bg, 0.91)
	panelSoft := rgba(colors.Color8, 0.82)
	borderColor := rgba(accent, 0.52)
	shadow := "rgba(0,0,0,0.42)"
	softText := rgba(muted, 0.74)

	// Waybar strip.
	topBarH := maxPreview(32, sy(42))
	rect(rgba(bg, 0.86), 0, 0, width, topBarH)
	rect(rgba(accent, 0.78), 0, topBarH-maxPreview(2, sy(3)), width, maxPreview(2, sy(3)))
	text(fg, maxPreview(12, sy(15)), sx(34), sy(10), "1  2  3  4  work")
	text(accent, maxPreview(12, sy(15)), sx(785), sy(10), label)
	text(softText, maxPreview(11, sy(14)), sx(1390), sy(10), "CPU 18%   MEM 42%   NET 1.2M   BAT 100%")

	marginX := sx(34)
	marginBottom := sy(30)
	gapX := sx(22)
	gapY := sy(22)
	contentY := topBarH + sy(22)
	panelW := (width - 2*marginX - gapX) / 2
	panelH := (height - contentY - marginBottom - gapY) / 2
	leftX := marginX
	rightX := leftX + panelW + gapX
	topY := contentY
	bottomY := topY + panelH + gapY
	winR := maxPreview(12, sy(18))
	titleH := maxPreview(32, sy(43))
	padX := sx(20)
	padY := sy(16)

	window := func(x, y, w, h int, title string) {
		roundRect(shadow, x+sx(8), y+sy(10), w, h, winR)
		roundRect(panel, x, y, w, h, winR)
		roundRect(panelSoft, x, y, w, titleH+sy(5), winR)
		rect(panelSoft, x, y+titleH/2, w, titleH/2+sy(5))
		border(borderColor, x, y, w, h, winR)
		line(rgba(accent, 0.56), x, y+titleH, x+w, y+titleH, maxPreview(1, sy(2)))
		dotY := y + titleH/2
		dotR := maxPreview(3, sy(5))
		draw("stroke 'none' fill '%s' circle %d,%d %d,%d", colors.Color1, x+sx(18), dotY, x+sx(18)+dotR, dotY)
		draw("stroke 'none' fill '%s' circle %d,%d %d,%d", colors.Color3, x+sx(35), dotY, x+sx(35)+dotR, dotY)
		draw("stroke 'none' fill '%s' circle %d,%d %d,%d", colors.Color2, x+sx(52), dotY, x+sx(52)+dotR, dotY)
		text(fg, maxPreview(12, sy(16)), x+sx(74), y+sy(12), title)
	}

	// Top-left: LazyVim explorer with a code buffer.
	window(leftX, topY, panelW, panelH, "nvim  ~/.config/omarchy")
	nvimContentY := topY + titleH
	nvimSideW := panelW * 30 / 100
	rect(rgba(colors.Color8, 0.70), leftX, nvimContentY, nvimSideW, panelH-titleH)
	line(rgba(accent, 0.36), leftX+nvimSideW, nvimContentY, leftX+nvimSideW, topY+panelH, maxPreview(1, sx(2)))
	text(accent, maxPreview(12, sy(16)), leftX+padX, nvimContentY+padY, "EXPLORER")
	for i, row := range []struct {
		name  string
		color string
	}{
		{"~/.config/omarchy", fg},
		{"  backgrounds/", colors.Color4},
		{"  default/", colors.Color4},
		{"  themes/", colors.Color4},
		{"  current/theme.lua", accent},
		{"  waybar.css", colors.Color6},
		{"  hyprland.conf", colors.Color5},
		{"  ghostty.conf", colors.Color3},
	} {
		y := nvimContentY + padY + sy(32+i*28)
		if i == 4 {
			roundRect(rgba(accent, 0.18), leftX+sx(10), y-sy(5), nvimSideW-sx(20), sy(26), maxPreview(4, sy(6)))
		}
		text(row.color, maxPreview(10, sy(14)), leftX+padX, y, row.name)
	}
	codeX := leftX + nvimSideW + sx(22)
	codeY := nvimContentY + padY
	text(softText, maxPreview(10, sy(13)), codeX, codeY, "theme.lua")
	line(rgba(accent, 0.28), codeX, codeY+sy(25), leftX+panelW-sx(18), codeY+sy(25), maxPreview(1, sy(1)))
	codeLines := []struct {
		line  string
		color string
	}{
		{"local colors = {", colors.Color5},
		{fmt.Sprintf("  background = \"%s\",", colors.Background), fg},
		{fmt.Sprintf("  foreground = \"%s\",", colors.Foreground), fg},
		{fmt.Sprintf("  accent     = \"%s\",", colors.Accent), accent},
		{"}", colors.Color5},
		{"require(\"lazyvim\").setup({", colors.Color4},
		{"  explorer = true,", fg},
		{"  transparent = false,", fg},
		{"})", colors.Color4},
	}
	for i, row := range codeLines {
		y := codeY + sy(52+i*29)
		text(softText, maxPreview(10, sy(13)), codeX, y, fmt.Sprintf("%2d", i+1))
		text(row.color, maxPreview(10, sy(14)), codeX+sx(45), y, row.line)
	}
	rect(rgba(accent, 0.24), leftX, topY+panelH-sy(28), panelW, sy(28))
	text(fg, maxPreview(9, sy(12)), leftX+sx(18), topY+panelH-sy(22), "NORMAL  theme.lua  utf-8  lua")

	// Top-right: btop system dashboard.
	window(rightX, topY, panelW, panelH, "btop")
	btopY := topY + titleH + padY
	text(accent, maxPreview(13, sy(18)), rightX+padX, btopY, "CPU")
	text(softText, maxPreview(10, sy(13)), rightX+sx(112), btopY+sy(3), "Ryzen 7 7840U  18%  47C")
	chartX := rightX + padX
	chartY := btopY + sy(34)
	chartW := panelW - 2*padX
	chartH := sy(92)
	rect(rgba(colors.Color8, 0.50), chartX, chartY, chartW, chartH)
	for i, h := range []int{28, 52, 38, 74, 46, 62, 34, 80, 45, 58, 36, 69, 50, 42, 76, 31, 57, 66} {
		barW := chartW / 23
		x := chartX + sx(12) + i*(barW+sx(6))
		bar(accent, x, chartY+chartH-sy(h), barW, sy(h))
	}
	statsY := chartY + chartH + sy(24)
	for i, stat := range []struct {
		label string
		value string
		color string
	}{
		{"MEM", "7.9G / 18.8G", colors.Color2},
		{"DISK", "41% used", colors.Color3},
		{"NET", "1.2M down", colors.Color6},
	} {
		x := rightX + padX + i*(chartW/3)
		text(stat.color, maxPreview(11, sy(15)), x, statsY, stat.label)
		text(fg, maxPreview(10, sy(13)), x, statsY+sy(24), stat.value)
		bar(rgba(stat.color, 0.34), x, statsY+sy(49), chartW/4, sy(11))
		bar(stat.color, x, statsY+sy(49), chartW/(7-i), sy(11))
	}
	tableY := statsY + sy(88)
	text(softText, maxPreview(10, sy(13)), rightX+padX, tableY, "PID      CPU%   MEM%   Command")
	for i, row := range []string{"2451     8.3    6.2    nvim", "1822     5.7    4.8    ghostty", "3104     2.2    3.1    btop", "1440     1.8    2.6    hyprland"} {
		color := fg
		if i == 0 {
			color = accent
		}
		text(color, maxPreview(10, sy(13)), rightX+padX, tableY+sy(27+i*25), row)
	}

	// Bottom-left: shell with lsd -la output.
	window(leftX, bottomY, panelW, panelH, "ghostty  zsh")
	termY := bottomY + titleH + padY
	text(accent, maxPreview(12, sy(16)), leftX+padX, termY, "prettyletto@omarchy")
	text(fg, maxPreview(12, sy(16)), leftX+sx(190), termY, "~/.config/omarchy > lsd -la")
	lsRows := []struct {
		perm  string
		size  string
		name  string
		color string
	}{
		{"drwxr-xr-x", "4.0K", "backgrounds", colors.Color4},
		{"drwxr-xr-x", "4.0K", "themes", colors.Color4},
		{"drwxr-xr-x", "4.0K", "current", colors.Color4},
		{"-rw-r--r--", "2.8K", "colors.toml", colors.Color3},
		{"-rw-r--r--", "7.1K", "waybar.css", colors.Color6},
		{"-rw-r--r--", "1.6K", "hyprland.conf", colors.Color5},
		{"-rw-r--r--", "1.2K", "ghostty.conf", colors.Color2},
		{"-rw-r--r--", "930B", "mako.ini", colors.Color1},
	}
	for i, row := range lsRows {
		y := termY + sy(42+i*31)
		text(softText, maxPreview(10, sy(13)), leftX+padX, y, row.perm)
		text(softText, maxPreview(10, sy(13)), leftX+sx(170), y, row.size)
		text(row.color, maxPreview(11, sy(15)), leftX+sx(245), y, row.name)
	}
	text(accent, maxPreview(12, sy(16)), leftX+padX, bottomY+panelH-sy(45), "~ >")
	rect(rgba(accent, 0.86), leftX+sx(70), bottomY+panelH-sy(42), sx(10), sy(19))

	// Bottom-right: file manager.
	window(rightX, bottomY, panelW, panelH, "files  ~/Pictures/Bgs")
	filesY := bottomY + titleH
	fileSideW := panelW * 24 / 100
	rect(rgba(colors.Color8, 0.58), rightX, filesY, fileSideW, panelH-titleH)
	for i, place := range []string{"Home", "Desktop", "Documents", "Downloads", "Pictures", "Projects"} {
		y := filesY + sy(26+i*31)
		color := softText
		if place == "Pictures" {
			roundRect(rgba(accent, 0.20), rightX+sx(10), y-sy(5), fileSideW-sx(20), sy(26), maxPreview(4, sy(6)))
			color = accent
		}
		text(color, maxPreview(10, sy(14)), rightX+padX, y, place)
	}
	pathX := rightX + fileSideW + sx(22)
	pathY := filesY + sy(18)
	roundRect(rgba(colors.Color8, 0.46), pathX, pathY, panelW-fileSideW-sx(42), sy(34), maxPreview(8, sy(10)))
	text(fg, maxPreview(10, sy(13)), pathX+sx(14), pathY+sy(8), "/home/prettyletto/Pictures/Bgs")
	gridX := pathX
	gridY := pathY + sy(58)
	cardW := (panelW - fileSideW - sx(72)) / 3
	cardH := sy(96)
	items := []struct {
		name  string
		color string
	}{
		{"ChatGPT Image", accent},
		{"violet-wall", colors.Color5},
		{"omarchy", colors.Color4},
		{"tokyo-night", colors.Color6},
		{"preview.png", colors.Color3},
		{"lock.png", colors.Color2},
	}
	for i, item := range items {
		col := i % 3
		row := i / 3
		x := gridX + col*(cardW+sx(16))
		y := gridY + row*(cardH+sy(48))
		roundRect(rgba(colors.Color8, 0.54), x, y, cardW, cardH, maxPreview(8, sy(12)))
		rect(rgba(item.color, 0.28), x+sx(10), y+sy(10), cardW-sx(20), cardH-sy(30))
		border(rgba(item.color, 0.44), x, y, cardW, cardH, maxPreview(8, sy(12)))
		text(item.color, maxPreview(9, sy(12)), x+sx(6), y+cardH+sy(10), item.name)
	}

	args := []string{
		sourcePath,
		"-resize", fmt.Sprintf("%dx%d^", width, height),
		"-gravity", "center",
		"-extent", fmt.Sprintf("%dx%d", width, height),
	}
	for _, draw := range draws {
		args = append(args, "-draw", draw)
	}
	for _, op := range texts {
		args = append(args,
			"-fill", op.color,
			"-gravity", "northwest",
			"-pointsize", fmt.Sprintf("%d", op.size),
			"-annotate", fmt.Sprintf("+%d+%d", op.x, op.y), op.text,
		)
	}
	args = append(args, outputPath)

	cmd := exec.CommandContext(ctx, magick, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("magick preview failed: %v: %s", err, string(output))
	}
	return nil
}

func rgba(hex string, alpha float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return fmt.Sprintf("rgba(0,0,0,%.2f)", alpha)
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)
	return fmt.Sprintf("rgba(%d,%d,%d,%.2f)", r, g, b, alpha)
}

func maxPreview(a, b int) int {
	if a > b {
		return a
	}
	return b
}
