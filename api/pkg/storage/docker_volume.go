package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/plutov/formulosity/api/pkg/types"
)


func SaveFile(file *types.File) (string, error) {
    uploadDir := os.Getenv("UPLOADS_DIR") 

    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        return "", fmt.Errorf("failed to create upload directory: %v", err)
    }

    sanitizedFilename := filepath.Base(file.Name)

    filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizedFilename)
    fullPath := filepath.Join(uploadDir, filename)

    outFile, err := os.Create(fullPath)
    if err != nil {
        return "", fmt.Errorf("failed to create file: %v", err)
    }
    defer outFile.Close()

    if _, err := io.Copy(outFile, file.Data); err != nil {
        return "", fmt.Errorf("failed to copy file: %v", err)
    }

    return outFile.Name(), nil
}