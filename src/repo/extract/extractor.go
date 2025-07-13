package extract

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"
	"sigs.k8s.io/yaml"
)

const (
	tqJsonPath = "tq.json"
	tqYamlPath = "tq.yaml"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) ExtractConfig(repoDir string) (tqsdk.Space, error) {
	// Try tq.json first
	space, err := e.extractConfigFromFile(repoDir, tqJsonPath, true)
	if err == nil {
		return space, nil
	}

	// If tq.json not found, try tq.yaml
	if os.IsNotExist(err) {
		space, yamlErr := e.extractConfigFromFile(repoDir, tqYamlPath, false)
		if yamlErr == nil {
			return space, nil
		}
		// Return original JSON error if YAML also fails
		return tqsdk.Space{}, domain.ErrNoTqJsonFound
	}

	return tqsdk.Space{}, err
}

func (e *Extractor) extractConfigFromFile(repoDir, filename string, isJSON bool) (tqsdk.Space, error) {
	configFile := filepath.Join(repoDir, filename)

	data, err := os.ReadFile(configFile)
	if err != nil {
		return tqsdk.Space{}, err
	}

	var space tqsdk.Space
	if isJSON {
		if err := json.Unmarshal(data, &space); err != nil {
			return tqsdk.Space{}, domain.ErrTqIsNotValidJson
		}
	} else {
		// Use yaml.Unmarshal from sigs.k8s.io/yaml which handles both JSON and YAML
		if err := yaml.Unmarshal(data, &space); err != nil {
			return tqsdk.Space{}, domain.ErrTqIsNotValidJson
		}
	}

	if err := space.Validate(); err != nil {
		return space, err
	}
	return space, nil
}
