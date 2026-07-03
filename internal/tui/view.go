package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/omarchy-themegen/internal/gen"
	"github.com/prettyletto/omarchy-themegen/internal/omarchy"
	"github.com/prettyletto/omarchy-themegen/internal/preview"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).MarginBottom(1)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	accentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
)

const inlinePreviewRows = 14
const inlinePreviewStartRow = 4
const comparisonVisibleDirections = 3

func (m Model) View() string {
	switch m.step {
	case stepValidating:
		return m.viewValidating()
	case stepError:
		return m.viewError()
	case stepGeneration:
		return m.viewGeneration()
	case stepModeSelect:
		return m.viewModeSelect()
	case stepComparison:
		return m.viewComparison()
	case stepGroupSelect:
		return m.viewGroupSelect()
	case stepOverrideSelect:
		return m.viewOverrideSelect()
	case stepNaming:
		return m.viewNaming()
	case stepConfirmExport:
		return m.viewConfirmExport()
	case stepExporting:
		return m.viewExporting()
	case stepResult:
		return m.viewResult()
	case stepApplyConfirm:
		return m.viewApplyConfirm()
	case stepApplying:
		return m.viewApplying()
	case stepDone:
		return m.viewDone()
	default:
		return "Unknown state"
	}
}

func (m Model) viewValidating() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("omarchy-themegen"))
	b.WriteString("\n\n")
	b.WriteString(m.spinner.View() + " Validating source image...\n")
	b.WriteString(dimStyle.Render(m.imagePath))
	return b.String()
}

func (m Model) viewError() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("omarchy-themegen"))
	b.WriteString("\n\n")
	if m.err != nil {
		b.WriteString(errorStyle.Render("Error: " + m.err.Error()))
	} else {
		b.WriteString(errorStyle.Render("Error: unknown error"))
	}
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Press Enter to exit."))
	return b.String()
}

func (m Model) viewGeneration() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("omarchy-themegen"))
	b.WriteString("\n\n")
	b.WriteString(m.spinner.View() + " Generating theme directions...\n")
	return b.String()
}

func (m Model) viewComparison() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select a Direction"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("j/k to move • Tab to cycle • Enter to select • b browser • i hide/show image • q to quit"))
	b.WriteString("\n\n")
	if m.inlinePreviewEnabled {
		path := m.selectedDirectionPreviewPath()
		b.WriteString(m.inlinePreviewPlaceholder(path))
	}
	if m.previewMessage != "" {
		b.WriteString(dimStyle.Render(m.previewMessage))
		b.WriteString("\n")
	}
	if m.message != "" {
		b.WriteString(warnStyle.Render(m.message))
		b.WriteString("\n")
	}
	if m.previewMessage != "" || m.message != "" {
		b.WriteString("\n")
	}

	start, end := visibleDirectionRange(m.selected, len(m.directions), comparisonVisibleDirections)
	if start > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  ↑ %d more direction(s)", start)))
		b.WriteString("\n\n")
	}

	for i := start; i < end; i++ {
		d := m.directions[i]
		prefix := "  "
		label := fmt.Sprintf("Direction %d: %s", d.ID, d.Label)
		if d.LightMode {
			label += " (light)"
		}

		if i == m.selected {
			prefix = "▸ "
			b.WriteString(accentStyle.Render(prefix + label))
		} else {
			b.WriteString(dimStyle.Render(prefix + label))
		}
		b.WriteString("\n")

		colors := d.Colors
		b.WriteString(dimStyle.Render(fmt.Sprintf("     bg %s  fg %s  accent %s", colors.Background, colors.Foreground, colors.Accent)))
		b.WriteString("\n")
		swatches := []struct {
			name string
			hex  string
		}{
			{"bg", colors.Background},
			{"fg", colors.Foreground},
			{"ac", colors.Accent},
		}
		for _, sw := range swatches {
			rgb, err := gen.ParseHex(sw.hex)
			var swatchStyle lipgloss.Style
			bgColor := lipgloss.Color(sw.hex)
			if i == m.selected {
				swatchStyle = lipgloss.NewStyle().Background(bgColor).Padding(0, 1)
			} else {
				swatchStyle = lipgloss.NewStyle().Background(bgColor).Padding(0, 1).Faint(true)
			}
			var label string
			if err == nil {
				label = fmt.Sprintf(" %s L:%.0f ", sw.name, rgb.ToHSL().L*100)
			} else {
				label = fmt.Sprintf(" %s ?? ", sw.name)
			}
			b.WriteString("     " + swatchStyle.Render(label))
		}
		b.WriteString("\n")

		b.WriteString("     ")
		termHexes := []string{
			colors.Color0, colors.Color1, colors.Color2, colors.Color3,
			colors.Color4, colors.Color5, colors.Color6, colors.Color7,
		}
		for _, h := range termHexes {
			style := lipgloss.NewStyle().Background(lipgloss.Color(h)).Padding(0, 1)
			b.WriteString(style.Render(" "))
		}
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("     0 %s  1 %s  2 %s  3 %s", colors.Color0, colors.Color1, colors.Color2, colors.Color3)))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("     4 %s  5 %s  6 %s  7 %s", colors.Color4, colors.Color5, colors.Color6, colors.Color7)))
		b.WriteString("\n")

		if len(d.Warnings) > 0 && i == m.selected {
			for _, w := range d.Warnings {
				b.WriteString(warnStyle.Render("     ⚠ " + w))
				b.WriteString("\n")
			}
		}
		if i == m.selected && m.directionPreviews != nil {
			if path := m.directionPreviews[d.ID]; path != "" {
				b.WriteString(dimStyle.Render("     Preview: " + path))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}
	if end < len(m.directions) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  ↓ %d more direction(s)", len(m.directions)-end)))
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render(fmt.Sprintf("\nShowing %d-%d of %d • Direction %d selected", start+1, end, len(m.directions), m.selected+1)))
	if m.browserURL != "" {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Browser: " + m.browserURL))
	}

	return b.String()
}

