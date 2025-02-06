package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLoadPrices(t *testing.T) {
	// Create a temporary test config file
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "test_price.yaml")
	testConfig := `prices:
  deposit_10:
    amount: 1000
    stripe_id: "price_test_10"
    available: true
    description: "Test $10 Deposit"
  deposit_30:
    amount: 3000
    stripe_id: "price_test_30"
    available: true
    description: "Test $30 Deposit"`

	err := os.WriteFile(testConfigPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset the singleton instance
	prices = nil
	pricesOnce = sync.Once{}

	// Test loading the config
	err = LoadPrices(testConfigPath)
	if err != nil {
		t.Fatalf("LoadPrices failed: %v", err)
	}

	// Test GetPrice
	price10, exists := GetPrice(Deposit10USD)
	if !exists {
		t.Error("Expected price config to exist")
	}
	if price10.Amount != 1000 {
		t.Errorf("Expected amount 1000, got %d", price10.Amount)
	}
	if price10.StripeID != "price_test_10" {
		t.Errorf("Expected stripe_id 'price_test_10', got %s", price10.StripeID)
	}

	// Test GetPriceByAmount
	price30, exists := GetPriceByAmount(3000)
	if !exists {
		t.Error("Expected to find price by amount 3000")
	}
	if price30.StripeID != "price_test_30" {
		t.Errorf("Expected stripe_id 'price_test_30', got %s", price30.StripeID)
	}

	// Test GetStripePrice
	stripeID, exists := GetStripePrice(1000)
	if !exists {
		t.Error("Expected to find Stripe price ID")
	}
	if stripeID != "price_test_10" {
		t.Errorf("Expected stripe_id 'price_test_10', got %s", stripeID)
	}

	// Test GetAvailablePrices
	availablePrices := GetAvailablePrices()
	if len(availablePrices) != 2 {
		t.Errorf("Expected 2 available prices, got %d", len(availablePrices))
	}

	// Test GetPrices
	simplePrices := GetPrices()
	if len(simplePrices) != 2 {
		t.Errorf("Expected 2 simple prices, got %d", len(simplePrices))
	}
}

func TestLoadPricesInvalidPath(t *testing.T) {
	// Reset the singleton instance
	prices = nil
	pricesOnce = sync.Once{}

	err := LoadPrices("nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error when loading from invalid path")
	}
}

func TestLoadPricesInvalidYAML(t *testing.T) {
	// Create a temporary test config file with invalid YAML
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "invalid_price.yaml")
	invalidConfig := `prices:
  deposit_10:
    amount: "invalid" # Should be number
    available: true`

	err := os.WriteFile(testConfigPath, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset the singleton instance
	prices = nil
	pricesOnce = sync.Once{}

	err = LoadPrices(testConfigPath)
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}

func TestGetNonExistentPrice(t *testing.T) {
	// Create a temporary test config file
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "test_price.yaml")
	testConfig := `prices:
  deposit_10:
    amount: 1000
    stripe_id: "price_test_10"
    available: true
    description: "Test $10 Deposit"`

	err := os.WriteFile(testConfigPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset the singleton instance
	prices = nil
	pricesOnce = sync.Once{}

	// Load the config
	err = LoadPrices(testConfigPath)
	if err != nil {
		t.Fatalf("LoadPrices failed: %v", err)
	}

	// Test getting non-existent price
	_, exists := GetPrice("non_existent_price")
	if exists {
		t.Error("Expected non-existent price to return exists=false")
	}

	// Test getting non-existent amount
	_, exists = GetPriceByAmount(9999)
	if exists {
		t.Error("Expected non-existent amount to return exists=false")
	}

	// Test getting non-existent Stripe price
	_, exists = GetStripePrice(9999)
	if exists {
		t.Error("Expected non-existent Stripe price to return exists=false")
	}
}
