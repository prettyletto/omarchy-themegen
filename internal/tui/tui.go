package tui

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/prettyletto/omarchy-themegen/internal/export"
	"github.com/prettyletto/omarchy-themegen/internal/gen"
	"github.com/prettyletto/omarchy-themegen/internal/image"
	"github.com/prettyletto/omarchy-themegen/internal/preview"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
	"github.com/prettyletto/omarchy-themegen/internal/validate"
)

type step int

const (
	stepValidating step = iota
	stepError
	stepGeneration
	stepModeSelect
	stepComparison
	stepGroupSelect
	stepOverrideSelect
	stepNaming
	stepConfirmExport
	stepExporting
	stepResult
	stepApplyConfirm
	stepApplying
	stepDone
)

type genDoneMsg struct {
	candidates []gen.PaletteCandidate
	opts       *gen.GenerationOptions
	err        error
}

type exportDoneMsg struct {
	result        *export.ExportResult
	archiveResult *export.ArchiveResult
	err           error
}

type applyDoneMsg struct {
	err error
}

type composedPreviewDoneMsg struct {
	key  string
	path string
	err  error
}

type browserStateMsg struct {
	state *preview.SelectionState
}

type Model struct {
	step step

	imagePath string
	imgResult *image.Result
	genOpts   *gen.GenerationOptions

	candidates []gen.PaletteCandidate
	directions []theme.Direction
	selected   int

	composition    *theme.Composition
	mixMode        bool
	groupCursor    int
	overrideCursor int

	themeName    string
	normalized   string
	exportPath   string
	forceExport  bool
	archiveMode  bool
	reproducible bool
	livePreview  bool

	exportResult  *export.ExportResult
	archiveResult *export.ArchiveResult
	postResult    *validate.PostExportResult

	err     error
	message string

	textInput textinput.Model
	spinner   spinner.Model

	width    int
	height   int
	quitting bool

	lightMode bool
	seed      int

	termCaps             preview.Capability
	inlinePreviewEnabled bool
	previewCache         *preview.Cache
	previewDir           string
	directionPreviews    map[int]string
	composedPreviewPath  string
	composedPreviewKey   string
	composedPreviewBusy  bool
	previewMessage       string
	browserServer        *preview.BrowserServer
	browserURL           string
	program              *tea.Program
}

func Run(imagePath string, archiveMode bool) error {
	ti := textinput.New()
	ti.Placeholder = "Type a theme name..."
	ti.CharLimit = 64
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	previewDir, err := os.MkdirTemp("", "omarchy-themegen-preview-*")
	if err != nil {
		return fmt.Errorf("create preview cache: %w", err)
	}
	cache, err := preview.NewCache(previewDir)
	if err != nil {
		_ = os.RemoveAll(previewDir)
		return err
	}
	termCaps := preview.DetectCapability()

	m := &Model{
		step:                 stepValidating,
		imagePath:            imagePath,
		selected:             -1,
		textInput:            ti,
		spinner:              s,
		archiveMode:          archiveMode,
		termCaps:             termCaps,
		inlinePreviewEnabled: preview.InlineImageSupported(termCaps),
		previewCache:         cache,
		previewDir:           previewDir,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p
	_, err = p.Run()

	if m.browserServer != nil {
		m.browserServer.Stop()
	}
	if cache != nil {
		cache.Cleanup()
	} else if previewDir != "" {
		_ = os.RemoveAll(previewDir)
	}

	return err
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		runValidation(m.imagePath),
	)
}

func runValidation(path string) tea.Cmd {
	return func() tea.Msg {
		result := image.Validate(path)
		if !result.Valid {
			return genDoneMsg{err: fmt.Errorf("validation: %s", strings.Join(result.Errors, "; "))}
		}
		return genDoneMsg{}
	}
}

