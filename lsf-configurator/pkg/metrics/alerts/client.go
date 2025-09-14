package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type alertClient struct {
	alertApiUrl string
	username    string
	password    string
}

func NewAlertClient(alertApiUrl string, username string, password string) AlertClient {
	return &alertClient{
		alertApiUrl: alertApiUrl,
		username:    username,
		password:    password,
	}
}

func (c *alertClient) EnsureAlertConnector(ctx context.Context, connectorName, indexName string) (string, error) {
	// 1Check if connector exists
	url := fmt.Sprintf("%s/api/actions/connectors", c.alertApiUrl)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("kbn-xsrf", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to list connectors: %s", string(data))
	}

	var connectors []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &connectors); err != nil {
		return "", err
	}

	for _, c := range connectors {
		if c.Name == connectorName {
			return c.ID, nil // connector exists
		}
	}

	// 2️⃣ Create connector if it does not exist
	body := fmt.Sprintf(`{
		"name": "%s",
		"connector_type_id": ".index",
		"config": { "index": "%s" },
		"secrets": {}
	}`, connectorName, indexName)

	req, _ = http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/actions/connector", c.alertApiUrl), strings.NewReader(body))
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ = io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to create connector: %s", string(data))
	}

	var respJSON struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &respJSON); err != nil {
		return "", err
	}

	log.Printf("Index connector %s created with ID %s", connectorName, respJSON.ID)
	return respJSON.ID, nil
}

func (c *alertClient) CreateAlert(ctx context.Context, serviceName string, latencyThresholdMs int, connectorId string) (string, error) {
	url := fmt.Sprintf("%s/api/alerting/rule", c.alertApiUrl)

	// Kibana ES query for latency
	esQuery := fmt.Sprintf(`{"query":{"range":{"transaction.duration.us":{"gt":%d}}}}`, latencyThresholdMs*1000)

	// Request body
	reqBody := map[string]interface{}{
		"rule_type_id": ".es-query",
		"consumer":     "alerts",
		"name":         fmt.Sprintf("Latency Rule - %s", serviceName),
		"tags":         []string{"latency", serviceName},
		"enabled":      true,
		"schedule": map[string]string{
			"interval": "1m",
		},
		"notify_when": "onActionGroupChange",
		"params": map[string]interface{}{
			"index":     []string{"traces-*"},
			"timeField": "@timestamp",
			"esQuery":   esQuery,
			"size":      1,
		},
		"actions": []map[string]interface{}{
			{
				"group": "query matched",
				"id":    connectorId,
				"params": map[string]interface{}{
					"documents": []map[string]interface{}{
						{
							"rule":       fmt.Sprintf("High Latency - %s", serviceName),
							"status":     "firing",
							"service":    serviceName,
							"latency":    "{{context.value}}",
							"@timestamp": "{{context.date}}",
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to create alert: %s", string(data))
	}

	// Parse returned rule ID
	var respJSON struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &respJSON); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return respJSON.ID, nil
}
