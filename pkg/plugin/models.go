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
	Types   []string        `json:"-"`
	Results [][]interface{} `json:"results"`
	Error   string          `json:"error"`
	Detail  string          `json:"detail"`
}

// UnmarshalJSON handles PostHog's types field which can be either:
// - ["String", "Int64"] (old format)
// - [["String", "String"], ["Int64", "Int64"]] (new format: [clickhouse_type, posthog_type])
func (r *HogQLAPIResponse) UnmarshalJSON(data []byte) error {
	type Alias HogQLAPIResponse
	aux := &struct {
		Types json.RawMessage `json:"types"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(aux.Types) == 0 {
		return nil
	}

	// Try array of strings first: ["String", "Int64"]
	var stringTypes []string
	if err := json.Unmarshal(aux.Types, &stringTypes); err == nil {
		r.Types = stringTypes
		return nil
	}

	// Try array of arrays: [["String", "String"], ["Int64", "Int64"]]
	var arrayTypes [][]string
	if err := json.Unmarshal(aux.Types, &arrayTypes); err == nil {
		r.Types = make([]string, len(arrayTypes))
		for i, pair := range arrayTypes {
			if len(pair) > 0 {
				r.Types[i] = pair[0] // Use the first element (ClickHouse type)
			}
		}
		return nil
	}

	return fmt.Errorf("unexpected types format: %s", string(aux.Types))
}
