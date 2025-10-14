package results

import (
	"lsf-configurator/pkg/core"

	"github.com/go-redis/redis/v7"
)

type redisResultsClient struct {
	RedisUrl string
	RedisCli redis.Client
}

const resultsKeyPrefix = "result:"

func NewRedisResultsClient(redisUrl string) core.ResultsClient {
	return &redisResultsClient{
		RedisUrl: redisUrl,
		RedisCli: *redis.NewClient(&redis.Options{Addr: redisUrl}),
	}
}

func (rc *redisResultsClient) GetAppResults(appId string, count int) ([]string, error) {
	return rc.RedisCli.LRange(resultsKeyPrefix+appId, 0, int64(count-1)).Result()
}
