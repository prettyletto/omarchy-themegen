package gen

import (
	"crypto/sha256"
	"fmt"
	"os"
)

const GeneratorVersion = "0.2.0-dev"

type GenerationOptions struct {
	SourcePath    string
	Fingerprint   string
	Seed          int
	LightMode     bool
	GeneratorVer  string
}

func NewGenerationOptions(sourcePath string, seed int, lightMode bool) (*GenerationOptions, error) {
	fp, err := fingerprintFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("cannot fingerprint source image: %w", err)
	}

	return &GenerationOptions{
		SourcePath:   sourcePath,
		Fingerprint:  fp,
		Seed:         seed,
		LightMode:    lightMode,
		GeneratorVer: GeneratorVersion,
	}, nil
}

func fingerprintFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", h), nil
}
