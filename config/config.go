package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"tubr/store"
)

var CONFIG_PATH = flag.String("cfgPath", "./config.json", "config file path")

type Config struct {
	PostgresConfig store.PostgresConfig `json:"postgres"`
}

func FromENV() *Config {
	var config Config

	fs, err := os.Open(*CONFIG_PATH)
	if err != nil {
		log.Fatalf("error: failed to open config file: %+v", err)
	}

	cfgData, err := io.ReadAll(fs)
	if err != nil {
		log.Fatalf("error: failed to read config: %+v", err)
	}

	if err = json.Unmarshal(cfgData, &config); err != nil {
		log.Fatalf("error: failed to decode config: %+v", err)
	}

	fmt.Println(config)

	return &config
}