func visibleDirectionRange(selected, total, visible int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if visible <= 0 || visible >= total {
		return 0, total
	}
	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}
	start := selected - visible/2
	if start < 0 {
		start = 0
	}
	if start+visible > total {
		start = total - visible
	}
	return start, start + visible
}

func (m Model) viewNaming() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Name Your Theme"))
	b.WriteString("\n")

	if m.composition != nil {
		b.WriteString(dimStyle.Render("Mode: component-mix"))
	} else if m.selected >= 0 && m.selected < len(m.directions) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("Direction: %d (%s)", m.directions[m.selected].ID, m.directions[m.selected].Label)))
	}
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(errorStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	b.WriteString("Theme name: ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.textInput.Value() != "" {
		normalized := theme.NormalizeForExport(strings.TrimSpace(m.textInput.Value()))
		if normalized != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf("Export name: %s", normalized)))
			b.WriteString("\n")
			b.WriteString(dimStyle.Render(fmt.Sprintf("Path: ~/.config/omarchy/themes/%s", normalized)))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Enter to confirm • Esc to go back"))

	return b.String()
}

func (m Model) viewConfirmExport() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Confirm Export"))
	b.WriteString("\n\n")

	if m.composition != nil {
		b.WriteString("Mode:       component-mix\n")
		for gid, did := range m.composition.GroupSources {
			b.WriteString(fmt.Sprintf("  %s → direction %d\n", gid, did))
		}
		if len(m.composition.Overrides) > 0 {
			b.WriteString("Overrides:\n")
			for surf, did := range m.composition.Overrides {
				b.WriteString(fmt.Sprintf("  %s → direction %d\n", surf, did))
			}
		}
	} else if m.selected >= 0 && m.selected < len(m.directions) {
		b.WriteString(fmt.Sprintf("Direction:  %d (%s)\n", m.directions[m.selected].ID, m.directions[m.selected].Label))
	}
	b.WriteString(fmt.Sprintf("Theme name: %s\n", m.themeName))
	b.WriteString(fmt.Sprintf("Export as:  %s\n", m.normalized))
	b.WriteString(fmt.Sprintf("Path:       %s\n", m.exportPath))

	if m.forceExport {
		b.WriteString(warnStyle.Render("\n⚠ Target already exists. A timestamped backup will be created."))
	}

	b.WriteString("\n\nArtifacts:\n")
	if m.archiveMode {
		b.WriteString("  [x] Finished archive (a to toggle)\n")
	} else {
		b.WriteString("  [ ] Finished archive (a to toggle)\n")
	}
	if m.reproducible {
		b.WriteString(warnStyle.Render("  [x] Reproducible archive — includes source image (p to toggle)\n"))
	} else {
		b.WriteString("  [ ] Reproducible archive (p to toggle)\n")
	}
	if m.livePreview {
		b.WriteString(warnStyle.Render("  [x] Live Hyprland screenshot preview — applies theme first (l to toggle)\n"))
	} else {
		b.WriteString("  [ ] Live Hyprland screenshot preview — applies theme first (l to toggle)\n")
	}
	if m.reproducible && !m.reproducibleConfirmed() {
		b.WriteString(warnStyle.Render("  ⚠ Requires confirmation\n"))
	}

	b.WriteString("\n")
	b.WriteString("Export is separate from apply.\n")
	b.WriteString(dimStyle.Render("\nEnter to confirm export • Esc to go back"))

	return b.String()
}

