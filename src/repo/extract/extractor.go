package extract

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/treenq/treenq/src/domain"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

const tqRelativePath = "tq.json"

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) ExtractConfig(repoDir string) (tqsdk.Space, error) {
	configFile := filepath.Join(repoDir, tqRelativePath)

	data, err := os.ReadFile(configFile)
	if os.IsNotExist(err) {
		return tqsdk.Space{}, domain.ErrNoConfigFileFound
	}
	if err != nil {
		return tqsdk.Space{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var space tqsdk.Space
	if err := json.Unmarshal(data, &space); err != nil {
		return tqsdk.Space{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return space, nil
}
