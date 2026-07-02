package preview

import (
	"os"
	"os/exec"
	"strings"
)

type Capability int

const (
	CapNone Capability = iota
	CapKitty
	CapITerm2
	CapSixel
)

func (c Capability) String() string {
	switch c {
	case CapNone:
		return "none"
	case CapKitty:
		return "kitty"
	case CapITerm2:
		return "iterm2"
	case CapSixel:
		return "sixel"
	default:
		return "unknown"
	}
}

func (c Capability) Supported() bool {
	return c != CapNone
}

func DetectCapability() Capability {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Kitty protocol
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return CapKitty
	}

	// WezTerm supports kitty protocol
	if strings.Contains(strings.ToLower(termProgram), "wezterm") {
		return CapKitty
	}

	// Ghostty supports kitty protocol
	if strings.Contains(strings.ToLower(termProgram), "ghostty") {
		return CapKitty
	}

	// iTerm2
	if strings.Contains(strings.ToLower(termProgram), "iterm") || strings.Contains(strings.ToLower(termProgram), "apple_terminal") {
		return CapITerm2
	}

	// Sixel support (foot, xterm with sixel, etc.)
	if strings.Contains(strings.ToLower(term), "foot") || strings.Contains(strings.ToLower(term), "xterm") {
		// Check if sixel is actually enabled
		return CapSixel
	}

	// Check for imgcat (iTerm2 utility)
	if _, err := exec.LookPath("imgcat"); err == nil {
		return CapITerm2
	}

	return CapNone
}

type DisplayResult struct {
	Cap      Capability
	Fallback string
	Message  string
}

func DisplayCapability() DisplayResult {
	c := DetectCapability()
	r := DisplayResult{Cap: c}

	switch c {
	case CapNone:
		r.Fallback = "ANSI color swatches only"
		r.Message = "Terminal image preview not supported. Using text/ANSI fallback."
	case CapKitty:
		r.Message = "Kitty image protocol detected."
	case CapITerm2:
		r.Message = "iTerm2 image protocol detected."
	case CapSixel:
		r.Message = "Sixel graphics detected."
	}

	return r
}