func runGeneration(path string, lightMode bool, seed int) tea.Cmd {
	return func() tea.Msg {
		opts, err := gen.NewGenerationOptions(path, seed, lightMode)
		if err != nil {
			return genDoneMsg{err: err}
		}
		colors, err := gen.ExtractDominantColors(path, 12)
		if err != nil {
			return genDoneMsg{err: fmt.Errorf("color extraction: %w", err)}
		}
		candidates, err := gen.GeneratePalettes(colors, opts)
		if err != nil {
			return genDoneMsg{err: fmt.Errorf("palette generation: %w", err)}
		}
		return genDoneMsg{candidates: candidates, opts: opts}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.drawInlinePreviewCmd()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		// In naming mode, forward printable keys to text input first
		if m.step == stepNaming && m.isTextInputKey(msg) {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
		return m.handleKey(msg)

	case genDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = stepError
			return m, nil
		}

		if m.step == stepValidating {
			m.imgResult = image.Validate(m.imagePath)
			m.step = stepGeneration
			return m, tea.Batch(m.spinner.Tick, runGeneration(m.imagePath, m.lightMode, m.seed))
		}

		m.candidates = msg.candidates
		m.genOpts = msg.opts
		m.directions = buildDirections(msg.candidates, m.genOpts)
		m.generateDirectionPreviewFiles()
		m.composition = buildComposition(m.directions)
		m.step = stepModeSelect
		if m.selected < 0 {
			m.selected = 0
		}
		return m, nil

	case exportDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = stepError
			return m, nil
		}
		m.exportResult = msg.result
		if msg.archiveResult != nil {
			m.archiveResult = msg.archiveResult
		}
		m.postResult = validate.PostExport(m.exportPath, m.normalized)
		m.step = stepResult
		return m, nil

	case applyDoneMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Apply failed: %v", msg.err)
		} else {
			m.message = "Theme applied successfully."
		}
		m.step = stepDone
		return m, nil

	case composedPreviewDoneMsg:
		if m.composition == nil {
			return m, nil
		}
		if msg.key != m.compositionPreviewKey() {
			return m, nil
		}
		if msg.err != nil {
			m.composedPreviewBusy = false
			m.previewMessage = fmt.Sprintf("Composed preview failed: %v", msg.err)
			return m, nil
		}
		m.composedPreviewBusy = false
		m.composedPreviewKey = msg.key
		m.composedPreviewPath = msg.path
		m.previewMessage = "Composed preview updated"
		return m, m.drawInlinePreviewCmd()

	case browserStateMsg:
		return m.applyBrowserState(msg.state)
	}

	return m, nil
}

func (m Model) isTextInputKey(msg tea.KeyMsg) bool {
	s := msg.String()
	switch s {
	case "enter", "esc", "ctrl+c", "up", "down", "tab":
		return false
	}
	return true
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Sequence(m.clearInlinePreviewCmd(), tea.Quit)

	case "esc":
		return m.handleEsc()

	case "q":
		if m.step == stepNaming {
			break
		}
		if m.step != stepModeSelect {
			m.quitting = true
			return m, tea.Sequence(m.clearInlinePreviewCmd(), tea.Quit)
		}

	case "enter":
		return m.handleEnter()

	case "up", "k":
		switch m.step {
		case stepModeSelect:
			m.mixMode = false
		case stepComparison:
			if m.selected > 0 {
				m.selected--
			}
			return m, m.drawInlinePreviewCmd()
		case stepGroupSelect:
			if m.groupCursor > 0 {
				m.groupCursor--
			}
		case stepOverrideSelect:
			if m.overrideCursor > 0 {
				m.overrideCursor--
			}
		}

	case "down", "j":
		switch m.step {
		case stepModeSelect:
			m.mixMode = true
		case stepComparison:
			if m.selected < len(m.directions)-1 {
				m.selected++
			}
			return m, m.drawInlinePreviewCmd()
		case stepGroupSelect:
			if m.groupCursor < len(theme.AllGroups)-1 {
				m.groupCursor++
			}
		case stepOverrideSelect:
			if m.overrideCursor < len(m.allOverrideSurfaces())-1 {
				m.overrideCursor++
			}
		}

	case "tab":
		if m.step == stepComparison && len(m.directions) > 0 {
			m.selected = (m.selected + 1) % len(m.directions)
			return m, m.drawInlinePreviewCmd()
		}

	case "i":
		if m.step == stepComparison || m.step == stepGroupSelect || m.step == stepOverrideSelect {
			if !preview.InlineImageSupported(m.termCaps) {
				m.message = "Terminal image preview is not supported here. Use browser preview instead."
				return m, nil
			}
			m.inlinePreviewEnabled = !m.inlinePreviewEnabled
			if m.inlinePreviewEnabled {
				m.message = "Inline image preview enabled."
				if m.step == stepGroupSelect || m.step == stepOverrideSelect {
					return m, m.requestComposedPreviewCmd()
				}
				return m, m.drawInlinePreviewCmd()
			}
			m.message = "Inline image preview disabled."
			return m, m.clearInlinePreviewCmd()
		}

	case "1", "2", "3", "4", "5":
		dirID := int(msg.Runes[0] - '0')
		if !theme.ValidDirectionID(dirID) || dirID > len(m.directions) {
			return m, nil
		}
		if m.step == stepGroupSelect {
			m.ensureComposition("component-mix")
			g := theme.AllGroups[m.groupCursor]
			m.composition.SetGroupSource(g.ID, dirID)
			m.message = ""
			return m, m.requestComposedPreviewCmd()
		}
		if m.step == stepOverrideSelect {
			m.ensureComposition("component-mix")
			surfaces := m.allOverrideSurfaces()
			if m.overrideCursor < len(surfaces) {
				surf := surfaces[m.overrideCursor]
				m.composition.SetOverride(surf, dirID)
				m.message = fmt.Sprintf("Override: %s → direction %d", surf, dirID)
			}
			return m, m.requestComposedPreviewCmd()
		}

	case "d":
		if m.step == stepOverrideSelect {
			m.ensureComposition("component-mix")
			surfaces := m.allOverrideSurfaces()
			if m.overrideCursor < len(surfaces) {
				surf := surfaces[m.overrideCursor]
				m.composition.ClearOverride(surf)
				m.message = fmt.Sprintf("Cleared override for %s", surf)
			}
			return m, m.requestComposedPreviewCmd()
		}

	case "a":
		if m.step == stepConfirmExport {
			m.archiveMode = !m.archiveMode
			return m, nil
		}

	case "l":
		if m.step == stepConfirmExport {
			m.livePreview = !m.livePreview
			return m, nil
		}

	case "p":
		if m.step == stepConfirmExport {
			m.reproducible = !m.reproducible
			if m.reproducible {
				m.message = "Reproducible archive will include source image bytes."
			} else {
				m.message = ""
			}
			return m, nil
		}

	case "r":
		if m.step == stepGroupSelect {
			m.ensureComposition("component-mix")
			// Reset all groups to direction 1
			for _, g := range theme.AllGroups {
				m.composition.SetGroupSource(g.ID, 1)
			}
			m.message = "All groups reset to direction 1"
			return m, m.requestComposedPreviewCmd()
		}

	case "w":
		if m.step == stepGroupSelect || m.step == stepOverrideSelect {
			m.composition = nil
			m.mixMode = false
			m.step = stepModeSelect
			m.message = ""
			return m, m.clearInlinePreviewCmd()
		}

	case "b":
		return m.handleBrowserToggle()
	}
	return m, nil
}

