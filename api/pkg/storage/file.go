package storage

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/plutov/formulosity/api/pkg/types"
)

type File struct {
	Logger    *slog.Logger
	uploadDir string
}

func (p *File) Init() error {
	p.uploadDir = os.Getenv("UPLOADS_DIR")
	if p.uploadDir == "" {
		p.uploadDir = "./uploads"
	}

	if err := os.MkdirAll(p.uploadDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create upload directory: %v", err)
	}

	return nil
}

func (p *File) SaveFile(file *types.File) (string, error) {
	sanitizedFilename, err := sanitizeFilename(file.Name)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizedFilename)
	fullPath := filepath.Join(p.uploadDir, filename)

	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			p.Logger.Error("unable to close file", "err", err)
		}
	}()

	dataSize, err := io.Copy(outFile, file.Data)
	if err != nil {
		if rmErr := os.Remove(fullPath); rmErr != nil {
			return "", rmErr
		}
		return "", err
	}

	if dataSize == 0 {
		if rmErr := os.Remove(fullPath); rmErr != nil {
			return "", rmErr
		}
		return "", fmt.Errorf("file content is empty")
	}

	return outFile.Name(), nil
}

func sanitizeFilename(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("filename is empty")
	}

	p := bluemonday.NewPolicy()

	name = filepath.Base(name)
	sanitized := p.Sanitize(name)

	if regexp.MustCompile(`[^\w\-.]`).MatchString(sanitized) {
		return "", fmt.Errorf("invalid characters in filename")
	}
	sanitized = regexp.MustCompile(`[^\w\-.]`).ReplaceAllString(sanitized, "_")

	return sanitized, nil
}

func (p *File) IsFileExist(fileName string) (bool, string, error) {
	fullPath := filepath.Join(p.uploadDir, fileName)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, fullPath, nil
}
