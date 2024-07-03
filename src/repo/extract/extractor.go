package extract

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/google/uuid"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

//go:embed template.txt
var emptyTqTemplate []byte

type Extractor struct {
	builderDirPrefix string
	builderPackage   string
	tpl              *template.Template
}

func NewExtractor(builderDirPrefix string, builderPackage string) *Extractor {
	tpl := template.Must(template.New("builder").Parse(string(emptyTqTemplate)))
	return &Extractor{builderDirPrefix: builderDirPrefix, builderPackage: builderPackage, tpl: tpl}
}

const tqRelativePath = "tq"
const tqBuildLauncherFile = "builder.go"

func (e *Extractor) Open() (string, error) {
	id := uuid.NewString()
	if err := e.createBuilder(id); err != nil {
		return "", fmt.Errorf("failed to open new builder: %w", err)
	}
	return id, nil
}

func (e *Extractor) Close(id string) error {
	builderDir := e.getBuilderPath(id)
	if err := removeBuilder(builderDir); err != nil {
		return fmt.Errorf("failed to close builder: %w", err)
	}

	return nil
}

func (e *Extractor) ExtractConfig(id string, repoDir string) (tqsdk.App, error) {
	builderDir := e.getBuilderPath(id)
	repoConfigDir := filepath.Join(repoDir, tqRelativePath)
	targetDir := filepath.Join(builderDir, tqRelativePath)

	if err := os.MkdirAll(targetDir, 0766); err != nil {
		return tqsdk.App{}, fmt.Errorf("failed to create tq module dir: %w", err)
	}

	if err := copyDirectory(repoConfigDir, targetDir); err != nil {
		return tqsdk.App{}, fmt.Errorf("failed to copy build config: %w", err)
	}
	defer func() {
		os.RemoveAll(targetDir)
	}()

	builderLauncherPath := filepath.Join(builderDir, tqBuildLauncherFile)
	output, err := exec.Command("go", "run", builderLauncherPath).Output()
	if err != nil {
		return tqsdk.App{}, fmt.Errorf("failed to exctract build config: %w", err)
	}

	var res tqsdk.App
	if err := json.Unmarshal(output, &res); err != nil {
		return tqsdk.App{}, fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	return res, nil
}

func (e *Extractor) getBuilderPath(id string) string {
	return filepath.Join(e.builderDirPrefix, id)
}

func (e *Extractor) createBuilder(id string) error {
	builderDir := e.getBuilderPath(id)

	if err := os.MkdirAll(builderDir, 0766); err != nil {
		return fmt.Errorf("failed to create builder dir: %w", err)
	}

	f, err := os.Create(filepath.Join(builderDir, tqBuildLauncherFile))
	if err != nil {
		return fmt.Errorf("failed to create builder file: %w", err)
	}
	defer f.Close()

	if err := e.tpl.Execute(f, map[string]string{"ID": id, "Package": e.builderPackage}); err != nil {
		return fmt.Errorf("failed to write builder template: %w", err)
	}

	return nil
}

func removeBuilder(dst string) error {
	err := os.RemoveAll(dst)
	if err != nil {
		return fmt.Errorf("failed to remove builder dir: %w", err)
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
