package export

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ArchiveResult struct {
	Path string
}

func CreateArchive(themeDir, normalizedName, archivePath string) (*ArchiveResult, error) {
	// Resolve archive path
	arcPath := archivePath
	if !filepath.IsAbs(arcPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("cannot get current directory: %w", err)
		}
		arcPath = filepath.Join(cwd, arcPath)
	}

	f, err := os.Create(arcPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create archive: %w", err)
	}

	gzw := gzip.NewWriter(f)
	tw := tar.NewWriter(gzw)

	prefix := normalizedName + "/"

	err = filepath.Walk(themeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == themeDir {
			return nil
		}

		relPath, err := filepath.Rel(themeDir, path)
		if err != nil {
			return fmt.Errorf("cannot compute relative path: %w", err)
		}

		archiveName := prefix + relPath

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("cannot create tar header: %w", err)
		}
		header.Name = archiveName

		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("cannot read symlink: %w", err)
			}
			header.Linkname = link
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("cannot write tar header: %w", err)
		}

		if info.Mode().IsRegular() {
			src, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("cannot open file for archive: %w", err)
			}
			if _, err := io.Copy(tw, src); err != nil {
				src.Close()
				return fmt.Errorf("cannot write file to archive: %w", err)
			}
			src.Close()
		}

		return nil
	})

	if err != nil {
		tw.Close()
		gzw.Close()
		f.Close()
		os.Remove(arcPath)
		return nil, fmt.Errorf("archive creation failed: %w", err)
	}

	if err := tw.Close(); err != nil {
		gzw.Close()
		f.Close()
		os.Remove(arcPath)
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzw.Close(); err != nil {
		f.Close()
		os.Remove(arcPath)
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(arcPath)
		return nil, fmt.Errorf("failed to close archive file: %w", err)
	}

	return &ArchiveResult{Path: arcPath}, nil
}
