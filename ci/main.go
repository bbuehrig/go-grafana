package main

import (
	"context"
	"fmt"
	"os"

	// Make sure to run: go get dagger.io/dagger
	dagger "dagger.io/dagger"
)

func main() {
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Mount source code
	src := client.Host().Directory(".")

	platforms := []struct {
		name     string
		platform string
	}{
		{"amd64", "linux/amd64"},
		{"arm64", "linux/arm64"},
	}

	for _, p := range platforms {
		imageTag := fmt.Sprintf("go-grafana-monitor:%s", p.name)
		image := client.Container(dagger.ContainerOpts{Platform: dagger.Platform(p.platform)}).Build(src)
		_, err := image.Export(ctx, imageTag)
		if err != nil {
			panic(fmt.Sprintf("Failed to export image for %s: %v", p.name, err))
		}
		fmt.Printf("Image for %s exported locally as %s\n", p.name, imageTag)
	}
}
