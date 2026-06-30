package export

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func CreateBackup(dirPath string) (string, error) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return "", nil
	}

	parent := filepath.Dir(dirPath)
	base := filepath.Base(dirPath)

	for attempt := 0; attempt < 10; attempt++ {
		var timestamp string
		if attempt == 0 {
			timestamp = time.Now().UTC().Format("20060102T150405Z")
		} else {
			timestamp = time.Now().UTC().Format("20060102T150405.000000Z")
			time.Sleep(time.Millisecond)
		}
		backupName := fmt.Sprintf("%s.backup-%s", base, timestamp)
		backupPath := filepath.Join(parent, backupName)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			if err := os.Rename(dirPath, backupPath); err != nil {
				return "", fmt.Errorf("cannot rename %s to %s: %w", dirPath, backupPath, err)
			}
			return backupPath, nil
		}
	}

	return "", fmt.Errorf("failed to create unique backup for %s after 10 attempts", dirPath)
}
