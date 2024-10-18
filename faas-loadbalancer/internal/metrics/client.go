package metrics

import (
	"context"
	"encoding/json"
	"errors"

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

func (c metricsClient) GetNodeMetrics() (map[string]NodeMetrics, error) {
	size := 1
	nameField := "kubernetes.node.name"
	res, err := c.client.Search().
		Size(0).
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
				"nodes": {
					Terms: &types.TermsAggregation{
						Field: &nameField,
					},
					Aggregations: map[string]types.Aggregations{
						"latest_metrics": {
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
				},
			},
		}).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	aggregation := res.Aggregations["nodes"]
	nodesAgg, ok := aggregation.(*types.StringTermsAggregate)
	if !ok {
		return nil, errors.New("incorrect aggregation type")
	}

	buckets, ok := nodesAgg.Buckets.([]types.StringTermsBucket)
	if !ok {
		return nil, errors.New("incorrect bucket type")
	}
	nodeMetricsMap := make(map[string]NodeMetrics)
	for _, bucket := range buckets {
		aggregation = bucket.Aggregations["latest_metrics"]
		bucketAgg := aggregation.(*types.TopHitsAggregate)
		source := bucketAgg.Hits.Hits[0].Source_

		nodeMetrics, err := unmarshalSource(source)
		if err != nil {
			return nil, err
		}

		nodeMetricsMap[bucket.Key.(string)] = nodeMetrics
	}

	return nodeMetricsMap, nil
}

func (c metricsClient) Test() (string, error) {
	size := 1
	nameField := "kubernetes.node.name"
	res, _ := c.client.Search().
		Size(0).
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
				"nodes": {
					Terms: &types.TermsAggregation{
						Field: &nameField,
					},
					Aggregations: map[string]types.Aggregations{
						"latest_metrics": {
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
				},
			},
		}).
		Do(context.Background())

	aggregation := res.Aggregations["nodes"]
	nodesAgg, _ := aggregation.(*types.StringTermsAggregate)

	buckets, _ := nodesAgg.Buckets.([]types.StringTermsBucket)
	for _, bucket := range buckets {
		aggregation = bucket.Aggregations["latest_metrics"]
		bucketAgg := aggregation.(*types.TopHitsAggregate)
		source := bucketAgg.Hits.Hits[0].Source_
		return string(source), nil
	}

	//
	// for i, v := range res.Aggregations["nodes"] {

	// }

	return "asd", nil
}

func unmarshalSource(source json.RawMessage) (NodeMetrics, error) {
	var sourceMap map[string]json.RawMessage
	err := json.Unmarshal(source, &sourceMap)
	if err != nil {
		return NodeMetrics{}, err
	}

	var k8s map[string]json.RawMessage
	err = json.Unmarshal(sourceMap["k8s"], &k8s)
	if err != nil {
		return NodeMetrics{}, err
	}

	var nodeMetrics NodeMetrics
	err = json.Unmarshal(k8s["node"], &nodeMetrics)
	if err != nil {
		return nodeMetrics, err
	}

	return nodeMetrics, nil
}
