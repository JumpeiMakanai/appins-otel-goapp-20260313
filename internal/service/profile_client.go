package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Profile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Team string `json:"team"`
}

type ProfileClient struct {
	BaseURL string
	Client  *http.Client
}

func NewProfileClient(baseURL string) *ProfileClient {
	return &ProfileClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (c *ProfileClient) GetProfile(ctx context.Context, id int) (*Profile, error) {
	url := fmt.Sprintf("%s/profile?id=%d", c.BaseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile service returned status %d", resp.StatusCode)
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}