func (m Model) viewExporting() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Exporting"))
	b.WriteString("\n\n")
	b.WriteString(m.spinner.View() + " Writing theme files...\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("Target: %s", m.exportPath)))
	return b.String()
}

func (m Model) viewResult() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Export Complete"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Theme:    %s\n", m.normalized))
	b.WriteString(fmt.Sprintf("Path:     %s\n", m.exportPath))
	if m.exportResult != nil && m.exportResult.BackupPath != "" {
		b.WriteString(fmt.Sprintf("Backup:   %s\n", m.exportResult.BackupPath))
	}
	if m.archiveResult != nil && m.archiveResult.Path != "" {
		b.WriteString(fmt.Sprintf("Archive:  %s\n", m.archiveResult.Path))
	}
	if m.exportResult != nil {
		for _, w := range m.exportResult.Warnings {
			b.WriteString(warnStyle.Render("Warning:  " + w))
			b.WriteString("\n")
		}
	}

	if m.postResult != nil {
		if m.postResult.OmarchyInstalled {
			b.WriteString(successStyle.Render("Omarchy:  detected"))
		} else {
			b.WriteString(warnStyle.Render("Omarchy:  not detected (reduced confidence)"))
		}
		b.WriteString("\n")
		for _, w := range m.postResult.Warnings {
			b.WriteString(warnStyle.Render("Warning:  " + w))
			b.WriteString("\n")
		}
		if !m.postResult.Passed {
			for _, e := range m.postResult.Errors {
				b.WriteString(errorStyle.Render("Error:    " + e))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Enter to continue to apply • Esc to quit"))

	return b.String()
}

func (m Model) viewApplyConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Apply Theme"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Apply %s with Omarchy?\n", m.normalized))
	b.WriteString("\n")
	b.WriteString(warnStyle.Render("This will restart themed components and may trigger hooks.\n"))

	disc := omarchy.Discover()
	if !disc.Installed {
		b.WriteString(warnStyle.Render("\nOmarchy is not installed. Apply is unavailable.\n"))
		for _, diag := range disc.Diagnostics {
			b.WriteString(dimStyle.Render("  " + diag + "\n"))
		}
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Esc to go back"))
		return b.String()
	}

	b.WriteString(fmt.Sprintf("\nOmarchy:    %s (confidence: %s)\n", disc.BinaryPath, disc.Confidence()))
	b.WriteString(dimStyle.Render("\nEnter to confirm apply • Esc to go back"))

	return b.String()
}

func (m Model) viewApplying() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Applying"))
	b.WriteString("\n\n")
	b.WriteString(m.spinner.View() + " Running omarchy theme set...\n")
	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Done"))
	b.WriteString("\n\n")
	if m.message != "" {
		b.WriteString(successStyle.Render(m.message))
	} else {
		b.WriteString(successStyle.Render("Theme exported to " + m.exportPath))
	}
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Press Enter or Esc to exit."))
	return b.String()
}

func (m Model) viewModeSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select Mode"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("j/k to move • Enter to confirm • b for browser preview • q to quit"))
	b.WriteString("\n\n")

	if m.mixMode {
		b.WriteString(dimStyle.Render("  Whole Theme\n"))
		b.WriteString(accentStyle.Render("▸ Component Mix"))
		b.WriteString("\n")
		b.WriteString("  Choose a different direction per surface group.\n")
	} else {
		b.WriteString(accentStyle.Render("▸ Whole Theme"))
		b.WriteString("\n")
		b.WriteString("  Select one complete direction for all surfaces.\n")
		b.WriteString(dimStyle.Render("  Component Mix\n"))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Recommended: Whole Theme for a consistent look."))
	b.WriteString(m.viewDirectionPaletteLegend())
	return b.String()
}

