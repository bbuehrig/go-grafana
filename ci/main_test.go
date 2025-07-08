package main

import (
	"context"
	"os"
	"testing"

	dagger "dagger.io/dagger"
)

func TestRunPipeline(t *testing.T) {
	if os.Getenv("CI_DAGGER_TEST") == "" {
		t.Skip("Skipping Dagger pipeline test; set CI_DAGGER_TEST=1 to run.")
	}
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Fatalf("Failed to connect to Dagger: %v", err)
	}
	defer client.Close()

	if err := RunPipeline(ctx, client); err != nil {
		t.Errorf("Pipeline failed: %v", err)
	}
}
