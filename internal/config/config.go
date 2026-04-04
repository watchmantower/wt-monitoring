package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ServerID                          string
	APIKey                            string
	APIURL                            string
	FallbackIntervalSeconds           int
	HTTPTimeoutSeconds                int
	MaxProcesses                      int
	ProcessCollectionIntervalMultiple int
}

const DefaultFallbackIntervalSeconds = 10
const DefaultHTTPTimeoutSeconds = 5
const DefaultMaxProcesses = 20
const DefaultProcessCollectionIntervalMultiple = 5

func Parse() Config {
	serverID := flag.String("server_id", "", "Server ID")
	apiKey := flag.String("api_key", "", "API Key")
	apiURL := flag.String("api_url", "", "WT API URL")
	fallbackIntervalSeconds := flag.Int("fallback_interval", DefaultFallbackIntervalSeconds, "Fallback interval in seconds when the API response is unavailable")
	httpTimeoutSeconds := flag.Int("http_timeout", DefaultHTTPTimeoutSeconds, "HTTP request timeout in seconds")
	maxProcesses := flag.Int("max_processes", DefaultMaxProcesses, "Maximum number of processes to include in metrics")
	processCollectionIntervalMultiple := flag.Int("process_interval_multiple", DefaultProcessCollectionIntervalMultiple, "Collect process metrics once every N loops")
	flag.Parse()

	return Config{
		ServerID:                          *serverID,
		APIKey:                            *apiKey,
		APIURL:                            *apiURL,
		FallbackIntervalSeconds:           *fallbackIntervalSeconds,
		HTTPTimeoutSeconds:                *httpTimeoutSeconds,
		MaxProcesses:                      *maxProcesses,
		ProcessCollectionIntervalMultiple: *processCollectionIntervalMultiple,
	}
}

func (c Config) Validate() error {
	if c.ServerID == "" || c.APIKey == "" {
		return fmt.Errorf("server_id and api_key are required")
	}

	if c.APIURL == "" {
		return fmt.Errorf("api_url is required")
	}

	if c.FallbackIntervalSeconds <= 0 {
		return fmt.Errorf("fallback_interval must be greater than 0")
	}

	if c.HTTPTimeoutSeconds <= 0 {
		return fmt.Errorf("http_timeout must be greater than 0")
	}

	if c.MaxProcesses < 0 {
		return fmt.Errorf("max_processes cannot be negative")
	}

	if c.ProcessCollectionIntervalMultiple <= 0 {
		return fmt.Errorf("process_interval_multiple must be greater than 0")
	}

	return nil
}
