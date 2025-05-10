package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/big"

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
	size := 100
	appNameField := "labels.app_name"
	traceIdField := "trace.id"
	timestampField := "timestamp.us"

	res, err := c.client.Search().
		Index(tracesIndex).
		Request(&search.Request{
			Query: &types.Query{
				Bool: &types.BoolQuery{ // Combine Term and Exists queries
					Must: []types.Query{
						{
							Term: map[string]types.TermQuery{
								appNameField: types.TermQuery{
									Value: appId,
								},
							},
						},
					},
				},
			},
			Aggregations: map[string]types.Aggregations{
				"traces": {
					Terms: &types.TermsAggregation{
						Field: &traceIdField,
						Size:  &size,
					},
					Aggregations: map[string]types.Aggregations{
						"first_span": {
							TopHits: &types.TopHitsAggregation{
								Sort: []types.SortCombinations{
									types.SortOptions{
										SortOptions: map[string]types.FieldSort{
											timestampField: { // Updated field
												Order: &sortorder.Asc,
											},
										},
									},
								},
							},
						},
						"all_spans": {
							TopHits: &types.TopHitsAggregation{
								Sort: []types.SortCombinations{
									types.SortOptions{
										SortOptions: map[string]types.FieldSort{
											timestampField: { // Updated field
												Order: &sortorder.Desc,
											},
										},
									},
								},
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

	totalDuration := big.NewFloat(0)
	traceCount := 0

	for _, bucket := range tracesAgg.Buckets.([]types.StringTermsBucket) {
		firstSpanAgg := bucket.Aggregations["first_span"].(*types.TopHitsAggregate)
		allSpansAgg := bucket.Aggregations["all_spans"].(*types.TopHitsAggregate)

		if len(firstSpanAgg.Hits.Hits) == 0 || len(allSpansAgg.Hits.Hits) == 0 {
			continue
		}

		firstTimestampUs, err := unmarshalSpanTimestamp(firstSpanAgg.Hits.Hits[0].Source_)
		if err != nil {
			return 0, err
		}

		// Find the span that ends the latest
		var latestEndTimeUs float64 = 0
		for _, hit := range allSpansAgg.Hits.Hits {

			spanTimestampUs, err := unmarshalSpanTimestamp(hit.Source_)
			if err != nil {
				log.Print("Error unmarshalling span timestamp:", err)
				continue
			}

			spanDurationUs, err := unmarshalSpanDuration(hit.Source_)
			if err != nil {
				log.Print("Error unmarshalling span duration:", err)
				continue
			}

			endTimeUs := spanTimestampUs + spanDurationUs
			if endTimeUs > latestEndTimeUs {
				latestEndTimeUs = endTimeUs
			}
		}

		if latestEndTimeUs == 0 || latestEndTimeUs < firstTimestampUs {
			log.Print("No valid end time found for trace")
			continue
		}

		duration := (latestEndTimeUs - firstTimestampUs) / 1e6
		totalDuration.Add(totalDuration, big.NewFloat(duration))
		traceCount++
	}

	if traceCount == 0 {
		return 0, errors.New("no traces found")
	}

	averageRuntime, _ := new(big.Float).Quo(totalDuration, big.NewFloat(float64(traceCount))).Float64()
	return averageRuntime, nil
}

func unmarshalSpanTimestamp(jsonMessage json.RawMessage) (float64, error) {
	var span map[string]interface{}
	if err := json.Unmarshal(jsonMessage, &span); err != nil {
		return 0, err
	}

	spanTimestamp, ok := span["timestamp"].(map[string]interface{})
	if !ok {
		return 0, errors.New("missing or invalid span timestamp")
	}

	spanTimestampUs, ok := spanTimestamp["us"].(float64)
	if !ok {
		return 0, errors.New("missing or invalid first span timestamp in microseconds")
	}

	// For debugging purposes, you can log the span ID and timestamp
	spanField, ok := span["span"].(map[string]interface{})
	if !ok {
		return 0, errors.New("missing or invalid span object")
	}

	if spanId, ok := spanField["id"].(string); ok {
		log.Printf("Span ID: %s, Timestamp (us): %f", spanId, spanTimestampUs)
	}

	return spanTimestampUs, nil
}

func unmarshalSpanDuration(jsonMessage json.RawMessage) (float64, error) {
	var span map[string]interface{}
	if err := json.Unmarshal(jsonMessage, &span); err != nil {
		return 0, err
	}

	spanField, ok := span["span"].(map[string]interface{})
	if !ok {
		return 0, errors.New("missing or invalid span object")
	}

	durationField, ok := spanField["duration"].(map[string]interface{})
	if !ok {
		return 0, errors.New("missing or invalid duration object")
	}

	spanDuration, ok := durationField["us"].(float64)
	if !ok {
		return 0, errors.New("missing or invalid span duration")
	}

	// For debugging purposes, you can log the span ID and timestamp
	if spanId, ok := spanField["id"].(string); ok {
		log.Printf("Span ID: %s, Duration (us): %f", spanId, spanDuration)
	}

	return spanDuration, nil
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