func (m Model) handleEsc() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepNaming:
		m.textInput.SetValue("")
		m.message = ""
		if m.composition != nil {
			// In component-mix, go back to override select
			m.step = stepOverrideSelect
			return m, m.drawInlinePreviewCmd()
		} else {
			m.step = stepComparison
			return m, m.drawInlinePreviewCmd()
		}
	case stepConfirmExport:
		m.step = stepNaming
	case stepApplyConfirm:
		m.step = stepResult
	case stepOverrideSelect:
		m.step = stepGroupSelect
		return m, m.drawInlinePreviewCmd()
	case stepGroupSelect:
		m.step = stepModeSelect
	default:
		m.quitting = true
		return m, tea.Sequence(m.clearInlinePreviewCmd(), tea.Quit)
	}
	if m.step == stepModeSelect {
		return m, m.clearInlinePreviewCmd()
	}
	return m, nil
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepError:
		m.quitting = true
		return m, tea.Quit

	case stepModeSelect:
		if m.mixMode {
			m.step = stepGroupSelect
			m.groupCursor = 0
			m.message = ""
			m.ensureComposition("component-mix")
			return m, m.requestComposedPreviewCmd()
		}
		m.composition = nil
		m.step = stepComparison
		m.selected = 0
		return m, m.drawInlinePreviewCmd()

	case stepComparison:
		if m.selected >= 0 && m.selected < len(m.directions) {
			m.step = stepNaming
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, tea.Sequence(m.clearInlinePreviewCmd(), textinput.Blink)
		}
		return m, nil

	case stepGroupSelect:
		// Validate all groups have direction
		for _, g := range theme.AllGroups {
			if _, ok := m.composition.GroupSources[g.ID]; !ok {
				m.message = fmt.Sprintf("Group %s has no direction assigned. Press a number key %s.", g.Label, theme.DirectionRangeLabel)
				return m, nil
			}
		}
		m.message = ""
		m.step = stepOverrideSelect
		return m, m.drawInlinePreviewCmd()

	case stepOverrideSelect:
		// Override selection is optional; proceed to naming
		m.message = ""
		m.step = stepNaming
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, tea.Sequence(m.clearInlinePreviewCmd(), textinput.Blink)

	case stepNaming:
		name := strings.TrimSpace(m.textInput.Value())
		if name == "" {
			m.message = "Theme name cannot be empty."
			return m, nil
		}
		m.themeName = name
		m.normalized = theme.NormalizeForExport(name)
		if m.normalized == "" {
			m.message = "Theme name normalizes to empty. Choose a different name."
			return m, nil
		}

		home, _ := os.UserHomeDir()
		m.exportPath = filepath.Join(home, ".config", "omarchy", "themes", m.normalized)

		if _, err := os.Stat(m.exportPath); err == nil {
			m.forceExport = true
		} else {
			m.forceExport = false
		}
		m.message = ""
		m.step = stepConfirmExport
		return m, nil

	case stepConfirmExport:
		return m.triggerExport()

	case stepResult:
		m.step = stepApplyConfirm
		return m, nil

	case stepApplyConfirm:
		return m.triggerApply()

	case stepDone:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) triggerExport() (tea.Model, tea.Cmd) {
	m.step = stepExporting

	if m.composition != nil {
		return m.triggerCompositionExport()
	}

	dir := m.directions[m.selected]
	tm, err := theme.NewThemeModelFromDirection(m.themeName, m.imagePath, m.imgResult, dir)
	if err != nil {
		return m, func() tea.Msg { return exportDoneMsg{err: err} }
	}
	tm.Version = "1.0.0"

	if err := validate.PreExport(tm); err != nil {
		return m, func() tea.Msg { return exportDoneMsg{err: err} }
	}

	archive := m.archiveMode
	livePreview := m.livePreview
	exportPath := m.exportPath
	normalizedName := m.normalized
	forceOverwrite := m.forceExport

	return m, func() tea.Msg {
		var result *export.ExportResult
		var err error
		if livePreview {
			result, err = export.ThemeDirectoryWithLivePreview(tm, exportPath, forceOverwrite)
		} else {
			result, err = export.ThemeDirectory(tm, exportPath, forceOverwrite)
		}
		msg := exportDoneMsg{result: result, err: err}
		if err == nil && archive {
			arcResult, aErr := export.CreateArchive(exportPath, normalizedName, normalizedName+".tar.gz")
			if aErr == nil {
				msg.archiveResult = arcResult
			}
		}
		return msg
	}
}

