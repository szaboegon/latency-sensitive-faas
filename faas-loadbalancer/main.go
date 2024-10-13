package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

func main() {
	// Initialize the Elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200", // Replace with your Elasticsearch endpoint
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	// Example query to retrieve APM metrics
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"service.name": "your-service-name", // Replace with your service name
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("apm-*-metrics-*"), // Replace with your APM metrics index pattern
		es.Search.WithBody(&buf),
	)
	if err != nil {
		log.Fatalf("Error performing search request: %s", err)
	}
	defer res.Body.Close()

	// Handle the response
	if res.IsError() {
		log.Fatalf("Error response: %s", res.Status())
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	// Print the hits
	fmt.Printf("Search results:\n%s\n", response)
}
