package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/user/truckllm/internal/config"
)

type HTTPStore struct {
	endpoint string
	apiKey   string
}

func NewHTTPStore(endpoint, apiKey string) *HTTPStore {
	return &HTTPStore{endpoint: endpoint, apiKey: apiKey}
}

func (h *HTTPStore) GetUsage(provider, keyID, metric string) (float64, error) {
	url := fmt.Sprintf("%s/usage/%s/%s/%s", h.endpoint, provider, keyID, metric)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+h.apiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(string(body), 64)
}

func (h *HTTPStore) SetUsage(provider, keyID, metric string, value float64) error {
	url := fmt.Sprintf("%s/usage/%s/%s/%s", h.endpoint, provider, keyID, metric)
	data := map[string]float64{"value": value}
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+h.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func (h *HTTPStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	url := fmt.Sprintf("%s/usage/%s/%s/%s/increment", h.endpoint, provider, keyID, metric)
	data := map[string]float64{"delta": delta}
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+h.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func (h *HTTPStore) LoadConfig() (*config.Config, error) {
	url := fmt.Sprintf("%s/config", h.endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+h.apiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	err = json.Unmarshal(body, &cfg)
	return &cfg, err
}

func (h *HTTPStore) SaveConfig(cfg *config.Config) error {
	url := fmt.Sprintf("%s/config", h.endpoint)
	jsonData, _ := json.Marshal(cfg)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+h.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
