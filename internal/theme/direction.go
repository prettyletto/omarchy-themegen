package theme

import "github.com/anomalyco/omarchy-themegen/internal/image"

type Direction struct {
	ID           int
	Label        string
	Fingerprint  string
	Colors       *Colors
	Warnings     []string
	LightMode    bool
}

type DirectionSet struct {
	Directions []Direction
	ImageResult *image.Result
}

func NewThemeModelFromDirection(name, sourceImage string, imgResult *image.Result, dir Direction) (*ThemeModel, error) {
	if name == "" {
		return nil, errNameRequired
	}

	normalized := normalizeThemeName(name)
	if normalized == "" {
		return nil, errNameEmpty
	}

	if errs := ValidateColors(dir.Colors); len(errs) > 0 {
		return nil, errBadColors
	}

	return &ThemeModel{
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
	}, nil
}
