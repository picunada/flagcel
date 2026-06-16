package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/open-feature/go-sdk/openfeature"
	flagcel "github.com/picunada/flagcel/sdks/go"
)

func main() {
	endpoint := env("FLAGCEL_ENDPOINT", "http://localhost:8080/api/v1")
	apiKey := os.Getenv("FLAGCEL_API_KEY")
	flagKey := env("FLAGCEL_FLAG_KEY", "new-checkout")

	provider, err := flagcel.NewProvider(
		endpoint,
		apiKey,
		flagcel.WithPollInterval(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Shutdown()

	if err := openfeature.SetProviderAndWait(provider); err != nil {
		log.Fatal(err)
	}

	client := openfeature.NewClient("flagcel-local-example")
	evalCtx := openfeature.NewEvaluationContext(
		env("TARGETING_KEY", "u_123"),
		map[string]any{
			"user": map[string]any{
				"id":      env("TARGETING_KEY", "u_123"),
				"country": env("USER_COUNTRY", "US"),
			},
			"request": map[string]any{
				"path": env("REQUEST_PATH", "/checkout"),
			},
		},
	)

	details, err := client.BooleanValueDetails(
		context.Background(),
		flagKey,
		false,
		evalCtx,
	)
	if err != nil {
		log.Printf("evaluation used default: %v", err)
	}

	fmt.Printf("flag=%s value=%t reason=%s variant=%s\n",
		flagKey,
		details.Value,
		details.Reason,
		details.Variant,
	)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
