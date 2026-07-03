package theme

import (
	"fmt"
	"strings"

	"github.com/prettyletto/omarchy-themegen/internal/image"
)

const DirectionCount = 5

const DirectionRangeLabel = "1-5"

type SurfaceGroup struct {
	ID       string
	Label    string
	Surfaces []string
}

var (
	GroupDesktopShell = SurfaceGroup{
		ID:    "desktop-shell",
		Label: "Desktop Shell",
		Surfaces: []string{
			"waybar", "hyprland", "hyprlock", "mako",
			"walker", "swayosd", "hyprland-preview-share-picker",
		},
	}
	GroupTerminalsAndTUI = SurfaceGroup{
		ID:    "terminals-and-tui",
		Label: "Terminals And TUI",
		Surfaces: []string{
			"ghostty", "alacritty", "foot", "kitty", "btop", "terminal-palette",
		},
	}
	GroupEditor = SurfaceGroup{
		ID:       "editor",
		Label:    "Editor",
		Surfaces: []string{"neovim"},
	}
	GroupAssetsAndSystem = SurfaceGroup{
		ID:    "assets-and-system",
		Label: "Assets And System",
		Surfaces: []string{
			"wallpaper-background", "preview-assets", "icons",
			"light-mode", "chromium", "keyboard-rgb",
			"helix", "gum", "obsidian",
		},
	}

	AllGroups = []SurfaceGroup{
		GroupDesktopShell, GroupTerminalsAndTUI, GroupEditor, GroupAssetsAndSystem,
	}
)

func GroupByID(id string) (SurfaceGroup, bool) {
	id = strings.ToLower(id)
	for _, g := range AllGroups {
		if g.ID == id || strings.ToLower(g.Label) == id {
			return g, true
		}
	}
	return SurfaceGroup{}, false
}

func ValidSurface(name string) bool {
	name = strings.ToLower(name)
	for _, g := range AllGroups {
		for _, s := range g.Surfaces {
			if s == name {
				return true
			}
		}
	}
	return false
}

type Composition struct {
	Mode         string         // "whole-theme" or "component-mix"
	Directions   []Direction    // source directions
	GroupSources map[string]int // group ID -> direction ID
	Overrides    map[string]int // surface name -> direction ID
	Warnings     []string
}

func NewComposition(mode string) *Composition {
	return &Composition{
		Mode:         mode,
		GroupSources: make(map[string]int),
		Overrides:    make(map[string]int),
	}
}

func (c *Composition) SetGroupSource(groupID string, dirID int) error {
	if _, ok := GroupByID(groupID); !ok {
		return fmt.Errorf("unknown surface group: %s", groupID)
	}
	if !ValidDirectionID(dirID) {
		return fmt.Errorf("invalid direction %d (must be %s)", dirID, DirectionRangeLabel)
	}
	c.GroupSources[groupID] = dirID
	return nil
}

func (c *Composition) SetOverride(surfaceName string, dirID int) error {
	if !ValidSurface(surfaceName) {
		return fmt.Errorf("unknown or unsupported surface: %s", surfaceName)
	}
	if !ValidDirectionID(dirID) {
		return fmt.Errorf("invalid direction %d (must be %s)", dirID, DirectionRangeLabel)
	}
	c.Overrides[surfaceName] = dirID
	return nil
}

func ValidDirectionID(dirID int) bool {
	return dirID >= 1 && dirID <= DirectionCount
}

func (c *Composition) ClearOverride(surfaceName string) {
	delete(c.Overrides, surfaceName)
}

func (c *Composition) Validate() []string {
	var errs []string

	if c.Mode == "component-mix" {
		for _, g := range AllGroups {
			if _, ok := c.GroupSources[g.ID]; !ok {
				errs = append(errs, fmt.Sprintf("missing source direction for group: %s", g.Label))
			}
		}
	}

	for surface, dirID := range c.Overrides {
		if !ValidDirectionID(dirID) {
			errs = append(errs, fmt.Sprintf("override for %s has invalid direction %d", surface, dirID))
		}
		if !ValidSurface(surface) {
			errs = append(errs, fmt.Sprintf("override targets unsupported surface: %s", surface))
		}
	}

	return errs
}

func (c *Composition) Resolve(name, sourceImage string, imgResult *image.Result) (*ThemeModel, error) {
	if errs := c.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("composition validation failed: %s", strings.Join(errs, "; "))
	}

	if len(c.Directions) == 0 {
		return nil, fmt.Errorf("no directions available for composition")
	}

	if c.Mode == "whole-theme" {
		return c.resolveWholeTheme(name, sourceImage, imgResult)
	}

	return c.resolveComponentMix(name, sourceImage, imgResult)
}

