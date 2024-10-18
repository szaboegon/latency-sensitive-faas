package metrics

import (
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
)

var metricsIndex string = ".ds-metrics-apm.*"

type ApiKeyConfig struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	ApiKey              string `json:"api_key"`
	Encoded             string `json:"encoded"`
	BeatsLogstashFormat string `json:"beats_logstash_format"`
}

type metricsClient struct {
	client *elasticsearch.TypedClient
}

func NewMetricsReader(apiKeyConfig ApiKeyConfig) (MetricsReader, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
		//APIKey: apiKeyConfig.ApiKey,
	}
	es, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		return nil, err
	}

	return &metricsClient{
		client: es,
	}, nil
}

func (c metricsClient) GetNodeCpuUtilizations() (map[string]NodeMetrics, error) {
	size := 1
	_, err := c.client.Search().
		Index(metricsIndex).
		Request(&search.Request{
			Query: &types.Query{
				Bool: &types.BoolQuery{
					Must: []types.Query{
						{
							Exists: &types.ExistsQuery{
								Field: "k8s.node.cpu.utilization",
							},
						},
						{
							Exists: &types.ExistsQuery{
								Field: "kubernetes.node.name",
							},
						},
					},
				},
			},
			Aggregations: map[string]types.Aggregations{
				"cpu_utilization": {
					TopHits: &types.TopHitsAggregation{
						Sort: []types.SortCombinations{
							types.SortOptions{
								SortOptions: map[string]types.FieldSort{
									"@timestamp": types.FieldSort{
										Order: &sortorder.Desc,
									},
								},
							},
						},
						Size: &size,
					},
				},
			},
		}).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c metricsClient) Test() (string, error) {
	// size := 1
	// res, err := c.client.Search().
	// 	Index(metricsIndex).
	// 	Request(&search.Request{
	// 		Query: &types.Query{
	// 			Bool: &types.BoolQuery{
	// 				Must: []types.Query{
	// 					{
	// 						Exists: &types.ExistsQuery{
	// 							Field: "k8s.node.cpu.utilization",
	// 						},
	// 					},
	// 					{
	// 						Exists: &types.ExistsQuery{
	// 							Field: "kubernetes.node.name",
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 		Aggregations: map[string]types.Aggregations{
	// 			"cpu_utilization": {
	// 				TopHits: &types.TopHitsAggregation{
	// 					Sort: []types.SortCombinations{
	// 						types.SortOptions{
	// 							SortOptions: map[string]types.FieldSort{
	// 								"@timestamp": types.FieldSort{
	// 									Order: &sortorder.Desc,
	// 								},
	// 							},
	// 						},
	// 					},
	// 					Size: &size,
	// 				},
	// 			},
	// 		},
	// 	}).
	// 	Do(context.Background())
	res, err := c.client.Search().
		Index(metricsIndex).Do(context.Background())
	if err != nil {
		return "", err
	}

	test, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return string(test), nil
}
