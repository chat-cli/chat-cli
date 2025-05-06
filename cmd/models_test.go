package cmd

import (
	"testing"
)

func TestModelsCmd(t *testing.T) {
	if modelsCmd.Use == "" {
		t.Error("Expected modelsCmd to have non-empty Use")
	}
}