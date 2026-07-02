package preview

import (
	"encoding/base64"
	"fmt"
	"os"
)

func InlineImage(path string, cap Capability, cols, rows int) string {
	if path == "" || !InlineImageSupported(cap) {
		return ""
	}
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 20
	}
	switch cap {
	case CapKitty:
		encodedPath := base64.StdEncoding.EncodeToString([]byte(path))
		return fmt.Sprintf("\x1b_Ga=T,f=100,t=f,c=%d,r=%d;%s\x1b\\", cols, rows, encodedPath)
	case CapITerm2:
		data, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		return fmt.Sprintf("\x1b]1337;File=inline=1;width=%dch;height=%d:%s\a", cols, rows, encoded)
	default:
		return ""
	}
}

func InlineImageSupported(cap Capability) bool {
	return cap == CapKitty || cap == CapITerm2
}

func ClearInlineImages(cap Capability) string {
	switch cap {
	case CapKitty:
		return "\x1b_Ga=d,d=A\x1b\\"
	default:
		return ""
	}
}
