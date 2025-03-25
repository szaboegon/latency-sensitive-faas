package routing

import (
	"bytes"
	"encoding/json"
	"log"
	"lsf-configurator/pkg/core"
	"net/http"
)

const (
	SetRouteTableApi = "/routing_table/{appId}"
)

type RouteConfigurator struct {
	LoadBalancerUrls []string
	httpClient       *http.Client
}

func NewRouteConfigurator(loadBalancerUrls []string) core.RoutingClient {
	return &RouteConfigurator{
		LoadBalancerUrls: loadBalancerUrls,
		httpClient:       &http.Client{},
	}
}

// TODO improve on error handling
func (rc *RouteConfigurator) SetRoutingTable(appId string, rt core.RoutingTable) error {
	data, err := json.Marshal(rt)
	if err != nil {
		return err
	}
	for _, url := range rc.LoadBalancerUrls {
		req, err := http.NewRequest(http.MethodPut, url+SetRouteTableApi, bytes.NewBuffer(data))
		if err != nil {
			return err
		}

		resp, err := rc.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			log.Printf("failed to set routing table in load balancer with url %s. status code: %v", url, resp.StatusCode)
		}
	}
	return nil
}