func (m Model) viewGroupSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Assign Surface Groups"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("j/k to move • 1/2/3 assign • r reset • b browser • i hide/show image • Enter done • w back"))
	b.WriteString("\n\n")

	if m.inlinePreviewEnabled {
		path := m.currentComposedPreview()
		b.WriteString(m.inlinePreviewPlaceholder(path))
	}
	if m.message != "" {
		b.WriteString(warnStyle.Render(m.message))
		b.WriteString("\n\n")
	}
	b.WriteString(m.viewDirectionPaletteLegend())
	if len(m.directions) > 0 {
		b.WriteString("\n")
	}

	for i, g := range theme.AllGroups {
		prefix := "  "
		if i == m.groupCursor {
			prefix = accentStyle.Render("▸ ")
		}

		dirID, assigned := m.composition.GroupSources[g.ID]
		if assigned {
			b.WriteString(fmt.Sprintf("%s%s → Direction %d\n", prefix, g.Label, dirID))
		} else {
			b.WriteString(fmt.Sprintf("%s%s → (unassigned)\n", prefix, g.Label))
		}

		// Show surfaces in group
		b.WriteString(dimStyle.Render(fmt.Sprintf("     %s", strings.Join(g.Surfaces, ", "))))
		b.WriteString("\n\n")
	}

	return b.String()
}

func (m Model) viewDirectionPaletteLegend() string {
	if len(m.directions) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Generated Directions:"))
	b.WriteString("\n")
	for _, d := range m.directions {
		colors := d.Colors
		b.WriteString(fmt.Sprintf("  Direction %d: %s  ", d.ID, d.Label))
		for _, h := range []string{colors.Background, colors.Foreground, colors.Accent, colors.Color1, colors.Color2, colors.Color3, colors.Color4, colors.Color5, colors.Color6} {
			b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(h)).Padding(0, 1).Render(" "))
		}
		b.WriteString(dimStyle.Render(fmt.Sprintf("  bg %s  fg %s  ac %s", colors.Background, colors.Foreground, colors.Accent)))
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) viewOverrideSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Per-Surface Overrides"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("j/k to move • 1/2/3 set • d clear • b browser • i hide/show image • Enter skip • w back"))
	b.WriteString("\n\n")
	if m.inlinePreviewEnabled {
		path := m.currentComposedPreview()
		b.WriteString(m.inlinePreviewPlaceholder(path))
	}
	if m.message != "" {
		b.WriteString(warnStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	surfaces := m.allOverrideSurfaces()
	for i, s := range surfaces {
		prefix := "  "
		if i == m.overrideCursor {
			prefix = accentStyle.Render("▸ ")
		}
		if ov, ok := m.composition.Overrides[s]; ok {
			b.WriteString(fmt.Sprintf("%s%s → Direction %d (overridden)\n", prefix, s, ov))
		} else {
			b.WriteString(dimStyle.Render(fmt.Sprintf("%s%s (group default)\n", prefix, s)))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Changes apply to the final Theme Model on export."))
	return b.String()
}

func (m Model) selectedDirectionPreviewPath() string {
	if m.selected < 0 || m.selected >= len(m.directions) || m.directionPreviews == nil {
		return ""
	}
	return m.directionPreviews[m.directions[m.selected].ID]
}

func (m Model) inlinePreviewPlaceholder(path string) string {
	if path == "" || !preview.InlineImageSupported(m.termCaps) {
		return ""
	}
	return strings.Repeat("\n", inlinePreviewRows)
}

func (m Model) inlinePreviewOutput(path string) string {
	if path == "" || !preview.InlineImageSupported(m.termCaps) {
		return ""
	}
	cols := m.width - 4
	if cols > 72 {
		cols = 72
	}
	if cols < 40 {
		cols = 40
	}
	img := preview.InlineImage(path, m.termCaps, cols, inlinePreviewRows)
	if img == "" {
		return ""
	}
	return fmt.Sprintf("\x1b[s%s\x1b[%d;1H%s\x1b[u", preview.ClearInlineImages(m.termCaps), inlinePreviewStartRow, img)
}
