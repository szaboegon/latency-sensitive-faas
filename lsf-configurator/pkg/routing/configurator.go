package routing

import (
	"encoding/json"
	"fmt"
	"log"
	"lsf-configurator/pkg/core"

	redis "github.com/go-redis/redis/v7"
)

type RouteConfigurator struct {
	RedisUrl string
	RedisCli redis.Client
}

func NewRouteConfigurator(redisUrl string) core.RoutingClient {
	return &RouteConfigurator{
		RedisUrl: redisUrl,
		RedisCli: *redis.NewClient(&redis.Options{Addr: redisUrl}),
	}
}

func (rc *RouteConfigurator) SetRoutingTable(appId string, fc core.FunctionComposition) error {
	rtDto := make(RoutingTable)

	for component, routes := range fc.Components {
		for _, r := range routes {
			routeDto := Route{
				Component: r.To,
				Url:       getFunctionUrl(core.UniqueFcId(appId, r.Function), fc.NameSpace),
			}
			rtDto[string(component)] = append(rtDto[string(component)], routeDto)
		}
	}

	data, err := json.Marshal(rtDto)
	if err != nil {
		return err
	}

	err = rc.RedisCli.Set(fc.Id, data, 0).Err()
	if err != nil {
		return err
	}

	log.Printf("Routing table for %s set successfully, rt: %s", fc.Id, data)
	return nil
}

func getFunctionUrl(fcId, namespace string) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local", fcId, namespace)
}
