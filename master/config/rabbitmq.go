package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type RabbitMQConfig struct {
	RabbitMQ struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Vhost    string `yaml:"vhost"`
	} `yaml:"rabbitmq"`
}

// 获取 RabbitMQ 配置的函数
func GetRabbitMQConfig() *RabbitMQConfig {
	config := &RabbitMQConfig{}
	if os.Getenv("env") == "test" {
		return config
	}
	yamlFile, err := os.ReadFile("config/yamls/rabbitmq.yaml")
	if err != nil {
		log.Fatalf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return config
}
