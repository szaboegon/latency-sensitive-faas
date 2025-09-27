package config

import (
	"log"
	"os"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
)

type Configuration struct {
	UploadDir              string `env:"UPLOAD_DIR"`
	TemplatesPath          string `env:"TEMPLATES_PATH"`
	ImageRegistry          string `env:"IMAGE_REGISTRY"`
	ImageRepository        string `env:"IMAGE_REPOSITORY"`
	RedisUrl               string `env:"REDIS_URL"`
	VerboseLogs            bool   `env:"VERBOSE_LOGS"`
	MetricsBackendAddress  string `env:"METRICS_BACKEND_ADDRESS"`
	TektonNamespace        string `env:"TEKTON_NAMESPACE"`
	TektonPipeline         string `env:"TEKTON_PIPELINE"`
	TektonNotifyURL        string `env:"TEKTON_NOTIFY_URL"`
	TektonWorkspacePVC     string `env:"TEKTON_WORKSPACE_PVC"`
	TektonServiceAccount   string `env:"TEKTON_SERVICE_ACCOUNT"`
	TektonConcurrencyLimit int    `env:"TEKTON_CONCURRENCY_LIMIT"`
	DatabasePath           string `env:"DATABASE_PATH"`
	AlertingApiUrl         string `env:"ALERTING_API_URL"`
	AlertingUsername       string `env:"ALERTING_USERNAME"`
	AlertingPassword       string `env:"ALERTING_PASSWORD"`
	AlertsIndex            string `env:"ALERTS_INDEX" default:"latency-alerts"`
	AlertingConnector      string `env:"ALERTING_CONNECTOR" default:"lsf-alerts-connector"`
	LocalMode              bool   `env:"LOCAL_MODE" default:"false"`
	DeployNamespace        string `env:"DEPLOY_NAMESPACE" default:"application"`
}

func Init() Configuration {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found. Using system environment variables")
	}

	conf := Configuration{}
	confVal := reflect.ValueOf(&conf).Elem()
	confType := confVal.Type()

	for i := 0; i < confVal.NumField(); i++ {
		field := confType.Field(i)
		envVar := field.Tag.Get("env")

		if envVar == "" {
			continue // Skip fields without an env tag
		}

		envValue := os.Getenv(envVar)
		if envValue == "" {
			envValue = field.Tag.Get("default")
		}

		if envValue == "" {
			continue
		}

		fieldVal := confVal.Field(i)

		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(envValue)

		case reflect.Int:
			intValue, err := strconv.Atoi(envValue)
			if err == nil {
				fieldVal.SetInt(int64(intValue))
			} else {
				log.Fatalf("Invalid int value for %s: %s", envVar, envValue)
			}

		case reflect.Bool:
			boolValue, err := strconv.ParseBool(envValue)
			if err == nil {
				fieldVal.SetBool(boolValue)
			} else {
				log.Fatalf("Invalid bool value for %s: %s", envVar, envValue)
			}
		}
	}

	return conf
}