func (m *Model) triggerCompositionExport() (tea.Model, tea.Cmd) {
	tm, err := m.composition.Resolve(m.themeName, m.imagePath, m.imgResult)
	if err != nil {
		return m, func() tea.Msg { return exportDoneMsg{err: err} }
	}
	tm.Version = "1.0.0"

	if err := validate.PreExport(tm); err != nil {
		return m, func() tea.Msg { return exportDoneMsg{err: err} }
	}

	archive := m.archiveMode
	livePreview := m.livePreview
	exportPath := m.exportPath
	normalizedName := m.normalized
	forceOverwrite := m.forceExport

	return m, func() tea.Msg {
		var result *export.ExportResult
		var err error
		if livePreview {
			result, err = export.ThemeDirectoryWithLivePreview(tm, exportPath, forceOverwrite)
		} else {
			result, err = export.ThemeDirectory(tm, exportPath, forceOverwrite)
		}
		msg := exportDoneMsg{result: result, err: err}
		if err == nil && archive {
			arcResult, aErr := export.CreateArchive(exportPath, normalizedName, normalizedName+".tar.gz")
			if aErr == nil {
				msg.archiveResult = arcResult
			}
		}
		return msg
	}
}

func (m *Model) triggerApply() (tea.Model, tea.Cmd) {
	m.step = stepApplying

	return m, func() tea.Msg {
		if _, err := exec.LookPath("omarchy"); err != nil {
			return applyDoneMsg{err: fmt.Errorf("omarchy is not installed")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "omarchy", "theme", "set", m.normalized)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return applyDoneMsg{err: fmt.Errorf("%v: %s", err, string(output))}
		}
		return applyDoneMsg{}
	}
}

func buildDirections(candidates []gen.PaletteCandidate, opts *gen.GenerationOptions) []theme.Direction {
	var dirs []theme.Direction
	for _, c := range candidates {
		dirs = append(dirs, theme.Direction{
			ID:          c.ID,
			Label:       c.Label,
			Fingerprint: opts.Fingerprint,
			Colors:      c.Colors,
			Warnings:    c.Warnings,
			LightMode:   opts.LightMode,
		})
	}
	return dirs
}

