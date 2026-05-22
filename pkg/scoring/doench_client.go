package scoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Doench Rule Set 2 (Azimuth) on-target scoring via Python HTTP server.
// Input: 30-mer sequences (4nt upstream + 20nt guide + NGG PAM + 3nt downstream)
// Output: scores 0-100 (higher = better on-target cutting efficiency)

var doenchServerURL = "http://127.0.0.1:8080"

var doenchClient = &http.Client{
	Timeout: 30 * time.Second,
}

func init() {
	if url := os.Getenv("DOENCH_SERVER_URL"); url != "" {
		doenchServerURL = url
	}
}

type doenchRequest struct {
	Sequences []string `json:"sequences"`
}

type doenchResponse struct {
	Scores map[string]float64 `json:"scores"`
	Error  string             `json:"error,omitempty"`
}

// DoenchScore sends 30-mer sequences to the Python scoring server and returns
// a map of sequence → score. Scores are 0-100 scale. -1 = invalid (contains N).
func DoenchScore(sequences []string) (map[string]float64, error) {
	if len(sequences) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(doenchRequest{Sequences: sequences})
	if err != nil {
		return nil, fmt.Errorf("doench: marshal request: %w", err)
	}

	resp, err := doenchClient.Post(doenchServerURL+"/score", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("doench: server unreachable at %s: %w", doenchServerURL, err)
	}
	defer resp.Body.Close()

	var result doenchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("doench: decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("doench: server error: %s", result.Error)
	}

	return result.Scores, nil
}

// DoenchHealthCheck verifies the Python scoring server is running.
func DoenchHealthCheck() error {
	resp, err := doenchClient.Get(doenchServerURL + "/health")
	if err != nil {
		return fmt.Errorf("doench server not reachable at %s: %w", doenchServerURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("doench server unhealthy: status %d", resp.StatusCode)
	}
	return nil
}
