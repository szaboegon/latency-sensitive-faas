package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
)

const (
	metricsIndex = ".ds-metrics-apm.*"
	tracesIndex  = ".ds-traces-apm*"
)

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

func NewMetricsReader(backendAddr string) (MetricsReader, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			backendAddr,
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

func (c metricsClient) QueryNodeMetrics() ([]NodeMetrics, error) {
	size := 1
	nameField := "kubernetes.node.name"
	cpuUtilField := "k8s.node.cpu.utilization"
	nodesAggName := "nodes"
	metricsAggName := "latest_metrics"
	res, err := c.client.Search().
		Size(0).
		Index(metricsIndex).
		Request(&search.Request{
			Query: &types.Query{
				Bool: &types.BoolQuery{
					Must: []types.Query{
						{
							Exists: &types.ExistsQuery{
								Field: cpuUtilField,
							},
						},
						{
							Exists: &types.ExistsQuery{
								Field: nameField,
							},
						},
					},
				},
			},
			Aggregations: map[string]types.Aggregations{
				nodesAggName: {
					Terms: &types.TermsAggregation{
						Field: &nameField,
					},
					Aggregations: map[string]types.Aggregations{
						metricsAggName: {
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

	aggregation := res.Aggregations[nodesAggName]
	nodesAgg, ok := aggregation.(*types.StringTermsAggregate)
	if !ok {
		return nil, errors.New("incorrect aggregation type")
	}

	buckets, ok := nodesAgg.Buckets.([]types.StringTermsBucket)
	if !ok {
		return nil, errors.New("incorrect bucket type")
	}
	ret := []NodeMetrics{}
	for _, bucket := range buckets {
		aggregation = bucket.Aggregations[metricsAggName]
		bucketAgg := aggregation.(*types.TopHitsAggregate)
		source := bucketAgg.Hits.Hits[0].Source_

		nodeMetrics, err := unmarshalSource(source)
		if err != nil {
			return nil, err
		}
		nodeMetrics.Node = Node(bucket.Key.(string))
		ret = append(ret, nodeMetrics)
	}

	return ret, nil
}

// TODO needs to be tested, to ensure query and calculation is correct
func (c metricsClient) QueryAverageAppRuntime(appId string) (float64, error) {
	size := 1
	appNameField := "labels.app_name"
	traceIdField := "trace.id"
	timestampField := "@timestamp"

	res, err := c.client.Search().
		Index(tracesIndex).
		Request(&search.Request{
			Query: &types.Query{
				Term: map[string]types.TermQuery{
					appNameField: {
						Value: appId,
					},
				},
			},
			Aggregations: map[string]types.Aggregations{
				"traces": {
					Terms: &types.TermsAggregation{
						Field: &traceIdField,
					},
					Aggregations: map[string]types.Aggregations{
						"first_span": {
							TopHits: &types.TopHitsAggregation{
								Sort: []types.SortCombinations{
									types.SortOptions{
										SortOptions: map[string]types.FieldSort{
											timestampField: {
												Order: &sortorder.Asc,
											},
										},
									},
								},
								Size: &size,
							},
						},
						"last_span": {
							TopHits: &types.TopHitsAggregation{
								Sort: []types.SortCombinations{
									types.SortOptions{
										SortOptions: map[string]types.FieldSort{
											timestampField: {
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
		return 0, err
	}

	tracesAgg, ok := res.Aggregations["traces"].(*types.StringTermsAggregate)
	if !ok {
		return 0, errors.New("incorrect aggregation type for traces")
	}

	var totalDuration float64
	var traceCount int

	for _, bucket := range tracesAgg.Buckets.([]types.StringTermsBucket) {
		firstSpanAgg := bucket.Aggregations["first_span"].(*types.TopHitsAggregate)
		lastSpanAgg := bucket.Aggregations["last_span"].(*types.TopHitsAggregate)

		if len(firstSpanAgg.Hits.Hits) == 0 || len(lastSpanAgg.Hits.Hits) == 0 {
			continue
		}

		var firstSpan, lastSpan map[string]interface{}
		if err := json.Unmarshal(firstSpanAgg.Hits.Hits[0].Source_, &firstSpan); err != nil {
			return 0, err
		}
		if err := json.Unmarshal(lastSpanAgg.Hits.Hits[0].Source_, &lastSpan); err != nil {
			return 0, err
		}

		firstTimestamp, ok := firstSpan[timestampField].(string)
		if !ok {
			return 0, errors.New("missing or invalid first span timestamp")
		}
		lastTimestamp, ok := lastSpan[timestampField].(string)
		if !ok {
			return 0, errors.New("missing or invalid last span timestamp")
		}

		lastSpanSpanField, ok := lastSpan["span"].(map[string]interface{})
		if !ok {
			return 0, errors.New("missing or invalid span object in last span")
		}

		lastSpanDurationField, ok := lastSpanSpanField["duration"].(map[string]interface{})
		if !ok {
			return 0, errors.New("missing or invalid last span duration")
		}

		lastSpanDuration, ok := lastSpanDurationField["us"].(float64)
		if !ok {
			return 0, errors.New("missing or invalid last span duration")
		}

		firstTime, err := time.Parse(time.RFC3339, firstTimestamp)
		if err != nil {
			return 0, errors.New("invalid first span timestamp format")
		}
		lastTime, err := time.Parse(time.RFC3339, lastTimestamp)
		if err != nil {
			return 0, errors.New("invalid last span timestamp format")
		}

		// Calculate trace duration including the last span's duration
		duration := lastTime.Sub(firstTime).Seconds() + (lastSpanDuration / 1e6) // Convert microseconds to seconds
		totalDuration += duration
		traceCount++
	}

	if traceCount == 0 {
		return 0, errors.New("no traces found")
	}

	averageRuntime := totalDuration / float64(traceCount)
	return averageRuntime, nil
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
