package config

import (
	"os"
	"reflect"
	"strconv"

	"github.com/apex/log"
	"github.com/joho/godotenv"
)

type Configuration struct {
	UploadDir        string `env:"UPLOAD_DIR"`
	TemplatesPath    string `env:"TEMPLATES_PATH"`
	ImageRegistry    string `env:"IMAGE_REGISTRY"`
	RegistryUser     string `env:"REGISTRY_USER"`
	RegistryPassword string `env:"REGISTRY_PASSWORD"`
	BuilderImage     string `env:"BUILDER_IMAGE"`
}

func Init() Configuration {
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found. Using system environment variables")
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
			continue // Skip empty environment variables
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
