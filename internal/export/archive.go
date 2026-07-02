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

type ExtraFile struct {
	Name string
	Path string
}

func CreateArchive(themeDir, normalizedName, archivePath string) (*ArchiveResult, error) {
	arcPath, err := resolveArchivePath(archivePath)
	if err != nil {
		return nil, err
	}

	if err := writeTarGz(arcPath, normalizedName, func(tw *tar.Writer, prefix string) error {
		return walkThemeDir(tw, themeDir, prefix)
	}); err != nil {
		return nil, err
	}

	return &ArchiveResult{Path: arcPath}, nil
}

func CreateArchiveWithExtras(themeDir, normalizedName, archivePath string, extras []ExtraFile) error {
	arcPath, err := resolveArchivePath(archivePath)
	if err != nil {
		return err
	}

	return writeTarGz(arcPath, normalizedName, func(tw *tar.Writer, prefix string) error {
		if err := walkThemeDir(tw, themeDir, prefix); err != nil {
			return err
		}
		for _, ef := range extras {
			if err := addFileToTar(tw, prefix+ef.Name, ef.Path); err != nil {
				return err
			}
		}
		return nil
	})
}

func resolveArchivePath(archivePath string) (string, error) {
	if filepath.IsAbs(archivePath) {
		return archivePath, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current directory: %w", err)
	}
	return filepath.Join(cwd, archivePath), nil
}

func writeTarGz(arcPath, prefix string, writeFn func(*tar.Writer, string) error) error {
	f, err := os.Create(arcPath)
	if err != nil {
		return fmt.Errorf("cannot create archive: %w", err)
	}

	gzw := gzip.NewWriter(f)
	tw := tar.NewWriter(gzw)
	pfx := prefix + "/"

	err = writeFn(tw, pfx)
	if err != nil {
		tw.Close()
		gzw.Close()
		f.Close()
		os.Remove(arcPath)
		return fmt.Errorf("archive creation failed: %w", err)
	}

	if err := tw.Close(); err != nil {
		gzw.Close()
		f.Close()
		os.Remove(arcPath)
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzw.Close(); err != nil {
		f.Close()
		os.Remove(arcPath)
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(arcPath)
		return fmt.Errorf("failed to close archive file: %w", err)
	}

	return nil
}

func walkThemeDir(tw *tar.Writer, themeDir, prefix string) error {
	return filepath.Walk(themeDir, func(path string, info os.FileInfo, err error) error {
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
		return addFileToTar(tw, prefix+relPath, path)
	})
}

func addFileToTar(tw *tar.Writer, archiveName, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot stat file for archive: %w", err)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("cannot create tar header: %w", err)
	}
	header.Name = archiveName

	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(filePath)
		if err != nil {
			return fmt.Errorf("cannot read symlink: %w", err)
		}
		header.Linkname = link
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("cannot write tar header: %w", err)
	}

	if info.Mode().IsRegular() {
		src, err := os.Open(filePath)
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
}