func (c *Composition) resolveWholeTheme(name, sourceImage string, imgResult *image.Result) (*ThemeModel, error) {
	dir := c.Directions[0]

	normalized := normalizeThemeName(name)
	if normalized == "" {
		return nil, errNameEmpty
	}

	tm := &ThemeModel{
		Name:           name,
		NormalizedName: normalized,
		Version:        "0.2.0-dev",
		SourceImage:    sourceImage,
		ImageResult:    imgResult,
		Colors:         dir.Colors,
		Mode:           "whole-theme",
		DirectionID:    dir.ID,
		DirectionLabel: dir.Label,
		LightMode:      dir.LightMode,
	}

	// Record provenance: all groups use the same direction
	tm.GroupSelections = make(map[string]int)
	for _, g := range AllGroups {
		tm.GroupSelections[g.ID] = dir.ID
	}

	return tm, nil
}

func (c *Composition) resolveComponentMix(name, sourceImage string, imgResult *image.Result) (*ThemeModel, error) {
	normalized := normalizeThemeName(name)
	if normalized == "" {
		return nil, errNameEmpty
	}

	// Determine master direction from Assets And System group
	masterDirID := c.GroupSources[GroupAssetsAndSystem.ID]
	if masterDirID == 0 {
		for _, g := range AllGroups {
			if d, ok := c.GroupSources[g.ID]; ok {
				masterDirID = d
				break
			}
		}
	}
	if masterDirID < 1 || masterDirID > len(c.Directions) {
		return nil, fmt.Errorf("cannot determine master direction for component mix")
	}

	masterDir := c.Directions[masterDirID-1]

	// Start with master direction's colors
	merged := *masterDir.Colors

	// Apply group-specific color roles
	if dirID := c.GroupSources[GroupTerminalsAndTUI.ID]; dirID > 0 && dirID <= len(c.Directions) {
		termDir := c.Directions[dirID-1]
		merged.Cursor = termDir.Colors.Cursor
		for i, src := range []*string{
			&merged.Color0, &merged.Color1, &merged.Color2, &merged.Color3,
			&merged.Color4, &merged.Color5, &merged.Color6, &merged.Color7,
			&merged.Color8, &merged.Color9, &merged.Color10, &merged.Color11,
			&merged.Color12, &merged.Color13, &merged.Color14, &merged.Color15,
		} {
			dst := []*string{
				&termDir.Colors.Color0, &termDir.Colors.Color1, &termDir.Colors.Color2, &termDir.Colors.Color3,
				&termDir.Colors.Color4, &termDir.Colors.Color5, &termDir.Colors.Color6, &termDir.Colors.Color7,
				&termDir.Colors.Color8, &termDir.Colors.Color9, &termDir.Colors.Color10, &termDir.Colors.Color11,
				&termDir.Colors.Color12, &termDir.Colors.Color13, &termDir.Colors.Color14, &termDir.Colors.Color15,
			}
			*src = *dst[i]
		}
	}

	if dirID := c.GroupSources[GroupDesktopShell.ID]; dirID > 0 && dirID <= len(c.Directions) {
		dskDir := c.Directions[dirID-1]
		merged.Accent = dskDir.Colors.Accent
	}

	// Apply per-surface overrides (win over group)
	// neovim → affects accent (Editor gets full palette, accent is enough)
	for surface, dirID := range c.Overrides {
		if dirID > 0 && dirID <= len(c.Directions) && surface == "neovim" {
			ovDir := c.Directions[dirID-1]
			merged.Accent = ovDir.Colors.Accent
		}
	}

	label := "Mixed (" + masterDir.Label + " base)"

	tm := &ThemeModel{
		Name:           name,
		NormalizedName: normalized,
		Version:        "0.2.0-dev",
		SourceImage:    sourceImage,
		ImageResult:    imgResult,
		Colors:         &merged,
		Mode:           "component-mix",
		DirectionID:    masterDir.ID,
		DirectionLabel: label,
		LightMode:      masterDir.LightMode,
	}

	tm.GroupSelections = make(map[string]int)
	for k, v := range c.GroupSources {
		tm.GroupSelections[k] = v
	}
	tm.Overrides = make(map[string]int)
	for k, v := range c.Overrides {
		tm.Overrides[k] = v
	}
	tm.CompositionWarnings = append([]string{}, c.Warnings...)

	return tm, nil
}
