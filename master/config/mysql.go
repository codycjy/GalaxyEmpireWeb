package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type MysqlConfig struct {
	DB struct {
		Host       string            `yaml:"host"`
		Port       string            `yaml:"port"`
		Username   string            `yaml:"username"`
		Password   string            `yaml:"password"`
		Database   string            `yaml:"database"`
		Parameters map[string]string `yaml:"parameters"`
	} `yaml:"db"`
}

func LoadConfig(filename string) (*MysqlConfig, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &MysqlConfig{}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetDSN(filename string) string {
	config, err := LoadConfig(filename)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	var params []string
	for key, value := range config.DB.Parameters {
		params = append(params, fmt.Sprintf("%s=%s", key, value))
	}
	paramStr := strings.Join(params, "&")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		config.DB.Username, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Database, paramStr)

	return dsn
}