func buildComposition(directions []theme.Direction) *theme.Composition {
	c := theme.NewComposition("whole-theme")
	c.Directions = directions
	return c
}

func (m Model) clearInlinePreviewCmd() tea.Cmd {
	seq := preview.ClearInlineImages(m.termCaps)
	if seq == "" {
		return nil
	}
	return func() tea.Msg {
		_, _ = os.Stdout.WriteString(seq)
		return nil
	}
}

func (m Model) drawInlinePreviewCmd() tea.Cmd {
	if !m.inlinePreviewEnabled {
		return nil
	}
	seq := m.inlinePreviewOutput(m.currentInlinePreviewPath())
	if seq == "" {
		return nil
	}
	return func() tea.Msg {
		// Let Bubble Tea finish painting the text buffer, then draw the terminal
		// image into the blank rows reserved by View().
		time.Sleep(25 * time.Millisecond)
		_, _ = os.Stdout.WriteString(seq)
		return nil
	}
}

func (m Model) currentInlinePreviewPath() string {
	switch m.step {
	case stepComparison:
		return m.selectedDirectionPreviewPath()
	case stepGroupSelect, stepOverrideSelect:
		return m.cachedComposedPreviewPath()
	default:
		return ""
	}
}

func (m *Model) ensureComposition(mode string) {
	if m.composition == nil {
		m.composition = buildComposition(m.directions)
	}
	m.composition.Mode = mode
	m.composition.Directions = m.directions
	if m.composition.GroupSources == nil {
		m.composition.GroupSources = make(map[string]int)
	}
	if m.composition.Overrides == nil {
		m.composition.Overrides = make(map[string]int)
	}
}

func (m *Model) generateDirectionPreviewFiles() {
	m.directionPreviews = make(map[int]string)
	m.previewMessage = ""
	if m.previewCache == nil || len(m.directions) == 0 {
		return
	}
	previewDir := filepath.Join(m.previewDir, "directions")
	paths, err := preview.GenerateDirectionPreviews(previewDir, m.imagePath, m.directions)
	if err != nil {
		m.previewMessage = fmt.Sprintf("Preview generation failed: %v", err)
		return
	}
	for i, path := range paths {
		if i < len(m.directions) {
			m.directionPreviews[m.directions[i].ID] = path
		}
	}
	m.previewMessage = fmt.Sprintf("Generated %d direction previews", len(paths))
}

func (m Model) cachedComposedPreviewPath() string {
	if m.composedPreviewPath == "" || m.composedPreviewKey != m.compositionPreviewKey() {
		return ""
	}
	return m.composedPreviewPath
}

func (m Model) shouldReserveComposedPreviewRows() bool {
	return m.composedPreviewBusy || m.cachedComposedPreviewPath() != ""
}

func (m *Model) requestComposedPreviewCmd() tea.Cmd {
	if m.composition == nil || len(m.directions) == 0 || m.imgResult == nil {
		return nil
	}
	if !m.allGroupsAssigned() {
		m.composedPreviewBusy = false
		m.composedPreviewPath = ""
		m.composedPreviewKey = ""
		m.previewMessage = "Assign all surface groups to render the component mix preview."
		return m.clearInlinePreviewCmd()
	}
	key := m.compositionPreviewKey()
	if m.composedPreviewPath != "" && m.composedPreviewKey == key {
		m.composedPreviewBusy = false
		return m.drawInlinePreviewCmd()
	}
	if m.composedPreviewBusy && m.composedPreviewKey == key {
		return nil
	}
	m.composedPreviewBusy = true
	m.composedPreviewKey = key
	m.composedPreviewPath = ""
	m.previewMessage = "Rendering component mix preview..."

	directions := append([]theme.Direction(nil), m.directions...)
	imagePath := m.imagePath
	imgResult := m.imgResult
	basePreviewDir := m.previewDir
	groups := make(map[string]int, len(m.composition.GroupSources))
	for gid, did := range m.composition.GroupSources {
		groups[gid] = did
	}
	overrides := make(map[string]int, len(m.composition.Overrides))
	for surface, did := range m.composition.Overrides {
		overrides[surface] = did
	}

	renderCmd := func() tea.Msg {
		comp := theme.NewComposition("component-mix")
		comp.Directions = directions
		for _, g := range theme.AllGroups {
			dirID := groups[g.ID]
			if err := comp.SetGroupSource(g.ID, dirID); err != nil {
				return composedPreviewDoneMsg{key: key, err: err}
			}
		}
		for surface, dirID := range overrides {
			if dirID > 0 {
				if err := comp.SetOverride(surface, dirID); err != nil {
					return composedPreviewDoneMsg{key: key, err: err}
				}
			}
		}
		tm, err := comp.Resolve("Preview", imagePath, imgResult)
		if err != nil {
			return composedPreviewDoneMsg{key: key, err: err}
		}
		previewDir := filepath.Join(basePreviewDir, "composed")
		if err := os.MkdirAll(previewDir, 0755); err != nil {
			return composedPreviewDoneMsg{key: key, err: err}
		}
		path := filepath.Join(previewDir, key+".png")
		if err := preview.GenerateComposedPreview(path, imagePath, tm); err != nil {
			return composedPreviewDoneMsg{key: key, err: err}
		}
		return composedPreviewDoneMsg{key: key, path: path}
	}
	return tea.Batch(m.clearInlinePreviewCmd(), renderCmd)
}

