package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"server-monitor/internal/logx"
	"server-monitor/internal/model"
	"time"
)

type Sender struct {
	client           *http.Client
	fallbackInterval int
}

func New(timeoutSeconds int, fallbackInterval int) *Sender {
	return &Sender{
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		fallbackInterval: fallbackInterval,
	}
}

func (s *Sender) SendMetrics(apiURL string, apiKey string, metrics model.Metrics) (int, error) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return s.fallbackInterval, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return s.fallbackInterval, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return s.fallbackInterval, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s.fallbackInterval, fmt.Errorf("server error: %s", resp.Status)
	}

	var apiResponse model.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return s.fallbackInterval, err
	}

	if apiResponse.Interval == 0 {
		logx.Warn("API response is missing interval, using fallback interval.")
		apiResponse.Interval = s.fallbackInterval
	}

	return apiResponse.Interval, nil
}
