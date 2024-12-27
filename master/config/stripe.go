package config

import (
	"log"
	"os"
	"sync"
)

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

var (
	stripeConfig *StripeConfig
	stripeOnce   sync.Once
)

func GetStripeConfig() *StripeConfig {
	stripeOnce.Do(func() {
		stripeConfig = &StripeConfig{
			SecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
			WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		}

		// For development, you can get these from stripe cli:
		// stripe listen --forward-to localhost:8080/api/v1/payment/webhook
		if os.Getenv("ENV") == "development" && stripeConfig.WebhookSecret == "" {
			log.Printf("[StripeConfig] Running in development mode without webhook secret")
		}
	})
	return stripeConfig
}
