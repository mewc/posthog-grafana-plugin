package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PostHogClient struct {
	baseURL    string
	projectID  string
	apiKey     string
	httpClient *http.Client
}

func NewPostHogClient(baseURL, projectID, apiKey string) *PostHogClient {
	return &PostHogClient{
		baseURL:   baseURL,
		projectID: projectID,
		apiKey:    apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *PostHogClient) ExecuteHogQL(ctx context.Context, query string) (*HogQLAPIResponse, error) {
	reqBody := HogQLAPIRequest{
		Query: HogQLQueryBody{
			Kind:  "HogQLQuery",
			Query: query,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/projects/%s/query/", c.baseURL, c.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiResp HogQLAPIResponse
		if json.Unmarshal(respBody, &apiResp) == nil && apiResp.Detail != "" {
			return nil, fmt.Errorf("PostHog API error (%d): %s", resp.StatusCode, apiResp.Detail)
		}
		return nil, fmt.Errorf("PostHog API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp HogQLAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &apiResp, nil
}

func (c *PostHogClient) TestConnection(ctx context.Context) error {
	_, err := c.ExecuteHogQL(ctx, "SELECT 1")
	return err
}
