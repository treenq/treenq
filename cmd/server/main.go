package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/api"
)

// CopyFile copies a file from src to dst. If dst does not exist, it will be created.
// If it exists, it will be truncated.
func CopyFile(src, dst string) error {
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

func ReadBuildConfig() (tqsdk.Resource, error) {
	f, err := os.Open("tq.json")
	if err != nil {
		return tqsdk.Resource{}, err
	}

	var res tqsdk.Resource
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		return tqsdk.Resource{}, err
	}

	return res, nil
}

func Deploy() error {
	// get sdk config folder
	if err := CopyFile("../../pkg/example/tq/ex.go", "../../pkg/builder/tq/build.go"); err != nil {
		return fmt.Errorf("failed to copy build file: %v", err)
	}

	// write config data
	cmd := exec.Command("go", "run", "../../pkg/builder/builder.go")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to exctract build config: %w", err)
	}

	// read config file
	conf, err := ReadBuildConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	connStringValue := ""
	if conf.Size > 0 {
		connStringValue = "postgres://some.com"
	}
	buildFlags := fmt.Sprintf("-X 'github.com/treenq/treenq/pkg/sdk.connStr=%s'", connStringValue)

	// build image passing the config values
	cmd = exec.Command("go", "build", "-ldflags", buildFlags, "../../pkg/example/main.go")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build target app: %w", err)
	}

	return nil
}

func main() {
	// if err := Deploy(); err != nil {
	// 	log.Fatalln(err)
	// }

	m := api.New()
	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Println(err)
	}
}
