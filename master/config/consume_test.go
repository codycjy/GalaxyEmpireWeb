package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLoadConsumes(t *testing.T) {
	// Create a temporary test config file
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "test_consume.yaml")
	testConfig := `consumes:
  extend_account:
    amount: 1000
    available: true
    description: "Test Extend Account"
    duration_days: 31`

	err := os.WriteFile(testConfigPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset the singleton instance
	consumes = nil
	consumeOnce = sync.Once{}

	// Test loading the config
	err = LoadConsumes(testConfigPath)
	if err != nil {
		t.Fatalf("LoadConsumes failed: %v", err)
	}

	// Test getting a consume config
	consume, exists := GetConsume(ExtendAccount)
	if !exists {
		t.Error("Expected consume config to exist")
	}

	// Verify the loaded values
	if consume.Amount != 1000 {
		t.Errorf("Expected amount 1000, got %d", consume.Amount)
	}
	if consume.Description != "Test Extend Account" {
		t.Errorf("Expected description 'Test Extend Account', got %s", consume.Description)
	}
	if consume.DurationDays != 31 {
		t.Errorf("Expected duration_days 31, got %d", consume.DurationDays)
	}
	if !consume.Available {
		t.Error("Expected consume to be available")
	}
}

func TestLoadConsumesInvalidPath(t *testing.T) {
	// Reset the singleton instance
	consumes = nil
	consumeOnce = sync.Once{}

	err := LoadConsumes("nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error when loading from invalid path")
	}
}

func TestLoadConsumesInvalidYAML(t *testing.T) {
	// Create a temporary test config file with invalid YAML
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "invalid_consume.yaml")
	invalidConfig := `consumes:
  extend_account:
    amount: "invalid" # Should be number
    available: true`

	err := os.WriteFile(testConfigPath, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset the singleton instance
	consumes = nil
	consumeOnce = sync.Once{}

	err = LoadConsumes(testConfigPath)
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}
