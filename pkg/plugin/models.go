package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PostHogSettings struct {
	PostHogURL string `json:"posthogUrl"`
	ProjectID  string `json:"projectId"`
}

type PostHogSecrets struct {
	APIKey string `json:"apiKey"`
}

func LoadSettings(source backend.DataSourceInstanceSettings) (*PostHogSettings, *PostHogSecrets, error) {
	settings := PostHogSettings{}
	if err := json.Unmarshal(source.JSONData, &settings); err != nil {
		return nil, nil, fmt.Errorf("could not unmarshal settings: %w", err)
	}

	secrets := &PostHogSecrets{
		APIKey: source.DecryptedSecureJSONData["apiKey"],
	}

	return &settings, secrets, nil
}

type PostHogQuery struct {
	QueryType string `json:"queryType"`
	RawHogQL  string `json:"rawHogQL"`
}

type HogQLAPIRequest struct {
	Query HogQLQueryBody `json:"query"`
}

type HogQLQueryBody struct {
	Kind  string `json:"kind"`
	Query string `json:"query"`
}

type HogQLAPIResponse struct {
	Columns []string        `json:"columns"`
	Types   []string        `json:"types"`
	Results [][]interface{} `json:"results"`
	Error   string          `json:"error"`
	Detail  string          `json:"detail"`
}
