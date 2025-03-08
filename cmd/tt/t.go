package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
)

func main() {
	if buildah.InitReexec() {
		return
	}
	unshare.MaybeReexecUsingUserNamespace(false)
	buildStoreOptions, err := storage.DefaultStoreOptions()
	ifErr(err)
	buildStore, err := storage.GetStore(buildStoreOptions)
	ifErr(err)

	id, ref, err := imagebuildah.BuildDockerfiles(context.Background(), buildStore, define.BuildOptions{
		Registry:     "jopa",
		Output:       "jopa2",
		Out:          os.Stdout,
		Err:          os.Stderr,
		ReportWriter: os.Stdout,
		IgnoreFile:   "./.dockerignore",
	}, "./Dockerfile")
	ifErr(err)
	fmt.Println(id)
	fmt.Println(ref)
}

func ifErr(err error) {
	if err != nil {
		panic(err)
	}
}
