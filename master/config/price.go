package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

// Price represents a basic price structure
type Price struct {
	StripeID string
	Amount   int64
}

// PriceConfig represents the full price configuration
type PriceConfig struct {
	Amount      int64  `yaml:"amount"`
	StripeID    string `yaml:"stripe_id"`
	Available   bool   `yaml:"available"`
	Description string `yaml:"description"`
}

type PriceType string

const (
	Deposit10USD PriceType = "deposit_10"
	Deposit30USD PriceType = "deposit_30"
	Deposit50USD PriceType = "deposit_50"
	ExtendPrice  PriceType = "extend_account"
)

type priceConfiguration struct {
	Prices map[PriceType]PriceConfig `yaml:"prices"`
}

var (
	prices     *priceConfiguration
	pricesOnce sync.Once
)

// LoadPrices loads the price configuration from the YAML file
func LoadPrices(configPath string) error {
	if configPath == "" {
		configPath = filepath.Join("config", "yaml", "price.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config priceConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	pricesOnce.Do(func() {
		prices = &config
	})

	return nil
}

func initPrice() {
	pricesOnce.Do(func() {
		if err := LoadPrices(""); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	})
}

// GetPrice returns the price configuration for a given price type
func GetPrice(priceType PriceType) (PriceConfig, bool) {
	initPrice()
	price, exists := prices.Prices[priceType]
	return price, exists
}

// GetPriceByAmount returns the price configuration for a given amount
func GetPriceByAmount(amount int64) (PriceConfig, bool) {
	initPrice()

	for _, price := range prices.Prices {
		if price.Amount == amount && price.Available {
			return price, true
		}
	}
	return PriceConfig{}, false
}

// GetStripePrice returns the Stripe price ID for a given amount
func GetStripePrice(amount int64) (string, bool) {
	price, exists := GetPriceByAmount(amount)
	if !exists || price.StripeID == "" {
		return "", false
	}
	return price.StripeID, true
}

// GetAvailablePrices returns all available prices
func GetAvailablePrices() []PriceConfig {
	if prices == nil {
		LoadPrices("")
	}

	var availablePrices []PriceConfig
	for _, price := range prices.Prices {
		if price.Available {
			availablePrices = append(availablePrices, price)
		}
	}
	return availablePrices
}

// GetPrices returns all available prices in the simplified Price format
func GetPrices() []Price {
	if prices == nil {
		LoadPrices("")
	}

	var result []Price
	for _, p := range prices.Prices {
		if p.Available {
			result = append(result, Price{
				StripeID: p.StripeID,
				Amount:   p.Amount,
			})
		}
	}
	return result
}
