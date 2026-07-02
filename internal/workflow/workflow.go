package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prettyletto/omarchy-themegen/internal/export"
	"github.com/prettyletto/omarchy-themegen/internal/gen"
	"github.com/prettyletto/omarchy-themegen/internal/image"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
	"github.com/prettyletto/omarchy-themegen/internal/validate"
)

type Options struct {
	ImagePath    string
	ThemeName    string
	OutputDir    string
	Seed         int
	LightMode    bool
	DirectionID  int // 1-3 for whole-theme
	Mode         string
	GroupSources map[string]int
	Overrides    map[string]int
	Overwrite    bool
	Archive      bool
	ArchivePath  string
	ArchiveOnly  bool
	RecipePath   string
	Reproducible bool
	LivePreview  bool
}

type Result struct {
	ThemeName        string
	NormalizedName   string
	Mode             string
	DirectionID      int
	DirectionLabel   string
	LightMode        bool
	ExportPath       string
	BackupPath       string
	ArchivePath      string
	RecipePath       string
	ReproduciblePath string
	OmarchyDetected  bool
	Warnings         []string
	Errors           []string
	GroupSelections  map[string]int
	Overrides        map[string]int
	Success          bool
}

func Run(opts Options) *Result {
	r := &Result{Success: true, Mode: opts.Mode}

	// Validate image
	imgResult := image.Validate(opts.ImagePath)
	if !imgResult.Valid {
		r.Errors = append(r.Errors, imgResult.Errors...)
		r.Success = false
		return r
	}
	r.Warnings = append(r.Warnings, imgResult.Warnings...)

	// Generate directions
	genOpts, err := gen.NewGenerationOptions(opts.ImagePath, opts.Seed, opts.LightMode)
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
		r.Success = false
		return r
	}

	colors, err := gen.ExtractDominantColors(opts.ImagePath, 12)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("color extraction: %v", err))
		r.Success = false
		return r
	}

	candidates, err := gen.GeneratePalettes(colors, genOpts)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("palette generation: %v", err))
		r.Success = false
		return r
	}

	var dirs []theme.Direction
	for _, c := range candidates {
		dirs = append(dirs, theme.Direction{
			ID: c.ID, Label: c.Label, Fingerprint: genOpts.Fingerprint,
			Colors: c.Colors, Warnings: c.Warnings, LightMode: genOpts.LightMode,
		})
		for _, w := range c.Warnings {
			r.Warnings = append(r.Warnings, fmt.Sprintf("direction %d: %s", c.ID, w))
		}
	}

	// Resolve composition
	var tm *theme.ThemeModel
	if opts.Mode == "component-mix" {
		comp := theme.NewComposition("component-mix")
		comp.Directions = dirs
		for gid, did := range opts.GroupSources {
			if err := comp.SetGroupSource(gid, did); err != nil {
				r.Errors = append(r.Errors, err.Error())
				r.Success = false
				return r
			}
		}
		for surf, did := range opts.Overrides {
			if err := comp.SetOverride(surf, did); err != nil {
				r.Errors = append(r.Errors, err.Error())
				r.Success = false
				return r
			}
		}
		tm, err = comp.Resolve(opts.ThemeName, opts.ImagePath, imgResult)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Sprintf("composition: %v", err))
			r.Success = false
			return r
		}
		r.GroupSelections = tm.GroupSelections
		r.Overrides = tm.Overrides
	} else {
		if opts.DirectionID < 1 || opts.DirectionID > len(dirs) {
			r.Errors = append(r.Errors, fmt.Sprintf("invalid direction %d", opts.DirectionID))
			r.Success = false
			return r
		}
		d := dirs[opts.DirectionID-1]
		tm, err = theme.NewThemeModelFromDirection(opts.ThemeName, opts.ImagePath, imgResult, d)
		if err != nil {
			r.Errors = append(r.Errors, err.Error())
			r.Success = false
			return r
		}
	}

	tm.Version = "1.0.0"
	r.ThemeName = tm.Name
	r.NormalizedName = tm.NormalizedName
	r.DirectionID = tm.DirectionID
	r.DirectionLabel = tm.DirectionLabel
	r.LightMode = tm.LightMode

	// Pre-export validation
	if err := validate.PreExport(tm); err != nil {
		r.Errors = append(r.Errors, err.Error())
		r.Success = false
		return r
	}

	// Determine export directory
	exportDir := opts.OutputDir
	if exportDir == "" {
		home, _ := os.UserHomeDir()
		exportDir = filepath.Join(home, ".config", "omarchy", "themes", tm.NormalizedName)
	}

	// Archive-only mode: export to temp dir, archive, then cleanup
	if opts.ArchiveOnly {
		tmpDir, err := os.MkdirTemp("", "omarchy-themegen-")
		if err != nil {
			r.Errors = append(r.Errors, err.Error())
			r.Success = false
			return r
		}
		exportDir = filepath.Join(tmpDir, tm.NormalizedName)
		defer os.RemoveAll(tmpDir)
	}

	// Export
	var exportResult *export.ExportResult
	if opts.LivePreview {
		exportResult, err = export.ThemeDirectoryWithLivePreview(tm, exportDir, opts.Overwrite)
	} else {
		exportResult, err = export.ThemeDirectory(tm, exportDir, opts.Overwrite)
	}
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("export: %v", err))
		if _, statErr := os.Stat(exportDir); statErr == nil {
			r.Warnings = append(r.Warnings, fmt.Sprintf("partial directory may exist at %s", exportDir))
		}
		r.Success = false
		return r
	}
	r.ExportPath = exportResult.Path
	r.BackupPath = exportResult.BackupPath
	r.Warnings = append(r.Warnings, exportResult.Warnings...)

	// Post-export validation
	postResult := validate.PostExport(exportDir, tm.NormalizedName)
	r.OmarchyDetected = postResult.OmarchyInstalled
	r.Warnings = append(r.Warnings, postResult.Warnings...)
	if !postResult.Passed {
		r.Errors = append(r.Errors, postResult.Errors...)
		r.Success = false
	}

	// Archive
	if opts.Archive {
		arcPath := opts.ArchivePath
		if arcPath == "" {
			arcPath = tm.NormalizedName + ".tar.gz"
		}
		arcResult, err := export.CreateArchive(exportDir, tm.NormalizedName, arcPath)
		if err != nil {
			r.Warnings = append(r.Warnings, fmt.Sprintf("archive: %v", err))
		} else {
			r.ArchivePath = arcResult.Path
		}
	}

	// Recipe
	if opts.RecipePath != "" {
		recipe := export.BuildRecipe(tm, genOpts)
		rPath := opts.RecipePath
		if err := export.WriteRecipe(recipe, rPath); err != nil {
			r.Warnings = append(r.Warnings, fmt.Sprintf("recipe: %v", err))
		} else {
			r.RecipePath = rPath
		}
	}

	// Reproducible archive
	if opts.Reproducible && opts.Overwrite {
		repPath := tm.NormalizedName + ".reproducible.tar.gz"
		if err := createReproducibleArchive(exportDir, tm, genOpts, repPath); err != nil {
			r.Warnings = append(r.Warnings, fmt.Sprintf("reproducible archive: %v", err))
		} else {
			r.ReproduciblePath = repPath
		}
	}

	return r
}

func createReproducibleArchive(exportDir string, tm *theme.ThemeModel, opts *gen.GenerationOptions, archivePath string) error {
	recipe := export.BuildRecipe(tm, opts)
	tmpDir, err := os.MkdirTemp("", "reproducible-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	recipePath := filepath.Join(tmpDir, "recipe.json")
	if err := export.WriteRecipe(recipe, recipePath); err != nil {
		return err
	}

	srcData, err := os.ReadFile(tm.SourceImage)
	if err != nil {
		return err
	}
	srcName := filepath.Base(tm.SourceImage)
	if err := os.WriteFile(filepath.Join(tmpDir, srcName), srcData, 0644); err != nil {
		return err
	}

	return export.CreateArchiveWithExtras(exportDir, tm.NormalizedName, archivePath, []export.ExtraFile{
		{Name: "recipe.json", Path: recipePath},
		{Name: srcName, Path: filepath.Join(tmpDir, srcName)},
	})
}
