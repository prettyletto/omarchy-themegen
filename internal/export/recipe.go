package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/prettyletto/omarchy-themegen/internal/gen"
	"github.com/prettyletto/omarchy-themegen/internal/theme"
)

type Recipe struct {
	GeneratorVersion string         `json:"generator_version"`
	Fingerprint      string         `json:"fingerprint"`
	Seed             int            `json:"seed"`
	LightMode        bool           `json:"light_mode"`
	Mode             string         `json:"mode"`
	DirectionID      int            `json:"direction_id,omitempty"`
	DirectionLabel   string         `json:"direction_label,omitempty"`
	ThemeName        string         `json:"theme_name,omitempty"`
	GroupSelections  map[string]int `json:"group_selections,omitempty"`
	Overrides        map[string]int `json:"overrides,omitempty"`
}

func BuildRecipe(tm *theme.ThemeModel, opts *gen.GenerationOptions) *Recipe {
	r := &Recipe{
		GeneratorVersion: gen.GeneratorVersion,
		Fingerprint:      opts.Fingerprint,
		Seed:             opts.Seed,
		LightMode:        opts.LightMode,
		Mode:             tm.Mode,
		DirectionID:      tm.DirectionID,
		DirectionLabel:   tm.DirectionLabel,
	}

	if len(tm.GroupSelections) > 0 {
		r.GroupSelections = make(map[string]int)
		for k, v := range tm.GroupSelections {
			r.GroupSelections[k] = v
		}
	}
	if len(tm.Overrides) > 0 {
		r.Overrides = make(map[string]int)
		for k, v := range tm.Overrides {
			r.Overrides[k] = v
		}
	}

	return r
}

func (r *Recipe) IncludeThemeName(name string) {
	r.ThemeName = name
}

func WriteRecipe(recipe *Recipe, outputPath string) error {
	data, err := json.MarshalIndent(recipe, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal recipe: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("cannot create recipe directory: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write recipe: %w", err)
	}
	return nil
}

func LoadRecipe(path string) (*Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read recipe: %w", err)
	}
	var r Recipe
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("cannot parse recipe: %w", err)
	}
	if r.Fingerprint == "" {
		return nil, fmt.Errorf("recipe missing required fingerprint")
	}
	if r.GeneratorVersion == "" {
		return nil, fmt.Errorf("recipe missing required generator version")
	}
	if r.Mode == "" {
		r.Mode = "whole-theme"
	}
	return &r, nil
}
