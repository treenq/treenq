package extract

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

// copyFile copies a file from src to dst. If dst does not exist, it will be created.
// If it exists, it will be truncated.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if err = dstFile.Sync(); err != nil {
		return err
	}

	return nil
}

// copyDirectory copies all files from srcDir to dstDir.
func copyDirectory(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err = copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err = copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

type Extractor struct {
	builderDir string
}

func NewExtractor(builderDir string) *Extractor {
	return &Extractor{builderDir: builderDir}
}

const tqRelativePath = "tq"
const tqBuildLauncherFile = "builder.go"

func (e *Extractor) ExtractConfig(repoDir string) (tqsdk.Resource, error) {
	repoConfigDir := filepath.Join(repoDir, tqRelativePath)
	targetDir := filepath.Join(e.builderDir, tqRelativePath)

	if err := os.MkdirAll(targetDir, 0766); err != nil {
		return tqsdk.Resource{}, fmt.Errorf("failed to create tq module dir: %w", err)
	}

	if err := copyDirectory(repoConfigDir, targetDir); err != nil {
		return tqsdk.Resource{}, fmt.Errorf("failed to copy build config: %w", err)
	}

	builderLauncherPath := filepath.Join(e.builderDir, tqBuildLauncherFile)
	defer func() {
		os.RemoveAll(targetDir)
	}()
	output, err := exec.Command("go", "run", builderLauncherPath).Output()
	if err != nil {
		return tqsdk.Resource{}, fmt.Errorf("failed to exctract build config: %w", err)
	}

	var res tqsdk.Resource
	if err := json.Unmarshal(output, &res); err != nil {
		return tqsdk.Resource{}, fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	return res, nil
}
