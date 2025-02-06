package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

type ConsumeConfig struct {
	Amount       int64  `yaml:"amount"`
	Available    bool   `yaml:"available"`
	Description  string `yaml:"description"`
	DurationDays int    `yaml:"duration_days"`
}

type ConsumeType string

const (
	ExtendAccount ConsumeType = "extend_account"
)

type consumeConfiguration struct {
	Consumes map[ConsumeType]ConsumeConfig `yaml:"consumes"`
}

var (
	consumes    *consumeConfiguration
	consumeOnce sync.Once
)

func LoadConsumes(configPath string) error {
	if configPath == "" {
		configPath = filepath.Join("config", "yaml", "consume.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config consumeConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	consumeOnce.Do(func() {
		consumes = &config
	})

	return nil
}

func initConsume() {
	consumeOnce.Do(func() {
		if err := LoadConsumes(""); err != nil {
			log.Fatalf("Failed to load consume config: %v", err)
		}
	})
}

func GetConsume(consumeType ConsumeType) (ConsumeConfig, bool) {
	initConsume()
	consume, exists := consumes.Consumes[consumeType]
	return consume, exists
}