func (m *Model) compositionPreviewKey() string {
	if m.composition == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(m.composition.Mode)
	for _, g := range theme.AllGroups {
		b.WriteString(fmt.Sprintf("|g:%s=%d", g.ID, m.composition.GroupSources[g.ID]))
	}
	for _, surface := range m.allOverrideSurfaces() {
		b.WriteString(fmt.Sprintf("|o:%s=%d", surface, m.composition.Overrides[surface]))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("%x", sum)[:16]
}

func (m *Model) allGroupsAssigned() bool {
	if m.composition == nil {
		return false
	}
	for _, g := range theme.AllGroups {
		dirID, ok := m.composition.GroupSources[g.ID]
		if !ok || dirID == 0 {
			return false
		}
	}
	return true
}

func (m *Model) allOverrideSurfaces() []string {
	var surfaces []string
	for _, g := range theme.AllGroups {
		surfaces = append(surfaces, g.Surfaces...)
	}
	return surfaces
}

func (m *Model) reproducibleConfirmed() bool {
	return m.reproducible && m.forceExport
}

func (m *Model) handleBrowserToggle() (tea.Model, tea.Cmd) {
	if m.browserServer != nil {
		m.browserServer.Stop()
		m.browserServer = nil
		m.browserURL = ""
		m.message = "Browser preview stopped."
		return m, nil
	}

	if len(m.directions) == 0 {
		m.message = "No directions available. Wait for generation to complete."
		return m, nil
	}

	prog := m.program
	server := preview.NewBrowserServer(m.imagePath, m.directions, func(state *preview.SelectionState) {
		if prog != nil {
			prog.Send(browserStateMsg{state: state})
		}
	})
	url, err := server.Start()
	if err != nil {
		m.message = fmt.Sprintf("Failed to start browser preview: %v", err)
		return m, nil
	}

	m.browserServer = server
	m.browserURL = url
	m.message = fmt.Sprintf("Browser preview: %s", url)
	return m, nil
}

func (m *Model) applyBrowserState(state *preview.SelectionState) (tea.Model, tea.Cmd) {
	if state == nil {
		return m, nil
	}
	if state.Mode == "component-mix" {
		m.mixMode = true
		m.ensureComposition("component-mix")
		m.composition.GroupSources = make(map[string]int)
		m.composition.Overrides = make(map[string]int)
		m.step = stepGroupSelect
		for gid, did := range state.Groups {
			if did > 0 {
				if err := m.composition.SetGroupSource(gid, did); err != nil {
					m.message = fmt.Sprintf("Browser selection ignored: %v", err)
				}
			}
		}
		for surf, did := range state.Overrides {
			if did > 0 {
				if err := m.composition.SetOverride(surf, did); err != nil {
					m.message = fmt.Sprintf("Browser selection ignored: %v", err)
				}
			}
		}
		return m, m.requestComposedPreviewCmd()
	} else if state.Mode == "whole-theme" {
		m.composition = nil
		m.mixMode = false
		m.step = stepComparison
		m.selected = state.Selected - 1
		if m.selected < 0 {
			m.selected = 0
		}
		if len(m.directions) > 0 && m.selected >= len(m.directions) {
			m.selected = len(m.directions) - 1
		}
	}
	return m, m.drawInlinePreviewCmd()
}
