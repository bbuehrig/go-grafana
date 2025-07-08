package main

import (
	"context"
	"fmt"
	"os"

	// Make sure to run: go get dagger.io/dagger
	dagger "dagger.io/dagger"
)

func RunPipeline(ctx context.Context, client *dagger.Client) error {
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
		container := client.Container(dagger.ContainerOpts{Platform: dagger.Platform(p.platform)}).
			From("golang:1.24-alpine").
			WithMountedDirectory("/app", src).
			WithWorkdir("/app")

		// Run tests before build
		out, err := container.WithExec([]string{"go", "test", "-v", "./..."}).Stdout(ctx)
		if err != nil {
			return fmt.Errorf("Tests failed for %s: %v\n%s", p.name, err, out)
		}
		fmt.Printf("Tests passed for %s:\n%s\n", p.name, out)

		// Build binary
		buildContainer := container.WithExec([]string{"go", "build", "-o", "go-grafana", "."})
		final := client.Container(dagger.ContainerOpts{Platform: dagger.Platform(p.platform)}).
			From("gcr.io/distroless/base").
			WithWorkdir("/app").
			WithFile("/app/go-grafana", buildContainer.File("/app/go-grafana"))

		_, err = final.Export(ctx, imageTag)
		if err != nil {
			return fmt.Errorf("Failed to export image for %s: %v", p.name, err)
		}
		fmt.Printf("Image for %s exported locally as %s\n", p.name, imageTag)
	}
	return nil
}

func main() {
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if err := RunPipeline(ctx, client); err != nil {
		panic(err)
	}
}
